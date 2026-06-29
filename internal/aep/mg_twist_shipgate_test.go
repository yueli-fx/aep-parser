// internal/aep/mg_twist_shipgate_test.go
//
// AE ship gate for the Twist filter (`ADBE Vector Filter - Twist`) from scratch
// (MG roadmap S5, specs/2026-06-12-from-scratch-mg-roadmap.md). Verified at the
// capability's surface per delivery-contract red line 4: a 400×400 white
// rectangle gets a Twist filter (Angle=150). The gate renders the frame and
// asserts the path was actually twisted into a pinwheel — the original
// axis-aligned edges are gone and the shape's corners swung off-axis — proving
// the distortion reshaped the path, not just that the value round-tripped.
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

func buildMGTwistDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGTW", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// Dark BG so the twisted silhouette reads unambiguously.
	bg, err := aep.NewShapeLayer(comp, "BG")
	if err != nil {
		t.Fatalf("NewShapeLayer BG: %v", err)
	}
	bgRect, err := bg.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect BG: %v", err)
	}
	if err := bgRect.SetSize([2]float64{2200, 1300}); err != nil {
		t.Fatalf("BG rect SetSize: %v", err)
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

	// A 400×400 white rectangle + Twist (Angle=150). Centered at 960,540 →
	// originally spans x∈[760,1160], y∈[340,740]. Twist rotates the path more in
	// the center than at the edges, bowing the straight edges into spirals and
	// swinging the corners off the axis-aligned diagonal.
	card, err := aep.NewShapeLayer(comp, "CARD")
	if err != nil {
		t.Fatalf("NewShapeLayer CARD: %v", err)
	}
	rect, err := card.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect CARD: %v", err)
	}
	if err := rect.SetSize([2]float64{400, 400}); err != nil {
		t.Fatalf("CARD rect SetSize: %v", err)
	}
	fill, err := card.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill CARD: %v", err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("CARD SetColor: %v", err)
	}
	tw, err := card.RootGroup().AddTwist()
	if err != nil {
		t.Fatalf("AddTwist: %v", err)
	}
	if err := tw.SetAngle(150); err != nil {
		t.Fatalf("SetAngle: %v", err)
	}
	if err := card.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("CARD Position: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if err := aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}
	return rp
}

func runMGTwistGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_twist_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_twist.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGTwistDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_twist_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_twist_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_twist.done")
	framePNG := filepath.Join(tempDir, "mg_twist_frame.png")

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
	t.Logf("mg twist %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg twist %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0. Original card spans x∈[760,1160], y∈[340,740].
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}

	// Diagnostic grid over the card region so the swirl can be eyeballed in the
	// log alongside the pixel assertions.
	var sb strings.Builder
	for y := 300; y <= 780; y += 30 {
		fmt.Fprintf(&sb, "y=%4d ", y)
		for x := 700; x <= 1220; x += 20 {
			if whiteNear(img, x, y, 4) {
				sb.WriteByte('#')
			} else {
				sb.WriteByte('.')
			}
		}
		sb.WriteByte('\n')
	}
	t.Logf("%s twist silhouette:\n%s", ver, sb.String())

	const win = 10
	// Center stays white (twist pins the center).
	if !whiteNear(img, 960, 540, win) {
		t.Errorf("%s center not white — card body not rendered", ver)
	}
	// The original axis-aligned corners are vacated by the twist: the four
	// original sharp-corner regions go dark (an un-twisted square would keep
	// them white).
	cornerPts := [][2]int{{775, 355}, {1145, 355}, {775, 725}, {1145, 725}}
	cornersVacated := 0
	for _, pt := range cornerPts {
		if !whiteNear(img, pt[0], pt[1], win) {
			cornersVacated++
		}
	}
	// Twist has a handedness: the swirl is NOT mirror-symmetric about the card's
	// vertical center axis (x=960). A plain / puckered / bloated square is
	// mirror-symmetric, so a broken twist that renders the untwisted square
	// would score ~0 here. Count grid samples whose left/right mirror disagrees.
	asym := 0
	for y := 340; y <= 740; y += 20 {
		for x := 720; x < 960; x += 20 {
			if whiteNear(img, x, y, 6) != whiteNear(img, 1920-x, y, 6) {
				asym++
			}
		}
	}
	t.Logf("%s twist: corners-vacated=%d/4 mirror-asym=%d", ver, cornersVacated, asym)
	if cornersVacated != 4 {
		t.Errorf("%s only %d/4 original corners vacated — path not twisted", ver, cornersVacated)
	}
	if asym < 20 {
		t.Errorf("%s mirror-asym=%d too low — swirl handedness absent (twist no-op?)", ver, asym)
	}

	// Resave proof: the Twist filter survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("CARD") == nil {
		t.Fatal("resaved: CARD layer missing")
	}
}

func TestMGTwist_AEShipGate_AE2020(t *testing.T) {
	runMGTwistGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGTwist_AEShipGate_AE2025(t *testing.T) {
	runMGTwistGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
