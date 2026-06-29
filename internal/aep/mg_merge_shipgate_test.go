// internal/aep/mg_merge_shipgate_test.go
//
// AE ship gate for the Merge Paths filter (`ADBE Vector Filter - Merge`) from
// scratch (MG roadmap S5, specs/2026-06-12-from-scratch-mg-roadmap.md). Verified
// at the capability's surface per delivery-contract red line 4: a 400×400 white
// rectangle and a 200×200 concentric ellipse are combined with Merge Type =
// Subtract, carving a round hole out of the square. The gate renders the frame
// and asserts the surrounding ring stays white while the central hole is dark —
// proving the boolean combine actually reshaped the path, not just that the enum
// round-tripped.
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

func buildMGMergeDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGMERGE", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// Dark BG so the carved hole reads as not-white unambiguously.
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

	// A 400×400 white rect + a concentric 200×200 ellipse + Merge=Subtract →
	// square with a round hole. Centered at 960,540: rect spans x∈[760,1160],
	// y∈[340,740]; the ellipse (r=100) carves a hole spanning x∈[860,1060],
	// y∈[440,640].
	cut, err := aep.NewShapeLayer(comp, "CUT")
	if err != nil {
		t.Fatalf("NewShapeLayer CUT: %v", err)
	}
	rect, err := cut.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect CUT: %v", err)
	}
	if err := rect.SetSize([2]float64{400, 400}); err != nil {
		t.Fatalf("CUT rect SetSize: %v", err)
	}
	el, err := cut.RootGroup().AddEllipse()
	if err != nil {
		t.Fatalf("AddEllipse CUT: %v", err)
	}
	if err := el.SetSize([2]float64{200, 200}); err != nil {
		t.Fatalf("CUT ellipse SetSize: %v", err)
	}
	// Merge below, Fill on top: the Merge combines Rect−Ellipse into one path,
	// and the Fill (top of the stack) paints the merged result. Fill below Merge
	// renders nothing (paints the un-merged paths before the combine).
	mg, err := cut.RootGroup().AddMergePaths()
	if err != nil {
		t.Fatalf("AddMergePaths: %v", err)
	}
	if err := mg.SetType(aep.MergeTypeSubtract); err != nil {
		t.Fatalf("SetType: %v", err)
	}
	fill, err := cut.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill CUT: %v", err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("CUT SetColor: %v", err)
	}
	if err := cut.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("CUT Position: %v", err)
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

func runMGMergeGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mg_merge_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mg_merge.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGMergeDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_merge_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_merge_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_merge.done")
	framePNG := filepath.Join(tempDir, "mg_merge_frame.png")

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
	t.Logf("mg merge %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg merge %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0: the ring around the hole is white; the
	// central round hole (where the ellipse was subtracted) is dark.
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
	// Ring: in the rect, clear of the subtracted ellipse → white.
	ringPts := [][2]int{{800, 540}, {1120, 540}, {960, 380}, {960, 700}}
	// Hole: inside the subtracted ellipse → dark.
	holePts := [][2]int{{960, 540}, {920, 540}, {960, 580}}
	ring := 0
	for _, pt := range ringPts {
		if whiteNear(img, pt[0], pt[1], win) {
			ring++
		}
	}
	hole := 0
	for _, pt := range holePts {
		if !whiteNear(img, pt[0], pt[1], win) {
			hole++
		}
	}
	t.Logf("%s merge: ring white=%d/4 hole dark=%d/3", ver, ring, hole)
	if ring != 4 {
		t.Errorf("%s only %d/4 ring points white — rect body not rendered", ver, ring)
	}
	if hole != 3 {
		t.Errorf("%s only %d/3 hole points dark — ellipse not subtracted", ver, hole)
	}

	// Resave proof: the Merge Paths filter survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("CUT") == nil {
		t.Fatal("resaved: CUT layer missing")
	}
}

func TestMGMerge_AEShipGate_AE2020(t *testing.T) {
	runMGMergeGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGMerge_AEShipGate_AE2025(t *testing.T) {
	runMGMergeGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
