// internal/aep/flame_demo_shipgate_test.go
//
// Procedural-FX generator (spec 2026-06-18-procedural-fx-generator), flame demo.
// v3 = MULTI-LAYER composite for depth (validates technique T3 additive-depth,
// docs/fx-techniques.md): 3 Fractal Noise->Tritone->Turbulent Displace fire layers
// at different noise scales (big tongues / mid / fine core), concentric masks for
// outer->inner temperature zones, Add-blended so overlaps build a white-hot core,
// + a top Glo2 Glow adjustment + black bg. Motion is SHARED across layers (coherent
// — per-layer rates desync and shimmer late). Builds 100% in Go, AE 2020+2025 render
// a frame to PNG, pixel-checked (red line 4). User-accepted on real machine 2026-06-18.
//
// Param matchNames (probe_flame_params.jsx): Fractal Noise 0004 Contrast / 0005
// Brightness / 0009 Uniform-Scaling / 0011 Scale-Width / 0012 Scale-Height / 0015
// Complexity / 0013 Offset-Turbulence(vec) / 0023 Evolution. Turbulent Displace 0002
// Amount / 0003 Size / 0006 Evolution. Tritone 0001 Highlights / 0002 Midtones /
// 0003 Shadows. Glo2 0002 Threshold / 0003 Radius / 0004 Intensity.
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

type flameLayerCfg struct {
	name                                             string
	contrast, brightness, scaleW, scaleH, complexity float64
	dispAmt, dispSize, maskScale                     float64
	tShadow, tMid, tHigh                             []float64 // [A,R,G,B] 0-255
}

// flameScalePath shrinks a path toward (cx,cy) by f (concentric flame zones).
func flameScalePath(p aep.BezierPath, cx, cy, f float64) aep.BezierPath {
	out := aep.BezierPath{Closed: p.Closed}
	for _, v := range p.Vertices {
		out.Vertices = append(out.Vertices, [2]float64{cx + (v[0]-cx)*f, cy + (v[1]-cy)*f})
	}
	return out
}

func buildFlameDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "FLAME", 1080, 1920, 30, 4)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	// Build top->bottom (new layers append below): Glow adj, core, mid, base, BG.
	if _, err := aep.NewAdjustmentLayer(comp, "Glow"); err != nil {
		t.Fatalf("NewAdjustmentLayer: %v", err)
	}
	cfgs := []flameLayerCfg{
		{name: "FireCore", contrast: 185, brightness: -18, scaleW: 30, scaleH: 420, complexity: 6,
			dispAmt: 38, dispSize: 28, maskScale: 0.5,
			tShadow: []float64{255, 30, 0, 0}, tMid: []float64{255, 255, 150, 25}, tHigh: []float64{255, 255, 250, 235}},
		{name: "FireMid", contrast: 158, brightness: -24, scaleW: 50, scaleH: 330, complexity: 6,
			dispAmt: 50, dispSize: 38, maskScale: 0.76,
			tShadow: []float64{255, 16, 0, 0}, tMid: []float64{255, 245, 80, 0}, tHigh: []float64{255, 255, 195, 70}},
		{name: "FireBase", contrast: 128, brightness: -30, scaleW: 82, scaleH: 260, complexity: 5,
			dispAmt: 62, dispSize: 46, maskScale: 1.0,
			tShadow: []float64{255, 14, 0, 0}, tMid: []float64{255, 175, 26, 0}, tHigh: []float64{255, 255, 110, 8}},
	}
	for _, c := range cfgs {
		if _, err := aep.NewSolidLayer(comp, c.name, 1080, 1920, [3]float64{0, 0, 0}); err != nil {
			t.Fatalf("NewSolidLayer %s: %v", c.name, err)
		}
	}
	if _, err := aep.NewSolidLayer(comp, "BG", 1080, 1920, [3]float64{0, 0, 0}); err != nil {
		t.Fatalf("NewSolidLayer BG: %v", err)
	}

	// Elided effect params need the parade parsed -> Reopen before AddEffect.
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	fc := rp.Compositions[0]
	set := func(label string, l *aep.Layer, fx *aep.Effect, mn string, v any) {
		t.Helper()
		if _, err := aep.SetEffectParam(l, fx, mn, v); err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}

	// Top Glow adjustment — glows the Add-composited fire below.
	glowL := fc.LayerByName("Glow")
	gl, err := aep.AddEffect(glowL, "ADBE Glo2")
	if err != nil {
		t.Fatalf("AddEffect Glo2: %v", err)
	}
	set("Glow Threshold", glowL, gl, "ADBE Glo2-0002", 50.0)
	set("Glow Radius", glowL, gl, "ADBE Glo2-0003", 55.0)
	set("Glow Intensity", glowL, gl, "ADBE Glo2-0004", 1.5)

	flamePath := aep.BezierPath{
		Vertices: [][2]float64{
			{540, 250}, {700, 760}, {812, 1260}, {700, 1700},
			{540, 1785}, {380, 1700}, {268, 1260}, {380, 760},
		},
		Closed: true,
	}
	// SHARED, calm motion across all layers (coherent -> no desync shimmer).
	const evoEnd, offEndY, tdEvoEnd = 540.0, 540.0, 360.0

	for _, c := range cfgs {
		l := fc.LayerByName(c.name)
		if l == nil {
			t.Fatalf("reopened layer %s missing", c.name)
		}
		fn, err := aep.AddEffect(l, aep.EffectFractalNoise)
		if err != nil {
			t.Fatalf("AddEffect FractalNoise %s: %v", c.name, err)
		}
		set("FN Contrast", l, fn, "ADBE Fractal Noise-0004", c.contrast)
		set("FN Brightness", l, fn, "ADBE Fractal Noise-0005", c.brightness)
		set("FN UniformScale", l, fn, "ADBE Fractal Noise-0009", 0.0)
		set("FN ScaleW", l, fn, "ADBE Fractal Noise-0011", c.scaleW)
		set("FN ScaleH", l, fn, "ADBE Fractal Noise-0012", c.scaleH)
		set("FN Complexity", l, fn, "ADBE Fractal Noise-0015", c.complexity)

		tr, err := aep.AddEffect(l, "ADBE Tritone")
		if err != nil {
			t.Fatalf("AddEffect Tritone %s: %v", c.name, err)
		}
		set("Tritone Hi", l, tr, "ADBE Tritone-0001", c.tHigh)
		set("Tritone Mid", l, tr, "ADBE Tritone-0002", c.tMid)
		set("Tritone Sh", l, tr, "ADBE Tritone-0003", c.tShadow)

		td, err := aep.AddEffect(l, aep.EffectTurbulentDisplace)
		if err != nil {
			t.Fatalf("AddEffect TurbulentDisplace %s: %v", c.name, err)
		}
		set("TD Amount", l, td, "ADBE Turbulent Displace-0002", c.dispAmt)
		set("TD Size", l, td, "ADBE Turbulent Displace-0003", c.dispSize)

		// Concentric mask -> outer/mid/inner temperature zones (layered, not solid fill).
		mp := flameScalePath(flamePath, 540, 1080, c.maskScale)
		feather := 110 * c.maskScale
		if feather < 60 {
			feather = 60
		}
		mask, err := aep.AddMask(l, c.name+"Mask", mp)
		if err != nil {
			t.Fatalf("AddMask %s: %v", c.name, err)
		}
		if err := mask.SetFeather([2]float64{feather, feather}); err != nil {
			t.Fatalf("SetFeather %s: %v", c.name, err)
		}

		if err := l.SetBlendingMode(aep.BlendingModeAdd); err != nil {
			t.Fatalf("SetBlendingMode %s: %v", c.name, err)
		}

		if _, err := aep.AnimateEffectParam(l, fn, "ADBE Fractal Noise-0023",
			[]aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 4, Value: evoEnd}}); err != nil {
			t.Fatalf("Animate FN Evolution %s: %v", c.name, err)
		}
		if _, err := aep.AnimateEffectParamVec(l, fn, "ADBE Fractal Noise-0013",
			[]aep.VectorKeyframe{{Time: 0, Value: []float64{540, 960}}, {Time: 4, Value: []float64{540, offEndY}}}); err != nil {
			t.Fatalf("Animate FN Offset %s: %v", c.name, err)
		}
		if _, err := aep.AnimateEffectParam(l, td, "ADBE Turbulent Displace-0006",
			[]aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 4, Value: tdEvoEnd}}); err != nil {
			t.Fatalf("Animate TD Evolution %s: %v", c.name, err)
		}
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
	framePNG2 := filepath.Join(tempDir, "flame_frame2.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"png":%q,"t":1.0,"png2":%q,"t2":3.0}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(framePNG), toFwd(framePNG2))
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
	if warm < 50 {
		t.Errorf("flame %s: too few fire-colored pixels = %d (Tritone not applied?)", ver, warm)
	}

	// Motion: frame at t=1 must differ from t=3 (Evolution animating).
	img2, err2 := decodePNG(framePNG2)
	if err2 != nil {
		t.Errorf("flame %s: decode frame2: %v", ver, err2)
		return
	}
	if keep2data, rerr := os.ReadFile(framePNG2); rerr == nil {
		_ = os.WriteFile(filepath.Join("..", "..", "tmp_debug", "flame_"+ver+"_t3.png"), keep2data, 0644)
	}
	diff := 0
	for y := b.Min.Y; y < b.Max.Y; y += 17 {
		for x := b.Min.X; x < b.Max.X; x += 17 {
			r1, g1, b1, _ := img.At(x, y).RGBA()
			r2, g2, b2, _ := img2.At(x, y).RGBA()
			if absDiff(int(r1>>8), int(r2>>8))+absDiff(int(g1>>8), int(g2>>8))+absDiff(int(b1>>8), int(b2>>8)) > 24 {
				diff++
			}
		}
	}
	t.Logf("flame %s: %d sample points differ between t=1 and t=3", ver, diff)
	if diff < 20 {
		t.Errorf("flame %s: frames at t=1,t=3 nearly identical (diff=%d) — Evolution not animating", ver, diff)
	}
}

func decodePNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func absDiff(a, b int) int {
	if a < b {
		return b - a
	}
	return a - b
}

func TestFlameDemo_AEShipGate_AE2025(t *testing.T) {
	runFlameDemoGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
func TestFlameDemo_AEShipGate_AE2020(t *testing.T) {
	runFlameDemoGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
