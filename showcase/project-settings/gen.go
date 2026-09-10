// showcase/project-settings/gen.go — from-scratch (no AE) showcase of
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
// Run from repo root: `go run ./showcase/project-settings`.
package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const outPath = "showcase/project-settings/project_settings.aep"

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

	// All settings below have AE-DOM readback confirmed (AE 2020 + 2025). The
	// nnhd display group (time/frames-count/feet/use-feet) was a false-green
	// boundary until 2026-06-14: AE reads them from the legacy nhed header, not
	// nnhd, so the setters now mirror BOTH (incidents/nnhd-display-settings-layout-re.md).
	// Values chosen NON-default so a wrong readback shows as a real mismatch.
	fmt.Println("applying project settings:")
	try("SetBitsPerChannel(16)", p.SetBitsPerChannel(aep.BPC16))
	try("SetLinearBlending(true)", p.SetLinearBlending(true))
	try("SetExpressionEngine(javascript-1.0)", p.SetExpressionEngine("javascript-1.0"))
	try("SetFootageTimecodeDisplayStartType(UseSourceMedia)", p.SetFootageTimecodeDisplayStartType(aep.FootageTimecodeDisplayStartTypeUseSourceMedia))
	try("SetTimeDisplayType(Timecode)", p.SetTimeDisplayType(aep.TimeDisplayTypeTimecode))
	try("SetFramesCountType(Start0)", p.SetFramesCountType(aep.FramesCountTypeStart0))
	try("SetFramesUseFeetFrames(true)", p.SetFramesUseFeetFrames(true))
	try("SetFeetFramesFilmType(MM35)", p.SetFeetFramesFilmType(aep.FeetFramesFilmTypeMM35))

	out, err := os.Create(outPath)
	must(err)
	defer out.Close()
	must(p.WriteAEP(out))
	fmt.Printf("wrote %s\n", outPath)
}
