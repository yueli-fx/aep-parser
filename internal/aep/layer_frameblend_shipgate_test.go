// internal/aep/layer_frameblend_shipgate_test.go
//
// AE ship gate for the layer-set FRAME-BLEND setters (verify=roundtrip → ae-accept):
//
//   - SetFrameBlendEnabled    (ldta switch bit @0x27)
//   - SetFrameBlendPixelMotion (ldta mode bit @0x25; false=Frame Mix, true=Pixel Motion)
//
// Frame blending is only valid on layers with temporal frames — video footage OR
// precomp/comp layers. A solid/still cannot frame-blend, so the carrier uses
// PRECOMP layers (comp-as-layer), which build entirely from-scratch in Go and need
// no external video file. MAIN holds two precomp layers sourcing inner comp SRC:
//   LMIX → enabled + !pixelMotion → AE FrameBlendingType.FRAME_MIX
//   LPXM → enabled +  pixelMotion → AE FrameBlendingType.PIXEL_MOTION
// The FRAME_MIX vs PIXEL_MOTION contrast proves BOTH bits (a no-op or a single-bit
// error would collapse to NO_FRAME_BLEND or the wrong mode). DOM-readback only, no
// resave (idta-batch lesson). Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func buildFrameBlendDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	main, err := aep.NewComposition(p, "MAIN", 640, 360, 30, 4)
	if err != nil {
		t.Fatalf("NewComposition MAIN: %v", err)
	}
	src, err := aep.NewComposition(p, "SRC", 320, 180, 30, 4)
	if err != nil {
		t.Fatalf("NewComposition SRC: %v", err)
	}
	if _, err := aep.NewSolidLayer(src, "SRCs", 320, 180, [3]float64{0.3, 0.6, 0.9}); err != nil {
		t.Fatalf("NewSolidLayer SRCs: %v", err)
	}
	if _, err := aep.NewPrecompLayer(main, src, "LMIX"); err != nil {
		t.Fatalf("NewPrecompLayer LMIX: %v", err)
	}
	if _, err := aep.NewPrecompLayer(main, src, "LPXM"); err != nil {
		t.Fatalf("NewPrecompLayer LPXM: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	mc := rp.CompositionByName("MAIN")
	if mc == nil {
		t.Fatal("reopened MAIN missing")
	}
	mix := mc.LayerByName("LMIX")
	pxm := mc.LayerByName("LPXM")
	if mix == nil || pxm == nil {
		t.Fatal("reopened LMIX/LPXM missing")
	}
	mustMut := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustMut("LMIX SetFrameBlendEnabled", mix.SetFrameBlendEnabled(true))
	mustMut("LMIX SetFrameBlendPixelMotion", mix.SetFrameBlendPixelMotion(false))
	mustMut("LPXM SetFrameBlendEnabled", pxm.SetFrameBlendEnabled(true))
	mustMut("LPXM SetFrameBlendPixelMotion", pxm.SetFrameBlendPixelMotion(true))
	return rp
}

func runLayerFrameBlendGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/frameblend_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_frameblend.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildFrameBlendDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "frameblend_in.aep")
	doneFile := filepath.Join(tempDir, "frameblend.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q}`, toFwd(inputAEP), toFwd(doneFile))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("layer frameblend %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("layer frameblend %s ship gate FAIL:\n%s", ver, body)
	}
}

func TestLayerFrameBlend_AEShipGate_AE2020(t *testing.T) {
	runLayerFrameBlendGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestLayerFrameBlend_AEShipGate_AE2025(t *testing.T) {
	runLayerFrameBlendGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
