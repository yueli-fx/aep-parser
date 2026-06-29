// internal/aep/layer_av_fields2_shipgate_test.go
//
// AE ship gate (batch 2 of the roundtrip→ae-accept backfill arc): more
// AV-layer setters that until now only had Go round-trip coverage —
// SetInPoint / SetOutPoint (timeline trim, seconds) / SetPreserveTransparency /
// SetSamplingBicubic / SetIsGuide / SetIsAdjust / SetLabel.
//
// Flow: all-Go solid project → Reopen (parsed layer) → set 7 fields → AE must
// accept the file, read each back through the scripting API (DOM), and the
// values must survive AE's own resave (Go re-decodes getters/struct fields).
package aep_test

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runLayerAVFields2ShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "layer_av_fields2.aep")
	resavedAEP := filepath.Join(tempDir, "layer_av_fields2.resaved.aep")
	doneFile := filepath.Join(tempDir, "layer_av_fields2.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/layer_av_fields2_args.json`

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1280, 720, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "BG", 1280, 720, [3]float64{1, 0, 0}); err != nil {
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
	L := comp.Layers[0]

	must := func(name string, err error) {
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	must("SetInPoint", L.SetInPoint(1.0))
	must("SetOutPoint", L.SetOutPoint(3.0))
	must("SetPreserveTransparency", L.SetPreserveTransparency(true))
	must("SetSamplingBicubic", L.SetSamplingBicubic(true))
	must("SetIsGuide", L.SetIsGuide(true))
	must("SetIsAdjust", L.SetIsAdjust(true))
	must("SetLabel", L.SetLabel(9))

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

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_layer_av_fields2.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Fatalf("layer AV fields2 ship gate FAIL:\n%s", string(content))
	}

	raw, err := os.ReadFile(resavedAEP)
	if err != nil {
		t.Fatalf("read resaved: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("parse resaved: %v", err)
	}
	rc := re.CompositionByName("Main")
	if rc == nil || len(rc.Layers) == 0 {
		t.Fatal("resaved: comp/layer missing")
	}
	rl := rc.Layers[0]
	if math.Abs(rl.InPoint()-1.0) > 0.01 {
		t.Errorf("resaved InPoint = %v, want 1.0", rl.InPoint())
	}
	if math.Abs(rl.OutPoint()-3.0) > 0.01 {
		t.Errorf("resaved OutPoint = %v, want 3.0", rl.OutPoint())
	}
	if !rl.PreserveTransparency {
		t.Errorf("resaved PreserveTransparency = false, want true")
	}
	if !rl.SamplingBicubic {
		t.Errorf("resaved SamplingBicubic = false, want true")
	}
	if !rl.IsGuide {
		t.Errorf("resaved IsGuide = false, want true")
	}
	if !rl.IsAdjust {
		t.Errorf("resaved IsAdjust = false, want true")
	}
	if rl.Label != 9 {
		t.Errorf("resaved Label = %d, want 9", rl.Label)
	}
}

func TestLayerAVFields2_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runLayerAVFields2ShipGate(t, aep.TargetAE2025, aeExe)
}

func TestLayerAVFields2_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runLayerAVFields2ShipGate(t, aep.TargetAE2020, aeExe)
}
