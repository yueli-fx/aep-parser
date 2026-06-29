// internal/aep/shape_ellkf_shipgate_test.go
//
// AE ship gate: Ellipse keyframe persistence gate.
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

// runV2_2EllKfShipGate: Ellipse with animated Size (non-spatial Vec2) +
// Position (spatial Vec2 motion-path, bpk 104). Position ldat differs from AE
// only in the auto-computed ~0 spatial tangent (AE recomputes on load); the
// gate confirms AE accepts it and round-trips the keyframe VALUES.
func runV2_2EllKfShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_ellkf.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_ellkf.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_ellkf.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_ellkf_args.json`

	p := aep.NewProject(target)
	comp, _ := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	l, _ := aep.NewShapeLayer(comp, "EllAnim")
	ell, _ := l.RootGroup().AddEllipse()
	_ = ell.Size().AddKeyframeLinear(0, [2]float64{50, 50})
	_ = ell.Size().AddKeyframeLinear(2, [2]float64{300, 200})
	_ = ell.Position().AddKeyframeLinear(0, [2]float64{10, 20})
	_ = ell.Position().AddKeyframeLinear(2, [2]float64{70, 90})

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
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, `E:/projects/tools/aep-parser/test_data/verify_v2_2_ellkf.jsx`, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("ell-kf ship gate FAIL:\n%s", string(content))
	}

	root := parseAEP(t, resavedAEP)
	// Position keyframes survive: spatial block value at 0x38 (kf0 [10,20]).
	if kfl := findShipList(root, "ADBE Vector Ellipse Position"); kfl != nil {
		if lhd3 := findShipChunk(kfl, rifx.IDLhd3); lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
			t.Errorf("resaved Ellipse Position numKf != 2")
		}
		if ldat := findShipChunk(kfl, rifx.IDLdat); ldat != nil && len(ldat.Data) >= 0x48 {
			x := math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[0x38:0x40]))
			y := math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[0x40:0x48]))
			if math.Abs(x-10) > 0.5 || math.Abs(y-20) > 0.5 {
				t.Errorf("resaved Ellipse Position kf0 = (%.3g,%.3g), want (10,20)", x, y)
			}
		}
	} else {
		t.Errorf("resaved Ellipse Position keyframes dropped")
	}
}

func TestV2_2_EllKf_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2EllKfShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_EllKf_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2EllKfShipGate(t, aep.TargetAE2020, aeExe)
}
