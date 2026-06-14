// internal/aep/mg_gradient_hilite_shipgate_test.go
//
// AE ship gate for a radial gradient's HIGHLIGHT (`ADBE Vector Grad HiLite
// Length` / `Angle`) from scratch. Verified at the capability's surface per
// delivery-contract red line 4: a 400×400 rect filled with a red→blue RADIAL
// gradient centred at the shape origin, then HiLite Length = 70 (Angle = 0)
// shifts the bright red centre off the geometric centre along the highlight
// axis. The gate renders the frame and asserts the fill is symmetric along ONE
// axis only — the cardinal pair on the highlight axis differs sharply (the red
// hotspot moved toward one side, the opposite side is blue) while the
// perpendicular pair stays matched. A centred radial (HiLite Length 0) would
// make all four cardinals equal (the radial gate); a linear ramp would split
// both pairs. Only a real highlight offset produces the one-axis split.
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

func buildMGGradientHiliteDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGHILITE", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// 400×400 rect centred at 960,540. Local space [-200,200]; radial centre at
	// [0,0], outer radius 200 (End=[200,0]). HiLite Length 70 shifts the red
	// hotspot 0.7·radius off-centre along the highlight axis.
	ramp, err := aep.NewShapeLayer(comp, "HILITE")
	if err != nil {
		t.Fatalf("NewShapeLayer HILITE: %v", err)
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
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}}, // red (centre / hotspot)
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
	if err := gf.SetHighlightLength(70); err != nil {
		t.Fatalf("SetHighlightLength: %v", err)
	}
	if err := gf.SetHighlightAngle(0); err != nil {
		t.Fatalf("SetHighlightAngle: %v", err)
	}
	if err := ramp.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("HILITE Position: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	return rp
}

func runMGGradientHiliteGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_gradient_hilite_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_gradient_hilite.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGGradientHiliteDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_gradient_hilite_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_gradient_hilite_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_gradient_hilite.done")
	framePNG := filepath.Join(tempDir, "mg_gradient_hilite_frame.png")

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
	t.Logf("mg gradient-hilite %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg gradient-hilite %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0. Sample the four cardinal points at radius
	// 130 from the geometric centre (960,540). With the hotspot shifted along
	// the highlight axis, exactly one cardinal pair must split (one side red,
	// the other blue) while the perpendicular pair stays matched.
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
	lR, lG, lB := avgRGB(img, 830, 540, win) // left
	rR, rG, rB := avgRGB(img, 1090, 540, win) // right
	uR, uG, uB := avgRGB(img, 960, 410, win) // up
	dR, dG, dB := avgRGB(img, 960, 670, win) // down
	t.Logf("%s hilite: L(r=%d,g=%d,b=%d) R(r=%d,g=%d,b=%d) U(r=%d,g=%d,b=%d) D(r=%d,g=%d,b=%d)",
		ver, lR, lG, lB, rR, rG, rB, uR, uG, uB, dR, dG, dB)

	dx := absInt(lR - rR) // red split along the horizontal (X) axis
	dy := absInt(uR - dR) // red split along the vertical (Y) axis
	// Exactly one axis must split sharply; the perpendicular axis stays matched.
	if max(dx, dy) < 60 {
		t.Errorf("%s no axis split (dx=%d dy=%d) — highlight not shifting the hotspot (looks centred radial)", ver, dx, dy)
	}
	if min(dx, dy) > 45 {
		t.Errorf("%s both axes split (dx=%d dy=%d) — fill is linear, not a single-axis highlight", ver, dx, dy)
	}
	// The matched (perpendicular) pair must also agree on blue so it is a genuine
	// symmetric pair, not noise.
	if dx < dy { // X is the matched axis
		if absInt(lB-rB) > 45 || absInt(lG-rG) > 45 {
			t.Errorf("%s matched X axis not symmetric (L b=%d g=%d, R b=%d g=%d)", ver, lB, lG, rB, rG)
		}
	} else { // Y is the matched axis
		if absInt(uB-dB) > 45 || absInt(uG-dG) > 45 {
			t.Errorf("%s matched Y axis not symmetric (U b=%d g=%d, D b=%d g=%d)", ver, uB, uG, dB, dG)
		}
	}

	// Resave proof: the highlighted radial gradient survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("HILITE") == nil {
		t.Fatal("resaved: HILITE layer missing")
	}
}

func TestMGGradientHilite_AEShipGate_AE2020(t *testing.T) {
	runMGGradientHiliteGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGGradientHilite_AEShipGate_AE2025(t *testing.T) {
	runMGGradientHiliteGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
