// internal/aep/shape_stroke_shipgate_test.go
//
// AE ship gate: Stroke static gate.
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
)

// runV2_2StrokeShipGate is the V2.2.1 focused Stroke gate: one ShapeLayer with
// a Rect (geometry) + Stroke. Asserts AE accepts the stroke (no silent-drop)
// and the re-saved Color/Width/Opacity cdat round-trip. Values differ from the
// embed fixture so the cdat overwrite is genuinely exercised. Color uses AE's
// [A,R,G,B]×255 f64 encoding (RE'd from the stroke tolerance fixture).
func runV2_2StrokeShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_stroke.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_stroke.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_stroke.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_stroke_args.json`

	color := [4]float64{1, 0, 0, 1} // red, distinct from embed fixture's [0,0,1,1]
	width, opacity := 4.0, 60.0

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "Stroke_Static")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	r, err := l.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	_ = r.SetSize([2]float64{200, 100})
	stroke, err := l.RootGroup().AddStroke()
	if err != nil {
		t.Fatalf("AddStroke: %v", err)
	}
	_ = stroke.SetColor(color)
	_ = stroke.SetWidth(width)
	_ = stroke.SetOpacity(opacity)
	// Line Cap / Join non-default (no coupling) → assert strict round-trip.
	// Miter Limit read empirically: AE may normalize it to 4 on save when
	// Line Join != Miter (RE'd hidden-property behavior).
	lineCap := aep.StrokeLineCapProjecting // 3
	lineJoin := aep.StrokeLineJoinRound    // 2
	miterLimit := 12.0
	_ = stroke.SetLineCap(lineCap)
	_ = stroke.SetLineJoin(lineJoin)
	_ = stroke.SetMiterLimit(miterLimit)

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

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2_stroke.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("stroke ship gate FAIL:\n%s", string(content))
	}

	// Decode re-saved cdats. Color = [A,R,G,B]×255 f64; Width/Opacity = f64.
	root := parseAEP(t, resavedAEP)
	colCdat := streamCdat(root, "ADBE Vector Stroke Color")
	wCdat := streamCdat(root, "ADBE Vector Stroke Width")
	opCdat := streamCdat(root, "ADBE Vector Stroke Opacity")
	if colCdat == nil || wCdat == nil || opCdat == nil {
		t.Fatalf("resaved stroke cdat missing (color=%v width=%v op=%v) — stroke dropped?",
			colCdat != nil, wCdat != nil, opCdat != nil)
	}
	rdF64 := func(b []byte, off int) float64 {
		return math.Float64frombits(binary.BigEndian.Uint64(b[off : off+8]))
	}
	// Color stored ARGB×255: [0]=A*255 [8]=R*255 [16]=G*255 [24]=B*255.
	wantARGB := []float64{color[3] * 255, color[0] * 255, color[1] * 255, color[2] * 255}
	for i, w := range wantARGB {
		if got := rdF64(colCdat, i*8); math.Abs(got-w) > 1.0 {
			t.Errorf("resaved stroke color[%d] = %.3g, want %.3g (ARGB×255)", i, got, w)
		}
	}
	if got := rdF64(wCdat, 0); math.Abs(got-width) > 0.01 {
		t.Errorf("resaved stroke width = %.3g, want %.3g", got, width)
	}
	if got := rdF64(opCdat, 0); math.Abs(got-opacity) > 0.01 {
		t.Errorf("resaved stroke opacity = %.3g, want %.3g", got, opacity)
	}

	// Line Cap / Line Join: non-default, no coupling → must survive AE resave.
	capCdat := streamCdat(root, "ADBE Vector Stroke Line Cap")
	joinCdat := streamCdat(root, "ADBE Vector Stroke Line Join")
	if capCdat == nil || joinCdat == nil {
		t.Fatalf("resaved Line Cap/Join cdat missing (cap=%v join=%v) — AE dropped the slot?",
			capCdat != nil, joinCdat != nil)
	}
	if got := rdF64(capCdat, 0); math.Abs(got-float64(lineCap)) > 0.01 {
		t.Errorf("resaved Line Cap = %.3g, want %d (Projecting)", got, lineCap)
	}
	if got := rdF64(joinCdat, 0); math.Abs(got-float64(lineJoin)) > 0.01 {
		t.Errorf("resaved Line Join = %.3g, want %d (Round)", got, lineJoin)
	}
	// Miter Limit: empirical. AE keeps it only when Line Join = Miter; with
	// Join = Round it normalizes to default 4. Accept either our value or 4,
	// and surface which so the behavior is recorded rather than silently passing.
	if miterCdat := streamCdat(root, "ADBE Vector Stroke Miter Limit"); miterCdat != nil {
		got := rdF64(miterCdat, 0)
		switch {
		case math.Abs(got-miterLimit) < 0.01:
			t.Logf("resaved Miter Limit = %.3g (our value preserved)", got)
		case math.Abs(got-4) < 0.01:
			t.Logf("resaved Miter Limit = 4 (AE normalized; Join != Miter hides it)")
		default:
			t.Errorf("resaved Miter Limit = %.3g, want %.3g or normalized 4", got, miterLimit)
		}
	}
}

func TestV2_2_Stroke_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2StrokeShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_Stroke_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2StrokeShipGate(t, aep.TargetAE2020, aeExe)
}
