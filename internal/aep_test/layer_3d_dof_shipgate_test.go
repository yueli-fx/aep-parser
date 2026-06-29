// internal/aep/layer_3d_dof_shipgate_test.go
//
// AE ship gate for camera Depth of Field (roadmap priority 2: the 景深 of the
// "景深视差推拉镜" milestone). The camera-option setters (SetCameraDepthOfField /
// SetCameraFocusDistance / SetCameraAperture / SetCameraBlurLevel) existed but
// were only DOM-readback-checked — never render-verified that they actually
// produce a defocus blur (delivery-contract red line 4 gap). This gate builds
// two 3D boxes — a SHARP one at the camera's focus distance (z=-500) and a far
// BLUR one (z=+1400) — under a camera with DoF on + a wide aperture, then
// asserts on pixels that the far box's edges are blurred (a wide partial-grey
// transition band) while the focused box stays crisp.
//
// All setters run on the reopened camera (from-scratch camera options elide —
// see camera-light showcase); the scene is otherwise Go-built.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// grayBandLeft counts the partial-brightness pixels immediately left of a white
// run's start on scanline y — the defocus falloff width. A crisp edge has ~1-3
// (anti-alias only); a blurred edge has many. Scans left from start-1 while the
// red channel is in (dark, white) open range, stopping at full dark.
func grayBandLeft(img image.Image, y, start int) int {
	b := img.Bounds()
	const dark, white = 40, 215
	n := 0
	for x := start - 1; x >= b.Min.X; x-- {
		r, _, _, _ := img.At(x, y).RGBA()
		v := int(r >> 8)
		if v <= dark {
			break
		}
		if v < white {
			n++
		}
	}
	return n
}

func runLayer3DDoFGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/3d_dof_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_3d_dof.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "DOF", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	bg, _ := aep.NewShapeLayer(comp, "BG")
	bgr, _ := bg.RootGroup().AddRect()
	_ = bgr.SetSize([2]float64{2400, 1400})
	bgf, _ := bg.RootGroup().AddFill()
	_ = bgf.SetColor([4]float64{0.05, 0.05, 0.08, 1})
	_ = bg.Position().SetStaticValue([2]float64{960, 540})
	mk := func(name string, x float64) {
		l, _ := aep.NewShapeLayer(comp, name)
		r, _ := l.RootGroup().AddRect()
		_ = r.SetSize([2]float64{300, 300})
		f, _ := l.RootGroup().AddFill()
		_ = f.SetColor([4]float64{1, 1, 1, 1})
		_ = l.Position().SetStaticValue([2]float64{x, 540})
	}
	mk("SHARP", 640)
	mk("BLUR", 1280)
	if _, err := aep.NewCameraLayer(comp, "Cam"); err != nil {
		t.Fatalf("NewCameraLayer: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	c := rp.Compositions[0]
	sh := c.LayerByName("SHARP")
	if err := sh.SetIs3D(true); err != nil {
		t.Fatalf("SHARP SetIs3D: %v", err)
	}
	if err := sh.SetPosition([]float64{640, 540, -500}); err != nil {
		t.Fatalf("SHARP SetPosition: %v", err)
	}
	bl := c.LayerByName("BLUR")
	if err := bl.SetIs3D(true); err != nil {
		t.Fatalf("BLUR SetIs3D: %v", err)
	}
	if err := bl.SetPosition([]float64{1280, 540, 1400}); err != nil {
		t.Fatalf("BLUR SetPosition: %v", err)
	}
	cam := c.LayerByName("Cam")
	if err := cam.SetPosition([]float64{960, 540, -1800}); err != nil {
		t.Fatalf("Cam SetPosition: %v", err)
	}
	// Camera at z=-1800, SHARP at z=-500 → focus distance 1300 (the near plane).
	if err := cam.SetCameraDepthOfField(true); err != nil {
		t.Fatalf("SetCameraDepthOfField: %v", err)
	}
	if err := cam.SetCameraFocusDistance(1300); err != nil {
		t.Fatalf("SetCameraFocusDistance: %v", err)
	}
	if err := cam.SetCameraAperture(300); err != nil {
		t.Fatalf("SetCameraAperture: %v", err)
	}
	_ = cam.SetCameraBlurLevel(200)
	if err := aep.MoveToEnd(c.LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "dof_in.aep")
	resavedAEP := filepath.Join(tempDir, "dof_resaved.aep")
	doneFile := filepath.Join(tempDir, "dof.done")
	png := filepath.Join(tempDir, "dof.png")

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
	t.Logf("3d dof %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("3d dof %s ship gate FAIL:\n%s", ver, body)
	}

	f, err := os.Open(png)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode: %v", ver, err)
	}
	runs := whiteRuns(img, 540)
	if len(runs) != 2 {
		t.Fatalf("%s expected 2 boxes on scanline 540, got %d runs %v", ver, len(runs), runs)
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i][0] < runs[j][0] })
	// Left run = SHARP (x≈640 region), right run = BLUR.
	sharpBand := grayBandLeft(img, 540, runs[0][0])
	blurBand := grayBandLeft(img, 540, runs[1][0])
	t.Logf("%s SHARP edge band=%d  BLUR edge band=%d", ver, sharpBand, blurBand)
	if blurBand < 10 {
		t.Errorf("%s far box not blurred (band=%d) — DoF/aperture not producing defocus", ver, blurBand)
	}
	if blurBand < 3*sharpBand {
		t.Errorf("%s far box not clearly blurrier than near (blur=%d sharp=%d) — focus/DoF not differentiating depth", ver, blurBand, sharpBand)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("Cam") == nil {
		t.Fatalf("resaved Cam missing")
	}
}

func TestLayer3DDoF_AEShipGate_AE2020(t *testing.T) {
	runLayer3DDoFGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestLayer3DDoF_AEShipGate_AE2025(t *testing.T) {
	runLayer3DDoFGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
