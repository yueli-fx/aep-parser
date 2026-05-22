package aep_test

import (
	"bytes"
	"fmt"
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestLayerFieldDecoding(t *testing.T) {
	rb := &rifxBuilder{}
	ldta := buildLdta160(ldtaOpts{
		LayerID: 42, ParentID: 100, SourceID: 7,
		Quality:        uint16(aep.LayerQualityBest),
		StartTime:      0.0,
		InPoint:        0.0,
		OutPoint:       5.0,
		StretchDividend: 1,
		StretchDivisor:  1,
		Flag25:         0x40,                                // bit6 = sampling bicubic
		Flag26:         0x80 | 0x08 | 0x02 | 0x10,           // null, solo, adjust, markers-locked
		Flag27:         0x80 | 0x40 | 0x20 | 0x08 | 0x04 | 0x02 | 0x01, // collapse, shy, lock, motion-blur, fx, audio, visible
		Label:          8,
		BlendingMode:   byte(aep.BlendingModeOverlay),
		PreserveTransparency: true,
		TrackMatte:     byte(aep.TrackMatteAlphaInverse),
		TypeByte:       0,
	})
	// Empty property tree just so parseProperties doesn't fail.
	tdgp := rb.listChunk("LIST", "tdgp", rb.chunk("tdmn", []byte("ADBE Group End")))
	data := wrapLayerInComp(rb, ldta, tdgp)

	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if len(proj.Compositions) != 1 || len(proj.Compositions[0].Layers) != 1 {
		t.Fatal("comp/layer missing")
	}
	l := proj.Compositions[0].Layers[0]

	if l.ID != 42 {
		t.Errorf("ID = %d, want 42", l.ID)
	}
	if l.ParentID != 100 {
		t.Errorf("ParentID = %d, want 100", l.ParentID)
	}
	if l.SourceID != 7 {
		t.Errorf("SourceID = %d, want 7", l.SourceID)
	}
	if l.Quality != aep.LayerQualityBest {
		t.Errorf("Quality = %d, want Best (2)", l.Quality)
	}
	if l.Label != 8 {
		t.Errorf("Label = %d, want 8", l.Label)
	}
	if l.BlendingMode != aep.BlendingModeOverlay {
		t.Errorf("BlendingMode = %d, want Overlay (7)", l.BlendingMode)
	}
	if !l.PreserveTransparency {
		t.Error("PreserveTransparency = false, want true")
	}
	if l.TrackMatte != aep.TrackMatteAlphaInverse {
		t.Errorf("TrackMatte = %d, want AlphaInverse (2)", l.TrackMatte)
	}
	if math.Abs(l.Duration-5.0) > 1e-6 {
		t.Errorf("Duration = %v, want 5.0", l.Duration)
	}
	if l.Stretch != 1.0 {
		t.Errorf("Stretch = %v, want 1.0", l.Stretch)
	}

	// Flag bits
	checks := []struct {
		got  bool
		want bool
		name string
	}{
		{l.SamplingBicubic, true, "SamplingBicubic"},
		{l.IsNull, true, "IsNull"},
		{l.Solo, true, "Solo"},
		{l.IsAdjust, true, "IsAdjust"},
		{l.MarkersLocked, true, "MarkersLocked"},
		{l.CollapseTransform, true, "CollapseTransform"},
		{l.Shy, true, "Shy"},
		{l.Locked, true, "Locked"},
		{l.MotionBlur, true, "MotionBlur"},
		{l.EffectsEnabled, true, "EffectsEnabled"},
		{l.AudioEnabled, true, "AudioEnabled"},
		{l.Visible, true, "Visible"},
		{l.Is3D, false, "Is3D"},
		{l.IsGuide, false, "IsGuide"},
		{l.FrameBlendEnabled, false, "FrameBlendEnabled"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestLayerComment(t *testing.T) {
	rb := &rifxBuilder{}
	ldta := buildLdta160(ldtaOpts{
		LayerID: 1, SourceID: 1,
		StretchDividend: 1, StretchDivisor: 1,
		OutPoint: 1.0,
	})
	tdgp := rb.listChunk("LIST", "tdgp", rb.chunk("tdmn", []byte("ADBE Group End")))
	// cmta as AE writes it: UTF-8 with CRLF, double-NUL terminator.
	cmta := rb.chunk("cmta", []byte("line1\r\nline2\x00\x00"))
	data := wrapLayerInComp(rb, ldta, tdgp, cmta)

	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	l := proj.Compositions[0].Layers[0]
	if l.Comment != "line1\nline2" {
		t.Errorf("Comment = %q, want %q", l.Comment, "line1\nline2")
	}
}

func TestLayerCommentEmpty(t *testing.T) {
	// No cmta sibling at all → empty.
	rb := &rifxBuilder{}
	ldta := buildLdta160(ldtaOpts{
		LayerID: 1, SourceID: 1,
		StretchDividend: 1, StretchDivisor: 1,
		OutPoint: 1.0,
	})
	tdgp := rb.listChunk("LIST", "tdgp", rb.chunk("tdmn", []byte("ADBE Group End")))
	data := wrapLayerInComp(rb, ldta, tdgp)
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if got := proj.Compositions[0].Layers[0].Comment; got != "" {
		t.Errorf("Comment = %q, want \"\"", got)
	}
}

func TestLayerAutoOrient(t *testing.T) {
	// Bit layout (per py-aep LdtaChunk):
	//   ldta[0x25] bit 4 (0x10) = CharactersTowardCamera
	//   ldta[0x26] bit 5 (0x20) = CameraOrPointOfInterest
	//   ldta[0x26] bit 0 (0x01) = AlongPath
	// AE enforces mutual exclusivity — only one bit is ever set.
	cases := []struct {
		name   string
		flag25 byte
		flag26 byte
		want   aep.AutoOrientType
	}{
		{"none", 0, 0, aep.AutoOrientNone},
		{"along-path", 0, 0x01, aep.AutoOrientAlongPath},
		{"camera-or-poi", 0, 0x20, aep.AutoOrientCameraOrPointOfInterest},
		{"characters-toward-camera", 0x10, 0, aep.AutoOrientCharactersTowardCamera},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rb := &rifxBuilder{}
			ldta := buildLdta160(ldtaOpts{
				LayerID: 1, SourceID: 1,
				StretchDividend: 1, StretchDivisor: 1,
				OutPoint: 1.0,
				Flag25:   c.flag25,
				Flag26:   c.flag26,
			})
			tdgp := rb.listChunk("LIST", "tdgp", rb.chunk("tdmn", []byte("ADBE Group End")))
			data := wrapLayerInComp(rb, ldta, tdgp)
			proj, err := aep.FromReader(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("FromReader: %v", err)
			}
			got := proj.Compositions[0].Layers[0].AutoOrient
			if got != c.want {
				t.Errorf("AutoOrient = %v, want %v", got, c.want)
			}
		})
	}
}

func TestLayerParent(t *testing.T) {
	// Three sibling layers in one comp: L10 (no parent), L20 (parent=L10),
	// L30 (parent=999 — orphan, no such layer).
	cases := []struct {
		name        string
		layerIndex  int
		wantNil     bool
		wantParent  uint32 // expected parent's ID when wantNil == false
	}{
		{"root layer has no parent", 0, true, 0},
		{"child resolves to parent", 1, false, 10},
		{"orphan reference returns nil", 2, true, 0},
	}

	rb := &rifxBuilder{}
	mk := func(id, parentID uint32) []byte {
		return buildLdta160(ldtaOpts{
			LayerID: id, ParentID: parentID, SourceID: 1,
			StretchDividend: 1, StretchDivisor: 1,
			OutPoint: 1.0,
		})
	}
	data := buildMultiLayerComp(rb, [][]byte{
		mk(10, 0),
		mk(20, 10),
		mk(30, 999),
	})
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layers := proj.Compositions[0].Layers
	if len(layers) != 3 {
		t.Fatalf("expected 3 layers, got %d", len(layers))
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := layers[c.layerIndex].Parent()
			if c.wantNil {
				if got != nil {
					t.Errorf("Parent() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("Parent() = nil, want layer with ID %d", c.wantParent)
			}
			if got.ID != c.wantParent {
				t.Errorf("Parent().ID = %d, want %d", got.ID, c.wantParent)
			}
		})
	}
}

func TestLayerParentNoComp(t *testing.T) {
	// Synthetic layer with no owning comp — Parent() must safely return nil.
	l := &aep.Layer{ID: 1, ParentID: 1}
	if got := l.Parent(); got != nil {
		t.Errorf("Parent() = %v, want nil", got)
	}
}

func TestLayerSourceComposition(t *testing.T) {
	// Project layout:
	//   Comp1 (id=1) — holds 3 layers
	//     L0: SourceID=2   → points to Comp2 (valid pre-comp)
	//     L1: SourceID=3   → points to Footage (non-comp source)
	//     L2: SourceID=999 → missing item
	//   Comp2 (id=2) — empty pre-comp target
	//   Footage (id=3)
	rb := &rifxBuilder{}
	mk := func(id, sourceID uint32) []byte {
		return buildLdta160(ldtaOpts{
			LayerID: id, SourceID: sourceID,
			StretchDividend: 1, StretchDivisor: 1,
			OutPoint: 1.0,
		})
	}
	data := buildMultiCompProject(rb, 2, [][]byte{
		mk(10, 2),
		mk(20, 3),
		mk(30, 999),
	})
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if len(proj.Compositions) != 2 {
		t.Fatalf("expected 2 comps, got %d", len(proj.Compositions))
	}
	if len(proj.Footage) != 1 {
		t.Fatalf("expected 1 footage, got %d", len(proj.Footage))
	}
	layers := proj.Compositions[0].Layers
	if len(layers) != 3 {
		t.Fatalf("expected 3 layers in comp1, got %d", len(layers))
	}

	cases := []struct {
		name       string
		layerIndex int
		wantNil    bool
		wantCompID uint32
	}{
		{"valid pre-comp source", 0, false, 2},
		{"footage source returns nil", 1, true, 0},
		{"missing source returns nil", 2, true, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := layers[c.layerIndex].SourceComposition()
			if c.wantNil {
				if got != nil {
					t.Errorf("SourceComposition() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("SourceComposition() = nil, want comp with ID %d", c.wantCompID)
			}
			if got.ID != c.wantCompID {
				t.Errorf("SourceComposition().ID = %d, want %d", got.ID, c.wantCompID)
			}
		})
	}
}

func TestLayerSourceCompositionNoOwner(t *testing.T) {
	// Synthetic layer with no owning comp/project — must return nil safely.
	l := &aep.Layer{ID: 1, SourceID: 1}
	if got := l.SourceComposition(); got != nil {
		t.Errorf("SourceComposition() = %v, want nil", got)
	}
}

func TestJSONParentName(t *testing.T) {
	// Same scaffolding as TestLayerParent: L10 (root), L20 (parent=L10),
	// L30 (parent=999, orphan). After JSON marshal:
	//   L10 → no parent_id, no parent_name
	//   L20 → parent_id=10, parent_name="" (no Utf8 name on L10 here)
	//   L30 → parent_id=999, no parent_name (orphan)
	// Then a second pass with a Utf8 name on L10 to verify parent_name
	// actually resolves to the parent's Name.
	rb := &rifxBuilder{}
	mk := func(id, parentID uint32) []byte {
		return buildLdta160(ldtaOpts{
			LayerID: id, ParentID: parentID, SourceID: 1,
			StretchDividend: 1, StretchDivisor: 1,
			OutPoint: 1.0,
		})
	}
	data := buildMultiLayerComp(rb, [][]byte{mk(10, 0), mk(20, 10), mk(30, 999)})
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	// Assign a Name to L10 directly on the parsed model — equivalent to a
	// Utf8 sibling in the Layr LIST, but simpler for the table here.
	proj.Compositions[0].Layers[0].Name = "background"

	jp := proj.ToJSON()
	jl := jp.Compositions[0].Layers

	cases := []struct {
		idx            int
		wantParentID   uint32
		wantParentName string
	}{
		{0, 0, ""},            // root: no parent
		{1, 10, "background"}, // child: resolved
		{2, 999, ""},          // orphan: parent_id present, parent_name empty
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("layer_%d", c.idx), func(t *testing.T) {
			if jl[c.idx].ParentID != c.wantParentID {
				t.Errorf("ParentID = %d, want %d", jl[c.idx].ParentID, c.wantParentID)
			}
			if jl[c.idx].ParentName != c.wantParentName {
				t.Errorf("ParentName = %q, want %q", jl[c.idx].ParentName, c.wantParentName)
			}
		})
	}
}

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

func TestTrackMatteLayerReal(t *testing.T) {
	proj, comp := openTrackMatteAE24(t)

	// All 4 source / baseline layers have no track matte set
	for _, name := range []string{"solidA", "solidB", "solidC", "mt_baseline"} {
		l := layerBySourceName(proj, comp, name)
		if l == nil {
			t.Errorf("layer with source=%q missing", name)
			continue
		}
		if l.TrackMatteLayerID != 0 {
			t.Errorf("%s TrackMatteLayerID = %d, want 0 (no explicit matte)", name, l.TrackMatteLayerID)
		}
		if l.TrackMatte != aep.TrackMatteNone {
			t.Errorf("%s TrackMatte = %v, want None", name, l.TrackMatte)
		}
		if l.TrackMatteLayer() != nil {
			t.Errorf("%s TrackMatteLayer() = %v, want nil", name, l.TrackMatteLayer())
		}
	}

	type matteCase struct {
		layerSource    string // identify dependent layer by its source name
		wantMode       aep.TrackMatteType
		wantMatteSrc   string // source footage name of the matte-source layer
	}
	cases := []matteCase{
		{"mt_alpha_to_solidA", aep.TrackMatteAlpha, "solidA"},
		{"mt_luma_to_solidB", aep.TrackMatteLuma, "solidB"},
		{"mt_alphainv_to_solidC", aep.TrackMatteAlphaInverse, "solidC"},
	}
	for _, c := range cases {
		l := layerBySourceName(proj, comp, c.layerSource)
		if l == nil {
			t.Errorf("layer with source=%q missing", c.layerSource)
			continue
		}
		if l.TrackMatte != c.wantMode {
			t.Errorf("%s TrackMatte = %v, want %v", c.layerSource, l.TrackMatte, c.wantMode)
		}
		if l.TrackMatteLayerID == 0 {
			t.Errorf("%s TrackMatteLayerID = 0, want non-zero", c.layerSource)
			continue
		}
		src := l.TrackMatteLayer()
		if src == nil {
			t.Errorf("%s TrackMatteLayer() = nil, want layer with source=%q", c.layerSource, c.wantMatteSrc)
			continue
		}
		wantSrc := layerBySourceName(proj, comp, c.wantMatteSrc)
		if wantSrc == nil || src.ID != wantSrc.ID {
			t.Errorf("%s TrackMatteLayer().ID = %d, want layer with source=%q (id=%d)",
				c.layerSource, src.ID, c.wantMatteSrc,
				func() uint32 { if wantSrc != nil { return wantSrc.ID }; return 0 }())
		}
	}
}

func TestSetTrackMatteLayerRoundtrip(t *testing.T) {
	proj, comp := openTrackMatteAE24(t)
	mtBaseline := layerBySourceName(proj, comp, "mt_baseline")
	solidA := layerBySourceName(proj, comp, "solidA")
	if mtBaseline == nil || solidA == nil {
		t.Fatalf("fixture layers missing (mtBaseline=%v solidA=%v)", mtBaseline, solidA)
	}
	if mtBaseline.TrackMatteLayerID != 0 {
		t.Fatalf("mt_baseline should start with no matte; got TMLayerID=%d", mtBaseline.TrackMatteLayerID)
	}

	// Assign solidA as matte source for mt_baseline with Luma mode.
	if err := mtBaseline.SetTrackMatteLayer(solidA.ID, aep.TrackMatteLuma); err != nil {
		t.Fatalf("SetTrackMatteLayer: %v", err)
	}
	if mtBaseline.TrackMatteLayerID != solidA.ID {
		t.Errorf("after Set: TMLayerID = %d, want %d", mtBaseline.TrackMatteLayerID, solidA.ID)
	}
	if mtBaseline.TrackMatte != aep.TrackMatteLuma {
		t.Errorf("after Set: TrackMatte = %v, want Luma", mtBaseline.TrackMatte)
	}
	if got := mtBaseline.TrackMatteLayer(); got == nil || got.ID != solidA.ID {
		t.Errorf("after Set: TrackMatteLayer() = %v, want solidA", got)
	}

	// Roundtrip.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var comp2 *aep.Composition
	for _, c := range proj2.Compositions {
		if c.Name == "RE_TRACKMATTE" {
			comp2 = c
			break
		}
	}
	mt2 := layerBySourceName(proj2, comp2, "mt_baseline")
	solidA2 := layerBySourceName(proj2, comp2, "solidA")
	if mt2 == nil || solidA2 == nil {
		t.Fatalf("roundtrip layers missing")
	}
	if mt2.TrackMatteLayerID != solidA2.ID {
		t.Errorf("roundtrip TMLayerID = %d, want %d", mt2.TrackMatteLayerID, solidA2.ID)
	}
	if mt2.TrackMatte != aep.TrackMatteLuma {
		t.Errorf("roundtrip TrackMatte = %v, want Luma", mt2.TrackMatte)
	}

	// Clear and roundtrip.
	if err := mt2.ClearTrackMatteLayer(); err != nil {
		t.Fatalf("ClearTrackMatteLayer: %v", err)
	}
	if mt2.TrackMatteLayerID != 0 || mt2.TrackMatte != aep.TrackMatteNone {
		t.Errorf("after Clear: TMLayerID=%d TrackMatte=%v", mt2.TrackMatteLayerID, mt2.TrackMatte)
	}
	var buf2 bytes.Buffer
	if err := proj2.WriteAEP(&buf2); err != nil {
		t.Fatalf("WriteAEP after clear: %v", err)
	}
	proj3, err := aep.FromReader(bytes.NewReader(buf2.Bytes()))
	if err != nil {
		t.Fatalf("re-parse after clear: %v", err)
	}
	var comp3 *aep.Composition
	for _, c := range proj3.Compositions {
		if c.Name == "RE_TRACKMATTE" {
			comp3 = c
			break
		}
	}
	mt3 := layerBySourceName(proj3, comp3, "mt_baseline")
	if mt3.TrackMatteLayerID != 0 || mt3.TrackMatte != aep.TrackMatteNone {
		t.Errorf("after clear+roundtrip: TMLayerID=%d TrackMatte=%v", mt3.TrackMatteLayerID, mt3.TrackMatte)
	}
}

func TestSetTrackMatteLayerRejectsInvalid(t *testing.T) {
	proj, comp := openTrackMatteAE24(t)
	mtBaseline := layerBySourceName(proj, comp, "mt_baseline")
	if mtBaseline == nil {
		t.Fatalf("mt_baseline missing")
	}
	// Self-matte rejected.
	if err := mtBaseline.SetTrackMatteLayer(mtBaseline.ID, aep.TrackMatteAlpha); err == nil {
		t.Error("self-matte: expected error")
	}
	// Unknown ID rejected.
	if err := mtBaseline.SetTrackMatteLayer(99999, aep.TrackMatteAlpha); err == nil {
		t.Error("unknown ID: expected error")
	}
	// Non-text layer without ldta returns error.
	stub := &aep.Layer{Name: "stub"}
	if err := stub.SetTrackMatteLayer(1, aep.TrackMatteAlpha); err == nil {
		t.Error("layer without ldta: expected error")
	}
}

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

// TestGeometryOptionsTypedSettersRoundtrip exercises Layer.Geometry*
// typed getters/setters against re_geometry_options.aep. AE 24+
// Advanced 3D renderer enabled in fixture; Plane Curvature / Plane
// Subdivision / Bevel Direction present.
func TestGeometryOptionsTypedSettersRoundtrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_geometry_options.aep")
	if err != nil {
		t.Skipf("re_geometry_options.aep not present")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "GEO_OPT" {
			comp = c
		}
	}
	if comp == nil {
		t.Fatal("GEO_OPT comp missing")
	}
	var solid *aep.Layer
	for _, l := range comp.Layers {
		if l.GeometryPlaneCurvature() != nil {
			solid = l
			break
		}
	}
	if solid == nil {
		t.Fatal("3D solid with geometry options not found")
	}
	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Errorf("%s: %v", label, err)
		}
	}
	mustNoErr("SetGeometryPlaneCurvature", solid.SetGeometryPlaneCurvature(0.4))
	mustNoErr("SetGeometryPlaneSubdivision", solid.SetGeometryPlaneSubdivision(8))
	mustNoErr("SetGeometryBevelDirection", solid.SetGeometryBevelDirection(2))

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var reSolid *aep.Layer
	for _, c := range re.Compositions {
		for _, l := range c.Layers {
			if p := l.GeometryPlaneCurvature(); p != nil {
				if v, ok := p.StaticValue.(float64); ok && math.Abs(v-0.4) < 1e-3 {
					reSolid = l
				}
			}
		}
	}
	if reSolid == nil {
		t.Fatal("post-roundtrip solid missing")
	}
	if v := reSolid.GeometryPlaneSubdivision().StaticValue.(float64); v != 8 {
		t.Errorf("post-roundtrip Subdivision = %g, want 8", v)
	}
	if v := reSolid.GeometryBevelDirection().StaticValue.(float64); v != 2 {
		t.Errorf("post-roundtrip BevelDirection = %g, want 2", v)
	}
}

