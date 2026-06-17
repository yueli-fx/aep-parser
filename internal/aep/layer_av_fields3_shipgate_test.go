// internal/aep/layer_av_fields3_shipgate_test.go
//
// AE ship gate (batch 3 of the roundtrip→ae-accept backfill arc): the
// length-variable SetName (rewrites a Utf8 chunk + forces parent-LIST size
// recompute, a different write path from the fixed-width batches 1-2),
// SetStartTime, and the structural SetParent.
//
// SetComment was originally bundled here but FAILED the gate: on a from-scratch
// layer with no native cmta, the new cmta is appended to the Layr LIST tail and
// AE silently ignores it (DOM comment reads back empty) — see
// incidents/layer-setcomment-cmta-append-position.md. It stays at verify=roundtrip.
//
// Two-layer fixture: BG + FG solids. FG is renamed, time-shifted, and parented
// to BG. AE must accept the resized file, read each back through the DOM (incl.
// the parent layer reference), and the values must survive resave.
package aep_test

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func runLayerAVFields3ShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "layer_av_fields3.aep")
	resavedAEP := filepath.Join(tempDir, "layer_av_fields3.resaved.aep")
	doneFile := filepath.Join(tempDir, "layer_av_fields3.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/layer_av_fields3_args.json`

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1280, 720, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "BG", 1280, 720, [3]float64{1, 0, 0}); err != nil {
		t.Fatalf("NewSolidLayer BG: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "FG", 640, 360, [3]float64{0, 0, 1}); err != nil {
		t.Fatalf("NewSolidLayer FG: %v", err)
	}
	p, err = aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	comp = p.CompositionByName("Main")
	if comp == nil || len(comp.Layers) != 2 {
		t.Fatalf("re-parse: want 2 layers, got %d", len(comp.Layers))
	}
	var bg, fg *aep.Layer
	for _, l := range comp.Layers {
		switch l.Name {
		case "BG":
			bg = l
		case "FG":
			fg = l
		}
	}
	if bg == nil || fg == nil {
		t.Fatal("re-parse: BG/FG layer missing")
	}

	must := func(name string, err error) {
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	must("SetName", fg.SetName("FG_Renamed"))
	must("SetStartTime", fg.SetStartTime(0.5))
	must("SetParent", fg.SetParent(bg.ID))

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

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_layer_av_fields3.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Fatalf("layer AV fields3 ship gate FAIL:\n%s", string(content))
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
	if rc == nil || len(rc.Layers) != 2 {
		t.Fatalf("resaved: want 2 layers, got %d", len(rc.Layers))
	}
	var rbg, rfg *aep.Layer
	for _, l := range rc.Layers {
		switch l.Name {
		case "BG":
			rbg = l
		case "FG_Renamed":
			rfg = l
		}
	}
	if rbg == nil {
		t.Fatal("resaved: BG layer missing")
	}
	if rfg == nil {
		t.Fatal("resaved: FG_Renamed layer missing (SetName lost)")
	}
	if math.Abs(rfg.StartTime-0.5) > 0.01 {
		t.Errorf("resaved StartTime = %v, want 0.5", rfg.StartTime)
	}
	if rfg.ParentID != rbg.ID {
		t.Errorf("resaved FG ParentID = %d, want BG ID %d", rfg.ParentID, rbg.ID)
	}
}

func TestLayerAVFields3_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runLayerAVFields3ShipGate(t, aep.TargetAE2025, aeExe)
}

func TestLayerAVFields3_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runLayerAVFields3ShipGate(t, aep.TargetAE2020, aeExe)
}
