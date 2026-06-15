// SetMaskPath — rewrite an existing mask's path geometry in place.
//
// AddMask can only CREATE a mask; reshaping a parsed mask's outline was
// previously refused (the path lives in a variable-length om-s/shap/kfl
// subtree). SetMaskPath rebuilds the "ADBE Mask Shape" om-s with the new path
// (reusing makeMaskShapeOmS — same encodeBezier + mask-strictness lhd3/shph
// patching AddMask uses, so vertex counts other than 4 work) and swaps it into
// the mask atom group. WriteAEP recomputes the enclosing LIST sizes (the om-s
// grows/shrinks with the vertex count, exactly like AddMask's structural splice).
//
// Coordinates are layer pixels (scaled to the source-fraction unit via
// maskLayerDims, identical to AddMask). The mask must have been round-tripped
// through the parser (it needs its atom-group chunk back-ref).
package serializer

import (
	"fmt"

	"github.com/example/aep-parser/internal/codec"
	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// SetMaskPath replaces mask's static path with path (layer-pixel coordinates).
// (Full contract lives on the aep.SetMaskPath facade — docgen source.)
func SetMaskPath(layer *Layer, mask *Mask, path BezierPath) error {
	if layer == nil || mask == nil {
		return fmt.Errorf("SetMaskPath: nil layer or mask")
	}
	if len(path.Vertices) < 2 {
		return fmt.Errorf("SetMaskPath: path needs >= 2 vertices, got %d", len(path.Vertices))
	}
	mb := maskBack(mask)
	if mb == nil || mb.atomTdgp == nil {
		return fmt.Errorf("SetMaskPath: mask %q has no atom-group chunk (round-trip through Reopen first)", mask.Name)
	}

	lw, lh := maskLayerDims(layer)
	if lw <= 0 || lh <= 0 {
		return fmt.Errorf("SetMaskPath: cannot resolve layer %q pixel dimensions", layer.Name)
	}
	scaled := scaleMaskPath(path, lw, lh)

	// Locate the existing "ADBE Mask Shape" om-s in the atom group.
	kids := mb.atomTdgp.Children
	shapeIdx := -1
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == "ADBE Mask Shape" &&
			kids[i+1].IsList() && kids[i+1].FormType == rifx.IDOmS {
			shapeIdx = i + 1
			break
		}
	}
	if shapeIdx < 0 {
		return fmt.Errorf("SetMaskPath: mask %q has no Mask Shape om-s", mask.Name)
	}

	// Atomic structural mutation (CLAUDE.md #1): swap the rebuilt om-s in, then
	// re-parse the atom to validate it before keeping the change — roll back to
	// the pre-call om-s / shph back-ref / warnings on any failure.
	oldOmS := kids[shapeIdx]
	oldShph := mb.shph
	oldWarn := warningsLen(layer)
	newOmS := makeMaskShapeOmS(scaled)
	mb.atomTdgp.Children[shapeIdx] = newOmS
	rollback := func() {
		mb.atomTdgp.Children[shapeIdx] = oldOmS
		mb.shph = oldShph
		rollbackWarnings(layer, oldWarn)
	}

	shap := firstShap(newOmS)
	if shap == nil {
		rollback()
		return fmt.Errorf("SetMaskPath: rebuilt om-s has no shap")
	}

	// Re-parse the modified atom (mirrors AddMask): confirm it decodes to one
	// mask with the requested vertex count and produces no parser warnings.
	comp := scene.LayerComp(layer)
	if comp == nil || scene.CompositionProj(comp) == nil {
		rollback()
		return fmt.Errorf("SetMaskPath: layer %q has no composition/project back-ref", layer.Name)
	}
	ctx := newParseCtxFPS(comp.TickRate, comp.FrameRate, comp.Name, &scene.CompositionProj(comp).Warnings)
	tmpLayr := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
		makeTdmn(MatchNameGroupMaskParade),
		{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
			makeTdmn("ADBE Mask Atom"), makeMaskMkif(mask.Index), mb.atomTdgp,
		}},
	}}
	probes := parseMasks(tmpLayr, ctx)
	if len(probes) != 1 || len(probes[0].Vertices) != len(path.Vertices) {
		rollback()
		return fmt.Errorf("SetMaskPath: reshaped mask re-parse failed (%d masks, %d vertices, want 1/%d)",
			len(probes), probeVerts(probes), len(path.Vertices))
	}
	if newWarn := newWarningsSince(layer, oldWarn); len(newWarn) > 0 {
		rollback()
		return fmt.Errorf("SetMaskPath: produced %d parser warning(s), rolled back: %v", len(newWarn), newWarn)
	}

	// Refresh the scene-side path fields + the shph back-ref from the new shap.
	refreshMaskShap(mask, mb, shap)
	mask.PathKeyframes = nil // SetMaskPath writes a static path
	return nil
}

