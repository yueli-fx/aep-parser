package aep_test

import (
	"bytes"
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestLayerPropertyAccessors(t *testing.T) {
	proj, err := aep.FromReader(bytes.NewReader(buildKeyframedAEP()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]

	// Generic lookup.
	if p := layer.PropertyByMatchName("ADBE Opacity"); p == nil {
		t.Error("PropertyByMatchName(Opacity) = nil, want a property")
	}
	if p := layer.PropertyByMatchName("does-not-exist"); p != nil {
		t.Errorf("PropertyByMatchName(unknown) = %v, want nil", p)
	}

	// Named accessors should return non-nil for properties present in the
	// synthetic AEP (Position + Opacity were added by buildKeyframedAEP).
	if p := layer.Position(); p == nil {
		t.Error("Position() = nil")
	} else if p.MatchName != aep.MatchNamePosition {
		t.Errorf("Position().MatchName = %q, want %q", p.MatchName, aep.MatchNamePosition)
	}
	if p := layer.Opacity(); p == nil {
		t.Error("Opacity() = nil")
	} else if p.MatchName != aep.MatchNameOpacity {
		t.Errorf("Opacity().MatchName = %q, want %q", p.MatchName, aep.MatchNameOpacity)
	}

	// Accessors for properties NOT in this synthetic file should return nil
	// rather than panic.
	if p := layer.Scale(); p != nil {
		t.Errorf("Scale() = %v, want nil (none added to synthetic AEP)", p)
	}
	if p := layer.Rotation(); p != nil {
		t.Errorf("Rotation() = %v, want nil", p)
	}
	if p := layer.AnchorPoint(); p != nil {
		t.Errorf("AnchorPoint() = %v, want nil", p)
	}
}

// TestLayerSetters round-trips every Layer.SetX(): flip the value, write
// the project back, re-parse, confirm the new value persisted. Covers
// flag-bit setters (Visible/Solo/Shy/…) and byte-field setters
// (BlendingMode/TrackMatte/Label/Quality/PreserveTransparency).
func TestLayerSetters(t *testing.T) {
	// Build via the standard keyframed fixture which produces a layer
	// with a full-length ldta — the setters need byte offsets up to
	// 0x6B (TrackMatte) so the minimal 0x2C ldta from buildExtendedAEP
	// wouldn't suffice.
	data := buildLdtaFullLengthAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]

	// Drive every setter to a non-default value, then re-read.
	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("SetVisible(false)", layer.SetVisible(false))
	mustNoErr("SetSolo(true)", layer.SetSolo(true))
	mustNoErr("SetShy(true)", layer.SetShy(true))
	mustNoErr("SetLocked(true)", layer.SetLocked(true))
	mustNoErr("SetEffectsEnabled(false)", layer.SetEffectsEnabled(false))
	mustNoErr("SetMotionBlur(true)", layer.SetMotionBlur(true))
	mustNoErr("SetAudioEnabled(false)", layer.SetAudioEnabled(false))
	mustNoErr("SetFrameBlendEnabled(true)", layer.SetFrameBlendEnabled(true))
	mustNoErr("SetCollapseTransform(true)", layer.SetCollapseTransform(true))
	mustNoErr("SetIs3D(true)", layer.SetIs3D(true))
	mustNoErr("SetIsAdjust(true)", layer.SetIsAdjust(true))
	mustNoErr("SetIsGuide(true)", layer.SetIsGuide(true))
	mustNoErr("SetMarkersLocked(true)", layer.SetMarkersLocked(true))
	mustNoErr("SetSamplingBicubic(true)", layer.SetSamplingBicubic(true))
	mustNoErr("SetFrameBlendPixelMotion(true)", layer.SetFrameBlendPixelMotion(true))
	mustNoErr("SetBlendingMode", layer.SetBlendingMode(aep.BlendingModeMultiply))
	mustNoErr("SetTrackMatte", layer.SetTrackMatte(aep.TrackMatteAlphaInverse))
	mustNoErr("SetLabel", layer.SetLabel(12))
	mustNoErr("SetQuality", layer.SetQuality(aep.LayerQualityBest))
	mustNoErr("SetPreserveTransparency(true)", layer.SetPreserveTransparency(true))

	// In-memory Go fields should already reflect new values.
	checks := []struct {
		name string
		got  any
		want any
	}{
		{"Visible", layer.Visible, false},
		{"Solo", layer.Solo, true},
		{"Shy", layer.Shy, true},
		{"Locked", layer.Locked, true},
		{"EffectsEnabled", layer.EffectsEnabled, false},
		{"MotionBlur", layer.MotionBlur, true},
		{"AudioEnabled", layer.AudioEnabled, false},
		{"FrameBlendEnabled", layer.FrameBlendEnabled, true},
		{"CollapseTransform", layer.CollapseTransform, true},
		{"Is3D", layer.Is3D, true},
		{"IsAdjust", layer.IsAdjust, true},
		{"IsGuide", layer.IsGuide, true},
		{"MarkersLocked", layer.MarkersLocked, true},
		{"SamplingBicubic", layer.SamplingBicubic, true},
		{"FrameBlendPixelMotion", layer.FrameBlendPixelMotion, true},
		{"BlendingMode", layer.BlendingMode, aep.BlendingModeMultiply},
		{"TrackMatte", layer.TrackMatte, aep.TrackMatteAlphaInverse},
		{"Label", layer.Label, uint8(12)},
		{"Quality", layer.Quality, aep.LayerQualityBest},
		{"PreserveTransparency", layer.PreserveTransparency, true},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("in-memory after Set: %s = %v, want %v", c.name, c.got, c.want)
		}
	}

	// Round-trip through bytes — re-parsing the output should observe
	// the same values, which proves the setters wrote the right bits.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	l2 := proj2.Compositions[0].Layers[0]
	rt := []struct {
		name string
		got  any
		want any
	}{
		{"Visible", l2.Visible, false},
		{"Solo", l2.Solo, true},
		{"Shy", l2.Shy, true},
		{"Locked", l2.Locked, true},
		{"EffectsEnabled", l2.EffectsEnabled, false},
		{"MotionBlur", l2.MotionBlur, true},
		{"AudioEnabled", l2.AudioEnabled, false},
		{"FrameBlendEnabled", l2.FrameBlendEnabled, true},
		{"CollapseTransform", l2.CollapseTransform, true},
		{"Is3D", l2.Is3D, true},
		{"IsAdjust", l2.IsAdjust, true},
		{"IsGuide", l2.IsGuide, true},
		{"MarkersLocked", l2.MarkersLocked, true},
		{"SamplingBicubic", l2.SamplingBicubic, true},
		{"FrameBlendPixelMotion", l2.FrameBlendPixelMotion, true},
		{"BlendingMode", l2.BlendingMode, aep.BlendingModeMultiply},
		{"TrackMatte", l2.TrackMatte, aep.TrackMatteAlphaInverse},
		{"Label", l2.Label, uint8(12)},
		{"Quality", l2.Quality, aep.LayerQualityBest},
		{"PreserveTransparency", l2.PreserveTransparency, true},
	}
	for _, c := range rt {
		if c.got != c.want {
			t.Errorf("after round-trip: %s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

// TestLayerTimeSetters covers SetStartTime / SetInPoint / SetOutPoint /
// SetStretch against the synthetic full-length ldta fixture, with
// WriteAEP roundtrip.
func TestLayerTimeSetters(t *testing.T) {
	data := buildLdtaFullLengthAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("SetStartTime", layer.SetStartTime(1.25))
	mustNoErr("SetInPoint", layer.SetInPoint(0.5))
	mustNoErr("SetOutPoint", layer.SetOutPoint(3.75))
	mustNoErr("SetStretch", layer.SetStretch(2.0))

	if math.Abs(layer.StartTime-1.25) > 1e-3 {
		t.Errorf("in-mem StartTime = %g, want 1.25", layer.StartTime)
	}
	wantDur := 3.75 - 0.5
	if math.Abs(layer.Duration-wantDur) > 1e-3 {
		t.Errorf("in-mem Duration = %g, want %g", layer.Duration, wantDur)
	}
	if math.Abs(layer.Stretch-2.0) > 1e-3 {
		t.Errorf("in-mem Stretch = %g, want 2.0", layer.Stretch)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	l2 := proj2.Compositions[0].Layers[0]
	if math.Abs(l2.StartTime-1.25) > 1e-3 {
		t.Errorf("roundtrip StartTime = %g", l2.StartTime)
	}
	if math.Abs(l2.Duration-wantDur) > 1e-3 {
		t.Errorf("roundtrip Duration = %g, want %g", l2.Duration, wantDur)
	}
	if math.Abs(l2.Stretch-2.0) > 1e-3 {
		t.Errorf("roundtrip Stretch = %g", l2.Stretch)
	}
}

// TestLayerSetParentAndSource covers SetParent + SetSource length-
// preserving 4-byte writes + ID validation.
func TestLayerSetParentAndSource(t *testing.T) {
	proj, err := aep.FromReader(bytes.NewReader(buildLdtaFullLengthAEP()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]

	// SetParent=0 (no parent) should always succeed.
	if err := layer.SetParent(0); err != nil {
		t.Errorf("SetParent(0): %v", err)
	}
	if layer.ParentID != 0 {
		t.Errorf("ParentID after SetParent(0) = %d", layer.ParentID)
	}

	// Self-parenting rejected (only meaningful when layer has a non-zero ID;
	// the buildKeyframedAEP fixture writes a real ID).
	proj2, _ := aep.FromReader(bytes.NewReader(buildKeyframedAEP()))
	layer2 := proj2.Compositions[0].Layers[0]
	if layer2.ID != 0 {
		if err := layer2.SetParent(layer2.ID); err == nil {
			t.Error("SetParent(own ID): expected error")
		}
	}

	// SetParent with a non-existent ID inside a real comp → error.
	if err := layer2.SetParent(99999); err == nil {
		t.Error("SetParent(non-existent ID): expected error")
	}

	// SetSource with sourceID=0 should always work.
	if err := layer.SetSource(0); err != nil {
		t.Errorf("SetSource(0): %v", err)
	}
	if layer.SourceID != 0 {
		t.Errorf("SourceID after SetSource(0) = %d", layer.SourceID)
	}

	// Roundtrip the ParentID=0 change through WriteAEP.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj3, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if proj3.Compositions[0].Layers[0].ParentID != 0 {
		t.Errorf("roundtrip ParentID = %d", proj3.Compositions[0].Layers[0].ParentID)
	}
}

// TestLayerSetAutoOrient covers the 3-bit mutually-exclusive
// auto-orient enum write.
func TestLayerSetAutoOrient(t *testing.T) {
	proj, err := aep.FromReader(bytes.NewReader(buildLdtaFullLengthAEP()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	cases := []aep.AutoOrientType{
		aep.AutoOrientCharactersTowardCamera,
		aep.AutoOrientCameraOrPointOfInterest,
		aep.AutoOrientAlongPath,
		aep.AutoOrientNone,
	}
	for _, want := range cases {
		if err := layer.SetAutoOrient(want); err != nil {
			t.Fatalf("SetAutoOrient(%s): %v", want, err)
		}
		if layer.AutoOrient != want {
			t.Errorf("in-mem after Set: AutoOrient = %s, want %s", layer.AutoOrient, want)
		}
		var buf bytes.Buffer
		if err := proj.WriteAEP(&buf); err != nil {
			t.Fatalf("WriteAEP: %v", err)
		}
		proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatalf("re-parse: %v", err)
		}
		if got := proj2.Compositions[0].Layers[0].AutoOrient; got != want {
			t.Errorf("roundtrip: AutoOrient = %s, want %s", got, want)
		}
	}
}

// TestLayerSetNameAndComment covers length-variable text writes on
// the layer's Utf8 (name) and cmta (comment) chunks. Round-trips
// through WriteAEP to confirm the resized chunks parse back.
func TestLayerSetNameAndComment(t *testing.T) {
	data := buildLdtaFullLengthAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]

	// The synthetic fixture has no Utf8 nor cmta chunk by default;
	// SetName should refuse (no Utf8 to mutate), SetComment should
	// succeed (it can insert a fresh cmta).
	if err := layer.SetName("X"); err == nil {
		t.Error("SetName on layer without Utf8: expected error")
	}
	if err := layer.SetComment("first line\nsecond line"); err != nil {
		t.Fatalf("SetComment (insert): %v", err)
	}
	if layer.Comment != "first line\nsecond line" {
		t.Errorf("after SetComment: Comment = %q", layer.Comment)
	}

	// Round-trip with the inserted cmta and verify reparse.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	l2 := proj2.Compositions[0].Layers[0]
	if l2.Comment != "first line\nsecond line" {
		t.Errorf("after roundtrip: Comment = %q", l2.Comment)
	}

	// Now a layer with a real Utf8 chunk: re_text.aep variants.
	rp, err := aep.Open("../../test_data/re_text.aep")
	if err != nil {
		t.Skipf("re_text.aep not present; rerun JSX")
	}
	var rc *aep.Composition
	for _, c := range rp.Compositions {
		if c.Name == "RE_TEXT" {
			if rc == nil || len(c.Layers) > len(rc.Layers) {
				rc = c
			}
		}
	}
	if rc == nil {
		t.Fatal("RE_TEXT comp missing from re_text.aep")
	}
	var target *aep.Layer
	for _, l := range rc.Layers {
		if l.Name == "baseline_A" {
			target = l
			break
		}
	}
	if target == nil {
		t.Skip("baseline_A layer not found")
	}
	if err := target.SetName("renamed-baseline"); err != nil {
		t.Fatalf("SetName: %v", err)
	}
	if err := target.SetComment("changed via SetComment"); err != nil {
		t.Fatalf("SetComment: %v", err)
	}

	var rbuf bytes.Buffer
	if err := rp.WriteAEP(&rbuf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	rp2, err := aep.FromReader(bytes.NewReader(rbuf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var found *aep.Layer
	for _, c := range rp2.Compositions {
		if c.Name == "RE_TEXT" {
			for _, l := range c.Layers {
				if l.Name == "renamed-baseline" {
					found = l
					break
				}
			}
		}
	}
	if found == nil {
		t.Fatal("after roundtrip: layer name didn't persist")
	}
	if found.Comment != "changed via SetComment" {
		t.Errorf("after roundtrip: Comment = %q", found.Comment)
	}
}

// TestLayerSettersRejectMissingLdta covers the error path: a layer
// built outside the parser has no underlying ldta chunk, so every
// setter must refuse with an error instead of panicking.
func TestLayerSettersRejectMissingLdta(t *testing.T) {
	l := &aep.Layer{Name: "standalone"}
	if err := l.SetVisible(false); err == nil {
		t.Error("SetVisible on layer without ldta: expected error, got nil")
	}
	if err := l.SetBlendingMode(aep.BlendingModeAdd); err == nil {
		t.Error("SetBlendingMode on layer without ldta: expected error, got nil")
	}
	if err := l.SetLabel(5); err == nil {
		t.Error("SetLabel on layer without ldta: expected error, got nil")
	}
}
