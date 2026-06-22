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
// Two-phase: buildRgbzure creates the comp + three precomp layers (needs comp ② to
// exist); finishRgbzure applies transforms + Fill after the reopen.
package main

import (
	"fmt"

	aep "github.com/example/aep-parser/internal/aep"
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
