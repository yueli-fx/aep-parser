// flightdeck/showcase/booyah-clone/gen_glitch_text.go — comp ⑩ "グリッチテキスト".
// The MONSTER comp: 27 layers, ~100 masks, 11 Displacement Maps — the work/risk
// centre of the whole replication (Phase 3, plan Task 3.1). Built in stages:
//   STEP 1 (this file's first commit): no-mask SKELETON — 27 layers with correct
//     source refs, blend modes, timing (start/in/out) and visibility. Anchors the
//     layer structure and tests that 27 layers survive AE (multi-layer-silent-drop).
//   STEP 2 (later): batch masks (copyMasksFromOriginal) on L0–L20.
//   STEP 3 (later): effect chains (Geometry2 / Glo2 / Displacement Map / Fractal
//     Noise / Gaussian Blur), incl. the Displacement-Map layer references.
//
// Source map — driven by reading the original comp's layers via the oracle:
//   precomp layers      → the clone comp of the same name (⑦ ここは開けない / ②
//                         テキスト / ⑧ RGBズレ / ③ マップ用ノイズ), built earlier
//   footage-backed adj  → own NewAdjustmentLayer (footage-share = own solid, a
//                         render-neutral structural delta, same call as ③)
//   footage-backed av   → own NewSolidLayer (L24/L26; Fractal Noise paints pixels)
//
// Two-phase: buildGlitchText creates the comp + 27 layers (sources only; needs the
// child comps ⑦②⑧③ to exist already); finishGlitchText applies blend/timing/
// visibility after the reopen (those ldta edits are unreliable on un-Reopened
// embed-template clones — same discipline as ⑤⑥⑦).
package main

import (
	"encoding/binary"
	"fmt"
	"math"

	aep "github.com/example/aep-parser/internal/aep"
)

const glitchTextCompName = "グリッチテキスト"

func buildGlitchText(p *aep.Project, orc *oracle) {
	comp, err := aep.NewComposition(p, glitchTextCompName, 1920, 1080, cloneFps, 6)
	must(err)
	orig := orc.mustComp(glitchTextCompName)

	for i, ol := range orig.Layers {
		switch {
		case ol.SourceComposition() != nil:
			// Precomp layer → the clone comp of the same name (built earlier in main).
			src := ol.SourceComposition()
			child := p.CompositionByName(src.Name)
			if child == nil {
				panic(fmt.Sprintf("comp ⑩ L%d: child comp %q not found in clone", i, src.Name))
			}
			_, err := aep.NewPrecompLayer(comp, child, layerName(ol, i))
			must(err)
		case ol.IsAdjust:
			// Footage-backed adjustment layer → own adjustment (own solid source).
			_, err := aep.NewAdjustmentLayer(comp, layerName(ol, i))
			must(err)
		default:
			// Footage-backed AV solid (L24/L26) → own black solid; Fractal Noise paints it.
			_, err := aep.NewSolidLayer(comp, layerName(ol, i), 1920, 1080, [3]float64{0, 0, 0})
			must(err)
		}
	}
	fmt.Printf("  ⑩ グリッチテキスト: skeleton %d layers built (sources only; masks+effects deferred)\n", len(orig.Layers))
}

// layerName returns the original layer's display name, or a synthetic stable name
// when the original is empty (AE needs a non-empty solid/adjustment source name; the
// glitch comp's adjustment layers are all unnamed in the original).
func layerName(ol *aep.Layer, i int) string {
	if ol.Name != "" {
		return ol.Name
	}
	return fmt.Sprintf("L%02d", i)
}

