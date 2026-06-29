// internal/aep/layer_3d_light_shipgate_test.go
//
// AE ship gate for from-scratch 3D LIGHTING (roadmap priority 2, the 光照 half of
// the remaining Material-Options item). A from-scratch 3D gray PANEL is lit by a
// from-scratch POINT light (NewLightLayer + SetLightKind(Point) + SetPosition +
// SetLightIntensity) placed up-left and in front of the panel. The whole scene
// is Go-built — no JSX mutation.
//
// This needs ZERO new serializer code: SetLightKind writes ldta @0x88
// (length-preserving) and AE's default material has "Accepts Lights"=ON, so a 3D
// layer is lit without touching the (empty, from-scratch) Material Options group.
// 4414 = ExtendScript LightType.POINT — SetLightKind from scratch is genuine.
//
// Per delivery-contract red line 4 the gate renders the frame and asserts on
// pixels that the panel carries a brightness FALLOFF: the up-left interior
// (near the light) is markedly brighter than the down-right interior (far from
// it). A flat-shaded 3D panel (no light effect) or a 2D layer would render
// uniform — the gradient is the proof the light illuminates our 3D layer.
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

// avgLuma returns the mean red-channel (=luma on a gray fill) brightness over
// the rectangle [x0,x1)×[y0,y1), sampled every 4px.
func avgLuma(img image.Image, x0, y0, x1, y1 int) float64 {
	sum, n := 0.0, 0
	for y := y0; y < y1; y += 4 {
		for x := x0; x < x1; x += 4 {
			r, _, _, _ := img.At(x, y).RGBA()
			sum += float64(r >> 8)
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func runLayer3DLightGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/3d_light_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_3d_light.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "LIT3D", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	// Dark backdrop so the panel reads clearly against it.
	bg, _ := aep.NewShapeLayer(comp, "BG")
	bgr, _ := bg.RootGroup().AddRect()
	_ = bgr.SetSize([2]float64{2400, 1400})
	bgf, _ := bg.RootGroup().AddFill()
	_ = bgf.SetColor([4]float64{0.02, 0.02, 0.03, 1})
	_ = bg.Position().SetStaticValue([2]float64{960, 540})
	// Mid-gray panel (headroom for the falloff without clipping to white).
	panel, _ := aep.NewShapeLayer(comp, "PANEL")
	pr, _ := panel.RootGroup().AddRect()
	_ = pr.SetSize([2]float64{1200, 1000})
	pf, _ := panel.RootGroup().AddFill()
	_ = pf.SetColor([4]float64{0.5, 0.5, 0.5, 1})
	_ = panel.Position().SetStaticValue([2]float64{960, 540})
	if _, err := aep.NewLightLayer(comp, "Lamp"); err != nil {
		t.Fatalf("NewLightLayer: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	c := rp.Compositions[0]
	pl := c.LayerByName("PANEL")
	if err := pl.SetIs3D(true); err != nil {
		t.Fatalf("PANEL SetIs3D: %v", err)
	}
	lamp := c.LayerByName("Lamp")
	if err := lamp.SetLightKind(aep.LightKindPoint); err != nil {
		t.Fatalf("SetLightKind: %v", err)
	}
	// Point light up-left and in front of the panel (z<0 toward the viewer).
	if err := lamp.SetPosition([]float64{450, 300, -700}); err != nil {
		t.Fatalf("Lamp SetPosition: %v", err)
	}
	if p := lamp.LightIntensity(); p != nil {
		_ = lamp.SetLightIntensity(150)
	}
	if err := aep.MoveToEnd(c.LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "lit_in.aep")
	resavedAEP := filepath.Join(tempDir, "lit_resaved.aep")
	doneFile := filepath.Join(tempDir, "lit.done")
	png := filepath.Join(tempDir, "lit.png")

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
	t.Logf("3d light %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("3d light %s ship gate FAIL:\n%s", ver, body)
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
	// Panel spans roughly x[360,1560] y[40,1040] (1200×1000 centred at 960,540).
	// Sample interior corners: near-light (top-left) vs far (bottom-right).
	near := avgLuma(img, 420, 120, 620, 320)
	far := avgLuma(img, 1300, 760, 1500, 960)
	t.Logf("%s panel luma near-light=%.1f far=%.1f (ratio %.2f)", ver, near, far, near/far)
	if near < 30 {
		t.Errorf("%s near-light region too dark (%.1f) — panel not rendering/lit", ver, near)
	}
	if near < 1.3*far {
		t.Errorf("%s no lighting falloff (near=%.1f far=%.1f, ratio %.2f) — light not illuminating the 3D panel", ver, near, far, near/far)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("PANEL") == nil || re.Compositions[0].LayerByName("Lamp") == nil {
		t.Fatalf("resaved PANEL/Lamp missing")
	}
}

func TestLayer3DLight_AEShipGate_AE2020(t *testing.T) {
	runLayer3DLightGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestLayer3DLight_AEShipGate_AE2025(t *testing.T) {
	runLayer3DLightGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
