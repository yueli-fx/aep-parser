package codec

import (
	"encoding/binary"
	"math"
	"strings"
)

// FormatOptions is a typed view of the Ropt chunk (format-specific render
// options). Kind names the active format; only that format's fields are
// populated. Read-only (P3 §3A slice-4). HDR10 metadata + XML format options
// (AVI/H264) are deferred.
type FormatOptions struct {
	Kind string // cineon / jpeg / openexr / png / targa / tiff

	// cineon (sDPX)
	TenBitBlackPoint      int
	TenBitWhitePoint      int
	ConvertedBlackPoint   float64
	ConvertedWhitePoint   float64
	CurrentGamma          float64
	HighlightExpansion    int
	LogarithmicConversion bool
	CineonFileFormat      int

	// jpeg
	Quality int

	// openexr (oEXR)
	ThirtyTwoBitFloat bool
	LuminanceChroma   bool

	// targa (TPIC)
	BitsPerPixel   int
	RLECompression bool

	// tiff (TIF )
	LZWCompression bool
	IBMPCByteOrder bool

	// png (png!)
	Width  int
	Height int

	// shared: BitDepth (cineon/png), Compression (openexr/png)
	BitDepth    int
	Compression int
}

// codec_output_module.go — pure byte decoding of the per-output-module
// settings: the 128B OutputModuleSettingsItem (ldat under each LItm item's
// LIST:list) and the 154B Roou chunk. No scene coupling (codec_ boundary).
// Byte layout mirrors py-aep binary/render_chunks.py; offsets verified by
// field-size accounting + tmp_debug probe + golden cross-check.

const OutputModuleSettingsItemSize = 128

// OutputModuleSettingsItem (128B) field offsets.
const (
	OmsFlagByte07         RenderSettingOffset = 7  // bit7 preserve_rgb / bit6 include_source_xmp / bit4 use_region_of_interest / bit3 use_comp_frame_number
	OmsPostRenderCompID   RenderSettingOffset = 8  // u32
	OmsChannels           RenderSettingOffset = 19 // u8
	OmsResizeQuality      RenderSettingOffset = 23 // u8
	OmsResize             RenderSettingOffset = 27 // u8 bool
	OmsLockAspectRatio    RenderSettingOffset = 29 // u8 bool
	OmsFlagByte22         RenderSettingOffset = 31 // bit0 Crop
	OmsCropTop            RenderSettingOffset = 32 // u16
	OmsCropLeft           RenderSettingOffset = 34 // u16
	OmsCropBottom         RenderSettingOffset = 36 // u16
	OmsCropRight          RenderSettingOffset = 38 // u16
	OmsOutputAudio        RenderSettingOffset = 42 // u8
	OmsIncludeProjectLink RenderSettingOffset = 47 // u8 bool
	OmsPostRenderAction   RenderSettingOffset = 48 // u32
	OmsConvertLinear      RenderSettingOffset = 91 // u8
	OmsColorSpaceWorking  RenderSettingOffset = 93 // u8
)

// Roou (154B) field offsets.
const (
	RouoVideoCodec      RenderSettingOffset = 4   // 4 ascii
	RouoStartingNumber  RenderSettingOffset = 16  // u32
	RouoFormatID        RenderSettingOffset = 26  // 4 ascii
	RouoWidth           RenderSettingOffset = 36  // u16
	RouoHeight          RenderSettingOffset = 40  // u16
	RouoDepth           RenderSettingOffset = 71  // u8
	RouoColorPremult    RenderSettingOffset = 77  // u8
	RouoAudioSampleRate RenderSettingOffset = 100 // f64
	RouoAudioDisabledHi RenderSettingOffset = 108 // u8
	RouoAudioFormat     RenderSettingOffset = 109 // u8
	RouoAudioBitDepth   RenderSettingOffset = 111 // u8
	RouoAudioChannels   RenderSettingOffset = 113 // u8
)

// OmSettings holds the decoded 128B output-module settings block.
type OmSettings struct {
	UseCompFrameNumber  bool
	UseRegionOfInterest bool
	IncludeSourceXMP    bool
	PreserveRGB         bool
	Channels            int
	ResizeQuality       int
	Resize              bool
	LockAspectRatio     bool
	Crop                bool
	CropTop             int
	CropLeft            int
	CropBottom          int
	CropRight           int
	OutputAudio         int
	IncludeProjectLink  bool
	PostRenderAction    uint32
	ConvertToLinear     int
}