func finishGlitchText(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(glitchTextCompName)
	if comp == nil {
		panic("comp ⑩: not found after reopen")
	}
	orig := orc.mustComp(glitchTextCompName)
	if len(comp.Layers) != len(orig.Layers) {
		panic(fmt.Sprintf("comp ⑩: layer count %d != original %d (silent drop?)", len(comp.Layers), len(orig.Layers)))
	}

	cx, cy := float64(comp.Width)/2, float64(comp.Height)/2
	for i, l := range comp.Layers {
		ol := orig.Layers[i]
		// Timing: source-offset start (@0x0C) + in/out trim. The short adjustment
		// spans (e.g. L0 [1.2→1.3]) are the staccato glitch slices — set in then out
		// so the duration recompute lands on the trimmed span.
		must(l.SetStartTime(ol.StartTime))
		must(l.SetInPoint(ol.InPoint()))
		must(l.SetOutPoint(ol.OutPoint()))
		// Blend mode (L7 = Add; every other layer Normal).
		must(l.SetBlendingMode(ol.BlendingMode))
		// Visibility (L24/L25/L26 are hidden — video switch off — in the original).
		must(l.SetVisible(ol.Visible))

		// Non-default layer opacity. The original dims several layers — L23 横ブラー
		// (heavily-blurred text) at 9 %, L8 glow at 60 % — and leaving them at 100 %
		// washes the whole comp into a cyan glow that buries the text. SetLayerTransform
		// rewrites the entire transform group, so reproduce the centred transform
		// (anchor = source-centre fraction, position = comp centre — all these layers
		// sit at [960,540]) alongside the opacity. Skip default-100 %, hidden, and
		// keyframed/expression opacity (none of the dimmed layers are animated).
		if !ol.Visible {
			continue
		}
		op := findProp(findGroup(ol.PropertyTree(), "ADBE Transform Group"), "ADBE Opacity")
		if op == nil || len(op.Keyframes) > 0 || op.Expression != "" || op.StaticValue == nil {
			continue
		}
		pct := toScalar(op.StaticValue) * 100 // parser opacity is 0..1; SetLayerTransform wants percent
		if pct >= 99.999 {
			continue // 100 % = default, no-op
		}
		tr := aep.NewLayerTransform()
		must(tr.AnchorPoint().SetStaticValue([2]float64{0.5, 0.5}))
		must(tr.Position().SetStaticValue([2]float64{cx, cy}))
		must(tr.Opacity().SetStaticValue(pct))
		must(aep.SetLayerTransform(l, tr))
	}
	fmt.Printf("  ⑩ グリッチテキスト: + blend/timing/visibility/opacity on %d layers\n", len(comp.Layers))
}

// finishGlitchTextMasks (STEP 2) rebuilds every original mask on the matching
// clone layer. All 131 masks in ⑩ are axis-aligned rectangles (audited: 4
// vertices, Add mode, unit-square ldat) — the tearing slices that gate the
// displacement/scale glitch. AE stores a mask outline as a bbox (shph @0x04 =
// L,T,R,B fractions of the SOURCE pixel space) plus a unit-square ldat; the
// parser's Mask.Vertices surfaces only the normalized ldat (all masks read
// identically), so the real per-mask geometry must be reconstructed from the
// shph bbox. We map each bbox to a pixel rectangle and feed it to AddMask, which
// re-divides by the (1920×1080) source dims back to the original fractions —
// a faithful round-trip for these rectangular masks.
func finishGlitchTextMasks(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(glitchTextCompName)
	if comp == nil {
		panic("comp ⑩: not found after reopen (masks)")
	}
	orig := orc.mustComp(glitchTextCompName)
	w, h := float64(comp.Width), float64(comp.Height) // source = own 1920×1080 solid
	total := 0
	for i, l := range comp.Layers {
		total += copyMasksFromOriginal(l, orig.Layers[i], w, h)
	}
	fmt.Printf("  ⑩ グリッチテキスト: + %d masks (bbox-rect tearing slices) across L0–L20\n", total)
}

// Curated per-effect tuned-param sets (match-name suffixes). Only real settable
// params are listed — group/marker leaves (-0000, Sub Settings, Evolution Options,
// etc.) and the Displacement-Map layer-reference (-0001, a self-ref auto-bound by
// AddEffect) are deliberately omitted so SetEffectParam never clobbers a group.
// Params present in the set but ELIDED in the original are skipped (leave default).
var effectScalarParams = map[string][]string{
	"ADBE Geometry2":        {"-0003"},                                                                // Scale Height
	"ADBE Glo2":             {"-0002", "-0003", "-0004"},                                              // Threshold, Radius, Intensity
	"ADBE Gaussian Blur 2":  {"-0001", "-0002"},                                                       // Blurriness, Blur Dimensions
	"ADBE Displacement Map": {"-0003", "-0005", "-0007"},                                              // Max H, Max V, Edge Behavior
	"ADBE Fractal Noise":    {"-0002", "-0004", "-0005", "-0009", "-0010", "-0011", "-0012", "-0015"}, // NoiseType, Contrast, Brightness, UniformScaling, Scale, ScaleW, ScaleH, Complexity
	"ADBE Exposure2":        {"-0003"},                                                                // Exposure (⑪ L1, 9kf; wiggle expr added separately)
	"ADBE Venetian Blinds":  {"-0001", "-0002", "-0003"},                                              // Completion, Direction, Width
	"ADBE Ramp":             {"-0005"},                                                                // Ramp Shape (1 = linear, 2 = radial)
}

