// internal/aep/text_run_shipgate_test.go
//
// AE ship gate for batch-7 single-run text SetRun* style setters. A from-scratch
// NewTextLayer + SetText("Ag") makes a SINGLE-run text doc; each SetRun0 attribute
// therefore surfaces as a whole-document textDocument property that AE reads back
// on BOTH AE2020 and AE2025 (no characterRange — that's AE2022+ only). Style is
// applied after Reopen (parse-the-clone surfaces Runs[0]), mirroring
// mg_text_style_shipgate_test.go. DOM-value verification → ae-accept (not render).
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runTextRunGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/text_run_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_text_run.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "TXT", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	tl, err := aep.NewTextLayer(comp, "T")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if err := tl.SetText("Ag"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	rl := rp.Compositions[0].LayerByName("T")
	if rl == nil {
		t.Fatal("reopened text layer missing")
	}
	must := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	must("SetRunFillColor", rl.SetRunFillColor(0, [4]float64{0.2, 0.4, 0.8, 1}))
	must("SetRunApplyStroke", rl.SetRunApplyStroke(0, true))
	must("SetRunStrokeColor", rl.SetRunStrokeColor(0, [4]float64{0.9, 0.1, 0.3, 1}))
	must("SetRunStrokeWidth", rl.SetRunStrokeWidth(0, 5))
	must("SetRunStrokeOverFill", rl.SetRunStrokeOverFill(0, false))
	must("SetRunFauxBold", rl.SetRunFauxBold(0, true))
	must("SetRunFauxItalic", rl.SetRunFauxItalic(0, true))
	must("SetRunBaselineShift", rl.SetRunBaselineShift(0, 8))
	must("SetRunAutoLeading", rl.SetRunAutoLeading(0, false))
	must("SetRunLeading", rl.SetRunLeading(0, 80))
	// Uncertain on AE2020 (logged by verify JSX; trimmed if not reflected).
	must("SetRunHorizontalScale", rl.SetRunHorizontalScale(0, 50))
	must("SetRunVerticalScale", rl.SetRunVerticalScale(0, 150))
	must("SetRunTsume", rl.SetRunTsume(0, 0.5)) // AE tsume DOM is 0..1

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "text_in.aep")
	resavedAEP := filepath.Join(tempDir, "text_resaved.aep")
	doneFile := filepath.Join(tempDir, "text.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
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

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("%s text run AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s text run ship gate FAIL:\n%s", ver, body)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	if re.Compositions[0].LayerByName("T") == nil {
		t.Fatalf("%s resaved: text layer T missing", ver)
	}
}

func TestTextRun_AEShipGate_AE2020(t *testing.T) {
	runTextRunGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestTextRun_AEShipGate_AE2025(t *testing.T) {
	runTextRunGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
