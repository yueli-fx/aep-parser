// flightdeck/showcase/glitch/gen.go — from-scratch (no AE) PLUGIN-FREE glitch on
// readable "GLITCH" text, following the all-native Booyah Glitch recipe
// (samples/motionbox/glitch/booyah-glitch). Validates glitch technique atoms
// (docs/fx-techniques.md) on rendered pixels (delivery contract red line 4):
//   T16 rgb-channel-split   — text pre-comp instanced 3×, Fill'd pure R/G/B, offset
//                             via a Transform effect, Add-blended => chromatic split
//   T2  displacement-glitch — Displacement Map (src = HIDDEN big-block Fractal Noise)
//                             on an adjustment => clean horizontal slice tear; the bg
//                             stays clean because the noise layer is map-only (eye off)
//   T5  emissive-glow       — Glo2 neon bloom on the bright text
// (T17 scanlines-crt dropped — Venetian Blinds CRT lines muddied/darkened the word;
//  it reads much cleaner without. T18 temporal-glitch is time-domain, not single-
//  frame-verifiable. Both stay sample-observed, not shown here.)
//
// Text can't be recoloured/offset directly (NewTextLayer exposes only SetText), so the
// word lives in a pre-comp ("TXT") instanced 3×; each instance is a normal AV layer.
// A fresh pre-comp instance has NO materialized layer transform (l.Position() is nil),
// so position + scale are done with a Transform EFFECT (ADBE Geometry2) whose point
// params are comp-fraction coords ([0.5,0.5] = centre, per Booyah's saved values).
//
// Run from repo root: `go run ./flightdeck/showcase/glitch`.
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/glitch/glitch.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type mark struct {
	name string
	col  []float64  // Fill color [A,R,G,B] 0-255
	pos  []float64  // Transform Position, comp-fraction [x,y]
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)

	// Source pre-comp: the readable word.
	txt, err := aep.NewComposition(p, "TXT", 1920, 1080, 30, 4)
	must(err)
	if _, err := aep.NewTextLayer(txt, "word"); err != nil {
		panic(err)
	}

	comp, err := aep.NewComposition(p, "GLITCH", 1920, 1080, 30, 4)
	must(err)

	// top -> bottom (new layers append below)
	_, err = aep.NewAdjustmentLayer(comp, "Glow")
	must(err)
	_, err = aep.NewAdjustmentLayer(comp, "Displace")
	must(err)

	// Point text is left-anchored, so the word starts at Position.x and runs right;
	// base x is shifted left of centre so "GLITCH" sits roughly centred. R/G/B get a
	// small ± offset around the base for the chromatic split.
	marks := []mark{
		{"G_R", []float64{255, 255, 0, 0}, []float64{0.308, 0.496}},
		{"G_G", []float64{255, 0, 255, 40}, []float64{0.3, 0.5}},
		{"G_B", []float64{255, 0, 130, 255}, []float64{0.292, 0.505}},
	}
	for _, m := range marks {
		_, err := aep.NewPrecompLayer(comp, txt, m.name)
		must(err)
	}

	_, err = aep.NewSolidLayer(comp, "Map", 1920, 1080, [3]float64{0, 0, 0})
	must(err)
	_, err = aep.NewSolidLayer(comp, "BG", 1920, 1080, [3]float64{0.02, 0.02, 0.05})
	must(err)

	rp, err := aep.Reopen(p)
	must(err)
	must(rp.CompositionByName("TXT").LayerByName("word").SetText("GLITCH"))
	fc := rp.CompositionByName("GLITCH")
	set := func(l *aep.Layer, fx *aep.Effect, mn string, v any) {
		if _, err := aep.SetEffectParam(l, fx, mn, v); err != nil {
			panic(fmt.Sprintf("%s %s: %v", l.Name, mn, err))
		}
	}

	// Map = high-contrast big horizontal blocks => clean slice tears (not fine static).
	// Eye off: used only as the displacement source, so the bg stays clean.
	mapL := fc.LayerByName("Map")
	fn, err := aep.AddEffect(mapL, aep.EffectFractalNoise)
	must(err)
	set(mapL, fn, "ADBE Fractal Noise-0004", 320.0) // Contrast (hard edges)
	set(mapL, fn, "ADBE Fractal Noise-0009", 0.0)   // Uniform Scaling off
	set(mapL, fn, "ADBE Fractal Noise-0011", 700.0) // Scale Width (very wide)
	set(mapL, fn, "ADBE Fractal Noise-0012", 9.0)   // Scale Height (thin bands)
	set(mapL, fn, "ADBE Fractal Noise-0015", 1.0)   // Complexity (blocky)
	mapL.SetVisible(false)

	dispL := fc.LayerByName("Displace")
	dm, err := aep.AddEffect(dispL, "ADBE Displacement Map")
	must(err)
	must(aep.SetEffectLayerParam(dispL, dm, "ADBE Displacement Map-0001", mapL))
	set(dispL, dm, "ADBE Displacement Map-0003", 26.0) // Max Horizontal Displacement
	set(dispL, dm, "ADBE Displacement Map-0005", 0.0)  // Max Vertical (pure h-slices)

	glowL := fc.LayerByName("Glow")
	gl, err := aep.AddEffect(glowL, "ADBE Glo2")
	must(err)
	set(glowL, gl, "ADBE Glo2-0002", 70.0) // Glow Threshold (low -> strong bloom)
	set(glowL, gl, "ADBE Glo2-0003", 42.0) // Glow Radius
	set(glowL, gl, "ADBE Glo2-0004", 2.6)  // Glow Intensity

	// (Scanlines / Venetian Blinds removed — the word reads much cleaner & brighter
	// without the CRT line overlay. T17 stays sample-observed, not shown here.)

	// Each text instance: Transform (scale up + RGB offset) + Fill (channel) + Add.
	for _, m := range marks {
		l := fc.LayerByName(m.name)
		tr, err := aep.AddEffect(l, "ADBE Geometry2")
		must(err)
		set(l, tr, "ADBE Geometry2-0011", 0.0)   // Uniform Scale off
		set(l, tr, "ADBE Geometry2-0004", 250.0) // Scale Width %
		set(l, tr, "ADBE Geometry2-0003", 250.0) // Scale Height %
		set(l, tr, "ADBE Geometry2-0002", m.pos) // Position (comp fraction)

		ff, err := aep.AddEffect(l, "ADBE Fill")
		must(err)
		set(l, ff, "ADBE Fill-0002", m.col)
		must(l.SetBlendingMode(aep.BlendingModeAdd))
	}

	_, err = aep.AnimateEffectParam(dispL, dm, "ADBE Displacement Map-0003",
		[]aep.ScalarKeyframe{{Time: 0, Value: 6}, {Time: 1.5, Value: 30}, {Time: 3, Value: 12}})
	must(err)

	f, err := os.Create(outPath)
	must(err)
	defer f.Close()
	must(rp.WriteAEP(f))
	fmt.Println("wrote", outPath)
}
