// internal/aep/shape_layrxfkf_shipgate_test.go
//
// AE ship gate: Layer Position and Transform keyframe gates.
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

// runV2_2LayrPosKfShipGate is the V2.2.1 Layr Transform Position keyframe gate
// (Path B): a ShapeLayer whose combined Transform Position ("ADBE Position") is
// animated (2 linear keyframes). Persisted as a bpk-128 spatial dim-3 block
// (value@0x38 X/Y/Z, Z=0; motion-path marker@0x08) injected into the embedded
// transform-group template. Confirms AE accepts the path end-to-end + the
// re-saved keyframes round-trip their VALUES.
func runV2_2LayrPosKfShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_layrposkf.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_layrposkf.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_layrposkf.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_layrposkf_args.json`

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, _ := comp.NewShapeLayer("PosAnim")
	r, _ := l.RootGroup().AddRect()
	_ = r.SetSize([2]float64{200, 100})
	_ = l.Position().AddKeyframeLinear(0, [2]float64{100, 200})
	_ = l.Position().AddKeyframeLinear(2, [2]float64{700, 400})

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

	runAeRunShipGate(t, aeExe, `E:/projects/tools/aep-parser/test_data/verify_v2_2_layrposkf.jsx`, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("layr-pos-kf ship gate FAIL:\n%s", string(content))
	}

	root := parseAEP(t, resavedAEP)
	kfl := findShipList(root, "ADBE Position")
	if kfl == nil {
		t.Fatalf("resaved Layr Position has no keyframe container — position keyframes dropped")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
		t.Errorf("resaved Layr Position numKf != 2")
	}
	// Spatial dim-3 block: value at block+0x38 (X), +0x40 (Y), +0x48 (Z=0).
	ldat := findShipChunk(kfl, rifx.IDLdat)
	if ldat != nil && len(ldat.Data) >= 0x48+8 {
		x := math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[0x38:0x40]))
		y := math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[0x40:0x48]))
		if math.Abs(x-100) > 0.5 || math.Abs(y-200) > 0.5 {
			t.Errorf("resaved Layr Position kf0 = (%.3g,%.3g), want (100,200)", x, y)
		}
	}
}

func TestV2_2_LayrPosKf_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2LayrPosKfShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_LayrPosKf_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2LayrPosKfShipGate(t, aep.TargetAE2020, aeExe)
}

// runV2_2XfKfShipGate is the V2.2.1 子项⑥ gate: a ShapeLayer whose Transform
// Anchor / Scale / Rotation / Opacity are each animated (2 linear keyframes).
// On disk: Anchor = 3D spatial motion-path (bpk-128, like Position); Scale = 3D
// non-spatial (bpk-128, ÷100, Z=1.0); Rotation/Opacity = 1D non-spatial (bpk-48;
// Rotation degrees, Opacity ÷100). Confirms AE accepts the channels + reads back
// user-facing values, and the re-saved keyframes round-trip.
func runV2_2XfKfShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_xfkf.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_xfkf.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_xfkf.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_xfkf_args.json`

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, _ := comp.NewShapeLayer("XfAnim")
	r, _ := l.RootGroup().AddRect()
	_ = r.SetSize([2]float64{200, 100})
	_ = l.AnchorPoint().AddKeyframeLinear(0, [2]float64{10, 20})
	_ = l.AnchorPoint().AddKeyframeLinear(2, [2]float64{50, 60})
	_ = l.Scale().AddKeyframeLinear(0, [2]float64{100, 100})
	_ = l.Scale().AddKeyframeLinear(2, [2]float64{150, 200})
	_ = l.Rotation().AddKeyframeLinear(0, 0)
	_ = l.Rotation().AddKeyframeLinear(2, 90)
	_ = l.Opacity().AddKeyframeLinear(0, 100)
	_ = l.Opacity().AddKeyframeLinear(2, 50)

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

	runAeRunShipGate(t, aeExe, `E:/projects/tools/aep-parser/test_data/verify_v2_2_xfkf.jsx`, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("xf-kf ship gate FAIL:\n%s", string(content))
	}

	// Re-parse: each channel keeps 2 keyframes; spot-check on-disk (raw) values.
	root := parseAEP(t, resavedAEP)
	for _, ch := range []struct {
		name string
		bpk  uint32
	}{
		{"ADBE Anchor Point", 128}, {"ADBE Scale", 128}, {"ADBE Rotate Z", 48}, {"ADBE Opacity", 48},
	} {
		kfl := findShipList(root, ch.name)
		if kfl == nil {
			t.Errorf("resaved %s: keyframes dropped", ch.name)
			continue
		}
		lhd3 := findShipChunk(kfl, rifx.IDLhd3)
		if lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
			t.Errorf("resaved %s numKf != 2", ch.name)
		}
		if lhd3 != nil && binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]) != ch.bpk {
			t.Errorf("resaved %s bpk = %d, want %d", ch.name, binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]), ch.bpk)
		}
	}
}

func TestV2_2_XfKf_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2XfKfShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_XfKf_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2XfKfShipGate(t, aep.TargetAE2020, aeExe)
}
