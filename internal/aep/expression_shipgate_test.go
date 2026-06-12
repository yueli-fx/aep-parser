// internal/aep/expression_shipgate_test.go
//
// AE ship gate for expression ACTIVATION (MG roadmap S2,
// specs/2026-06-12-from-scratch-mg-roadmap.md). SetExpression has always
// round-tripped in Go; what was never true is AE *evaluating* it — the
// tdb4 @0x78 enabled byte was written inverted, leaving every expression
// render-dead (delivery-contract red-line-1 origin case).
//
// Per red line 4 the gate verifies the capability surface — the RENDERED
// result of an expression — not just the readback flag:
//
//   - ON:  eccentric dot (circle 300px above anchor), rotation expression
//     "time*90" ENABLED → at t=2s rotated 180°, the dot hangs BELOW its
//     anchor: pixels at (660,840), nothing at (660,240).
//   - OFF: same geometry at x=1260 with the same expression DISABLED →
//     stays at (1260,240).
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func buildExpressionDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "EXPR", 1920, 1080, 30, 4)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	bg, err := aep.NewShapeLayer(comp, "BG")
	if err != nil {
		t.Fatalf("NewShapeLayer BG: %v", err)
	}
	rect, err := bg.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := rect.SetSize([2]float64{2200, 1300}); err != nil {
		t.Fatalf("rect SetSize: %v", err)
	}
	bgFill, err := bg.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill BG: %v", err)
	}
	if err := bgFill.SetColor([4]float64{0.05, 0.05, 0.08, 1}); err != nil {
		t.Fatalf("BG SetColor: %v", err)
	}
	if err := bg.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("BG Position: %v", err)
	}

	addEccentricDot := func(name string, color [4]float64, x float64) {
		dl, err := aep.NewShapeLayer(comp, name)
		if err != nil {
			t.Fatalf("NewShapeLayer %s: %v", name, err)
		}
		el, err := dl.RootGroup().AddEllipse()
		if err != nil {
			t.Fatalf("AddEllipse %s: %v", name, err)
		}
		if err := el.SetSize([2]float64{72, 72}); err != nil {
			t.Fatalf("%s SetSize: %v", name, err)
		}
		if err := el.SetPosition([2]float64{0, -300}); err != nil {
			t.Fatalf("%s SetPosition: %v", name, err)
		}
		f, err := dl.RootGroup().AddFill()
		if err != nil {
			t.Fatalf("AddFill %s: %v", name, err)
		}
		if err := f.SetColor(color); err != nil {
			t.Fatalf("%s SetColor: %v", name, err)
		}
		if err := dl.Position().SetStaticValue([2]float64{x, 540}); err != nil {
			t.Fatalf("%s Position: %v", name, err)
		}
	}
	addEccentricDot("ON", [4]float64{1.0, 0.55, 0.1, 1}, 660)   // amber
	addEccentricDot("OFF", [4]float64{0.25, 0.85, 1.0, 1}, 1260) // cyan

	// Expressions need parsed back-refs (tdbs); attach after Reopen.
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	cc := rp.Compositions[0]
	for _, spec := range []struct {
		layer   string
		enabled bool
	}{{"ON", true}, {"OFF", false}} {
		rot := cc.LayerByName(spec.layer).Rotation()
		if rot == nil {
			t.Fatalf("%s: no rotation property", spec.layer)
		}
		if err := rot.SetExpression("time*90"); err != nil {
			t.Fatalf("%s SetExpression: %v", spec.layer, err)
		}
		if err := rot.SetExpressionEnabled(spec.enabled); err != nil {
			t.Fatalf("%s SetExpressionEnabled: %v", spec.layer, err)
		}
	}
	if err := aep.MoveToEnd(cc.LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}
	return rp
}

func runExpressionGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/expression_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_expression.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildExpressionDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "expr_in.aep")
	resavedAEP := filepath.Join(tempDir, "expr_resaved.aep")
	doneFile := filepath.Join(tempDir, "expr.done")
	framePNG := filepath.Join(tempDir, "expr_frame.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"png":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(framePNG))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 240)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("expression %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("expression %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at t=2s.
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const tol = 40
	amber := [3]uint8{255, 140, 26}
	cyan := [3]uint8{64, 217, 255}
	bgCol := [3]uint8{13, 13, 20}
	checks := []struct {
		name string
		x, y int
		want [3]uint8
	}{
		{"ON rotated 180° → below anchor", 660, 840, amber},
		{"ON origin vacated", 660, 240, bgCol},
		{"OFF stays above anchor", 1260, 240, cyan},
		{"OFF below stays background", 1260, 840, bgCol},
	}
	for _, c := range checks {
		r, g, b, _ := img.At(c.x, c.y).RGBA()
		got := [3]int{int(r >> 8), int(g >> 8), int(b >> 8)}
		for i := range 3 {
			d := got[i] - int(c.want[i])
			if d < -tol || d > tol {
				t.Errorf("%s render %s @(%d,%d): got RGB%v want ~%v", ver, c.name, c.x, c.y, got, c.want)
				break
			}
		}
	}

	// Resave proof: expression source + enabled state survive AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	cc := re.Compositions[0]
	for _, want := range []struct {
		layer   string
		enabled bool
	}{{"ON", true}, {"OFF", false}} {
		l := cc.LayerByName(want.layer)
		if l == nil || l.Rotation() == nil {
			t.Errorf("resaved %s: rotation stream gone (expression dropped on resave)", want.layer)
			continue
		}
		rot := l.Rotation()
		if rot.Expression != "time*90" || rot.ExpressionEnabled != want.enabled {
			t.Errorf("resaved %s: expr=%q enabled=%v, want time*90/%v", want.layer, rot.Expression, rot.ExpressionEnabled, want.enabled)
		}
	}
}

func TestExpression_AEShipGate_AE2020(t *testing.T) {
	runExpressionGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestExpression_AEShipGate_AE2025(t *testing.T) {
	runExpressionGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
