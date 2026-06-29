// internal/aep/shape_gradient_stroke_shipgate_test.go
//
// AE ship gate for gradient stroke (AddGradientStroke). Builds one ShapeLayer
// (rect + gradient stroke, 3 custom color stops + a non-default alpha ramp), has
// AE open + resave it, and decodes the resaved Grad Colors stops via the read
// path. Proves AE accepts the from-scratch G-Stroke body + generated prop.map
// XML (no silent-drop) AND preserves the custom color + alpha stops. Gated by
// AE_SHIP_GATE.
//
// Reuses verify_v2_2_gradstroke.jsx (layer "GradStroke_Static" with rect +
// G-Stroke). Symmetric to the gradient-fill ship gate.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/codec"
)

func runV2_2GradientStrokeShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_gradstroke.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_gradstroke.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_gradstroke.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/generated/args/v2_2_gradstroke_gate_args.json`

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "GradStroke_Static") // name expected by verify_v2_2_gradstroke.jsx
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, _ := l.RootGroup().AddRect()
	_ = rect.SetSize([2]float64{200, 100})
	gs, _ := l.RootGroup().AddGradientStroke()
	if err := gs.SetColorStops([]codec.GradientColorStop{
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
		{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{0, 1, 0}},
		{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
	}); err != nil {
		t.Fatalf("SetColorStops: %v", err)
	}
	// Non-default alpha ramp — exercises the alpha write path (else untested).
	if err := gs.SetAlphaStops([]codec.GradientAlphaStop{
		{Offset: 0, Midpoint: 0.5, Alpha: 1.0},
		{Offset: 1, Midpoint: 0.5, Alpha: 0.3},
	}); err != nil {
		t.Fatalf("SetAlphaStops: %v", err)
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
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	jsxPath := `E:/projects/tools/aep-parser/test_data/generators/verify_v2_2_gradstroke.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("gradient stroke ship gate FAIL:\n%s", string(content))
	}

	// Decode the resaved gradient via the read path; confirm AE preserved the
	// 3 custom color stops (red / green / blue) + the non-default last alpha.
	proj, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("open resaved: %v", err)
	}
	var grad *codec.Gradient
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
	if n := len(grad.AlphaStops); n != 2 {
		t.Fatalf("resaved alpha stops = %d, want 2", n)
	} else if a := grad.AlphaStops[n-1].Alpha; abs(a-0.3) > 0.02 {
		t.Errorf("resaved last alpha = %g, want ~0.3", a)
	}
}

func TestV2_2_GradientStroke_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2GradientStrokeShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_GradientStroke_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2GradientStrokeShipGate(t, aep.TargetAE2020, aeExe)
}
