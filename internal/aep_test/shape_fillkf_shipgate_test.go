// internal/aep/shape_fillkf_shipgate_test.go
//
// AE ship gate: Fill Color keyframe persistence gate.
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

// runV2_2FillKfShipGate is the V2.2.1 Color-keyframe gate: a ShapeLayer with a
// Rect + Fill whose Color is animated (2 linear keyframes). Color keyframes use
// the spatial-style ldat block (value@0x38, bpk 152) with [A,R,G,B]×255 values
// — byte-identical to AE's own encoding (verified offline). Confirms AE accepts
// the color-keyframe path end-to-end + the re-saved keyframes round-trip.
func runV2_2FillKfShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_fillkf.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_fillkf.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_fillkf.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/generated/args/v2_2_fillkf_args.json`

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, _ := aep.NewShapeLayer(comp, "FillKf")
	r, _ := l.RootGroup().AddRect()
	_ = r.SetSize([2]float64{200, 100})
	fill, _ := l.RootGroup().AddFill()
	_ = fill.Color().AddKeyframeLinear(0, [4]float64{1, 0, 0, 1})
	_ = fill.Color().AddKeyframeLinear(2, [4]float64{0, 0, 1, 1})

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`, toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, `E:/projects/tools/aep-parser/test_data/generators/verify_v2_2_fillkf.jsx`, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("fill-kf ship gate FAIL:\n%s", string(content))
	}

	root := parseAEP(t, resavedAEP)
	kfl := findShipList(root, "ADBE Vector Fill Color")
	if kfl == nil {
		t.Fatalf("resaved Fill Color has no keyframe container — color keyframes dropped")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
		t.Errorf("resaved Fill Color numKf != 2")
	}
	// Color block is spatial-style: value (ARGB×255) at block+0x38.
	ldat := findShipChunk(kfl, rifx.IDLdat)
	if ldat != nil && len(ldat.Data) >= 0x38+32 {
		a := math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[0x38:0x40]))
		rr := math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[0x40:0x48]))
		if math.Abs(a-255) > 1 || math.Abs(rr-255) > 1 { // kf0 [1,0,0,1] → A=255,R=255
			t.Errorf("resaved Fill Color kf0 ARGB = (%.3g,%.3g,..), want A=255,R=255", a, rr)
		}
	}
}

func TestV2_2_FillKf_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2FillKfShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_FillKf_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2FillKfShipGate(t, aep.TargetAE2020, aeExe)
}
