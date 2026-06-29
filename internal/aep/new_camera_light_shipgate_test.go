// internal/aep/new_camera_light_shipgate_test.go
//
// AE ship gate for NewCameraLayer / NewLightLayer. Builds a fresh project with
// one comp containing a Go-created Camera + Light layer, WriteAEP, and has AE
// open it — proving AE ACCEPTS the cloned-template layers (no silent-drop /
// corrupt) and reports them as CameraLayer / LightLayer with the right names.
// Resaves so the Go side confirms AE kept them.
//
// Gated by AE_SHIP_GATE. Templates extracted from test_data/re_cameralight.aep.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runCameraLightGate(t *testing.T, target aep.AETarget, aeExe, ver string) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/camera_light_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_camera_light.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	camLayer, err := aep.NewCameraLayer(comp, "Cam1")
	if err != nil {
		t.Fatalf("NewCameraLayer: %v", err)
	}
	lightLayer, err := aep.NewLightLayer(comp, "Light1")
	if err != nil {
		t.Fatalf("NewLightLayer: %v", err)
	}
	// Exercise from-scratch SetLightKind: template default is Parallel; set Spot
	// so the gate proves AE accepts a kind-patched clone and reports SPOT
	// (verify_camera_light.jsx checks light.lightType).
	if err := lightLayer.SetLightKind(aep.LightKindSpot); err != nil {
		t.Fatalf("SetLightKind: %v", err)
	}
	// Exercise from-scratch Camera/Light Options setters (the parse-the-clone
	// property tree). verify_camera_light.jsx reads these back from AE's DOM.
	for _, e := range []struct {
		label string
		fn    func() error
	}{
		{"SetCameraZoom", func() error { return camLayer.SetCameraZoom(850) }},
		{"SetCameraDepthOfField", func() error { return camLayer.SetCameraDepthOfField(true) }},
		{"SetCameraFocusDistance", func() error { return camLayer.SetCameraFocusDistance(1200) }},
		{"SetCameraAperture", func() error { return camLayer.SetCameraAperture(180) }},
		// From-scratch Iris*/Highlight* (synthesis-inserted leaves; DoF on above).
		{"SetIrisShape", func() error { return camLayer.SetIrisShape(4) }},
		{"SetIrisRotation", func() error { return camLayer.SetIrisRotation(25) }},
		{"SetIrisRoundness", func() error { return camLayer.SetIrisRoundness(60) }},
		{"SetIrisAspectRatio", func() error { return camLayer.SetIrisAspectRatio(1.8) }},
		{"SetIrisDiffractionFringe", func() error { return camLayer.SetIrisDiffractionFringe(30) }},
		{"SetIrisHighlightGain", func() error { return camLayer.SetIrisHighlightGain(40) }},
		{"SetIrisHighlightThreshold", func() error { return camLayer.SetIrisHighlightThreshold(0.7) }},
		{"SetIrisHighlightSaturation", func() error { return camLayer.SetIrisHighlightSaturation(50) }},
		{"SetLightIntensity", func() error { return lightLayer.SetLightIntensity(65) }},
		{"SetLightConeAngle", func() error { return lightLayer.SetLightConeAngle(72) }},
		{"SetLightConeFeather", func() error { return lightLayer.SetLightConeFeather(35) }},
		// Raw cdat [A,R,G,B] 0..255; R=51/G=102/B=204 → AE DOM [0.2,0.4,0.8].
		{"SetLightColor", func() error { return lightLayer.SetLightColor([]float64{255, 51, 102, 204}) }},
		// Remaining light options (template-present slots; gate them too).
		{"SetLightFalloffType", func() error { return lightLayer.SetLightFalloffType(2) }}, // Smooth
		{"SetLightFalloffStart", func() error { return lightLayer.SetLightFalloffStart(100) }},
		{"SetLightFalloffDistance", func() error { return lightLayer.SetLightFalloffDistance(750) }}, // 500 is AE's default (elided on resave)
		{"SetLightCastsShadows", func() error { return lightLayer.SetLightCastsShadows(true) }},
		{"SetLightShadowDarkness", func() error { return lightLayer.SetLightShadowDarkness(80) }},
		{"SetLightShadowDiffusion", func() error { return lightLayer.SetLightShadowDiffusion(15) }},
	} {
		if err := e.fn(); err != nil {
			t.Fatalf("%s: %v", e.label, err)
		}
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "camlight_in.aep")
	resavedAEP := filepath.Join(tempDir, "camlight_resaved.aep")
	doneFile := filepath.Join(tempDir, "camlight.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
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
	t.Logf("%s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s camera/light ship gate FAIL:\n%s", ver, body)
	}

	// Preservation proof: AE's resave kept both layers with their types.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	var cam, light bool
	for _, l := range re.Compositions[0].Layers {
		if l.Name == "Cam1" && l.Type == aep.LayerTypeCamera {
			cam = true
			// Option values survived AE's resave (read via the parsed property tree).
			chkOpt(t, ver, "Cam1.Zoom", l.CameraZoom(), 850)
			chkOpt(t, ver, "Cam1.FocusDistance", l.CameraFocusDistance(), 1200)
			chkOpt(t, ver, "Cam1.Aperture", l.CameraAperture(), 180)
			// From-scratch Iris*/Highlight* (synthesis-inserted) survived resave.
			chkOpt(t, ver, "Cam1.IrisShape", l.IrisShape(), 4)
			chkOpt(t, ver, "Cam1.IrisRotation", l.IrisRotation(), 25)
			chkOpt(t, ver, "Cam1.IrisRoundness", l.IrisRoundness(), 60)
			chkOptApprox(t, ver, "Cam1.IrisAspectRatio", l.IrisAspectRatio(), 1.8)
			chkOpt(t, ver, "Cam1.IrisDiffractionFringe", l.IrisDiffractionFringe(), 30)
			chkOpt(t, ver, "Cam1.IrisHighlightGain", l.IrisHighlightGain(), 40)
			chkOptApprox(t, ver, "Cam1.IrisHighlightThreshold", l.IrisHighlightThreshold(), 0.7)
			chkOpt(t, ver, "Cam1.IrisHighlightSaturation", l.IrisHighlightSaturation(), 50)
		}
		if l.Name == "Light1" && l.Type == aep.LayerTypeLight {
			light = true
			if l.LightKind != aep.LightKindSpot {
				t.Errorf("%s resaved: Light 'Light1' LightKind = %v, want spot", ver, l.LightKind)
			}
			chkOpt(t, ver, "Light1.Intensity", l.LightIntensity(), 65)
			chkOpt(t, ver, "Light1.ConeAngle", l.LightConeAngle(), 72)
			chkOpt(t, ver, "Light1.ConeFeather", l.LightConeFeather(), 35)
			// From-scratch Light Color survived AE's resave (tolerant compare —
			// AE round-trips colour through float32).
			if lc := l.LightColor(); lc == nil {
				t.Errorf("%s resaved: Light1.Color property nil", ver)
			} else if rgba, ok := lc.StaticValue.([]float64); !ok || len(rgba) < 4 ||
				abs(rgba[0]-255) > 0.5 || abs(rgba[1]-51) > 0.5 || abs(rgba[2]-102) > 0.5 || abs(rgba[3]-204) > 0.5 {
				t.Errorf("%s resaved: Light1.Color = %v, want raw ~[255 51 102 204]", ver, lc.StaticValue)
			}
			// Remaining light options survived AE's resave.
			chkOpt(t, ver, "Light1.FalloffType", l.LightFalloffType(), 2)
			chkOpt(t, ver, "Light1.FalloffStart", l.LightFalloffStart(), 100)
			chkOpt(t, ver, "Light1.FalloffDistance", l.LightFalloffDistance(), 750)
			chkOpt(t, ver, "Light1.CastsShadows", l.LightCastsShadows(), 1)
			chkOpt(t, ver, "Light1.ShadowDarkness", l.LightShadowDarkness(), 80)
			chkOpt(t, ver, "Light1.ShadowDiffusion", l.LightShadowDiffusion(), 15)
		}
	}
	if !cam {
		t.Errorf("%s resaved: Camera 'Cam1' missing", ver)
	}
	if !light {
		t.Errorf("%s resaved: Light 'Light1' missing", ver)
	}
}

func chkOpt(t *testing.T, ver, label string, p *aep.Property, want float64) {
	t.Helper()
	if p == nil {
		t.Errorf("%s resaved: %s property nil", ver, label)
		return
	}
	got, ok := p.StaticValue.(float64)
	if !ok || got != want {
		t.Errorf("%s resaved: %s = %v, want %g", ver, label, p.StaticValue, want)
	}
}

// chkOptApprox tolerates AE's float32 round-trip (e.g. 1.8 → 1.79999995).
func chkOptApprox(t *testing.T, ver, label string, p *aep.Property, want float64) {
	t.Helper()
	if p == nil {
		t.Errorf("%s resaved: %s property nil", ver, label)
		return
	}
	got, ok := p.StaticValue.(float64)
	if !ok || abs(got-want) > 1e-4 {
		t.Errorf("%s resaved: %s = %v, want ~%g", ver, label, p.StaticValue, want)
	}
}

func TestNewCameraLight_AEShipGate_AE2020(t *testing.T) {
	runCameraLightGate(t, aep.TargetAE2020, ae2020(), "AE2020")
}

func TestNewCameraLight_AEShipGate_AE2025(t *testing.T) {
	runCameraLightGate(t, aep.TargetAE2025, ae2025(), "AE2025")
}
