// internal/aep/mg_star_shipgate_test.go
//
// AE ship gate for the PolyStar shape (`ADBE Vector Shape - Star`) from scratch
// (MG roadmap S5, specs/2026-06-12-from-scratch-mg-roadmap.md). Verified at the
// capability's surface per delivery-contract red line 4: a 5-point white star
// (OuterRadius=250, InnerRadius=100). The gate renders the frame and asserts the
// 5 tip directions are filled (white) at a mid radius while the 5 gap directions
// between them are empty (dark) — the alternating signature of a star, which a
// plain polygon/circle would not produce.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"image"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func buildMGStarDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGSTAR", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// Dark BG so the star reads as white-on-dark unambiguously.
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

	// A 5-point white star, OuterRadius=250 / InnerRadius=100, centered at 960,540.
	star, err := aep.NewShapeLayer(comp, "STAR")
	if err != nil {
		t.Fatalf("NewShapeLayer STAR: %v", err)
	}
	st, err := star.RootGroup().AddStar()
	if err != nil {
		t.Fatalf("AddStar: %v", err)
	}
	if err := st.SetPoints(5); err != nil {
		t.Fatalf("SetPoints: %v", err)
	}
	if err := st.SetOuterRadius(250); err != nil {
		t.Fatalf("SetOuterRadius: %v", err)
	}
	if err := st.SetInnerRadius(100); err != nil {
		t.Fatalf("SetInnerRadius: %v", err)
	}
	fill, err := star.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill STAR: %v", err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("STAR SetColor: %v", err)
	}
	if err := star.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("STAR Position: %v", err)
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

func runMGStarGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mg_star_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mg_star.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGStarDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_star_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_star_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_star.done")
	framePNG := filepath.Join(tempDir, "mg_star_frame.png")

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
	t.Logf("mg star %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg star %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0: at a mid radius (200px, between inner 100 and
	// outer 250), the 5 tip directions are white (inside the arms) and the 5 gap
	// directions between them are dark (outside the star). Rotation=0 → tip 0
	// points straight up (screen angle -90°); tips every 72°, gaps offset 36°.
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const cx, cy, r, win = 960, 540, 200, 10
	sampleAt := func(deg float64) (int, int) {
		rad := deg * math.Pi / 180
		return cx + int(math.Round(r*math.Cos(rad))), cy + int(math.Round(r*math.Sin(rad)))
	}
	tipsWhite, gapsDark := 0, 0
	for k := 0; k < 5; k++ {
		tx, ty := sampleAt(-90 + 72*float64(k))
		if whiteNear(img, tx, ty, win) {
			tipsWhite++
		}
		gx, gy := sampleAt(-90 + 36 + 72*float64(k))
		if !whiteNear(img, gx, gy, win) {
			gapsDark++
		}
	}
	centerWhite := whiteNear(img, cx, cy, win)
	t.Logf("%s star: tips white=%d/5 gaps dark=%d/5 centerWhite=%v", ver, tipsWhite, gapsDark, centerWhite)
	if tipsWhite != 5 {
		t.Errorf("%s only %d/5 star tips filled — star not rendered", ver, tipsWhite)
	}
	if gapsDark != 5 {
		t.Errorf("%s only %d/5 gaps dark — not a star (polygon/circle would fill the gaps)", ver, gapsDark)
	}
	if !centerWhite {
		t.Errorf("%s center not white — star interior not filled", ver)
	}

	// Resave proof: the star survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("STAR") == nil {
		t.Fatal("resaved: STAR layer missing")
	}
}

func TestMGStar_AEShipGate_AE2020(t *testing.T) {
	runMGStarGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGStar_AEShipGate_AE2025(t *testing.T) {
	runMGStarGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
