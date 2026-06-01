package aep

import "encoding/binary"

// codec_render_settings.go — pure byte decoding of the render-queue settings
// ldat item. No scene coupling (codec_ boundary). Byte layout mirrors py-aep
// binary/render_chunks.py RenderSettingsItem (2246 bytes, big-endian); offsets
// verified by field-size accounting (sum == 2246) + golden cross-check.

const renderSettingsItemSize = 2246

// Field offsets within one RenderSettingsItem block. slice-1 (top-level) +
// slice-2 (render settings). All big-endian; verified by py-aep field-size
// accounting (sum == 2246) + golden cross-check.
const (
	rsFlagByte              = 0x07 // u8; bit 2 = queue_item_notify
	rsCompID                = 0x08 // u32
	rsStatus                = 0x0C // u32
	rsTimeSpanStartDividend = 0x14 // u32
	rsTimeSpanStartDivisor  = 0x18 // u32
	rsTimeSpanDurDividend   = 0x1C // u32
	rsTimeSpanDurDivisor    = 0x20 // u32
	rsFieldRender           = 0x32 // u16
	rsPulldown              = 0x36 // u16
	rsQuality               = 0x38 // u16 (0xFFFF = current)
	rsResolutionX           = 0x3A // u16
	rsResolutionY           = 0x3C // u16
	rsEffects               = 0x40 // u16 (0xFFFF = current)
	rsProxyUse              = 0x44 // u16 (0xFFFF = current)
	rsMotionBlur            = 0x48 // u16 (0xFFFF = current)
	rsFrameBlending         = 0x4C // u16 (0xFFFF = current)
	rsLogType               = 0x50 // u16 (raw; py-aep maps to 3xxx enum)
	rsSkipExistingFiles     = 0x54 // u16 (bool)
	rsTemplateName          = 0x5A // 64 bytes, windows-1252, NUL-padded
	rsTemplateNameLen       = 64
	rsUseThisFrameRate      = 2144 // u16 (FrameRateSetting: 0=comp, 1=this)
	rsTimeSpanSource        = 2148 // u16
	rsSoloSwitches          = 2164 // u16 (0xFFFF = current)
	rsDiskCache             = 2168 // u16 (0xFFFF = current)
	rsGuideLayers           = 2172 // u16 (0xFFFF = current)
	rsColorDepth            = 2180 // u16 (0xFFFF = current)
	rsElapsedSeconds        = 2202 // u32
)

// time_span_source enum values (py-aep TimeSpanSource).
const (
	timeSpanLengthOfComp = 0
	timeSpanWorkAreaOnly = 1
	timeSpanCustom       = 2
)

// renderSettings holds the decoded fields of one settings block.
type renderSettings struct {
	compID            uint32
	status            uint32
	templateName      string
	timeSpanSource    uint16
	tsStartDividend   uint32
	tsStartDivisor    uint32
	tsDurationDivdend uint32
	tsDurationDivisor uint32

	// slice-2 render settings (raw u16 unless noted).
	queueItemNotify   bool
	logType           uint16
	elapsedSeconds    uint32
	fieldRender       uint16
	pulldown          uint16
	quality           uint16
	resolutionX       uint16
	resolutionY       uint16
	effects           uint16
	proxyUse          uint16
	motionBlur        uint16
	frameBlending     uint16
	skipExistingFiles bool
	useThisFrameRate  uint16
	soloSwitches      uint16
	diskCache         uint16
	guideLayers       uint16
	colorDepth        uint16
}

// decodeRenderSettings decodes one 2246-byte settings block. Returns false if
// the block is too short to hold the fields we read.
func decodeRenderSettings(b []byte) (renderSettings, bool) {
	if len(b) < renderSettingsItemSize {
		return renderSettings{}, false
	}
	u16 := func(off int) uint16 { return binary.BigEndian.Uint16(b[off:]) }
	rs := renderSettings{
		compID:            binary.BigEndian.Uint32(b[rsCompID:]),
		status:            binary.BigEndian.Uint32(b[rsStatus:]),
		templateName:      decodeWin1252(b[rsTemplateName : rsTemplateName+rsTemplateNameLen]),
		timeSpanSource:    u16(rsTimeSpanSource),
		tsStartDividend:   binary.BigEndian.Uint32(b[rsTimeSpanStartDividend:]),
		tsStartDivisor:    binary.BigEndian.Uint32(b[rsTimeSpanStartDivisor:]),
		tsDurationDivdend: binary.BigEndian.Uint32(b[rsTimeSpanDurDividend:]),
		tsDurationDivisor: binary.BigEndian.Uint32(b[rsTimeSpanDurDivisor:]),
		queueItemNotify:   b[rsFlagByte]&(1<<2) != 0,
		logType:           u16(rsLogType),
		elapsedSeconds:    binary.BigEndian.Uint32(b[rsElapsedSeconds:]),
		fieldRender:       u16(rsFieldRender),
		pulldown:          u16(rsPulldown),
		quality:           u16(rsQuality),
		resolutionX:       u16(rsResolutionX),
		resolutionY:       u16(rsResolutionY),
		effects:           u16(rsEffects),
		proxyUse:          u16(rsProxyUse),
		motionBlur:        u16(rsMotionBlur),
		frameBlending:     u16(rsFrameBlending),
		skipExistingFiles: u16(rsSkipExistingFiles) != 0,
		useThisFrameRate:  u16(rsUseThisFrameRate),
		soloSwitches:      u16(rsSoloSwitches),
		diskCache:         u16(rsDiskCache),
		guideLayers:       u16(rsGuideLayers),
		colorDepth:        u16(rsColorDepth),
	}
	return rs, true
}

// currentSettingsInt maps a raw u16 render-setting value to py-aep's NUMBER
// semantics: 0xFFFF ("current settings") becomes -1, everything else is the
// value as-is.
func currentSettingsInt(v uint16) int {
	if v == 0xFFFF {
		return -1
	}
	return int(v)
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
