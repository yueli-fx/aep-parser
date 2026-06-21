package scene

import (
	"encoding/binary"
	"math"
)

// Footage convenience helpers — convenience accessors for footage items
// (asset type, has audio, start frame, etc.). Most read from sspc
// at byte offsets sourced from the reference parser's `binary/footage_chunks.py`:
//
//	@0x78  (120)  footage_missing_at_save (1 byte bool, 1 = missing)
//	@0xA0  (160)  audio_sample_rate (8 byte f64 BE; 0 = no audio)
//	@0xAC  (172)  start_frame (uint32 BE)
//	@0xB0  (176)  end_frame   (uint32 BE)
//
// Synthesized fixtures with a short sspc (< 222 bytes) return zero
// values for these helpers — the layout doesn't apply.

const (
	sspcOffFootageMissing  = 0x78
	sspcOffAudioSampleRate = 0xA0
	sspcOffStartFrame      = 0xAC
	sspcOffEndFrame        = 0xB0
)

// AssetType returns the footage kind as a string: "placeholder" /
// "solid" / "file". Footage items with neither flag default to "file".
func (f *Footage) AssetType() string {
	switch {
	case f.IsPlaceholder:
		return "placeholder"
	case f.IsSolid:
		return "solid"
	default:
		return "file"
	}
}

// File returns the footage source file path. Empty string for solids
// and placeholders. Alias of [Footage.Path], provided as a convenience.
func (f *Footage) File() string {
	return f.Path
}

func (f *Footage) SspcData() []byte {
	if f.back == nil {
		return nil
	}
	return f.back.SspcData()
}

// FootageMissing reports whether the footage source file was missing
// at the time AE saved the project (per the `footage_missing_at_save`
// flag in sspc @0x78). Solids and placeholders never have a file, so
// this is always false for them (regardless of the on-disk flag bit).
func (f *Footage) FootageMissing() bool {
	if f.IsSolid || f.IsPlaceholder {
		return false
	}
	d := f.SspcData()
	if len(d) <= sspcOffFootageMissing {
		return false
	}
	return d[sspcOffFootageMissing] != 0
}

// HasAudio reports whether the footage has an audio stream — true
// when sspc's audio_sample_rate (8 byte f64 BE at @0xA0) is non-zero.
// Solids and placeholders never have audio.
func (f *Footage) HasAudio() bool {
	if f.IsSolid || f.IsPlaceholder {
		return false
	}
	d := f.SspcData()
	if len(d) < sspcOffAudioSampleRate+8 {
		return false
	}
	bits := binary.BigEndian.Uint64(d[sspcOffAudioSampleRate : sspcOffAudioSampleRate+8])
	rate := math.Float64frombits(bits)
	return rate > 0
}

// StartFrame returns the footage start frame (sspc @0xAC, uint32 BE).
// 0 for non-sequence footage and for fixtures without a full-size sspc.
func (f *Footage) StartFrame() int {
	d := f.SspcData()
	if len(d) < sspcOffStartFrame+4 {
		return 0
	}
	return int(binary.BigEndian.Uint32(d[sspcOffStartFrame : sspcOffStartFrame+4]))
}

// EndFrame returns the footage end frame (sspc @0xB0, uint32 BE).
// 0 for non-sequence footage and for fixtures without a full-size sspc.
func (f *Footage) EndFrame() int {
	d := f.SspcData()
	if len(d) < sspcOffEndFrame+4 {
		return 0
	}
	return int(binary.BigEndian.Uint32(d[sspcOffEndFrame : sspcOffEndFrame+4]))
}
