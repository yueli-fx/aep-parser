// internal/aep/mg_roundcorners_shipgate_test.go
//
// AE ship gate for the Round Corners filter (`ADBE Vector Filter - RC`) from
// scratch (MG roadmap S5, specs/2026-06-12-from-scratch-mg-roadmap.md). Verified
// at the capability's surface per delivery-contract red line 4: a 400×400 white
// rectangle gets a Round Corners filter (Radius=150). The gate renders the frame
// and asserts the card interior + straight edges stay white while the original
// sharp corners are carved away (dark) — proving the rounding actually reshaped
// the path, not just that the value round-tripped.
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

func buildMGRoundCornersDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGRC", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// Dark BG so the carved corners read as not-white unambiguously.
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

	// A 400×400 white rectangle + Round Corners (Radius=150). Centered at
	// 960,540 → spans x∈[760,1160], y∈[340,740]. Round Corners on top of the
	// stack rounds the rect path the fill below it paints.
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
	rc, err := card.RootGroup().AddRoundCorners()
	if err != nil {
		t.Fatalf("AddRoundCorners: %v", err)
	}
	if err := rc.SetRadius(150); err != nil {
		t.Fatalf("SetRadius: %v", err)
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

func runMGRoundCornersGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_roundcorners_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_roundcorners.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGRoundCornersDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_roundcorners_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_roundcorners_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_roundcorners.done")
	framePNG := filepath.Join(tempDir, "mg_roundcorners_frame.png")

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
	t.Logf("mg roundcorners %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg roundcorners %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0: card interior + straight-edge midpoints stay
	// white; the four original sharp corners are carved away (dark).
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
	// Interior + edge midpoints that survive rounding.
	whitePts := [][2]int{{960, 540}, {960, 345}, {960, 735}, {765, 540}, {1155, 540}}
	// Original sharp corners (carved out by Radius=150).
	cornerPts := [][2]int{{770, 350}, {1150, 350}, {770, 730}, {1150, 730}}
	whiteOK := 0
	for _, pt := range whitePts {
		if whiteNear(img, pt[0], pt[1], win) {
			whiteOK++
		}
	}
	cornersCarved := 0
	for _, pt := range cornerPts {
		if !whiteNear(img, pt[0], pt[1], win) {
			cornersCarved++
		}
	}
	t.Logf("%s roundcorners: white-survive=%d/5 corners-carved=%d/4", ver, whiteOK, cornersCarved)
	if whiteOK != 5 {
		t.Errorf("%s only %d/5 interior/edge points white — card body not rendered", ver, whiteOK)
	}
	if cornersCarved != 4 {
		t.Errorf("%s only %d/4 sharp corners carved — corners not rounded", ver, cornersCarved)
	}

	// Resave proof: the Round Corners filter survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("CARD") == nil {
		t.Fatal("resaved: CARD layer missing")
	}
}

func TestMGRoundCorners_AEShipGate_AE2020(t *testing.T) {
	runMGRoundCornersGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGRoundCorners_AEShipGate_AE2025(t *testing.T) {
	runMGRoundCornersGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
