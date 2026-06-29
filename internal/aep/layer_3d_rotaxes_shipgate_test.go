// internal/aep/layer_3d_rotaxes_shipgate_test.go
//
// AE render-pixel ship gate closing out roadmap priority 2 (3D layers): the
// remaining per-axis rotation channels Rotate X, Orientation, and Rotate Z on a
// from-scratch 3D layer. Like Rotate Y (layer_3d_rotatey_shipgate_test.go) these
// need ZERO new write code — the from-scratch transform template already carries
// every 3D rotation/orientation slot, so after Reopen the existing setter
// overwrites a cdat length-preservingly. Per delivery-contract red line 4 each
// axis renders the frame and asserts on pixels:
//
//   Rotate X=50°       box rotates about the horizontal axis; under perspective
//                      the top and bottom EDGES foreshorten to different WIDTHS
//                      (trapezoid on its side). A flat rect has topW ≈ botW.
//   Orientation Y=50°  same Y-axis tumble as Rotate Y: the near vertical edge
//                      renders TALLER than the far edge (trapezoid). leftH ≠ rightH.
//   Rotate Z=45°       in-plane spin: the axis-aligned square becomes a DIAMOND,
//                      so the center column is far taller than the near-edge
//                      columns. A non-rotated square has near-constant column heights.
//
// DISK-CACHE GOTCHA (see mg_text_style_shipgate_test.go / re-fixture.md): a
// from-scratch single-comp project always gets comp.id=1, and saveFrameToPng keys
// its persistent on-disk frame cache on (comp.id, render-time) ACROSS processes.
// Three sibling axes rendered at t=0 would collide and false-green. Defenses, both
// applied: each axis renders at a UNIQUE time and clearAEDiskCache runs before each
// AE launch.
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

// rowWidth returns the count of near-white pixels in row y (the box's horizontal
// extent there). Companion to colHeight (layer_3d_rotatey_shipgate_test.go).
func rowWidth(img image.Image, y int) int {
	b := img.Bounds()
	n := 0
	for x := b.Min.X; x < b.Max.X; x++ {
		r, g, bb, _ := img.At(x, y).RGBA()
		if r>>8 >= 200 && g>>8 >= 200 && bb>>8 >= 200 {
			n++
		}
	}
	return n
}

// whiteRowRange returns the topmost and bottommost rows containing any near-white
// pixel.
func whiteRowRange(img image.Image) (int, int) {
	b := img.Bounds()
	lo, hi := -1, -1
	for y := b.Min.Y; y < b.Max.Y; y++ {
		if rowWidth(img, y) > 0 {
			if lo < 0 {
				lo = y
			}
			hi = y
		}
	}
	return lo, hi
}

type rot3DAxis struct {
	comp   string                     // composition name (also the disk-cache isolation handle)
	prop   string                     // transform-group matchname for JSX readback
	idx    int                        // array index for multi-component props (Orientation); 0 for scalars
	expect float64                    // expected readback value
	time   float64                    // unique render time (disk-cache key isolation)
	apply  func(b *aep.Layer) error   // the per-axis setter under test (after Reopen)
	check  func(t *testing.T, ver string, img image.Image)
}

func runLayer3DRotAxisGate(t *testing.T, aeExe, ver string, target aep.AETarget, ax rot3DAxis) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/3d_rotaxes_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_3d_rotaxes.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, ax.comp, 1920, 1080, 30, 5)
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
	if err := ax.apply(b); err != nil {
		t.Fatalf("apply %s: %v", ax.prop, err)
	}
	// Camera close in (strong perspective) so the foreshortening is pronounced.
	if err := c.LayerByName("Cam").SetPosition([]float64{960, 540, -800}); err != nil {
		t.Fatalf("Cam SetPosition: %v", err)
	}
	if err := aep.MoveToEnd(c.LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "rot_in.aep")
	resavedAEP := filepath.Join(tempDir, "rot_resaved.aep")
	doneFile := filepath.Join(tempDir, "rot.done")
	png := filepath.Join(tempDir, "rot.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"png":%q,"comp":%q,"prop":%q,"idx":%d,"expect":%g,"time":%g}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(png),
		ax.comp, ax.prop, ax.idx, ax.expect, ax.time)
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	// Disk-cache key is (comp.id=1, render-time); clear it so a sibling axis's
	// stale frame at a reused key can't false-green this render.
	clearAEDiskCache(t)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 240)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("3d %s %s AE readback:\n%s", ax.prop, ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("3d %s %s ship gate FAIL:\n%s", ax.prop, ver, body)
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
	ax.check(t, ver, img)

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("BOX") == nil {
		t.Fatalf("resaved BOX missing")
	}
}

// --- Rotate X: top/bottom edge widths foreshorten (trapezoid on its side) ---

