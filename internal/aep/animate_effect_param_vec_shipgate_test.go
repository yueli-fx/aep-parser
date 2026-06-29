// internal/aep/animate_effect_param_vec_shipgate_test.go
//
// AE render gate for animated COLOR and POINT effect parameters
// (AnimateEffectParamVec) — the color/2D-3D-point counterpart of the scalar
// AnimateEffectParam gate. Per delivery-contract red line 4, both are verified
// at the render surface across time, on a 100% Go-built file:
//
//   FILLANIM comp: white solid + Fill effect, Fill Color animated
//     red([A,R,G,B]=255,255,0,0) @0s -> blue(255,0,0,255) @2s.
//     frame t=0 center = RED, frame t=2 center = BLUE.
//
//   RAMPANIM comp: solid + Gradient Ramp (Start Color white, End Color black,
//     Radial), Start-of-Ramp POINT animated left(0.25,0.5) @0s -> right
//     (0.75,0.5) @2s. The bright radial centre sweeps L->R:
//     t=0 left bright / right dark, t=2 left dark / right bright.
//
// Sampling two times proves AE evaluated each multi-component param's keyframes
// over time (not just that the stream round-tripped). Resave reopen proves the
// animated streams survive AE's re-encode.
//
// Gated by AE_SHIP_GATE. Uses test_data/verify_anim_effect_vec.jsx.
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

func buildVecAnimRenderDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)

	// FILLANIM: white solid + Fill effect with animated Color.
	fc, err := aep.NewComposition(p, "FILLANIM", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition FILLANIM: %v", err)
	}
	if _, err := aep.NewSolidLayer(fc, "F", 1920, 1080, [3]float64{1, 1, 1}); err != nil {
		t.Fatalf("NewSolidLayer F: %v", err)
	}

	// RAMPANIM: solid + Gradient Ramp with animated Start-of-Ramp point.
	rc, err := aep.NewComposition(p, "RAMPANIM", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition RAMPANIM: %v", err)
	}
	if _, err := aep.NewSolidLayer(rc, "R", 1920, 1080, [3]float64{0.5, 0.5, 0.5}); err != nil {
		t.Fatalf("NewSolidLayer R: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}

	// Fill color animation.
	fl := compByName(rp, "FILLANIM").LayerByName("F")
	ffx, err := aep.AddEffect(fl, aep.EffectFill)
	if err != nil {
		t.Fatalf("AddEffect(Fill): %v", err)
	}
	if _, err := aep.AnimateEffectParamVec(fl, ffx, "ADBE Fill-0002", []aep.VectorKeyframe{
		{Time: 0, Value: []float64{255, 255, 0, 0}}, // red
		{Time: 2, Value: []float64{255, 0, 0, 255}}, // blue
	}); err != nil {
		t.Fatalf("AnimateEffectParamVec(Fill Color): %v", err)
	}

	// Gradient Ramp: white->black radial, start point swept L->R. Solid source
	// is full-comp, so point on-disk units = fraction of 1920x1080.
	rl := compByName(rp, "RAMPANIM").LayerByName("R")
	rfx, err := aep.AddEffect(rl, aep.EffectGradientRamp)
	if err != nil {
		t.Fatalf("AddEffect(Ramp): %v", err)
	}
	set := func(mn string, v any) {
		t.Helper()
		if _, err := aep.SetEffectParam(rl, rfx, mn, v); err != nil {
			t.Fatalf("SetEffectParam(%s): %v", mn, err)
		}
	}
	set("ADBE Ramp-0002", []float64{255, 255, 255, 255})  // Start Color white
	set("ADBE Ramp-0004", []float64{255, 0, 0, 0})        // End Color black
	set("ADBE Ramp-0005", 2.0)                            // Ramp Shape = Radial
	set("ADBE Ramp-0003", []float64{0.5, 1.0})            // End of Ramp = (960,1080) → radius
	if _, err := aep.AnimateEffectParamVec(rl, rfx, "ADBE Ramp-0001", []aep.VectorKeyframe{
		{Time: 0, Value: []float64{0.25, 0.5}}, // left  (480,540)
		{Time: 2, Value: []float64{0.75, 0.5}}, // right (1440,540)
	}); err != nil {
		t.Fatalf("AnimateEffectParamVec(Ramp Start): %v", err)
	}
	return rp
}

func runVecAnimGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/anim_effect_vec_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_anim_effect_vec.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildVecAnimRenderDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "anim_vec_in.aep")
	resavedAEP := filepath.Join(tempDir, "anim_vec_resaved.aep")
	doneFile := filepath.Join(tempDir, "anim_vec.done")
	fillT0 := filepath.Join(tempDir, "fill_t0.png")
	fillT2 := filepath.Join(tempDir, "fill_t2.png")
	rampT0 := filepath.Join(tempDir, "ramp_t0.png")
	rampT2 := filepath.Join(tempDir, "ramp_t2.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,`+
		`"fill_t0":%q,"fill_t2":%q,"ramp_t0":%q,"ramp_t2":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP),
		toFwd(fillT0), toFwd(fillT2), toFwd(rampT0), toFwd(rampT2))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 300)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("anim vec %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("anim vec %s ship gate FAIL:\n%s", ver, body)
	}

	const win = 12
	// Fill color: red @ t0, blue @ t2 at comp centre.
	r0, g0, b0 := avgRGBFile(t, fillT0, 960, 540, win)
	r2, g2, b2 := avgRGBFile(t, fillT2, 960, 540, win)
	t.Logf("%s FILL t0=(%d,%d,%d) t2=(%d,%d,%d)", ver, r0, g0, b0, r2, g2, b2)
	if !(r0 > 150 && b0 < 100) {
		t.Errorf("%s FILL t0 = (%d,%d,%d), want RED (R>150,B<100)", ver, r0, g0, b0)
	}
	if !(b2 > 150 && r2 < 100) {
		t.Errorf("%s FILL t2 = (%d,%d,%d), want BLUE (B>150,R<100)", ver, r2, g2, b2)
	}

	// Ramp point: bright radial centre sweeps L->R.
	lt0, _, _ := avgRGBFile(t, rampT0, 480, 540, win)
	rt0, _, _ := avgRGBFile(t, rampT0, 1440, 540, win)
	lt2, _, _ := avgRGBFile(t, rampT2, 480, 540, win)
	rt2, _, _ := avgRGBFile(t, rampT2, 1440, 540, win)
	t.Logf("%s RAMP t0 L=%d R=%d  t2 L=%d R=%d", ver, lt0, rt0, lt2, rt2)
	if !(lt0 > 150 && rt0 < 90) {
		t.Errorf("%s RAMP t0 L=%d R=%d, want L bright(>150) R dark(<90)", ver, lt0, rt0)
	}
	if !(rt2 > 150 && lt2 < 90) {
		t.Errorf("%s RAMP t2 L=%d R=%d, want L dark(<90) R bright(>150)", ver, lt2, rt2)
	}

	// Resave proof: both animated effect params survive AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	checkAnimated := func(comp, layer, param string, comps int) {
		l := compByName(re, comp).LayerByName(layer)
		if l == nil {
			t.Errorf("resaved: %s/%s missing", comp, layer)
			return
		}
		for _, e := range l.Effects {
			for _, pr := range e.Parameters {
				if pr.MatchName == param {
					if len(pr.Keyframes) < 2 {
						t.Errorf("resaved %s: %d keyframes, want >=2", param, len(pr.Keyframes))
					}
					if pr.Components != comps {
						t.Errorf("resaved %s: Components=%d, want %d", param, pr.Components, comps)
					}
					return
				}
			}
		}
		t.Errorf("resaved: param %s not found", param)
	}
	checkAnimated("FILLANIM", "F", "ADBE Fill-0002", 4)
	checkAnimated("RAMPANIM", "R", "ADBE Ramp-0001", 2)
}

func avgRGBFile(t *testing.T, path string, px, py, win int) (int, int, int) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("rendered frame missing: %v", err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return avgRGB(img, px, py, win)
}

func TestAnimEffectVec_AEShipGate_AE2020(t *testing.T) {
	runVecAnimGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestAnimEffectVec_AEShipGate_AE2025(t *testing.T) {
	runVecAnimGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
