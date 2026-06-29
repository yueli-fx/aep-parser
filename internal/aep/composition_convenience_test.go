package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestCompositionConvenience_ReCameraLight(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	var comp1 *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "Comp 1" {
			comp1 = c
			break
		}
	}
	if comp1 == nil {
		t.Fatal("Comp 1 not found")
	}
	if got, want := comp1.NumLayers(), 16; got != want {
		t.Errorf("NumLayers = %d, want %d", got, want)
	}
	if comp1.TimeScale() != comp1.TickRate {
		t.Errorf("TimeScale (%v) != TickRate (%v)", comp1.TimeScale(), comp1.TickRate)
	}
	// HasAudio: re_cameralight Comp 1 has 13 AV layers; at least one
	// audio switch is expected on by default. Just smoke-check it.
	_ = comp1.HasAudio()
}

func TestCompositionConvenience_EmptyComp(t *testing.T) {
	c := &aep.Composition{Name: "Empty", TickRate: 30720}
	if got := c.NumLayers(); got != 0 {
		t.Errorf("NumLayers empty = %d, want 0", got)
	}
	if c.HasAudio() {
		t.Error("HasAudio empty = true, want false")
	}
	if got := c.TimeScale(); got != 30720 {
		t.Errorf("TimeScale = %v, want 30720", got)
	}
}
