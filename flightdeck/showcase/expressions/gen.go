// flightdeck/showcase/expressions/gen.go — from-scratch (no AE) showcase of
// expression activation. Three "orbit" layers each hold a dot offset upward
// inside the layer; a `time*N` expression on the layer's Rotation spins it, so
// the dot orbits its centre. Rendered at t=1s the three dots sit at different
// angles (90° / 180° / 270°) — proving the expression drives the value, not a
// static keyframe. A static white pip marks each orbit centre for context.
//
// Expressions are set AFTER Reopen (on the reopened comp's Rotation property),
// matching the S2 expression ship-gate. Writes expressions.aep next to this file.
// Run from repo root: `go run ./flightdeck/showcase/expressions`. render.jsx renders t=1s.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/expressions/expressions.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type orbit struct {
	name  string
	cx    float64
	color [4]float64
	expr  string // rotation expression
}

var orbits = []orbit{
	{"ORB_90", 480, [4]float64{1.0, 0.55, 0.1, 1}, "time*90"},   // amber, 90°/s
	{"ORB_180", 960, [4]float64{0.25, 0.85, 1.0, 1}, "time*180"}, // cyan, 180°/s
	{"ORB_270", 1440, [4]float64{1.0, 0.2, 0.55, 1}, "time*270"}, // magenta, 270°/s
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "Expressions", 1920, 1080, 30, 4)
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

	for _, o := range orbits {
		// static centre pip
		pip, err := aep.NewShapeLayer(comp, o.name+"_centre")
		must(err)
		pe, err := pip.RootGroup().AddEllipse()
		must(err)
		must(pe.SetSize([2]float64{24, 24}))
		pf, err := pip.RootGroup().AddFill()
		must(err)
		must(pf.SetColor([4]float64{0.5, 0.5, 0.55, 1}))
		must(pip.Position().SetStaticValue([2]float64{o.cx, 540}))

		// orbiting dot: offset 260px up inside the layer, layer at the centre.
		dl, err := aep.NewShapeLayer(comp, o.name)
		must(err)
		e, err := dl.RootGroup().AddEllipse()
		must(err)
		must(e.SetSize([2]float64{80, 80}))
		must(e.SetPosition([2]float64{0, -260})) // offset up so rotation orbits it
		f, err := dl.RootGroup().AddFill()
		must(err)
		must(f.SetColor(o.color))
		must(dl.Position().SetStaticValue([2]float64{o.cx, 540}))
	}

	// Reopen, then activate the rotation expressions (S2 path).
	rp, err := aep.Reopen(p)
	must(err)
	cc := rp.Compositions[0]
	for _, o := range orbits {
		rot := cc.LayerByName(o.name).Rotation()
		must(rot.SetExpression(o.expr))
		must(rot.SetExpressionEnabled(true))
	}
	must(aep.MoveToEnd(cc.LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s (%d layers)\n", outPath, len(cc.Layers))
}
