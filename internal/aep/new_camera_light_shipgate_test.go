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

	aep "github.com/example/aep-parser/internal/aep"
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
		{"SetLightIntensity", func() error { return lightLayer.SetLightIntensity(65) }},
		{"SetLightConeAngle", func() error { return lightLayer.SetLightConeAngle(72) }},
		{"SetLightConeFeather", func() error { return lightLayer.SetLightConeFeather(35) }},
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
		}
		if l.Name == "Light1" && l.Type == aep.LayerTypeLight {
			light = true
			if l.LightKind != aep.LightKindSpot {
				t.Errorf("%s resaved: Light 'Light1' LightKind = %v, want spot", ver, l.LightKind)
			}
			chkOpt(t, ver, "Light1.Intensity", l.LightIntensity(), 65)
			chkOpt(t, ver, "Light1.ConeAngle", l.LightConeAngle(), 72)
			chkOpt(t, ver, "Light1.ConeFeather", l.LightConeFeather(), 35)
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

func TestNewCameraLight_AEShipGate_AE2020(t *testing.T) {
	runCameraLightGate(t, aep.TargetAE2020, ae2020(), "AE2020")
}

func TestNewCameraLight_AEShipGate_AE2025(t *testing.T) {
	runCameraLightGate(t, aep.TargetAE2025, ae2025(), "AE2025")
}
