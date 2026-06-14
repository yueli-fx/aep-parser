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
	// Iris*/Highlight* are synthesis-inserted from-scratch (AE elides these
	// DoF-bokeh controls, so the template has no slots — spliced + reset).
	if cam.IrisShape() == nil || cam.IrisHighlightSaturation() == nil {
		t.Fatal("fresh camera: Iris leaves missing — splice failed")
	}
	must("SetIrisShape", cam.SetIrisShape(4))
	must("SetIrisRotation", cam.SetIrisRotation(25))
	must("SetIrisRoundness", cam.SetIrisRoundness(60))
	must("SetIrisDiffractionFringe", cam.SetIrisDiffractionFringe(30))
	must("SetIrisHighlightGain", cam.SetIrisHighlightGain(40))
	must("SetIrisHighlightSaturation", cam.SetIrisHighlightSaturation(50))
	// Light options.
	must("SetLightIntensity", light.SetLightIntensity(65))
	must("SetLightConeAngle", light.SetLightConeAngle(72))
	must("SetLightConeFeather", light.SetLightConeFeather(35))
	// Light Color is synthesis-inserted from-scratch (AE elides the default
	// white, so the template has no slot): the leaf is spliced in + reset to
	// white, lighting up SetLightColor on a fresh light.
	if light.LightColor() == nil {
		t.Fatal("fresh light: LightColor() nil — leaf splice failed")
	}
	// Raw cdat convention: [A,R,G,B] in 0..255 (alpha first).
	must("SetLightColor", light.SetLightColor([]float64{255, 51, 102, 204}))
	must("SetLightFalloffType", light.SetLightFalloffType(2))
	must("SetLightFalloffStart", light.SetLightFalloffStart(100))
	must("SetLightFalloffDistance", light.SetLightFalloffDistance(750))
	must("SetLightCastsShadows", light.SetLightCastsShadows(true))
	must("SetLightShadowDarkness", light.SetLightShadowDarkness(80))
	must("SetLightShadowDiffusion", light.SetLightShadowDiffusion(15))

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
	chk("IrisShape", reCam.IrisShape(), 4)
	chk("IrisRotation", reCam.IrisRotation(), 25)
	chk("IrisRoundness", reCam.IrisRoundness(), 60)
	chk("IrisDiffractionFringe", reCam.IrisDiffractionFringe(), 30)
	chk("IrisHighlightGain", reCam.IrisHighlightGain(), 40)
	chk("IrisHighlightSaturation", reCam.IrisHighlightSaturation(), 50)
	chk("LightIntensity", reLight.LightIntensity(), 65)
	chk("LightConeAngle", reLight.LightConeAngle(), 72)
	chk("LightConeFeather", reLight.LightConeFeather(), 35)
	chk("LightFalloffType", reLight.LightFalloffType(), 2)
	chk("LightFalloffStart", reLight.LightFalloffStart(), 100)
	chk("LightFalloffDistance", reLight.LightFalloffDistance(), 750)
	chk("LightCastsShadows", reLight.LightCastsShadows(), 1)
	chk("LightShadowDarkness", reLight.LightShadowDarkness(), 80)
	chk("LightShadowDiffusion", reLight.LightShadowDiffusion(), 15)

	lc := reLight.LightColor()
	if lc == nil {
		t.Fatal("round-trip: LightColor() nil")
	}
	rgba, ok := lc.StaticValue.([]float64)
	if !ok || len(rgba) != 4 {
		t.Fatalf("round-trip: LightColor StaticValue = %v (%T), want 4 floats", lc.StaticValue, lc.StaticValue)
	}
	for i, want := range []float64{255, 51, 102, 204} {
		if rgba[i] != want {
			t.Errorf("round-trip: LightColor[%d] = %g, want %g", i, rgba[i], want)
		}
	}
}
