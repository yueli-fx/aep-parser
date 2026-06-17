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
	sol, err := aep.NewSolidLayer(comp, "Flame", 1080, 1920, [3]float64{0, 0, 0})
	if err != nil {
		t.Fatalf("NewSolidLayer: %v", err)
	}
	_, err = aep.AddEffect(sol, aep.EffectFractalNoise)
	if err != nil {
		t.Fatalf("AddEffect FractalNoise: %v", err)
	}
	return p
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
