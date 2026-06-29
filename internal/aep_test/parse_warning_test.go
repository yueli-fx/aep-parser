package aep_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// TestParseEmitsWarningOnInconsistentKeyframeStream pins the regression for
// silent-failure paths in parseKeyframes. Two corrupt fixtures are parsed and
// each must surface a warning in Project.Warnings; the property itself must
// still exist (with no keyframes) so callers can detect the partial decode.
func TestParseEmitsWarningOnInconsistentKeyframeStream(t *testing.T) {
	// Case A: lhd3 header is shorter than 0x14 — header itself is malformed.
	t.Run("short_lhd3_header", func(t *testing.T) {
		rb := &rifxBuilder{}
		leaf := rb.buildCorruptKeyframedLeaf("ADBE Opacity", 0x01, make([]byte, 48),
			make([]byte, 0x10)) // only 16 bytes, need >= 20
		var tdgpBody []byte
		tdgpBody = append(tdgpBody, leaf...)
		tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
		data := wrapAsLayer(tdgpBody)

		proj, err := aep.FromReader(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("FromReader: %v", err)
		}
		if len(proj.Warnings) == 0 {
			t.Fatalf("expected a warning for short lhd3, got none")
		}
		matched := false
		for _, w := range proj.Warnings {
			if strings.Contains(w, "ADBE Opacity") && strings.Contains(w, "lhd3") {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("expected warning mentioning ADBE Opacity + lhd3, got: %v", proj.Warnings)
		}
		op := proj.Compositions[0].Layers[0].Opacity()
		if op == nil {
			t.Fatalf("Opacity property should still surface (just without keyframes)")
		}
		if len(op.Keyframes) != 0 {
			t.Errorf("Opacity should have 0 keyframes after corrupt parse, got %d", len(op.Keyframes))
		}
	})

	// Case B: lhd3 header is sane but claims more keyframes than ldat holds.
	t.Run("count_exceeds_ldat", func(t *testing.T) {
		rb := &rifxBuilder{}
		// Claim 99 keyframes × 48 bpk = 4752 bytes, but ldat only has 48.
		leaf := rb.buildCorruptKeyframedLeaf("ADBE Opacity", 0x01, make([]byte, 48),
			buildLhd3(99, 48))
		var tdgpBody []byte
		tdgpBody = append(tdgpBody, leaf...)
		tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
		data := wrapAsLayer(tdgpBody)

		proj, err := aep.FromReader(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("FromReader: %v", err)
		}
		matched := false
		for _, w := range proj.Warnings {
			if strings.Contains(w, "ADBE Opacity") && strings.Contains(w, "inconsistent") {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("expected warning mentioning ADBE Opacity + inconsistent, got: %v", proj.Warnings)
		}
		op := proj.Compositions[0].Layers[0].Opacity()
		if op == nil {
			t.Fatalf("Opacity property should still surface")
		}
		if len(op.Keyframes) != 0 {
			t.Errorf("Opacity should have 0 keyframes after corrupt parse, got %d", len(op.Keyframes))
		}
	})
}

// TestParseWarnsOnBpkLayoutMismatch pins the pre-flight bpk-vs-layout check.
// Constructs a 3D non-spatial property (Scale, header byte 0x00) with bpk
// truncated below its expected 128 to force the mismatch. decodeEasing's
// silent fallback used to swallow this; now parseKeyframes warns once upfront.
func TestParseWarnsOnBpkLayoutMismatch(t *testing.T) {
	rb := &rifxBuilder{}

	blk := make([]byte, 48)
	binary.BigEndian.PutUint32(blk[0:], uint32(math.Round(1.0*8000)))
	binary.BigEndian.PutUint64(blk[0x08:], math.Float64bits(5.0))

	leaf := rb.buildCorruptKeyframedLeaf("ADBE Scale", 0x03, blk, buildLhd3(1, 48))
	var tdgpBody []byte
	tdgpBody = append(tdgpBody, leaf...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	data := wrapAsLayer(tdgpBody)

	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}

	matched := false
	for _, w := range proj.Warnings {
		if strings.Contains(w, "ADBE Scale") && strings.Contains(w, "non-spatial 3D") {
			matched = true
			break
		}
	}
	if !matched {
		t.Errorf("expected warning about ADBE Scale bpk vs non-spatial 3D layout, got: %v", proj.Warnings)
	}

	scale := proj.Compositions[0].Layers[0].PropertyByMatchName("ADBE Scale")
	if scale == nil || len(scale.Keyframes) != 1 {
		t.Fatalf("Scale should still surface 1 partial keyframe, got %v", scale)
	}
	kf := scale.Keyframes[0]
	if got := kf.Time; math.Abs(got-1.0) > 1e-6 {
		t.Errorf("partial keyframe Time = %v, want 1.0", got)
	}
	if len(kf.InTemporalEase) != 0 {
		t.Errorf("partial keyframe should have empty InTemporalEase (decodeEasing skipped), got %v", kf.InTemporalEase)
	}
}

// TestParseCleanProjectHasNoWarnings pins the "clean parse stays quiet"
// contract: Warnings must be nil/empty when no anomalies are detected, so
// callers can use len(proj.Warnings) > 0 as a binary "anything weird?" signal.
func TestParseCleanProjectHasNoWarnings(t *testing.T) {
	data := buildKeyframedAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if len(proj.Warnings) != 0 {
		t.Errorf("clean parse produced unexpected warnings: %v", proj.Warnings)
	}
}
