// showcase/booyah-clone/gen_shape_katamari.go — comp ⑥ "シェイプの塊".
// Seven precomp layers, all referencing comp ⑤ プリコンポジション 1 (NewPrecompLayer,
// reusing ⑤'s nesting approach). Each copy slides horizontally (Position 2kf, absolute
// coords 541→1340) and flickers (Opacity 34/41/39/39/39/44 kf; L6 has none), shrunk by a
// per-layer static Scale (L0 24 %, L1-L5 33 %, L6 75 %) and some rotated 90° (L0/L3/L4).
// Per-layer start offsets stagger the seven nested timelines. Values pulled from the
// original via the oracle.
//
// Unlike comp ⑤, Position here is stored as ABSOLUTE coords with real keyframes (no
// separated-read 0,0 trap), so no explicit centering is needed for the animated layers;
// only L6 (Position elided → AE default = comp centre) is centred explicitly.
//
// Transform replication copies EVERY channel — static Scale + Rotation included, not just
// the animated Position/Opacity — because dropping the static channels is what spread comp
// ⑤'s nested copies too wide (incidents/layer-replication-drops-static-transform-channels.md).
// Anchor is written as a FRACTION-of-source (0.5 = centre) for precomp/AV layers, NOT pixels
// (incidents/setlayertransform-av-anchor-fraction.md).
//
// Two-phase: buildShapeKatamari creates the comp + the seven precomp layers (needs comp ⑤
// to already exist — built earlier in main); finishShapeKatamari applies the transforms
// after the reopen.
package main

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const shapeKatamariCompName = "シェイプの塊"

func buildShapeKatamari(p *aep.Project, orc *oracle) {
	comp, err := aep.NewComposition(p, shapeKatamariCompName, 1920, 1080, cloneFps, 6)
	must(err)
	child := p.CompositionByName(precomp1CompName) // comp ⑤, built earlier in main
	if child == nil {
		panic("comp ⑥: child comp ⑤ プリコンポジション 1 not found in project")
	}
	// 7 precomp layers; NewPrecompLayer appends at the bottom, so L0..L6 keep order.
	for range 7 {
		_, err := aep.NewPrecompLayer(comp, child, child.Name)
		must(err)
	}
	_ = orc
}

func finishShapeKatamari(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(shapeKatamariCompName)
	if comp == nil {
		panic("comp ⑥: not found after reopen")
	}
	origComp := orc.mustComp(shapeKatamariCompName)
	cx, cy := float64(comp.Width)/2, float64(comp.Height)/2

	for i, l := range comp.Layers {
		otg := findGroup(origComp.Layers[i].PropertyTree(), "ADBE Transform Group")
		tr := aep.NewLayerTransform()

		// Anchor: original elides it (AE default = source centre); for a precomp/AV
		// layer SetLayerTransform wants the anchor as a FRACTION (0.5 = centre), not px.
		must(tr.AnchorPoint().SetStaticValue([2]float64{0.5, 0.5}))

		// Static Scale (parser reads a fraction; SetLayerTransform.Scale wants percent → ×100).
		if sc := findProp(otg, "ADBE Scale"); sc != nil {
			s := toFloats(sc.StaticValue)
			must(tr.Scale().SetStaticValue([2]float64{s[0] * 100, s[1] * 100}))
		}
		// Static Rotate Z (degrees) — only L0/L3/L4 carry it; the rest elide (= 0).
		if r := findProp(otg, "ADBE Rotate Z"); r != nil {
			must(tr.Rotation().SetStaticValue(toScalar(r.StaticValue)))
		}
		// Position: L0-L5 animate with absolute-coord keyframes; L6 elides → comp centre.
		if pos := findProp(otg, "ADBE Position"); pos != nil && len(pos.Keyframes) > 0 {
			for _, kf := range pos.Keyframes {
				v := toFloats(kf.Value)
				must(tr.Position().AddKeyframeLinear(kf.Time, [2]float64{v[0], v[1]}))
			}
		} else {
			must(tr.Position().SetStaticValue([2]float64{cx, cy}))
		}
		// Opacity: L0-L5 flicker (kf, fraction → percent ×100); L6 elides → leave default 100 %.
		if op := findProp(otg, "ADBE Opacity"); op != nil && len(op.Keyframes) > 0 {
			for _, kf := range op.Keyframes {
				must(tr.Opacity().AddKeyframeLinear(kf.Time, toScalar(kf.Value)*100))
			}
		}

		// Start offset staggers each nested ⑤ copy (SetStartTime = ldta edit, post-reopen).
		must(l.SetStartTime(origComp.Layers[i].StartTime))
		must(aep.SetLayerTransform(comp.Layers[i], tr))
	}

	fmt.Printf("  ⑥ シェイプの塊: 7 precomp layers → comp ⑤ (Position 2kf + Opacity 34/41/39/39/39/44 kf, static Scale/Rotation, staggered starts)\n")
}
