package aep_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestCompositionLayerByID(t *testing.T) {
	comp := &aep.Composition{
		Layers: []*aep.Layer{
			{ID: 10, Name: "first"},
			{ID: 20, Name: "second"},
			{ID: 30, Name: "third"},
		},
	}
	if got := comp.LayerByID(20); got == nil || got.Name != "second" {
		t.Errorf("LayerByID(20) = %v, want layer named \"second\"", got)
	}
	if got := comp.LayerByID(999); got != nil {
		t.Errorf("LayerByID(999) = %v, want nil", got)
	}
	// Edge case: ID 0 is a valid "no parent" sentinel — should NOT match
	// any layer (real AE projects start IDs at 1).
	if got := comp.LayerByID(0); got != nil {
		t.Errorf("LayerByID(0) = %v, want nil", got)
	}
}

// TestCompositionSetters round-trips every Composition.SetX(): change
// the value, WriteAEP, re-parse, confirm persistence.
func TestCompositionSetters(t *testing.T) {
	data := buildMinimalAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	comp := proj.Compositions[0]

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("SetBGColor", comp.SetBGColor([3]uint8{0x33, 0x66, 0x99}))
	mustNoErr("SetSize", comp.SetSize(3840, 2160))
	mustNoErr("SetShutterAngle", comp.SetShutterAngle(360))
	mustNoErr("SetShutterPhase", comp.SetShutterPhase(-180))
	mustNoErr("SetMotionBlurAdaptiveSampleLimit", comp.SetMotionBlurAdaptiveSampleLimit(256))
	mustNoErr("SetMotionBlurSamplesPerFrame", comp.SetMotionBlurSamplesPerFrame(32))
	mustNoErr("SetWorkArea", comp.SetWorkArea(1.5, 4.5))

	if comp.BGColor != [3]uint8{0x33, 0x66, 0x99} {
		t.Errorf("BGColor in-mem = %v", comp.BGColor)
	}
	if comp.Width != 3840 || comp.Height != 2160 {
		t.Errorf("Size in-mem = %dx%d, want 3840x2160", comp.Width, comp.Height)
	}
	if comp.ShutterAngle != 360 {
		t.Errorf("ShutterAngle in-mem = %d", comp.ShutterAngle)
	}
	if comp.ShutterPhase != -180 {
		t.Errorf("ShutterPhase in-mem = %d", comp.ShutterPhase)
	}
	if comp.MotionBlurAdaptiveSampleLimit != 256 {
		t.Errorf("MotionBlurAdaptiveSampleLimit in-mem = %d", comp.MotionBlurAdaptiveSampleLimit)
	}
	if comp.MotionBlurSamplesPerFrame != 32 {
		t.Errorf("MotionBlurSamplesPerFrame in-mem = %d", comp.MotionBlurSamplesPerFrame)
	}
	if math.Abs(comp.WorkAreaStart-1.5) > 1e-6 || math.Abs(comp.WorkAreaEnd-4.5) > 1e-6 {
		t.Errorf("WorkArea in-mem = (%g, %g), want (1.5, 4.5)", comp.WorkAreaStart, comp.WorkAreaEnd)
	}

	// Round-trip
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	c2 := proj2.Compositions[0]
	if c2.BGColor != [3]uint8{0x33, 0x66, 0x99} {
		t.Errorf("after roundtrip: BGColor = %v", c2.BGColor)
	}
	if c2.Width != 3840 || c2.Height != 2160 {
		t.Errorf("after roundtrip: Size = %dx%d, want 3840x2160", c2.Width, c2.Height)
	}
	if c2.ShutterAngle != 360 {
		t.Errorf("after roundtrip: ShutterAngle = %d", c2.ShutterAngle)
	}
	if c2.ShutterPhase != -180 {
		t.Errorf("after roundtrip: ShutterPhase = %d", c2.ShutterPhase)
	}
	if c2.MotionBlurAdaptiveSampleLimit != 256 {
		t.Errorf("after roundtrip: MotionBlurAdaptiveSampleLimit = %d", c2.MotionBlurAdaptiveSampleLimit)
	}
	if c2.MotionBlurSamplesPerFrame != 32 {
		t.Errorf("after roundtrip: MotionBlurSamplesPerFrame = %d", c2.MotionBlurSamplesPerFrame)
	}
	if math.Abs(c2.WorkAreaStart-1.5) > 1e-6 || math.Abs(c2.WorkAreaEnd-4.5) > 1e-6 {
		t.Errorf("after roundtrip: WorkArea = (%g, %g)", c2.WorkAreaStart, c2.WorkAreaEnd)
	}
}

// TestCompositionNameAndTiming covers SetName (length-variable Utf8)
// + SetFrameRate + SetDuration (length-preserving cdta writes).
func TestCompositionNameAndTiming(t *testing.T) {
	data := buildMinimalAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	comp := proj.Compositions[0]

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("SetName", comp.SetName("Renamed Comp"))
	mustNoErr("SetFrameRate", comp.SetFrameRate(60.0))
	mustNoErr("SetDuration", comp.SetDuration(7.5))

	if comp.Name != "Renamed Comp" {
		t.Errorf("in-mem Name = %q", comp.Name)
	}
	if math.Abs(comp.FrameRate-60.0) > 1e-6 {
		t.Errorf("in-mem FrameRate = %g", comp.FrameRate)
	}
	if math.Abs(comp.Duration-7.5) > 1e-3 {
		t.Errorf("in-mem Duration = %g", comp.Duration)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	c2 := proj2.Compositions[0]
	if c2.Name != "Renamed Comp" {
		t.Errorf("roundtrip Name = %q", c2.Name)
	}
	if math.Abs(c2.FrameRate-60.0) > 1e-6 {
		t.Errorf("roundtrip FrameRate = %g", c2.FrameRate)
	}
	if math.Abs(c2.Duration-7.5) > 1e-3 {
		t.Errorf("roundtrip Duration = %g", c2.Duration)
	}
}

func TestCompositionSettersRejectMissingCdta(t *testing.T) {
	c := &aep.Composition{Name: "standalone"}
	if err := c.SetBGColor([3]uint8{1, 2, 3}); err == nil {
		t.Error("SetBGColor on comp without cdta: expected error")
	}
	if err := c.SetSize(1920, 1080); err == nil {
		t.Error("SetSize on comp without cdta: expected error")
	}
	if err := c.SetShutterAngle(180); err == nil {
		t.Error("SetShutterAngle on comp without cdta: expected error")
	}
	if err := c.SetWorkArea(0, 1); err == nil {
		t.Error("SetWorkArea on comp without cdta: expected error")
	}
}

// TestCompositionExtraFlagSetters covers SetFrameBlending /
// SetPreserveNestedResolution / SetPixelAspect against re_batch3.aep.
func TestCompositionExtraFlagSetters(t *testing.T) {
	proj, err := aep.FromReader(bytes.NewReader(buildMinimalAEP()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	comp := proj.Compositions[0]

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("SetFrameBlending", comp.SetFrameBlending(true))
	mustNoErr("SetPreserveNestedResolution", comp.SetPreserveNestedResolution(true))
	mustNoErr("SetPixelAspect", comp.SetPixelAspect(2.0))

	if math.Abs(comp.PixelAspect-2.0) > 1e-6 {
		t.Errorf("in-mem PixelAspect = %g", comp.PixelAspect)
	}

	// Round-trip.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if math.Abs(proj2.Compositions[0].PixelAspect-2.0) > 1e-6 {
		t.Errorf("roundtrip PixelAspect = %g", proj2.Compositions[0].PixelAspect)
	}
}