// TestMaterialOptionsTypedSettersRoundtrip exercises Layer.Material* typed
// getters/setters against re_material_options.aep, which has 3D-enabled
// solids carrying the full 17-property materialOption tree.
func TestMaterialOptionsTypedSettersRoundtrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_material_options.aep")
	if err != nil {
		t.Skipf("re_material_options.aep not present")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "MAT_OPT" {
			comp = c
		}
	}
	if comp == nil {
		t.Fatal("MAT_OPT comp missing")
	}
	// Find the 3D solid with custom material (has Specular = 0.6).
	var solid *aep.Layer
	for _, l := range comp.Layers {
		if sp := l.MaterialSpecular(); sp != nil {
			if v, ok := sp.StaticValue.(float64); ok && v > 0.5 && v < 0.7 {
				solid = l
			}
		}
	}
	if solid == nil {
		t.Fatal("3D solid with custom Specular not found")
	}

	// Verify getters surface fixture values.
	if v := solid.MaterialCastsShadows().StaticValue.(float64); v != 2 {
		t.Errorf("CastsShadows = %g, want 2 (Only)", v)
	}
	if v := solid.MaterialAmbient().StaticValue.(float64); math.Abs(v-0.5) > 1e-3 {
		t.Errorf("Ambient = %g, want ~0.5", v)
	}

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Errorf("%s: %v", label, err)
		}
	}
	// Exercise material setters. NOTE: AE prunes properties from
	// serialization when they sit at default — AcceptsShadows on this
	// custom layer is missing because the JSX wrote it = 1 (default).
	// Shininess Coefficient is also pruned on the custom layer (JSX
	// didn't touch it). Find the sister default-valued layer to test
	// those two.
	mustNoErr("SetMaterialCastsShadows", solid.SetMaterialCastsShadows(aep.MaterialCastsOn))
	mustNoErr("SetMaterialLightTransmission", solid.SetMaterialLightTransmission(0.3))
	mustNoErr("SetMaterialAcceptsLights", solid.SetMaterialAcceptsLights(true))
	mustNoErr("SetMaterialAppearsInReflections", solid.SetMaterialAppearsInReflections(false))
	mustNoErr("SetMaterialAmbient", solid.SetMaterialAmbient(0.75))
	mustNoErr("SetMaterialDiffuse", solid.SetMaterialDiffuse(0.65))
	mustNoErr("SetMaterialSpecular", solid.SetMaterialSpecular(0.55))
	mustNoErr("SetMaterialMetal", solid.SetMaterialMetal(0.1))
	mustNoErr("SetMaterialReflection", solid.SetMaterialReflection(0.4))
	mustNoErr("SetMaterialFresnel", solid.SetMaterialFresnel(0.2))
	mustNoErr("SetMaterialTransparency", solid.SetMaterialTransparency(0.0))
	mustNoErr("SetMaterialTranspRolloff", solid.SetMaterialTranspRolloff(0.0))
	mustNoErr("SetMaterialIndexOfRefraction", solid.SetMaterialIndexOfRefraction(1.5))
	mustNoErr("SetMaterialShadowColor", solid.SetMaterialShadowColor([]float64{0.5, 0, 0, 1}))
	if g := solid.MaterialGlossiness(); g == nil {
		t.Error("Glossiness getter returned nil")
	}

	// Find the default-valued 3D solid that still carries AcceptsShadows
	// and Shininess (since they're at default, weren't pruned).
	var defSolid *aep.Layer
	for _, l := range comp.Layers {
		if l == solid {
			continue
		}
		if l.MaterialAcceptsShadows() != nil && l.MaterialShininess() != nil {
			defSolid = l
		}
	}
	if defSolid != nil {
		mustNoErr("SetMaterialAcceptsShadows", defSolid.SetMaterialAcceptsShadows(false))
		mustNoErr("SetMaterialShininess", defSolid.SetMaterialShininess(45))
	} else {
		t.Log("default-valued 3D solid not found; skipping AcceptsShadows/Shininess setter test")
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
	var reSolid *aep.Layer
	for _, c := range re.Compositions {
		for _, l := range c.Layers {
			if sp := l.MaterialSpecular(); sp != nil {
				if v, ok := sp.StaticValue.(float64); ok && math.Abs(v-0.55) < 1e-3 {
					reSolid = l
				}
			}
		}
	}
	if reSolid == nil {
		t.Fatal("post-roundtrip solid missing")
	}
	if v := reSolid.MaterialCastsShadows().StaticValue.(float64); v != 1 {
		t.Errorf("post-roundtrip CastsShadows = %g, want 1", v)
	}
	if v := reSolid.MaterialIndexOfRefraction().StaticValue.(float64); math.Abs(v-1.5) > 1e-3 {
		t.Errorf("post-roundtrip IOR = %g, want 1.5", v)
	}

	// Negative: 2D layer (no materialOption) should error.
	var solid2D *aep.Layer
	for _, l := range comp.Layers {
		if l.MaterialSpecular() == nil {
			solid2D = l
			break
		}
	}
	if solid2D != nil {
		if err := solid2D.SetMaterialSpecular(0.5); err == nil {
			t.Error("SetMaterialSpecular on 2D layer should error")
		}
	}
}

