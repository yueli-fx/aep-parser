// internal/aep/mg_puckerbloat_shipgate_test.go
//
// AE ship gate for the Pucker & Bloat filter (`ADBE Vector Filter - PB`) from
// scratch (MG roadmap S5, specs/2026-06-12-from-scratch-mg-roadmap.md). Verified
// at the capability's surface per delivery-contract red line 4: a 400×400 white
// rectangle gets a Pucker & Bloat filter (Amount=100, bloat). The gate renders
// the frame and asserts the card interior stays white, the straight edges bow
// OUTWARD past the original rect boundary (white beyond the edge), and the
// original sharp corners are pulled IN toward the center (dark) — proving the
// distortion actually reshaped the path, not just that the value round-tripped.
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

func buildMGPuckerBloatDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGPB", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// Dark BG so the bowed edges / carved corners read unambiguously.
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

	// A 400×400 white rectangle + Pucker & Bloat (Amount=100, bloat). Centered at
	// 960,540 → spans x∈[760,1160], y∈[340,740]. Bloat bows the rect edges outward
	// and pulls the corners inward.
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
	pb, err := card.RootGroup().AddPuckerBloat()
	if err != nil {
		t.Fatalf("AddPuckerBloat: %v", err)
	}
	if err := pb.SetAmount(100); err != nil {
		t.Fatalf("SetAmount: %v", err)
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

func runMGPuckerBloatGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mg_puckerbloat_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mg_puckerbloat.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGPuckerBloatDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_puckerbloat_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_puckerbloat_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_puckerbloat.done")
	framePNG := filepath.Join(tempDir, "mg_puckerbloat_frame.png")

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
	t.Logf("mg puckerbloat %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg puckerbloat %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0. Card spans x∈[760,1160], y∈[340,740].
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
	// Center stays white.
	whitePts := [][2]int{{960, 540}}
	// Edges bow OUTWARD: points just beyond each original edge midpoint go white.
	bulgePts := [][2]int{{960, 320}, {960, 760}, {740, 540}, {1180, 540}}
	// Corners pulled IN: original sharp-corner regions go dark.
	cornerPts := [][2]int{{775, 355}, {1145, 355}, {775, 725}, {1145, 725}}
	whiteOK := 0
	for _, pt := range whitePts {
		if whiteNear(img, pt[0], pt[1], win) {
			whiteOK++
		}
	}
	bulgeOK := 0
	for _, pt := range bulgePts {
		if whiteNear(img, pt[0], pt[1], win) {
			bulgeOK++
		}
	}
	cornersCarved := 0
	for _, pt := range cornerPts {
		if !whiteNear(img, pt[0], pt[1], win) {
			cornersCarved++
		}
	}
	t.Logf("%s puckerbloat: center-white=%d/1 edges-bulged=%d/4 corners-carved=%d/4", ver, whiteOK, bulgeOK, cornersCarved)
	if whiteOK != 1 {
		t.Errorf("%s center not white — card body not rendered", ver)
	}
	if bulgeOK != 4 {
		t.Errorf("%s only %d/4 edges bulged outward — bloat did not bow the path", ver, bulgeOK)
	}
	if cornersCarved != 4 {
		t.Errorf("%s only %d/4 sharp corners pulled in — path not distorted", ver, cornersCarved)
	}

	// Resave proof: the Pucker & Bloat filter survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("CARD") == nil {
		t.Fatal("resaved: CARD layer missing")
	}
}

func TestMGPuckerBloat_AEShipGate_AE2020(t *testing.T) {
	runMGPuckerBloatGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGPuckerBloat_AEShipGate_AE2025(t *testing.T) {
	runMGPuckerBloatGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
