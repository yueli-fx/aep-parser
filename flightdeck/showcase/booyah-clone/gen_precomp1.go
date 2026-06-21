// flightdeck/showcase/booyah-clone/gen_precomp1.go — comp ⑤ "プリコンポジション 1".
// Three precomp layers, all referencing comp ① シェイイイイプ (NewPrecompLayer — the
// first nesting in this replication). The three copies stagger comp ①'s glitch: L0
// slides horizontally (Position 2kf 778→1204), L1 flickers (Opacity 13kf), L2 sits
// static. Small per-layer start offsets stagger the nested timeline. Values pulled
// from the original via the oracle.
//
// Centering: L0's Position is animated with ABSOLUTE coords (no centering needed),
// but L1 and L2 store Position SEPARATED (Position_0/Position_1, read as 0,0) while
// being visually centred — the same separated-Position read trap as comp ④
// ([[shape-layer-position-default-offscreen]] Case 2). So L1/L2 are centred
// explicitly at comp-centre. Native 30 fps.
//
// Two-phase: buildPrecomp1 creates the comp + the three precomp layers (needs comp ①
// to already exist in the project — it's built earlier in main); finishPrecomp1
// applies SetLayerTransform after the reopen.
package main

import (
	"fmt"

	aep "github.com/example/aep-parser/internal/aep"
)

const precomp1CompName = "プリコンポジション 1"

func buildPrecomp1(p *aep.Project, orc *oracle) {
	comp, err := aep.NewComposition(p, precomp1CompName, 1920, 1080, 30, 6)
	must(err)
	child := p.CompositionByName(shapeIiipCompName) // comp ①, built earlier in main
	if child == nil {
		panic("comp ⑤: child comp ① not found in project")
	}
	// 3 precomp layers; NewPrecompLayer appends at the bottom, so L0/L1/L2 keep order.
	// Start offsets are applied post-reopen via SetStartTime (precomp layers are
	// embed-template clones, so the pre-reopen scene StartTime field is not honored).
	for range 3 {
		_, err := aep.NewPrecompLayer(comp, child, child.Name)
		must(err)
	}
	_ = orc
}

// finishPrecomp1 runs after the reopen: layer transforms via SetLayerTransform
// (needs parsed layers). Layers are addressed by index (all three share comp ①'s
// name, so LayerByName would be ambiguous).
func finishPrecomp1(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(precomp1CompName)
	if comp == nil {
		panic("comp ⑤: not found after reopen")
	}
	origComp := orc.mustComp(precomp1CompName)
	cx, cy := float64(comp.Width)/2, float64(comp.Height)/2

	// Start offsets stagger each nested comp ① copy's glitch (SetStartTime = ldta edit).
	for i, l := range comp.Layers {
		must(l.SetStartTime(origComp.Layers[i].StartTime))
	}

	// L0: Position animated 2kf (absolute coords — no centering).
	otg0 := findGroup(origComp.Layers[0].PropertyTree(), "ADBE Transform Group")
	tr0 := aep.NewLayerTransform()
	for _, kf := range findProp(otg0, "ADBE Position").Keyframes {
		v := toFloats(kf.Value)
		must(tr0.Position().AddKeyframeLinear(kf.Time, [2]float64{v[0], v[1]}))
	}
	must(aep.SetLayerTransform(comp.Layers[0], tr0))

	// L1: centred Position (separated-read trap) + Opacity 13kf flicker.
	otg1 := findGroup(origComp.Layers[1].PropertyTree(), "ADBE Transform Group")
	tr1 := aep.NewLayerTransform()
	must(tr1.Position().SetStaticValue([2]float64{cx, cy}))
	for _, kf := range findProp(otg1, "ADBE Opacity").Keyframes {
		must(tr1.Opacity().AddKeyframeLinear(kf.Time, toScalar(kf.Value)*100))
	}
	must(aep.SetLayerTransform(comp.Layers[1], tr1))

	// L2: centred Position, static (no animation).
	tr2 := aep.NewLayerTransform()
	must(tr2.Position().SetStaticValue([2]float64{cx, cy}))
	must(aep.SetLayerTransform(comp.Layers[2], tr2))

	fmt.Printf("  ⑤ プリコンポジション 1: 3 precomp layers → comp ① (L0 Position 2kf, L1 Opacity 13kf+centred, L2 static centred)\n")
}
