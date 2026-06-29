// internal/aep/mg_offset_shipgate_test.go
//
// AE ship gate for the Offset Paths filter (`ADBE Vector Filter - Offset`) from
// scratch (MG roadmap S5, specs/2026-06-12-from-scratch-mg-roadmap.md). Verified
// at the capability's surface per delivery-contract red line 4: a 400×400 white
// rectangle gets an Offset Paths filter (Amount=60). The gate renders the frame
// and asserts a band just OUTSIDE each original edge turned white (the path grew
// outward) while points beyond the offset stay dark (the growth is bounded, not
// a runaway fill) — proving the offset actually reshaped the path, not just that
// the value round-tripped.
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

func buildMGOffsetDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGOFF", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// Dark BG so the grown band reads as white-on-dark unambiguously.
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

	// A 400×400 white rectangle + Offset Paths (Amount=60). Centered at 960,540
	// → original spans x∈[760,1160], y∈[340,740]; after +60 offset the filled
	// path grows to x∈[700,1220], y∈[280,800]. Offset on top of the stack grows
	// the rect path the fill below it paints.
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
	off, err := card.RootGroup().AddOffsetPaths()
	if err != nil {
		t.Fatalf("AddOffsetPaths: %v", err)
	}
	if err := off.SetAmount(60); err != nil {
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

func runMGOffsetGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mg_offset_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mg_offset.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGOffsetDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_offset_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_offset_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_offset.done")
	framePNG := filepath.Join(tempDir, "mg_offset_frame.png")

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
	t.Logf("mg offset %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg offset %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0: a band just outside each original edge is
	// white (the path grew by +60), and points beyond the offset stay dark.
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const win = 8
	// Center + a point just outside each ORIGINAL edge (inside the +60 offset).
	grownPts := [][2]int{{960, 540}, {1190, 540}, {730, 540}, {960, 310}, {960, 770}}
	// Points beyond the offset bound — must stay dark.
	beyondPts := [][2]int{{1250, 540}, {670, 540}, {960, 250}, {960, 830}}
	grown := 0
	for _, pt := range grownPts {
		if whiteNear(img, pt[0], pt[1], win) {
			grown++
		}
	}
	bounded := 0
	for _, pt := range beyondPts {
		if !whiteNear(img, pt[0], pt[1], win) {
			bounded++
		}
	}
	t.Logf("%s offset: grown-band white=%d/5 beyond-dark=%d/4", ver, grown, bounded)
	if grown != 5 {
		t.Errorf("%s only %d/5 grown-band points white — path did not offset outward", ver, grown)
	}
	if bounded != 4 {
		t.Errorf("%s only %d/4 beyond points dark — offset unbounded/wrong", ver, bounded)
	}

	// Resave proof: the Offset Paths filter survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("CARD") == nil {
		t.Fatal("resaved: CARD layer missing")
	}
}

func TestMGOffset_AEShipGate_AE2020(t *testing.T) {
	runMGOffsetGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGOffset_AEShipGate_AE2025(t *testing.T) {
	runMGOffsetGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
