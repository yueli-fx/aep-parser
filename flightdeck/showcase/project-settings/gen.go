// flightdeck/showcase/project-settings/gen.go — from-scratch (no AE) showcase of
// PROJECT SETTINGS setters (Project.Set*). Project settings have no rendered
// visual → 📋 readback showcase: gen.go applies a distinctive set and verify.jsx
// dumps app.project DOM back to .done for the user to read-check (not a png).
//
// Setters are applied NON-FATALLY: a from-scratch project may lack the backing
// chunk for some settings (ExEn / nnhd / lnrb …), so each call is logged as
// applied / skipped rather than panicking. The INDEX table lists only the ones
// whose AE-DOM readback was confirmed.
//
// Writes project_settings.aep next to this file.
// Run from repo root: `go run ./flightdeck/showcase/project-settings`.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outPath = "flightdeck/showcase/project-settings/project_settings.aep"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// try applies a setter non-fatally, recording the outcome.
func try(name string, err error) {
	if err != nil {
		fmt.Printf("  SKIP %s: %v\n", name, err)
		return
	}
	fmt.Printf("  ok   %s\n", name)
}

func main() {
	p := aep.NewProject(aep.TargetAE2020)
	// A project needs at least one comp to be a valid .aep.
	comp, err := aep.NewComposition(p, "ProjSettings", 1920, 1080, 30, 5)
	must(err)
	bg, err := aep.NewShapeLayer(comp, "BG")
	must(err)
	r, err := bg.RootGroup().AddRect()
	must(err)
	must(r.SetSize([2]float64{400, 400}))
	f, err := bg.RootGroup().AddFill()
	must(err)
	must(f.SetColor([4]float64{0.25, 0.6, 0.95, 1}))
	must(bg.Position().SetStaticValue([2]float64{960, 540}))

	// Only setters whose AE-DOM readback is EXACT are applied here. Excluded as
	// found boundaries (Go byte round-trip green but AE DOM did NOT reflect them,
	// red line 4a): SetTimeDisplayType / SetFeetFramesFilmType (both nnhd byte 8 —
	// suspected bit-packing / setter-interaction) and SetFramesCountType. See
	// INDEX.md. Note: SetFootageTimecodeDisplayStartType (nnhd byte 9) DID take.
	fmt.Println("applying project settings:")
	try("SetBitsPerChannel(16)", p.SetBitsPerChannel(aep.BPC16))
	try("SetLinearBlending(true)", p.SetLinearBlending(true))
	try("SetExpressionEngine(javascript-1.0)", p.SetExpressionEngine("javascript-1.0"))
	try("SetFootageTimecodeDisplayStartType(UseSourceMedia)", p.SetFootageTimecodeDisplayStartType(aep.FootageTimecodeDisplayStartTypeUseSourceMedia))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(p.WriteAEP(out))
	fmt.Printf("wrote %s\n", outPath)
}
