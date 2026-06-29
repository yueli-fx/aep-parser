// flightdeck/showcase/expressions/gen.go — from-scratch (no AE) showcase of the
// expression vocabulary proven by the S2 gate + its 2026-06-14 follow-up.
//
// Top row — three "orbit" layers each hold a dot offset upward; a `time*N`
// expression on Rotation spins it, so the dot orbits its centre (frozen at a
// different angle per speed at the render frame).
//
// Bottom row — the four richer idioms (each best SEEN by scrubbing the .aep in
// AE, where loop/wiggle motion is visible; the PNG is just a snapshot):
//   LINK   follower dot = leader.position + offset (cross-layer reference)
//   LOOP   dot on looping position keyframes (loopOut("cycle"))
//   WIG    dot displaced by wiggle(freq, amp)
//   SLIDER dot.x driven by a Slider Control value via effect(1)(1)
//
// Expressions + AddEffect are set AFTER Reopen (matching the ship-gates). Writes
// expressions.aep next to this file. Run: `go run ./flightdeck/showcase/expressions`.
// render.jsx renders t=1s.
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
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
	expr  string
}

var orbits = []orbit{
	{"ORB_90", 480, [4]float64{1.0, 0.55, 0.1, 1}, "time*90"},
	{"ORB_180", 960, [4]float64{0.25, 0.85, 1.0, 1}, "time*180"},
	{"ORB_270", 1440, [4]float64{1.0, 0.2, 0.55, 1}, "time*270"},
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

	// addDot builds a filled circle layer; returns the layer for positioning.
	addDot := func(name string, color [4]float64, size float64) *aep.ShapeLayer {
		l, err := aep.NewShapeLayer(comp, name)
		must(err)
		e, err := l.RootGroup().AddEllipse()
		must(err)
		must(e.SetSize([2]float64{size, size}))
		f, err := l.RootGroup().AddFill()
		must(err)
		must(f.SetColor(color))
		return l
	}

	// --- Top row: time*N orbits (centres y=360, dot offset 150px up). ---
	for _, o := range orbits {
		pip := addDot(o.name+"_centre", [4]float64{0.5, 0.5, 0.55, 1}, 22)
		must(pip.Position().SetStaticValue([2]float64{o.cx, 360}))

		l, err := aep.NewShapeLayer(comp, o.name)
		must(err)
		e, err := l.RootGroup().AddEllipse()
		must(err)
		must(e.SetSize([2]float64{68, 68}))
		must(e.SetPosition([2]float64{0, -150})) // offset up so rotation orbits it
		f, err := l.RootGroup().AddFill()
		must(err)
		must(f.SetColor(o.color))
		must(l.Position().SetStaticValue([2]float64{o.cx, 360}))
	}

	// --- Bottom row: the four vocab idioms (y≈820). ---
	const by = 820

	// LINK: a leader sweeping horizontally + a follower rigidly 110px above it.
	lead := addDot("LEAD", [4]float64{1.0, 0.55, 0.1, 1}, 56) // amber leader
	must(lead.Position().AddKeyframeLinear(0, [2]float64{300, by}))
	must(lead.Position().AddKeyframeLinear(2, [2]float64{560, by}))
	must(lead.Position().AddKeyframeLinear(4, [2]float64{300, by}))
	link := addDot("LINK", [4]float64{0.25, 0.85, 1.0, 1}, 56) // cyan follower
	must(link.Position().SetStaticValue([2]float64{430, by}))

	// LOOP: a dot that bobs up/down on a 1s loop.
	loop := addDot("LOOP", [4]float64{0.25, 1.0, 0.5, 1}, 64) // green
	must(loop.Position().AddKeyframeLinear(0, [2]float64{760, by + 60}))
	must(loop.Position().AddKeyframeLinear(1, [2]float64{760, by - 60}))

	// WIG: a dot jittering around a fixed anchor.
	wig := addDot("WIG", [4]float64{1.0, 0.2, 0.78, 1}, 64) // magenta
	must(wig.Position().SetStaticValue([2]float64{1120, by}))

	// SLIDER: a dot whose x is driven by a Slider Control value.
	sld := addDot("SLD", [4]float64{1.0, 0.9, 0.16, 1}, 64) // yellow
	must(sld.Position().SetStaticValue([2]float64{1500, by}))

	// --- Activate expressions + slider AFTER Reopen (ship-gate path). ---
	rp, err := aep.Reopen(p)
	must(err)
	cc := rp.Compositions[0]

	for _, o := range orbits {
		rot := cc.LayerByName(o.name).Rotation()
		must(rot.SetExpression(o.expr))
		must(rot.SetExpressionEnabled(true))
	}

	sldL := cc.LayerByName("SLD")
	fx, err := aep.AddEffect(sldL, aep.EffectSliderControl)
	must(err)
	_, err = aep.SetEffectParam(sldL, fx, "ADBE Slider Control-0001", 1500.0)
	must(err)

	for _, spec := range []struct{ layer, expr string }{
		{"LINK", `thisComp.layer("LEAD").transform.position + [0,-110]`},
		{"LOOP", `loopOut("cycle")`},
		{"WIG", `wiggle(3, 70)`},
		{"SLD", `[effect(1)(1), 820]`},
	} {
		pos := cc.LayerByName(spec.layer).Position()
		must(pos.SetExpression(spec.expr))
		must(pos.SetExpressionEnabled(true))
	}

	must(aep.MoveToEnd(cc.LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s (%d layers)\n", outPath, len(cc.Layers))
}
