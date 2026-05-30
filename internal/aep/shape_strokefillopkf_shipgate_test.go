// internal/aep/shape_strokefillopkf_shipgate_test.go
//
// V2.2 Phase 5 Task 5.2 — AE ship gate: Rect sub-property, Stroke, and Fill Opacity keyframe gates.
// Gated by AE_SHIP_GATE env var; CI / no-AE runs SKIP cleanly.
package aep_test

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

// runV2_2RectSubKfShipGate is the V2.2.1 子项⑦ gate: a Rect whose Position
// (spatial Vec2 motion-path, bpk-104) + Roundness (1D non-spatial, bpk-48) are
// animated. Persisted via the richer rect body template (Size/Position/Roundness
// cdat slots). Confirms AE accepts the channels + the re-saved keyframes survive.
func runV2_2RectSubKfShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_rectsubkf.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_rectsubkf.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_rectsubkf.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_rectsubkf_args.json`

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, _ := comp.NewShapeLayer("RectSub")
	r, _ := l.RootGroup().AddRect()
	_ = r.SetSize([2]float64{200, 100})
	_ = r.Position().AddKeyframeLinear(0, [2]float64{0, 0})
	_ = r.Position().AddKeyframeLinear(2, [2]float64{40, 50})
	_ = r.Roundness().AddKeyframeLinear(0, 0)
	_ = r.Roundness().AddKeyframeLinear(2, 20)

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

	runAeRunShipGate(t, aeExe, `E:/projects/tools/aep-parser/test_data/verify_v2_2_rectsubkf.jsx`, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("rect-sub-kf ship gate FAIL:\n%s", string(content))
	}

	root := parseAEP(t, resavedAEP)
	for _, ch := range []struct {
		name string
		bpk  uint32
	}{{"ADBE Vector Rect Position", 104}, {"ADBE Vector Rect Roundness", 48}} {
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

func TestV2_2_RectSubKf_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2RectSubKfShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_RectSubKf_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2RectSubKfShipGate(t, aep.TargetAE2020, aeExe)
}

// runV2_2StrokeKfShipGate is the V2.2.1 子项⑧ gate: a Stroke whose Opacity (raw
// %) + Width (raw px) are animated — both 1D non-spatial (bpk-48). The stroke
// body template already carries both cdat slots, so they flip in place (no new
// template). Confirms AE accepts the channels + the re-saved keyframes survive.
func runV2_2StrokeKfShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_strokekf.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_strokekf.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_strokekf.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_strokekf_args.json`

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, _ := comp.NewShapeLayer("StrokeAnim")
	r, _ := l.RootGroup().AddRect()
	_ = r.SetSize([2]float64{200, 100})
	stroke, _ := l.RootGroup().AddStroke()
	_ = stroke.Opacity().AddKeyframeLinear(0, 100)
	_ = stroke.Opacity().AddKeyframeLinear(2, 50)
	_ = stroke.Width().AddKeyframeLinear(0, 5)
	_ = stroke.Width().AddKeyframeLinear(2, 20)

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

	runAeRunShipGate(t, aeExe, `E:/projects/tools/aep-parser/test_data/verify_v2_2_strokekf.jsx`, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("stroke-kf ship gate FAIL:\n%s", string(content))
	}

	root := parseAEP(t, resavedAEP)
	for _, name := range []string{"ADBE Vector Stroke Opacity", "ADBE Vector Stroke Width"} {
		kfl := findShipList(root, name)
		if kfl == nil {
			t.Errorf("resaved %s: keyframes dropped", name)
			continue
		}
		lhd3 := findShipChunk(kfl, rifx.IDLhd3)
		if lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
			t.Errorf("resaved %s numKf != 2", name)
		}
		if lhd3 != nil && binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]) != 48 {
			t.Errorf("resaved %s bpk = %d, want 48", name, binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
		}
	}
}

func TestV2_2_StrokeKf_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2StrokeKfShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_StrokeKf_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2StrokeKfShipGate(t, aep.TargetAE2020, aeExe)
}

// runV2_2FillOpKfShipGate is the V2.2.1 子项⑨ gate: a Fill whose Opacity (raw %)
// is animated — 1D non-spatial (bpk-48). Persisted via the richer fill body
// template (Color + Opacity cdat slots). Confirms AE acceptance + kf survival.
func runV2_2FillOpKfShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_fillopkf.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_fillopkf.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_fillopkf.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_fillopkf_args.json`

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, _ := comp.NewShapeLayer("FillAnim")
	r, _ := l.RootGroup().AddRect()
	_ = r.SetSize([2]float64{200, 100})
	fill, _ := l.RootGroup().AddFill()
	_ = fill.Opacity().AddKeyframeLinear(0, 100)
	_ = fill.Opacity().AddKeyframeLinear(2, 40)

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

	runAeRunShipGate(t, aeExe, `E:/projects/tools/aep-parser/test_data/verify_v2_2_fillopkf.jsx`, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("fill-op-kf ship gate FAIL:\n%s", string(content))
	}

	root := parseAEP(t, resavedAEP)
	kfl := findShipList(root, "ADBE Vector Fill Opacity")
	if kfl == nil {
		t.Fatalf("resaved Fill Opacity: keyframes dropped")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
		t.Errorf("resaved Fill Opacity numKf != 2")
	}
	if lhd3 != nil && binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]) != 48 {
		t.Errorf("resaved Fill Opacity bpk = %d, want 48", binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
	}
}

func TestV2_2_FillOpKf_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2FillOpKfShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_FillOpKf_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2FillOpKfShipGate(t, aep.TargetAE2020, aeExe)
}
