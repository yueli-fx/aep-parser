package aep

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/rifx"
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

// Revision returns the project's file revision counter — incremented
// by AE on each save. Read-only here since the head chunk also carries
// the next-item-id counter that our writer manages independently.
// Returns 0 when the file has no head chunk or its data is too short.
func (p *Project) Revision() uint16 {
	if p.root == nil {
		return 0
	}
	head := p.root.FindFirst(chunkIDHead)
	if head == nil || len(head.Data) < 20 {
		return 0
	}
	return binary.BigEndian.Uint16(head.Data[18:20])
}

// ──────────────────────────────────────────────
// Toggle flag chunks: lnrb / lnrp
// ──────────────────────────────────────────────

// rootChunkPresent reports whether root has a direct child chunk
// (not LIST) with the given id.
func (p *Project) rootChunkPresent(id rifx.ChunkID) bool {
	if p.root == nil {
		return false
	}
	for _, c := range p.root.Children {
		if !c.IsList() && c.ID == id {
			return true
		}
	}
	return false
}

// setRootFlagChunk adds (when on=true) or removes (when on=false) a
// presence-encoded flag chunk on root.
//
// **Layout** (RE'd 2026-05-26 against AE 2025-saved fixture
// `re_linear_blending_on.aep`): the chunk carries **1 byte 0x01**, NOT
// zero-length. Empty payload makes AE reject the file with "文件数据
// 丢失"/"file data missing".
//
// **Position** matters: AE inserts lnrb / lnrp immediately AFTER the
// root `cpid` chunk (color-management profile id) and before `dwga`.
// Append-to-end likewise causes "file data missing". We insert right
// after the existing cpid (or fall back to before dwga, or append if
// neither anchor exists).
func (p *Project) setRootFlagChunk(id rifx.ChunkID, on bool) error {
	if p.root == nil {
		return fmt.Errorf("project: cannot toggle %s — no root chunk (built outside parser)", id)
	}
	idx := -1
	for i, c := range p.root.Children {
		if !c.IsList() && c.ID == id {
			idx = i
			break
		}
	}
	if on && idx < 0 {
		insertAt := flagChunkInsertPosition(p.root)
		newChunk := &rifx.Chunk{ID: id, Data: []byte{0x01}}
		p.root.Children = append(p.root.Children, nil)
		copy(p.root.Children[insertAt+1:], p.root.Children[insertAt:])
		p.root.Children[insertAt] = newChunk
	}
	if !on && idx >= 0 {
		p.root.Children = append(p.root.Children[:idx], p.root.Children[idx+1:]...)
	}
	return nil
}

// flagChunkInsertPosition returns the index where a new lnrb / lnrp
// flag chunk should be inserted on root. Per RE: after `cpid`,
// otherwise before `dwga`, otherwise at end.
func flagChunkInsertPosition(root *rifx.Chunk) int {
	for i, c := range root.Children {
		if !c.IsList() && c.ID == chunkIDCpid {
			return i + 1
		}
	}
	for i, c := range root.Children {
		if !c.IsList() && c.ID == rifx.IDDwga {
			return i
		}
	}
	return len(root.Children)
}

// chunkIDCpid is the root-level color-profile id chunk, used as the
// insertion anchor for lnrb / lnrp flag chunks. Not exposed in rifx
// since no other code path needs it.
var chunkIDCpid = rifx.ChunkID{'c', 'p', 'i', 'd'}

// LinearBlending reports whether the project uses linear blending
// (presence of lnrb chunk under root).
func (p *Project) LinearBlending() bool {
	return p.rootChunkPresent(rifx.IDLnrb)
}

// SetLinearBlending toggles the lnrb chunk under root.
func (p *Project) SetLinearBlending(v bool) error {
	return p.setRootFlagChunk(rifx.IDLnrb, v)
}

// LinearizeWorkingSpace reports whether the working color space is
// linearized for blending (presence of lnrp chunk under root).
func (p *Project) LinearizeWorkingSpace() bool {
	return p.rootChunkPresent(rifx.IDLnrp)
}

// SetLinearizeWorkingSpace toggles the lnrp chunk under root.
func (p *Project) SetLinearizeWorkingSpace(v bool) error {
	return p.setRootFlagChunk(rifx.IDLnrp, v)
}

// ──────────────────────────────────────────────
// acer (1 byte bool)
// ──────────────────────────────────────────────

// CompensateForSceneReferredProfiles reports whether AE compensates
// for scene-referred profiles when rendering (acer byte 0 != 0).
// Returns false when the chunk is absent.
func (p *Project) CompensateForSceneReferredProfiles() bool {
	if p.acerChunk == nil || len(p.acerChunk.Data) < 1 {
		return false
	}
	return p.acerChunk.Data[0] != 0
}

