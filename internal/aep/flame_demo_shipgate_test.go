// internal/aep/flame_demo_shipgate_test.go
//
// Phase 0 (make-or-break) of the procedural-FX generator (spec 2026-06-18-
// procedural-fx-generator): can the library DETERMINISTICALLY build a .aep that
// renders as a recognizable flame? Builds a solid + native effect stack
// (Fractal Noise -> Tint -> Turbulent Displace, Evolution animated) 100% in Go,
// has AE 2020+2025 render a frame to PNG, and checks the rendered pixels (red
// line 4: verify RENDERED output, not stored values). Mirrors the proven
// orbit_demo render harness. Gated by AE_SHIP_GATE.
//
// Fractal Noise param matchNames (probe_flame_params.jsx, AE2025):
//   0004 Contrast · 0005 Brightness · 0009 Uniform-Scaling(off to split W/H) ·
//   0010 Scale · 0011 Scale-Width · 0012 Scale-Height · 0015 Complexity ·
//   0023 Evolution(animate) · 0029 Opacity
// Turbulent Displace: 0002 Amount · 0003 Size · 0006 Evolution
// Tint: 0001 Map-Black-To · 0002 Map-White-To · 0003 Amount-to-Tint
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

func buildFlameDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "FLAME", 1080, 1920, 30, 4)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "Flame", 1080, 1920, [3]float64{0, 0, 0}); err != nil {
		t.Fatalf("NewSolidLayer: %v", err)
	}
	// Elided effect params (Contrast, Tint colors) need the parade parsed, so
	// Reopen before AddEffect/SetEffectParam (proven pattern, animate_effect_param_vec_test).
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	sol := rp.Compositions[0].LayerByName("Flame")
	if sol == nil {
		t.Fatal("reopened Flame layer missing")
	}
	set := func(label string, fx *aep.Effect, mn string, v any) {
		t.Helper()
		if _, err := aep.SetEffectParam(sol, fx, mn, v); err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}

	// 1) Fractal Noise tuned to tall, high-contrast flame-like streaks.
	fn, err := aep.AddEffect(sol, aep.EffectFractalNoise)
	if err != nil {
		t.Fatalf("AddEffect FractalNoise: %v", err)
	}
	set("FN Contrast", fn, "ADBE Fractal Noise-0004", 200.0)
	set("FN Brightness", fn, "ADBE Fractal Noise-0005", 0.0)
	set("FN UniformScale-off", fn, "ADBE Fractal Noise-0009", 0.0)
	set("FN ScaleWidth", fn, "ADBE Fractal Noise-0011", 50.0)
	set("FN ScaleHeight", fn, "ADBE Fractal Noise-0012", 300.0)
	set("FN Complexity", fn, "ADBE Fractal Noise-0015", 6.0)

	// 2) Tint: black -> deep red, white -> orange-yellow ([A,R,G,B] 0-255).
	tn, err := aep.AddEffect(sol, aep.EffectTint)
	if err != nil {
		t.Fatalf("AddEffect Tint: %v", err)
	}
	set("Tint Black", tn, "ADBE Tint-0001", []float64{255, 12, 0, 0})
	set("Tint White", tn, "ADBE Tint-0002", []float64{255, 255, 190, 40})
	set("Tint Amount", tn, "ADBE Tint-0003", 100.0)

	// 3) Turbulent Displace: organic wavering of the fire edges.
	td, err := aep.AddEffect(sol, aep.EffectTurbulentDisplace)
	if err != nil {
		t.Fatalf("AddEffect TurbulentDisplace: %v", err)
	}
	set("TD Amount", td, "ADBE Turbulent Displace-0002", 45.0)
	set("TD Size", td, "ADBE Turbulent Displace-0003", 30.0)

	// 4) Feathered teardrop mask -> clip the fire texture to a flame silhouette
	//    (wide bottom, narrow tip). Vertices in layer px (1080x1920), center x=540.
	flamePath := aep.BezierPath{
		Vertices: [][2]float64{
			{540, 250},  // tip
			{700, 760},
			{812, 1260},
			{700, 1700},
			{540, 1785}, // bottom center
			{380, 1700},
			{268, 1260},
			{380, 760},
		},
		Closed: true,
	}
	mask, err := aep.AddMask(sol, "FlameMask", flamePath)
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	if err := mask.SetFeather([2]float64{95, 95}); err != nil {
		t.Fatalf("SetFeather: %v", err)
	}

	return rp
}

func runFlameDemoGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/flame_demo_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_flame_demo.jsx`
	toFwd := func(s string) string { return strings.ReplaceAll(s, `\`, `/`) }

	p := buildFlameDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "flame_in.aep")
	doneFile := filepath.Join(tempDir, "flame.done")
	framePNG := filepath.Join(tempDir, "flame_frame.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"png":%q,"t":2.0}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(framePNG))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 240)

	body, _ := os.ReadFile(doneFile)
	t.Logf("flame %s:\n%s", ver, string(body))
	if l := strings.SplitN(string(body), "\n", 2); len(l) == 0 || strings.TrimSpace(l[0]) != "PASS" {
		t.Fatalf("flame %s render FAIL:\n%s", ver, string(body))
	}

	// Persist the rendered frame outside t.TempDir so it can be eyeballed.
	keep := filepath.Join("..", "..", "tmp_debug", "flame_"+ver+".png")
	if data, rerr := os.ReadFile(framePNG); rerr == nil {
		_ = os.WriteFile(keep, data, 0644)
		t.Logf("flame %s frame kept at %s", ver, keep)
	}

	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("png: %v", err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	bright, warm := 0, 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y += 17 {
		for x := b.Min.X; x < b.Max.X; x += 17 {
			r, g, bl, _ := img.At(x, y).RGBA()
			r8, g8, b8 := int(r>>8), int(g>>8), int(bl>>8)
			if r8 > 40 || g8 > 40 || b8 > 40 {
				bright++
			}
			if r8 > 80 && r8-b8 > 40 && r8 >= g8 {
				warm++
			}
		}
	}
	t.Logf("flame %s: bright=%d warm=%d (frame %s)", ver, bright, warm, keep)
	if bright == 0 {
		t.Errorf("flame %s: frame all black — effect not rendering", ver)
	}
}

func TestFlameDemo_AEShipGate_AE2025(t *testing.T) {
	runFlameDemoGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
func TestFlameDemo_AEShipGate_AE2020(t *testing.T) {
	runFlameDemoGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
