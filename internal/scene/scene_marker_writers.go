package scene

import (
	"fmt"
	"math"
)

// Marker write API. Layer markers and composition markers share the
// same Marker type and same writers — both reach here through the
// underlying ldat / NmHd / Utf8 chunk references captured at parse time.
//
// Time / Duration / Label are length-preserving (fixed-offset byte
// writes). Comment / Chapter / URL / FrameTarget / CuePointName are
// length-variable — the underlying Utf8 chunk's bytes can grow or
// shrink; WriteAEP recomputes all parent LIST sizes.

// SetTime writes a new marker time (seconds) to the ldat keyframe slot
// using the owning composition's TickRate. length-preserving.
func (m *Marker) SetTime(seconds float64) error {
	if m.back == nil {
		return fmt.Errorf("marker: no ldat reference (built outside parser?)")
	}
	if err := m.back.SetTime(seconds); err != nil {
		return err
	}
	rate := m.back.CompTickRate()
	if rate == 0 {
		rate = aeLegacyTimeBase
	}
	ticks := uint32(math.Round(seconds * rate))
	m.Time = float64(ticks) / rate
	return nil
}

// SetDuration writes a new marker duration (seconds) to NmHd @0x08.
// Encoded as uint32 BE in 600ths-of-a-second per py-aep documentation;
// see decodeNmHd doc comment for the offset rationale.
// length-preserving (4 bytes).
//
// `seconds == 0` produces a point marker. Negative durations clamp to 0.
func (m *Marker) SetDuration(seconds float64) error {
	if m.back == nil {
		return fmt.Errorf("marker: no NmHd reference (built outside parser?)")
	}
	if seconds < 0 {
		seconds = 0
	}
	if err := m.back.SetDuration(seconds); err != nil {
		return err
	}
	const nmHdDurationBase = 600.0
	ticks := uint32(math.Round(seconds * nmHdDurationBase))
	m.Duration = float64(ticks) / nmHdDurationBase
	return nil
}

// SetLabel writes a new timeline label-color index (0..16) to NmHd
// @0x10. Indices outside 0..16 are written verbatim (AE shows index 0
// for unknown values but the byte round-trips).
// length-preserving (1 byte).
func (m *Marker) SetLabel(index uint8) error {
	if m.back == nil {
		return fmt.Errorf("marker: no NmHd reference (built outside parser?)")
	}
	if err := m.back.SetLabel(index); err != nil {
		return err
	}
	m.Label = index
	return nil
}

// SetComment rewrites the marker's primary comment text (first Utf8
// child of the Nmrd block). length-variable — the Utf8 chunk's data
// slice is replaced wholesale; WriteAEP recomputes ancestor LIST sizes.
//
// Use SetChapter / SetURL / SetFrameTarget / SetCuePointName for the
// other four Utf8 slots (they fill in declaration order, so missing
// earlier slots are created as empty when a later slot is written).
func (m *Marker) SetComment(s string) error {
	if m.back == nil {
		return fmt.Errorf("marker: no Nmrd reference (built outside parser?)")
	}
	if err := m.back.SetComment(s); err != nil {
		return err
	}
	m.Comment = s
	return nil
}

// SetChapter rewrites the marker's chapter-link text (second Utf8 in
// the Nmrd block).
func (m *Marker) SetChapter(s string) error {
	if m.back == nil {
		return fmt.Errorf("marker: no Nmrd reference (built outside parser?)")
	}
	if err := m.back.SetChapter(s); err != nil {
		return err
	}
	m.Chapter = s
	return nil
}

// SetURL rewrites the marker's web-target URL (third Utf8).
func (m *Marker) SetURL(s string) error {
	if m.back == nil {
		return fmt.Errorf("marker: no Nmrd reference (built outside parser?)")
	}
	if err := m.back.SetURL(s); err != nil {
		return err
	}
	m.URL = s
	return nil
}

// SetFrameTarget rewrites the marker's frame-target id (fourth Utf8).
func (m *Marker) SetFrameTarget(s string) error {
	if m.back == nil {
		return fmt.Errorf("marker: no Nmrd reference (built outside parser?)")
	}
	if err := m.back.SetFrameTarget(s); err != nil {
		return err
	}
	m.FrameTarget = s
	return nil
}

// SetCuePointName rewrites the marker's cue-point name (fifth Utf8 —
// legacy Flash-era; rarely populated in modern AE projects).
func (m *Marker) SetCuePointName(s string) error {
	if m.back == nil {
		return fmt.Errorf("marker: no Nmrd reference (built outside parser?)")
	}
	if err := m.back.SetCuePointName(s); err != nil {
		return err
	}
	m.CuePointName = s
	return nil
}
