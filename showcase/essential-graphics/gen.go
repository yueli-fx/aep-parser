// showcase/essential-graphics/gen.go — from-scratch (no AE) showcase of
// Essential Graphics: AddEssentialProperty + SetMotionGraphicsTemplateName. EG
// controllers live in the EG panel (no rendered visual) → 📋 readback showcase:
// gen.go promotes three effect params to EG controllers and names the template;
// verify.jsx dumps the comp's EG DOM (template name + controller count + names)
// to .done for the user to read-check (not a png).
//
// Mirrors the proven EG ship-gate path (NewSolidLayer + Reopen + AddEffect +
// AddEssentialProperty + SetMotionGraphicsTemplateName); each controller is added
// NON-FATALLY so a wrong param match-name is reported rather than panicking.
//
// Writes essential_graphics.aep next to this file.
// Run from repo root: `go run ./showcase/essential-graphics`.
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const outPath = "showcase/essential-graphics/essential_graphics.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "EssentialGraphics", 1920, 1080, 30, 5)
	must(err)
	if _, err := aep.NewSolidLayer(comp, "host", 1920, 1080, [3]float64{0.2, 0.4, 0.8}); err != nil {
		must(err)
	}
	// Reopen so the host layer is parsed (EG splice needs a parsed host).
	p, err = aep.Reopen(p)
	must(err)
	comp = p.CompositionByName("EssentialGraphics")
	if comp == nil || len(comp.Layers) == 0 {
		panic("re-parse: comp/layer missing")
	}
	layer := comp.Layers[0]

	// EXACTLY the ship-gate-proven case: ONE Slider Control exposed as ONE EG
	// controller. (An earlier 3-controller version — slider + materialized color
	// + checkbox — loaded and DOM-read fine but CRASHED AE when the user expanded
	// the Essential Graphics panel: that combination is beyond the gate's coverage
	// [gate only ever did 1 slider, and never opened the panel]. Reduced here to
	// the single proven controller; the multi-/color-/checkbox-controller panel
	// crash is a found boundary — see INDEX.md + cockpit RE candidates.)
	fx, err := aep.AddEffect(layer, aep.EffectSliderControl)
	must(err)
	_, err = aep.AddEssentialProperty(layer, fx, "ADBE Slider Control-0001", "Blur Amount")
	must(err)
	fmt.Println("  ok   controller \"Blur Amount\" (single slider, gate-proven)")
	added := 1
	must(comp.SetMotionGraphicsTemplateName("Showcase EG Template"))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(p.WriteAEP(out))
	fmt.Printf("wrote %s (%d controllers)\n", outPath, added)
}
