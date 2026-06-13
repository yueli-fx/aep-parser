// flightdeck/showcase/comp-settings/gen.go — from-scratch (no AE) showcase of
// COMPOSITION SETTINGS setters. Comp settings have no rendered visual, so this is
// a 📋 readback showcase: gen.go applies a distinctive set of Composition.Set*
// values and verify.jsx dumps the comp DOM back to the .done log for the user to
// read-check (not a rendered png).
//
// Writes comp_settings.aep next to this file.
// Run from repo root: `go run ./flightdeck/showcase/comp-settings`.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/comp-settings/comp_settings.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "CompSettings", 1920, 1080, 30, 10)
	must(err)

	// A trivial visible layer so the comp is non-empty.
	bg, err := aep.NewShapeLayer(comp, "BG")
	must(err)
	r, err := bg.RootGroup().AddRect()
	must(err)
	must(r.SetSize([2]float64{400, 400}))
	f, err := bg.RootGroup().AddFill()
	must(err)
	must(f.SetColor([4]float64{0.25, 0.6, 0.95, 1}))
	must(bg.Position().SetStaticValue([2]float64{960, 540}))

	// Distinctive, non-default settings — only setters whose AE-DOM readback is
	// EXACT are shown here. (SetShutterAngle/SetShutterPhase read back scaled
	// ~1.2× and SetResolutionFactor makes AE's resolutionFactor DOM throw a
	// divide-by-zero — both excluded as found boundaries; see INDEX.md.)
	must(comp.SetCompMotionBlur(true))               // DOM motionBlur = true
	must(comp.SetMotionBlurSamplesPerFrame(24))      // DOM motionBlurSamplesPerFrame
	must(comp.SetMotionBlurAdaptiveSampleLimit(192)) // DOM motionBlurAdaptiveSampleLimit
	must(comp.SetBGColor([3]uint8{40, 20, 80}))      // DOM bgColor (~[0.157,0.078,0.314])
	must(comp.SetWorkArea(1.0, 6.0))                 // DOM workAreaStart=1, duration=5
	must(comp.SetHideShyLayers(true))                // DOM hideShyLayers = true
	must(comp.SetPreserveNestedFrameRate(true))      // DOM preserveNestedFrameRate = true

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(p.WriteAEP(out))
	fmt.Printf("wrote %s\n", outPath)
}
