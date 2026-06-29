// flightdeck/showcase/booyah-clone/gen_kakuh.go — comp ④ "カクッ".
// Two near-identical shape layers ("シェイプレイヤー 3" over "シェイプレイヤー 2"),
// each a STROKED rectangle (1635×810, no fill, 0.8px stroke) cut by a Trim Paths
// filter (Start 97, Offset 5.3) — a thin partial rectangle outline. The layer
// "snaps" in via a 2-kf Scale pop (0.64→1.0) and flickers via a 20-kf Opacity
// stutter (identical on both layers); L0 additionally rotates 180°. Values pulled
// from the original via the oracle.
//
// Position stays at the NewShapeLayer default (0,0) — the original's layer Position
// is (0,0) too (stored separated; we write a unified (0,0), render-identical), so no
// centering (unlike comp ①). Stroke colour is the AddStroke default (black): the
// original elided its stroke colour (AE create-time default), which we can't recover
// — a logged fidelity delta on a 0.8px element. 29.97 fps (cloneFps).
//
// Two-phase: buildKakuh creates the comp + the two shape layers (rect/stroke/trim
// build pre-reopen); finishKakuh applies SetLayerTransform, which needs a parsed
// layer (see the Reopen in main).
package main

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const kakuhCompName = "カクッ"

// L0 first, L1 second: NewShapeLayer appends at the bottom, so "…3" lands at
// index 0 (top) and "…2" at index 1, matching the original (L0="3" over L1="2").
var kakuhLayerNames = []string{"シェイプレイヤー 3", "シェイプレイヤー 2"}

func buildKakuh(p *aep.Project, orc *oracle) {
	comp, err := aep.NewComposition(p, kakuhCompName, 1920, 1080, cloneFps, 6)
	must(err)
	for _, name := range kakuhLayerNames {
		sl, err := aep.NewShapeLayer(comp, name)
		must(err)
		root := sl.RootGroup()
		rect, err := root.AddRect()
		must(err)
		must(rect.SetSize([2]float64{1635, 810}))
		stroke, err := root.AddStroke()
		must(err)
		must(stroke.SetWidth(0.8))
		trim, err := root.AddTrim()
		must(err)
		must(trim.SetStart(97))
		must(trim.SetOffset(5.3))
	}
}

// finishKakuh runs after the project is reopened: layer-level Scale (2kf) + Opacity
// (20kf) + (L0) Rotate Z, applied via SetLayerTransform (needs a parsed layer).
func finishKakuh(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(kakuhCompName)
	if comp == nil {
		panic("comp ④: not found after reopen")
	}
	origComp := orc.mustComp(kakuhCompName)

	for i, name := range kakuhLayerNames {
		layer := comp.LayerByName(name)
		if layer == nil {
			panic("comp ④: layer not found after reopen: " + name)
		}
		otg := findGroup(origComp.Layers[i].PropertyTree(), "ADBE Transform Group")
		tr := aep.NewLayerTransform()

		// Position: centre on the comp. The original's layer is visually centred
		// (an AE-created shape layer defaults to position = comp-centre), but its
		// Transform stores Position SEPARATED (Position_0/Position_1) which the
		// parser surfaces as 0,0 — so trusting that read put our clone's layer at
		// the top-left corner. Centre explicitly (rect is centred at shape-(0,0) with
		// a (0,0) anchor, so a comp-centre layer position centres the rect).
		must(tr.Position().SetStaticValue([2]float64{float64(comp.Width) / 2, float64(comp.Height) / 2}))

		// Scale (2kf, ease in the original → linear approx; 3D value, take X/Y).
		// Parser reports scale as a fraction (1.0 = 100%); SetLayerTransform's Scale
		// is percent, so ×100 (same convention as Opacity below).
		for _, kf := range findProp(otg, "ADBE Scale").Keyframes {
			v := toFloats(kf.Value)
			must(tr.Scale().AddKeyframeLinear(kf.Time, [2]float64{v[0] * 100, v[1] * 100}))
		}
		// Opacity (20kf; parser 0..1 → SetLayerTransform percent).
		for _, kf := range findProp(otg, "ADBE Opacity").Keyframes {
			must(tr.Opacity().AddKeyframeLinear(kf.Time, toScalar(kf.Value)*100))
		}
		// Rotate Z (static; L0 = 180, L1 absent = default 0).
		if rz := findProp(otg, "ADBE Rotate Z"); rz != nil && rz.StaticValue != nil {
			must(tr.Rotation().SetStaticValue(toScalar(rz.StaticValue)))
		}
		must(aep.SetLayerTransform(layer, tr))
	}

	fmt.Printf("  ④ カクッ: 2 stroked-rect+trim shape layers + Scale(2kf)/Opacity(20kf), L0 RotateZ=180\n")
}
