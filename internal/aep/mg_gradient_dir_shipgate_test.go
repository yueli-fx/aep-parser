// internal/aep/mg_gradient_dir_shipgate_test.go
//
// AE ship gate for gradient ramp DIRECTION (`ADBE Vector Grad Start Pt` /
// `End Pt`) from scratch (MG roadmap S5, specs/2026-06-12-from-scratch-mg-roadmap.md).
// Verified at the capability's surface per delivery-contract red line 4: a
// 400×400 rect filled with a red→blue gradient whose ramp is set DIAGONAL
// (Start=[-150,-150] → End=[150,150]). The gate renders the frame and asserts
// the top-left corner is red, the bottom-right is blue, and the other two
// corners are mid-purple — a pattern only a diagonal ramp produces (a default
// horizontal ramp would make the top-right blue and the bottom-left red).
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

	aep "github.com/example/aep-parser/internal/aep"
)

func buildMGGradientDirDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGGRAD", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// A 400×400 rect filled with a red→blue gradient, ramp set diagonal. Centered
	// at 960,540 → comp span x∈[760,1160], y∈[340,740]. Start/End Pt are in the
	// shape's local space ([-200,200] for a centered 400px rect); [-150,-150]→
	// [150,150] is the main diagonal.
	ramp, err := aep.NewShapeLayer(comp, "RAMP")
	if err != nil {
		t.Fatalf("NewShapeLayer RAMP: %v", err)
	}
	rect, err := ramp.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := rect.SetSize([2]float64{400, 400}); err != nil {
		t.Fatalf("rect SetSize: %v", err)
	}
	gf, err := ramp.RootGroup().AddGradientFill()
	if err != nil {
		t.Fatalf("AddGradientFill: %v", err)
	}
	if err := gf.SetColorStops([]aep.GradientColorStop{
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}}, // red
		{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}}, // blue
	}); err != nil {
		t.Fatalf("SetColorStops: %v", err)
	}
	if err := gf.SetStartPoint([2]float64{-150, -150}); err != nil {
		t.Fatalf("SetStartPoint: %v", err)
	}
	if err := gf.SetEndPoint([2]float64{150, 150}); err != nil {
		t.Fatalf("SetEndPoint: %v", err)
	}
	if err := ramp.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("RAMP Position: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	return rp
}

// avgRGB returns the mean 8-bit R,G,B over a (2*win+1)² window centered at (px,py).
func avgRGB(img image.Image, px, py, win int) (int, int, int) {
	b := img.Bounds()
	var sr, sg, sb, n int
	for y := py - win; y <= py+win; y++ {
		for x := px - win; x <= px+win; x++ {
			if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
				continue
			}
			r, g, bb, _ := img.At(x, y).RGBA()
			sr += int(r >> 8)
			sg += int(g >> 8)
			sb += int(bb >> 8)
			n++
		}
	}
	if n == 0 {
		return 0, 0, 0
	}
	return sr / n, sg / n, sb / n
}

func runMGGradientDirGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_gradient_dir_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_gradient_dir.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGGradientDirDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_gradient_dir_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_gradient_dir_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_gradient_dir.done")
	framePNG := filepath.Join(tempDir, "mg_gradient_dir_frame.png")

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
	t.Logf("mg gradient-dir %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg gradient-dir %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0: the red→blue ramp runs along the main
	// diagonal. TL=red, BR=blue, TR & BL=mid-purple (a horizontal ramp would make
	// TR blue and BL red).
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const win = 12
	tlR, _, tlB := avgRGB(img, 800, 380, win)
	brR, _, brB := avgRGB(img, 1120, 700, win)
	trR, _, trB := avgRGB(img, 1120, 380, win)
	blR, _, blB := avgRGB(img, 800, 700, win)
	t.Logf("%s gradient-dir corners: TL(r=%d,b=%d) BR(r=%d,b=%d) TR(r=%d,b=%d) BL(r=%d,b=%d)",
		ver, tlR, tlB, brR, brB, trR, trB, blR, blB)
	if tlR-tlB < 50 {
		t.Errorf("%s top-left not red (r=%d b=%d) — ramp start wrong", ver, tlR, tlB)
	}
	if brB-brR < 50 {
		t.Errorf("%s bottom-right not blue (r=%d b=%d) — ramp end wrong", ver, brR, brB)
	}
	if absInt(trR-trB) > 60 {
		t.Errorf("%s top-right not mid (r=%d b=%d) — ramp is horizontal, not diagonal", ver, trR, trB)
	}
	if absInt(blR-blB) > 60 {
		t.Errorf("%s bottom-left not mid (r=%d b=%d) — ramp is horizontal, not diagonal", ver, blR, blB)
	}

	// Resave proof: the gradient (with direction) survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("RAMP") == nil {
		t.Fatal("resaved: RAMP layer missing")
	}
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestMGGradientDir_AEShipGate_AE2020(t *testing.T) {
	runMGGradientDirGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGGradientDir_AEShipGate_AE2025(t *testing.T) {
	runMGGradientDirGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