// effectVecParams: multi-component (2D point / 4D color) params per effect, static
// or animated — SetEffectParam (static) / AnimateEffectParamVec (keyframed).
var effectVecParams = map[string][]string{
	"ADBE Fractal Noise": {"-0013"},                             // Offset Turbulence (2D pan)
	"ADBE Ramp":          {"-0001", "-0002", "-0003", "-0004"},  // Start of Ramp (2D anim), Start Color (4D), End of Ramp (2D), End Color (4D)
}

// unsupportedEffects are effect match-names absent from the embedded template set
// (AddEffect would refuse them). applyEffectsFromOriginal skips + logs these so the
// layer still builds; the missing effect is an honest, logged fidelity gap.
var unsupportedEffects = map[string]bool{
	"ADBE Noise2": true, // ⑪ L0 grain "微妙なノイズ" — not in the embedded library
}

// effectExprParams: expression-driven params per effect (set value 0 + expr).
var effectExprParams = map[string][]string{
	"ADBE Fractal Noise": {"-0023"}, // Evolution = time*N
}

// ok panics on a non-nil error and returns the value — for (*Property, error)
// builder calls whose Property we don't need.
func ok[T any](v T, err error) T { must(err); return v }

// finishGlitchTextFx (STEP 3) rebuilds every layer's effect chain: Geometry2
// Scale Height (L0–L6) / Glo2 (L7×2, L8) / Displacement Map (L10–L20; the layer
// reference is a SELF-ref in the original — tdpi==own layer id — which AddEffect
// binds to the host by default, so no SetEffectLayerParam is needed) / Gaussian
// Blur (L23, L24) / Fractal Noise with Evolution expression (L24, L26). All via
// the same SetEffectParam / AnimateEffectParam patterns proven on comps ③ and ⑦.
func finishGlitchTextFx(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(glitchTextCompName)
	if comp == nil {
		panic("comp ⑩: not found after reopen (fx)")
	}
	orig := orc.mustComp(glitchTextCompName)
	nfx := 0
	for i, l := range comp.Layers {
		nfx += applyEffectsFromOriginal(l, orig.Layers[i])
	}
	fmt.Printf("  ⑩ グリッチテキスト: + %d effects (Geometry2/Glo2/DisplacementMap[self-ref]/GaussianBlur/FractalNoise)\n", nfx)
}

// applyEffectsFromOriginal mirrors origLayer's effect parade onto newLayer using
// the curated param sets, and returns the effect count added.
func applyEffectsFromOriginal(newLayer, origLayer *aep.Layer) int {
	added := 0
	for _, of := range origLayer.Effects {
		if unsupportedEffects[of.MatchName] {
			fmt.Printf("    ⚠ skip unsupported effect %s on %q (not in embedded set)\n", of.MatchName, newLayer.Name)
			continue
		}
		fx, err := aep.AddEffect(newLayer, of.MatchName)
		must(err)
		added++
		mn := of.MatchName

		for _, suf := range effectScalarParams[mn] {
			op := fxParam(of, mn+suf)
			if op == nil {
				continue // elided in original → leave AE default
			}
			if len(op.Keyframes) > 0 {
				ok(aep.AnimateEffectParam(newLayer, fx, mn+suf, scalarKfs(op.Keyframes)))
			} else if op.StaticValue != nil {
				ok(aep.SetEffectParam(newLayer, fx, mn+suf, toScalar(op.StaticValue)))
			}
		}
		for _, suf := range effectVecParams[mn] {
			op := fxParam(of, mn+suf)
			if op == nil {
				continue
			}
			if len(op.Keyframes) > 0 {
				ok(aep.AnimateEffectParamVec(newLayer, fx, mn+suf, vecKfs(op.Keyframes)))
			} else if op.StaticValue != nil {
				ok(aep.SetEffectParam(newLayer, fx, mn+suf, toFloats(op.StaticValue)))
			}
		}
		for _, suf := range effectExprParams[mn] {
			op := fxParam(of, mn+suf)
			if op == nil || op.Expression == "" {
				continue
			}
			pr := ok(aep.SetEffectParam(newLayer, fx, mn+suf, 0.0))
			must(pr.SetExpression(op.Expression))
			must(pr.SetExpressionEnabled(true))
		}
	}
	return added
}

