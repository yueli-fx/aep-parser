// internal/aep/layer_flags2_shipgate_test.go
//
// AE ship gate for layer-set flags that need a real carrier (verify=roundtrip):
//
//   - SetEffectsEnabled   → AE effectsActive (read-only = switch ON && ≥1 enabled
//                           effect). EON: effect + switch on → true; EOFF: effect
//                           + SetEffectsEnabled(false) → false (contrast).
//   - SetIsNull           → AE nullLayer (flip a solid's null bit).
//
// SetTimeRemapEnabled is NOT here — confirmed false-green: it sets a static value
// 0.0, but AE (and our own TimeRemapEnabled getter) represent enabled time-remap
// as the property carrying 2 identity keyframes, so AE reads timeRemapEnabled=
// false. See incident layer-settimeremapenabled-needs-keyframes.
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

func buildLayerFlags2Demo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	main, err := aep.NewComposition(p, "MAIN", 640, 360, 30, 4)
	if err != nil {
		t.Fatalf("NewComposition MAIN: %v", err)
	}
	if _, err := aep.NewSolidLayer(main, "EON", 200, 200, [3]float64{0.9, 0.2, 0.2}); err != nil {
		t.Fatalf("NewSolidLayer EON: %v", err)
	}
	if _, err := aep.NewSolidLayer(main, "EOFF", 200, 200, [3]float64{0.2, 0.8, 0.3}); err != nil {
		t.Fatalf("NewSolidLayer EOFF: %v", err)
	}
	if _, err := aep.NewSolidLayer(main, "NUL", 100, 100, [3]float64{0.2, 0.4, 0.9}); err != nil {
		t.Fatalf("NewSolidLayer NUL: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	cc := rp.CompositionByName("MAIN")
	if cc == nil {
		t.Fatal("reopened MAIN missing")
	}
	must := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	// Effects switch: both get an effect; only EOFF turns the switch off.
	if _, err := aep.AddEffect(cc.LayerByName("EON"), aep.EffectGaussianBlur); err != nil {
		t.Fatalf("AddEffect EON: %v", err)
	}
	if _, err := aep.AddEffect(cc.LayerByName("EOFF"), aep.EffectGaussianBlur); err != nil {
		t.Fatalf("AddEffect EOFF: %v", err)
	}
	must("SetEffectsEnabled", cc.LayerByName("EOFF").SetEffectsEnabled(false))
	must("SetIsNull", cc.LayerByName("NUL").SetIsNull(true))

	return rp
}

func runLayerFlags2Gate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/layer_flags2_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_layer_flags2.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildLayerFlags2Demo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "layer_flags2_in.aep")
	resavedAEP := filepath.Join(tempDir, "layer_flags2_resaved.aep")
	doneFile := filepath.Join(tempDir, "layer_flags2.done")

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
	t.Logf("layer flags2 %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("layer flags2 %s ship gate FAIL:\n%s", ver, body)
	}
}

func TestLayerFlags2_AEShipGate_AE2020(t *testing.T) {
	runLayerFlags2Gate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestLayerFlags2_AEShipGate_AE2025(t *testing.T) {
	runLayerFlags2Gate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
