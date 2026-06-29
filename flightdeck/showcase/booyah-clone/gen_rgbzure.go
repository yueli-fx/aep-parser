// flightdeck/showcase/booyah-clone/gen_rgbzure.go — comp ⑧ "RGBズレ" (RGB shift).
// Three precomp layers, all referencing comp ② テキスト, each tinted by a Fill effect
// to one RGB channel and staggered in time → chromatic-aberration glitch on the text:
//   L0 "B": Fill colour [A,R,G,B] = [255,0,131,255] (blue),  Opacity 23kf, start 0.300
//   L1 "R": Fill DEFAULT colour (red; Fill-0002 elided),     Opacity 24kf, start 0.067
//   L2 "G": Fill colour [255,0,255,86]            (green),   Opacity 25kf, start -0.134
// All three are centred (Position static 960,540). New dimension vs ⑦: per-layer Fill
// colour (ADBE Fill), reusing the SetEffectParam colour path. Values via the oracle.
//
// Colour encoding is [A,R,G,B] each 0-255 — the same form the parser reads, so the
// oracle's value is passed through verbatim. L1 keeps the Fill default (red) by NOT
// setting Fill-0002, mirroring the original's elision; the AE DOM verify confirms red.
//
// RGB JITTER: the original drives the chromatic-aberration *motion* with an expression,
// not a static offset — each layer's separated X-position runs wiggle(24,12) (Y static).
// We keep Position UNIFIED and reproduce the X-only jitter with [wiggle(24,12)[0], cy, 0]
// (render-equivalent; separated-dims is a documented mechanism delta). The original ALSO
// wiggles Opacity, but expr-on-keyframed-opacity drops the layer in AE, so that flicker
// is an omitted gap (see finishRgbzureExpr). Applied after a second reopen.
//
// Three-phase: buildRgbzure creates the comp + three precomp layers (needs comp ② to
// exist); finishRgbzure applies transforms + Fill after reopen 1; finishRgbzureExpr
// attaches the wiggle expressions after reopen 2 (SetExpression needs the Position/
// Opacity channels that SetLayerTransform materialized to be re-parsed).
package main

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const rgbzureCompName = "RGBズレ"

func buildRgbzure(p *aep.Project, orc *oracle) {
	comp, err := aep.NewComposition(p, rgbzureCompName, 1920, 1080, cloneFps, 6)
	must(err)
	child := p.CompositionByName(komakoCompName) // comp ② テキスト, built earlier in main
	if child == nil {
		panic("comp ⑧: child comp ② テキスト not found in project")
	}
	for range 3 {
		_, err := aep.NewPrecompLayer(comp, child, child.Name)
		must(err)
	}
	_ = orc
}

func finishRgbzure(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(rgbzureCompName)
	if comp == nil {
		panic("comp ⑧: not found after reopen")
	}
	origComp := orc.mustComp(rgbzureCompName)
	cx, cy := float64(comp.Width)/2, float64(comp.Height)/2

	for i, l := range comp.Layers {
		ol := origComp.Layers[i]
		otg := findGroup(ol.PropertyTree(), "ADBE Transform Group")
		tr := aep.NewLayerTransform()
		must(tr.AnchorPoint().SetStaticValue([2]float64{0.5, 0.5}))

		// Position: static comp-centre (original stores 960,540 on all three).
		if pos := findProp(otg, "ADBE Position"); pos != nil && len(pos.Keyframes) == 0 {
			v := toFloats(pos.StaticValue)
			must(tr.Position().SetStaticValue([2]float64{v[0], v[1]}))
		} else {
			must(tr.Position().SetStaticValue([2]float64{cx, cy}))
		}
		// Opacity keyframes (fraction → percent ×100). Times are copied verbatim
		// (the RGB fade sits at comp-time ~2.5-3.5s).
		if op := findProp(otg, "ADBE Opacity"); op != nil {
			for _, kf := range op.Keyframes {
				must(tr.Opacity().AddKeyframeLinear(kf.Time, toScalar(kf.Value)*100))
			}
		}

		must(l.SetStartTime(ol.StartTime))
		must(aep.SetLayerTransform(comp.Layers[i], tr))

		// Fill effect: tint each copy to one RGB channel. Copy the colour from the
		// original where present; where elided (L1 "R") leave the Fill default (red).
		fx, err := aep.AddEffect(comp.Layers[i], "ADBE Fill")
		must(err)
		if of := ol.Effects[0]; of != nil {
			if col := fxParam(of, "ADBE Fill-0002"); col != nil {
				// SetStaticValue wants []float64 of len == Components (4 for ARGB colour).
				c := append([]float64(nil), toFloats(col.StaticValue)...)
				_, err := aep.SetEffectParam(comp.Layers[i], fx, "ADBE Fill-0002", c)
				must(err)
			}
		}
	}

	fmt.Printf("  ⑧ RGBズレ: 3 precomp layers → comp ② + Fill (B=[0,131,255] / R=default red / G=[0,255] α86), staggered + Opacity fade\n")
}

// finishRgbzureExpr runs after the second reopen: attach the X-position wiggle that
// drives the RGB horizontal jitter. The original separates Position and runs
// wiggle(24,12) on X only (Y static); we keep Position unified and reproduce X-only
// jitter as [wiggle(24,12)[0], cy, 0] (the three copies are centred, so Y=cy, Z=0).
//
// Two AE quirks pinned by ship-gate bisection (see the ⑧ row in INDEX.md):
//  1. value FREEZES wiggle. An array expression that passes other components through
//     `value` ([wiggle(24,12)[0], value[1], value[2]]) makes AE evaluate the whole
//     thing to the base value — X stays 960, no jitter (valueAtTime confirmed it static
//     at every sample). Using LITERAL pass-through components (cy, 0) lets the wiggle
//     evaluate. A bare wiggle(24,12) also works but jitters Y too (not faithful).
//  2. Opacity wiggle OMITTED. The original also runs wiggle(27,33) on Opacity, but here
//     Opacity is keyframed (the fade) and SetExpression on a SetLayerTransform-
//     materialized KEYFRAMED property makes AE silently drop the whole layer (verified:
//     op-only variant → comp layers=0). Static-opacity expressions are fine (⑤ L2). So
//     the opacity flicker-wiggle is a documented gap; the fade keyframes are kept.
func finishRgbzureExpr(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(rgbzureCompName)
	if comp == nil {
		panic("comp ⑧: not found after reopen 2")
	}
	origComp := orc.mustComp(rgbzureCompName)
	cy := float64(comp.Height) / 2
	for i, l := range comp.Layers {
		otg := findGroup(origComp.Layers[i].PropertyTree(), "ADBE Transform Group")
		pos0 := findProp(otg, "ADBE Position_0")
		if pos0 == nil || pos0.Expression == "" {
			panic(fmt.Sprintf("comp ⑧ L%d: original Position_0 expression missing", i))
		}
		pos := l.Position()
		if pos == nil {
			panic(fmt.Sprintf("comp ⑧ L%d: Position not materialized (SetLayerTransform should have)", i))
		}
		must(pos.SetExpression(fmt.Sprintf("[%s[0], %g, 0]", pos0.Expression, cy)))
	}
	fmt.Printf("  ⑧ RGBズレ: + Position X-wiggle (opacity flicker-wiggle omitted: expr-on-keyframed-opacity drops layer)\n")
}
