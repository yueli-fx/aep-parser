// internal/aep/mg_mask_path_shipgate_test.go
//
// AE ship gate for SetMaskPath (priority-4 mask,
// specs/2026-06-14-remaining-capability-roadmap.md). AddMask could only create a
// mask; SetMaskPath reshapes an existing mask's outline (variable-length om-s
// rebuild, reusing makeMaskShapeOmS so non-4-vertex paths get the mask-strictness
// lhd3/shph patching).
//
// Per delivery-contract red line 4, verified at the capability's surface: a white
// 600×600 shape rect masked first with a full-rect (reveals everything) then
// RESHAPED via SetMaskPath into a 3-vertex triangle. The gate asserts the
// revealed region is triangular — points inside the triangle are white, while the
// ORIGINAL rectangle's corners (inside the old full-rect mask, outside the
// triangle) are now dark — proving SetMaskPath actually rewrote AE's clip path
// (vertex count 4→3), not just that the value round-tripped.
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

func buildMGMaskPathDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MASKPATH", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

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

	card, err := aep.NewShapeLayer(comp, "CARD")
	if err != nil {
		t.Fatalf("NewShapeLayer CARD: %v", err)
	}
	r, err := card.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect CARD: %v", err)
	}
	if err := r.SetSize([2]float64{600, 600}); err != nil {
		t.Fatalf("CARD rect size: %v", err)
	}
	fill, err := card.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill CARD: %v", err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("CARD fill color: %v", err)
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

	l := rp.Compositions[0].LayerByName("CARD")
	if l == nil {
		t.Fatal("CARD layer missing after reopen")
	}
	// Mask created as a full 600×600 rect (covers the whole white rect)...
	fullRect := aep.BezierPath{
		Vertices: [][2]float64{{-300, -300}, {300, -300}, {300, 300}, {-300, 300}},
		Closed:   true,
	}
	m, err := aep.AddMask(l, "M", fullRect)
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	// ...then reshaped into a triangle (apex top-centre, base along the bottom).
	tri := aep.BezierPath{
		Vertices: [][2]float64{{0, -260}, {260, 260}, {-260, 260}},
		Closed:   true,
	}
	if err := aep.SetMaskPath(l, m, tri); err != nil {
		t.Fatalf("SetMaskPath: %v", err)
	}
	return rp
}

func runMGMaskPathGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_mask_path_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_mask_path.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGMaskPathDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_mask_path_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_mask_path_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_mask_path.done")
	framePNG := filepath.Join(tempDir, "mg_mask_path_frame.png")

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
	t.Logf("mg mask path %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg mask path %s ship gate FAIL:\n%s", ver, body)
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
	const win = 8
	// Inside the triangle (centre + lower-centre) → revealed white.
	insidePts := [][2]int{{960, 540}, {960, 760}}
	// Original rectangle corners (inside the old full-rect mask region
	// 660..1260 × 240..840, outside the triangle) → now masked dark. White here
	// would mean the rect mask was never reshaped.
	cornerPts := [][2]int{{660, 300}, {1260, 300}, {760, 360}}

	inside := 0
	for _, pt := range insidePts {
		if whiteNear(img, pt[0], pt[1], win) {
			inside++
		}
	}
	corner := 0
	for _, pt := range cornerPts {
		if !whiteNear(img, pt[0], pt[1], win) {
			corner++
		}
	}
	t.Logf("%s mask path: inside-triangle white=%d/2 old-corner dark=%d/3", ver, inside, corner)
	if inside != 2 {
		t.Errorf("%s only %d/2 inside-triangle points white — reshaped mask doesn't reveal", ver, inside)
	}
	if corner != 3 {
		t.Errorf("%s only %d/3 old-corner points dark — mask still a rectangle (SetMaskPath didn't reshape)", ver, corner)
	}

	// Resave proof: the reshaped path (3 vertices) survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	l := re.Compositions[0].LayerByName("CARD")
	if l == nil || len(l.Masks) != 1 {
		t.Fatal("resaved: CARD mask missing")
	}
	if got := len(l.Masks[0].Vertices); got != 3 {
		t.Errorf("resaved mask vertices = %d, want 3 (triangle)", got)
	}
}

func TestMGMaskPath_AEShipGate_AE2020(t *testing.T) {
	runMGMaskPathGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGMaskPath_AEShipGate_AE2025(t *testing.T) {
	runMGMaskPathGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
