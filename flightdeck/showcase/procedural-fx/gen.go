// flightdeck/showcase/procedural-fx/gen.go — from-scratch (no AE) showcase of the
// procedural-FX generator's v1 flame (spec 2026-06-18-procedural-fx-generator,
// Phase 0). Builds a single solid + native effect stack (Fractal Noise tuned tall
// + Tint fire palette + Turbulent Displace + feathered teardrop mask, Evolution
// animated) that renders as a recognizable, boiling flame. Writes flame.aep next
// to this file. Run from repo root: `go run ./flightdeck/showcase/procedural-fx`.
//
// This is the deterministic recipe the AI layer (later phases) will parameterize;
// here every knob is hand-set. Same recipe as TestFlameDemo_AEShipGate (gated).
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/procedural-fx/flame.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "FLAME", 1080, 1920, 30, 4)
	must(err)
	_, err = aep.NewSolidLayer(comp, "Flame", 1080, 1920, [3]float64{0, 0, 0})
	must(err)

	rp, err := aep.Reopen(p)
	must(err)
	sol := rp.Compositions[0].LayerByName("Flame")
	if sol == nil {
		panic("Flame layer missing after Reopen")
	}
	set := func(fx *aep.Effect, mn string, v any) {
		if _, err := aep.SetEffectParam(sol, fx, mn, v); err != nil {
			panic(fmt.Sprintf("SetEffectParam %s: %v", mn, err))
		}
	}

	fn, err := aep.AddEffect(sol, aep.EffectFractalNoise)
	must(err)
	set(fn, "ADBE Fractal Noise-0004", 200.0) // Contrast
	set(fn, "ADBE Fractal Noise-0005", 0.0)   // Brightness
	set(fn, "ADBE Fractal Noise-0009", 0.0)   // Uniform Scaling off
	set(fn, "ADBE Fractal Noise-0011", 50.0)  // Scale Width
	set(fn, "ADBE Fractal Noise-0012", 300.0) // Scale Height (tall)
	set(fn, "ADBE Fractal Noise-0015", 6.0)   // Complexity

	tn, err := aep.AddEffect(sol, aep.EffectTint)
	must(err)
	set(tn, "ADBE Tint-0001", []float64{255, 12, 0, 0})    // black -> deep red
	set(tn, "ADBE Tint-0002", []float64{255, 255, 190, 40}) // white -> orange-yellow
	set(tn, "ADBE Tint-0003", 100.0)

	td, err := aep.AddEffect(sol, aep.EffectTurbulentDisplace)
	must(err)
	set(td, "ADBE Turbulent Displace-0002", 45.0) // Amount
	set(td, "ADBE Turbulent Displace-0003", 30.0) // Size

	flamePath := aep.BezierPath{
		Vertices: [][2]float64{
			{540, 250}, {700, 760}, {812, 1260}, {700, 1700},
			{540, 1785}, {380, 1700}, {268, 1260}, {380, 760},
		},
		Closed: true,
	}
	mask, err := aep.AddMask(sol, "FlameMask", flamePath)
	must(err)
	must(mask.SetFeather([2]float64{95, 95}))

	_, err = aep.AnimateEffectParam(sol, fn, "ADBE Fractal Noise-0023",
		[]aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 4, Value: 1440}})
	must(err)
	_, err = aep.AnimateEffectParam(sol, td, "ADBE Turbulent Displace-0006",
		[]aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 4, Value: 720}})
	must(err)

	f, err := os.Create(outPath)
	must(err)
	defer f.Close()
	must(rp.WriteAEP(f))
	fmt.Println("wrote", outPath)
}
