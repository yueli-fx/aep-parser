// flightdeck/showcase/essential-graphics/gen.go — from-scratch (no AE) showcase of
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
// Run from repo root: `go run ./flightdeck/showcase/essential-graphics`.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/essential-graphics/essential_graphics.aep"

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

	// Promote one param from each of three control effects to an EG controller.
	// A color param is elided by default → materialize it with SetEffectParam
	// before exposing (the elision rule); slider/checkbox expose directly.
	controllers := []struct {
		effect      string
		paramMN     string
		eName       string
		materialize any // non-nil → SetEffectParam first to create the value stream
	}{
		{aep.EffectSliderControl, "ADBE Slider Control-0001", "Blur Amount", nil},
		{aep.EffectColorControl, "ADBE Color Control-0001", "Accent Color", []float64{1, 0.4, 0.2, 1}},
		{aep.EffectCheckboxControl, "ADBE Checkbox Control-0001", "Enable Glow", nil},
	}
	added := 0
	for _, c := range controllers {
		fx, err := aep.AddEffect(layer, c.effect)
		if err != nil {
			fmt.Printf("  SKIP AddEffect(%s): %v\n", c.effect, err)
			continue
		}
		if c.materialize != nil {
			if _, err := aep.SetEffectParam(layer, fx, c.paramMN, c.materialize); err != nil {
				fmt.Printf("  SKIP SetEffectParam(%q): %v\n", c.eName, err)
				continue
			}
		}
		if _, err := aep.AddEssentialProperty(layer, fx, c.paramMN, c.eName); err != nil {
			fmt.Printf("  SKIP AddEssentialProperty(%q): %v\n", c.eName, err)
			continue
		}
		fmt.Printf("  ok   controller %q\n", c.eName)
		added++
	}
	must(comp.SetMotionGraphicsTemplateName("Showcase EG Template"))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(p.WriteAEP(out))
	fmt.Printf("wrote %s (%d controllers)\n", outPath, added)
}
