package codec

import "encoding/binary"

// codec_render_settings.go — pure byte decoding of the render-queue settings
// ldat item. No scene coupling (codec_ boundary). Byte layout mirrors the
// reference parser's binary/render_chunks.py RenderSettingsItem (2246 bytes,
// big-endian); offsets verified by field-size accounting (sum == 2246) + golden cross-check.

const RenderSettingsItemSize = 2246

// RenderSettingOffset is a typed byte offset into one of the render-queue
// fixed-width settings structures (the RenderSettingsItem block, the
// OutputModuleSettingsItem block, or the Roou chunk). Typing the offsets keeps
// the length-preserving patch helpers from taking a bare int and documents
// intent at the call boundary (§F D-U1).
type RenderSettingOffset int

// Field offsets within one RenderSettingsItem block. slice-1 (top-level) +
// slice-2 (render settings). All big-endian; verified by field-size
// accounting (sum == 2246) + golden cross-check.
const (
	RsFlagByte              RenderSettingOffset = 0x07 // u8; bit 2 = queue_item_notify
	RsCompID                RenderSettingOffset = 0x08 // u32
	RsStatus                RenderSettingOffset = 0x0C // u32
	RsTimeSpanStartDividend RenderSettingOffset = 0x14 // u32
	RsTimeSpanStartDivisor  RenderSettingOffset = 0x18 // u32
	RsTimeSpanDurDividend   RenderSettingOffset = 0x1C // u32
	RsTimeSpanDurDivisor    RenderSettingOffset = 0x20 // u32
	RsFieldRender           RenderSettingOffset = 0x32 // u16
	RsPulldown              RenderSettingOffset = 0x36 // u16
	RsQuality               RenderSettingOffset = 0x38 // u16 (0xFFFF = current)
	RsResolutionX           RenderSettingOffset = 0x3A // u16
	RsResolutionY           RenderSettingOffset = 0x3C // u16
	RsEffects               RenderSettingOffset = 0x40 // u16 (0xFFFF = current)
	RsProxyUse              RenderSettingOffset = 0x44 // u16 (0xFFFF = current)
	RsMotionBlur            RenderSettingOffset = 0x48 // u16 (0xFFFF = current)
	RsFrameBlending         RenderSettingOffset = 0x4C // u16 (0xFFFF = current)
	RsLogType               RenderSettingOffset = 0x50 // u16 (raw; the reference parser maps to 3xxx enum)
	RsSkipExistingFiles     RenderSettingOffset = 0x54 // u16 (bool)
	RsTemplateName          RenderSettingOffset = 0x5A // 64 bytes, windows-1252, NUL-padded
	RsTemplateNameLen                           = 64   // length (not an offset)
	RsUseThisFrameRate      RenderSettingOffset = 2144 // u16 (FrameRateSetting: 0=comp, 1=this)
	RsTimeSpanSource        RenderSettingOffset = 2148 // u16
	RsSoloSwitches          RenderSettingOffset = 2164 // u16 (0xFFFF = current)
	RsDiskCache             RenderSettingOffset = 2168 // u16 (0xFFFF = current)
	RsGuideLayers           RenderSettingOffset = 2172 // u16 (0xFFFF = current)
	RsColorDepth            RenderSettingOffset = 2180 // u16 (0xFFFF = current)
	RsElapsedSeconds        RenderSettingOffset = 2202 // u32
)

// time_span_source enum values (the reference parser's TimeSpanSource).
const (
	TimeSpanLengthOfComp = 0
	TimeSpanWorkAreaOnly = 1
	TimeSpanCustom       = 2
)

// RenderSettingsBlock holds the decoded fields of one settings block.
type RenderSettingsBlock struct {
	CompID            uint32
	Status            uint32
	TemplateName      string
	TimeSpanSource    uint16
	TsStartDividend   uint32
	TsStartDivisor    uint32
	TsDurationDivdend uint32
	TsDurationDivisor uint32

	// slice-2 render settings (raw u16 unless noted).
	QueueItemNotify   bool
	LogType           uint16
	ElapsedSeconds    uint32
	FieldRender       uint16
	Pulldown          uint16
	Quality           uint16
	ResolutionX       uint16
	ResolutionY       uint16
	Effects           uint16
	ProxyUse          uint16
	MotionBlur        uint16
	FrameBlending     uint16
	SkipExistingFiles bool
	UseThisFrameRate  uint16
	SoloSwitches      uint16
	DiskCache         uint16
	GuideLayers       uint16
	ColorDepth        uint16
}

// decodeRenderSettings decodes one 2246-byte settings block. Returns false if
// the block is too short to hold the fields we read.
func DecodeRenderSettings(b []byte) (RenderSettingsBlock, bool) {
	if len(b) < RenderSettingsItemSize {
		return RenderSettingsBlock{}, false
	}
	u16 := func(off RenderSettingOffset) uint16 { return binary.BigEndian.Uint16(b[off:]) }
	rs := RenderSettingsBlock{
		CompID:            binary.BigEndian.Uint32(b[RsCompID:]),
		Status:            binary.BigEndian.Uint32(b[RsStatus:]),
		TemplateName:      DecodeWin1252(b[RsTemplateName : RsTemplateName+RsTemplateNameLen]),
		TimeSpanSource:    u16(RsTimeSpanSource),
		TsStartDividend:   binary.BigEndian.Uint32(b[RsTimeSpanStartDividend:]),
		TsStartDivisor:    binary.BigEndian.Uint32(b[RsTimeSpanStartDivisor:]),
		TsDurationDivdend: binary.BigEndian.Uint32(b[RsTimeSpanDurDividend:]),
		TsDurationDivisor: binary.BigEndian.Uint32(b[RsTimeSpanDurDivisor:]),
		QueueItemNotify:   b[RsFlagByte]&(1<<2) != 0,
		LogType:           u16(RsLogType),
		ElapsedSeconds:    binary.BigEndian.Uint32(b[RsElapsedSeconds:]),
		FieldRender:       u16(RsFieldRender),
		Pulldown:          u16(RsPulldown),
		Quality:           u16(RsQuality),
		ResolutionX:       u16(RsResolutionX),
		ResolutionY:       u16(RsResolutionY),
		Effects:           u16(RsEffects),
		ProxyUse:          u16(RsProxyUse),
		MotionBlur:        u16(RsMotionBlur),
		FrameBlending:     u16(RsFrameBlending),
		SkipExistingFiles: u16(RsSkipExistingFiles) != 0,
		UseThisFrameRate:  u16(RsUseThisFrameRate),
		SoloSwitches:      u16(RsSoloSwitches),
		DiskCache:         u16(RsDiskCache),
		GuideLayers:       u16(RsGuideLayers),
		ColorDepth:        u16(RsColorDepth),
	}
	return rs, true
}

// CurrentSettingsInt maps a raw u16 render-setting value to the reference
// parser's NUMBER semantics: 0xFFFF ("current settings") becomes -1,
// everything else is the value as-is.
func CurrentSettingsInt(v uint16) int {
	if v == 0xFFFF {
		return -1
	}
	return int(v)
}

// DecodeWin1252 decodes a NUL-padded windows-1252 byte slice to a string.
// Bytes < 0x80 are ASCII; 0x80–0xFF are mapped to their Unicode code points
// (latin-1 superset, sufficient for render-settings template names). Trailing
// NULs are trimmed.
func DecodeWin1252(b []byte) string {
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
