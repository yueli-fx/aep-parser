// AddMask — splice a fresh mask atom into a layer's "ADBE Mask Parade".
//
// AE stores each mask as a (tdmn "ADBE Mask Atom", mkif[48B], LIST:tdgp)
// TRIPLE inside the parade tdgp, terminated by a lone "ADBE Group End" tdmn
// sentinel — one chunk more than the Effect Parade's (tdmn, sspc) pair: the
// mkif carries mode / inverted / locked / motion-blur / internal index /
// label color. The atom tdgp holds tdsb + tdsn (the user-visible mask name —
// AE persists it here, omtn stays empty) + the "ADBE Mask Shape" om-s + Group
// End; Feather / Opacity / Expansion are default-elided exactly like effect
// params, so a fresh atom omits them.
//
// Unlike AddEffect there is no embedded template: every chunk of the atom is
// built from scratch. The "ADBE Mask Shape" om-s reuses the ship-gated shape
// path encoding (encodeBezier — mask ldat triples follow the same
// [anchor, this-out-control, next-in-control] absolute bbox-normalized layout
// RE'd for "ADBE Vector Shape"; verified against re_batch.aep masks with
// known JSX input), which is also why the mask path is parameterizable at
// creation time even though mutating an EXISTING mask's path stays refused
// (structural write).
package serializer

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// makeMaskMkif builds the 48-byte mkif info chunk for a fresh mask atom.
// Canonical non-zero bytes replicated from AE-2020-saved atoms (re_batch.aep):
//
//	@0x00 inverted=0  @0x01 locked=0  @0x02 motion-blur=0 (same as layer)
//	@0x04 u32 mode (MaskModeAdd)
//	@0x08 u32 internal mask index (monotonic per layer, AE keeps gaps)
//	@0x0C "om" tag bytes 6F 6D 00 00
//	@0x10 0x0E…  @0x18 0x0F…  (constants, semantics unknown)
//	@0x20 u32 save-timestamp (AE stamps save time; any value accepted)
//	@0x24 65 C0 00 00 (constant)
//	@0x2C FF + RGB label color (AE's first-mask yellow E4 D8 4C)
func makeMaskMkif(index uint32) *rifx.Chunk {
	d := make([]byte, 48)
	binary.BigEndian.PutUint32(d[0x04:0x08], uint32(MaskModeAdd))
	binary.BigEndian.PutUint32(d[0x08:0x0C], index)
	copy(d[0x0C:0x10], []byte{0x6F, 0x6D, 0x00, 0x00})
	d[0x10] = 0x0E
	d[0x18] = 0x0F
	copy(d[0x20:0x28], []byte{0x5F, 0xDB, 0x1D, 0x04, 0x65, 0xC0, 0x00, 0x00})
	d[0x2C] = 0xFF
	d[0x2D], d[0x2E], d[0x2F] = 0xE4, 0xD8, 0x4C
	return &rifx.Chunk{ID: rifx.IDMkif, Data: d}
}

