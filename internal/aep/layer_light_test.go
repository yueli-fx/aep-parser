package aep_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestLightKindReal(t *testing.T) {
	proj := openWave2AE24(t)
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_LIGHTS" && (comp == nil || c.ID > comp.ID) {
			comp = c
		}
	}
	if comp == nil {
		t.Fatalf("RE_LIGHTS comp missing")
	}
	byName := map[string]*aep.Layer{}
	for _, l := range comp.Layers {
		byName[l.Name] = l
	}
	cases := []struct {
		layerName string
		want      aep.LightKind
	}{
		{"light_parallel", aep.LightKindParallel},
		{"light_spot", aep.LightKindSpot},
		{"light_point", aep.LightKindPoint},
		{"light_ambient", aep.LightKindAmbient},
		{"light_spot_shadows_on", aep.LightKindSpot},
		{"light_spot_shadows_off", aep.LightKindSpot},
	}
	for _, c := range cases {
		l := byName[c.layerName]
		if l == nil {
			t.Errorf("layer %q missing", c.layerName)
			continue
		}
		if l.LightKind != c.want {
			t.Errorf("%s LightKind = %v, want %v", c.layerName, l.LightKind, c.want)
		}
	}
}

func TestSetLightKindRoundtrip(t *testing.T) {
	proj := openWave2AE24(t)
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_LIGHTS" && (comp == nil || c.ID > comp.ID) {
			comp = c
		}
	}
	var parallel *aep.Layer
	for _, l := range comp.Layers {
		if l.Name == "light_parallel" {
			parallel = l
			break
		}
	}
	if parallel == nil {
		t.Fatalf("light_parallel missing")
	}
	parallelID := parallel.ID
	// Flip Parallel → Ambient
	if err := parallel.SetLightKind(aep.LightKindAmbient); err != nil {
		t.Fatalf("SetLightKind: %v", err)
	}
	if parallel.LightKind != aep.LightKindAmbient {
		t.Errorf("after Set: LightKind = %v, want Ambient", parallel.LightKind)
	}

	// Roundtrip
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var found *aep.Layer
	for _, c := range proj2.Compositions {
		for _, l := range c.Layers {
			if l.ID == parallelID {
				found = l
				break
			}
		}
	}
	if found == nil {
		t.Fatalf("roundtrip: parallel layer (id=%d) missing", parallelID)
	}
	if found.LightKind != aep.LightKindAmbient {
		t.Errorf("roundtrip LightKind = %v, want Ambient", found.LightKind)
	}

	// Invalid kind rejected
	if err := parallel.SetLightKind(aep.LightKind(99)); err == nil {
		t.Error("invalid LightKind: expected error")
	}
}

func TestTimeRemapEnabledReal(t *testing.T) {
	proj := openWave2AE24(t)
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_TIMEREMAP" && (comp == nil || c.ID > comp.ID) {
			comp = c
		}
	}
	if comp == nil {
		t.Fatalf("RE_TIMEREMAP comp missing")
	}
	byName := map[string]*aep.Layer{}
	for _, l := range comp.Layers {
		byName[l.Name] = l
	}
	if l := byName["tr_baseline"]; l == nil {
		t.Error("tr_baseline missing")
	} else if l.TimeRemapEnabled() {
		t.Errorf("tr_baseline TimeRemapEnabled() = true, want false")
	}
	if l := byName["tr_remap_on"]; l == nil {
		t.Error("tr_remap_on missing")
	} else if !l.TimeRemapEnabled() {
		t.Errorf("tr_remap_on TimeRemapEnabled() = false, want true")
	}
}

// TestCameraLightAccessorsReal exercises Layer.CameraZoom() and friends
// against test_data/re_cameralight.aep (built by /tmp/re_cameralight.jsx
// in AE 2020 — see CLAUDE.md for fixture regeneration). Skips silently
// if the fixture is absent.
func TestCameraLightAccessorsReal(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_CL" {
			comp = c
			break
		}
	}
	if comp == nil {
		t.Fatal("RE_CL comp missing")
	}
	byName := map[string]*aep.Layer{}
	for _, l := range comp.Layers {
		byName[l.Name] = l
	}

	if cam := byName["MyCamera"]; cam == nil {
		t.Fatal("MyCamera missing")
	} else {
		check := func(label string, p *aep.Property, want float64) {
			t.Helper()
			if p == nil {
				t.Errorf("%s = nil", label)
				return
			}
			got, ok := p.StaticValue.(float64)
			if !ok {
				t.Errorf("%s static is %T, want float64", label, p.StaticValue)
				return
			}
			if math.Abs(got-want) > 1e-6 {
				t.Errorf("%s = %g, want %g", label, got, want)
			}
		}
		check("CameraZoom", cam.CameraZoom(), 1000)
		check("CameraAperture", cam.CameraAperture(), 50)
		check("CameraFocusDistance", cam.CameraFocusDistance(), 1500)
		check("CameraBlurLevel", cam.CameraBlurLevel(), 75)
		check("CameraDepthOfField", cam.CameraDepthOfField(), 1) // toggle = 1
	}

	if spot := byName["MySpot"]; spot == nil {
		t.Fatal("MySpot missing")
	} else {
		if p := spot.LightIntensity(); p == nil || p.StaticValue != 150.0 {
			t.Errorf("LightIntensity = %v, want 150", staticValueOf(p))
		}
		if p := spot.LightConeAngle(); p == nil || p.StaticValue != 80.0 {
			t.Errorf("LightConeAngle = %v, want 80", staticValueOf(p))
		}
		if p := spot.LightConeFeather(); p == nil || p.StaticValue != 40.0 {
			t.Errorf("LightConeFeather = %v, want 40", staticValueOf(p))
		}
		if p := spot.LightShadowDarkness(); p == nil || p.StaticValue != 60.0 {
			t.Errorf("LightShadowDarkness = %v, want 60", staticValueOf(p))
		}
		// LightColor is 4D — AE 2020 stores [A, R, G, B] in 0..255.
		if p := spot.LightColor(); p == nil || p.Components != 4 {
			t.Errorf("LightColor = %v (want 4D property)", p)
		}
	}
}

