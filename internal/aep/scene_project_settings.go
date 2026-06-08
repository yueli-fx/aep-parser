package aep

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// Project-level single-field setting chunks — py-aep parity P1 1D.
//
// All chunks sit as direct children of the root RIFX (Egg!) LIST:
//
//	head  — version_word + counters; we expose Revision() R only
//	lnrb  — flag chunk: linear_blending (presence = true; toggle = add/remove)
//	lnrp  — flag chunk: linearize_working_space (same semantics)
//	acer  — 1 byte bool: compensate_for_scene_referred_profiles
//	adfr  — 8 byte f64 BE: audio_sample_rate (Hz)
//	dwga  — byte 0: working_gamma selector (0 → 2.2, ≠0 → 2.4)
//	gpuG LIST → Utf8 — gpu_accel_type (UUID-style device id; length-variable string)
//	ExEn LIST → Utf8 — expression_engine ("extendscript" / "javascript-1.0")
//
// Source: py-aep `models/project.py` + `binary/scalar_chunks.py` +
// `binary/misc_chunks.py`. Verified against re_cameralight.aep dumps.

// ──────────────────────────────────────────────
// Revision (head[18..19] uint16 BE)
// ──────────────────────────────────────────────

// XmpPacket returns the project's XMP metadata packet (trailing UTF-8
// XML after the RIFX root chunk). Returns "" when the file has no XMP
// trailer or the project was built outside the parser.
//
// py-aep parity: matches `Project.xmp_packet` reader. Currently R only —
// SetXmpPacket would require care to keep the AEP loader happy (AE may
// validate XML structure). Round-tripping through WriteAEP preserves
// the original bytes verbatim via Chunk.Trailing.
func (p *Project) XmpPacket() string {
	pb := p.projectBack()
	if pb == nil || pb.root == nil || len(pb.root.Trailing) == 0 {
		return ""
	}
	return string(pb.root.Trailing)
}

// Revision returns the project's file revision counter — incremented
// by AE on each save. Read-only here since the head chunk also carries
// the next-item-id counter that our writer manages independently.
// Returns 0 when the file has no head chunk or its data is too short.
func (p *Project) Revision() uint16 {
	pb := p.projectBack()
	if pb == nil || pb.root == nil {
		return 0
	}
	head := pb.root.FindFirst(chunkIDHead)
	if head == nil || len(head.Data) < 20 {
		return 0
	}
	return binary.BigEndian.Uint16(head.Data[18:20])
}

// Toggle flag chunks lnrb / lnrp (presence-encoded) live in
// write_project_settings.go — toggling them mutates the root chunk tree,
// which belongs to the write stage rather than this scene accessor file.

// ──────────────────────────────────────────────
// acer (1 byte bool)
// ──────────────────────────────────────────────

// CompensateForSceneReferredProfiles reports whether AE compensates
// for scene-referred profiles when rendering (acer byte 0 != 0).
// Returns false when the chunk is absent.
func (p *Project) CompensateForSceneReferredProfiles() bool {
	pb := p.projectBack()
	if pb == nil || pb.acerChunk == nil || len(pb.acerChunk.Data) < 1 {
		return false
	}
	return pb.acerChunk.Data[0] != 0
}

// SetCompensateForSceneReferredProfiles writes the acer byte. Refuses
// when the acer chunk is absent (no slot to mutate; AE 23+ writes it
// by default — files without it are rare).
func (p *Project) SetCompensateForSceneReferredProfiles(v bool) error {
	if p.back == nil {
		return fmt.Errorf("project: no acer chunk — cannot SetCompensateForSceneReferredProfiles")
	}
	return p.back.SetCompensateForSceneReferredProfiles(v)
}

// ──────────────────────────────────────────────
// adfr (8 byte f64 BE — audio sample rate in Hz)
// ──────────────────────────────────────────────

// validAudioSampleRates is the set of values AE's Project Settings →
// Audio dialog exposes. SetAudioSampleRate rejects other values.
var validAudioSampleRates = []float64{22050, 32000, 44100, 48000, 96000}

