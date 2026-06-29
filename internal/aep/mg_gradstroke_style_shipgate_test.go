// internal/aep/mg_gradstroke_style_shipgate_test.go
//
// AE ship gate for gradient-stroke GEOMETRY (width / cap / join / miter) from
// scratch. A 300×300 rect with a horizontal red→blue gradient stroke set to
// width 60 (cap Projecting, join Bevel, miter 10). Per delivery-contract red
// line 4 the headline — stroke WIDTH — is pixel-proven: a point 25px outside the
// rect's left path edge lands inside the 60px band (810±30 → 780..840) and must
// render red, while a point 40px outside (x=770, beyond the band) must be
// background. An 18px stroke (the template default, band 801..819) would leave
// the 25px-outside point on the background, which the test rejects. Cap / join /
// miter are read back from AE's DOM (their render effect is subtle corner/cap
// geometry; value-level here).
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

func buildMGGradStrokeStyleDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGGSTYLE", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "GSWIDTH")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, err := l.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := rect.SetSize([2]float64{300, 300}); err != nil {
		t.Fatalf("rect SetSize: %v", err)
	}
	gs, err := l.RootGroup().AddGradientStroke()
	if err != nil {
		t.Fatalf("AddGradientStroke: %v", err)
	}
	if err := gs.SetColorStops([]aep.GradientColorStop{
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}}, // red (left)
		{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}}, // blue (right)
	}); err != nil {
		t.Fatalf("SetColorStops: %v", err)
	}
	if err := gs.SetStartPoint([2]float64{-150, 0}); err != nil {
		t.Fatalf("SetStartPoint: %v", err)
	}
	if err := gs.SetEndPoint([2]float64{150, 0}); err != nil {
		t.Fatalf("SetEndPoint: %v", err)
	}
	if err := gs.SetStrokeWidth(60); err != nil {
		t.Fatalf("SetStrokeWidth: %v", err)
	}
	if err := gs.SetLineCap(aep.StrokeLineCapProjecting); err != nil {
		t.Fatalf("SetLineCap: %v", err)
	}
	if err := gs.SetLineJoin(aep.StrokeLineJoinBevel); err != nil {
		t.Fatalf("SetLineJoin: %v", err)
	}
	if err := gs.SetMiterLimit(10); err != nil {
		t.Fatalf("SetMiterLimit: %v", err)
	}
	if err := l.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("Position: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	return rp
}

func runMGGradStrokeStyleGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_gradstroke_style_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_gradstroke_style.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGGradStrokeStyleDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_gradstroke_style_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_gradstroke_style_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_gradstroke_style.done")
	framePNG := filepath.Join(tempDir, "mg_gradstroke_style_frame.png")

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
	t.Logf("mg gradstroke-style %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg gradstroke-style %s ship gate FAIL:\n%s", ver, body)
	}

	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	// Rect centre 960,540, size 300 → left path edge at comp x=810. A 60px stroke
	// band spans 780..840; an 18px default band would span 801..819.
	const win = 3
	inR, _, inB := avgRGB(img, 785, 540, win)  // 25px outside edge → inside 60px band
	outR, _, outB := avgRGB(img, 768, 540, win) // 42px outside edge → beyond the band
	t.Logf("%s width: inBand(r=%d,b=%d) beyond(r=%d,b=%d)", ver, inR, inB, outR, outB)
	if inR < 120 || inR-inB < 60 {
		t.Errorf("%s point 25px outside edge not red (r=%d b=%d) — stroke width didn't widen to 60", ver, inR, inB)
	}
	if outR > 60 {
		t.Errorf("%s point 42px outside edge not background (r=%d) — band wider than expected", ver, outR)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("GSWIDTH") == nil {
		t.Fatal("resaved: GSWIDTH layer missing")
	}
}

func TestMGGradStrokeStyle_AEShipGate_AE2020(t *testing.T) {
	runMGGradStrokeStyleGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGGradStrokeStyle_AEShipGate_AE2025(t *testing.T) {
	runMGGradStrokeStyleGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
