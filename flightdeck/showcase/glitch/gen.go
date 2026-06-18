// flightdeck/showcase/glitch/gen.go — from-scratch (no AE) PLUGIN-FREE glitch,
// following the all-native Booyah Glitch recipe (samples/motionbox/glitch/booyah-
// glitch). Validates the glitch technique atoms (docs/fx-techniques.md) on rendered
// pixels (delivery contract red line 4):
//   T16 rgb-channel-split   — 3 bar copies Fill'd pure R/G/B, offset, Add-blended
//                             => white core + chromatic fringe (chromatic aberration)
//   T2  displacement-glitch — Displacement Map (src = high-contrast horizontal-streak
//                             Fractal Noise) on an adjustment layer => horizontal tear
//   T17 scanlines-crt       — Venetian Blinds on a dark solid => thin scanline stripes
//   T5  emissive-glow       — Glo2 bloom on the bright bars
// (T18 temporal-glitch is time-domain — not single-frame pixel-verifiable — left out.)
//
// Everything native => no plugin needed to render (Booyah's whole point). AddEffect /
// SetEffectParam / SetEffectLayerParam / AnimateEffectParam / SetBlendingMode / shape
// Fill+Position are all double-version ship-gate verified; this is the COMBINATION,
// rendered + eyeballed via render.jsx.
//
// Run from repo root: `go run ./flightdeck/showcase/glitch`.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/glitch/glitch.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "GLITCH", 1920, 1080, 30, 4)
	must(err)

	// top -> bottom (new layers append below)
	_, err = aep.NewSolidLayer(comp, "Scanlines", 1920, 1080, [3]float64{0.05, 0.05, 0.07})
	must(err)
	_, err = aep.NewAdjustmentLayer(comp, "Glow")
	must(err)
	_, err = aep.NewAdjustmentLayer(comp, "Displace")
	must(err)

	// RGB-split bar trio (same geometry, pure R/G/B fill, offset, Add).
	type mark struct {
		name string
		rgba [4]float64 // SetColor 0-1
		pos  [2]float64
	}
	marks := []mark{
		{"Mark_R", [4]float64{1, 0, 0, 1}, [2]float64{992, 528}},
		{"Mark_G", [4]float64{0, 1, 0, 1}, [2]float64{960, 540}},
		{"Mark_B", [4]float64{0, 0.35, 1, 1}, [2]float64{928, 552}},
	}
	for _, m := range marks {
		l, err := aep.NewShapeLayer(comp, m.name)
		must(err)
		g := l.RootGroup()
		r, err := g.AddRect()
		must(err)
		must(r.SetSize([2]float64{1120, 230}))
		bar2, err := g.AddRect()
		must(err)
		must(bar2.SetSize([2]float64{760, 70}))
		f, err := g.AddFill()
		must(err)
		must(f.SetColor(m.rgba))
		must(l.Position().SetStaticValue(m.pos))
	}

	_, err = aep.NewSolidLayer(comp, "Noise", 1920, 1080, [3]float64{0, 0, 0})
	must(err)
	_, err = aep.NewSolidLayer(comp, "BG", 1920, 1080, [3]float64{0.02, 0.02, 0.035})
	must(err)

	rp, err := aep.Reopen(p)
	must(err)
	fc := rp.Compositions[0]
	set := func(l *aep.Layer, fx *aep.Effect, mn string, v any) {
		if _, err := aep.SetEffectParam(l, fx, mn, v); err != nil {
			panic(fmt.Sprintf("%s %s: %v", l.Name, mn, err))
		}
	}

	// Noise: high-contrast horizontal streaks => horizontal displacement tear.
	noiseL := fc.LayerByName("Noise")
	fn, err := aep.AddEffect(noiseL, aep.EffectFractalNoise)
	must(err)
	set(noiseL, fn, "ADBE Fractal Noise-0004", 150.0) // Contrast
	set(noiseL, fn, "ADBE Fractal Noise-0009", 0.0)   // Uniform Scaling off
	set(noiseL, fn, "ADBE Fractal Noise-0011", 520.0) // Scale Width (wide)
	set(noiseL, fn, "ADBE Fractal Noise-0012", 32.0)  // Scale Height (thin) -> streaks
	set(noiseL, fn, "ADBE Fractal Noise-0015", 2.0)   // Complexity (blocky)
	set(noiseL, fn, "ADBE Fractal Noise-0029", 20.0)  // Opacity (faint texture)

	// Displace adjustment: tear everything below using Noise as the map.
	dispL := fc.LayerByName("Displace")
	dm, err := aep.AddEffect(dispL, "ADBE Displacement Map")
	must(err)
	must(aep.SetEffectLayerParam(dispL, dm, "ADBE Displacement Map-0001", noiseL))
	set(dispL, dm, "ADBE Displacement Map-0003", 48.0) // Max Horizontal Displacement
	set(dispL, dm, "ADBE Displacement Map-0005", 7.0)  // Max Vertical Displacement

	// Glow adjustment: bloom the bright bars.
	glowL := fc.LayerByName("Glow")
	gl, err := aep.AddEffect(glowL, "ADBE Glo2")
	must(err)
	set(glowL, gl, "ADBE Glo2-0002", 110.0) // Glow Threshold (low -> more bloom)
	set(glowL, gl, "ADBE Glo2-0003", 48.0)  // Glow Radius
	set(glowL, gl, "ADBE Glo2-0004", 1.6)   // Glow Intensity

	// Scanlines: thin Venetian-Blinds stripes over everything.
	scanL := fc.LayerByName("Scanlines")
	vb, err := aep.AddEffect(scanL, "ADBE Venetian Blinds")
	must(err)
	set(scanL, vb, "ADBE Venetian Blinds-0001", 34.0) // Transition Completion (lighter)
	set(scanL, vb, "ADBE Venetian Blinds-0002", 0.0)  // Direction (horizontal blinds)
	set(scanL, vb, "ADBE Venetian Blinds-0003", 5.0)  // Width (fine)
	set(scanL, vb, "ADBE Venetian Blinds-0004", 1.0)  // Feather

	// Fill each bar pure R/G/B + Add blend (RGB-split chromatic aberration).
	fills := map[string][]float64{
		"Mark_R": {255, 255, 0, 0},
		"Mark_G": {255, 0, 255, 0},
		"Mark_B": {255, 0, 90, 255},
	}
	for name, col := range fills {
		l := fc.LayerByName(name)
		ff, err := aep.AddEffect(l, "ADBE Fill")
		must(err)
		set(l, ff, "ADBE Fill-0002", col)
		must(l.SetBlendingMode(aep.BlendingModeAdd))
	}

	// Animate the tear (Max Horizontal Displacement jitters) — motion for the .aep,
	// not needed for the single-frame gate.
	_, err = aep.AnimateEffectParam(dispL, dm, "ADBE Displacement Map-0003",
		[]aep.ScalarKeyframe{{Time: 0, Value: 8}, {Time: 1.5, Value: 42}, {Time: 3, Value: 14}})
	must(err)

	f, err := os.Create(outPath)
	must(err)
	defer f.Close()
	must(rp.WriteAEP(f))
	fmt.Println("wrote", outPath)
}
