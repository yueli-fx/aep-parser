// internal/aep/project_settings_shipgate_test.go
//
// Automated AE ship gate for the from-scratch-writable PROJECT-level settings
// (Project.Set*). These were proven AE2020+AE2025 DOM-readback-correct by the
// project-settings showcase (flightdeck/showcase/project-settings, RE-fixed
// 2026-06-14 for the nnhd→legacy-nhed display group) but had no automated Go
// _AEShipGate test, so the caps were capped at verify=roundtrip — a
// verification-hygiene hole. This gate sets a NON-DEFAULT value for each and
// verifies AE's app.project DOM readback, double-version.
//
//   - SetBitsPerChannel(BPC16)                          → bitsPerChannel == 16
//   - SetTimeDisplayType(Frames)                        → timeDisplayType == FRAMES
//   - SetFramesCountType(Start1)                        → framesCountType == FC_START_1 (+displayStartFrame==1)
//   - SetFramesUseFeetFrames(true)                      → framesUseFeetFrames == true
//   - SetFeetFramesFilmType(MM16)                       → feetFramesFilmType == MM16
//   - SetFootageTimecodeDisplayStartType(UseSourceMedia)→ FTCS_USE_SOURCE_MEDIA
//   - SetLinearBlending(true)                           → linearBlending == true
//   - SetExpressionEngine("javascript-1.0")             → expressionEngine == "javascript-1.0"
//
// Carrier: a from-scratch project with one shape layer (project settings have
// no rendered visual — DOM readback is their surface). Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func buildProjectSettingsDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)

	// A project needs at least one comp/layer to be a valid .aep AE opens cleanly.
	comp, err := aep.NewComposition(p, "PS", 640, 360, 30, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	bg, err := aep.NewShapeLayer(comp, "BG")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	r, err := bg.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := r.SetSize([2]float64{200, 200}); err != nil {
		t.Fatalf("SetSize: %v", err)
	}

	// Apply each setter fatally — they are proven to succeed from-scratch
	// (showcase), so any error here is a real regression worth failing on.
	must := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	must("SetBitsPerChannel", p.SetBitsPerChannel(aep.BPC16))
	must("SetTimeDisplayType", p.SetTimeDisplayType(aep.TimeDisplayTypeFrames))
	must("SetFramesCountType", p.SetFramesCountType(aep.FramesCountTypeStart1))
	must("SetFramesUseFeetFrames", p.SetFramesUseFeetFrames(true))
	must("SetFeetFramesFilmType", p.SetFeetFramesFilmType(aep.FeetFramesFilmTypeMM16))
	must("SetFootageTimecodeDisplayStartType", p.SetFootageTimecodeDisplayStartType(aep.FootageTimecodeDisplayStartTypeUseSourceMedia))
	must("SetLinearBlending", p.SetLinearBlending(true))
	must("SetExpressionEngine", p.SetExpressionEngine("javascript-1.0"))

	return p
}

func runProjectSettingsGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/project_settings_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_project_settings.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildProjectSettingsDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "project_settings_in.aep")
	resavedAEP := filepath.Join(tempDir, "project_settings_resaved.aep")
	doneFile := filepath.Join(tempDir, "project_settings.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 240)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("project_settings %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("project_settings %s ship gate FAIL:\n%s", ver, body)
	}

	// Resave proof: settings survive AE's re-encode and the Go parser reads them
	// back from the resaved file.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.BitsPerChannel != aep.BPC16 {
		t.Errorf("resaved BitsPerChannel = %v, want BPC16", re.BitsPerChannel)
	}
	if re.TimeDisplayType() != aep.TimeDisplayTypeFrames {
		t.Errorf("resaved TimeDisplayType = %v, want Frames", re.TimeDisplayType())
	}
	if re.FramesCountType() != aep.FramesCountTypeStart1 {
		t.Errorf("resaved FramesCountType = %v, want Start1", re.FramesCountType())
	}
	if !re.FramesUseFeetFrames() {
		t.Errorf("resaved FramesUseFeetFrames = false, want true")
	}
	if re.FeetFramesFilmType() != aep.FeetFramesFilmTypeMM16 {
		t.Errorf("resaved FeetFramesFilmType = %v, want MM16", re.FeetFramesFilmType())
	}
	if re.FootageTimecodeDisplayStartType() != aep.FootageTimecodeDisplayStartTypeUseSourceMedia {
		t.Errorf("resaved FootageTimecodeDisplayStartType = %v, want UseSourceMedia", re.FootageTimecodeDisplayStartType())
	}
	if !re.LinearBlending() {
		t.Errorf("resaved LinearBlending = false, want true")
	}
	if re.ExpressionEngine() != "javascript-1.0" {
		t.Errorf("resaved ExpressionEngine = %q, want javascript-1.0", re.ExpressionEngine())
	}
}

func TestProjectSettings_AEShipGate_AE2020(t *testing.T) {
	runProjectSettingsGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestProjectSettings_AEShipGate_AE2025(t *testing.T) {
	runProjectSettingsGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
