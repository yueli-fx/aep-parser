package aep

import (
	"encoding/binary"
	"math"
	"strings"
)

// codec_output_module.go — pure byte decoding of the per-output-module
// settings: the 128B OutputModuleSettingsItem (ldat under each LItm item's
// LIST:list) and the 154B Roou chunk. No scene coupling (codec_ boundary).
// Byte layout mirrors py-aep binary/render_chunks.py; offsets verified by
// field-size accounting + tmp_debug probe + golden cross-check.

const outputModuleSettingsItemSize = 128

// OutputModuleSettingsItem (128B) field offsets.
const (
	omsFlagByte07         = 7  // bit7 preserve_rgb / bit6 include_source_xmp / bit4 use_region_of_interest / bit3 use_comp_frame_number
	omsPostRenderCompID   = 8  // u32
	omsChannels           = 19 // u8
	omsResizeQuality      = 23 // u8
	omsResize             = 27 // u8 bool
	omsLockAspectRatio    = 29 // u8 bool
	omsFlagByte22         = 31 // bit0 crop
	omsCropTop            = 32 // u16
	omsCropLeft           = 34 // u16
	omsCropBottom         = 36 // u16
	omsCropRight          = 38 // u16
	omsOutputAudio        = 42 // u8
	omsIncludeProjectLink = 47 // u8 bool
	omsPostRenderAction   = 48 // u32
	omsConvertLinear      = 91 // u8
	omsColorSpaceWorking  = 93 // u8
)

// Roou (154B) field offsets.
const (
	roouVideoCodec      = 4   // 4 ascii
	roouStartingNumber  = 16  // u32
	roouFormatID        = 26  // 4 ascii
	roouWidth           = 36  // u16
	roouHeight          = 40  // u16
	roouDepth           = 71  // u8
	roouColorPremult    = 77  // u8
	roouAudioSampleRate = 100 // f64
	roouAudioDisabledHi = 108 // u8
	roouAudioFormat     = 109 // u8
	roouAudioBitDepth   = 111 // u8
	roouAudioChannels   = 113 // u8
)

// omSettings holds the decoded 128B output-module settings block.
type omSettings struct {
	useCompFrameNumber  bool
	useRegionOfInterest bool
	includeSourceXMP    bool
	preserveRGB         bool
	channels            int
	resizeQuality       int
	resize              bool
	lockAspectRatio     bool
	crop                bool
	cropTop             int
	cropLeft            int
	cropBottom          int
	cropRight           int
	outputAudio         int
	includeProjectLink  bool
	postRenderAction    uint32
	convertToLinear     int
}

// decodeOMSettings decodes one 128B OutputModuleSettingsItem block.
func decodeOMSettings(b []byte) (omSettings, bool) {
	if len(b) < outputModuleSettingsItemSize {
		return omSettings{}, false
	}
	f7 := b[omsFlagByte07]
	f22 := b[omsFlagByte22]
	u16 := func(off int) int { return int(binary.BigEndian.Uint16(b[off:])) }
	return omSettings{
		useCompFrameNumber:  f7&(1<<3) != 0,
		useRegionOfInterest: f7&(1<<4) != 0,
		includeSourceXMP:    f7&(1<<6) != 0,
		preserveRGB:         f7&(1<<7) != 0,
		channels:            int(b[omsChannels]),
		resizeQuality:       int(b[omsResizeQuality]),
		resize:              b[omsResize] != 0,
		lockAspectRatio:     b[omsLockAspectRatio] != 0,
		crop:                f22&(1<<0) != 0,
		cropTop:             u16(omsCropTop),
		cropLeft:            u16(omsCropLeft),
		cropBottom:          u16(omsCropBottom),
		cropRight:           u16(omsCropRight),
		outputAudio:         int(b[omsOutputAudio]),
		includeProjectLink:  b[omsIncludeProjectLink] != 0,
		postRenderAction:    binary.BigEndian.Uint32(b[omsPostRenderAction:]),
		convertToLinear:     int(b[omsConvertLinear]),
	}, true
}

// roouSettings holds the decoded fields of a Roou chunk (output options).
type roouSettings struct {
	videoCodec      string
	startingNumber  uint32
	formatID        string
	width           int
	height          int
	depth           int
	colorPremult    int
	audioSampleRate float64
	audioFormat     int
	audioBitDepth   int
	audioChannels   int
	audioEnabled    bool
}

// decodeRoou decodes a Roou chunk body (>=114 bytes used).
func decodeRoou(b []byte) (roouSettings, bool) {
	if len(b) < 114 {
		return roouSettings{}, false
	}
	return roouSettings{
		videoCodec:      ascii4(b[roouVideoCodec:]),
		startingNumber:  binary.BigEndian.Uint32(b[roouStartingNumber:]),
		formatID:        ascii4(b[roouFormatID:]),
		width:           int(binary.BigEndian.Uint16(b[roouWidth:])),
		height:          int(binary.BigEndian.Uint16(b[roouHeight:])),
		depth:           int(b[roouDepth]),
		colorPremult:    int(b[roouColorPremult]),
		audioSampleRate: math.Float64frombits(binary.BigEndian.Uint64(b[roouAudioSampleRate:])),
		audioFormat:     int(b[roouAudioFormat]),
		audioBitDepth:   int(b[roouAudioBitDepth]),
		audioChannels:   int(b[roouAudioChannels]),
		audioEnabled:    b[roouAudioDisabledHi] != 0xFF,
	}, true
}

// ascii4 decodes a 4-byte ASCII code (NUL-trimmed) used for codec/format ids.
func ascii4(b []byte) string {
	if len(b) < 4 {
		return ""
	}
	return strings.TrimRight(string(b[:4]), "\x00")
}

// decodeRoptFormatOptions decodes a Ropt chunk into typed format options,
// dispatching on the 4-byte format_code. Returns nil for unknown/too-short
// bodies (e.g. XML-based formats AVI/H264/QuickTime carry no Ropt variant).
// Field offsets mirror py-aep binary/render_chunks.py Ropt variant classes;
// cineon offsets verified against the format_options/cineon/* fixtures.
func decodeRoptFormatOptions(d []byte) *FormatOptions {
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
