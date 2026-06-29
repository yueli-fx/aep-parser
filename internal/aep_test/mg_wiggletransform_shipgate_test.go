// internal/aep/mg_wiggletransform_shipgate_test.go
//
// AE ship gate for the Wiggle Transform filter (`ADBE Vector Filter - Wiggler`)
// from scratch (MG roadmap S5, specs/2026-06-12-from-scratch-mg-roadmap.md).
// Verified at the capability's surface per delivery-contract red line 4: a
// 200×200 white rectangle nominally centered at (960,540) gets a Wiggle
// Transform (Position amplitude [220,220], Rotation 70, Wiggles/Second 2, Seed
// 8). The gate renders frame 0 and asserts the white blob's centroid is randomly
// displaced FAR from the nominal center — proving the filter's nested transform
// actually jittered the shape, not just that the values round-tripped. The
// displacement is seed-deterministic, so both AE versions land the same centroid.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"image"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func buildMGWiggleTransformDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGWX", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// Dark BG so the displaced blob reads unambiguously.
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

	// A 200×200 white rect nominally centered at (960,540), + Wiggle Transform
	// with a large Position/Rotation amplitude. At frame 0 the seed-deterministic
	// wiggle displaces (and rotates) the whole shape away from the center.
	card, err := aep.NewShapeLayer(comp, "CARD")
	if err != nil {
		t.Fatalf("NewShapeLayer CARD: %v", err)
	}
	rect, err := card.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect CARD: %v", err)
	}
	if err := rect.SetSize([2]float64{200, 200}); err != nil {
		t.Fatalf("CARD rect SetSize: %v", err)
	}
	fill, err := card.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill CARD: %v", err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("CARD SetColor: %v", err)
	}
	wx, err := card.RootGroup().AddWiggleTransform()
	if err != nil {
		t.Fatalf("AddWiggleTransform: %v", err)
	}
	if err := wx.SetWigglesPerSecond(2); err != nil {
		t.Fatalf("SetWigglesPerSecond: %v", err)
	}
	if err := wx.SetRandomSeed(8); err != nil {
		t.Fatalf("SetRandomSeed: %v", err)
	}
	if err := wx.Transform().SetPosition([2]float64{220, 220}); err != nil {
		t.Fatalf("Transform SetPosition: %v", err)
	}
	if err := wx.Transform().SetRotation(70); err != nil {
		t.Fatalf("Transform SetRotation: %v", err)
	}
	if err := card.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("CARD Position: %v", err)
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

func runMGWiggleTransformGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mg_wiggletransform_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mg_wiggletransform.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGWiggleTransformDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_wiggletransform_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_wiggletransform_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_wiggletransform.done")
	framePNG := filepath.Join(tempDir, "mg_wiggletransform_frame.png")

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
	t.Logf("mg wiggletransform %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg wiggletransform %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0. The white rect's centroid should be far from
	// the nominal center (960,540) — the wiggle displaced it.
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}

	// Centroid of the bright (white rect) pixels across the frame.
	b := img.Bounds()
	var sumX, sumY, n float64
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			// bright = near-white (the card); BG is ~0.05.
			if r>>8 > 200 && g>>8 > 200 && bl>>8 > 200 {
				sumX += float64(x)
				sumY += float64(y)
				n++
			}
		}
	}
	if n < 500 {
		t.Fatalf("%s only %0.f white px — card not rendered", ver, n)
	}
	cx, cy := sumX/n, sumY/n
	disp := math.Hypot(cx-960, cy-540)
	t.Logf("%s wiggletransform: centroid=(%.1f,%.1f) displacement-from-center=%.1f px (whitePx=%.0f)", ver, cx, cy, disp, n)
	// A no-op (zero amplitude) leaves the centroid at (960,540) → disp ≈ 0. The
	// Position amplitude [220,220] displaces it by tens-to-hundreds of px.
	if disp < 50 {
		t.Errorf("%s centroid displacement=%.1f too small — wiggle transform did not move the shape (no-op?)", ver, disp)
	}

	// Resave proof: the Wiggle Transform filter survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("CARD") == nil {
		t.Fatal("resaved: CARD layer missing")
	}
}

func TestMGWiggleTransform_AEShipGate_AE2020(t *testing.T) {
	runMGWiggleTransformGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGWiggleTransform_AEShipGate_AE2025(t *testing.T) {
	runMGWiggleTransformGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
