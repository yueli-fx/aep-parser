// internal/aep/shape_gradient_stroke_anim_shipgate_test.go
//
// AE ship gate for ANIMATED gradient STROKE color stops (roadmap priority-1
// leftover: gradient FILL animated stops were gated, the STROKE host was not).
// GradientStrokeNode.AddGradientKeyframe reuses the exact `ADBE Vector Grad
// Colors` stream + animateGradientStops helper the fill uses (the stream is
// identical on fill and stroke), so this is a small symmetric extension.
//
// Per delivery-contract red line 4, verified at the capability's surface: a
// from-scratch shape layer = a 900px rect outlined by an 80px gradient stroke
// whose stops are keyframed (kf0@0s = R/B/G left→right, kf1@1s = G/R/B). Sampling
// the TOP stroke band (horizontal ramp: left edge = offset 0%, centre = 50%,
// right = 100%):
//
//	t=0:  left=RED   centre=BLUE  right=GREEN
//	t=1s: left=GREEN centre=RED   right=BLUE
//
// The left-edge RED→GREEN swap between two frames of the SAME stroke is the
// animation proof; resave confirms the stream keeps 2 keyframes.
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

func runGradientStrokeAnimGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/grad_stroke_anim_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_grad_stroke_anim.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildGradStrokeAnim(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "grad_stroke_anim_in.aep")
	resavedAEP := filepath.Join(tempDir, "grad_stroke_anim_resaved.aep")
	doneFile := filepath.Join(tempDir, "grad_stroke_anim.done")
	png0 := filepath.Join(tempDir, "grad_stroke_anim_t0.png")
	png1 := filepath.Join(tempDir, "grad_stroke_anim_t1.png")

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
	t.Logf("grad-stroke-anim %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("grad-stroke-anim %s ship gate FAIL:\n%s", ver, body)
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
	// Rect 900px centred at 960,540 → top edge at y=90; the 80px stroke band spans
	// y∈[50,130]. Sample the top band centreline (y=90) at left/centre/right x,
	// dodging the corners (x∈[510,1410]). Horizontal ramp: x→offset.
	const y = 90
	const lx, cx, rx = 580, 960, 1340
	img0 := decode(png0)
	img1 := decode(png1)
	l0, c0, r0 := dominantChannel(img0, lx, y), dominantChannel(img0, cx, y), dominantChannel(img0, rx, y)
	l1, c1, r1 := dominantChannel(img1, lx, y), dominantChannel(img1, cx, y), dominantChannel(img1, rx, y)
	t.Logf("%s t=0 L=%c C=%c R=%c | t=1 L=%c C=%c R=%c", ver, l0, c0, r0, l1, c1, r1)

	if l0 != 'R' || c0 != 'B' || r0 != 'G' {
		t.Errorf("%s t=0 stroke ramp = %c/%c/%c, want R/B/G", ver, l0, c0, r0)
	}
	if l1 != 'G' || c1 != 'R' || r1 != 'B' {
		t.Errorf("%s t=1 stroke ramp = %c/%c/%c, want G/R/B (stops did not animate)", ver, l1, c1, r1)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("GRADSTROKE") == nil {
		t.Fatalf("resaved GRADSTROKE layer missing")
	}
}

func TestGradientStrokeAnim_AEShipGate_AE2020(t *testing.T) {
	runGradientStrokeAnimGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestGradientStrokeAnim_AEShipGate_AE2025(t *testing.T) {
	runGradientStrokeAnimGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
