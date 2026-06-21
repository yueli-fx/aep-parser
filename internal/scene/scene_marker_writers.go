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

// @summary     Set the marker's time
// @param       seconds  the new marker time in seconds
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestMarkerFields_AEShipGate_AE2020,TestMarkerFields_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving (ldat tick write); verified by reading the
//   marker's keyTime back through the AE DOM
// @alias       marker time,标记时间,cue time,时间标记
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

// @summary     Set the marker's duration
// @description Encoded as a uint32 BE count of 600ths-of-a-second at NmHd
//   offset 0x08. A duration of 0 produces a point marker; negative values
//   clamp to 0.
// @param       seconds  the new marker duration in seconds
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestMarkerFields_AEShipGate_AE2020,TestMarkerFields_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving (4 bytes); verified by reading the marker's
//   duration back through the AE DOM
// @alias       marker duration,标记时长,cue duration,区间标记
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

// @summary     Set the marker's timeline label color
// @description Indices outside the 0..16 label range are written verbatim —
//   AE displays index 0 for unrecognized values but the byte still
//   round-trips.
// @param       index  the new label-color index (0..16)
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestMarkerFields_AEShipGate_AE2020,TestMarkerFields_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving (1 byte at NmHd offset 0x10); verified by
//   reading the marker's label back through the AE DOM
// @alias       marker label,标记颜色,label color,时间线颜色
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

// @summary     Set the marker's primary comment text
// @description Rewrites the first Utf8 child of the Nmrd block. Use
//   SetChapter / SetURL / SetFrameTarget / SetCuePointName for the other
//   four Utf8 slots — they fill in declaration order, so writing a later
//   slot creates any missing earlier slots as empty.
// @param       s  the new comment text
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestMarker_AEShipGate_AE2020,TestMarker_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable — the Utf8 chunk's data is replaced wholesale
//   and ancestor LIST sizes are recomputed; gated by calling this after
//   AddMarker
// @alias       marker comment,标记注释,cue point comment,标记文本
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

// @summary     Set the marker's chapter-link text
// @description Rewrites the second Utf8 child of the Nmrd block.
// @param       s  the new chapter-link text
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestMarkerFields_AEShipGate_AE2020,TestMarkerFields_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable — the Utf8 chunk's data is replaced wholesale
//   and ancestor LIST sizes are recomputed; verified by reading the
//   marker's chapter back through the AE DOM
// @alias       marker chapter,章节链接,chapter link
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

// @summary     Set the marker's web-target URL
// @description Rewrites the third Utf8 child of the Nmrd block.
// @param       s  the new URL
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestMarkerFields_AEShipGate_AE2020,TestMarkerFields_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable — the Utf8 chunk's data is replaced wholesale
//   and ancestor LIST sizes are recomputed; verified by reading the
//   marker's URL back through the AE DOM
// @alias       marker url,web link,网址,超链接
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

// @summary     Set the marker's frame-target id
// @description Rewrites the fourth Utf8 child of the Nmrd block.
// @param       s  the new frame-target id
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestMarkerFields_AEShipGate_AE2020,TestMarkerFields_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable — the Utf8 chunk's data is replaced wholesale
//   and ancestor LIST sizes are recomputed; verified by reading the
//   marker's frame target back through the AE DOM
// @alias       marker frame target,frame target,帧目标
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

// @summary     Set the marker's cue-point name
// @description Rewrites the fifth Utf8 child of the Nmrd block — a
//   legacy Flash-era field, rarely populated in modern AE projects.
// @param       s  the new cue-point name
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestMarkerFields_AEShipGate_AE2020,TestMarkerFields_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable — the Utf8 chunk's data is replaced wholesale
//   and ancestor LIST sizes are recomputed; verified by reading the
//   marker's cue-point name back through the AE DOM
// @alias       cue point name,提示点名称,flash cue
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
