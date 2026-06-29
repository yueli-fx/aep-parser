// internal/aep/layer_3d_rotatey_shipgate_test.go
//
// AE ship gate for 3D Rotate Y (perspective tumble) on a from-scratch layer
// (roadmap priority 2). A 400px shape box is flipped to 3D and given RotateY=50°
// via the existing per-axis setter; a Go-created camera at [960,540,-1600]
// supplies perspective. Like Position Z, this needs ZERO new write code — the
// from-scratch transform template already carries the Rotate X/Y/Orientation
// slots, so after Reopen SetRotateY overwrites an existing cdat (length-
// preserving). Per delivery-contract red line 4 the gate renders the frame and
// asserts on pixels that the box becomes a foreshortened TRAPEZOID:
//
//	RotateY=50° turns the box's right edge away from the camera, so the
//	left (near) edge renders TALLER than the right (far) edge. A 2D layer (or a
//	dropped RotateY) would render an axis-aligned rectangle (equal heights).
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

// colHeight returns the count of near-white pixels in column x (the box's
// vertical extent there).
func colHeight(img image.Image, x int) int {
	b := img.Bounds()
	n := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		r, g, bb, _ := img.At(x, y).RGBA()
		if r>>8 >= 200 && g>>8 >= 200 && bb>>8 >= 200 {
			n++
		}
	}
	return n
}

// whiteColRange returns the leftmost and rightmost columns containing any
// near-white pixel.
func whiteColRange(img image.Image) (int, int) {
	b := img.Bounds()
	lo, hi := -1, -1
	for x := b.Min.X; x < b.Max.X; x++ {
		if colHeight(img, x) > 0 {
			if lo < 0 {
				lo = x
			}
			hi = x
		}
	}
	return lo, hi
}

func runLayer3DRotateYGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/3d_roty_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_3d_roty.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "ROTY", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	bg, _ := aep.NewShapeLayer(comp, "BG")
	bgr, _ := bg.RootGroup().AddRect()
	_ = bgr.SetSize([2]float64{2400, 1400})
	bgf, _ := bg.RootGroup().AddFill()
	_ = bgf.SetColor([4]float64{0.05, 0.05, 0.08, 1})
	_ = bg.Position().SetStaticValue([2]float64{960, 540})
	box, _ := aep.NewShapeLayer(comp, "BOX")
	r, _ := box.RootGroup().AddRect()
	_ = r.SetSize([2]float64{400, 400})
	f, _ := box.RootGroup().AddFill()
	_ = f.SetColor([4]float64{1, 1, 1, 1})
	_ = box.Position().SetStaticValue([2]float64{960, 540})
	if _, err := aep.NewCameraLayer(comp, "Cam"); err != nil {
		t.Fatalf("NewCameraLayer: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	c := rp.Compositions[0]
	b := c.LayerByName("BOX")
	if err := b.SetIs3D(true); err != nil {
		t.Fatalf("SetIs3D: %v", err)
	}
	if err := b.SetPosition([]float64{960, 540, 0}); err != nil {
		t.Fatalf("SetPosition: %v", err)
	}
	if err := b.SetRotateY(50); err != nil {
		t.Fatalf("SetRotateY: %v", err)
	}
	// Camera close in (strong perspective) so the RotateY trapezoid is pronounced.
	if err := c.LayerByName("Cam").SetPosition([]float64{960, 540, -800}); err != nil {
		t.Fatalf("Cam SetPosition: %v", err)
	}
	if err := aep.MoveToEnd(c.LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "roty_in.aep")
	resavedAEP := filepath.Join(tempDir, "roty_resaved.aep")
	doneFile := filepath.Join(tempDir, "roty.done")
	png := filepath.Join(tempDir, "roty.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"png":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(png))
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
	t.Logf("3d rotateY %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("3d rotateY %s ship gate FAIL:\n%s", ver, body)
	}

	fp, err := os.Open(png)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(fp)
	fp.Close()
	if err != nil {
		t.Fatalf("%s decode: %v", ver, err)
	}
	lo, hi := whiteColRange(img)
	if lo < 0 {
		t.Fatalf("%s box not rendered (no white pixels)", ver)
	}
	// Sample heights a few px inside each edge to dodge corner anti-aliasing.
	leftH := colHeight(img, lo+5)
	rightH := colHeight(img, hi-5)
	tall, short := leftH, rightH
	if rightH > leftH {
		tall, short = rightH, leftH
	}
	ratio := float64(tall) / float64(short+1)
	t.Logf("%s box cols [%d,%d] leftH=%d rightH=%d (tall/short %.2f)", ver, lo, hi, leftH, rightH, ratio)
	if leftH == 0 || rightH == 0 {
		t.Fatalf("%s degenerate box heights (left=%d right=%d)", ver, leftH, rightH)
	}
	// RotateY under perspective renders a TRAPEZOID — the near vertical edge is
	// magnified taller than the far edge (direction depends on rotation sign).
	// A flat rect (2D layer / dropped RotateY) would have leftH ≈ rightH (≈1.0).
	if ratio < 1.2 {
		t.Errorf("%s box edges near-symmetric (leftH=%d rightH=%d, ratio %.2f) — RotateY not applied / not 3D / no perspective", ver, leftH, rightH, ratio)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("BOX") == nil {
		t.Fatalf("resaved BOX missing")
	}
}

func TestLayer3DRotateY_AEShipGate_AE2020(t *testing.T) {
	runLayer3DRotateYGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestLayer3DRotateY_AEShipGate_AE2025(t *testing.T) {
	runLayer3DRotateYGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
