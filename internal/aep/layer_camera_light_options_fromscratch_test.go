// internal/aep/layer_camera_light_options_fromscratch_test.go
//
// Proves the Camera/Light Options setters work on a FROM-SCRATCH layer
// (NewCameraLayer / NewLightLayer) — not just on a parsed fixture. The fresh
// layer's cloned template now has its property tree parsed + leaf backrefs wired
// (newTemplatedLayer), so SetCameraZoom / SetLightIntensity / … reach the cdats
// and survive a write→re-parse round-trip.
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestCameraLightOptions_FromScratch_Roundtrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "OPTS", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	cam, err := aep.NewCameraLayer(comp, "Cam")
	if err != nil {
		t.Fatalf("NewCameraLayer: %v", err)
	}
	light, err := aep.NewLightLayer(comp, "Lite")
	if err != nil {
		t.Fatalf("NewLightLayer: %v", err)
	}

	must := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	// Camera options.
	must("SetCameraZoom", cam.SetCameraZoom(850))
	must("SetCameraDepthOfField", cam.SetCameraDepthOfField(true))
	must("SetCameraFocusDistance", cam.SetCameraFocusDistance(1200))
	must("SetCameraAperture", cam.SetCameraAperture(180))
	// Light options.
	must("SetLightIntensity", light.SetLightIntensity(65))
	must("SetLightConeAngle", light.SetLightConeAngle(72))
	must("SetLightConeFeather", light.SetLightConeFeather(35))

	// In-memory mirror (the fresh layer's parsed property tree reflects the write).
	if got := cam.CameraZoom().StaticValue.(float64); got != 850 {
		t.Errorf("in-mem CameraZoom = %g, want 850", got)
	}
	if got := light.LightIntensity().StaticValue.(float64); got != 65 {
		t.Errorf("in-mem LightIntensity = %g, want 65", got)
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rc := re.Compositions[0]
	reCam, reLight := rc.LayerByName("Cam"), rc.LayerByName("Lite")
	if reCam == nil || reLight == nil {
		t.Fatalf("round-trip: layer missing (cam=%v light=%v)", reCam, reLight)
	}

	chk := func(label string, p *aep.Property, want float64) {
		t.Helper()
		if p == nil {
			t.Errorf("%s: property nil after round-trip", label)
			return
		}
		got, ok := p.StaticValue.(float64)
		if !ok || got != want {
			t.Errorf("%s = %v, want %g", label, p.StaticValue, want)
		}
	}
	chk("CameraZoom", reCam.CameraZoom(), 850)
	chk("CameraDepthOfField", reCam.CameraDepthOfField(), 1)
	chk("CameraFocusDistance", reCam.CameraFocusDistance(), 1200)
	chk("CameraAperture", reCam.CameraAperture(), 180)
	chk("LightIntensity", reLight.LightIntensity(), 65)
	chk("LightConeAngle", reLight.LightConeAngle(), 72)
	chk("LightConeFeather", reLight.LightConeFeather(), 35)
}
