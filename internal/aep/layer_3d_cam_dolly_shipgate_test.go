// internal/aep/layer_3d_cam_dolly_shipgate_test.go
//
// AE ship gate for VISIBLE 3D (roadmap priority 2): a from-scratch shape layer
// flipped to 3D (Is3D bit only — z stays 0) genuinely responds to a Go-created
// camera. The JSX dollies the camera Z between two renders; this asserts on the
// rendered pixels (red line 4) that the box scales with camera distance:
//
//   - NEAR (camera z=-700):  the 300px box renders LARGE (~426px wide).
//   - FAR  (camera z=-2400): it renders SMALL (~124px wide).
//
// A 2D layer ignores the camera and would render at a constant width — the
// >2× near/far size ratio is the proof the layer is truly 3D and camera-driven.
// This is the visible counterpart to TestLayer3DEnable (DOM-only); together they
// close "图层 3D flag" with both an acceptance and a render proof.
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

// boxWidth3D returns the width (px) of the white box band on scanline y.
func boxWidth3D(img image.Image, y int) int {
	b := img.Bounds()
	minX, maxX := -1, -1
	for x := b.Min.X; x < b.Max.X; x++ {
		r, g, bb, _ := img.At(x, y).RGBA()
		if r>>8 >= 200 && g>>8 >= 200 && bb>>8 >= 200 {
			if minX < 0 {
				minX = x
			}
			maxX = x
		}
	}
	if minX < 0 {
		return 0
	}
	return maxX - minX + 1
}

func runLayer3DCamDollyGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/3d_cam_dolly_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_3d_cam_dolly.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "CAM3D", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	// BG (2D, full-frame dark) — ignores the camera, stays a backdrop.
	bg, _ := aep.NewShapeLayer(comp, "BG")
	bgr, _ := bg.RootGroup().AddRect()
	_ = bgr.SetSize([2]float64{2200, 1300})
	bgf, _ := bg.RootGroup().AddFill()
	_ = bgf.SetColor([4]float64{0.05, 0.05, 0.08, 1})
	_ = bg.Position().SetStaticValue([2]float64{960, 540})
	// BOX (white 300px square) — flipped to 3D after Reopen.
	box, _ := aep.NewShapeLayer(comp, "BOX")
	br, _ := box.RootGroup().AddRect()
	_ = br.SetSize([2]float64{300, 300})
	bf, _ := box.RootGroup().AddFill()
	_ = bf.SetColor([4]float64{1, 1, 1, 1})
	_ = box.Position().SetStaticValue([2]float64{960, 540})
	// Camera on top (affects 3D layers below it).
	if _, err := aep.NewCameraLayer(comp, "Cam"); err != nil {
		t.Fatalf("NewCameraLayer: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if err := rp.Compositions[0].LayerByName("BOX").SetIs3D(true); err != nil {
		t.Fatalf("SetIs3D: %v", err)
	}
	if err := aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "cam3d_in.aep")
	doneFile := filepath.Join(tempDir, "cam3d.done")
	pngNear := filepath.Join(tempDir, "cam3d_near.png")
	pngFar := filepath.Join(tempDir, "cam3d_far.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"pngNear":%q,"pngFar":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(pngNear), toFwd(pngFar))
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
	t.Logf("3d cam-dolly %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("3d cam-dolly %s ship gate FAIL:\n%s", ver, body)
	}

	decodeWidth := func(path string) int {
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("%s rendered frame missing (%s): %v", ver, filepath.Base(path), err)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			t.Fatalf("%s decode %s: %v", ver, filepath.Base(path), err)
		}
		return boxWidth3D(img, 540)
	}
	near := decodeWidth(pngNear)
	far := decodeWidth(pngFar)
	t.Logf("%s box width: near=%d far=%d (ratio %.2f)", ver, near, far, float64(near)/float64(far+1))

	if near < 380 {
		t.Errorf("%s near box width=%d, want >380 (3D box should render large up close)", ver, near)
	}
	if far == 0 {
		t.Errorf("%s far box width=0 — box vanished / layer dropped", ver)
	}
	if far > 180 {
		t.Errorf("%s far box width=%d, want <180 (3D box should shrink at distance)", ver, far)
	}
	if far > 0 && near < 2*far {
		t.Errorf("%s near/far width ratio %.2f too small — box not scaling with camera (2D, not 3D?)", ver, float64(near)/float64(far))
	}
}

func TestLayer3DCamDolly_AEShipGate_AE2020(t *testing.T) {
	runLayer3DCamDollyGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestLayer3DCamDolly_AEShipGate_AE2025(t *testing.T) {
	runLayer3DCamDollyGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
