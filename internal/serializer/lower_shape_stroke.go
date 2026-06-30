// Code moved from lower_shape_node.go; keep behavior-only edits out of split commits.
package serializer

import (
	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

func lowerStrokeNode(s *StrokeNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	// Dashes enabled → swap to the dashed template (carries Dash 1 / Gap 1
	// slots); else the solid template (Dashes group is an empty placeholder).
	clone := cloneShapeStrokeBody
	if s.Dashes() != nil && s.Dashes().Enabled() {
		clone = cloneShapeStrokeDashedBody
	}
	body, err := clone()
	if err != nil {
		return nil, err
	}
	if s.Color().Mode() == codec.StreamModeAnimated && s.Color().HasKeyframes() {
		if err := injectAnimatedColor(body, "ADBE Vector Stroke Color", s.Color().Keyframes(), ctx); err != nil {
			return nil, err
		}
	} else {
		cv, _ := s.Color().StaticValue()
		overwriteShapeStreamCdat(body, "ADBE Vector Stroke Color", encodeShapeColorBE(cv))
	}

	// Opacity (raw %) + Width (raw px) — 1D non-spatial scalars (bpk-48,
	// value@0x08, no normalization; RE'd from v2_2_stroke_kf_re.aep). The stroke //nolint:jargon
	// body template already carries both cdat slots, so animated streams flip in
	// place (previously collapsed to the first keyframe value).
	if err := lowerShapeScalar(body, "ADBE Vector Stroke Opacity", s.Opacity(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Stroke Width", s.Width(), ctx); err != nil {
		return nil, err
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Stroke Line Cap", encodeF64sBE(float64(s.LineCap())))
	overwriteShapeStreamCdat(body, "ADBE Vector Stroke Line Join", encodeF64sBE(float64(s.LineJoin())))
	overwriteShapeStreamCdat(body, "ADBE Vector Stroke Miter Limit", encodeF64sBE(s.MiterLimit()))
	overwriteShapeStreamCdat(body, "ADBE Vector Blend Mode", encodeF64sBE(float64(s.BlendMode())))
	overwriteShapeStreamCdat(body, "ADBE Vector Composite Order", encodeF64sBE(float64(s.CompositeOrder())))
	lowerStrokeTaper(body, s.Taper())
	lowerStrokeWave(body, s.Wave())
	lowerStrokeDashes(body, s.Dashes())
	return body, nil
}

// lowerStrokeDashes overwrites the Dashes group's Dash 1 / Gap 1 cdats inside
// the embedded dashed-stroke body. No-op when dashes are disabled (the solid
// template was cloned and has no Dash/Gap slots to overwrite). Offset is not
// modeled (AE keeps it hidden / script-ungettable — no template slot exists).
func lowerStrokeDashes(strokeBody *rifx.Chunk, d *StrokeDashes) {
	if d == nil || !d.Enabled() {
		return
	}
	g := findGroupBody(strokeBody, "ADBE Vector Stroke Dashes")
	if g == nil {
		return
	}
	overwriteShapeStreamCdat(g, "ADBE Vector Stroke Dash 1", encodeF64sBE(d.Dash()))
	overwriteShapeStreamCdat(g, "ADBE Vector Stroke Gap 1", encodeF64sBE(d.Gap()))
}

// lowerStrokeTaper overwrites the Taper group's %-mode scalar cdats inside the
// embedded stroke body. The body template (re-extracted from a fixture with the
// Taper group's % controls set non-default) carries the 6 always-active slots:
// Start/End Length, Start/End Width, Start/End Ease — each a float64 BE at
// cdat[0:8]. Length Units / StartWidthPx / EndWidthPx are AE-elided in % mode
// and absent from the template (not currently modeled).
func lowerStrokeTaper(strokeBody *rifx.Chunk, t *StrokeTaper) {
	g := findGroupBody(strokeBody, "ADBE Vector Stroke Taper")
	if g == nil || t == nil {
		return
	}
	overwriteShapeStreamCdat(g, "ADBE Vector Taper Start Length", encodeF64sBE(t.StartLength()))
	overwriteShapeStreamCdat(g, "ADBE Vector Taper End Length", encodeF64sBE(t.EndLength()))
	overwriteShapeStreamCdat(g, "ADBE Vector Taper Start Width", encodeF64sBE(t.StartWidth()))
	overwriteShapeStreamCdat(g, "ADBE Vector Taper End Width", encodeF64sBE(t.EndWidth()))
	overwriteShapeStreamCdat(g, "ADBE Vector Taper Start Ease", encodeF64sBE(t.StartEase()))
	overwriteShapeStreamCdat(g, "ADBE Vector Taper End Ease", encodeF64sBE(t.EndEase()))
}

// lowerStrokeWave overwrites the Wave group's Wavelength-mode scalar cdats
// (Amount / Wavelength / Phase). Units / Cycles are AE-elided in Wavelength mode
// and absent from the template (not currently modeled).
func lowerStrokeWave(strokeBody *rifx.Chunk, w *StrokeWave) {
	g := findGroupBody(strokeBody, "ADBE Vector Stroke Wave")
	if g == nil || w == nil {
		return
	}
	overwriteShapeStreamCdat(g, "ADBE Vector Taper Wave Amount", encodeF64sBE(w.Amount()))
	overwriteShapeStreamCdat(g, "ADBE Vector Taper Wavelength", encodeF64sBE(w.Wavelength()))
	overwriteShapeStreamCdat(g, "ADBE Vector Taper Wave Phase", encodeF64sBE(w.Phase()))
}

// findGroupBody returns the LIST(tdgp) group body following the tdmn matching
// groupName among body.Children (one level), or nil. Used to descend into a
// nested shape group (Stroke Taper / Wave) before overwriting its sub-stream
// cdats with overwriteShapeStreamCdat.
func findGroupBody(body *rifx.Chunk, groupName string) *rifx.Chunk {
	kids := body.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == groupName {
			if next := kids[i+1]; next.IsList() && next.FormType == rifx.IDTdgp {
				return next
			}
		}
	}
	return nil
}

// lowerVectorGroup wraps shape-node children into the Root Vectors Group's
// inner LIST(tdgp). The on-disk shape (every AE-saved fixture observed —
// tolerance.aep + re_shapes.aep) is a 5-level nesting:
//
//	[LIST tdgp]                      ← Root Vectors Group body (this return)
//	  tdsb(0x00000401) + tdsn
//	  tdmn("ADBE Vector Group")
//	  [LIST tdgp]                    ← Vector Group body (3-child fixed routing)
//	    tdsb(0x00000001) + tdsn
//	    tdmn("ADBE Vectors Group")
//	    [LIST tdgp]                  ← Vectors Group body (holds user shapes)
//	      tdsb(0x00000401) + tdsn
//	      N × (tdmn(<shape>) + LIST(tdgp, shape body))
//	      tdmn("ADBE Group End")
//	    tdmn("ADBE Vector Transform Group") + empty LIST(tdgp)
//	    tdmn("ADBE Vector Materials Group") + empty LIST(tdgp)
//	    tdmn("ADBE Group End")
//	  tdmn("ADBE Group End")
//
// Flattening everything into the Root Vectors Group body directly made AE 2025
// parse the file without exception but silently drop the layer from
// comp.layers. The wrappers are structural: AE Shape Layer's Contents always
// holds one or more "ADBE Vector Group" entries (each is what UI shows as
// "Group N"), and each Vector Group always carries the 3-child fixed routing
// (Vectors Group for shape kids + Transform + Materials).
//
// This implementation maps the runtime `shapeRootGroup.Children = [Rect,
// Fill, ...]` to a SINGLE Vector Group wrapper (semantic = AE's
// auto-created "Group 1"). A future version may expose multiple
// user-named groups.
//
// Children render in order: Children[0] = bottom, Children[len-1] = top.