// TestTransformTypedSettersRoundtrip exercises Layer transform-group
// typed setters against re_cameralight.aep. MyCamera is 3D-enabled in
// the fixture, so RotateX/Y/Orientation are present too.
func TestTransformTypedSettersRoundtrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_CL" {
			comp = c
		}
	}
	if comp == nil {
		t.Fatal("RE_CL comp missing")
	}
	var cam *aep.Layer
	for _, l := range comp.Layers {
		if l.Name == "MyCamera" {
			cam = l
		}
	}
	if cam == nil {
		t.Fatal("MyCamera missing")
	}

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Errorf("%s: %v", label, err)
		}
	}
	mustNoErr("SetAnchorPoint", cam.SetAnchorPoint([]float64{100, 200, 300}))
	mustNoErr("SetPosition", cam.SetPosition([]float64{500, 600, -700}))
	mustNoErr("SetScale", cam.SetScale([]float64{2, 2, 2}))
	mustNoErr("SetRotation", cam.SetRotation(45))
	mustNoErr("SetOpacity", cam.SetOpacity(0.5))
	// Camera in this fixture has RotateZ-only (no Is3D rotation axes in the
	// property dump); RotateX / RotateY / Orientation may or may not be
	// present. Probe and skip if absent (typed setter surfaces this).
	if err := cam.SetRotateX(30); err != nil {
		t.Logf("SetRotateX: %v (expected if not 3D)", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var reCam *aep.Layer
	for _, c := range re.Compositions {
		for _, l := range c.Layers {
			if l.Name == "MyCamera" {
				reCam = l
			}
		}
	}
	if reCam == nil {
		t.Fatal("post-roundtrip MyCamera missing")
	}
	got, _ := reCam.Position().StaticValue.([]float64)
	want := []float64{500, 600, -700}
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("post-roundtrip Position = %v, want %v", got, want)
	}
	if v, _ := reCam.Rotation().StaticValue.(float64); v != 45 {
		t.Errorf("post-roundtrip Rotation = %g, want 45", v)
	}
	if v, _ := reCam.Opacity().StaticValue.(float64); v != 0.5 {
		t.Errorf("post-roundtrip Opacity = %g, want 0.5", v)
	}
	gotS, _ := reCam.Scale().StaticValue.([]float64)
	if len(gotS) != 3 || gotS[0] != 2 {
		t.Errorf("post-roundtrip Scale = %v, want [2 2 2]", gotS)
	}
}

