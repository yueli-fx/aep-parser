// internal/aep/essential_graphics_shipgate_test.go
//
// AE ship gate: Essential Graphics W acceptance + readback.
// Gated by AE_SHIP_GATE env var; CI / no-AE runs SKIP cleanly.
//
// All-Go-built variant (no fixture-derived IDs — gate-fixture ID-coincidence
// lesson): NewProject + NewComposition + NewSolidLayer + Reopen +
// AddEffect(Slider Control) + AddEssentialProperty + SetMotionGraphicsTemplateName.
// AE must accept the file, read the EG panel back through the scripting API
// (template name + controller count + controller name), and the controller
// must survive AE's own resave (Go re-decodes the resaved CIF3).
package aep_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func runEGAddShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "eg_add.aep")
	resavedAEP := filepath.Join(tempDir, "eg_add.resaved.aep")
	doneFile := filepath.Join(tempDir, "eg_add.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/eg_add_args.json`

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "host", 1920, 1080, [3]float64{0.2, 0.4, 0.8}); err != nil {
		t.Fatalf("NewSolidLayer: %v", err)
	}
	p, err = aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	comp = p.CompositionByName("Main")
	if comp == nil || len(comp.Layers) == 0 {
		t.Fatal("re-parse: comp/layer missing")
	}
	layer := comp.Layers[0]
	fx, err := aep.AddEffect(layer, aep.EffectSliderControl)
	if err != nil {
		t.Fatalf("AddEffect: %v", err)
	}
	ctrl, err := aep.AddEssentialProperty(layer, fx, "ADBE Slider Control-0001", "Exposed Slider")
	if err != nil {
		t.Fatalf("AddEssentialProperty: %v", err)
	}
	if err := comp.SetMotionGraphicsTemplateName("EG Gate Template"); err != nil {
		t.Fatalf("SetMotionGraphicsTemplateName: %v", err)
	}

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_eg_add.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Fatalf("EG add ship gate FAIL:\n%s", string(content))
	}

	// AE resave round-trip: the controller must still decode Go-side, with
	// the same identity AE was given.
	raw, err := os.ReadFile(resavedAEP)
	if err != nil {
		t.Fatalf("read resaved: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("parse resaved: %v", err)
	}
	rc := re.CompositionByName("Main")
	if rc == nil {
		t.Fatal("resaved: comp Main missing")
	}
	if rc.MotionGraphicsTemplateName != "EG Gate Template" {
		t.Errorf("resaved template name = %q, want EG Gate Template", rc.MotionGraphicsTemplateName)
	}
	if got := rc.MotionGraphicsTemplateControllerCount(); got != 1 {
		t.Fatalf("resaved ControllerCount = %d, want 1", got)
	}
	g := rc.EssentialGraphicsControllers[0]
	if g.Name != "Exposed Slider" || g.Type != aep.EGSlider {
		t.Errorf("resaved controller = {%q %v}, want {Exposed Slider slider}", g.Name, g.Type)
	}
	if g.UUID != ctrl.UUID {
		t.Errorf("resaved controller UUID = %s, want %s (AE re-keyed the controller)", g.UUID, ctrl.UUID)
	}
}

func TestEGAdd_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runEGAddShipGate(t, aep.TargetAE2025, aeExe)
}

func TestEGAdd_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runEGAddShipGate(t, aep.TargetAE2020, aeExe)
}
