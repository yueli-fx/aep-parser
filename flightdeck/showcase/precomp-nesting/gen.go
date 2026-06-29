// flightdeck/showcase/precomp-nesting/gen.go — from-scratch (no AE) showcase of
// precomp nesting. A "Badge" child comp holds a self-contained composed scene
// (three magenta-disc + white-star badges across the frame, no background →
// transparent); the parent comp nests it once over a dark BG. The parent render
// shows the whole nested scene, proving the precomp layer resolved to — and
// rendered — the child composition.
//
// NewPrecompLayer needs the comps fully parsed, so it runs after Reopen (matching
// the S4 precomp ship-gate). Writes precomp_nesting.aep next to this file.
// Run from repo root: `go run ./flightdeck/showcase/precomp-nesting`.
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/precomp-nesting/precomp_nesting.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)

	parent, err := aep.NewComposition(p, "PrecompMain", 1920, 1080, 30, 3)
	must(err)
	child, err := aep.NewComposition(p, "Badge", 1920, 1080, 30, 3)
	must(err)
	_ = parent

	// Child scene: three badges (magenta disc + white star) at x=480/960/1440, no
	// full-frame BG (transparent). Add disc then star per badge, then push every
	// disc to the bottom so its star renders on top.
	discNames := []string{}
	for i, x := range []float64{480, 960, 1440} {
		dn := fmt.Sprintf("Disc%d", i+1)
		disc, err := aep.NewShapeLayer(child, dn)
		must(err)
		de, err := disc.RootGroup().AddEllipse()
		must(err)
		must(de.SetSize([2]float64{260, 260}))
		df, err := disc.RootGroup().AddFill()
		must(err)
		must(df.SetColor([4]float64{1, 0.2, 0.55, 1}))
		must(disc.Position().SetStaticValue([2]float64{x, 540}))
		discNames = append(discNames, dn)

		st, err := aep.NewShapeLayer(child, fmt.Sprintf("Star%d", i+1))
		must(err)
		se, err := st.RootGroup().AddStar()
		must(err)
		must(se.SetPoints(5))
		must(se.SetOuterRadius(90))
		must(se.SetInnerRadius(40))
		sf, err := st.RootGroup().AddFill()
		must(err)
		must(sf.SetColor([4]float64{1, 1, 1, 1}))
		must(st.Position().SetStaticValue([2]float64{x, 540}))
	}

	// Reopen so item-list back-refs exist.
	rp, err := aep.Reopen(p)
	must(err)
	var parentR, childR *aep.Composition
	for _, c := range rp.Compositions {
		switch c.Name {
		case "PrecompMain":
			parentR = c
		case "Badge":
			childR = c
		}
	}
	// Discs to the bottom so each star renders on top of its disc.
	for _, dn := range discNames {
		must(aep.MoveToEnd(childR.LayerByName(dn)))
	}

	// Parent dark BG + one nested Badge instance (full frame).
	bg, err := aep.NewShapeLayer(parentR, "BG")
	must(err)
	bgr, err := bg.RootGroup().AddRect()
	must(err)
	must(bgr.SetSize([2]float64{2200, 1300}))
	bgf, err := bg.RootGroup().AddFill()
	must(err)
	must(bgf.SetColor([4]float64{0.07, 0.08, 0.11, 1}))
	must(bg.Position().SetStaticValue([2]float64{960, 540}))

	if _, err := aep.NewPrecompLayer(parentR, childR, "NESTED_Badge"); err != nil {
		panic(err)
	}
	// BG to the bottom so the nested badges render on top.
	must(aep.MoveToEnd(parentR.LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s (parent layers=%d, child layers=%d)\n",
		outPath, len(parentR.Layers), len(childR.Layers))
}