// maskShapeTdb4 is the 124-byte tdb4 of a static mask's "ADBE Mask Shape"
// om-s tdbs, byte-copied from an AE-2020-saved atom (re_mask_open.aep;
// re_batch.aep differs only at @0x0C..0x0F = 5D A8 — both accepted). It differs
// from the canonical scalar tdb4 makeTdb4 emits — @0x04..0x0B reads
// 00 07 00 01 00 02 00 07 (shape/path stream markers, vs 00 01 + header byte
// for value streams) and @0x10.. carries 0.0001 + four 1.0 doubles. AE 2020
// hard-crashes on open ("After Effects 已崩溃 (0 :: 42)") when a mask shape
// carries the canonical scalar head instead — the same head IS accepted inside
// a shape layer's "ADBE Vector Shape" (ship-gated), so the strictness is
// mask-specific: AE eagerly decodes mask outlines on project open.
var maskShapeTdb4 = []byte{
	0xDB, 0x99, 0x00, 0x01, 0x00, 0x07, 0x00, 0x01, 0x00, 0x02, 0x00, 0x07, 0x00, 0x00, 0x78, 0x00,
	0x3F, 0x1A, 0x36, 0xE2, 0xEB, 0x1C, 0x43, 0x2D, 0x3F, 0xF0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x3F, 0xF0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x3F, 0xF0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x3F, 0xF0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

// maskLayerDims resolves the divisor that maps layer pixels to the mask
// shape's on-disk unit. Mask coordinates are stored as FRACTIONS of the
// SOURCE ITEM's pixel space (every AE-saved solid-layer mask carries a
// 0..1-ish shph bbox; writing pixels there made AE read vertices ×layer-size
// — caught by the AE-2020 gate readback). Source-less layers (shape / text)
// have no item space: AE reads their stored values back 1:1 as pixels, so the
// divisor is 1 (gate-observed: a shape-layer mask written ÷compSize read back
// as the fraction values, not pixels).
func maskLayerDims(layer *Layer) (w, h float64) {
	if layer.SourceID == 0 {
		return 1, 1
	}
	comp := scene.LayerComp(layer)
	if comp == nil {
		return 0, 0
	}
	proj := scene.CompositionProj(comp)
	if proj == nil {
		return 0, 0
	}
	switch src := proj.AVItemByID(layer.SourceID).(type) {
	case *Footage:
		if src.Width > 0 && src.Height > 0 {
			return float64(src.Width), float64(src.Height)
		}
	case *Composition:
		return float64(src.Width), float64(src.Height)
	}
	return 0, 0
}

// makeMaskShapeOmS builds the static "ADBE Mask Shape" value: a LIST(om-s)
// holding the descriptor tdbs (tdsb + tdsn placeholder + the mask-specific
// tdb4 + 4B zero cdat) and a LIST(omks) with one LIST(shap) — the same shap
// structure LowerPathStream emits for a static shape path. The path arrives
// already scaled to layer-fraction units (see maskLayerDims). The mask closed
// flag additionally lives at shph[0x14] (the byte SetClosed patches);
// encodeBezier emits 0x01 there unconditionally, so open paths patch it back
// to 0.
func makeMaskShapeOmS(path BezierPath) *rifx.Chunk {
	innerTdbs := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdbs, Children: []*rifx.Chunk{
		makeTdsb(),
		makeTdsn(aeDefaultGroupName),
		{ID: rifx.IDtdb4, Data: append([]byte(nil), maskShapeTdb4...)},
		{ID: rifx.IDCdat, Data: make([]byte, 4)},
	}}

	shph, lhd3, ldat := encodeBezier(path)
	// Mask encoding deviates from the shape-path convention encodeBezier
	// follows, in ways that crash AE 2020 on open or throw 参数值无效 on the
	// mask-shape read (both gate-observed; ground truth = re_mask_open.aep,
	// AE-2020-saved open + closed masks across solid/shape layers):
	//   shph[3]  — 0x01 closed, 0x09 open (bit3 = OPEN; [0x14] stays 0x01 on
	//              every mask, open or not — it is NOT the closed flag).
	//   lhd3@0x14 — constant 4 (encodeBezier writes the vertex count, which
	//              only coincides at n=4 — n=3 masks hard-crashed AE 2020).
	//   lhd3@0x18 — constant 1 (not a closed flag on masks).
	//   lhd3@0x1C — 4·n (encodeBezier's constant 16 again only fits n=4).
	if len(shph.Data) > 0x14 {
		if path.Closed {
			shph.Data[3] = 0x01
		} else {
			shph.Data[3] = 0x09
		}
		shph.Data[0x14] = 0x01
	}
	if n := len(path.Vertices); len(lhd3.Data) >= 0x20 {
		binary.BigEndian.PutUint32(lhd3.Data[0x14:0x18], 4)
		binary.BigEndian.PutUint32(lhd3.Data[0x18:0x1C], 1)
		binary.BigEndian.PutUint32(lhd3.Data[0x1C:0x20], uint32(4*n))
	}
	kfList := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDkfl, Children: []*rifx.Chunk{lhd3, ldat}}
	shap := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDShap, Children: []*rifx.Chunk{
		shph, kfList, {ID: rifx.IDOmtn},
	}}
	omks := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDOmks, Children: []*rifx.Chunk{shap}}

	return &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDOmS, Children: []*rifx.Chunk{innerTdbs, omks}}
}

// ensureMaskParade returns the layer's Mask Parade, splicing a fresh empty
// parade group into the Layr property tree when absent — immediately before
// "ADBE Effect Parade" when the layer has one, else before "ADBE Transform
// Group" (AE's emitted group order: Mask Parade precedes both).
func ensureMaskParade(layer *Layer) (*AEPropertyGroup, func(), error) {
	if parade := layer.MaskParade(); parade != nil {
		return parade, func() {}, nil
	}
	return spliceEmptyParade(layer, "AddMask", MatchNameGroupMaskParade,
		[]string{MatchNameGroupEffectParade, MatchNameGroupTransform})
}

