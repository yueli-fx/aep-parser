// flightdeck/showcase/layers/gen.go — from-scratch (no AE) showcase of layer
// creation. Builds concentric solid layers (NewSolidLayer with colour + size,
// each centred → nested rectangles) plus a Null and an Adjustment layer. Only
// solids render; Null / Adjustment (and Camera / Light, not added here) are
// non-rendering layer types — render.jsx dumps the layer list to prove they were
// created. Writes layers.aep next to this file.
// Run from repo root: `go run ./flightdeck/showcase/layers`.
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/layers/layers.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "Layers", 1920, 1080, 30, 3)
	must(err)

	// Concentric solids: added front→back (first added = top of stack = front),
	// each centred at the comp middle → nested colour frames.
	solids := []struct {
		name string
		w, h int
		rgb  [3]float64
	}{
		{"Solid1_front", 560, 320, [3]float64{1, 0.78, 0.25}},  // amber (front)
		{"Solid2", 1000, 560, [3]float64{1, 0.3, 0.55}},        // pink
		{"Solid3", 1440, 800, [3]float64{0.3, 0.7, 0.85}},      // blue
		{"Solid4_back", 1920, 1080, [3]float64{0.1, 0.12, 0.2}}, // dark (full-frame back)
	}
	for _, s := range solids {
		if _, err := aep.NewSolidLayer(comp, s.name, s.w, s.h, s.rgb); err != nil {
			panic(err)
		}
	}

	// Non-rendering layer types (created, listed, but invisible in a flat render).
	if _, err := aep.NewNullLayer(comp, "Null_helper"); err != nil {
		panic(err)
	}
	if _, err := aep.NewAdjustmentLayer(comp, "Adjustment_top"); err != nil {
		panic(err)
	}

	rp, err := aep.Reopen(p)
	must(err)

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(rp.WriteAEP(out))
	fmt.Printf("wrote %s (%d layers)\n", outPath, len(rp.Compositions[0].Layers))
}