// AudioSampleRate returns the project's audio sample rate in Hz.
// Returns 0 when the adfr chunk is absent or too short.
func (p *Project) AudioSampleRate() float64 {
	pb := p.projectBack()
	if pb == nil || pb.adfrChunk == nil || len(pb.adfrChunk.Data) < 8 {
		return 0
	}
	bits := binary.BigEndian.Uint64(pb.adfrChunk.Data[0:8])
	return math.Float64frombits(bits)
}

// SetAudioSampleRate writes the adfr f64 BE. Refuses unknown rates
// outside AE's UI-supported set.
func (p *Project) SetAudioSampleRate(rate float64) error {
	if p.back == nil {
		return fmt.Errorf("project: no adfr chunk — cannot SetAudioSampleRate")
	}
	return p.back.SetAudioSampleRate(rate)
}

// ──────────────────────────────────────────────
// dwga (byte 0 selector — working gamma)
// ──────────────────────────────────────────────

// WorkingGamma returns the working gamma value (2.2 or 2.4) per the
// dwga byte 0 selector (0 → 2.2, non-zero → 2.4). Returns 2.2 default
// when dwga is absent.
func (p *Project) WorkingGamma() float64 {
	pb := p.projectBack()
	if pb == nil || pb.dwgaChunk == nil || len(pb.dwgaChunk.Data) < 1 {
		return 2.2
	}
	if pb.dwgaChunk.Data[0] == 0 {
		return 2.2
	}
	return 2.4
}

// SetWorkingGamma writes the dwga selector byte. Refuses values other
// than 2.2 and 2.4 (AE's only UI options).
func (p *Project) SetWorkingGamma(gamma float64) error {
	if p.back == nil {
		return fmt.Errorf("project: no dwga chunk — cannot SetWorkingGamma")
	}
	return p.back.SetWorkingGamma(gamma)
}

// ──────────────────────────────────────────────
// gpuG (Utf8 child — GPU acceleration device id, UUID-style)
// ──────────────────────────────────────────────

// GpuAccelType returns the project's GPU acceleration device id
// (UUID-style string inside the gpuG LIST → Utf8 child). py-aep models
// this as a labeled enum; we surface the raw string since AE writes a
// runtime-resolved device UUID. Empty string when the chunk is absent.
func (p *Project) GpuAccelType() string {
	pb := p.projectBack()
	if pb == nil || pb.gpugUtf8 == nil {
		return ""
	}
	return pb.gpugUtf8.Text()
}

// SetGpuAccelType replaces the gpuG Utf8 string in-place
// (length-variable splice; WriteAEP recomputes parent LIST size).
func (p *Project) SetGpuAccelType(s string) error {
	if p.back == nil {
		return fmt.Errorf("project: no gpuG/Utf8 chunk — cannot SetGpuAccelType")
	}
	return p.back.SetGpuAccelType(s)
}

// ──────────────────────────────────────────────
// ExEn (Utf8 child — expression engine)
// ──────────────────────────────────────────────

// ExpressionEngine returns the project's expression engine name
// ("extendscript" or "javascript-1.0"). Defaults to "extendscript"
// when the ExEn chunk is absent — matches py-aep's behavior.
func (p *Project) ExpressionEngine() string {
	pb := p.projectBack()
	if pb == nil || pb.exenUtf8 == nil {
		return "extendscript"
	}
	return pb.exenUtf8.Text()
}

// SetExpressionEngine writes the ExEn Utf8. Refuses values other than
// "extendscript" / "javascript-1.0" — matches py-aep's validator.
//
// When the project has no ExEn chunk yet (parser found none), we
// refuse rather than synthesize one — adding a new top-level LIST
// requires AE-side ship-gate verification.
func (p *Project) SetExpressionEngine(engine string) error {
	switch engine {
	case "extendscript", "javascript-1.0":
		// ok
	default:
		return fmt.Errorf("project: expression_engine %q invalid; must be \"extendscript\" or \"javascript-1.0\"", engine)
	}
	if p.back == nil {
		return fmt.Errorf("project: no ExEn chunk — cannot SetExpressionEngine on file without one")
	}
	return p.back.SetExpressionEngine(engine)
}