// SetMaskPathKeyframes replaces mask's path with an ANIMATED outline: N
// keyframes (>= 2), each a BezierPath snapshot at a time in seconds, with
// optional temporal ease. Layer-pixel coordinates (scaled via maskLayerDims,
// identical to SetMaskPath). Vertex counts may differ between keyframes.
// (Full contract lives on the aep.SetMaskPathKeyframes facade — docgen source.)
func SetMaskPathKeyframes(layer *Layer, mask *Mask, keys []MaskPathKey) error {
	if layer == nil || mask == nil {
		return fmt.Errorf("SetMaskPathKeyframes: nil layer or mask")
	}
	if len(keys) < 2 {
		return fmt.Errorf("SetMaskPathKeyframes: need >= 2 keyframes, got %d", len(keys))
	}
	for i, k := range keys {
		if len(k.Path.Vertices) < 2 {
			return fmt.Errorf("SetMaskPathKeyframes: keyframe %d path needs >= 2 vertices, got %d", i, len(k.Path.Vertices))
		}
	}
	mb := maskBack(mask)
	if mb == nil || mb.atomTdgp == nil {
		return fmt.Errorf("SetMaskPathKeyframes: mask %q has no atom-group chunk (round-trip through Reopen first)", mask.Name)
	}

	lw, lh := maskLayerDims(layer)
	if lw <= 0 || lh <= 0 {
		return fmt.Errorf("SetMaskPathKeyframes: cannot resolve layer %q pixel dimensions", layer.Name)
	}
	comp := scene.LayerComp(layer)
	if comp == nil || scene.CompositionProj(comp) == nil {
		return fmt.Errorf("SetMaskPathKeyframes: layer %q has no composition/project back-ref", layer.Name)
	}

	// Scale each keyframe path into the on-disk fraction unit and pack into the
	// codec keyframe shape the time-table encoder consumes.
	scaledKfs := make([]codec.StreamKeyframe[BezierPath], len(keys))
	for i, k := range keys {
		scaledKfs[i] = codec.StreamKeyframe[BezierPath]{
			Time:    k.Time,
			Value:   scaleMaskPath(k.Path, lw, lh),
			InEase:  k.InEase,
			OutEase: k.OutEase,
		}
	}

	// Locate the existing "ADBE Mask Shape" om-s in the atom group.
	kids := mb.atomTdgp.Children
	shapeIdx := -1
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == "ADBE Mask Shape" &&
			kids[i+1].IsList() && kids[i+1].FormType == rifx.IDOmS {
			shapeIdx = i + 1
			break
		}
	}
	if shapeIdx < 0 {
		return fmt.Errorf("SetMaskPathKeyframes: mask %q has no Mask Shape om-s", mask.Name)
	}

	// Atomic structural mutation (CLAUDE.md #1): swap the animated om-s in, then
	// re-parse the atom to validate before keeping the change — roll back to the
	// pre-call om-s / shph back-ref / warnings on any failure.
	oldOmS := kids[shapeIdx]
	oldShph := mb.shph
	oldWarn := warningsLen(layer)
	newOmS := makeMaskShapeOmSAnimated(scaledKfs, comp.TickRate)
	mb.atomTdgp.Children[shapeIdx] = newOmS
	rollback := func() {
		mb.atomTdgp.Children[shapeIdx] = oldOmS
		mb.shph = oldShph
		rollbackWarnings(layer, oldWarn)
	}

	ctx := newParseCtxFPS(comp.TickRate, comp.FrameRate, comp.Name, &scene.CompositionProj(comp).Warnings)
	tmpLayr := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
		makeTdmn(MatchNameGroupMaskParade),
		{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
			makeTdmn("ADBE Mask Atom"), makeMaskMkif(mask.Index), mb.atomTdgp,
		}},
	}}
	probes := parseMasks(tmpLayr, ctx)
	if len(probes) != 1 || len(probes[0].PathKeyframes) != len(keys) {
		rollback()
		return fmt.Errorf("SetMaskPathKeyframes: reshaped mask re-parse failed (%d masks, %d keyframes, want 1/%d)",
			len(probes), probeKfs(probes), len(keys))
	}
	for i := range keys {
		if got, want := len(probes[0].PathKeyframes[i].Vertices), len(keys[i].Path.Vertices); got != want {
			rollback()
			return fmt.Errorf("SetMaskPathKeyframes: keyframe %d re-parsed %d vertices, want %d", i, got, want)
		}
	}
	if newWarn := newWarningsSince(layer, oldWarn); len(newWarn) > 0 {
		rollback()
		return fmt.Errorf("SetMaskPathKeyframes: produced %d parser warning(s), rolled back: %v", len(newWarn), newWarn)
	}

	// Refresh scene-side fields from the re-parsed animated mask.
	mask.PathKeyframes = probes[0].PathKeyframes
	mask.Vertices = probes[0].Vertices
	mask.ShphRaw = probes[0].ShphRaw
	mask.Closed = probes[0].Closed
	if shap := firstShap(newOmS); shap != nil {
		for _, ch := range shap.Children {
			if ch.ID == rifx.IDShph {
				mb.shph = ch
			}
		}
	}
	return nil
}

