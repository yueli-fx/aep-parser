package aep_test

import (
	"bytes"
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