// ──────────────────────────────────────────────
// nnhd (40 bytes — project display settings)
// ──────────────────────────────────────────────
//
// Byte layout (from py-aep item_chunks.py NnhdChunk):
// - Bytes 0-7: reserved
// - Byte 8: _display_byte
//   - bit 7 = feet_frames_film_type (0=MM35, 1=MM16)
//   - bits 6-0 = time_display_type (0=TIMECODE, 1=FRAMES)
// - Byte 9: footage_timecode_display_start_type (0=Start0, 1=UseSourceMedia)
// - Byte 10: reserved
// - Byte 11: _feet_byte
//   - bit 0 = frames_use_feet_frames
// - Bytes 12-13: reserved
// - Bytes 14-15: timecode_default_base (u2 BE, 1-999)
// - Bytes 16-19: unknown (default 0x00000010)
// - Byte 20: frames_count_type (0=Start0, 1=Start1, 2=TimecodeConversion)
// - Bytes 21-23: reserved
// - Byte 24: bits_per_channel (already implemented)
// - Byte 25: transparency_grid_thumbnails (bool)
// - Bytes 26-39: unknown

// FeetFramesFilmType returns the film type for feet+frames timecode display.
// Returns FeetFramesFilmTypeMM35 (0) when nnhd chunk is absent.
func (p *Project) FeetFramesFilmType() FeetFramesFilmType {
	pb := p.projectBack()
	if pb == nil || pb.nnhdChunk == nil || len(pb.nnhdChunk.Data) < 9 {
		return FeetFramesFilmTypeMM35
	}
	// Byte 8, bit 7
	if pb.nnhdChunk.Data[8]&0x80 != 0 {
		return FeetFramesFilmTypeMM16
	}
	return FeetFramesFilmTypeMM35
}

// SetFeetFramesFilmType writes the film type to nnhd byte 8, bit 7.
func (p *Project) SetFeetFramesFilmType(v FeetFramesFilmType) error {
	if p.back == nil {
		return fmt.Errorf("project: no nnhd chunk — cannot SetFeetFramesFilmType")
	}
	return p.back.SetFeetFramesFilmType(v)
}

// FootageTimecodeDisplayStartType returns how timecode is displayed for footage.
// Returns FootageTimecodeDisplayStartTypeStart0 (0) when nnhd chunk is absent.
func (p *Project) FootageTimecodeDisplayStartType() FootageTimecodeDisplayStartType {
	pb := p.projectBack()
	if pb == nil || pb.nnhdChunk == nil || len(pb.nnhdChunk.Data) < 10 {
		return FootageTimecodeDisplayStartTypeStart0
	}
	return FootageTimecodeDisplayStartType(pb.nnhdChunk.Data[9])
}

// SetFootageTimecodeDisplayStartType writes the timecode display start type to nnhd byte 9.
func (p *Project) SetFootageTimecodeDisplayStartType(v FootageTimecodeDisplayStartType) error {
	if p.back == nil {
		return fmt.Errorf("project: no nnhd chunk — cannot SetFootageTimecodeDisplayStartType")
	}
	return p.back.SetFootageTimecodeDisplayStartType(v)
}

// TimecodeDefaultBase returns the default timecode base (1-999).
// Returns 0 when nnhd chunk is absent.
func (p *Project) TimecodeDefaultBase() int {
	pb := p.projectBack()
	if pb == nil || pb.nnhdChunk == nil || len(pb.nnhdChunk.Data) < 16 {
		return 0
	}
	return int(binary.BigEndian.Uint16(pb.nnhdChunk.Data[14:16]))
}

// SetTimecodeDefaultBase writes the timecode default base to nnhd bytes 14-15.
func (p *Project) SetTimecodeDefaultBase(v int) error {
	if p.back == nil {
		return fmt.Errorf("project: no nnhd chunk — cannot SetTimecodeDefaultBase")
	}
	return p.back.SetTimecodeDefaultBase(v)
}

// FramesCountType returns how frames are counted in the project.
// Returns FramesCountTypeStart0 (0) when nnhd chunk is absent.
func (p *Project) FramesCountType() FramesCountType {
	pb := p.projectBack()
	if pb == nil || pb.nnhdChunk == nil || len(pb.nnhdChunk.Data) < 21 {
		return FramesCountTypeStart0
	}
	return FramesCountType(pb.nnhdChunk.Data[20])
}

