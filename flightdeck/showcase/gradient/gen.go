// flightdeck/showcase/gradient/gen.go — from-scratch (no AE) showcase of gradient
// fills + strokes, ramp direction, type, and radial highlight. A 3×3 grid:
//   row 1 — LINEAR fills: horizontal · vertical · diagonal
//   row 2 — RADIAL fills: 2-stop concentric · 3-stop concentric · 2-stop + HiLite
//   row 3 — gradient STROKES: linear horizontal · radial · diagonal 3-stop
// proving SetColorStops + SetStartPoint/SetEndPoint (ramp geometry) +
// SetGradientType (linear vs radial) + SetHighlightLength/Angle (radial
// highlight) on both G-Fill and G-Stroke. Writes gradient.aep next to this file.
// Run from repo root: `go run ./flightdeck/showcase/gradient`.
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/gradient/gradient.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// 3 columns × 3 rows.
var cellPos = [][2]float64{
	{480, 200}, {960, 200}, {1440, 200},
	{480, 540}, {960, 540}, {1440, 540},
	{480, 880}, {960, 880}, {1440, 880},
}

const (
	cellW = 320
	cellH = 230
)

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

	// fillCell — a rect with a gradient FILL.
	fillCell := func(name string, idx int, build func(g *aep.GradientFillNode)) {
		l, err := aep.NewShapeLayer(comp, name)
		must(err)
		r, err := l.RootGroup().AddRect()
		must(err)
		must(r.SetSize([2]float64{cellW, cellH}))
		gf, err := l.RootGroup().AddGradientFill()
		must(err)
		build(gf)
		must(l.Position().SetStaticValue(cellPos[idx]))
	}

	// strokeCell — a rect with a gradient STROKE (18px baked ring).
	strokeCell := func(name string, idx int, build func(g *aep.GradientStrokeNode)) {
		l, err := aep.NewShapeLayer(comp, name)
		must(err)
		r, err := l.RootGroup().AddRect()
		must(err)
		must(r.SetSize([2]float64{cellW, cellH}))
		gs, err := l.RootGroup().AddGradientStroke()
		must(err)
		build(gs)
		must(l.Position().SetStaticValue(cellPos[idx]))
	}

	// ─── row 1 — LINEAR fills ──────────────────────────────────────────────
	fillCell("01_Horizontal", 0, func(gf *aep.GradientFillNode) {
		must(gf.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0.3, 0.2}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.2, 0.4, 1}},
		}))
		must(gf.SetStartPoint([2]float64{-150, 0}))
		must(gf.SetEndPoint([2]float64{150, 0}))
	})
	fillCell("02_Vertical", 1, func(gf *aep.GradientFillNode) {
		must(gf.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0.2, 0.7}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.2, 0.85, 0.95}},
		}))
		must(gf.SetStartPoint([2]float64{0, -105}))
		must(gf.SetEndPoint([2]float64{0, 105}))
	})
	fillCell("03_Diagonal", 2, func(gf *aep.GradientFillNode) {
		must(gf.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0.78, 0.2}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.3, 0.8, 0.4}},
		}))
		must(gf.SetStartPoint([2]float64{-150, -105}))
		must(gf.SetEndPoint([2]float64{150, 105}))
	})

	// ─── row 2 — RADIAL fills (+ highlight) ────────────────────────────────
	fillCell("04_Radial", 3, func(gf *aep.GradientFillNode) {
		must(gf.SetGradientType(aep.GradientRadial))
		must(gf.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0.3, 0.2}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.2, 0.4, 1}},
		}))
		must(gf.SetStartPoint([2]float64{0, 0}))   // centre
		must(gf.SetEndPoint([2]float64{150, 0}))   // outer radius
	})
	fillCell("05_RadialMulti", 4, func(gf *aep.GradientFillNode) {
		must(gf.SetGradientType(aep.GradientRadial))
		must(gf.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 1, 0.95}},
			{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{1, 0.55, 0.15}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.45, 0.2, 0.7}},
		}))
		must(gf.SetStartPoint([2]float64{0, 0}))
		must(gf.SetEndPoint([2]float64{150, 0}))
	})
	// 06 — RADIAL + HiLite: same red→blue radial as 04 but the bright centre is
	// shifted right (Angle 0 = +X) by 70% of the radius.
	fillCell("06_RadialHiLite", 5, func(gf *aep.GradientFillNode) {
		must(gf.SetGradientType(aep.GradientRadial))
		must(gf.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0.3, 0.2}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.2, 0.4, 1}},
		}))
		must(gf.SetStartPoint([2]float64{0, 0}))
		must(gf.SetEndPoint([2]float64{150, 0}))
		must(gf.SetHighlightLength(70))
		must(gf.SetHighlightAngle(0))
	})

	// ─── row 3 — gradient STROKES (18px ring) ──────────────────────────────
	strokeCell("07_StrokeLinear", 6, func(gs *aep.GradientStrokeNode) {
		must(gs.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0.3, 0.2}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.2, 0.4, 1}},
		}))
		must(gs.SetStartPoint([2]float64{-150, 0}))
		must(gs.SetEndPoint([2]float64{150, 0}))
	})
	strokeCell("08_StrokeRadial", 7, func(gs *aep.GradientStrokeNode) {
		must(gs.SetGradientType(aep.GradientRadial))
		must(gs.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0.85, 0.3}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.6, 0.2, 0.8}},
		}))
		must(gs.SetStartPoint([2]float64{0, 0}))
		must(gs.SetEndPoint([2]float64{220, 0})) // radius > rect → ring shows mid-ramp
	})
	strokeCell("09_StrokeDiag3", 8, func(gs *aep.GradientStrokeNode) {
		must(gs.SetColorStops([]aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0.25, 0.25}},
			{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{1, 0.9, 0.3}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0.25, 0.45, 1}},
		}))
		must(gs.SetStartPoint([2]float64{-150, -105}))
		must(gs.SetEndPoint([2]float64{150, 105}))
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