func TestAlternateSourceReal(t *testing.T) {
	proj, comp := openAlternateSourceAE24(t)

	// AE 25 wraps any AVItem passed to setAlternateSource() in an
	// auto-created precomp named "<slotName>_<originalName> 2", filed
	// under a "媒体替换合成" / "Media Replacement Comps" folder. The
	// persisted blsi points at the wrapper, not at the AVItem the script
	// passed. Our fixture set the alt source to ALT_SRC_B; AE stored the
	// wrapper comp "MediaSlot_ALT_SRC_B 2" instead.
	wrapper := proj.CompositionByName("MediaSlot_ALT_SRC_B 2")
	if wrapper == nil {
		t.Fatalf("expected AE-auto-created wrapper comp 'MediaSlot_ALT_SRC_B 2' missing — fixture or AE version differs?")
	}

	withAlt := layerByName(comp, "with_alt_b")
	baseline := layerByName(comp, "baseline_no_alt")
	if withAlt == nil || baseline == nil {
		t.Fatalf("fixture layers missing (withAlt=%v baseline=%v)", withAlt, baseline)
	}

	if !withAlt.HasAlternateSourceSlot() {
		t.Error("with_alt_b: HasAlternateSourceSlot() = false, want true")
	}
	if !baseline.HasAlternateSourceSlot() {
		t.Error("baseline_no_alt: HasAlternateSourceSlot() = false, want true (EGP slot is persisted even without an override)")
	}

	if withAlt.AlternateSourceID != wrapper.ID {
		t.Errorf("with_alt_b AlternateSourceID = %d, want %d (wrapper)", withAlt.AlternateSourceID, wrapper.ID)
	}
	got := withAlt.AlternateSource()
	if got == nil {
		t.Fatalf("with_alt_b AlternateSource() = nil, want wrapper comp")
	}
	if got.ItemID() != wrapper.ID || got.ItemName() != wrapper.Name {
		t.Errorf("with_alt_b AlternateSource() = {id=%d name=%q}, want {id=%d name=%q}",
			got.ItemID(), got.ItemName(), wrapper.ID, wrapper.Name)
	}

	if baseline.AlternateSourceID != 0 {
		t.Errorf("baseline_no_alt AlternateSourceID = %d, want 0", baseline.AlternateSourceID)
	}
	if baseline.AlternateSource() != nil {
		t.Errorf("baseline_no_alt AlternateSource() = %v, want nil", baseline.AlternateSource())
	}
}