// AddMask appends a mask with the given display name and static path to the
// layer's Mask Parade (auto-creating the parade for mask-less parsed layers)
// and returns the parsed *Mask.
// (Full contract + RE notes live on the aep.AddMask facade — docgen source.)
func AddMask(layer *Layer, name string, path BezierPath) (*Mask, error) {
	if layer == nil {
		return nil, fmt.Errorf("AddMask: layer is nil")
	}
	if layer.Type == LayerTypeCamera || layer.Type == LayerTypeLight {
		return nil, fmt.Errorf("AddMask: layer %q is a %s layer (AE does not allow masks on camera/light layers)", layer.Name, layer.Type)
	}
	if len(path.Vertices) == 0 {
		return nil, fmt.Errorf("AddMask: path has no vertices")
	}
	lw, lh := maskLayerDims(layer)
	if lw <= 0 || lh <= 0 {
		return nil, fmt.Errorf("AddMask: cannot resolve layer %q pixel dimensions (mask coordinates are stored as layer-space fractions)", layer.Name)
	}
	scaled := BezierPath{Closed: path.Closed,
		Vertices:    make([][2]float64, len(path.Vertices)),
		InTangents:  make([][2]float64, len(path.InTangents)),
		OutTangents: make([][2]float64, len(path.OutTangents)),
	}
	for i, v := range path.Vertices {
		scaled.Vertices[i] = [2]float64{v[0] / lw, v[1] / lh}
	}
	for i, v := range path.InTangents {
		scaled.InTangents[i] = [2]float64{v[0] / lw, v[1] / lh}
	}
	for i, v := range path.OutTangents {
		scaled.OutTangents[i] = [2]float64{v[0] / lw, v[1] / lh}
	}

	index := uint32(1)
	for _, m := range layer.Masks {
		if m.Index >= index {
			index = m.Index + 1
		}
	}
	if name == "" {
		name = fmt.Sprintf("Mask %d", index)
	}

	parade, undoParadeCreate, err := ensureMaskParade(layer)
	if err != nil {
		return nil, err
	}
	pgb := propertyGroupBack(parade)
	if pgb == nil || pgb.chunk == nil {
		undoParadeCreate()
		return nil, fmt.Errorf("AddMask: Mask Parade for layer %q has no chunk back-ref", layer.Name)
	}

	// Build the atom triple from scratch.
	tdmnCh := makeTdmn("ADBE Mask Atom")
	mkifCh := makeMaskMkif(index)
	atomTdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
		makeTdsb(),
		makeTdsn(name),
		makeTdmn("ADBE Mask Shape"),
		makeMaskShapeOmS(scaled),
		makeTdmn("ADBE Group End"),
	}}

	children := pgb.chunk.Children
	insertIdx := len(children)
	for i, ch := range children {
		if ch.ID == rifx.IDTdmn && string(bytes.TrimRight(ch.Data, "\x00")) == "ADBE Group End" {
			insertIdx = i
			break
		}
	}

	// Snapshot for atomic rollback.
	oldChunkChildren := append([]*rifx.Chunk(nil), children...)
	oldSceneChildren := append([]PropertyBase(nil), parade.Children...)
	oldMasks := append([]*Mask(nil), layer.Masks...)
	oldWarningsLen := warningsLen(layer)

	rollback := func() {
		pgb.chunk.Children = oldChunkChildren
		parade.Children = oldSceneChildren
		layer.Masks = oldMasks
		rollbackWarnings(layer, oldWarningsLen)
		undoParadeCreate()
	}

	// Chunk: splice (tdmn, mkif, tdgp) in just before the Group End sentinel.
	spliced := make([]*rifx.Chunk, 0, len(children)+3)
	spliced = append(spliced, children[:insertIdx]...)
	spliced = append(spliced, tdmnCh, mkifCh, atomTdgp)
	spliced = append(spliced, children[insertIdx:]...)
	pgb.chunk.Children = spliced

	// Scene: append a stand-in group node (the parade's scene children are the
	// mask atoms in order; the Group End sentinel is chunk-only).
	atomNode := &AEPropertyGroup{MatchName: "ADBE Mask Atom", Name: name}
	scene.SetPropertyGroupParent(atomNode, parade)
	scene.SetPropertyGroupBack(atomNode, &propertyGroupBackrefs{chunk: atomTdgp})
	parade.Children = append(parade.Children, atomNode)

	// Flat mirror: re-parse the spliced triple so the typed Mask's back-refs
	// point at the spliced chunks.
	var newMask *Mask
	if comp := scene.LayerComp(layer); comp != nil && scene.CompositionProj(comp) != nil {
		ctx := newParseCtxFPS(comp.TickRate, comp.FrameRate, comp.Name, &scene.CompositionProj(comp).Warnings)
		tmpParade := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{tdmnCh, mkifCh, atomTdgp}}
		tmpLayr := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
			makeTdmn(MatchNameGroupMaskParade), tmpParade,
		}}
		masks := parseMasks(tmpLayr, ctx)
		if len(masks) != 1 {
			rollback()
			return nil, fmt.Errorf("AddMask: spliced atom re-parse produced %d masks (want 1)", len(masks))
		}
		newMask = masks[0]
		layer.Masks = append(layer.Masks, newMask)
	}

	if newWarn := newWarningsSince(layer, oldWarningsLen); len(newWarn) > 0 {
		rollback()
		return nil, fmt.Errorf("AddMask: produced %d parser warning(s), rolled back: %v", len(newWarn), newWarn)
	}

	return newMask, nil
}