// scalarKfs converts parsed 1D keyframes to ScalarKeyframe inputs, copying the
// per-side temporal ease (zero = linear; L19 displacement carries real ease).
func scalarKfs(kfs []*aep.Keyframe) []aep.ScalarKeyframe {
	out := make([]aep.ScalarKeyframe, 0, len(kfs))
	for _, kf := range kfs {
		sk := aep.ScalarKeyframe{Time: kf.Time, Value: toScalar(kf.Value)}
		if len(kf.InTemporalEase) > 0 {
			sk.InEase = kf.InTemporalEase[0]
		}
		if len(kf.OutTemporalEase) > 0 {
			sk.OutEase = kf.OutTemporalEase[0]
		}
		out = append(out, sk)
	}
	return out
}

// vecKfs converts parsed multi-component keyframes to VectorKeyframe inputs
// (linear — the Offset Turbulence pan is a straight 2-kf slide, as in comp ③).
func vecKfs(kfs []*aep.Keyframe) []aep.VectorKeyframe {
	out := make([]aep.VectorKeyframe, 0, len(kfs))
	for _, kf := range kfs {
		out = append(out, aep.VectorKeyframe{Time: kf.Time, Value: toFloats(kf.Value)})
	}
	return out
}

// copyMasksFromOriginal rebuilds origLayer's masks on newLayer as bbox-rect
// pixel paths and returns the count added. w/h are the layer source pixel dims.
func copyMasksFromOriginal(newLayer, origLayer *aep.Layer, w, h float64) int {
	n := 0
	for _, om := range origLayer.Masks {
		rect, ok := maskBBoxRect(om, w, h)
		if !ok {
			panic(fmt.Sprintf("comp ⑩: mask %q has no decodable shph bbox", om.Name))
		}
		nm, err := aep.AddMask(newLayer, om.Name, rect)
		must(err)
		if om.Mode != aep.MaskModeAdd {
			must(nm.SetMode(om.Mode))
		}
		if om.Inverted {
			must(nm.SetInverted(true))
		}
		// Render-affecting mask attributes (omitting these rendered ⑪'s flare as a
		// hard bright rectangle — its mask has feather [1377,1377]). Feather/expansion
		// in pixels, opacity 0..1, all copied verbatim from the parsed original.
		if om.Feather[0] != 0 || om.Feather[1] != 0 {
			must(nm.SetFeather(om.Feather))
		}
		if om.Opacity != 1.0 {
			must(nm.SetOpacity(om.Opacity))
		}
		if om.Expansion != 0 {
			must(nm.SetExpansion(om.Expansion))
		}
		n++
	}
	return n
}

// maskBBoxRect decodes a mask's shph bounding box (@0x04 = four big-endian
// float32 L,T,R,B fractions of the source pixel space) into a closed pixel
// rectangle (top-left → top-right → bottom-right → bottom-left). Returns false
// when the shph is too short to carry a bbox.
func maskBBoxRect(m *aep.Mask, w, h float64) (aep.BezierPath, bool) {
	d := m.ShphRaw
	if len(d) < 0x14 {
		return aep.BezierPath{}, false
	}
	f := func(off int) float64 {
		return float64(math.Float32frombits(binary.BigEndian.Uint32(d[off : off+4])))
	}
	l, t, r, b := f(0x04), f(0x08), f(0x0C), f(0x10)
	x0, y0, x1, y1 := l*w, t*h, r*w, b*h
	return aep.BezierPath{
		Vertices: [][2]float64{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}},
		Closed:   true,
	}, true
}
