// internal/aep/mg_mask_pathkf_shipgate_test.go
//
// AE ship gate for SetMaskPathKeyframes (priority-4 mask,
// specs/2026-06-14-remaining-capability-roadmap.md). SetMaskPath reshapes a
// STATIC mask; SetMaskPathKeyframes writes an ANIMATED outline — N keyframes,
// each a path snapshot, emitted as the om-s time-table tdbs + one geometry shap
// per keyframe (byte-isomorphic to AE's own animated mask/shape path).
//
// Per delivery-contract red line 4, verified at the capability's surface: a white
// 600×600 shape rect masked by an ANIMATED rectangle that oscillates between the
// LEFT half (t=0) and the RIGHT half (t=2s). The gate renders both endpoint times
// and asserts the revealed region MOVED — at frame 0 the left half is white & the
// right dark, at t=2s it flips. A static mask cannot reveal opposite halves at two
// frames, so this proves AE honored the time table (not just value round-trip).
//
// SCALE (delivery-contract red line 2): SIX keyframes, not two — this crosses the
// lhd3 capacity-page boundary (ceil(6/4)=2 pages) under MASK strictness. The
// encodePathTimeTable page fix was only shape-path-gated to 6kf; mask is decoded
// eagerly by AE (capacity-field errors hard-crash 0::42 / corrupt), so the >4kf
// boundary needed its own mask gate. The two render endpoints land on keyframes
// (t=0 left, t=2s right); the intermediate oscillation just populates pages 1–2.
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

func buildMGMaskPathKfDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MASKPATHKF", 1920, 1080, 30, 5)
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
	// Mask coordinates are layer pixels with origin at the 600×600 card's centre.
	leftHalf := aep.BezierPath{
		Vertices: [][2]float64{{-300, -300}, {0, -300}, {0, 300}, {-300, 300}},
		Closed:   true,
	}
	rightHalf := aep.BezierPath{
		Vertices: [][2]float64{{0, -300}, {300, -300}, {300, 300}, {0, 300}},
		Closed:   true,
	}
	m, err := aep.AddMask(l, "M", leftHalf)
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	// Animate the outline: SIX keyframes oscillating left↔right (crosses the
	// lhd3 capacity page boundary, ceil(6/4)=2). Endpoints land on keyframes so
	// the render assertions stay exact: t=0 left, t=2s right.
	keys := []aep.MaskPathKey{
		{Time: 0, Path: leftHalf},
		{Time: 0.4, Path: rightHalf},
		{Time: 0.8, Path: leftHalf},
		{Time: 1.2, Path: rightHalf},
		{Time: 1.6, Path: leftHalf},
		{Time: 2, Path: rightHalf},
	}
	if err := aep.SetMaskPathKeyframes(l, m, keys); err != nil {
		t.Fatalf("SetMaskPathKeyframes: %v", err)
	}
	return rp
}

func runMGMaskPathKfGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mg_mask_pathkf_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mg_mask_pathkf.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGMaskPathKfDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_mask_pathkf_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_mask_pathkf_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_mask_pathkf.done")
	png0 := filepath.Join(tempDir, "mg_mask_pathkf_f0.png")
	png1 := filepath.Join(tempDir, "mg_mask_pathkf_f1.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"png0":%q,"png1":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(png0), toFwd(png1))
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
	t.Logf("mg mask pathkf %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg mask pathkf %s ship gate FAIL:\n%s", ver, body)
	}

	decode := func(path string) image.Image {
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("%s rendered frame %s missing: %v", ver, path, err)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			t.Fatalf("%s decode %s: %v", ver, path, err)
		}
		return img
	}
	img0 := decode(png0)
	img2 := decode(png1)

	const win = 8
	// Card spans comp x 660..1260. Left-centre ≈ 810, right-centre ≈ 1110.
	leftPt, rightPt := [2]int{810, 540}, [2]int{1110, 540}

	// Frame 0: left half revealed (white), right half masked (dark).
	f0LeftWhite := whiteNear(img0, leftPt[0], leftPt[1], win)
	f0RightDark := !whiteNear(img0, rightPt[0], rightPt[1], win)
	// t=2s: the outline has moved — right half revealed, left masked.
	f2LeftDark := !whiteNear(img2, leftPt[0], leftPt[1], win)
	f2RightWhite := whiteNear(img2, rightPt[0], rightPt[1], win)

	t.Logf("%s mask pathkf: f0[L white=%v R dark=%v] f2[L dark=%v R white=%v]",
		ver, f0LeftWhite, f0RightDark, f2LeftDark, f2RightWhite)
	if !f0LeftWhite {
		t.Errorf("%s frame0 left-centre not white — mask doesn't reveal left half at t=0", ver)
	}
	if !f0RightDark {
		t.Errorf("%s frame0 right-centre not dark — mask reveals right half at t=0 (should be left only)", ver)
	}
	if !f2LeftDark {
		t.Errorf("%s t=2s left-centre not dark — mask did not move off the left half (animation ignored)", ver)
	}
	if !f2RightWhite {
		t.Errorf("%s t=2s right-centre not white — mask did not move to the right half (animation ignored)", ver)
	}

	// Resave proof: AE re-encodes all 6 mask-shape keyframes (no page truncation).
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	l := re.Compositions[0].LayerByName("CARD")
	if l == nil || len(l.Masks) != 1 {
		t.Fatal("resaved: CARD mask missing")
	}
	if got := len(l.Masks[0].PathKeyframes); got != 6 {
		t.Errorf("resaved mask path keyframes = %d, want 6 (animated, 2 capacity pages)", got)
	}
}

func TestMGMaskPathKf_AEShipGate_AE2020(t *testing.T) {
	runMGMaskPathKfGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGMaskPathKf_AEShipGate_AE2025(t *testing.T) {
	runMGMaskPathKfGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
