// internal/aep/animate_effect_param_shipgate_test.go
//
// AE ship gate for AnimateEffectParam — keyframing a 1D-scalar effect parameter
// from scratch (the biggest "效果" gap: InsertKeyframe refuses a static param,
// so animated blur / animated Slider-Control rigs were impossible). The Go
// keyframe container is byte-identical to an AE-saved animated
// Gaussian-Blur-Blurriness fixture (RE-verified); this proves AE EVALUATES it
// (red line 4 — Go round-trip ≠ AE eval).
//
// A white 400px square with a Gaussian Blur whose Blurriness is KEYFRAMED
// 0@0s → 100@2s. At t=0 blur=0 (sharp edge); at t=2.5s blur=100 (the white
// bleeds ~100px past the original edge). Sampling 20px OUTSIDE the left edge:
// dark at t=0 (sharp), bright at t=2.5s (the keyframes grew the blur).
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

func buildAnimEffectDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "ANIMEFFECT", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	bg, _ := aep.NewShapeLayer(comp, "BG")
	bgRect, _ := bg.RootGroup().AddRect()
	bgRect.SetSize([2]float64{2200, 1300})
	bgFill, _ := bg.RootGroup().AddFill()
	bgFill.SetColor([4]float64{0, 0, 0, 1})
	bg.Position().SetStaticValue([2]float64{960, 540})

	sq, _ := aep.NewShapeLayer(comp, "SQ")
	r, _ := sq.RootGroup().AddRect()
	r.SetSize([2]float64{400, 400})
	fill, _ := sq.RootGroup().AddFill()
	fill.SetColor([4]float64{1, 1, 1, 1})
	sq.Position().SetStaticValue([2]float64{960, 540})

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if err := aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd: %v", err)
	}
	l := rp.Compositions[0].LayerByName("SQ")
	fx, err := aep.AddEffect(l, "ADBE Gaussian Blur 2")
	if err != nil {
		t.Fatalf("AddEffect: %v", err)
	}
	if _, err := aep.AnimateEffectParam(l, fx, "ADBE Gaussian Blur 2-0001",
		[]aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 2, Value: 100}}); err != nil {
		t.Fatalf("AnimateEffectParam: %v", err)
	}
	return rp
}

func runAnimEffectGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/anim_effect_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_anim_effect.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildAnimEffectDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "anim_effect_in.aep")
	resavedAEP := filepath.Join(tempDir, "anim_effect_resaved.aep")
	doneFile := filepath.Join(tempDir, "anim_effect.done")
	png0 := filepath.Join(tempDir, "anim_effect_t0.png")
	png1 := filepath.Join(tempDir, "anim_effect_t1.png")

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
	t.Logf("anim-effect %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("anim-effect %s ship gate FAIL:\n%s", ver, body)
	}

	decode := func(path string) image.Image {
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("%s frame missing (%s): %v", ver, filepath.Base(path), err)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			t.Fatalf("%s decode %s: %v", ver, filepath.Base(path), err)
		}
		return img
	}
	img0 := decode(png0)
	img1 := decode(png1)
	// Square 400px centred at 960 → left edge x=760. Sample 20px outside (x=740).
	const sx, sy = 740, 540
	out0 := luminance(img0, sx, sy)
	out1 := luminance(img1, sx, sy)
	c0, c1 := luminance(img0, 960, 540), luminance(img1, 960, 540)
	t.Logf("%s outside-edge lum t=0=%d t=2.5=%d | centre t=0=%d t=2.5=%d", ver, out0, out1, c0, c1)

	if out0 > 40 {
		t.Errorf("%s t=0 outside-edge lum=%d, want dark (<40) — blur keyframe@0 should be 0", ver, out0)
	}
	if out1 < 60 {
		t.Errorf("%s t=2.5 outside-edge lum=%d, want bright (>60) — keyframes did not grow the blur", ver, out1)
	}
	if c0 < 200 || c1 < 200 {
		t.Errorf("%s square centre not bright (t0=%d t2.5=%d) — unexpected", ver, c0, c1)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	rl := re.Compositions[0].LayerByName("SQ")
	if rl == nil || len(rl.Effects) != 1 {
		t.Fatalf("resaved SQ effect missing")
	}
	var blur *aep.Property
	for _, pp := range rl.Effects[0].Parameters {
		if pp.MatchName == "ADBE Gaussian Blur 2-0001" {
			blur = pp
		}
	}
	if blur == nil || len(blur.Keyframes) != 2 {
		t.Errorf("resaved Blurriness keyframes = %v, want 2", blur)
	}
}

func TestAnimEffect_AEShipGate_AE2020(t *testing.T) {
	runAnimEffectGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestAnimEffect_AEShipGate_AE2025(t *testing.T) {
	runAnimEffectGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
