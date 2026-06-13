// flightdeck/showcase/shape-primitives/gen.go — from-scratch (no AE) showcase of
// the shape primitives + paint operators. Builds a 4×2 grid: top row = the four
// parametric shapes (Rect / Ellipse / Star / Rounded Rect), bottom row = paint
// variations (Stroke-only / Fill+Stroke / thick-stroked Star / Gradient Fill).
// Writes shape_primitives.aep next to this file.
// Run from repo root: `go run ./flightdeck/showcase/shape-primitives`.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/shape-primitives/shape_primitives.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var colX = []float64{320, 740, 1180, 1600}
var rowY = []float64{360, 720}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "ShapePrimitives", 1920, 1080, 30, 3)
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

	white := [4]float64{0.95, 0.96, 1, 1}
	amber := [4]float64{1, 0.78, 0.25, 1}
	teal := [4]float64{0.30, 0.85, 0.80, 1}
	pink := [4]float64{1, 0.45, 0.62, 1}
	green := [4]float64{0.55, 0.86, 0.45, 1}

	cell := func(name string, col, row int, build func(g *aep.VectorGroup)) {
		l, err := aep.NewShapeLayer(comp, name)
		must(err)
		build(l.RootGroup())
		must(l.Position().SetStaticValue([2]float64{colX[col], rowY[row]}))
	}

	// Row 1 — the four parametric shapes (white fill).
	cell("01_Rect", 0, 0, func(g *aep.VectorGroup) {
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{180, 140}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(white))
	})
	cell("02_Ellipse", 1, 0, func(g *aep.VectorGroup) {
		e, err := g.AddEllipse()
		must(err)
		must(e.SetSize([2]float64{180, 140}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(teal))
	})
	cell("03_Star", 2, 0, func(g *aep.VectorGroup) {
		s, err := g.AddStar()
		must(err)
		must(s.SetPoints(6))
		must(s.SetOuterRadius(95))
		must(s.SetInnerRadius(45))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(amber))
	})
	cell("04_RoundedRect", 3, 0, func(g *aep.VectorGroup) {
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{180, 140}))
		must(r.SetRoundness(45))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(pink))
	})

	// Row 2 — paint variations.
	// Stroke-only ellipse (thick ring, no fill).
	cell("05_StrokeOnly", 0, 1, func(g *aep.VectorGroup) {
		e, err := g.AddEllipse()
		must(err)
		must(e.SetSize([2]float64{160, 160}))
		st, err := g.AddStroke()
		must(err)
		must(st.SetColor(green))
		must(st.SetWidth(22))
	})
	// Fill + Stroke rect (two colours).
	cell("06_FillStroke", 1, 1, func(g *aep.VectorGroup) {
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{170, 150}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(amber))
		st, err := g.AddStroke()
		must(err)
		must(st.SetColor(white))
		must(st.SetWidth(14))
	})
	// Thick round-joined star outline (no fill).
	cell("07_StrokedStar", 2, 1, func(g *aep.VectorGroup) {
		s, err := g.AddStar()
		must(err)
		must(s.SetPoints(5))
		must(s.SetOuterRadius(95))
		must(s.SetInnerRadius(46))
		st, err := g.AddStroke()
		must(err)
		must(st.SetColor(pink))
		must(st.SetWidth(16))
		must(st.SetLineJoin(aep.StrokeLineJoinRound))
	})
	// Gradient Fill rect (diagonal red→blue ramp).
	cell("08_GradientFill", 3, 1, func(g *aep.VectorGroup) {
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{180, 160}))
		gf, err := g.AddGradientFill()
		must(err)
		must(gf.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0.3, 0.2}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.25, 0.5, 1}},
		}))
		must(gf.SetStartPoint([2]float64{-90, -80}))
		must(gf.SetEndPoint([2]float64{90, 80}))
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