func TestSetAlternateSourceRoundtrip(t *testing.T) {
	proj, comp := openAlternateSourceAE24(t)
	srcA := proj.CompositionByName("ALT_SRC_A")
	srcB := proj.CompositionByName("ALT_SRC_B")
	baseline := layerByName(comp, "baseline_no_alt")
	if srcA == nil || srcB == nil || baseline == nil {
		t.Fatalf("fixture missing (srcA=%v srcB=%v baseline=%v)", srcA, srcB, baseline)
	}
	if baseline.AlternateSourceID != 0 {
		t.Fatalf("baseline starts with override id=%d, want 0", baseline.AlternateSourceID)
	}

	if err := baseline.SetAlternateSource(srcA); err != nil {
		t.Fatalf("SetAlternateSource(srcA): %v", err)
	}
	if baseline.AlternateSourceID != srcA.ID {
		t.Errorf("after Set(srcA): AlternateSourceID = %d, want %d", baseline.AlternateSourceID, srcA.ID)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	srcA2 := proj2.CompositionByName("ALT_SRC_A")
	srcB2 := proj2.CompositionByName("ALT_SRC_B")
	comp2 := proj2.CompositionByName("RE_ALT_SOURCE_MAIN")
	baseline2 := layerByName(comp2, "baseline_no_alt")
	if baseline2 == nil {
		t.Fatalf("baseline_no_alt missing after roundtrip")
	}
	if baseline2.AlternateSourceID != srcA2.ID {
		t.Errorf("roundtrip AlternateSourceID = %d, want %d", baseline2.AlternateSourceID, srcA2.ID)
	}
	if got := baseline2.AlternateSource(); got == nil || got.ItemID() != srcA2.ID {
		t.Errorf("roundtrip AlternateSource() = %v, want srcA", got)
	}

	// Switch to srcB and roundtrip again.
	if err := baseline2.SetAlternateSource(srcB2); err != nil {
		t.Fatalf("SetAlternateSource(srcB): %v", err)
	}
	var buf2 bytes.Buffer
	if err := proj2.WriteAEP(&buf2); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj3, err := aep.FromReader(bytes.NewReader(buf2.Bytes()))
	if err != nil {
		t.Fatalf("re-parse after switch: %v", err)
	}
	comp3 := proj3.CompositionByName("RE_ALT_SOURCE_MAIN")
	srcB3 := proj3.CompositionByName("ALT_SRC_B")
	baseline3 := layerByName(comp3, "baseline_no_alt")
	if baseline3.AlternateSourceID != srcB3.ID {
		t.Errorf("after switch+roundtrip: AlternateSourceID = %d, want %d", baseline3.AlternateSourceID, srcB3.ID)
	}

	// Clear and roundtrip.
	if err := baseline3.ClearAlternateSource(); err != nil {
		t.Fatalf("ClearAlternateSource: %v", err)
	}
	if baseline3.AlternateSourceID != 0 {
		t.Errorf("after Clear: AlternateSourceID = %d, want 0", baseline3.AlternateSourceID)
	}
	if baseline3.AlternateSource() != nil {
		t.Errorf("after Clear: AlternateSource() != nil")
	}
	var buf3 bytes.Buffer
	if err := proj3.WriteAEP(&buf3); err != nil {
		t.Fatalf("WriteAEP after clear: %v", err)
	}
	proj4, err := aep.FromReader(bytes.NewReader(buf3.Bytes()))
	if err != nil {
		t.Fatalf("re-parse after clear: %v", err)
	}
	comp4 := proj4.CompositionByName("RE_ALT_SOURCE_MAIN")
	baseline4 := layerByName(comp4, "baseline_no_alt")
	if baseline4.AlternateSourceID != 0 {
		t.Errorf("after clear+roundtrip: AlternateSourceID = %d, want 0", baseline4.AlternateSourceID)
	}
}

func TestSetAlternateSourceRejectsInvalid(t *testing.T) {
	proj, comp := openAlternateSourceAE24(t)
	baseline := layerByName(comp, "baseline_no_alt")
	if baseline == nil {
		t.Fatalf("baseline_no_alt missing")
	}

	// Item id not in project.
	bogus := &aep.Composition{ID: 999999, Name: "BOGUS"}
	if err := baseline.SetAlternateSource(bogus); err == nil {
		t.Error("bogus item id: expected error")
	}

	// Layer without an EGP slot: pick any layer inside ALT_SRC_A (the
	// "innerA" solid that has no addToMotionGraphics applied).
	srcAComp := proj.CompositionByName("ALT_SRC_A")
	if srcAComp == nil || len(srcAComp.Layers) == 0 {
		t.Skip("ALT_SRC_A has no layers; can't test no-slot rejection")
	}
	inner := srcAComp.Layers[0]
	if inner.HasAlternateSourceSlot() {
		t.Skip("unexpected: innerA has EGP slot")
	}
	srcB := proj.CompositionByName("ALT_SRC_B")
	if err := inner.SetAlternateSource(srcB); err == nil {
		t.Error("no-slot SetAlternateSource: expected error")
	}
}

// TestLayerAddFontAndUse exercises Layer.AddFont — append a new font
// to the btdk Fonts table, then point an existing style run at the
// new index via SetRunFontIndex. Round-trip via WriteAEP to confirm
// the new font appears in re-parsed TextSource.Fonts.
func TestLayerAddFontAndUse(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_text.aep")
	if err != nil {
		t.Skipf("re_text.aep not present")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_TEXT" {
			if comp == nil || len(c.Layers) > len(comp.Layers) {
				comp = c
			}
		}
	}
	if comp == nil {
		t.Fatal("RE_TEXT comp missing")
	}
	var layer *aep.Layer
	for _, l := range comp.Layers {
		if l.Name == "baseline_A" {
			layer = l
			break
		}
	}
	if layer == nil || layer.TextSource == nil {
		t.Fatal("baseline_A text layer not found")
	}
	beforeCount := len(layer.TextSource.Fonts)

	newIdx, err := layer.AddFont("Arial-BoldMT")
	if err != nil {
		t.Fatalf("AddFont: %v", err)
	}
	if newIdx != beforeCount {
		t.Errorf("AddFont returned %d, want %d (= old len)", newIdx, beforeCount)
	}
	if len(layer.TextSource.Fonts) != beforeCount+1 {
		t.Errorf("Fonts len = %d, want %d", len(layer.TextSource.Fonts), beforeCount+1)
	}
	if layer.TextSource.Fonts[newIdx] != "Arial-BoldMT" {
		t.Errorf("Fonts[%d] = %q, want %q", newIdx, layer.TextSource.Fonts[newIdx], "Arial-BoldMT")
	}

	// Point run #0 at the new font.
	if err := layer.SetRunFontIndex(0, newIdx); err != nil {
		t.Fatalf("SetRunFontIndex: %v", err)
	}
	if layer.TextSource.Runs[0].FontIndex != newIdx || layer.TextSource.Runs[0].FontName != "Arial-BoldMT" {
		t.Errorf("after SetRunFontIndex: idx=%d name=%q", layer.TextSource.Runs[0].FontIndex, layer.TextSource.Runs[0].FontName)
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
	var layer2 *aep.Layer
	for _, c := range proj2.Compositions {
		if c.Name != "RE_TEXT" {
			continue
		}
		for _, l := range c.Layers {
			if l.Name == "baseline_A" {
				layer2 = l
				break
			}
		}
	}
	if layer2 == nil || layer2.TextSource == nil {
		t.Fatal("after roundtrip: baseline_A missing or undecoded")
	}
	if len(layer2.TextSource.Fonts) != beforeCount+1 {
		t.Errorf("roundtrip Fonts len = %d, want %d", len(layer2.TextSource.Fonts), beforeCount+1)
	}
	if layer2.TextSource.Fonts[newIdx] != "Arial-BoldMT" {
		t.Errorf("roundtrip Fonts[%d] = %q", newIdx, layer2.TextSource.Fonts[newIdx])
	}
	if layer2.TextSource.Runs[0].FontName != "Arial-BoldMT" {
		t.Errorf("roundtrip run[0].FontName = %q", layer2.TextSource.Runs[0].FontName)
	}
}
