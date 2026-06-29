// flightdeck/showcase/shape-filters/gen.go — from-scratch (no AE) showcase of the
// shape vector-filter family. Builds a 4×3 grid of shape layers on a dark BG,
// each cell demonstrating one filter, and writes shape_filters.aep next to this
// file. Run from the repo root: `go run ./flightdeck/showcase/shape-filters`.
//
// Every individual filter here is double-version ship-gate verified; this
// 12-in-one COMBINATION is rendered + eyeballed via render.jsx (delivery
// contract red line 4 — see INDEX.md for the layout the render is checked against).
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/shape-filters/shape_filters.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// cell center positions: 4 columns × 4 rows in a 1920×1080 frame. Row 4 holds
// the Wiggle modulation comparison (Points Corner/Smooth, Correlation low/high).
var colX = []float64{320, 740, 1180, 1600}
var rowY = []float64{160, 405, 650, 905}

func at(col, row int) [2]float64 { return [2]float64{colX[col], rowY[row]} }

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "ShapeFilterShowcase", 1920, 1080, 30, 3)
	must(err)

	// Dark background.
	bg, err := aep.NewShapeLayer(comp, "BG")
	must(err)
	bgRect, err := bg.RootGroup().AddRect()
	must(err)
	must(bgRect.SetSize([2]float64{2200, 1300}))
	bgFill, err := bg.RootGroup().AddFill()
	must(err)
	must(bgFill.SetColor([4]float64{0.07, 0.08, 0.11, 1}))
	must(bg.Position().SetStaticValue([2]float64{960, 540}))

	white := [4]float64{0.95, 0.96, 1, 1}
	amber := [4]float64{1, 0.78, 0.25, 1}
	teal := [4]float64{0.30, 0.85, 0.80, 1}
	pink := [4]float64{1, 0.45, 0.62, 1}
	green := [4]float64{0.55, 0.86, 0.45, 1}
	violet := [4]float64{0.66, 0.55, 1, 1}

	// helper: new shape layer placed at a grid cell, builder gets its root group.
	cell := func(name string, col, row int, build func(g *aep.VectorGroup)) {
		l, err := aep.NewShapeLayer(comp, name)
		must(err)
		build(l.RootGroup())
		must(l.Position().SetStaticValue(at(col, row)))
	}

	// Row 1 -----------------------------------------------------------------
	// (0,0) Plain Rect — reference (no filter).
	cell("01_PlainRect", 0, 0, func(g *aep.VectorGroup) {
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{150, 150}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(white))
	})

	// (1,0) Round Corners — squircle.
	cell("02_RoundCorners", 1, 0, func(g *aep.VectorGroup) {
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{160, 160}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(amber))
		rc, err := g.AddRoundCorners()
		must(err)
		must(rc.SetRadius(45))
	})

	// (2,0) Offset Paths — grown outline.
	cell("03_OffsetPaths", 2, 0, func(g *aep.VectorGroup) {
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{110, 110}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(teal))
		of, err := g.AddOffsetPaths()
		must(err)
		must(of.SetAmount(30))
	})

	// (3,0) Trim Paths — stroked arc (ellipse + stroke + trim).
	cell("04_TrimPaths", 3, 0, func(g *aep.VectorGroup) {
		e, err := g.AddEllipse()
		must(err)
		must(e.SetSize([2]float64{160, 160}))
		st, err := g.AddStroke()
		must(err)
		must(st.SetColor(pink))
		must(st.SetWidth(16))
		tr, err := g.AddTrim()
		must(err)
		must(tr.SetEnd(62))
	})

	// Row 2 -----------------------------------------------------------------
	// (0,1) ZigZag — starburst.
	cell("05_ZigZag", 0, 1, func(g *aep.VectorGroup) {
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{150, 150}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(green))
		z, err := g.AddZigZag()
		must(err)
		must(z.SetSize(18))
		must(z.SetDetail(6))
	})

	// (1,1) Pucker & Bloat — clover.
	cell("06_PuckerBloat", 1, 1, func(g *aep.VectorGroup) {
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{150, 150}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(violet))
		pb, err := g.AddPuckerBloat()
		must(err)
		must(pb.SetAmount(90))
	})

	// (2,1) Twist — pinwheel.
	cell("07_Twist", 2, 1, func(g *aep.VectorGroup) {
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{160, 160}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(amber))
		tw, err := g.AddTwist()
		must(err)
		must(tw.SetAngle(140))
	})

	// (3,1) Wiggle Paths — roughened edge.
	cell("08_WigglePaths", 3, 1, func(g *aep.VectorGroup) {
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{150, 150}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(teal))
		wg, err := g.AddWigglePaths()
		must(err)
		must(wg.SetSize(22))
		must(wg.SetDetail(25))
		must(wg.SetWigglesPerSecond(3))
		must(wg.SetRandomSeed(5))
	})

	// Row 3 -----------------------------------------------------------------
	// (0,2) Repeater — row of dots.
	cell("09_Repeater", 0, 2, func(g *aep.VectorGroup) {
		e, err := g.AddEllipse()
		must(err)
		must(e.SetSize([2]float64{34, 34}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(pink))
		rp, err := g.AddRepeater()
		must(err)
		must(rp.SetCopies(5))
		must(rp.Transform().SetPosition([2]float64{42, 0}))
	})

	// (1,2) Merge Paths — square with a punched hole (Subtract).
	cell("10_MergeSubtract", 1, 2, func(g *aep.VectorGroup) {
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{150, 150}))
		e, err := g.AddEllipse()
		must(err)
		must(e.SetSize([2]float64{80, 80}))
		mg, err := g.AddMergePaths()
		must(err)
		must(mg.SetType(aep.MergeTypeSubtract))
		f, err := g.AddFill() // fill ABOVE the merge (combine filter)
		must(err)
		must(f.SetColor(green))
	})

	// (2,2) PolyStar — 5-point star.
	cell("11_PolyStar", 2, 2, func(g *aep.VectorGroup) {
		s, err := g.AddStar()
		must(err)
		must(s.SetPoints(5))
		must(s.SetOuterRadius(88))
		must(s.SetInnerRadius(40))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(amber))
	})

	// (3,2) Combo — Twist THEN Wiggle Paths on one rect (stacked distort filters).
	cell("12_TwistPlusWiggle", 3, 2, func(g *aep.VectorGroup) {
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{150, 150}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(violet))
		tw, err := g.AddTwist()
		must(err)
		must(tw.SetAngle(110))
		wg, err := g.AddWigglePaths()
		must(err)
		must(wg.SetSize(12))
		must(wg.SetDetail(20))
		must(wg.SetRandomSeed(3))
	})

	// Row 4 — Wiggle modulation comparison (the family's last knobs). All four
	// share Size/Detail/Seed; only the demonstrated knob varies, so the visual
	// difference is attributable to it alone.
	wiggleBase := func(g *aep.VectorGroup, color [4]float64) *aep.WigglePathsNode {
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{150, 150}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(color))
		wg, err := g.AddWigglePaths()
		must(err)
		must(wg.SetSize(34))
		must(wg.SetDetail(8))
		must(wg.SetRandomSeed(5))
		return wg
	}

	// (0,3) Points = Corner (default) — sharp angular spikes.
	cell("13_WiggleCorner", 0, 3, func(g *aep.VectorGroup) {
		wiggleBase(g, teal)
	})
	// (1,3) Points = Smooth — same seed, rounded scalloped bumps.
	cell("14_WiggleSmooth", 1, 3, func(g *aep.VectorGroup) {
		wg := wiggleBase(g, teal)
		must(wg.SetPoints(aep.RoughenPointsSmooth))
	})
	// (2,3) Correlation = 0 — independent jitter, jagged edge.
	cell("15_WiggleCorrLow", 2, 3, func(g *aep.VectorGroup) {
		wg := wiggleBase(g, pink)
		must(wg.SetCorrelation(0))
	})
	// (3,3) Correlation = 100 — coherent boil, edge collapses to a near-rigid
	// offset (almost-clean square).
	cell("16_WiggleCorrHigh", 3, 3, func(g *aep.VectorGroup) {
		wg := wiggleBase(g, pink)
		must(wg.SetCorrelation(100))
	})

	// BG must render behind everything.
	rp, err := aep.Reopen(p)
	must(err)
	must(aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s (%d layers)\n", outPath, len(rp.Compositions[0].Layers))
}
