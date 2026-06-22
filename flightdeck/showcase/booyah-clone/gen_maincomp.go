// flightdeck/showcase/booyah-clone/gen_maincomp.go — comp ⑫ "メインコンプ！"
// (Phase 4 Task 4.2): the TOP composition that stacks the three branches and is the
// target of the terminal render-pixel comparison against the original.
//   L0 ⑩ グリッチテキスト, blend=Add, Position 3kf (small H-nudge) + Opacity wiggle(29,25)
//      + Venetian Blinds(14/90/9)
//   L1 ⑨ なんか周りのやつ, Opacity wiggle(29,09)
//   L2 ⑪ 背景変えるならココ, Exposure2(-1.45) + CurvesCustom (default instance —
//      the curve data is arbitrary-data with no scripting channel, spike 0.5 blocked)
//
// All three layers are precomp references (NewPrecompLayer → the clone comps of the
// same name). Opacity wiggles are on STATIC opacity (the ⑤ L2 safe pattern — AE keeps
// the expression; unlike ⑪'s keyframed-Exposure wiggle which drops the layer).
// Effects reuse the ⑩/⑪ helper (Venetian Blinds / Exposure2 / CurvesCustom default).
//
// Three-phase: buildMainComp (comp + 3 precomp layers); finishMainComp (transforms +
// blend + effects, post-reopen 1); finishMainCompExpr (Opacity wiggle, post-reopen 2,
// once SetLayerTransform has materialized the Opacity channel).
//
// Delta: L0 Position keyframes are written LINEAR (SetLayerTransform has no eased
// builder); the original carries a 0.333-influence ease on a 75px nudge — a subtle
// timing delta on a small move, logged.
package main

import (
	"fmt"

	aep "github.com/example/aep-parser/internal/aep"
)

const mainCompName = "メインコンプ！"

func buildMainComp(p *aep.Project, orc *oracle) {
	comp, err := aep.NewComposition(p, mainCompName, 1920, 1080, cloneFps, 6)
	must(err)
	orig := orc.mustComp(mainCompName)
	for i, ol := range orig.Layers {
		src := ol.SourceComposition()
		if src == nil {
			panic(fmt.Sprintf("comp ⑫ L%d: expected a precomp source", i))
		}
		child := p.CompositionByName(src.Name)
		if child == nil {
			panic(fmt.Sprintf("comp ⑫ L%d: child comp %q not found in clone", i, src.Name))
		}
		_, err := aep.NewPrecompLayer(comp, child, layerName(ol, i))
		must(err)
	}
	fmt.Printf("  ⑫ メインコンプ: %d precomp layers built (⑩/⑨/⑪)\n", len(orig.Layers))
}

func finishMainComp(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(mainCompName)
	if comp == nil {
		panic("comp ⑫: not found after reopen")
	}
	orig := orc.mustComp(mainCompName)
	cx, cy := float64(comp.Width)/2, float64(comp.Height)/2

	for i, l := range comp.Layers {
		ol := orig.Layers[i]
		otg := findGroup(ol.PropertyTree(), "ADBE Transform Group")
		tr := aep.NewLayerTransform()
		// Precomp/AV anchor = source-centre FRACTION (setlayertransform-av-anchor-fraction).
		must(tr.AnchorPoint().SetStaticValue([2]float64{0.5, 0.5}))

		// Position: keyframed (L0 3kf H-nudge) or static comp-centre. Linear (no eased
		// LayerTransform builder); copy the keyframe times/values from the original.
		if pos := findProp(otg, "ADBE Position"); pos != nil && len(pos.Keyframes) > 0 {
			for _, kf := range pos.Keyframes {
				v := toFloats(kf.Value)
				must(tr.Position().AddKeyframeLinear(kf.Time, [2]float64{v[0], v[1]}))
			}
		} else {
			must(tr.Position().SetStaticValue([2]float64{cx, cy}))
		}
		// Opacity: static 100 % (the wiggle expression is attached in reopen 2).
		must(tr.Opacity().SetStaticValue(100))
		must(aep.SetLayerTransform(l, tr))

		must(l.SetBlendingMode(ol.BlendingMode)) // L0 = Add; L1/L2 Normal
		applyEffectsFromOriginal(l, ol)          // L0 Venetian Blinds, L2 Exposure2 + CurvesCustom
	}
	fmt.Printf("  ⑫ メインコンプ: + transforms/blend/effects (L0 Add+Position 3kf+VenetianBlinds, L2 Exposure2 -1.45 + Curves default)\n")
}

// finishMainCompExpr runs after reopen 2: the static-Opacity wiggle expressions
// (L0 wiggle(29,25), L1 wiggle(29,09)). Static opacity → AE keeps the expression.
func finishMainCompExpr(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(mainCompName)
	if comp == nil {
		panic("comp ⑫: not found after reopen 2")
	}
	orig := orc.mustComp(mainCompName)
	for i, l := range comp.Layers {
		otg := findGroup(orig.Layers[i].PropertyTree(), "ADBE Transform Group")
		op := findProp(otg, "ADBE Opacity")
		if op == nil || op.Expression == "" {
			continue
		}
		clOp := l.Opacity()
		if clOp == nil {
			panic(fmt.Sprintf("comp ⑫ L%d: Opacity not materialized for wiggle", i))
		}
		must(clOp.SetExpression(op.Expression))
	}
	fmt.Printf("  ⑫ メインコンプ: + Opacity wiggle (L0 wiggle(29,25), L1 wiggle(29,09))\n")
}