// TestCameraLightTypedSettersRoundtrip exercises every typed Set* method
// added in layer_accessors.go against re_cameralight.aep, verifying both
// in-memory update and roundtrip persistence.
func TestCameraLightTypedSettersRoundtrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_CL" {
			comp = c
			break
		}
	}
	if comp == nil {
		t.Fatal("RE_CL comp missing")
	}
	byName := map[string]*aep.Layer{}
	for _, l := range comp.Layers {
		byName[l.Name] = l
	}
	cam := byName["MyCamera"]
	spot := byName["MySpot"]
	amb := byName["MyAmbient"]
	if cam == nil || spot == nil || amb == nil {
		t.Fatalf("layers missing: cam=%v spot=%v amb=%v", cam, spot, amb)
	}

	// Camera scalar setters.
	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Errorf("%s: %v", label, err)
		}
	}
	mustNoErr("SetCameraZoom", cam.SetCameraZoom(2000))
	mustNoErr("SetCameraFocusDistance", cam.SetCameraFocusDistance(3000))
	mustNoErr("SetCameraAperture", cam.SetCameraAperture(100))
	mustNoErr("SetCameraBlurLevel", cam.SetCameraBlurLevel(50))
	mustNoErr("SetCameraDepthOfField(false)", cam.SetCameraDepthOfField(false))

	// Light scalar setters (use MyAmbient — has 9 static props).
	mustNoErr("SetLightIntensity", amb.SetLightIntensity(80))
	mustNoErr("SetLightConeAngle", amb.SetLightConeAngle(45))
	mustNoErr("SetLightConeFeather", amb.SetLightConeFeather(20))
	mustNoErr("SetLightFalloffType", amb.SetLightFalloffType(2))
	mustNoErr("SetLightFalloffStart", amb.SetLightFalloffStart(100))
	mustNoErr("SetLightFalloffDistance", amb.SetLightFalloffDistance(1000))
	mustNoErr("SetLightShadowDarkness", amb.SetLightShadowDarkness(75))
	mustNoErr("SetLightShadowDiffusion", amb.SetLightShadowDiffusion(15))
	mustNoErr("SetLightCastsShadows(true)", amb.SetLightCastsShadows(true))

	// 4D color setter on MySpot.
	wantColor := []float64{200, 100, 50, 255}
	mustNoErr("SetLightColor", spot.SetLightColor(wantColor))

	// In-memory mirror checks.
	if got := cam.CameraZoom().StaticValue.(float64); got != 2000 {
		t.Errorf("in-mem CameraZoom = %g, want 2000", got)
	}
	if got := amb.LightCastsShadows().StaticValue.(float64); got != 1 {
		t.Errorf("in-mem CastsShadows = %g, want 1", got)
	}
	if got, _ := spot.LightColor().StaticValue.([]float64); len(got) != 4 || got[0] != 200 {
		t.Errorf("in-mem LightColor = %v, want %v", got, wantColor)
	}

	// Roundtrip.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var reComp *aep.Composition
	for _, c := range re.Compositions {
		if c.Name == "RE_CL" {
			reComp = c
		}
	}
	reByName := map[string]*aep.Layer{}
	for _, l := range reComp.Layers {
		reByName[l.Name] = l
	}
	reCam := reByName["MyCamera"]
	reAmb := reByName["MyAmbient"]
	reSpot := reByName["MySpot"]
	if got := reCam.CameraZoom().StaticValue.(float64); got != 2000 {
		t.Errorf("post-roundtrip CameraZoom = %g, want 2000", got)
	}
	if got := reCam.CameraFocusDistance().StaticValue.(float64); got != 3000 {
		t.Errorf("post-roundtrip FocusDistance = %g, want 3000", got)
	}
	if got := reAmb.LightIntensity().StaticValue.(float64); got != 80 {
		t.Errorf("post-roundtrip LightIntensity = %g, want 80", got)
	}
	if got := reAmb.LightCastsShadows().StaticValue.(float64); got != 1 {
		t.Errorf("post-roundtrip CastsShadows = %g, want 1", got)
	}
	if got, _ := reSpot.LightColor().StaticValue.([]float64); len(got) != 4 || got[0] != 200 || got[3] != 255 {
		t.Errorf("post-roundtrip LightColor = %v, want %v", got, wantColor)
	}

	// Error path: typed setter on a layer without that property.
	if err := amb.SetCameraZoom(50); err == nil {
		t.Error("SetCameraZoom on light layer should error")
	}
	if err := cam.SetLightColor([]float64{1, 2, 3}); err == nil {
		t.Error("SetLightColor on camera layer should error")
	}
	if err := cam.SetIrisShape(5); err == nil {
		// Iris properties are absent in this fixture's MyCamera (no DoF iris)
		// so this setter should error with "property not present".
		t.Error("SetIrisShape on camera w/o Iris sub-tree should error")
	}
}
