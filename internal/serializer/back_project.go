package serializer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// projectBackrefs holds the rifx.Chunk references that power Project's
// length-preserving write paths (BPC + all P1 project-settings setters,
// CMS / GPU / expression engine writers) plus the parser-owned root chunk
// and the cached root Fold LIST used by structural mutation.
//
// Lifecycle:
//   - Populated by parseProject when a Project is built from a parsed .aep
//     file.
//   - For projects built via NewProject the template-loader populates these
//     fields from the bundled .aep template (parsed re-parse closed loop).
//   - opaque is reserved for future V3 phases that need to round-trip
//     unrecognized root-level chunks (per CLAUDE.md hard constraint #5);
//     currently nil.
type projectBackrefs struct {
	// root keeps the original RIFX chunk tree so callers can mutate leaf
	// chunks (e.g. Footage.SetPath) and re-serialize via WriteAEP.
	root *rifx.Chunk

	// rootFold is the cached root Fold LIST reference; derived cache, never
	// owned (see Invariants #8). Used by structural mutation (NewComposition
	// / future NewFootage etc.).
	rootFold *rifx.Chunk

	// Project header chunks. nhed @0x0F + nnhd @0x18 both hold BPC enum
	// (0=8 / 1=16 / 2=32). SetBitsPerChannel writes both for consistency.
	nhedChunk *rifx.Chunk
	nnhdChunk *rifx.Chunk

	// Project-level single-field setting chunks (P1 1D, py-aep parity).
	// Captured by parseProject when present; mutated by Set* methods.
	// All exist as direct root children — see project_settings.go.
	acerChunk *rifx.Chunk // 1B bool — compensate_for_scene_referred_profiles
	adfrChunk *rifx.Chunk // 8B f64 BE — audio_sample_rate
	dwgaChunk *rifx.Chunk // 1-4B — byte 0 = working_gamma selector
	gpugUtf8  *rifx.Chunk // Utf8 inside gpuG LIST — gpu_accel_type (UUID)
	exenUtf8  *rifx.Chunk // Utf8 inside ExEn LIST — expression_engine
	cmsUtf8   *rifx.Chunk // Utf8 — CMS settings JSON (AE 24+)

	opaque map[rifx.ChunkID]*rifx.Chunk
}

var _ ProjectWriter = (*projectBackrefs)(nil)

// projectBack returns the concrete backrefs behind a Project's writer
// interface for serializer-stage (parse_/mutate_/write_) raw chunk access.
// Returns nil when the project was built outside the parser.
func projectBack(p *Project) *projectBackrefs {
	if pb, ok := scene.ProjectBack(p).(*projectBackrefs); ok {
		return pb
	}
	return nil
}

func (b *projectBackrefs) XmpPacket() string {
	if b == nil || b.root == nil || len(b.root.Trailing) == 0 {
		return ""
	}
	return string(b.root.Trailing)
}

// validAudioSampleRates is the set of values AE's Project Settings → Audio
// dialog exposes. SetAudioSampleRate rejects other values.
var validAudioSampleRates = []float64{22050, 32000, 44100, 48000, 96000}

// chunkIDHead is the root-level "head" chunk ChunkID. It holds project-level
// counters (max item ID + a save-sequence counter) that AE 25 validates on
// open — when a counter is below the actual item IDs, AE reports "文件数据丢失"
// (file data missing), observed in practice.
var chunkIDHead = rifx.ChunkID{'h', 'e', 'a', 'd'}

func (b *projectBackrefs) Revision() uint16 {
	if b == nil || b.root == nil {
		return 0
	}
	head := b.root.FindFirst(chunkIDHead)
	if head == nil || len(head.Data) < 20 {
		return 0
	}
	return binary.BigEndian.Uint16(head.Data[18:20])
}

func (b *projectBackrefs) VersionString() string {
	if b == nil || b.root == nil {
		return ""
	}
	head := b.root.FindFirst(chunkIDHead)
	if head == nil || len(head.Data) < 8 {
		return ""
	}
	w := binary.BigEndian.Uint32(head.Data[4:8])
	majorA := (w >> 26) & 0x1F
	majorB := (w >> 19) & 0x07
	minor := (w >> 15) & 0x0F
	build := w & 0xFF
	major := majorA*8 + majorB
	return fmt.Sprintf("%d.%dx%d", major, minor, build)
}

