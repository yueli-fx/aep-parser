// flightdeck/showcase/booyah-clone/gen_hiraku.go — comp ⑦ "ここは開けない方が身のため".
// Three precomp layers, all referencing comp ⑥ シェイプの塊 (reusing the NewPrecompLayer
// nesting). New dimension vs ⑤/⑥: a Glow effect (ADBE Glo2) on the two visible layers.
//   L0: static Position + Scale 101 %, Opacity 25kf (fades to 0), Glow Intensity 0.30
//   L1: static Position + Scale 101 %, Opacity 34kf,                Glow Intensity 0.35
//   L2: invisible (Opacity static 0), no Glow — replicated for structural fidelity.
// All Glow Radius = 0 (non-default; AE default is 25). Position here is STATIC absolute
// coords (not keyframed like ⑥, not the separated 0,0 trap of ⑤). Values via the oracle.
//
// Same transform discipline as ⑥: replicate every channel (static Scale included), anchor
// as a source-centre FRACTION for the precomp/AV layer (setlayertransform-av-anchor-fraction).
//
// Two-phase: buildHiraku creates the comp + three precomp layers (needs comp ⑥ to exist);
// finishHiraku applies transforms + Glow after the reopen (AddEffect needs a parsed layer).
package main

import (
	"fmt"

	aep "github.com/example/aep-parser/internal/aep"
)

const hirakuCompName = "ここは開けない方が身のため"

func buildHiraku(p *aep.Project, orc *oracle) {
	comp, err := aep.NewComposition(p, hirakuCompName, 1920, 1080, cloneFps, 6)
	must(err)
	child := p.CompositionByName(shapeKatamariCompName) // comp ⑥, built earlier in main
	if child == nil {
		panic("comp ⑦: child comp ⑥ シェイプの塊 not found in project")
	}
	for range 3 {
		_, err := aep.NewPrecompLayer(comp, child, child.Name)
		must(err)
	}
	_ = orc
}

func finishHiraku(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(hirakuCompName)
	if comp == nil {
		panic("comp ⑦: not found after reopen")
	}
	origComp := orc.mustComp(hirakuCompName)
	cx, cy := float64(comp.Width)/2, float64(comp.Height)/2

	for i, l := range comp.Layers {
		ol := origComp.Layers[i]
		otg := findGroup(ol.PropertyTree(), "ADBE Transform Group")
		tr := aep.NewLayerTransform()

		// Anchor: elided in original → source-centre FRACTION for a precomp/AV layer.
		must(tr.AnchorPoint().SetStaticValue([2]float64{0.5, 0.5}))

		// Static Scale (fraction → percent ×100); L0/L1 = 101 %, L2 elides → default 100 %.
		if sc := findProp(otg, "ADBE Scale"); sc != nil {
			s := toFloats(sc.StaticValue)
			must(tr.Scale().SetStaticValue([2]float64{s[0] * 100, s[1] * 100}))
		}
		// Static Rotate Z — all elide here (= 0), but copy if present for safety.
		if r := findProp(otg, "ADBE Rotate Z"); r != nil {
			must(tr.Rotation().SetStaticValue(toScalar(r.StaticValue)))
		}
		// Position: static absolute coords (L0/L1); elided (L2) → comp centre.
		if pos := findProp(otg, "ADBE Position"); pos != nil {
			if len(pos.Keyframes) > 0 {
				for _, kf := range pos.Keyframes {
					v := toFloats(kf.Value)
					must(tr.Position().AddKeyframeLinear(kf.Time, [2]float64{v[0], v[1]}))
				}
			} else {
				v := toFloats(pos.StaticValue)
				must(tr.Position().SetStaticValue([2]float64{v[0], v[1]}))
			}
		} else {
			must(tr.Position().SetStaticValue([2]float64{cx, cy}))
		}
		// Opacity: keyframed (L0 25kf / L1 34kf) or static (L2 = 0, invisible). ×100 → percent.
		if op := findProp(otg, "ADBE Opacity"); op != nil {
			if len(op.Keyframes) > 0 {
				for _, kf := range op.Keyframes {
					must(tr.Opacity().AddKeyframeLinear(kf.Time, toScalar(kf.Value)*100))
				}
			} else {
				must(tr.Opacity().SetStaticValue(toScalar(op.StaticValue) * 100))
			}
		}

		must(l.SetStartTime(ol.StartTime))
		must(aep.SetLayerTransform(comp.Layers[i], tr))

		// Glow (ADBE Glo2) on the layers that carry one (L0/L1). Copy Radius (-0003) and
		// Intensity (-0004) from the original; both stored only because non-default.
		if len(ol.Effects) > 0 {
			of := ol.Effects[0]
			fx, err := aep.AddEffect(comp.Layers[i], "ADBE Glo2")
			must(err)
			for _, mn := range []string{"ADBE Glo2-0003", "ADBE Glo2-0004"} {
				if op := fxParam(of, mn); op != nil {
					_, err := aep.SetEffectParam(comp.Layers[i], fx, mn, toScalar(op.StaticValue))
					must(err)
				}
			}
		}
	}

	fmt.Printf("  ⑦ ここは開けない方が身のため: 3 precomp layers → comp ⑥ (L0/L1 static Position+Scale 101%%+Opacity kf+Glow, L2 invisible)\n")
}