// SetFramesCountType writes the frames count type to nnhd byte 20.
func (p *Project) SetFramesCountType(v FramesCountType) error {
	if p.back == nil {
		return fmt.Errorf("project: no nnhd chunk — cannot SetFramesCountType")
	}
	return p.back.SetFramesCountType(v)
}

// DisplayStartFrame returns the display start frame (0 or 1).
// This is derived from frames_count_type % 2.
// Returns 0 when nnhd chunk is absent.
func (p *Project) DisplayStartFrame() int {
	return int(p.FramesCountType()) % 2
}

// SetDisplayStartFrame sets the display start frame (0 or 1).
// This modifies frames_count_type to preserve the value.
func (p *Project) SetDisplayStartFrame(v int) error {
	if v != 0 && v != 1 {
		return fmt.Errorf("project: display_start_frame %d invalid; must be 0 or 1", v)
	}
	if p.back == nil {
		return fmt.Errorf("project: no nnhd chunk — cannot SetFramesCountType")
	}
	return p.back.SetDisplayStartFrame(v)
}

// FramesUseFeetFrames returns whether frames use feet+frames display.
// Returns false when nnhd chunk is absent.
func (p *Project) FramesUseFeetFrames() bool {
	pb := p.projectBack()
	if pb == nil || pb.nnhdChunk == nil || len(pb.nnhdChunk.Data) < 12 {
		return false
	}
	// Byte 11, bit 0
	return pb.nnhdChunk.Data[11]&0x01 != 0
}

// SetFramesUseFeetFrames writes the frames_use_feet_frames flag to nnhd byte 11, bit 0.
func (p *Project) SetFramesUseFeetFrames(v bool) error {
	if p.back == nil {
		return fmt.Errorf("project: no nnhd chunk — cannot SetFramesUseFeetFrames")
	}
	return p.back.SetFramesUseFeetFrames(v)
}

// TimeDisplayType returns how time is displayed in the project.
// Returns TimeDisplayTypeTimecode (0) when nnhd chunk is absent.
func (p *Project) TimeDisplayType() TimeDisplayType {
	pb := p.projectBack()
	if pb == nil || pb.nnhdChunk == nil || len(pb.nnhdChunk.Data) < 9 {
		return TimeDisplayTypeTimecode
	}
	// Byte 8, bits 6-0
	return TimeDisplayType(pb.nnhdChunk.Data[8] & 0x7F)
}

// SetTimeDisplayType writes the time display type to nnhd byte 8, bits 6-0.
func (p *Project) SetTimeDisplayType(v TimeDisplayType) error {
	if p.back == nil {
		return fmt.Errorf("project: no nnhd chunk — cannot SetTimeDisplayType")
	}
	return p.back.SetTimeDisplayType(v)
}

// TransparencyGridThumbnails returns whether transparency grid is shown in thumbnails.
// Returns false when nnhd chunk is absent.
func (p *Project) TransparencyGridThumbnails() bool {
	pb := p.projectBack()
	if pb == nil || pb.nnhdChunk == nil || len(pb.nnhdChunk.Data) < 26 {
		return false
	}
	return pb.nnhdChunk.Data[25] != 0
}

// SetTransparencyGridThumbnails writes the transparency grid thumbnails flag to nnhd byte 25.
func (p *Project) SetTransparencyGridThumbnails(v bool) error {
	if p.back == nil {
		return fmt.Errorf("project: no nnhd chunk — cannot SetTransparencyGridThumbnails")
	}
	return p.back.SetTransparencyGridThumbnails(v)
}

// ──────────────────────────────────────────────
// CMS (Color Management System) settings (AE 24+)
// ──────────────────────────────────────────────
//
// CMS settings are stored as a Utf8 chunk containing JSON with keys:
// - colorManagementSystem: 0=Adobe, 1=OCIO
// - lutInterpolationMethod: 0=Trilinear, 1=Tetrahedral
// - ocioConfigurationFile: string path
//
// The Utf8 chunk is identified by the presence of "lutInterpolationMethod"
// or "colorManagementSystem" in its content.