func (b *projectBackrefs) EffectNames() []string {
	if b == nil || b.root == nil {
		return nil
	}
	return effectNamesFromRoot(b.root)
}

func (b *projectBackrefs) CompensateForSceneReferredProfiles() bool {
	if b == nil || b.acerChunk == nil || len(b.acerChunk.Data) < 1 {
		return false
	}
	return b.acerChunk.Data[0] != 0
}

func (b *projectBackrefs) AudioSampleRate() float64 {
	if b == nil || b.adfrChunk == nil || len(b.adfrChunk.Data) < 8 {
		return 0
	}
	bits := binary.BigEndian.Uint64(b.adfrChunk.Data[0:8])
	return math.Float64frombits(bits)
}

func (b *projectBackrefs) WorkingGamma() float64 {
	if b == nil || b.dwgaChunk == nil || len(b.dwgaChunk.Data) < 1 {
		return 2.2
	}
	if b.dwgaChunk.Data[0] == 0 {
		return 2.2
	}
	return 2.4
}

func (b *projectBackrefs) GpuAccelType() (string, bool) {
	if b == nil || b.gpugUtf8 == nil {
		return "", false
	}
	return b.gpugUtf8.Text(), true
}

func (b *projectBackrefs) ExpressionEngine() (string, bool) {
	if b == nil || b.exenUtf8 == nil {
		return "", false
	}
	return b.exenUtf8.Text(), true
}

func (b *projectBackrefs) NnhdByte(off int) (byte, bool) {
	if b == nil || b.nnhdChunk == nil || len(b.nnhdChunk.Data) <= off {
		return 0, false
	}
	return b.nnhdChunk.Data[off], true
}

func (b *projectBackrefs) NnhdUint32(off int) (uint32, bool) {
	if b == nil || b.nnhdChunk == nil || len(b.nnhdChunk.Data) < off+4 {
		return 0, false
	}
	return binary.BigEndian.Uint32(b.nnhdChunk.Data[off : off+4]), true
}

// mirrorNhed writes v to nhed[off] when the legacy nhed chunk is present and
// long enough. AE reads the project display settings from the 32-byte legacy
// nhed header; the 40-byte nnhd duplicates them. A setter that touches only
// nnhd is silently ignored by AE (it reopens with the stale nhed value), so
// every display setter must keep both in sync. The nhed↔nnhd offset map (RE
// 2026-06-14, AE 2020 self-saves): nhed[8]=time, [9]=footage, [11]bit0=useFeet,
// [12]=timecodeBase, [13]=feet-frames-per-foot, [14]=framesCount, [15]=bpc.
func (b *projectBackrefs) mirrorNhed(off int, v byte) {
	if b.nhedChunk != nil && len(b.nhedChunk.Data) > off {
		b.nhedChunk.Data[off] = v
	}
}

func (b *projectBackrefs) NnhdUint16(off int) (uint16, bool) {
	if b == nil || b.nnhdChunk == nil || len(b.nnhdChunk.Data) < off+2 {
		return 0, false
	}
	return binary.BigEndian.Uint16(b.nnhdChunk.Data[off : off+2]), true
}

func (b *projectBackrefs) CmsJSON() ([]byte, bool) {
	if b == nil || b.cmsUtf8 == nil {
		return nil, false
	}
	return b.cmsUtf8.Data, true
}

func (b *projectBackrefs) LinearBlendingFlag() bool {
	return b.rootChunkPresent(rifx.IDLnrb)
}

func (b *projectBackrefs) LinearizeWorkingSpaceFlag() bool {
	return b.rootChunkPresent(rifx.IDLnrp)
}

// rootChunkPresent reports whether root has a direct child chunk
// (not LIST) with the given id.
func (b *projectBackrefs) rootChunkPresent(id rifx.ChunkID) bool {
	if b == nil || b.root == nil {
		return false
	}
	for _, c := range b.root.Children {
		if !c.IsList() && c.ID == id {
			return true
		}
	}
	return false
}

