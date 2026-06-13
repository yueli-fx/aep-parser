// flightdeck/showcase/effects/gen.go — from-scratch (no AE) showcase of the
// AddEffect + SetEffectParam capability. Builds a 4×3 grid of identical source
// tokens (a teal rounded square with an amber stroke, so every effect has an
// edge + two colours to act on) on a dark BG, then drops one built-in effect on
// each cell and tunes it. Cell 1 is the untouched source for before/after.
//
// AddEffect needs a *parsed* layer, so all source layers are built first, then
// Reopen() upgrades them, then effects are spliced on the reopened comp.
//
// Every effect here (AddEffect 30-effect library + SetEffectParam all control
// types) is double-version ship-gate verified; this grid is the COMBINATION,
// rendered + eyeballed via render.jsx (delivery contract red line 4).
//
// Run from the repo root: `go run ./flightdeck/showcase/effects`.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/effects/effects.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var colX = []float64{320, 740, 1180, 1600}
var rowY = []float64{230, 540, 850}

func at(col, row int) [2]float64 { return [2]float64{colX[col], rowY[row]} }

var (
	teal  = [4]float64{0.30, 0.85, 0.80, 1}
	amber = [4]float64{1, 0.78, 0.25, 1}
)

// P is one SetEffectParam tweak: suffix is appended to the effect match-name
// (e.g. "-0001"), value is the on-disk StaticValue (scalar / [A,R,G,B]×255 /
// angle-deg / 0|1 bool / [x,y] fraction).
type P struct {
	suffix string
	v      any
}

type cell struct {
	name     string
	col, row int
	effect   string // "" = no effect (the reference source)
	params   []P
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "EffectsShowcase", 1920, 1080, 30, 3)
	must(err)

	// Dark background.
	bg, err := aep.NewShapeLayer(comp, "BG")
	must(err)
	br, err := bg.RootGroup().AddRect()
	must(err)
	must(br.SetSize([2]float64{2200, 1300}))
	bf, err := bg.RootGroup().AddFill()
	must(err)
	must(bf.SetColor([4]float64{0.07, 0.08, 0.11, 1}))
	must(bg.Position().SetStaticValue([2]float64{960, 540}))

	cells := []cell{
		{name: "01_Source", col: 0, row: 0},
		{name: "02_GaussianBlur", col: 1, row: 0, effect: aep.EffectGaussianBlur, params: []P{{"-0001", 30.0}}},
		{name: "03_DropShadow", col: 2, row: 0, effect: aep.EffectDropShadow, params: []P{
			{"-0001", []float64{255, 0, 0, 0}}, // shadow colour (black)
			{"-0003", 135.0},                   // direction
			{"-0004", 30.0},                    // distance
			{"-0005", 14.0},                    // softness
		}},
		{name: "04_Invert", col: 3, row: 0, effect: aep.EffectInvert},
		{name: "05_Tint", col: 0, row: 1, effect: aep.EffectTint, params: []P{{"-0003", 100.0}}}, // amount to tint
		{name: "06_Tritone", col: 1, row: 1, effect: aep.EffectTritone},
		{name: "07_WaveWarp", col: 2, row: 1, effect: aep.EffectWaveWarp}, // default wave warp = visible rippled edge (geometry distort)
		{name: "08_Brightness", col: 3, row: 1, effect: aep.EffectBrightnessContrast, params: []P{
			{"-0001", 60.0}, // brightness
			{"-0002", 40.0}, // contrast
		}},
		{name: "09_FractalNoise", col: 0, row: 2, effect: aep.EffectFractalNoise},
		{name: "10_GradientRamp", col: 1, row: 2, effect: aep.EffectGradientRamp},
		// Mosaic block-count params are control-type 1 (no generic SetEffectParam
		// template yet) — left at AE's default block grid, which is already visible.
		{name: "11_Mosaic", col: 2, row: 2, effect: aep.EffectMosaic},
		{name: "12_DirectionalBlur", col: 3, row: 2, effect: aep.EffectDirectionalBlur, params: []P{
			{"-0001", 45.0}, // direction
			{"-0002", 35.0}, // blur length
		}},
	}

	// Build all source tokens first (built layers, not yet parsed).
	for _, c := range cells {
		l, err := aep.NewShapeLayer(comp, c.name)
		must(err)
		g := l.RootGroup()
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{200, 200}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(teal))
		s, err := g.AddStroke()
		must(err)
		must(s.SetColor(amber))
		must(s.SetWidth(18))
		must(l.Position().SetStaticValue(at(c.col, c.row)))
	}

	// Reopen so every layer is parsed → AddEffect works.
	rp, err := aep.Reopen(p)
	must(err)
	rc := rp.Compositions[0]

	for _, c := range cells {
		if c.effect == "" {
			continue
		}
		l := rc.LayerByName(c.name)
		if l == nil {
			panic("layer not found: " + c.name)
		}
		fx, err := aep.AddEffect(l, c.effect)
		must(err)
		for _, pr := range c.params {
			if _, err := aep.SetEffectParam(l, fx, c.effect+pr.suffix, pr.v); err != nil {
				// Don't abort the whole build — surface the bad param so the
				// render.jsx param dump can be cross-checked next iteration.
				fmt.Printf("  WARN %s %s%s: %v\n", c.name, c.effect, pr.suffix, err)
			}
		}
	}

	must(aep.MoveToEnd(rc.LayerByName("BG")))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s (%d layers)\n", outPath, len(rc.Layers))
}