// makeMaskShapeOmSAnimated builds an ANIMATED "ADBE Mask Shape" om-s: the value
// tdbs becomes a TIME-table tdbs (tdsb + tdsn + the mask tdb4 with its
// static→animated flags patched + a LIST(kfl) time table — one 64B block per
// keyframe), and omks holds one mask-strictness shap per keyframe. Byte-
// isomorphic to AE's own animated shape path (re_path_anim.aep) — the static
// mask tdb4 (maskShapeTdb4) differs from that fixture's animated tdb4 only at
// the same three flag offsets @0x05/@0x44/@0x4f injectAnimatedStream patches
// (byte-verified), and the per-frame geometry reuses makeMaskShap's mask
// patches. See incidents/path-keyframe-write-re.md + add-mask-create-re.md.
func makeMaskShapeOmSAnimated(kfs []codec.StreamKeyframe[BezierPath], tickRate float64) *rifx.Chunk {
	tdb4 := append([]byte(nil), maskShapeTdb4...)
	if len(tdb4) > 0x4f {
		tdb4[0x05] &^= 0x01
		tdb4[0x44] = 0x01
		tdb4[0x4f] &^= 0x01
	}
	timeTdbs := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdbs, Children: []*rifx.Chunk{
		makeTdsb(),
		makeTdsn(aeDefaultGroupName),
		{ID: rifx.IDtdb4, Data: tdb4},
		encodePathTimeTable(kfs, &lowerCtx{tickRate: tickRate}),
	}}
	shaps := make([]*rifx.Chunk, len(kfs))
	for i, kf := range kfs {
		shaps[i] = makeMaskShap(kf.Value)
	}
	omks := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDOmks, Children: shaps}
	return &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDOmS, Children: []*rifx.Chunk{timeTdbs, omks}}
}

// probeKfs returns the keyframe count of the first probe mask, or -1.
func probeKfs(masks []*Mask) int {
	if len(masks) == 0 {
		return -1
	}
	return len(masks[0].PathKeyframes)
}

// probeVerts returns the vertex count of the first probe mask, or -1.
func probeVerts(masks []*Mask) int {
	if len(masks) == 0 {
		return -1
	}
	return len(masks[0].Vertices)
}

// scaleMaskPath divides every coordinate by the layer-fraction divisors (mirrors
// AddMask's scaling).
func scaleMaskPath(path BezierPath, lw, lh float64) BezierPath {
	out := BezierPath{Closed: path.Closed,
		Vertices:    make([][2]float64, len(path.Vertices)),
		InTangents:  make([][2]float64, len(path.InTangents)),
		OutTangents: make([][2]float64, len(path.OutTangents)),
	}
	for i, v := range path.Vertices {
		out.Vertices[i] = [2]float64{v[0] / lw, v[1] / lh}
	}
	for i, v := range path.InTangents {
		out.InTangents[i] = [2]float64{v[0] / lw, v[1] / lh}
	}
	for i, v := range path.OutTangents {
		out.OutTangents[i] = [2]float64{v[0] / lw, v[1] / lh}
	}
	return out
}

// firstShap returns the first LIST(shap) inside an om-s (via its omks).
func firstShap(omS *rifx.Chunk) *rifx.Chunk {
	for _, ch := range omS.Children {
		if ch.IsList() && ch.FormType == rifx.IDOmks {
			for _, s := range ch.Children {
				if s.IsList() && s.FormType == rifx.IDShap {
					return s
				}
			}
		}
	}
	return nil
}

// refreshMaskShap re-syncs Mask.Vertices / Closed / ShphRaw and the shph
// back-ref from a (new) shap — the static-path subset of fillFromShap that also
// re-points the back-ref (fillFromShap only sets it when nil).
func refreshMaskShap(mask *Mask, mb *maskBackrefs, shap *rifx.Chunk) {
	for _, ch := range shap.Children {
		switch {
		case ch.ID == rifx.IDShph:
			mb.shph = ch
			mask.ShphRaw = append([]byte(nil), ch.Data...)
			if len(ch.Data) >= 4 {
				mask.Closed = ch.Data[3]&0x08 == 0
			}
		case ch.IsList() && ch.FormType == rifx.IDkfl:
			mask.Vertices = decodeMaskVertices(ch)
		}
	}
}
