// internal/aep/lower_layer.go
//
// Phase 2 Task 2.3 — lowerShapeLayer produces a LIST(Layr) chunk from a
// runtime *ShapeLayer. Children per RE-S1 (empty ShapeLayer canonical):
//
//	ldta (160 B AE 2020 canonical)
//	Utf8 (layer name, length-variable)
//	LIST(tdgp) — layer Transform Group (V2.2: user-facing 2D 5-stream)
//	[if root has shapes:] tdmn("ADBE Root Vectors Group") + LIST(tdgp, root)
//
// Per RE-S2 the Phase-0 fixture observed a 6-axis 3D-compatible Transform
// schema (Position_0 / Position_1 / Orientation / RotateX / RotateY /
// Envir Appear). V2.2 instead emits the user-facing 2D form (Anchor /
// Position / Scale / Rotate Z / Opacity) per spec §3.3a + V1 parser
// convention — AE accepts both per the prompt's hot-path 2D guidance.
// Phase 4 roundtrip is the AE accept gate.
package aep

import (
	"encoding/binary"

	"github.com/example/aep-parser/internal/rifx"
)

// LowerShapeLayerForTest exports lowerShapeLayer for unit tests.
func LowerShapeLayerForTest(s *ShapeLayer) (*rifx.Chunk, error) {
	return lowerShapeLayer(s, NewLowerCtxForTest())
}

// lowerShapeLayer → LIST(Layr).
func lowerShapeLayer(s *ShapeLayer, ctx *lowerCtx) (*rifx.Chunk, error) {
	layr := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDLayr}

	// ldta (160 B AE 2020 canonical per RE-S1; capability matrix doesn't
	// branch — AE 2025's +4 zero-pad tail isn't ours to emit, the higher
	// AE just reads our 160 B fine).
	ldta := &rifx.Chunk{ID: rifx.IDLdta, Data: buildLdtaBytes(s, ctx)}
	layr.Children = append(layr.Children, ldta)

	// Utf8 (layer name). Length-variable; written verbatim.
	layr.Children = append(layr.Children, &rifx.Chunk{
		ID:   rifx.IDUtf8,
		Data: []byte(s.Name),
	})

	// Layer-level property groups live INSIDE an outer LIST(tdgp) — AE
	// rejects the flat-Layr-children form (Phase 5 ship gate FAIL repro;
	// scars/v2-2-aelayer-structure.md fix A). Outer body shape per
	// tolerance.aep:
	//   tdsb + tdsn("") + (tdmn + LIST(tdgp))* + tdmn("ADBE Group End")
	//
	// V2.2 minimum outer body emits: Root Vectors Group (when shapes
	// present) + Transform Group. Additional layer-property groups AE
	// emits at default (Layer Styles / Extrsn Options / Material Options
	// / Audio Group / Layer Sets) get added if AE 2020/25 still rejects.
	outer := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	outer.Children = append(outer.Children, makeTdsb(), makeTdsn(""))

	if s.shapeRootGroup != nil && len(s.shapeRootGroup.Children) > 0 {
		outer.Children = append(outer.Children, makeTdmn("ADBE Root Vectors Group"))
		rootGroupTdgp, err := lowerVectorGroup(s.shapeRootGroup, ctx)
		if err != nil {
			return nil, err
		}
		outer.Children = append(outer.Children, rootGroupTdgp)
	}

	transformWrapper, err := lowerLayerTransform(s.shapeTransform, ctx)
	if err != nil {
		return nil, err
	}
	// transformWrapper.Children[0] = tdmn("ADBE Transform Group");
	// transformWrapper.Children[1..] = the tdgp body chunks. Append both:
	// the tdmn becomes the property-group name marker, the second is the
	// body LIST(tdgp).
	outer.Children = append(outer.Children, transformWrapper.Children[0])
	transformBody := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: transformWrapper.Children[1:]}
	outer.Children = append(outer.Children, transformBody)

	outer.Children = append(outer.Children, makeTdmn("ADBE Group End"))
	layr.Children = append(layr.Children, outer)

	return layr, nil
}

