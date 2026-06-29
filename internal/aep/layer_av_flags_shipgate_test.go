// internal/aep/layer_av_flags_shipgate_test.go
//
// AE ship gate (batch 1 of the roundtrip→ae-accept backfill arc): the common
// AV-layer flag/enum setters that until now only had Go round-trip coverage —
// SetVisible / SetShy / SetSolo / SetLocked / SetMotionBlur / SetQuality /
// SetBlendingMode. Several share ldta @0x27 (single-byte bit field), so this
// also proves the multiple flags don't clobber each other through one AE trip.
//
// Flow: all-Go solid project → Reopen (parsed layer) → set 7 flags → AE must
// accept the file, read each flag/enum back through the scripting API (DOM),
// and the values must survive AE's own resave (Go re-decodes the struct fields).
package aep_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runLayerAVFlagsShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "layer_av_flags.aep")
	resavedAEP := filepath.Join(tempDir, "layer_av_flags.resaved.aep")
	doneFile := filepath.Join(tempDir, "layer_av_flags.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/generated/args/layer_av_flags_args.json`

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
	must("SetVisible", L.SetVisible(false))
	must("SetShy", L.SetShy(true))
	must("SetSolo", L.SetSolo(true))
	must("SetLocked", L.SetLocked(true))
	must("SetMotionBlur", L.SetMotionBlur(true))
	must("SetQuality", L.SetQuality(aep.LayerQualityDraft))
	must("SetBlendingMode", L.SetBlendingMode(aep.BlendingModeMultiply))

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
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	jsxPath := `E:/projects/tools/aep-parser/test_data/generators/verify_layer_av_flags.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Fatalf("layer AV flags ship gate FAIL:\n%s", string(content))
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
	if rl.Visible {
		t.Errorf("resaved Visible = true, want false")
	}
	if !rl.Shy {
		t.Errorf("resaved Shy = false, want true")
	}
	if !rl.Solo {
		t.Errorf("resaved Solo = false, want true")
	}
	if !rl.Locked {
		t.Errorf("resaved Locked = false, want true")
	}
	if !rl.MotionBlur {
		t.Errorf("resaved MotionBlur = false, want true")
	}
	if rl.Quality != aep.LayerQualityDraft {
		t.Errorf("resaved Quality = %v, want Draft", rl.Quality)
	}
	if rl.BlendingMode != aep.BlendingModeMultiply {
		t.Errorf("resaved BlendingMode = %v, want Multiply", rl.BlendingMode)
	}
}

func TestLayerAVFlags_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runLayerAVFlagsShipGate(t, aep.TargetAE2025, aeExe)
}

func TestLayerAVFlags_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runLayerAVFlagsShipGate(t, aep.TargetAE2020, aeExe)
}
