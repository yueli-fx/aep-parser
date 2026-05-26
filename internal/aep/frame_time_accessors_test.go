package aep_test

import (
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// Frame-time accessors are pure conversions over existing time fields,
// so we mostly verify the round-trip seconds ↔ frame relationship on
// a real fixture.

func TestFrameAccessors_Layer(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
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
	if comp1 == nil || comp1.FrameRate == 0 {
		t.Skip("Comp 1 not present or FrameRate=0")
	}
	if len(comp1.Layers) == 0 {
		t.Skip("Comp 1 has no layers")
	}
	l := comp1.Layers[0]
	if l.InPoint() < 0 {
		t.Errorf("InPoint = %v < 0; suspicious", l.InPoint())
	}
	if l.OutPoint() < l.InPoint() {
		t.Errorf("OutPoint %v < InPoint %v", l.OutPoint(), l.InPoint())
	}
	// FrameInPoint vs InPoint round-trip (within 1 frame).
	wantFrame := int(math.Round(l.InPoint() * comp1.FrameRate))
	if got := l.FrameInPoint(); got != wantFrame {
		t.Errorf("FrameInPoint = %d, want %d (InPoint=%v, fps=%v)", got, wantFrame, l.InPoint(), comp1.FrameRate)
	}
}

func TestFrameAccessors_LayerSetters(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	// Find any layer with an ldta we can write to.
	var l *aep.Layer
	for _, c := range proj.Compositions {
		if c.Name != "Comp 1" {
			continue
		}
		for _, ll := range c.Layers {
			if ll.Visible {
				l = ll
				break
			}
		}
		if l != nil {
			break
		}
	}
	if l == nil {
		t.Skip("no Comp 1 layer to test")
	}
	if err := l.SetFrameInPoint(30); err != nil {
		t.Fatalf("SetFrameInPoint(30): %v", err)
	}
	if got := l.FrameInPoint(); got != 30 {
		t.Errorf("after SetFrameInPoint(30): FrameInPoint = %d", got)
	}
	if err := l.SetFrameStartTime(5); err != nil {
		t.Fatalf("SetFrameStartTime(5): %v", err)
	}
	if got := l.FrameStartTime(); got != 5 {
		t.Errorf("after SetFrameStartTime(5): FrameStartTime = %d", got)
	}
}

func TestFrameAccessors_Composition(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	var c *aep.Composition
	for _, cc := range proj.Compositions {
		if cc.Name == "Comp 1" {
			c = cc
			break
		}
	}
	if c == nil || c.FrameRate == 0 {
		t.Skip("Comp 1 not present or FrameRate=0")
	}
	wantDur := int(math.Round(c.Duration * c.FrameRate))
	if got := c.FrameDuration(); got != wantDur {
		t.Errorf("FrameDuration = %d, want %d (Duration=%v, fps=%v)", got, wantDur, c.Duration, c.FrameRate)
	}
	wantWAStart := int(math.Round(c.WorkAreaStart * c.FrameRate))
	if got := c.WorkAreaStartFrame(); got != wantWAStart {
		t.Errorf("WorkAreaStartFrame = %d, want %d", got, wantWAStart)
	}
	// Round-trip set WorkArea via frames.
	if err := c.SetWorkAreaStartFrame(10); err != nil {
		t.Fatalf("SetWorkAreaStartFrame(10): %v", err)
	}
	if got := c.WorkAreaStartFrame(); got != 10 {
		t.Errorf("WorkAreaStartFrame = %d after set, want 10", got)
	}
	if err := c.SetWorkAreaDurationFrame(60); err != nil {
		t.Fatalf("SetWorkAreaDurationFrame(60): %v", err)
	}
	if got := c.WorkAreaDurationFrame(); got != 60 {
		t.Errorf("WorkAreaDurationFrame = %d after set, want 60", got)
	}
}

func TestFrameAccessors_KeyframeMarker(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	var kf *aep.Keyframe
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			for _, p := range l.Properties {
				if len(p.Keyframes) > 0 {
					kf = p.Keyframes[0]
					break
				}
			}
			if kf != nil {
				break
			}
		}
		if kf != nil {
			break
		}
	}
	if kf == nil {
		t.Skip("no keyframes in fixture")
	}
	// Smoke: FrameTime returns a sensible non-negative integer.
	if ft := kf.FrameTime(); ft < 0 {
		t.Errorf("Keyframe.FrameTime = %d (< 0)", ft)
	}
	// SetFrameTime round-trip.
	if err := kf.SetFrameTime(0); err != nil {
		t.Fatalf("SetFrameTime(0): %v", err)
	}
	if got := kf.FrameTime(); got != 0 {
		t.Errorf("Keyframe.FrameTime after Set(0) = %d", got)
	}
}

func TestFrameAccessors_StandaloneZero(t *testing.T) {
	// Layer with no comp owner: layerFps=0, frame accessors return 0,
	// setters error out cleanly.
	l := &aep.Layer{Name: "Solo"}
	if l.FrameInPoint() != 0 {
		t.Errorf("standalone FrameInPoint = %d, want 0", l.FrameInPoint())
	}
	if err := l.SetFrameInPoint(10); err == nil {
		t.Error("SetFrameInPoint on standalone layer: expected error")
	}
	// Keyframe with no compFps.
	k := &aep.Keyframe{}
	if k.FrameTime() != 0 {
		t.Errorf("standalone Keyframe.FrameTime = %d, want 0", k.FrameTime())
	}
	if err := k.SetFrameTime(5); err == nil {
		t.Error("Keyframe.SetFrameTime with compFps=0: expected error")
	}
}
