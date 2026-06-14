// flightdeck/showcase/gradient/gen.go — from-scratch (no AE) showcase of gradient
// fills, ramp direction, and ramp TYPE. A 3×2 grid of rects:
//   linear: horizontal · vertical · diagonal · 3-stop rainbow
//   radial: 2-stop concentric · 3-stop concentric
// proving SetColorStops + SetStartPoint/SetEndPoint (ramp geometry) +
// SetGradientType (linear vs radial). Writes gradient.aep next to this file.
// Run from repo root: `go run ./flightdeck/showcase/gradient`.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/gradient/gradient.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// 3 columns × 2 rows.
var cellPos = [][2]float64{
	{480, 360}, {960, 360}, {1440, 360},
	{480, 720}, {960, 720}, {1440, 720},
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "Gradient", 1920, 1080, 30, 3)
	must(err)

	bg, err := aep.NewShapeLayer(comp, "BG")
	must(err)
	bgr, err := bg.RootGroup().AddRect()
	must(err)
	must(bgr.SetSize([2]float64{2200, 1300}))
	bgf, err := bg.RootGroup().AddFill()
	must(err)
	must(bgf.SetColor([4]float64{0.07, 0.08, 0.11, 1}))
	must(bg.Position().SetStaticValue([2]float64{960, 540}))

	cell := func(name string, idx int, build func(g *aep.GradientFillNode)) {
		l, err := aep.NewShapeLayer(comp, name)
		must(err)
		r, err := l.RootGroup().AddRect()
		must(err)
		must(r.SetSize([2]float64{360, 280}))
		gf, err := l.RootGroup().AddGradientFill()
		must(err)
		build(gf)
		must(l.Position().SetStaticValue(cellPos[idx]))
	}

	// 1 — horizontal red→blue.
	cell("01_Horizontal", 0, func(gf *aep.GradientFillNode) {
		must(gf.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0.3, 0.2}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.2, 0.4, 1}},
		}))
		must(gf.SetStartPoint([2]float64{-170, 0}))
		must(gf.SetEndPoint([2]float64{170, 0}))
	})

	// 2 — vertical magenta→cyan.
	cell("02_Vertical", 1, func(gf *aep.GradientFillNode) {
		must(gf.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0.2, 0.7}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.2, 0.85, 0.95}},
		}))
		must(gf.SetStartPoint([2]float64{0, -130}))
		must(gf.SetEndPoint([2]float64{0, 130}))
	})

	// 3 — diagonal amber→green.
	cell("03_Diagonal", 2, func(gf *aep.GradientFillNode) {
		must(gf.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0.78, 0.2}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.3, 0.8, 0.4}},
		}))
		must(gf.SetStartPoint([2]float64{-170, -130}))
		must(gf.SetEndPoint([2]float64{170, 130}))
	})

	// 4 — 3-stop horizontal rainbow (red→yellow→blue).
	cell("04_ThreeStop", 3, func(gf *aep.GradientFillNode) {
		must(gf.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0.25, 0.25}},
			{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{1, 0.9, 0.3}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.25, 0.45, 1}},
		}))
		must(gf.SetStartPoint([2]float64{-170, 0}))
		must(gf.SetEndPoint([2]float64{170, 0}))
	})

	// 5 — RADIAL 2-stop: red centre → blue edge (concentric rings).
	cell("05_Radial", 4, func(gf *aep.GradientFillNode) {
		must(gf.SetGradientType(aep.GradientRadial))
		must(gf.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0.3, 0.2}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.2, 0.4, 1}},
		}))
		must(gf.SetStartPoint([2]float64{0, 0}))    // centre
		must(gf.SetEndPoint([2]float64{170, 0}))    // outer radius
	})

	// 6 — RADIAL 3-stop: white centre → orange → purple edge.
	cell("06_RadialMulti", 5, func(gf *aep.GradientFillNode) {
		must(gf.SetGradientType(aep.GradientRadial))
		must(gf.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 1, 0.95}},
			{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{1, 0.55, 0.15}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.45, 0.2, 0.7}},
		}))
		must(gf.SetStartPoint([2]float64{0, 0}))
		must(gf.SetEndPoint([2]float64{170, 0}))
	})

	rp, err := aep.Reopen(p)
	must(err)
	must(aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s (%d layers)\n", outPath, len(rp.Compositions[0].Layers))
}
