// internal/aep/shape_rectkf_shipgate_test.go
//
// V2.2 Phase 5 Task 5.2 — AE ship gate: Rect Size keyframe persistence gate.
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

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

// runV2_2RectKfShipGate is the V2.2.1 focused keyframe-persistence gate: a
// ShapeLayer with an animated Rect Size (2 linear keyframes) + Fill. Asserts
// AE accepts it and the re-saved Rect Size keyframe container (lhd3 numKf +
// ldat values) round-trips — proving shape sub-stream keyframes persist (the
// non-spatial Vec2 ldat is byte-identical to AE's own encoding).
func runV2_2RectKfShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_rectkf.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_rectkf.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_rectkf.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_rectkf_args.json`

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := comp.NewShapeLayer("RectKf")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	r, err := l.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	_ = r.Size().AddKeyframeLinear(0, [2]float64{50, 50})
	_ = r.Size().AddKeyframeLinear(2, [2]float64{300, 200})
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

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2_rectkf.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("rect-kf ship gate FAIL:\n%s", string(content))
	}

	// Decode the re-saved Rect Size keyframe container.
	root := parseAEP(t, resavedAEP)
	kfl := findShipList(root, "ADBE Vector Rect Size")
	if kfl == nil {
		t.Fatalf("resaved Rect Size has no LIST(list) keyframe container — keyframes dropped")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	ldat := findShipChunk(kfl, rifx.IDLdat)
	if lhd3 == nil || ldat == nil {
		t.Fatalf("resaved Rect Size kf container missing lhd3/ldat")
	}
	numKf := binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C])
	bpk := int(binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
	if numKf != 2 {
		t.Errorf("resaved Rect Size numKf = %d, want 2", numKf)
	}
	// Non-spatial Vec2: value at block+0x08 (2 f64). Check kf0=[50,50], kf1=[300,200].
	rdF64 := func(off int) float64 { return math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[off : off+8])) }
	if got0x, got0y := rdF64(0x08), rdF64(0x10); math.Abs(got0x-50) > 0.5 || math.Abs(got0y-50) > 0.5 {
		t.Errorf("resaved Rect Size kf0 = (%.3g,%.3g), want (50,50)", got0x, got0y)
	}
	if got1x, got1y := rdF64(bpk+0x08), rdF64(bpk+0x10); math.Abs(got1x-300) > 0.5 || math.Abs(got1y-200) > 0.5 {
		t.Errorf("resaved Rect Size kf1 = (%.3g,%.3g), want (300,200)", got1x, got1y)
	}
}

func TestV2_2_RectKf_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2RectKfShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_RectKf_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2RectKfShipGate(t, aep.TargetAE2020, aeExe)
}
