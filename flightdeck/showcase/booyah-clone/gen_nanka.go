// flightdeck/showcase/booyah-clone/gen_nanka.go — comp ⑨ "なんか周りのやつ".
// First MIXED-SOURCE comp: two precomp layers (src=④ カクッ) under a freshly-built
// shape layer with an ANIMATED Trim Paths reveal. Layout:
//   L0 precomp ④, centred, Opacity wiggle(29,38), horizontally mirrored
//   L1 precomp ④, centred, Opacity wiggle(29,38)
//   L2 shape "シェイプレイヤー 1": Rect 1856×1015 + animated Trim (Start 100→0,
//      Offset 0→720, bezier-eased 2kf) + Stroke; centred, Opacity 65% wiggle(29,09)
//
// New dimensions vs ⑧: (a) one comp drawing from two different source kinds (precomp +
// from-scratch shape); (b) animated shape Trim Paths with COPIED bezier ease — the
// first non-linear keyframe ease in this replication (Position/Opacity were all linear;
// only the text animators and now this trim carry real ease). Trim Start/Offset use
// PropertyStream.AddKeyframeWithEase (the eased sibling of AddKeyframeLinear), and the
// shape-scalar lower flips the static cdat to an animated keyframe container.
//
// L0 mirror: the original L0 is a 3D layer with Rotate Y = 180° (a horizontal flip).
// We reproduce the flip with Scale X = -100 % on a 2-D layer (render-equivalent for a
// flat layer; avoids the 3D-enable + RotateY channel-synthesis path) — a documented
// mechanism delta. Opacity wiggles are on STATIC opacity, so SetExpression is safe
// (unlike ⑧'s keyframed opacity which AE drops — see expression-enable-byte-pair).
//
// Three-phase: buildNanka (comp + 2 precomp + shape w/ animated trim, pre-reopen);
// finishNanka (SetLayerTransform after reopen 1); finishNankaExpr (opacity wiggle
// expressions after reopen 2).
package main

import (
	"fmt"

	aep "github.com/example/aep-parser/internal/aep"
)

const nankaCompName = "なんか周りのやつ"
const nankaShapeName = "シェイプレイヤー 1"

func buildNanka(p *aep.Project, orc *oracle) {
	comp, err := aep.NewComposition(p, nankaCompName, 1920, 1080, cloneFps, 6)
	must(err)
	child := p.CompositionByName(kakuhCompName) // comp ④ カクッ, built earlier
	if child == nil {
		panic("comp ⑨: child comp ④ カクッ not found in project")
	}
	// L0, L1 = precomp layers → ④ (NewPrecompLayer appends at bottom: L0 then L1).
	for range 2 {
		_, err := aep.NewPrecompLayer(comp, child, child.Name)
		must(err)
	}

	// L2 = shape layer with Rect + animated Trim + Stroke (order matches the original:
	// Rect defines the path, Trim trims it, Stroke renders the trimmed path = draw-on).
	sl, err := aep.NewShapeLayer(comp, nankaShapeName)
	must(err)
	root := sl.RootGroup()
	rect, err := root.AddRect()
	must(err)
	must(rect.SetSize([2]float64{1856, 1015.4375}))

	trim, err := root.AddTrim()
	must(err)
	origComp := orc.mustComp(nankaCompName)
	otrim := findGroup(findGroup(origComp.Layers[2].PropertyTree(), "ADBE Root Vectors Group"), "ADBE Vector Filter - Trim")
	for _, kf := range findProp(otrim, "ADBE Vector Trim Start").Keyframes {
		must(trim.Start().AddKeyframeWithEase(kf.Time, toScalar(kf.Value), kf.InTemporalEase[0], kf.OutTemporalEase[0]))
	}
	for _, kf := range findProp(otrim, "ADBE Vector Trim Offset").Keyframes {
		must(trim.Offset().AddKeyframeWithEase(kf.Time, toScalar(kf.Value), kf.InTemporalEase[0], kf.OutTemporalEase[0]))
	}

	_, err = root.AddStroke() // width elided in the original = AddStroke default
	must(err)
}

func finishNanka(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(nankaCompName)
	if comp == nil {
		panic("comp ⑨: not found after reopen")
	}
	cx, cy := float64(comp.Width)/2, float64(comp.Height)/2

	// L0/L1 precomp: centred (separated-position 0,0 read trap), anchor = source-centre
	// FRACTION (AV-layer anchor units, [[setlayertransform-av-anchor-fraction]]). Opacity
	// static (the wiggle is added in finishNankaExpr). L0 mirrored via Scale X = -100%.
	for i := 0; i < 2; i++ {
		tr := aep.NewLayerTransform()
		must(tr.AnchorPoint().SetStaticValue([2]float64{0.5, 0.5}))
		must(tr.Position().SetStaticValue([2]float64{cx, cy}))
		if i == 0 {
			must(tr.Scale().SetStaticValue([2]float64{-100, 100})) // Rotate Y 180° → mirror
		}
		must(aep.SetLayerTransform(comp.Layers[i], tr))
	}

	// L2 shape: centred (shape-layer Position default 0,0 → comp-centre puts the rect
	// centred), Opacity static 65% (wiggle added later).
	trS := aep.NewLayerTransform()
	must(trS.Position().SetStaticValue([2]float64{cx, cy}))
	must(trS.Opacity().SetStaticValue(65))
	must(aep.SetLayerTransform(comp.LayerByName(nankaShapeName), trS))

	_ = orc
	fmt.Printf("  ⑨ なんか周りのやつ: 2 precomp(④) + 1 shape層 (Rect+animated Trim+Stroke), L0 mirrored, centred\n")
}

// finishNankaExpr runs after the second reopen: opacity wiggle expressions (all on
// STATIC opacity, so AE-safe). Sources from the oracle verbatim.
func finishNankaExpr(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(nankaCompName)
	if comp == nil {
		panic("comp ⑨: not found after reopen 2")
	}
	origComp := orc.mustComp(nankaCompName)
	for i, l := range comp.Layers {
		otg := findGroup(origComp.Layers[i].PropertyTree(), "ADBE Transform Group")
		opOrig := findProp(otg, "ADBE Opacity")
		if opOrig == nil || opOrig.Expression == "" {
			continue
		}
		op := l.Opacity()
		if op == nil {
			panic(fmt.Sprintf("comp ⑨ L%d: Opacity not materialized", i))
		}
		must(op.SetExpression(opOrig.Expression))
	}
	fmt.Printf("  ⑨ なんか周りのやつ: + Opacity wiggle expr (L0/L1 wiggle(29,38), L2 wiggle(29,09))\n")
}
