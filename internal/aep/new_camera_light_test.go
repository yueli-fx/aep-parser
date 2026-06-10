// internal/aep/new_camera_light_test.go
//
// Go round-trip tests for NewCameraLayer / NewLightLayer (no AE). Proves the
// cloned template Layr splices in and survives WriteAEP → re-parse as a layer
// of the right type with the caller's name. AE acceptance is the ship-gate's job
// (new_camera_light_shipgate_test.go).
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestNewCameraLayer_RoundTrip(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2025)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	cam, err := aep.NewCameraLayer(comp, "Cam1")
	if err != nil {
		t.Fatalf("NewCameraLayer: %v", err)
	}
	if cam.Type != aep.LayerTypeCamera {
		t.Errorf("in-memory layer Type = %v, want Camera", cam.Type)
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
	var found *aep.Layer
	for _, l := range rc.Layers {
		if l.Name == "Cam1" {
			found = l
		}
	}
	if found == nil {
		t.Fatalf("camera layer not found after round-trip (have %d layers)", len(rc.Layers))
	}
	if found.Type != aep.LayerTypeCamera {
		t.Errorf("re-parsed Type = %v, want Camera", found.Type)
	}
}

func TestNewLightLayer_RoundTrip(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2025)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	lt, err := aep.NewLightLayer(comp, "Light1")
	if err != nil {
		t.Fatalf("NewLightLayer: %v", err)
	}
	if lt.Type != aep.LayerTypeLight {
		t.Errorf("in-memory layer Type = %v, want Light", lt.Type)
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
	var found *aep.Layer
	for _, l := range rc.Layers {
		if l.Name == "Light1" {
			found = l
		}
	}
	if found == nil {
		t.Fatalf("light layer not found after round-trip (have %d layers)", len(rc.Layers))
	}
	if found.Type != aep.LayerTypeLight {
		t.Errorf("re-parsed Type = %v, want Light", found.Type)
	}
}

func TestNewCameraLayer_EmptyName(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2025)
	comp, _ := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if _, err := aep.NewCameraLayer(comp, ""); err == nil {
		t.Error("NewCameraLayer empty name: want error, got nil")
	}
}