// decodeOMSettings decodes one 128B OutputModuleSettingsItem block.
func DecodeOMSettings(b []byte) (OmSettings, bool) {
	if len(b) < OutputModuleSettingsItemSize {
		return OmSettings{}, false
	}
	f7 := b[OmsFlagByte07]
	f22 := b[OmsFlagByte22]
	u16 := func(off RenderSettingOffset) int { return int(binary.BigEndian.Uint16(b[off:])) }
	return OmSettings{
		UseCompFrameNumber:  f7&(1<<3) != 0,
		UseRegionOfInterest: f7&(1<<4) != 0,
		IncludeSourceXMP:    f7&(1<<6) != 0,
		PreserveRGB:         f7&(1<<7) != 0,
		Channels:            int(b[OmsChannels]),
		ResizeQuality:       int(b[OmsResizeQuality]),
		Resize:              b[OmsResize] != 0,
		LockAspectRatio:     b[OmsLockAspectRatio] != 0,
		Crop:                f22&(1<<0) != 0,
		CropTop:             u16(OmsCropTop),
		CropLeft:            u16(OmsCropLeft),
		CropBottom:          u16(OmsCropBottom),
		CropRight:           u16(OmsCropRight),
		OutputAudio:         int(b[OmsOutputAudio]),
		IncludeProjectLink:  b[OmsIncludeProjectLink] != 0,
		PostRenderAction:    binary.BigEndian.Uint32(b[OmsPostRenderAction:]),
		ConvertToLinear:     int(b[OmsConvertLinear]),
	}, true
}

// RoouSettings holds the decoded fields of a Roou chunk (output options).
type RoouSettings struct {
	VideoCodec      string
	StartingNumber  uint32
	FormatID        string
	Width           int
	Height          int
	Depth           int
	ColorPremult    int
	AudioSampleRate float64
	AudioFormat     int
	AudioBitDepth   int
	AudioChannels   int
	AudioEnabled    bool
}

// DecodeRoou decodes a Roou chunk body (>=114 bytes used).
func DecodeRoou(b []byte) (RoouSettings, bool) {
	if len(b) < 114 {
		return RoouSettings{}, false
	}
	return RoouSettings{
		VideoCodec:      Ascii4(b[RouoVideoCodec:]),
		StartingNumber:  binary.BigEndian.Uint32(b[RouoStartingNumber:]),
		FormatID:        Ascii4(b[RouoFormatID:]),
		Width:           int(binary.BigEndian.Uint16(b[RouoWidth:])),
		Height:          int(binary.BigEndian.Uint16(b[RouoHeight:])),
		Depth:           int(b[RouoDepth]),
		ColorPremult:    int(b[RouoColorPremult]),
		AudioSampleRate: math.Float64frombits(binary.BigEndian.Uint64(b[RouoAudioSampleRate:])),
		AudioFormat:     int(b[RouoAudioFormat]),
		AudioBitDepth:   int(b[RouoAudioBitDepth]),
		AudioChannels:   int(b[RouoAudioChannels]),
		AudioEnabled:    b[RouoAudioDisabledHi] != 0xFF,
	}, true
}

// Ascii4 decodes a 4-byte ASCII code (NUL-trimmed) used for codec/format ids.
func Ascii4(b []byte) string {
	if len(b) < 4 {
		return ""
	}
	return strings.TrimRight(string(b[:4]), "\x00")
}

// DecodeRoptFormatOptions decodes a Ropt chunk into typed format options,
// dispatching on the 4-byte format_code. Returns nil for unknown/too-short
// bodies (e.g. XML-based formats AVI/H264/QuickTime carry no Ropt variant).
// Field offsets mirror py-aep binary/render_chunks.py Ropt variant classes;
// cineon offsets verified against the format_options/cineon/* fixtures.
func DecodeRoptFormatOptions(d []byte) *FormatOptions {
	if len(d) < 4 {
		return nil
	}
	be := binary.BigEndian
	f64 := func(off int) float64 { return math.Float64frombits(be.Uint64(d[off:])) }
	switch string(d[:4]) {
	case "sDPX": // Cineon / DPX
		if len(d) < 47 {
			return nil
		}
		return &FormatOptions{
			Kind:                  "cineon",
			TenBitBlackPoint:      int(be.Uint16(d[14:])),
			TenBitWhitePoint:      int(be.Uint16(d[16:])),
			ConvertedBlackPoint:   f64(18),
			ConvertedWhitePoint:   f64(26),
			CurrentGamma:          f64(34),
			HighlightExpansion:    int(be.Uint16(d[42:])),
			LogarithmicConversion: d[44] != 0,
			CineonFileFormat:      int(d[45]),
			BitDepth:              int(d[46]),
		}
	case "JPEG":
		if len(d) < 58 {
			return nil
		}
		return &FormatOptions{Kind: "jpeg", Quality: int(be.Uint16(d[52:]))}
	case "oEXR": // OpenEXR
		if len(d) < 17 {
			return nil
		}
		return &FormatOptions{
			Kind:              "openexr",
			Compression:       int(d[14]),
			ThirtyTwoBitFloat: d[15] != 0,
			LuminanceChroma:   d[16] != 0,
		}
	case "TPIC": // Targa
		if len(d) < 83 {
			return nil
		}
		return &FormatOptions{Kind: "targa", BitsPerPixel: int(d[77]), RLECompression: d[82] != 0}
	case "TIF ":
		if len(d) < 602 {
			return nil
		}
		return &FormatOptions{Kind: "tiff", IBMPCByteOrder: d[600] != 0, LZWCompression: d[601] != 0}
	case "png!":
		if len(d) < 34 {
			return nil
		}
		return &FormatOptions{
			Kind:        "png",
			Width:       int(be.Uint32(d[18:])),
			Height:      int(be.Uint32(d[22:])),
			BitDepth:    int(be.Uint16(d[28:])),
			Compression: int(be.Uint32(d[30:])),
		}
	}
	return nil
}
