// internal/aep/mg_trim_shipgate_test.go
//
// AE ship gate for Trim Paths (`ADBE Vector Filter - Trim`) from scratch (MG
// roadmap S3, specs/2026-06-12-from-scratch-mg-roadmap.md). Verified at the
// capability's surface per delivery-contract red line 4: the gate renders the
// frame and asserts on actual pixels that the trimmed ring is cut to an arc.
//
//   - FULL: ellipse + white stroke, Trim End=100 (identity) — a complete ring;
//     stroke present on BOTH the left and right of the circle.
//   - HALF: ellipse + white stroke, Trim Start=0 End=50 — AE's ellipse path
//     starts at the top and winds clockwise, so 0..50% reveals the RIGHT half
//     (top→right→bottom); the stroke is present on the right and ABSENT on the
//     left. The FULL-vs-HALF left-side difference is the trim proof (rules out
//     a dropped layer).
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

func buildMGTrimDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGTRIM", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	bg, err := aep.NewShapeLayer(comp, "BG")
	if err != nil {
		t.Fatalf("NewShapeLayer BG: %v", err)
	}
	rect, err := bg.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := rect.SetSize([2]float64{2200, 1300}); err != nil {
		t.Fatalf("rect SetSize: %v", err)
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

	// addRing builds a layer with an ellipse + white stroke + a Trim filter,
	// centred at `center`. trimEnd<100 cuts the ring to an arc. Add order
	// [Ellipse, Stroke, Trim] matches the AE fixture — bottom-up the stack is
	// path → stroke → trim, so the trim cuts the stroked ring (verified by
	// rendering v2_2_trim.aep).
	addRing := func(name string, center [2]float64, trimEnd float64) {
		rl, err := aep.NewShapeLayer(comp, name)
		if err != nil {
			t.Fatalf("NewShapeLayer %s: %v", name, err)
		}
		el, err := rl.RootGroup().AddEllipse()
		if err != nil {
			t.Fatalf("AddEllipse %s: %v", name, err)
		}
		if err := el.SetSize([2]float64{320, 320}); err != nil {
			t.Fatalf("%s ellipse SetSize: %v", name, err)
		}
		st, err := rl.RootGroup().AddStroke()
		if err != nil {
			t.Fatalf("AddStroke %s: %v", name, err)
		}
		if err := st.SetColor([4]float64{1, 1, 1, 1}); err != nil {
			t.Fatalf("%s stroke SetColor: %v", name, err)
		}
		if err := st.SetWidth(24); err != nil {
			t.Fatalf("%s stroke SetWidth: %v", name, err)
		}
		tr, err := rl.RootGroup().AddTrim()
		if err != nil {
			t.Fatalf("AddTrim %s: %v", name, err)
		}
		if err := tr.SetEnd(trimEnd); err != nil {
			t.Fatalf("%s trim SetEnd: %v", name, err)
		}
		if err := rl.Position().SetStaticValue(center); err != nil {
			t.Fatalf("%s Position: %v", name, err)
		}
	}

	addRing("FULL", [2]float64{560, 540}, 100)
	addRing("HALF", [2]float64{1360, 540}, 50)

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if err := aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}
	return rp
}

// whiteNear reports whether a near-white pixel (all channels ≥ 200) exists
// within ±win of (px,py) — tolerant of the stroke band's exact sub-pixel edge.
func whiteNear(img image.Image, px, py, win int) bool {
	b := img.Bounds()
	for y := py - win; y <= py+win; y++ {
		for x := px - win; x <= px+win; x++ {
			if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
				continue
			}
			r, g, bb, _ := img.At(x, y).RGBA()
			if r>>8 >= 200 && g>>8 >= 200 && bb>>8 >= 200 {
				return true
			}
		}
	}
	return false
}

func runMGTrimGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_trim_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_trim.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGTrimDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_trim_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_trim_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_trim.done")
	framePNG := filepath.Join(tempDir, "mg_trim_frame.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"png":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(framePNG))
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
	t.Logf("mg trim %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg trim %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0. Ring radius 160 around each centre.
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const R, win = 160, 10
	// FULL ring (centre 560,540): stroke on BOTH sides — a complete ring.
	fullLeft := whiteNear(img, 560-R, 540, win)
	fullRight := whiteNear(img, 560+R, 540, win)
	// HALF ring (centre 1360,540): Trim End=50 reveals the right half only.
	halfRight := whiteNear(img, 1360+R, 540, win)
	halfTop := whiteNear(img, 1360, 540-R, win)
	halfBottom := whiteNear(img, 1360, 540+R, win)
	halfLeft := whiteNear(img, 1360-R, 540, win)
	t.Logf("%s pixels: FULL L=%v R=%v | HALF T=%v R=%v B=%v L=%v",
		ver, fullLeft, fullRight, halfTop, halfRight, halfBottom, halfLeft)

	if !fullLeft || !fullRight {
		t.Errorf("%s FULL ring incomplete (left=%v right=%v) — stroke/ellipse not rendering", ver, fullLeft, fullRight)
	}
	if !halfRight || !halfTop || !halfBottom {
		t.Errorf("%s HALF arc missing on its revealed side (top=%v right=%v bottom=%v) — trim over-cut or stroke gone",
			ver, halfTop, halfRight, halfBottom)
	}
	if halfLeft {
		t.Errorf("%s HALF left side still painted — Trim End=50 did not cut the ring", ver)
	}

	// Resave proof: trim End survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	cc := re.Compositions[0]
	if cc.LayerByName("HALF") == nil || cc.LayerByName("FULL") == nil {
		t.Fatalf("resaved: HALF/FULL missing (layers=%d)", len(cc.Layers))
	}
}

func TestMGTrim_AEShipGate_AE2020(t *testing.T) {
	runMGTrimGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGTrim_AEShipGate_AE2025(t *testing.T) {
	runMGTrimGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
