package aep

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/rifx"
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
	if m.ldat == nil {
		return fmt.Errorf("marker: no ldat reference (built outside parser?)")
	}
	if seconds < 0 {
		return fmt.Errorf("marker: negative time %g not supported", seconds)
	}
	if m.ldatOffset+4 > len(m.ldat.Data) {
		return fmt.Errorf("marker: ldat offset %d out of range (len=%d)", m.ldatOffset, len(m.ldat.Data))
	}
	rate := m.tickRate
	if rate == 0 {
		rate = aeLegacyTimeBase
	}
	ticks := uint32(math.Round(seconds * rate))
	binary.BigEndian.PutUint32(m.ldat.Data[m.ldatOffset:m.ldatOffset+4], ticks)
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
	if m.nmHd == nil {
		return fmt.Errorf("marker: no NmHd reference (built outside parser?)")
	}
	if len(m.nmHd.Data) < 0x0C {
		return fmt.Errorf("marker: NmHd too short for Duration write (len=%d)", len(m.nmHd.Data))
	}
	if seconds < 0 {
		seconds = 0
	}
	const nmHdDurationBase = 600.0
	ticks := uint32(math.Round(seconds * nmHdDurationBase))
	binary.BigEndian.PutUint32(m.nmHd.Data[0x08:0x0C], ticks)
	m.Duration = float64(ticks) / nmHdDurationBase
	return nil
}

// SetLabel writes a new timeline label-color index (0..16) to NmHd
// @0x10. Indices outside 0..16 are written verbatim (AE shows index 0
// for unknown values but the byte round-trips).
// length-preserving (1 byte).
func (m *Marker) SetLabel(index uint8) error {
	if m.nmHd == nil {
		return fmt.Errorf("marker: no NmHd reference (built outside parser?)")
	}
	if len(m.nmHd.Data) < 0x11 {
		return fmt.Errorf("marker: NmHd too short for Label write (len=%d)", len(m.nmHd.Data))
	}
	m.nmHd.Data[0x10] = index
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
	return m.setNmrdUtf8(0, s, func() { m.Comment = s })
}

// SetChapter rewrites the marker's chapter-link text (second Utf8 in
// the Nmrd block).
func (m *Marker) SetChapter(s string) error {
	return m.setNmrdUtf8(1, s, func() { m.Chapter = s })
}

// SetURL rewrites the marker's web-target URL (third Utf8).
func (m *Marker) SetURL(s string) error {
	return m.setNmrdUtf8(2, s, func() { m.URL = s })
}

// SetFrameTarget rewrites the marker's frame-target id (fourth Utf8).
func (m *Marker) SetFrameTarget(s string) error {
	return m.setNmrdUtf8(3, s, func() { m.FrameTarget = s })
}

// SetCuePointName rewrites the marker's cue-point name (fifth Utf8 —
// legacy Flash-era; rarely populated in modern AE projects).
func (m *Marker) SetCuePointName(s string) error {
	return m.setNmrdUtf8(4, s, func() { m.CuePointName = s })
}

// setNmrdUtf8 mutates the n-th Utf8 child of the Marker's Nmrd block.
// If fewer Utf8 children exist than n+1, missing slots are filled with
// empty Utf8 chunks so the requested slot reaches the right index in
// declaration order (AE writes all 5 slots even when empty).
func (m *Marker) setNmrdUtf8(slot int, s string, onSuccess func()) error {
	if m.nmrd == nil {
		return fmt.Errorf("marker: no Nmrd reference (built outside parser?)")
	}
	// Locate existing Utf8 children in order — collect both their indices
	// in m.nmrd.Children and the count so we know whether to append.
	var utf8Idx []int
	for i, ch := range m.nmrd.Children {
		if ch.ID == rifx.IDUtf8 {
			utf8Idx = append(utf8Idx, i)
		}
	}
	for len(utf8Idx) <= slot {
		// Append an empty Utf8 chunk. WriteAEP will allocate the chunk
		// header and recompute sizes; we just need the Children entry.
		empty := &rifx.Chunk{ID: rifx.IDUtf8, Data: nil}
		m.nmrd.Children = append(m.nmrd.Children, empty)
		utf8Idx = append(utf8Idx, len(m.nmrd.Children)-1)
	}
	target := m.nmrd.Children[utf8Idx[slot]]
	target.Data = []byte(s)
	onSuccess()
	return nil
}