func rotateXAxis() rot3DAxis {
	return rot3DAxis{
		comp: "ROTX", prop: "ADBE Rotate X", idx: 0, expect: 50, time: 0.5,
		apply: func(b *aep.Layer) error { return b.SetRotateX(50) },
		check: func(t *testing.T, ver string, img image.Image) {
			lo, hi := whiteRowRange(img)
			if lo < 0 {
				t.Fatalf("%s box not rendered (no white pixels)", ver)
			}
			topW := rowWidth(img, lo+5)
			botW := rowWidth(img, hi-5)
			if topW == 0 || botW == 0 {
				t.Fatalf("%s degenerate box widths (top=%d bot=%d)", ver, topW, botW)
			}
			tall, short := topW, botW
			if botW > topW {
				tall, short = botW, topW
			}
			ratio := float64(tall) / float64(short+1)
			t.Logf("%s ROTX rows [%d,%d] topW=%d botW=%d (tall/short %.2f)", ver, lo, hi, topW, botW, ratio)
			if ratio < 1.2 {
				t.Errorf("%s box edges near-symmetric (topW=%d botW=%d, ratio %.2f) — RotateX not applied / not 3D / no perspective", ver, topW, botW, ratio)
			}
		},
	}
}

// --- Orientation Y=50: near/far vertical edge heights foreshorten (trapezoid) ---

func orientationAxis() rot3DAxis {
	return rot3DAxis{
		comp: "ORIENT", prop: "ADBE Orientation", idx: 1, expect: 50, time: 1.0,
		apply: func(b *aep.Layer) error { return b.SetOrientation([]float64{0, 50, 0}) },
		check: func(t *testing.T, ver string, img image.Image) {
			lo, hi := whiteColRange(img)
			if lo < 0 {
				t.Fatalf("%s box not rendered (no white pixels)", ver)
			}
			leftH := colHeight(img, lo+5)
			rightH := colHeight(img, hi-5)
			if leftH == 0 || rightH == 0 {
				t.Fatalf("%s degenerate box heights (left=%d right=%d)", ver, leftH, rightH)
			}
			tall, short := leftH, rightH
			if rightH > leftH {
				tall, short = rightH, leftH
			}
			ratio := float64(tall) / float64(short+1)
			t.Logf("%s ORIENT cols [%d,%d] leftH=%d rightH=%d (tall/short %.2f)", ver, lo, hi, leftH, rightH, ratio)
			if ratio < 1.2 {
				t.Errorf("%s box edges near-symmetric (leftH=%d rightH=%d, ratio %.2f) — Orientation not applied / not 3D / no perspective", ver, leftH, rightH, ratio)
			}
		},
	}
}

// --- Rotate Z=45: in-plane spin turns the square into a diamond ---

func rotateZAxis() rot3DAxis {
	return rot3DAxis{
		comp: "ROTZ", prop: "ADBE Rotate Z", idx: 0, expect: 45, time: 1.5,
		apply: func(b *aep.Layer) error { return b.SetRotation(45) },
		check: func(t *testing.T, ver string, img image.Image) {
			lo, hi := whiteColRange(img)
			if lo < 0 {
				t.Fatalf("%s box not rendered (no white pixels)", ver)
			}
			center := (lo + hi) / 2
			centerH := colHeight(img, center)
			edgeH := colHeight(img, lo+5)
			ratio := float64(centerH) / float64(edgeH+1)
			t.Logf("%s ROTZ cols [%d,%d] centerH=%d edgeH=%d (center/edge %.2f)", ver, lo, hi, centerH, edgeH, ratio)
			if centerH == 0 {
				t.Fatalf("%s box not rendered at center", ver)
			}
			// A 45° diamond: tall center column, near-zero edge columns. An
			// axis-aligned square would have center ≈ edge (ratio ≈ 1).
			if ratio < 1.5 {
				t.Errorf("%s box not diamond (centerH=%d edgeH=%d, ratio %.2f) — RotateZ not applied", ver, centerH, edgeH, ratio)
			}
		},
	}
}

func TestLayer3DRotateX_AEShipGate_AE2020(t *testing.T) {
	runLayer3DRotAxisGate(t, ae2020(), "AE2020", aep.TargetAE2020, rotateXAxis())
}
func TestLayer3DRotateX_AEShipGate_AE2025(t *testing.T) {
	runLayer3DRotAxisGate(t, ae2025(), "AE2025", aep.TargetAE2025, rotateXAxis())
}
func TestLayer3DOrientation_AEShipGate_AE2020(t *testing.T) {
	runLayer3DRotAxisGate(t, ae2020(), "AE2020", aep.TargetAE2020, orientationAxis())
}
func TestLayer3DOrientation_AEShipGate_AE2025(t *testing.T) {
	runLayer3DRotAxisGate(t, ae2025(), "AE2025", aep.TargetAE2025, orientationAxis())
}
func TestLayer3DRotateZ_AEShipGate_AE2020(t *testing.T) {
	runLayer3DRotAxisGate(t, ae2020(), "AE2020", aep.TargetAE2020, rotateZAxis())
}
func TestLayer3DRotateZ_AEShipGate_AE2025(t *testing.T) {
	runLayer3DRotAxisGate(t, ae2025(), "AE2025", aep.TargetAE2025, rotateZAxis())
}
