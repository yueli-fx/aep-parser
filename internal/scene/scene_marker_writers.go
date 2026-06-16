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
//
//aep:cap domain=comp tier=stable verify=roundtrip boundary="length-preserving(ldat tick write);无专门 AE gate→round-trip" alias="marker time,标记时间,cue time,时间标记"
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
//
//aep:cap domain=comp tier=stable verify=roundtrip boundary="length-preserving(4B NmHd);0 → point marker;负值夹至 0;无专门 AE gate→round-trip" alias="marker duration,标记时长,cue duration,区间标记"
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
//
//aep:cap domain=comp tier=stable verify=roundtrip boundary="length-preserving(1B NmHd @0x10);越界值写入并 round-trip;无专门 AE gate→round-trip" alias="marker label,标记颜色,label color,时间线颜色"
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
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestMarker_AEShipGate_AE2020,TestMarker_AEShipGate_AE2025 boundary="length-variable(Utf8 chunk 替换+父 LIST size 重算);AE gate 在 AddMarker 后调用" alias="marker comment,标记注释,cue point comment,标记文本"
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
//
//aep:cap domain=comp tier=stable verify=roundtrip boundary="length-variable(Utf8 chunk 替换+父 LIST size 重算);无专门 AE gate→round-trip" alias="marker chapter,章节链接,chapter link"
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
//
//aep:cap domain=comp tier=stable verify=roundtrip boundary="length-variable(Utf8 chunk 替换+父 LIST size 重算);无专门 AE gate→round-trip" alias="marker url,web link,网址,超链接"
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
//
//aep:cap domain=comp tier=stable verify=roundtrip boundary="length-variable(Utf8 chunk 替换+父 LIST size 重算);无专门 AE gate→round-trip" alias="marker frame target,frame target,帧目标"
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
//
//aep:cap domain=comp tier=stable verify=roundtrip boundary="length-variable(Utf8 chunk 替换+父 LIST size 重算);Flash 遗留字段,现代 AE 项目极少填;无专门 AE gate→round-trip" alias="cue point name,提示点名称,flash cue"
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
