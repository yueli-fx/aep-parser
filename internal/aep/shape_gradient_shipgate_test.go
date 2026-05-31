// internal/aep/shape_gradient_shipgate_test.go
//
// AE ship gate for gradient fill (SetGradient). Builds one ShapeLayer
// (rect + gradient fill, 3 custom color stops), has AE open + resave it, and
// decodes the resaved Grad Colors stops XML. Proves AE accepts the from-scratch
// gradient-fill body + generated prop.map XML (no silent-drop) AND preserves
// the custom stops. Gated by AE_SHIP_GATE.
//
// This is the empirical cross-version test: the only stops-bearing gradient
// fixture is AE 25.6-saved (AE 2020 can't open it), so whether AE 2020 accepts
// an AE25-shaped gradient body built from scratch is answered here.
//
// Reuses verify_v2_2_gradient.jsx (layer "Grad_Static" with rect + G-Fill).
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func runV2_2GradientFillShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_gradfill.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_gradfill.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_gradfill.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_gradient_gate_args.json`

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := comp.NewShapeLayer("Grad_Static") // name expected by verify_v2_2_gradient.jsx
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, _ := l.RootGroup().AddRect()
	_ = rect.SetSize([2]float64{200, 100})
	gf, _ := l.RootGroup().AddGradientFill()
	if err := gf.SetColorStops([]aep.GradientColorStop{
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
		{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{0, 1, 0}},
		{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
	}); err != nil {
		t.Fatalf("SetColorStops: %v", err)
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

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2_gradient.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("gradient ship gate FAIL:\n%s", string(content))
	}

	// Decode the resaved gradient via the read path; confirm AE preserved the
	// 3 custom color stops (red / green / blue).
	proj, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("open resaved: %v", err)
	}
	var grad *aep.Gradient
	for _, c := range proj.Compositions {
		for _, ly := range c.Layers {
			for _, pr := range ly.Properties {
				if pr.MatchName == "ADBE Vector Grad Colors" && pr.Gradient != nil {
					grad = pr.Gradient
				}
			}
		}
	}
	if grad == nil {
		t.Fatal("resaved file has no decodable gradient — AE dropped the stops?")
	}
	if len(grad.ColorStops) != 3 {
		t.Fatalf("resaved color stops = %d, want 3:\n%+v", len(grad.ColorStops), grad.ColorStops)
	}
	wants := [3][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
	for i, w := range wants {
		got := grad.ColorStops[i].Color
		for c := 0; c < 3; c++ {
			if abs(got[c]-w[c]) > 0.02 {
				t.Errorf("resaved color stop %d = %v, want ~%v", i, got, w)
				break
			}
		}
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func TestV2_2_GradientFill_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2GradientFillShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_GradientFill_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2GradientFillShipGate(t, aep.TargetAE2020, aeExe)
}
