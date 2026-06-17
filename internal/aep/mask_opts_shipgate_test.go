// internal/aep/mask_opts_shipgate_test.go
//
// AE ship gate for batch-10 Mask option setters: SetMode / SetInverted /
// SetLocked / SetColor / SetMaskMotionBlur / SetFeather / SetExpansion /
// SetClosed. A from-scratch shape layer gets a mask (AddMask, closed rect), then
// the options are set after Reopen (parse-the-clone surfaces Masks[0]); AE reads
// each back from its Mask DOM and resaves. ae-accept tier.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func runMaskOptsGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mask_opts_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mask_opts.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MSK", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewShapeLayer(comp, "S"); err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	// Reopen to surface the layer as a *Layer, then AddMask on it.
	rp0, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen0: %v", err)
	}
	l0 := rp0.Compositions[0].LayerByName("S")
	if l0 == nil {
		t.Fatal("reopened layer S missing")
	}
	if _, err := aep.AddMask(l0, "M", rectPath()); err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	// Reopen again to parse the new mask so the option setters reach it.
	rp, err := aep.Reopen(rp0)
	if err != nil {
		t.Fatalf("Reopen1: %v", err)
	}
	rl := rp.Compositions[0].LayerByName("S")
	if rl == nil || len(rl.Masks) == 0 {
		t.Fatal("reopened mask missing")
	}
	m := rl.Masks[0]
	must := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	must("SetMode", m.SetMode(aep.MaskModeSubtract))
	must("SetInverted", m.SetInverted(true))
	must("SetColor", m.SetColor([3]uint8{255, 128, 0}))
	must("SetMaskMotionBlur", m.SetMaskMotionBlur(aep.MaskMotionBlurOn))
	must("SetFeather", m.SetFeather([2]float64{10, 20}))
	must("SetExpansion", m.SetExpansion(15))
	must("SetClosed", m.SetClosed(false))
	must("SetLocked", m.SetLocked(true)) // last: locking doesn't block DOM reads

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "msk_in.aep")
	resavedAEP := filepath.Join(tempDir, "msk_resaved.aep")
	doneFile := filepath.Join(tempDir, "msk.done")

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
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
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
	t.Logf("%s mask opts AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s mask opts ship gate FAIL:\n%s", ver, body)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	if rm := re.Compositions[0].LayerByName("S"); rm == nil || len(rm.Masks) == 0 {
		t.Fatalf("%s resaved: mask missing", ver)
	}
}

func TestMaskOpts_AEShipGate_AE2020(t *testing.T) {
	runMaskOptsGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMaskOpts_AEShipGate_AE2025(t *testing.T) {
	runMaskOptsGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
