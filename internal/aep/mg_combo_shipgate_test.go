// internal/aep/mg_combo_shipgate_test.go
//
// AE ship gate for MG roadmap S6 — the end-to-end COMBINATION (capstone). Per
// delivery-contract red line 4 + the "composition is an independent deliverable"
// clause, a single all-Go-built project must combine every shipped slice and
// render correctly as a whole: a precomp (child badge) + Trim Paths (the badge's
// half-ring) + ease keyframes (the eased MOVER dot) + a Repeater (the DOTS row).
// The gate renders the mid-frame (t=2s) and pixel-asserts all four features
// coexist in one AE-accepted, rendered frame.
//
//   - MOVER (ease): amber dot, Position 2 keyframes (200,250)→(1700,250) with a
//     slow-out ease; at t=2s it must lag the linear midpoint (x≈950).
//   - BADGE (precomp + trim): a nested comp whose only content is an ellipse +
//     white stroke + Trim End=50 — centred at (960,540), its RIGHT arc visible
//     and LEFT side cut.
//   - DOTS (repeater): one white dot repeated 4× at +200px → x=660..1260, y=900.
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
	"github.com/yueli-fx/aep-parser/internal/codec"
)

func buildMGComboDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	scene, err := aep.NewComposition(p, "MGX_Scene", 1920, 1080, 30, 4)
	if err != nil {
		t.Fatalf("NewComposition scene: %v", err)
	}
	badge, err := aep.NewComposition(p, "MGX_Badge", 1920, 1080, 30, 4)
	if err != nil {
		t.Fatalf("NewComposition badge: %v", err)
	}

	// Child badge: ellipse + white stroke + Trim End=50 (half-ring), centred.
	// No BG → transparent everywhere else so it composites over the scene.
	bShape, err := aep.NewShapeLayer(badge, "Ring")
	if err != nil {
		t.Fatalf("NewShapeLayer Ring: %v", err)
	}
	el, err := bShape.RootGroup().AddEllipse()
	if err != nil {
		t.Fatalf("badge AddEllipse: %v", err)
	}
	if err := el.SetSize([2]float64{300, 300}); err != nil {
		t.Fatalf("badge ellipse SetSize: %v", err)
	}
	st, err := bShape.RootGroup().AddStroke()
	if err != nil {
		t.Fatalf("badge AddStroke: %v", err)
	}
	if err := st.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("badge stroke SetColor: %v", err)
	}
	if err := st.SetWidth(24); err != nil {
		t.Fatalf("badge stroke SetWidth: %v", err)
	}
	tr, err := bShape.RootGroup().AddTrim()
	if err != nil {
		t.Fatalf("badge AddTrim: %v", err)
	}
	if err := tr.SetEnd(50); err != nil {
		t.Fatalf("badge trim SetEnd: %v", err)
	}
	if err := bShape.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("badge Position: %v", err)
	}

	// Scene BG (dark).
	bg, err := aep.NewShapeLayer(scene, "BG")
	if err != nil {
		t.Fatalf("NewShapeLayer BG: %v", err)
	}
	bgRect, err := bg.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("BG AddRect: %v", err)
	}
	if err := bgRect.SetSize([2]float64{2200, 1300}); err != nil {
		t.Fatalf("BG SetSize: %v", err)
	}
	bgFill, err := bg.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("BG AddFill: %v", err)
	}
	if err := bgFill.SetColor([4]float64{0.05, 0.05, 0.08, 1}); err != nil {
		t.Fatalf("BG SetColor: %v", err)
	}
	if err := bg.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("BG Position: %v", err)
	}

	// MOVER: eased amber dot (top row, y=250).
	mover, err := aep.NewShapeLayer(scene, "MOVER")
	if err != nil {
		t.Fatalf("NewShapeLayer MOVER: %v", err)
	}
	mEl, err := mover.RootGroup().AddEllipse()
	if err != nil {
		t.Fatalf("MOVER AddEllipse: %v", err)
	}
	if err := mEl.SetSize([2]float64{72, 72}); err != nil {
		t.Fatalf("MOVER ellipse SetSize: %v", err)
	}
	mFill, err := mover.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("MOVER AddFill: %v", err)
	}
	if err := mFill.SetColor([4]float64{1.0, 0.55, 0.1, 1}); err != nil {
		t.Fatalf("MOVER SetColor: %v", err)
	}
	slowOut := codec.TemporalEase{Speed: 0, Influence: 0.9}
	if err := mover.Position().AddKeyframeWithEase(0, [2]float64{200, 250}, codec.TemporalEase{}, slowOut); err != nil {
		t.Fatalf("MOVER kf0: %v", err)
	}
	if err := mover.Position().AddKeyframeWithEase(4, [2]float64{1700, 250}, codec.TemporalEase{Speed: 0, Influence: 0.1}, codec.TemporalEase{}); err != nil {
		t.Fatalf("MOVER kf1: %v", err)
	}

	// DOTS: one white dot repeated 4× (bottom row, y=900).
	dots, err := aep.NewShapeLayer(scene, "DOTS")
	if err != nil {
		t.Fatalf("NewShapeLayer DOTS: %v", err)
	}
	dEl, err := dots.RootGroup().AddEllipse()
	if err != nil {
		t.Fatalf("DOTS AddEllipse: %v", err)
	}
	if err := dEl.SetSize([2]float64{60, 60}); err != nil {
		t.Fatalf("DOTS ellipse SetSize: %v", err)
	}
	dFill, err := dots.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("DOTS AddFill: %v", err)
	}
	if err := dFill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("DOTS SetColor: %v", err)
	}
	rep, err := dots.RootGroup().AddRepeater()
	if err != nil {
		t.Fatalf("DOTS AddRepeater: %v", err)
	}
	if err := rep.SetCopies(4); err != nil {
		t.Fatalf("DOTS SetCopies: %v", err)
	}
	if err := rep.Transform().SetPosition([2]float64{200, 0}); err != nil {
		t.Fatalf("DOTS repeater SetPosition: %v", err)
	}
	if err := dots.Position().SetStaticValue([2]float64{660, 900}); err != nil {
		t.Fatalf("DOTS Position: %v", err)
	}

	// Reopen so the comps are parsed, then nest the badge into the scene.
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	var sceneR, badgeR *aep.Composition
	for _, c := range rp.Compositions {
		switch c.Name {
		case "MGX_Scene":
			sceneR = c
		case "MGX_Badge":
			badgeR = c
		}
	}
	if sceneR == nil || badgeR == nil {
		t.Fatalf("reopened comps missing")
	}
	if _, err := aep.NewPrecompLayer(sceneR, badgeR, "BADGE"); err != nil {
		t.Fatalf("NewPrecompLayer: %v", err)
	}
	if err := aep.MoveToEnd(sceneR.LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}
	return rp
}

func runMGComboGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_combo_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_combo.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGComboDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_combo_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_combo_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_combo.done")
	framePNG := filepath.Join(tempDir, "mg_combo_frame.png")

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
	t.Logf("mg combo %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg combo %s ship gate FAIL:\n%s", ver, body)
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

	// ease: amber MOVER on row 250 must lag the linear midpoint (950).
	moverX := scanRowForColor(img, 250, [3]uint8{255, 140, 26}, 40)
	// trim+precomp: badge half-ring centred at (960,540), right arc present, left cut.
	const win = 12
	badgeRight := whiteNear(img, 1110, 540, win)
	badgeTop := whiteNear(img, 960, 390, win)
	badgeBottom := whiteNear(img, 960, 690, win)
	badgeLeft := whiteNear(img, 810, 540, win)
	// repeater: 4 dots on row 900.
	dotsPresent := 0
	for _, x := range []int{660, 860, 1060, 1260} {
		if whiteNear(img, x, 900, win) {
			dotsPresent++
		}
	}
	gapsClear := 0
	for _, x := range []int{760, 960} {
		if !whiteNear(img, x, 900, win) {
			gapsClear++
		}
	}
	t.Logf("%s combo: MOVER x=%d | badge R=%v T=%v B=%v L=%v | dots=%d/4 gaps=%d/2",
		ver, moverX, badgeRight, badgeTop, badgeBottom, badgeLeft, dotsPresent, gapsClear)

	if moverX < 0 || moverX >= 800 {
		t.Errorf("%s MOVER ease: x=%d, want a lagged value < 800 (linear midpoint is 950)", ver, moverX)
	}
	if !badgeRight || !badgeTop || !badgeBottom {
		t.Errorf("%s BADGE precomp+trim arc missing (R=%v T=%v B=%v)", ver, badgeRight, badgeTop, badgeBottom)
	}
	if badgeLeft {
		t.Errorf("%s BADGE trim did not cut the left side", ver)
	}
	if dotsPresent != 4 {
		t.Errorf("%s only %d/4 repeater dots rendered", ver, dotsPresent)
	}
	if gapsClear != 2 {
		t.Errorf("%s repeater gaps not clear (%d/2)", ver, gapsClear)
	}

	// Resave proof: the whole combined project survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	var sc *aep.Composition
	for _, c := range re.Compositions {
		if c.Name == "MGX_Scene" {
			sc = c
		}
	}
	if sc == nil || len(sc.Layers) != 4 {
		t.Fatalf("resaved scene missing or wrong layer count (%v)", sc)
	}
	badgeLayer := sc.LayerByName("BADGE")
	if badgeLayer == nil || badgeLayer.SourceComposition() == nil || badgeLayer.SourceComposition().Name != "MGX_Badge" {
		t.Errorf("resaved BADGE precomp source lost")
	}
}

func TestMGCombo_AEShipGate_AE2020(t *testing.T) {
	runMGComboGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGCombo_AEShipGate_AE2025(t *testing.T) {
	runMGComboGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
