// internal/aep/mg_gradient_radial_shipgate_test.go
//
// AE ship gate for gradient RADIAL type (`ADBE Vector Grad Type` = 2) from
// scratch. Verified at the capability's surface per delivery-contract red line
// 4: a 400×400 rect filled with a red→blue gradient set to RADIAL with the
// centre at the shape origin (Start=[0,0]) and the outer color reached at
// radius 200 (End=[200,0]). The gate renders the frame and asserts the fill is
// rotationally symmetric — the centre is red, points at equal radius in
// opposite directions are the SAME color, and the edge is blue. A linear ramp
// (the pre-type behavior) would make the left side red and the right blue
// (asymmetric), which the symmetry checks reject.
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

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func buildMGGradientRadialDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGRADIAL", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// 400×400 rect centred at 960,540 → comp span x∈[760,1160], y∈[340,740].
	// Local space is [-200,200]; radial centre at [0,0], outer radius 200.
	ramp, err := aep.NewShapeLayer(comp, "RADIAL")
	if err != nil {
		t.Fatalf("NewShapeLayer RADIAL: %v", err)
	}
	rect, err := ramp.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := rect.SetSize([2]float64{400, 400}); err != nil {
		t.Fatalf("rect SetSize: %v", err)
	}
	gf, err := ramp.RootGroup().AddGradientFill()
	if err != nil {
		t.Fatalf("AddGradientFill: %v", err)
	}
	if err := gf.SetColorStops([]aep.GradientColorStop{
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}}, // red (centre)
		{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}}, // blue (edge)
	}); err != nil {
		t.Fatalf("SetColorStops: %v", err)
	}
	if err := gf.SetGradientType(aep.GradientRadial); err != nil {
		t.Fatalf("SetGradientType: %v", err)
	}
	if err := gf.SetStartPoint([2]float64{0, 0}); err != nil {
		t.Fatalf("SetStartPoint: %v", err)
	}
	if err := gf.SetEndPoint([2]float64{200, 0}); err != nil {
		t.Fatalf("SetEndPoint: %v", err)
	}
	if err := ramp.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("RADIAL Position: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	return rp
}

func runMGGradientRadialGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mg_gradient_radial_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mg_gradient_radial.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGGradientRadialDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_gradient_radial_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_gradient_radial_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_gradient_radial.done")
	framePNG := filepath.Join(tempDir, "mg_gradient_radial_frame.png")

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
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
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
	t.Logf("mg gradient-radial %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg gradient-radial %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0: a radial red→blue fill centred at the rect
	// centre (960,540). The centre is red; the four cardinal points at radius
	// ~140px are the same mid color (rotational symmetry); the edge (~radius 195)
	// is blue. A linear horizontal ramp would make left red / right blue.
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const win = 10
	cR, _, cB := avgRGB(img, 960, 540, win)
	// Cardinal points at radius 140 from the centre.
	lR, lG, lB := avgRGB(img, 820, 540, win)  // left
	rR, rG, rB := avgRGB(img, 1100, 540, win) // right
	uR, uG, uB := avgRGB(img, 960, 400, win)  // up
	dR, dG, dB := avgRGB(img, 960, 680, win)  // down
	// Edge near radius ~195.
	eR, _, eB := avgRGB(img, 1150, 540, win)
	t.Logf("%s radial: centre(r=%d,b=%d) L(r=%d,g=%d,b=%d) R(r=%d,g=%d,b=%d) U(r=%d,g=%d,b=%d) D(r=%d,g=%d,b=%d) edge(r=%d,b=%d)",
		ver, cR, cB, lR, lG, lB, rR, rG, rB, uR, uG, uB, dR, dG, dB, eR, eB)

	if cR-cB < 50 {
		t.Errorf("%s centre not red (r=%d b=%d) — radial centre wrong", ver, cR, cB)
	}
	if eB-eR < 40 {
		t.Errorf("%s edge not blue (r=%d b=%d) — outer color wrong", ver, eR, eB)
	}
	// Rotational symmetry: opposite + orthogonal cardinal points must match.
	// (A linear horizontal ramp would make L red and R blue → large diffs.)
	if absInt(lR-rR) > 45 || absInt(lB-rB) > 45 {
		t.Errorf("%s left≠right (L r=%d b=%d, R r=%d b=%d) — fill is linear, not radial", ver, lR, lB, rR, rB)
	}
	if absInt(uR-dR) > 45 || absInt(uB-dB) > 45 {
		t.Errorf("%s up≠down (U r=%d b=%d, D r=%d b=%d) — fill not radially symmetric", ver, uR, uB, dR, dB)
	}
	if absInt(lR-uR) > 45 || absInt(lB-uB) > 45 {
		t.Errorf("%s left≠up (L r=%d b=%d, U r=%d b=%d) — fill not radially symmetric", ver, lR, lB, uR, uB)
	}
	if absInt(lG-rG) > 45 || absInt(uG-dG) > 45 {
		t.Errorf("%s green channel asymmetric — fill not radially symmetric", ver)
	}

	// Resave proof: the radial gradient survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("RADIAL") == nil {
		t.Fatal("resaved: RADIAL layer missing")
	}
}

func TestMGGradientRadial_AEShipGate_AE2020(t *testing.T) {
	runMGGradientRadialGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGGradientRadial_AEShipGate_AE2025(t *testing.T) {
	runMGGradientRadialGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
