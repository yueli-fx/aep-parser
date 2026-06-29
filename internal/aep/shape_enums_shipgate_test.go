// internal/aep/shape_enums_shipgate_test.go
//
// AE ship gate for shape-node enums (Rect/Ellipse Direction, Fill Blend Mode /
// Composite Order / Fill Rule, Stroke Blend Mode / Composite Order). Builds one
// ShapeLayer carrying rect+ellipse+fill+stroke with the enums set non-default,
// has AE open + resave it, and decodes the resaved cdats. Primarily proves AE
// accepts the enriched templates (no silent-drop); per-node value round-trip is
// covered by TestV2_2_ShapeEnums_Roundtrip. Gated by AE_SHIP_GATE.
//
// Reuses verify_v2_2_stroke.jsx (layer named "Stroke_Static" with rect+stroke;
// the extra ellipse+fill are ignored by that verifier).
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

func runV2_2ShapeEnumsShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_enums.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_enums.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_enums.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_stroke_args.json`

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "Stroke_Static") // name expected by verify_v2_2_stroke.jsx
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, _ := l.RootGroup().AddRect()
	_ = rect.SetSize([2]float64{200, 100})
	_ = rect.SetDirection(aep.ShapeDirectionReversed) // 3
	ell, _ := l.RootGroup().AddEllipse()
	_ = ell.SetDirection(aep.ShapeDirectionReversed) // 3
	fill, _ := l.RootGroup().AddFill()
	_ = fill.SetBlendMode(3)
	_ = fill.SetCompositeOrder(aep.ShapeCompositeOrderBelowPrevious) // 2
	_ = fill.SetFillRule(aep.FillRuleEvenOdd)                        // 2
	stroke, _ := l.RootGroup().AddStroke()
	_ = stroke.SetBlendMode(4)
	_ = stroke.SetCompositeOrder(aep.ShapeCompositeOrderBelowPrevious) // 2

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
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("shape-enums ship gate FAIL:\n%s", string(content))
	}

	root := parseAEP(t, resavedAEP)
	rdF64 := func(b []byte, off int) float64 {
		return math.Float64frombits(binary.BigEndian.Uint64(b[off : off+8]))
	}
	// Uniquely identifiable via first-match streamCdat:
	//   Fill Rule  — only on fill (2)
	//   Direction  — first match = rect (3)
	//   Composite Order — fill & stroke both 2
	//   Blend Mode — first match = fill (3); stroke's 4 covered by Go roundtrip
	checks := []struct {
		name string
		want float64
	}{
		{"ADBE Vector Shape Direction", 3},
		{"ADBE Vector Fill Rule", 2},
		{"ADBE Vector Blend Mode", 3},
		{"ADBE Vector Composite Order", 2},
	}
	for _, c := range checks {
		cdat := streamCdat(root, c.name)
		if cdat == nil {
			t.Errorf("resaved %q cdat missing — AE dropped the slot?", c.name)
			continue
		}
		if got := rdF64(cdat, 0); math.Abs(got-c.want) > 0.01 {
			t.Errorf("resaved %q = %.3g, want %.3g", c.name, got, c.want)
		}
	}
}

func TestV2_2_ShapeEnums_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2ShapeEnumsShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_ShapeEnums_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2ShapeEnumsShipGate(t, aep.TargetAE2020, aeExe)
}
