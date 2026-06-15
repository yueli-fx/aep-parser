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

	"github.com/example/aep-parser/internal/rifx"
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

	// Build + swap the new om-s; snapshot for rollback on re-parse failure.
	oldOmS := kids[shapeIdx]
	newOmS := makeMaskShapeOmS(scaled)
	mb.atomTdgp.Children[shapeIdx] = newOmS

	// Refresh the scene-side path fields + the shph back-ref from the new shap.
	shap := firstShap(newOmS)
	if shap == nil {
		mb.atomTdgp.Children[shapeIdx] = oldOmS
		return fmt.Errorf("SetMaskPath: rebuilt om-s has no shap")
	}
	refreshMaskShap(mask, mb, shap)
	mask.PathKeyframes = nil // SetMaskPath writes a static path
	return nil
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
