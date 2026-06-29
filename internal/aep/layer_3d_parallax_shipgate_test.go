// internal/aep/layer_3d_parallax_shipgate_test.go
//
// AE ship gate for per-layer 3D Z = PARALLAX (roadmap priority 2, the visible
// payoff of 3D layers). Two from-scratch shape boxes of the SAME nominal size
// (300px) are flipped to 3D and placed at different depths via SetPosition's 3rd
// component (NEAR z=-800 toward the camera, FAR z=+1200 away); a Go-created
// camera sits at [960,540,-1800]. The whole scene is Go-built — no JSX mutation.
// Per delivery-contract red line 4 the gate renders the frame and asserts on
// pixels that depth drives apparent size:
//
//   NEAR (z=-800) renders LARGE (~300px wide), FAR (z=+1200) renders SMALL
//   (~99px) — a >2× ratio between two identically-sized layers is the proof
//   per-layer Z is written and AE honours it (a 2D layer would ignore Z).
//
// Per-layer Z needs ZERO new write code: the from-scratch shape Position is
// already a 3-component spatial slot (lowerTransformVec2Spatial encode3D), so
// SetPosition([x,y,z]) on the reopened 3D layer is length-preserving. SetIs3D
// makes AE treat the layer as 3D (see layer-3d-enable-bit-materializes.md).
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

// whiteRuns returns the [start,end] x-extents of contiguous near-white runs on
// scanline y (used to measure each box independently).
func whiteRuns(img image.Image, y int) [][2]int {
	b := img.Bounds()
	var runs [][2]int
	inRun, st := false, 0
	for x := b.Min.X; x < b.Max.X; x++ {
		r, g, bb, _ := img.At(x, y).RGBA()
		white := r>>8 >= 200 && g>>8 >= 200 && bb>>8 >= 200
		if white && !inRun {
			inRun, st = true, x
		} else if !white && inRun {
			inRun = false
			runs = append(runs, [2]int{st, x - 1})
		}
	}
	if inRun {
		runs = append(runs, [2]int{st, b.Max.X - 1})
	}
	return runs
}

func runLayer3DParallaxGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/3d_parallax_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_3d_parallax.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "CAM3DP", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	bg, _ := aep.NewShapeLayer(comp, "BG")
	bgr, _ := bg.RootGroup().AddRect()
	_ = bgr.SetSize([2]float64{2400, 1400})
	bgf, _ := bg.RootGroup().AddFill()
	_ = bgf.SetColor([4]float64{0.05, 0.05, 0.08, 1})
	_ = bg.Position().SetStaticValue([2]float64{960, 540})
	mkBox := func(name string) {
		l, _ := aep.NewShapeLayer(comp, name)
		r, _ := l.RootGroup().AddRect()
		_ = r.SetSize([2]float64{300, 300})
		f, _ := l.RootGroup().AddFill()
		_ = f.SetColor([4]float64{1, 1, 1, 1})
		_ = l.Position().SetStaticValue([2]float64{960, 540})
	}
	mkBox("NEAR")
	mkBox("FAR")
	if _, err := aep.NewCameraLayer(comp, "Cam"); err != nil {
		t.Fatalf("NewCameraLayer: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	c := rp.Compositions[0]
	for _, b := range []struct {
		name string
		pos  []float64
	}{
		{"NEAR", []float64{650, 540, -800}},
		{"FAR", []float64{1270, 540, 1200}},
	} {
		l := c.LayerByName(b.name)
		if err := l.SetIs3D(true); err != nil {
			t.Fatalf("%s SetIs3D: %v", b.name, err)
		}
		if err := l.SetPosition(b.pos); err != nil {
			t.Fatalf("%s SetPosition 3-comp: %v", b.name, err)
		}
	}
	if err := c.LayerByName("Cam").SetPosition([]float64{960, 540, -1800}); err != nil {
		t.Fatalf("Cam SetPosition: %v", err)
	}
	if err := aep.MoveToEnd(c.LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "parallax_in.aep")
	resavedAEP := filepath.Join(tempDir, "parallax_resaved.aep")
	doneFile := filepath.Join(tempDir, "parallax.done")
	png := filepath.Join(tempDir, "parallax.png")

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
	t.Logf("3d parallax %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("3d parallax %s ship gate FAIL:\n%s", ver, body)
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
	for _, r := range runs {
		t.Logf("%s white run x=[%d,%d] width=%d", ver, r[0], r[1], r[1]-r[0]+1)
	}
	if len(runs) != 2 {
		t.Fatalf("%s expected 2 boxes on scanline 540, got %d runs %v", ver, len(runs), runs)
	}
	// Left run = NEAR (x≈650 projected), right run = FAR. Sort by start x.
	sort.Slice(runs, func(i, j int) bool { return runs[i][0] < runs[j][0] })
	nearW := runs[0][1] - runs[0][0] + 1
	farW := runs[1][1] - runs[1][0] + 1
	t.Logf("%s NEAR width=%d FAR width=%d (ratio %.2f)", ver, nearW, farW, float64(nearW)/float64(farW))
	if nearW < 2*farW {
		t.Errorf("%s NEAR/FAR width ratio too small (near=%d far=%d) — per-layer Z not driving depth", ver, nearW, farW)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	rc := re.Compositions[0]
	if rc.LayerByName("NEAR") == nil || rc.LayerByName("FAR") == nil {
		t.Fatalf("resaved NEAR/FAR missing")
	}
}

func TestLayer3DParallax_AEShipGate_AE2020(t *testing.T) {
	runLayer3DParallaxGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestLayer3DParallax_AEShipGate_AE2025(t *testing.T) {
	runLayer3DParallaxGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
