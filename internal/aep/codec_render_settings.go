package aep

import "encoding/binary"

// codec_render_settings.go — pure byte decoding of the render-queue settings
// ldat item. No scene coupling (codec_ boundary). Byte layout mirrors py-aep
// binary/render_chunks.py RenderSettingsItem (2246 bytes, big-endian); offsets
// verified by field-size accounting (sum == 2246) + golden cross-check.

const renderSettingsItemSize = 2246

// Field offsets within one RenderSettingsItem block (slice-1 subset).
const (
	rsCompID                = 0x08 // u32
	rsStatus                = 0x0C // u32
	rsTimeSpanStartDividend = 0x14 // u32
	rsTimeSpanStartDivisor  = 0x18 // u32
	rsTimeSpanDurDividend   = 0x1C // u32
	rsTimeSpanDurDivisor    = 0x20 // u32
	rsTemplateName          = 0x5A // 64 bytes, windows-1252, NUL-padded
	rsTemplateNameLen       = 64
	rsTimeSpanSource        = 2148 // u16
)

// time_span_source enum values (py-aep TimeSpanSource).
const (
	timeSpanLengthOfComp = 0
	timeSpanWorkAreaOnly = 1
	timeSpanCustom       = 2
)

// renderSettings holds the decoded slice-1 fields of one settings block.
type renderSettings struct {
	compID            uint32
	status            uint32
	templateName      string
	timeSpanSource    uint16
	tsStartDividend   uint32
	tsStartDivisor    uint32
	tsDurationDivdend uint32
	tsDurationDivisor uint32
}

// decodeRenderSettings decodes one 2246-byte settings block. Returns false if
// the block is too short to hold the fields we read.
func decodeRenderSettings(b []byte) (renderSettings, bool) {
	if len(b) < renderSettingsItemSize {
		return renderSettings{}, false
	}
	rs := renderSettings{
		compID:            binary.BigEndian.Uint32(b[rsCompID:]),
		status:            binary.BigEndian.Uint32(b[rsStatus:]),
		templateName:      decodeWin1252(b[rsTemplateName : rsTemplateName+rsTemplateNameLen]),
		timeSpanSource:    binary.BigEndian.Uint16(b[rsTimeSpanSource:]),
		tsStartDividend:   binary.BigEndian.Uint32(b[rsTimeSpanStartDividend:]),
		tsStartDivisor:    binary.BigEndian.Uint32(b[rsTimeSpanStartDivisor:]),
		tsDurationDivdend: binary.BigEndian.Uint32(b[rsTimeSpanDurDividend:]),
		tsDurationDivisor: binary.BigEndian.Uint32(b[rsTimeSpanDurDivisor:]),
	}
	return rs, true
}

// decodeWin1252 decodes a NUL-padded windows-1252 byte slice to a string.
// Bytes < 0x80 are ASCII; 0x80–0xFF are mapped to their Unicode code points
// (latin-1 superset, sufficient for render-settings template names). Trailing
// NULs are trimmed.
func decodeWin1252(b []byte) string {
	end := len(b)
	for end > 0 && b[end-1] == 0 {
		end--
	}
	b = b[:end]
	for _, c := range b {
		if c >= 0x80 {
			r := make([]rune, len(b))
			for i, x := range b {
				r[i] = rune(x)
			}
			return string(r)
		}
	}
	return string(b)
}
