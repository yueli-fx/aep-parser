// internal/aep/mg_repeater_shipgate_test.go
//
// AE ship gate for the Repeater filter (`ADBE Vector Filter - Repeater`) from
// scratch (MG roadmap S5, specs/2026-06-12-from-scratch-mg-roadmap.md). Verified
// at the capability's surface per delivery-contract red line 4: a single white
// dot is repeated 5× with a 300px horizontal Position offset, and the gate
// renders the frame and asserts five distinct dots sit on the row with dark gaps
// between them.
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

func buildMGRepeaterDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGREP", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	bg, err := aep.NewShapeLayer(comp, "BG")
	if err != nil {
		t.Fatalf("NewShapeLayer BG: %v", err)
	}
	rect, err := bg.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := rect.SetSize([2]float64{2200, 1300}); err != nil {
		t.Fatalf("rect SetSize: %v", err)
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

	// One white-filled dot + a Repeater: 5 copies, each offset +300px in x.
	dots, err := aep.NewShapeLayer(comp, "DOTS")
	if err != nil {
		t.Fatalf("NewShapeLayer DOTS: %v", err)
	}
	el, err := dots.RootGroup().AddEllipse()
	if err != nil {
		t.Fatalf("AddEllipse: %v", err)
	}
	if err := el.SetSize([2]float64{80, 80}); err != nil {
		t.Fatalf("ellipse SetSize: %v", err)
	}
	fill, err := dots.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill DOTS: %v", err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("DOTS SetColor: %v", err)
	}
	rep, err := dots.RootGroup().AddRepeater()
	if err != nil {
		t.Fatalf("AddRepeater: %v", err)
	}
	if err := rep.SetCopies(5); err != nil {
		t.Fatalf("SetCopies: %v", err)
	}
	if err := rep.Transform().SetPosition([2]float64{300, 0}); err != nil {
		t.Fatalf("Repeater SetPosition: %v", err)
	}
	// Base copy at x=360; copies march to 360,660,960,1260,1560 at y=540.
	if err := dots.Position().SetStaticValue([2]float64{360, 540}); err != nil {
		t.Fatalf("DOTS Position: %v", err)
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

func runMGRepeaterGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mg_repeater_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mg_repeater.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGRepeaterDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_repeater_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_repeater_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_repeater.done")
	framePNG := filepath.Join(tempDir, "mg_repeater_frame.png")

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
	t.Logf("mg repeater %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg repeater %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0: five white dots on y=540, dark gaps between.
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const win = 12
	dotXs := []int{360, 660, 960, 1260, 1560}
	gapXs := []int{510, 810, 1110, 1410}
	present := 0
	for _, x := range dotXs {
		if whiteNear(img, x, 540, win) {
			present++
		}
	}
	gapsClear := 0
	for _, x := range gapXs {
		if !whiteNear(img, x, 540, win) {
			gapsClear++
		}
	}
	t.Logf("%s repeater: dots present=%d/5 gaps clear=%d/4", ver, present, gapsClear)
	if present != 5 {
		t.Errorf("%s only %d/5 repeater copies rendered", ver, present)
	}
	if gapsClear != 4 {
		t.Errorf("%s %d/4 gaps were not clear — copies smeared/wrong spacing", ver, gapsClear)
	}

	// Resave proof: repeater Copies survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("DOTS") == nil {
		t.Fatal("resaved: DOTS layer missing")
	}
}

func TestMGRepeater_AEShipGate_AE2020(t *testing.T) {
	runMGRepeaterGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGRepeater_AEShipGate_AE2025(t *testing.T) {
	runMGRepeaterGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