// SetCompensateForSceneReferredProfiles writes the acer byte. Refuses
// when the acer chunk is absent (no slot to mutate; AE 23+ writes it
// by default — files without it are rare).
func (p *Project) SetCompensateForSceneReferredProfiles(v bool) error {
	if p.acerChunk == nil || len(p.acerChunk.Data) < 1 {
		return fmt.Errorf("project: no acer chunk — cannot SetCompensateForSceneReferredProfiles")
	}
	if v {
		p.acerChunk.Data[0] = 1
	} else {
		p.acerChunk.Data[0] = 0
	}
	return nil
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
	if p.adfrChunk == nil || len(p.adfrChunk.Data) < 8 {
		return 0
	}
	bits := binary.BigEndian.Uint64(p.adfrChunk.Data[0:8])
	return math.Float64frombits(bits)
}

// SetAudioSampleRate writes the adfr f64 BE. Refuses unknown rates
// outside AE's UI-supported set.
func (p *Project) SetAudioSampleRate(rate float64) error {
	if p.adfrChunk == nil || len(p.adfrChunk.Data) < 8 {
		return fmt.Errorf("project: no adfr chunk — cannot SetAudioSampleRate")
	}
	valid := false
	for _, v := range validAudioSampleRates {
		if rate == v {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("project: audio_sample_rate %v invalid; must be one of %v", rate, validAudioSampleRates)
	}
	binary.BigEndian.PutUint64(p.adfrChunk.Data[0:8], math.Float64bits(rate))
	return nil
}

// ──────────────────────────────────────────────
// dwga (byte 0 selector — working gamma)
// ──────────────────────────────────────────────

// WorkingGamma returns the working gamma value (2.2 or 2.4) per the
// dwga byte 0 selector (0 → 2.2, non-zero → 2.4). Returns 2.2 default
// when dwga is absent.
func (p *Project) WorkingGamma() float64 {
	if p.dwgaChunk == nil || len(p.dwgaChunk.Data) < 1 {
		return 2.2
	}
	if p.dwgaChunk.Data[0] == 0 {
		return 2.2
	}
	return 2.4
}

// SetWorkingGamma writes the dwga selector byte. Refuses values other
// than 2.2 and 2.4 (AE's only UI options).
func (p *Project) SetWorkingGamma(gamma float64) error {
	if p.dwgaChunk == nil || len(p.dwgaChunk.Data) < 1 {
		return fmt.Errorf("project: no dwga chunk — cannot SetWorkingGamma")
	}
	switch gamma {
	case 2.2:
		p.dwgaChunk.Data[0] = 0
	case 2.4:
		p.dwgaChunk.Data[0] = 1
	default:
		return fmt.Errorf("project: working_gamma %v invalid; must be 2.2 or 2.4", gamma)
	}
	return nil
}

// ──────────────────────────────────────────────
// gpuG (Utf8 child — GPU acceleration device id, UUID-style)
// ──────────────────────────────────────────────

// GpuAccelType returns the project's GPU acceleration device id
// (UUID-style string inside the gpuG LIST → Utf8 child). py-aep models
// this as a labeled enum; we surface the raw string since AE writes a
// runtime-resolved device UUID. Empty string when the chunk is absent.
func (p *Project) GpuAccelType() string {
	if p.gpugUtf8 == nil {
		return ""
	}
	return p.gpugUtf8.Text()
}

// SetGpuAccelType replaces the gpuG Utf8 string in-place
// (length-variable splice; WriteAEP recomputes parent LIST size).
func (p *Project) SetGpuAccelType(s string) error {
	if p.gpugUtf8 == nil {
		return fmt.Errorf("project: no gpuG/Utf8 chunk — cannot SetGpuAccelType")
	}
	p.gpugUtf8.Data = []byte(s)
	return nil
}

// ──────────────────────────────────────────────
// ExEn (Utf8 child — expression engine)
// ──────────────────────────────────────────────

// ExpressionEngine returns the project's expression engine name
// ("extendscript" or "javascript-1.0"). Defaults to "extendscript"
// when the ExEn chunk is absent — matches py-aep's behavior.
func (p *Project) ExpressionEngine() string {
	if p.exenUtf8 == nil {
		return "extendscript"
	}
	return p.exenUtf8.Text()
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
	if p.exenUtf8 == nil {
		return fmt.Errorf("project: no ExEn chunk — cannot SetExpressionEngine on file without one")
	}
	p.exenUtf8.Data = []byte(engine)
	return nil
}
