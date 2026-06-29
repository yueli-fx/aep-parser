// internal/aep/shape_gradient_anim_shipgate_test.go
//
// AE ship gate for ANIMATED gradient color stops (roadmap priority 1, the
// fixture-blocked item — unblocked by a user-authored AE oracle,
// test_data/fixtures/v2_2_gradient_anim_src.aep). Builds a from-scratch shape layer with
// a gradient fill whose stops are keyframed (kf0@0s = R/B/G left→right, kf1@1s =
// G/R/B), then renders BOTH frames and asserts on actual pixels (red line 4)
// that the stop colours swap over time:
//
//	horizontal ramp, left edge = offset 0%, centre = 50%, right edge = 100%.
//	t=0:  left=RED   centre=BLUE  right=GREEN
//	t=1s: left=GREEN centre=RED   right=BLUE
//
// The left-edge RED→GREEN swap between two frames of the SAME layer is the
// animation proof. Resave confirms the Grad Colors stream keeps 2 keyframes.
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

// dominantChannel reports which of R/G/B is the largest at (px,py), as 'R'/'G'/'B'.
func dominantChannel(img image.Image, px, py int) byte {
	r, g, b, _ := img.At(px, py).RGBA()
	r8, g8, b8 := r>>8, g>>8, b>>8
	if r8 >= g8 && r8 >= b8 {
		return 'R'
	}
	if g8 >= r8 && g8 >= b8 {
		return 'G'
	}
	return 'B'
}

func runGradientAnimGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/grad_anim_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_grad_anim.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildGradAnim(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "grad_anim_in.aep")
	resavedAEP := filepath.Join(tempDir, "grad_anim_resaved.aep")
	doneFile := filepath.Join(tempDir, "grad_anim.done")
	png0 := filepath.Join(tempDir, "grad_anim_t0.png")
	png1 := filepath.Join(tempDir, "grad_anim_t1.png")

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
	t.Logf("grad-anim %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("grad-anim %s ship gate FAIL:\n%s", ver, body)
	}

	decode := func(path string) image.Image {
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("%s rendered frame missing (%s): %v", ver, filepath.Base(path), err)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			t.Fatalf("%s decode %s: %v", ver, filepath.Base(path), err)
		}
		return img
	}
	// Rect 1000px centred at 960 → spans x∈[460,1460], y=540. Sample inside the
	// edges (left 4%, centre 50%, right 96%) to dodge edge anti-aliasing.
	const y = 540
	const lx, cx, rx = 500, 960, 1420
	img0 := decode(png0)
	img1 := decode(png1)
	l0, c0, r0 := dominantChannel(img0, lx, y), dominantChannel(img0, cx, y), dominantChannel(img0, rx, y)
	l1, c1, r1 := dominantChannel(img1, lx, y), dominantChannel(img1, cx, y), dominantChannel(img1, rx, y)
	t.Logf("%s t=0 L=%c C=%c R=%c | t=1 L=%c C=%c R=%c", ver, l0, c0, r0, l1, c1, r1)

	// t=0: R/B/G across the ramp.
	if l0 != 'R' || c0 != 'B' || r0 != 'G' {
		t.Errorf("%s t=0 ramp = %c/%c/%c, want R/B/G", ver, l0, c0, r0)
	}
	// t=1s: G/R/B — the stops swapped.
	if l1 != 'G' || c1 != 'R' || r1 != 'B' {
		t.Errorf("%s t=1 ramp = %c/%c/%c, want G/R/B (stops did not animate)", ver, l1, c1, r1)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("GRAD") == nil {
		t.Fatalf("resaved GRAD layer missing")
	}
}

func TestGradientAnim_AEShipGate_AE2020(t *testing.T) {
	runGradientAnimGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestGradientAnim_AEShipGate_AE2025(t *testing.T) {
	runGradientAnimGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