// buildLdtaBytes returns the 160-byte canonical ldta payload for a
// ShapeLayer. Fills the well-known offsets via ldta_layout.go constants;
// everything else is zero (AE-friendly default per RE-S1).
//
// Layer subtype byte (@0x80) = 4 (Shape) per ldta_layout.go ldtaLayerSubtype
// comment. Quality (@0x04) = 2 (Best) — AE's typical default. Visible bit
// is at byte 0x27 bit0 (per parse_layer.go decoder); default Visible = true
// → 0x01.
func buildLdtaBytes(s *ShapeLayer, ctx *lowerCtx) []byte {
	d := make([]byte, ldtaSize2020)

	// @0x00 — layer-local ID.
	binary.BigEndian.PutUint32(d[ldtaLayerID:ldtaLayerID+4], s.ID)

	// @0x04 — Quality. 2 = Best (AE default for new layers).
	binary.BigEndian.PutUint16(d[ldtaQuality:ldtaQuality+2], 2)

	// @0x08 — StretchDividend. 100 (= 100% stretch).
	binary.BigEndian.PutUint32(d[ldtaStretchDivd:ldtaStretchDivd+4], 100)
	// @0x6C — StretchDivisor. 100 (paired with dividend for 1× speed).
	binary.BigEndian.PutUint32(d[ldtaStretchDivs:ldtaStretchDivs+4], 100)

	// In/Out points: span the full source / comp default. AE writes
	// dividend=0, divisor=1 for fresh layers; out-point dividend = sentinel
	// (-1 cast to u32). Keep both divisors = 1 to avoid div-by-zero in
	// parsers.
	binary.BigEndian.PutUint32(d[ldtaInPointDivd:ldtaInPointDivd+4], 0)
	binary.BigEndian.PutUint32(d[ldtaInPointDivs:ldtaInPointDivs+4], 1)
	binary.BigEndian.PutUint32(d[ldtaOutPointDivd:ldtaOutPointDivd+4], 0xFFFFFFFF)
	binary.BigEndian.PutUint32(d[ldtaOutPointDivs:ldtaOutPointDivs+4], 1)

	// StartTime: 0/1.
	binary.BigEndian.PutUint32(d[ldtaStartTimeDivd:ldtaStartTimeDivd+4], 0)
	binary.BigEndian.PutUint32(d[ldtaStartTimeDivs:ldtaStartTimeDivs+4], 1)

	// Attr bytes — visible bit at byte 0x27 bit0 (parse_layer.go decoder).
	// Default Visible = true for fresh ShapeLayer.
	d[ldtaAttrByte2] = 0x01

	// @0x40 — legacy 32-byte name slot. Mirror up to 31 bytes of name +
	// NUL terminator. Parser ignores this in favor of the Utf8 chunk, but
	// AE's saved fixtures populate it.
	nameBytes := []byte(s.Name)
	maxName := 31
	if len(nameBytes) < maxName {
		maxName = len(nameBytes)
	}
	copy(d[ldtaLegacyName:ldtaLegacyName+maxName], nameBytes[:maxName])

	// @0x80 — LayerSubtype = Shape (4).
	binary.BigEndian.PutUint32(d[ldtaLayerSubtype:ldtaLayerSubtype+4], 4)

	// @0x84 — ParentID. 0 = no parent.
	binary.BigEndian.PutUint32(d[ldtaParentID:ldtaParentID+4], s.ParentID)

	_ = ctx // reserved for capability-driven branches (V3)
	return d
}

// lowerLayerTransform emits the Layer Transform Group LIST(tdgp). V2.2
// emits user-facing 2D form (5 streams: Anchor / Position / Scale /
// Rotate Z / Opacity) — matches V1 parser convention. The 6-axis schema
// observed in RE-S2 is AE's internal 3D-compatible form; runtime users
// drive 2D, lowering writes 2D.
//
// Structure per RE-S1: LIST(tdgp) holding tdsb + tdsn + N × (tdmn +
// LIST(tdbs, sub-stream)) + tdmn(Group End). Same shape as Vector Group
// — handled via direct child append (the typed lowering funcs return
// LIST(tdgp) wrappers with tdmn as child[0]; we inline).
func lowerLayerTransform(t *LayerTransform, ctx *lowerCtx) (*rifx.Chunk, error) {
	tdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	tdgp.Children = append(tdgp.Children, makeTdsb(), makeTdsn("Transform"))

	type stream2D struct {
		ps        *PropertyStream[[2]float64]
		matchName string
		display   string
	}
	type stream1D struct {
		ps        *PropertyStream[float64]
		matchName string
		display   string
	}
	streams2D := []stream2D{
		{t.anchorPoint, MatchNameAnchorPoint, "Anchor Point"},
		{t.position, MatchNamePosition, "Position"},
		{t.scale, MatchNameScale, "Scale"},
	}
	streams1D := []stream1D{
		{t.rotation, MatchNameRotateZ, "Rotation"},
		{t.opacity, MatchNameOpacity, "Opacity"},
	}
	for _, s := range streams2D {
		c, err := LowerVec2Stream(s.ps, s.matchName, s.display, ctx)
		if err != nil {
			return nil, err
		}
		tdgp.Children = append(tdgp.Children, c.Children...)
	}
	for _, s := range streams1D {
		c, err := LowerFloat64Stream(s.ps, s.matchName, s.display, ctx)
		if err != nil {
			return nil, err
		}
		tdgp.Children = append(tdgp.Children, c.Children...)
	}
	tdgp.Children = append(tdgp.Children, makeTdmn("ADBE Group End"))

	// Top-level tdmn naming the Transform Group lives in the parent Layr
	// directly above this tdgp — handled by lowerShapeLayer via child
	// ordering. We attach the tdmn here so the caller can pair-emit.
	// Wrap: return a LIST(tdgp) whose first child is tdmn(name).
	wrapper := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	wrapper.Children = append(wrapper.Children, makeTdmn("ADBE Transform Group"))
	wrapper.Children = append(wrapper.Children, tdgp.Children...)
	return wrapper, nil
}
