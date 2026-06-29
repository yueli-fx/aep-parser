// internal/aep/shape_path_open_shipgate_test.go
//
// AE ship gate: OPEN shape-path acceptance + closed-flag fidelity.
// Gated by AE_SHIP_GATE env var; CI / no-AE runs SKIP cleanly.
//
// Every prior shape-path gate builds Closed=true. This gate builds an OPEN path
// (SetClosed(false)) to close the long-standing coverage hole: encodeBezier
// writes shph[3]=0x00 for open paths (AE-native uses 0x09) plus n-dependent
// lhd3 fields. We verify (a) AE accepts our open-path bytes (no drop/crash) and
// (b) the open-ness survives an AE resave — Go re-decodes shph[3]. A reject or a
// closed=true readback would prove a real functional bug, not just a byte gap.
// See incidents/add-mask-create-re.md §finding-4 follow-up.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

func runV2_2PathOpenShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_path_open.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_path_open.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_path_open.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/generated/args/v2_2_path_open_args.json`

	verts := [][2]float64{{100, 100}, {300, 200}, {200, 400}} // open polyline, n=3

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "Path_Open")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	pth, err := l.RootGroup().AddPath()
	if err != nil {
		t.Fatalf("AddPath: %v", err)
	}
	if err := pth.SetVertices(verts); err != nil {
		t.Fatalf("SetVertices: %v", err)
	}
	if err := pth.SetClosed(false); err != nil {
		t.Fatalf("SetClosed(false): %v", err)
	}
	fill, err := l.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill: %v", err)
	}
	_ = fill.SetColor([4]float64{0.5, 0.5, 0.5, 1})

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

	jsxPath := `E:/projects/tools/aep-parser/test_data/generators/verify_v2_2_path_open.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("open path ship gate FAIL:\n%s", string(content))
	}

	assertResavedPathAnchors(t, resavedAEP, verts)
	assertResavedPathClosed(t, resavedAEP, false)
}

// assertResavedPathClosed decodes the first shph in an AE-resaved .aep and
// asserts its closed flag (shph[3]==0x01 closed / 0x09 open) matches want —
// proving AE preserved our open/closed intent across a save round-trip.
func assertResavedPathClosed(t *testing.T, path string, wantClosed bool) {
	t.Helper()
	root := parseAEP(t, path)
	shph := findShipChunk(root, rifx.IDShph)
	if shph == nil || len(shph.Data) < 4 {
		t.Fatalf("resaved: no shph (path dropped)")
	}
	got := shph.Data[3] == 0x01
	if got != wantClosed {
		t.Errorf("resaved shph[3]=0x%02x → closed=%v, want %v (AE re-encoded the open/closed flag)", shph.Data[3], got, wantClosed)
	}
}

func TestV2_2_PathOpen_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2PathOpenShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_PathOpen_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2PathOpenShipGate(t, aep.TargetAE2020, aeExe)
}
