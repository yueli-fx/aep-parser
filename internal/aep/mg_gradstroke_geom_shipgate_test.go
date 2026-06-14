// internal/aep/mg_gradstroke_geom_shipgate_test.go
//
// AE ship gate for gradient STROKE ramp geometry (`ADBE Vector Grad Type` /
// `Start Pt` / `End Pt`) from scratch. Verified at the capability's surface per
// delivery-contract red line 4 with two stroked rects in one comp:
//
//   - GSDIR: a 200×200 rect with an 18px gradient STROKE, red→blue LINEAR ramp
//     running left→right (Start=[-100,0] End=[100,0]). The left stroke band must
//     render red, the right band blue — proving SetStartPoint/SetEndPoint steer
//     the stroke's ramp direction.
//   - GSRAD: a 200×200 rect with an 18px gradient STROKE, red→blue RADIAL ramp
//     centred at the rect centre with radius 200 (End=[200,0]). The four
//     edge-midpoints of the stroke ring sit at equal radius (100 = 0.5·radius) →
//     they must render the SAME mid color (rotational symmetry). A linear ramp
//     would split left=red/right=blue, which the symmetry check rejects —
//     proving SetGradientType(radial) works on a stroke.
//
//   - GSHL: a 200×200 rect with an 18px gradient STROKE, red→blue RADIAL ramp
//     plus HiLite Length 70 / Angle 0 (+X). The highlight shifts the red hotspot
//     right, so the ring's right mid renders redder than its left mid (one-axis
//     split) while top/bottom stay matched — proving SetHighlightLength/Angle
//     work on a stroke.
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

func addGradStrokeRect(t *testing.T, comp *aep.Composition, name string, center [2]float64) *aep.GradientStrokeNode {
	t.Helper()
	l, err := aep.NewShapeLayer(comp, name)
	if err != nil {
		t.Fatalf("NewShapeLayer %s: %v", name, err)
	}
	rect, err := l.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("%s AddRect: %v", name, err)
	}
	if err := rect.SetSize([2]float64{200, 200}); err != nil {
		t.Fatalf("%s rect SetSize: %v", name, err)
	}
	gs, err := l.RootGroup().AddGradientStroke()
	if err != nil {
		t.Fatalf("%s AddGradientStroke: %v", name, err)
	}
	if err := gs.SetColorStops([]aep.GradientColorStop{
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}}, // red
		{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}}, // blue
	}); err != nil {
		t.Fatalf("%s SetColorStops: %v", name, err)
	}
	if err := l.Position().SetStaticValue(center); err != nil {
		t.Fatalf("%s Position: %v", name, err)
	}
	return gs
}

func buildMGGradStrokeGeomDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGGSTROKE", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// GSDIR — linear horizontal ramp on the left third of the comp.
	dir := addGradStrokeRect(t, comp, "GSDIR", [2]float64{360, 540})
	if err := dir.SetStartPoint([2]float64{-100, 0}); err != nil {
		t.Fatalf("GSDIR SetStartPoint: %v", err)
	}
	if err := dir.SetEndPoint([2]float64{100, 0}); err != nil {
		t.Fatalf("GSDIR SetEndPoint: %v", err)
	}

	// GSRAD — radial ramp centred at the rect centre, radius 200, in the middle.
	rad := addGradStrokeRect(t, comp, "GSRAD", [2]float64{960, 540})
	if err := rad.SetGradientType(aep.GradientRadial); err != nil {
		t.Fatalf("GSRAD SetGradientType: %v", err)
	}
	if err := rad.SetStartPoint([2]float64{0, 0}); err != nil {
		t.Fatalf("GSRAD SetStartPoint: %v", err)
	}
	if err := rad.SetEndPoint([2]float64{200, 0}); err != nil {
		t.Fatalf("GSRAD SetEndPoint: %v", err)
	}

	// GSHL — radial ramp + HiLite (Length 70 / Angle 0 = +X) on the right. The
	// highlight shifts the red hotspot right, so the ring's right mid renders
	// redder than its left mid (one-axis split), while top/bottom stay matched.
	hl := addGradStrokeRect(t, comp, "GSHL", [2]float64{1560, 540})
	if err := hl.SetGradientType(aep.GradientRadial); err != nil {
		t.Fatalf("GSHL SetGradientType: %v", err)
	}
	if err := hl.SetStartPoint([2]float64{0, 0}); err != nil {
		t.Fatalf("GSHL SetStartPoint: %v", err)
	}
	if err := hl.SetEndPoint([2]float64{200, 0}); err != nil {
		t.Fatalf("GSHL SetEndPoint: %v", err)
	}
	if err := hl.SetHighlightLength(70); err != nil {
		t.Fatalf("GSHL SetHighlightLength: %v", err)
	}
	if err := hl.SetHighlightAngle(0); err != nil {
		t.Fatalf("GSHL SetHighlightAngle: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	return rp
}

func runMGGradStrokeGeomGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_gradstroke_geom_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_gradstroke_geom.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGGradStrokeGeomDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_gradstroke_geom_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_gradstroke_geom_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_gradstroke_geom.done")
	framePNG := filepath.Join(tempDir, "mg_gradstroke_geom_frame.png")

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
	t.Logf("mg gradstroke-geom %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg gradstroke-geom %s ship gate FAIL:\n%s", ver, body)
	}

	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	// Tight window — the stroke band is only 18px wide.
	const win = 4

	// GSDIR (centre 360,540): left stroke band x≈260 red, right band x≈460 blue.
	dlR, _, dlB := avgRGB(img, 260, 540, win)
	drR, _, drB := avgRGB(img, 460, 540, win)
	t.Logf("%s GSDIR: left(r=%d,b=%d) right(r=%d,b=%d)", ver, dlR, dlB, drR, drB)
	if dlR-dlB < 50 {
		t.Errorf("%s GSDIR left stroke not red (r=%d b=%d) — ramp direction wrong", ver, dlR, dlB)
	}
	if drB-drR < 50 {
		t.Errorf("%s GSDIR right stroke not blue (r=%d b=%d) — ramp direction wrong", ver, drR, drB)
	}

	// GSRAD (centre 960,540): the four edge-midpoints of the 200×200 stroke ring
	// sit at radius 100 = 0.5·gradient-radius → same mid color (rotational
	// symmetry). A linear ramp would make left red / right blue.
	rlR, rlG, rlB := avgRGB(img, 860, 540, win)  // left
	rrR, rrG, rrB := avgRGB(img, 1060, 540, win) // right
	ruR, ruG, ruB := avgRGB(img, 960, 440, win)  // top
	rdR, rdG, rdB := avgRGB(img, 960, 640, win)  // bottom
	t.Logf("%s GSRAD ring mids: L(r=%d,g=%d,b=%d) R(r=%d,g=%d,b=%d) U(r=%d,g=%d,b=%d) D(r=%d,g=%d,b=%d)",
		ver, rlR, rlG, rlB, rrR, rrG, rrB, ruR, ruG, ruB, rdR, rdG, rdB)
	if absInt(rlR-rrR) > 45 || absInt(rlB-rrB) > 45 {
		t.Errorf("%s GSRAD left≠right (L r=%d b=%d, R r=%d b=%d) — stroke ramp is linear, not radial", ver, rlR, rlB, rrR, rrB)
	}
	if absInt(ruR-rdR) > 45 || absInt(ruB-rdB) > 45 {
		t.Errorf("%s GSRAD top≠bottom (U r=%d b=%d, D r=%d b=%d) — stroke ramp not radially symmetric", ver, ruR, ruB, rdR, rdB)
	}
	if absInt(rlR-ruR) > 45 || absInt(rlB-ruB) > 45 || absInt(rlG-ruG) > 45 {
		t.Errorf("%s GSRAD left≠top — stroke ramp not radially symmetric", ver)
	}
	// The ring sits mid-ramp (not pure red/blue), so neither channel dominates by
	// a wide margin — guards against a degenerate all-red / all-blue false pass.
	if absInt(rlR-rlB) > 170 {
		t.Errorf("%s GSRAD ring not mid-ramp (r=%d b=%d) — radius mapping off", ver, rlR, rlB)
	}

	// GSHL (centre 1560,540): radial ring + HiLite Length 70 / Angle 0 (+X). The
	// hotspot shifts right, so the ring's right mid is redder than its left mid
	// (one-axis split), while top and bottom stay matched.
	hlL_R, _, hlL_B := avgRGB(img, 1460, 540, win) // left mid
	hlR_R, _, hlR_B := avgRGB(img, 1660, 540, win) // right mid
	hlU_R, _, _ := avgRGB(img, 1560, 440, win)     // top mid
	hlD_R, _, _ := avgRGB(img, 1560, 640, win)     // bottom mid
	t.Logf("%s GSHL ring: L(r=%d,b=%d) R(r=%d,b=%d) U(r=%d) D(r=%d)",
		ver, hlL_R, hlL_B, hlR_R, hlR_B, hlU_R, hlD_R)
	if hlR_R-hlL_R < 40 {
		t.Errorf("%s GSHL right not redder than left (L r=%d, R r=%d) — highlight not shifting the hotspot on the stroke", ver, hlL_R, hlR_R)
	}
	if hlL_B-hlR_B < 30 {
		t.Errorf("%s GSHL left not bluer than right (L b=%d, R b=%d) — highlight axis wrong", ver, hlL_B, hlR_B)
	}
	if absInt(hlU_R-hlD_R) > 45 {
		t.Errorf("%s GSHL top≠bottom (U r=%d, D r=%d) — highlight broke the wrong axis", ver, hlU_R, hlD_R)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	for _, nm := range []string{"GSDIR", "GSRAD", "GSHL"} {
		if re.Compositions[0].LayerByName(nm) == nil {
			t.Fatalf("resaved: gradient-stroke layer %s missing", nm)
		}
	}
}

func TestMGGradStrokeGeom_AEShipGate_AE2020(t *testing.T) {
	runMGGradStrokeGeomGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGGradStrokeGeom_AEShipGate_AE2025(t *testing.T) {
	runMGGradStrokeGeomGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