// cmsSettings returns the CMS settings as a map, or nil if no CMS chunk.
// Records a parser Warning on malformed JSON so callers can surface the
// issue instead of seeing the chunk as silently absent.
func (p *Project) cmsSettings() map[string]interface{} {
	pb := p.projectBack()
	if pb == nil || pb.cmsUtf8 == nil {
		return nil
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(pb.cmsUtf8.Data, &settings); err != nil {
		p.Warnings = append(p.Warnings, fmt.Sprintf("project: CMS JSON parse failed: %v", err))
		return nil
	}
	return settings
}

// ColorManagementSystem returns the color management system used by the project.
// Returns ColorManagementSystemAdobe (0) when CMS chunk is absent.
func (p *Project) ColorManagementSystem() ColorManagementSystem {
	settings := p.cmsSettings()
	if settings == nil {
		return ColorManagementSystemAdobe
	}
	if v, ok := settings["colorManagementSystem"].(float64); ok {
		return ColorManagementSystem(v)
	}
	return ColorManagementSystemAdobe
}

// SetColorManagementSystem sets the color management system.
// Rejects values other than the defined enum members.
func (p *Project) SetColorManagementSystem(v ColorManagementSystem) error {
	switch v {
	case ColorManagementSystemAdobe, ColorManagementSystemOCIO:
	default:
		return fmt.Errorf("project: color_management_system %d invalid; must be Adobe(0) or OCIO(1)", v)
	}
	if p.back == nil {
		return fmt.Errorf("project: no CMS chunk — cannot set %s (only AE 24+ projects with CMS already saved are supported)", "colorManagementSystem")
	}
	return p.back.SetColorManagementSystem(v)
}

// LutInterpolationMethod returns the LUT interpolation method.
// Returns LutInterpolationMethodTrilinear (0) when CMS chunk is absent.
func (p *Project) LutInterpolationMethod() LutInterpolationMethod {
	settings := p.cmsSettings()
	if settings == nil {
		return LutInterpolationMethodTrilinear
	}
	if v, ok := settings["lutInterpolationMethod"].(float64); ok {
		return LutInterpolationMethod(v)
	}
	return LutInterpolationMethodTrilinear
}

// SetLutInterpolationMethod sets the LUT interpolation method.
// Rejects values other than the defined enum members.
func (p *Project) SetLutInterpolationMethod(v LutInterpolationMethod) error {
	switch v {
	case LutInterpolationMethodTrilinear, LutInterpolationMethodTetrahedral:
	default:
		return fmt.Errorf("project: lut_interpolation_method %d invalid; must be Trilinear(0) or Tetrahedral(1)", v)
	}
	if p.back == nil {
		return fmt.Errorf("project: no CMS chunk — cannot set %s (only AE 24+ projects with CMS already saved are supported)", "lutInterpolationMethod")
	}
	return p.back.SetLutInterpolationMethod(v)
}

// OcioConfigurationFile returns the OCIO configuration file path.
// Returns empty string when CMS chunk is absent.
func (p *Project) OcioConfigurationFile() string {
	settings := p.cmsSettings()
	if settings == nil {
		return ""
	}
	if v, ok := settings["ocioConfigurationFile"].(string); ok {
		return v
	}
	return ""
}

// SetOcioConfigurationFile sets the OCIO configuration file path.
func (p *Project) SetOcioConfigurationFile(v string) error {
	if p.back == nil {
		return fmt.Errorf("project: no CMS chunk — cannot set %s (only AE 24+ projects with CMS already saved are supported)", "ocioConfigurationFile")
	}
	return p.back.SetOcioConfigurationFile(v)
}

// WorkingSpace returns the working color space name (R only).
// Returns "None" when the chunk is absent or has no baseColorProfile.
func (p *Project) WorkingSpace() string {
	pb := p.projectBack()
	if pb == nil || pb.cmsUtf8 == nil {
		return "None"
	}
	if !strings.Contains(string(pb.cmsUtf8.Data), "baseColorProfile") {
		return "None"
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(pb.cmsUtf8.Data, &settings); err != nil {
		return "None"
	}
	profile, ok := settings["baseColorProfile"].(map[string]interface{})
	if !ok {
		return "None"
	}
	name, ok := profile["colorProfileName"].(string)
	if !ok {
		return "None"
	}
	return name
}
