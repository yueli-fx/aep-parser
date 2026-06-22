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
// explicitly at comp-centre. 29.97 fps (cloneFps).
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
	comp, err := aep.NewComposition(p, precomp1CompName, 1920, 1080, cloneFps, 6)
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

	// Anchor for a precomp layer = the SOURCE's centre. CAUTION: SetLayerTransform
	// encodes the anchor of an AV / precomp layer as a FRACTION-of-source (0.5 = centre),
	// NOT pixels like a shape/text layer (comp ②/④). Writing pixels (960,540) makes AE
	// read 960×1920 / 540×1080 ≈ (1.84M, 0.58M) — the pivot lands millions of px away and
	// the content renders off-screen (blank). The original elides anchor (AE default =
	// source centre); NewLayerTransform inits anchor to (0,0), which renders bottom-right
	// (the 0,0 case hides the ×dim bug since 0×N=0). So write the source-centre FRACTION.
	// (Library bug — SetLayerTransform's AV-layer anchor units differ from shape/text;
	// tracked in incidents/setlayertransform-av-anchor-fraction.md.)
	ax, ay := 0.5, 0.5

	// Start offsets stagger each nested comp ① copy's glitch (SetStartTime = ldta edit).
	for i, l := range comp.Layers {
		must(l.SetStartTime(origComp.Layers[i].StartTime))
	}

	// applyScaleRotation copies the original layer's static Scale (parser reads it
	// as a fraction; SetLayerTransform.Scale wants percent → ×100) and Rotate Z
	// (degrees). The original shrinks each nested ① copy (L0 44 %, L1 45 %) and
	// flips L1 180° — omitting these rendered the copies full-size & un-rotated, so
	// the composite glitch spread too wide vs the original (user-caught at frame 18).
	applyScaleRotation := func(tr *aep.LayerTransform, otg *aep.AEPropertyGroup) {
		if sc := findProp(otg, "ADBE Scale"); sc != nil {
			s := toFloats(sc.StaticValue)
			must(tr.Scale().SetStaticValue([2]float64{s[0] * 100, s[1] * 100}))
		}
		if r := findProp(otg, "ADBE Rotate Z"); r != nil {
			must(tr.Rotation().SetStaticValue(toScalar(r.StaticValue)))
		}
	}

	// L0: source-centre anchor + scale/rotation + Position animated 2kf (absolute coords).
	otg0 := findGroup(origComp.Layers[0].PropertyTree(), "ADBE Transform Group")
	tr0 := aep.NewLayerTransform()
	must(tr0.AnchorPoint().SetStaticValue([2]float64{ax, ay}))
	applyScaleRotation(tr0, otg0)
	for _, kf := range findProp(otg0, "ADBE Position").Keyframes {
		v := toFloats(kf.Value)
		must(tr0.Position().AddKeyframeLinear(kf.Time, [2]float64{v[0], v[1]}))
	}
	must(aep.SetLayerTransform(comp.Layers[0], tr0))

	// L1: source-centre anchor + scale/rotation + centred Position (separated-read trap) + Opacity 13kf.
	otg1 := findGroup(origComp.Layers[1].PropertyTree(), "ADBE Transform Group")
	tr1 := aep.NewLayerTransform()
	must(tr1.AnchorPoint().SetStaticValue([2]float64{ax, ay}))
	must(tr1.Position().SetStaticValue([2]float64{cx, cy}))
	applyScaleRotation(tr1, otg1)
	for _, kf := range findProp(otg1, "ADBE Opacity").Keyframes {
		must(tr1.Opacity().AddKeyframeLinear(kf.Time, toScalar(kf.Value)*100))
	}
	must(aep.SetLayerTransform(comp.Layers[1], tr1))

	// L2: source-centre anchor + centred Position, static (no animation).
	tr2 := aep.NewLayerTransform()
	must(tr2.AnchorPoint().SetStaticValue([2]float64{ax, ay}))
	must(tr2.Position().SetStaticValue([2]float64{cx, cy}))
	must(aep.SetLayerTransform(comp.Layers[2], tr2))

	fmt.Printf("  ⑤ プリコンポジション 1: 3 precomp layers → comp ① (L0 Position 2kf, L1 Opacity 13kf+centred, L2 static centred)\n")
}

// finishPrecomp1Expr runs after the second reopen: the original's L2 is not truly
// static — its Opacity (static 100 %) carries a wiggle(29,55) expression that flickers
// the third nested glitch copy. SetLayerTransform materialized L2's default Opacity, so
// after the reopen we can attach the expression (source from the oracle verbatim).
func finishPrecomp1Expr(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(precomp1CompName)
	if comp == nil {
		panic("comp ⑤: not found after reopen 2")
	}
	origComp := orc.mustComp(precomp1CompName)
	otg := findGroup(origComp.Layers[2].PropertyTree(), "ADBE Transform Group")
	opOrig := findProp(otg, "ADBE Opacity")
	if opOrig == nil || opOrig.Expression == "" {
		panic("comp ⑤ L2: original Opacity expression missing")
	}
	op := comp.Layers[2].Opacity()
	if op == nil {
		panic("comp ⑤ L2: Opacity not materialized (SetLayerTransform should have)")
	}
	must(op.SetExpression(opOrig.Expression))
	fmt.Printf("  ⑤ プリコンポジション 1: + L2 Opacity %s\n", opOrig.Expression)
}
