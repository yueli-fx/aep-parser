// internal/aep/solid_setters_shipgate_test.go
//
// AE ship gate: standalone Footage.SetSolidColor / SetSolidSize on a PARSED
// solid (not creation params — those ride TestNewSolidNull's gate).
// Gated by AE_SHIP_GATE env var; CI / no-AE runs SKIP cleanly.
//
// Flow: all-Go-built solid project → Reopen (parsed solid) → SetSolidColor +
// SetSolidSize → AE must accept the file, read the NEW color/dims through the
// scripting API (SolidSource.color, FootageItem.width/height), and the values
// must survive AE's own resave (Go re-decodes).
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

func runSolidSettersShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "solid_setters.aep")
	resavedAEP := filepath.Join(tempDir, "solid_setters.resaved.aep")
	doneFile := filepath.Join(tempDir, "solid_setters.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/solid_setters_args.json`

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
	var solid *aep.Footage
	for _, f := range p.Footage {
		if f.ID == comp.Layers[0].SourceID {
			solid = f
		}
	}
	if solid == nil {
		t.Fatal("re-parse: backing solid missing")
	}
	if err := solid.SetSolidColor([3]float64{0.1, 0.6, 0.9}); err != nil {
		t.Fatalf("SetSolidColor: %v", err)
	}
	if err := solid.SetSolidSize(640, 360); err != nil {
		t.Fatalf("SetSolidSize: %v", err)
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

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_solid_setters.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Fatalf("solid setters ship gate FAIL:\n%s", string(content))
	}

	raw, err := os.ReadFile(resavedAEP)
	if err != nil {
		t.Fatalf("read resaved: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("parse resaved: %v", err)
	}
	var rf *aep.Footage
	for _, f := range re.Footage {
		if f.IsSolid {
			rf = f
		}
	}
	if rf == nil {
		t.Fatal("resaved: solid footage missing")
	}
	for i, want := range [3]float64{0.1, 0.6, 0.9} {
		got := rf.SolidColor[i]
		if got < want-0.005 || got > want+0.005 {
			t.Errorf("resaved SolidColor[%d] = %v, want %v", i, got, want)
		}
	}
	if rf.Width != 640 || rf.Height != 360 {
		t.Errorf("resaved dims = %dx%d, want 640x360", rf.Width, rf.Height)
	}
}

func TestSolidSetters_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runSolidSettersShipGate(t, aep.TargetAE2025, aeExe)
}

func TestSolidSetters_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runSolidSettersShipGate(t, aep.TargetAE2020, aeExe)
}
