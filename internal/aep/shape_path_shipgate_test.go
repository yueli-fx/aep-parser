// internal/aep/shape_path_shipgate_test.go
//
// AE ship gate: Path + Fill focused gate.
// Gated by AE_SHIP_GATE env var; CI / no-AE runs SKIP cleanly.
package aep_test

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

// runV2_2PathShipGate is the V2.2.1 focused Path gate: one ShapeLayer with a
// Path (DISTINCT triangle, not the embed's square, so the geometry splice is
// actually exercised) + Fill. Asserts AE accepts it (from-scratch path CRASHED
// AE 2020) and the re-saved ldat anchors round-trip to the input vertices.
func runV2_2PathShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_path.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_path.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_path.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_path_args.json`

	verts := [][2]float64{{10, 20}, {70, 30}, {40, 90}}

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "Path_Static")
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
	_ = pth.SetClosed(true)
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
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2_path.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("path ship gate FAIL:\n%s", string(content))
	}

	assertResavedPathAnchors(t, resavedAEP, verts)
}

// assertResavedPathAnchors parses an AE-resaved .aep, decodes the path shph
// bbox + ldat vertex anchors (slot 0 of each 6-f32 vertex block, bbox-
// de-normalized), and asserts they match the input vertices — proving AE
// retained the geometry the splice injected.
func assertResavedPathAnchors(t *testing.T, path string, want [][2]float64) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open resaved: %v", err)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		t.Fatalf("parse resaved: %v", err)
	}
	shph := findShipChunk(root, rifx.IDShph)
	ldat := findShipChunk(root, rifx.IDLdat)
	if shph == nil || ldat == nil {
		t.Fatalf("resaved path: shph=%v ldat=%v (one missing → path dropped)", shph != nil, ldat != nil)
	}
	rf := func(b []byte, off int) float64 {
		return float64(math.Float32frombits(binary.BigEndian.Uint32(b[off : off+4])))
	}
	minX, minY := rf(shph.Data, 4), rf(shph.Data, 8)
	maxX, maxY := rf(shph.Data, 12), rf(shph.Data, 16)
	rx, ry := maxX-minX, maxY-minY
	if len(ldat.Data) < len(want)*24 {
		t.Fatalf("resaved ldat too short: %d B for %d verts", len(ldat.Data), len(want))
	}
	for i, w := range want {
		ax := rf(ldat.Data, i*24)*rx + minX
		ay := rf(ldat.Data, i*24+4)*ry + minY
		if math.Abs(ax-w[0]) > 0.5 || math.Abs(ay-w[1]) > 0.5 {
			t.Errorf("resaved vertex %d anchor = (%.3g,%.3g), want (%.3g,%.3g)", i, ax, ay, w[0], w[1])
		}
	}
}

func TestV2_2_Path_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2PathShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_Path_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2PathShipGate(t, aep.TargetAE2020, aeExe)
}
