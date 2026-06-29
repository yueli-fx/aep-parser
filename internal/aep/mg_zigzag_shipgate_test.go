// internal/aep/mg_zigzag_shipgate_test.go
//
// AE ship gate for the ZigZag filter (`ADBE Vector Filter - Zigzag`) from
// scratch (MG roadmap S5, specs/2026-06-12-from-scratch-mg-roadmap.md). Verified
// at the capability's surface per delivery-contract red line 4: a 400×400 white
// rectangle gets a ZigZag filter (Size=40, Detail=8). The gate renders the frame
// and proves the originally-straight top edge became jagged by scanning each
// column for the topmost white pixel — a straight edge yields a near-constant
// top-y, a zigzagged edge yields a large top-y spread (peaks and valleys).
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

func buildMGZigZagDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGZIG", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// Dark BG so the wavy edge reads as white-on-dark unambiguously.
	bg, err := aep.NewShapeLayer(comp, "BG")
	if err != nil {
		t.Fatalf("NewShapeLayer BG: %v", err)
	}
	bgRect, err := bg.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect BG: %v", err)
	}
	if err := bgRect.SetSize([2]float64{2200, 1300}); err != nil {
		t.Fatalf("BG rect SetSize: %v", err)
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

	// A 400×400 white rect + ZigZag (Size=40, Detail=8). Centered at 960,540 →
	// original spans x∈[760,1160], y∈[340,740]; the zigzag oscillates each edge
	// by ±40px. ZigZag on top of the stack distorts the rect path the fill paints.
	wave, err := aep.NewShapeLayer(comp, "WAVE")
	if err != nil {
		t.Fatalf("NewShapeLayer WAVE: %v", err)
	}
	rect, err := wave.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect WAVE: %v", err)
	}
	if err := rect.SetSize([2]float64{400, 400}); err != nil {
		t.Fatalf("WAVE rect SetSize: %v", err)
	}
	fill, err := wave.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill WAVE: %v", err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("WAVE SetColor: %v", err)
	}
	zz, err := wave.RootGroup().AddZigZag()
	if err != nil {
		t.Fatalf("AddZigZag: %v", err)
	}
	if err := zz.SetSize(40); err != nil {
		t.Fatalf("SetSize: %v", err)
	}
	if err := zz.SetDetail(8); err != nil {
		t.Fatalf("SetDetail: %v", err)
	}
	if err := wave.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("WAVE Position: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if err := aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}
	return rp
}

// topWhiteY returns the smallest y in [yLo,yHi] where (x,y) is white, or -1.
func topWhiteY(img image.Image, x, yLo, yHi int) int {
	b := img.Bounds()
	if x < b.Min.X || x >= b.Max.X {
		return -1
	}
	for y := yLo; y <= yHi; y++ {
		if y < b.Min.Y || y >= b.Max.Y {
			continue
		}
		r, g, bb, _ := img.At(x, y).RGBA()
		if r>>8 >= 200 && g>>8 >= 200 && bb>>8 >= 200 {
			return y
		}
	}
	return -1
}

func runMGZigZagGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_zigzag_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_zigzag.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGZigZagDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_zigzag_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_zigzag_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_zigzag.done")
	framePNG := filepath.Join(tempDir, "mg_zigzag_frame.png")

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
	t.Logf("mg zigzag %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg zigzag %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0: the top edge (originally a flat line at
	// y≈340) is now jagged — scanning each column for the topmost white pixel
	// across x∈[800,1120] yields a large spread (peaks ~y=300, valleys ~y=380),
	// whereas a straight edge would yield a near-constant top-y.
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	minY, maxY, found := 1<<30, -1, 0
	for x := 800; x <= 1120; x += 6 {
		ty := topWhiteY(img, x, 260, 430)
		if ty < 0 {
			continue
		}
		found++
		if ty < minY {
			minY = ty
		}
		if ty > maxY {
			maxY = ty
		}
	}
	spread := maxY - minY
	centerWhite := whiteNear(img, 960, 540, 10)
	t.Logf("%s zigzag: top-edge columns found=%d top-y spread=%d (min=%d max=%d) centerWhite=%v",
		ver, found, spread, minY, maxY, centerWhite)
	if found < 30 {
		t.Errorf("%s only %d top-edge columns had white — shape not rendered", ver, found)
	}
	if spread < 30 {
		t.Errorf("%s top-edge top-y spread=%d (<30) — edge is straight, not zigzagged", ver, spread)
	}
	if !centerWhite {
		t.Errorf("%s center not white — interior not filled", ver)
	}

	// Resave proof: the ZigZag filter survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("WAVE") == nil {
		t.Fatal("resaved: WAVE layer missing")
	}
}

func TestMGZigZag_AEShipGate_AE2020(t *testing.T) {
	runMGZigZagGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGZigZag_AEShipGate_AE2025(t *testing.T) {
	runMGZigZagGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