func (b *projectBackrefs) SetCompensateForSceneReferredProfiles(v bool) error {
	if b.acerChunk == nil || len(b.acerChunk.Data) < 1 {
		return fmt.Errorf("project: no acer chunk — cannot SetCompensateForSceneReferredProfiles")
	}
	if v {
		b.acerChunk.Data[0] = 1
	} else {
		b.acerChunk.Data[0] = 0
	}
	return nil
}

func (b *projectBackrefs) SetAudioSampleRate(rate float64) error {
	if b.adfrChunk == nil || len(b.adfrChunk.Data) < 8 {
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
	binary.BigEndian.PutUint64(b.adfrChunk.Data[0:8], math.Float64bits(rate))
	return nil
}

func (b *projectBackrefs) SetWorkingGamma(gamma float64) error {
	if b.dwgaChunk == nil || len(b.dwgaChunk.Data) < 1 {
		return fmt.Errorf("project: no dwga chunk — cannot SetWorkingGamma")
	}
	switch gamma {
	case 2.2:
		b.dwgaChunk.Data[0] = 0
	case 2.4:
		b.dwgaChunk.Data[0] = 1
	default:
		return fmt.Errorf("project: working_gamma %v invalid; must be 2.2 or 2.4", gamma)
	}
	return nil
}

func (b *projectBackrefs) SetGpuAccelType(s string) error {
	if b.gpugUtf8 == nil {
		return fmt.Errorf("project: no gpuG/Utf8 chunk — cannot SetGpuAccelType")
	}
	b.gpugUtf8.Data = []byte(s)
	return nil
}

func (b *projectBackrefs) SetExpressionEngine(engine string) error {
	switch engine {
	case "extendscript", "javascript-1.0":
		// ok
	default:
		return fmt.Errorf("project: expression_engine %q invalid; must be \"extendscript\" or \"javascript-1.0\"", engine)
	}
	if b.exenUtf8 == nil {
		return fmt.Errorf("project: no ExEn chunk — cannot SetExpressionEngine on file without one")
	}
	b.exenUtf8.Data = []byte(engine)
	return nil
}

func (b *projectBackrefs) SetFeetFramesFilmType(v FeetFramesFilmType) error {
	if b.nnhdChunk == nil || len(b.nnhdChunk.Data) < 20 {
		return fmt.Errorf("project: no nnhd chunk — cannot SetFeetFramesFilmType")
	}
	// AE stores film type as frames-per-foot (u32 BE @0x10): 35mm = 16, 16mm = 40.
	fpf := uint32(16)
	if v == FeetFramesFilmTypeMM16 {
		fpf = 40
	}
	binary.BigEndian.PutUint32(b.nnhdChunk.Data[16:20], fpf)
	b.mirrorNhed(13, byte(fpf))
	return nil
}

func (b *projectBackrefs) SetFootageTimecodeDisplayStartType(v FootageTimecodeDisplayStartType) error {
	if b.nnhdChunk == nil || len(b.nnhdChunk.Data) < 10 {
		return fmt.Errorf("project: no nnhd chunk — cannot SetFootageTimecodeDisplayStartType")
	}
	b.nnhdChunk.Data[9] = byte(v)
	b.mirrorNhed(9, byte(v))
	return nil
}

func (b *projectBackrefs) SetTimecodeDefaultBase(v int) error {
	if b.nnhdChunk == nil || len(b.nnhdChunk.Data) < 16 {
		return fmt.Errorf("project: no nnhd chunk — cannot SetTimecodeDefaultBase")
	}
	if v < 1 || v > 999 {
		return fmt.Errorf("project: timecode_default_base %d invalid; must be 1-999", v)
	}
	binary.BigEndian.PutUint16(b.nnhdChunk.Data[14:16], uint16(v))
	b.mirrorNhed(12, byte(v)) // nhed stores the base in a single byte
	return nil
}

func (b *projectBackrefs) SetFramesCountType(v FramesCountType) error {
	if b.nnhdChunk == nil || len(b.nnhdChunk.Data) < 21 {
		return fmt.Errorf("project: no nnhd chunk — cannot SetFramesCountType")
	}
	b.nnhdChunk.Data[20] = byte(v)
	b.mirrorNhed(14, byte(v))
	return nil
}

func (b *projectBackrefs) SetDisplayStartFrame(v int) error {
	if v != 0 && v != 1 {
		return fmt.Errorf("project: display_start_frame %d invalid; must be 0 or 1", v)
	}
	if b.nnhdChunk == nil || len(b.nnhdChunk.Data) < 21 {
		return fmt.Errorf("project: no nnhd chunk — cannot SetFramesCountType")
	}
	current := FramesCountType(b.nnhdChunk.Data[20])
	newValue := FramesCountType((int(current) &^ 1) | v)
	return b.SetFramesCountType(newValue)
}

func (b *projectBackrefs) SetFramesUseFeetFrames(v bool) error {
	if b.nnhdChunk == nil || len(b.nnhdChunk.Data) < 12 {
		return fmt.Errorf("project: no nnhd chunk — cannot SetFramesUseFeetFrames")
	}
	if v {
		b.nnhdChunk.Data[11] |= 0x01
		b.mirrorNhed(11, b.nnhdChunk.Data[11])
	} else {
		b.nnhdChunk.Data[11] &^= 0x01
		b.mirrorNhed(11, b.nnhdChunk.Data[11])
	}
	return nil
}

func (b *projectBackrefs) SetTimeDisplayType(v TimeDisplayType) error {
	if b.nnhdChunk == nil || len(b.nnhdChunk.Data) < 9 {
		return fmt.Errorf("project: no nnhd chunk — cannot SetTimeDisplayType")
	}
	// byte 8 holds the full time_display_type value (0=Timecode, 1=Frames);
	// it is NOT bit-packed with feet film type (that lives at nnhd[16-19]).
	b.nnhdChunk.Data[8] = byte(v)
	b.mirrorNhed(8, byte(v))
	return nil
}

func (b *projectBackrefs) SetTransparencyGridThumbnails(v bool) error {
	if b.nnhdChunk == nil || len(b.nnhdChunk.Data) < 26 {
		return fmt.Errorf("project: no nnhd chunk — cannot SetTransparencyGridThumbnails")
	}
	if v {
		b.nnhdChunk.Data[25] = 1
	} else {
		b.nnhdChunk.Data[25] = 0
	}
	b.mirrorNhed(16, b.nnhdChunk.Data[25])
	return nil
}

// updateCmsSetting updates a single key in the CMS settings JSON.
// Refuses when no CMS chunk is present — the chunk's on-disk container
// position is AE-version-dependent and not yet RE'd, so synthesizing a
// new one risks producing files AE rejects. To enable CMS settings,
// open the project in AE 24+, toggle one CMS field, save, and reopen.
func (b *projectBackrefs) updateCmsSetting(key string, value interface{}) error {
	if b.cmsUtf8 == nil {
		return fmt.Errorf("project: no CMS chunk — cannot set %s (only AE 24+ projects with CMS already saved are supported)", key)
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(b.cmsUtf8.Data, &settings); err != nil {
		return fmt.Errorf("project: failed to parse CMS settings: %w", err)
	}
	settings[key] = value
	data, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("project: failed to marshal CMS settings: %w", err)
	}
	b.cmsUtf8.Data = data
	return nil
}

func (b *projectBackrefs) SetColorManagementSystem(v ColorManagementSystem) error {
	switch v {
	case ColorManagementSystemAdobe, ColorManagementSystemOCIO:
	default:
		return fmt.Errorf("project: color_management_system %d invalid; must be Adobe(0) or OCIO(1)", v)
	}
	return b.updateCmsSetting("colorManagementSystem", int(v))
}

func (b *projectBackrefs) SetLutInterpolationMethod(v LutInterpolationMethod) error {
	switch v {
	case LutInterpolationMethodTrilinear, LutInterpolationMethodTetrahedral:
	default:
		return fmt.Errorf("project: lut_interpolation_method %d invalid; must be Trilinear(0) or Tetrahedral(1)", v)
	}
	return b.updateCmsSetting("lutInterpolationMethod", int(v))
}

func (b *projectBackrefs) SetOcioConfigurationFile(v string) error {
	return b.updateCmsSetting("ocioConfigurationFile", v)
}

// SetBitsPerChannel writes the project's color depth (8 / 16 / 32 bpc)
// to BOTH the nhed @0x0F and nnhd @0x18 header bytes. AE stores the
// enum redundantly; we keep both in sync. The scene-side Go field
// (Project.BitsPerChannel) is synced by the Project delegate, not here.
func (b *projectBackrefs) SetBitsPerChannel(bpc BitsPerChannel) error {
	if b.nhedChunk == nil || b.nnhdChunk == nil {
		return fmt.Errorf("project: header chunks missing (built outside parser?)")
	}
	if len(b.nhedChunk.Data) <= 0x0F {
		return fmt.Errorf("project: nhed too short (len=%d) for BitsPerChannel write", len(b.nhedChunk.Data))
	}
	if len(b.nnhdChunk.Data) <= 0x18 {
		return fmt.Errorf("project: nnhd too short (len=%d) for BitsPerChannel write", len(b.nnhdChunk.Data))
	}
	b.nhedChunk.Data[0x0F] = byte(bpc)
	b.nnhdChunk.Data[0x18] = byte(bpc)
	return nil
}

// SetLinearBlending toggles the lnrb flag chunk under root.
func (b *projectBackrefs) SetLinearBlending(v bool) error {
	return b.setRootFlagChunk(rifx.IDLnrb, v)
}

// SetLinearizeWorkingSpace toggles the lnrp flag chunk under root.
func (b *projectBackrefs) SetLinearizeWorkingSpace(v bool) error {
	return b.setRootFlagChunk(rifx.IDLnrp, v)
}

// setRootFlagChunk adds (when on=true) or removes (when on=false) a
// presence-encoded flag chunk on root.
//
// **Layout** (against AE 2025-saved fixture
// `re_linear_blending_on.aep`): the chunk carries **1 byte 0x01**, NOT
// zero-length. Empty payload makes AE reject the file with "文件数据
// 丢失"/"file data missing".
//
// **Position** matters: AE inserts lnrb / lnrp immediately AFTER the
// root `cpid` chunk (color-management profile id) and before `dwga`.
// Append-to-end likewise causes "file data missing". We insert right
// after the existing cpid (or fall back to before dwga, or append if
// neither anchor exists).
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

func (b *projectBackrefs) setRootFlagChunk(id rifx.ChunkID, on bool) error {
	if b.root == nil {
		return fmt.Errorf("project: cannot toggle %s — no root chunk (built outside parser)", id)
	}
	idx := -1
	for i, c := range b.root.Children {
		if !c.IsList() && c.ID == id {
			idx = i
			break
		}
	}
	if on && idx < 0 {
		insertAt := flagChunkInsertPosition(b.root)
		newChunk := &rifx.Chunk{ID: id, Data: []byte{0x01}}
		b.root.Children = append(b.root.Children, nil)
		copy(b.root.Children[insertAt+1:], b.root.Children[insertAt:])
		b.root.Children[insertAt] = newChunk
	}
	if !on && idx >= 0 {
		b.root.Children = append(b.root.Children[:idx], b.root.Children[idx+1:]...)
	}
	return nil
}

// WriteAEP runs the serializer-stage scene→chunk sync (shape layers, render
// queue, guides, head counters) for the owning project, then serializes the
// underlying RIFX tree. The scene Project.WriteAEP delegates here so these sync
// free functions (which reach the concrete back-refs) stay serializer-side.
func (b *projectBackrefs) WriteAEP(p *Project, w io.Writer) error {
	if b.root == nil {
		return fmt.Errorf("aep: project has no underlying RIFX tree (was it built from FromReader?)")
	}
	if err := syncShapeLayerChunks(p); err != nil {
		return fmt.Errorf("sync shape layers: %w", err)
	}
	syncRenderQueue(p)
	syncGuides(p)
	syncHeadCounters(p)
	return b.root.Write(w)
}
