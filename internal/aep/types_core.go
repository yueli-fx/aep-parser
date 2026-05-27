// Package aep parses Adobe After Effects .aep project files and supports
// limited write-back of footage paths.
//
// .aep files use the RIFX (Big-Endian RIFF) binary format. This package
// reads the chunk tree via internal/rifx, extracts compositions, footage,
// folders and layers, and can serialize a (possibly mutated) project back
// to a valid .aep file.
//
// Typical use:
//
//	project, err := aep.Open("my-project.aep")
//	if err != nil { ... }
//	for _, c := range project.Compositions { ... }
//
//	// Rewrite a footage path and save:
//	project.Footage[0].SetPath(`D:\new\location\file.png`)
//	f, _ := os.Create("modified.aep")
//	defer f.Close()
//	project.WriteAEP(f)
//
// Concurrency: Project (and everything reachable from it) is NOT safe for
// concurrent use. Multiple goroutines may safely READ disjoint subtrees of
// a parsed Project, but any Set* call (Footage.SetPath, Property.SetStaticValue,
// Keyframe.SetTime, Keyframe.SetValue) mutates the underlying rifx.Chunk
// bytes in place and races with concurrent readers/writers of the same
// chunk. Wrap mutation paths in your own synchronization.
package aep

import (
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// BitsPerChannel represents color depth.
type BitsPerChannel uint8

const (
	BPC8  BitsPerChannel = 0x00
	BPC16 BitsPerChannel = 0x01
	BPC32 BitsPerChannel = 0x02
)

func (b BitsPerChannel) String() string {
	switch b {
	case BPC8:
		return "8bpc"
	case BPC16:
		return "16bpc"
	case BPC32:
		return "32bpc"
	default:
		return fmt.Sprintf("unknown(%d)", b)
	}
}

// LayerType classifies the kind of layer.
type LayerType string

const (
	LayerTypeAV      LayerType = "av"         // audio/video source layer
	LayerTypeText    LayerType = "text"       // text layer
	LayerTypeShape   LayerType = "shape"      // shape layer
	LayerTypeNull    LayerType = "null"       // null object
	LayerTypeLight   LayerType = "light"      // 3D light
	LayerTypeCamera  LayerType = "camera"     // 3D camera
	LayerTypeAdjust  LayerType = "adjustment" // adjustment layer
	LayerType3DModel LayerType = "3d-model"   // 3D Model layer (AE 24+, py-aep ThreeDModelLayer)
	LayerTypeUnknown LayerType = "unknown"
)

// ItemType classifies a project item.
type ItemType string

const (
	ItemTypeComposition ItemType = "composition"
	ItemTypeFootage     ItemType = "footage"
	ItemTypeFolder      ItemType = "folder"
	ItemTypeUnknown     ItemType = "unknown"
)

// Project holds the fully parsed contents of an .aep file.
type Project struct {
	Compositions   []*Composition
	Footage        []*Footage
	Folders        []*Folder
	BitsPerChannel BitsPerChannel

	// Warnings collects non-fatal parsing anomalies — chunks whose header
	// looked sane enough to attempt decoding but whose payload did not
	// match the expected layout (length mismatch, impossible counts, etc).
	// Each entry is a human-readable string; the parser keeps going and
	// produces a best-effort Project. Empty (nil) on a clean parse.
	Warnings []string

	// root keeps the original RIFX chunk tree so callers can mutate leaf
	// chunks (e.g. Footage.SetPath) and re-serialize via WriteAEP.
	root *rifx.Chunk

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

	// V2: derived state for structural mutation (NewComposition / 未来 NewFootage etc.)
	nextItemID uint32      // monotonic Item ID counter; never reused (see Invariants #9)
	rootFold   *rifx.Chunk // cached root Fold LIST reference; derived cache, never owned (see Invariants #8)
	target     AETarget    // which AE-version template NewProject loaded; drives per-target builder chunk selection
}

// CompositionByID returns the first composition whose ID matches id, or
// nil if no such composition exists. ID 0 is treated as no-match (mirrors
// Composition.LayerByID; real AE projects start item IDs at 1).
// Footage and Folder items share the same ID namespace but are stored in
// separate slices, so an ID belonging to a non-composition item returns
// nil here.
func (p *Project) CompositionByID(id uint32) *Composition {
	if id == 0 {
		return nil
	}
	for _, c := range p.Compositions {
		if c.ID == id {
			return c
		}
	}
	return nil
}

// CompositionByName returns the first composition whose Name matches the
// given string, or nil if none does. Matching is exact (case-sensitive).
// Comp names aren't guaranteed unique in AE; use CompositionByID when
// you need precise identity.
func (p *Project) CompositionByName(name string) *Composition {
	for _, c := range p.Compositions {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// FootageByName returns the first footage item whose Name matches the
// given string, or nil if none does. Matching is exact (case-sensitive).
func (p *Project) FootageByName(name string) *Footage {
	for _, f := range p.Footage {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// Composition represents an After Effects composition.
//
// TickRate is the ticks-per-second base used for ALL keyframe / marker
// times within this composition's layers. AE stores this per-comp in
// cdta and it varies with fps and AE version. Decoding logic:
//
//	scale := cdta_uint32_BE @ 0xA8
//	rate  := cdta_uint32_BE @ 0x08
//	if scale <= 1 {           // modern comps (any fps)
//	    TickRate = rate       // typically fps × 512 or fps × 1024
//	} else {                  // legacy NTSC comps (older AE versions)
//	    TickRate = rate * 1000 / scale  // typically 8000 for 29.97
//	}
//
// For older 29.97-fps comps this is 8000; for new 30-fps comps it's
// 30720; etc.
type Composition struct {
	ID        uint32
	Name      string
	Width     uint16
	Height    uint16
	FrameRate float64 // frames per second
	Duration  float64 // seconds
	TickRate  float64 // keyframe/marker ticks per second

	// PixelAspect is the comp's pixel aspect ratio (1.0 = square, 2.0
	// = anamorphic, 0.91 = NTSC D1, etc.). Decoded from cdta @0x90 /
	// @0x94 as an integer numerator/denominator pair.
	PixelAspect float64

	// ResolutionFactor is the comp's preview-resolution downsample, X
	// and Y. AE Scripting's CompItem.resolutionFactor: [1,1]=Full,
	// [2,2]=Half, [3,3]=Third, [4,4]=Quarter, custom pairs allowed.
	// Decoded from cdta @0x00 / @0x02 (two uint16 BE). Default [1,1].
	ResolutionFactor [2]uint16

	// Renderer is the comp's 3D rendering engine (AE Scripting's
	// CompItem.renderer). Stored as an internal match-name string in
	// the `prin` chunk of the comp's `PRin` LIST sibling. Examples:
	//
	//   "ADBE Escher"   — Advanced 3D (AE 2025 default; alias "ADBE Picasso")
	//   "ADBE Ernst"    — Cinema 4D Engine
	//   "ADBE Standard" — Classic 3D (legacy; AE 2025 may auto-promote to Advanced)
	//
	// Empty when the comp has no PRin LIST (rare; only seen on
	// programmatically built comps that skipped the AE serialization).
	Renderer string

	// BGColor is the composition background color (R, G, B), 0..255 per
	// channel. Decoded from cdta @0x34/0x35/0x36 (3 × uint8). Defaults to
	// {0,0,0} (AE's default black).
	BGColor [3]uint8

	// WorkAreaStart / WorkAreaEnd are the render-queue partial range in
	// seconds (AE's "Work Area" — N and B in the timeline). Decoded from
	// cdta dividend/divisor pairs at @0x1C/0x20 (start) and @0x24/0x28
	// (end). When AE leaves the work area "unset", it writes
	// 0xFFFFFFFF for the end dividend; we substitute Duration so callers
	// always see the effective range.
	WorkAreaStart float64
	WorkAreaEnd   float64

	// DisplayStartTime is the comp's display-time origin in seconds (AE
	// scripting's CompItem.displayStartTime / displayStartFrame). Decoded
	// from cdta @0xA4 / @0xA8 (uint32 BE dividend / divisor pair). AE
	// writes both 0 when displayStart is 0; non-zero pair encodes
	// start-time in seconds. RE fixture: re_wave2_ae24.aep (RE_CDTA_DSF_120).
	DisplayStartTime float64

	// ShutterAngle is the motion-blur shutter angle in degrees (AE UI
	// range 0..720, default 180). Decoded from cdta @0xAE as uint16 BE.
	ShutterAngle uint16

	// ShutterPhase is the motion-blur shutter phase offset. Decoded from
	// cdta @0xB4 as int32 BE. Likely degrees in the AE UI's -360..+360
	// range (fixture value -90 plausibly matches AE's -90° UI display)
	// but the exact unit is not independently UI-verified — surface as
	// raw int32 and let callers interpret.
	ShutterPhase int32

	// MotionBlurAdaptiveSampleLimit / MotionBlurSamplesPerFrame are AE's
	// motion-blur quality knobs (Composition Settings → Advanced). Decoded
	// from cdta @0xC4 / @0xC8 as int32 BE. AE clamps these to positive
	// integers in the UI (defaults: 128 and 16).
	MotionBlurAdaptiveSampleLimit int32
	MotionBlurSamplesPerFrame     int32

	Layers []*Layer

	// Markers holds composition-level markers (AE's timeline-top chapter
	// markers, distinct from per-layer markers). Stored in the AEP as a
	// pseudo-layer (`LIST SecL` named "Markers") carrying an "ADBE Marker"
	// property — decoded with the same machinery as layer markers.
	Markers []*Marker

	// proj is the owning project, set by parseProject after the comp is
	// appended. Used by Layer.SourceComposition() to walk the project's
	// other comps via Project.CompositionByID. Unexported to keep the API
	// minimal; nil for comps built outside the parser.
	proj *Project

	// cdta is the underlying cdta chunk reference, captured by
	// parseComposition. Used by SetBGColor / SetShutterAngle /
	// SetMotionBlur* / SetWorkArea for length-preserving writes.
	// Nil for comps built outside the parser; setters refuse.
	cdta *rifx.Chunk

	// nameChunk is the comp's Utf8 name chunk (length-variable Set name).
	nameChunk *rifx.Chunk

	// V2: cached owning Item LIST chunk; populated by parseComposition.
	// Used by NewComposition (re-parse closed loop) + future structural
	// mutations. derived cache, never owned (see Invariant #8).
	itemList *rifx.Chunk

	// Item-level metadata shared with Footage / Folder. Populated by
	// parseItem from the surrounding Item LIST (cmta child + idta byte).
	Comment string // Item.comment — AE's project-panel comment column
	Label   uint8  // Item.label — project-panel color index 0..16 (idta @0x3A)

	// Underlying chunk refs for the comment / label writers.
	itemCmtaChunk  *rifx.Chunk // cmta sibling under the Item LIST
	itemIdtaChunk  *rifx.Chunk // idta sibling — Label byte at payload @0x3A
	itemLayrParent *rifx.Chunk // Item LIST itself — needed for cmta insertion when missing
}

// CdtaRawBytes returns the comp's cdta chunk Data slice, or nil if the
// comp has no cdta. Read-only access for debugging / RE tools — the
// underlying byte slice is the live chunk data; do not mutate.
func (c *Composition) CdtaRawBytes() []byte {
	if c.cdta == nil {
		return nil
	}
	return c.cdta.Data
}

// LayerByID returns the first layer in this composition whose ID matches
// id, or nil if no such layer exists. ID 0 is treated as no-match (it's
// the sentinel used by Layer.ParentID to mean "no parent"); real AE
// projects start layer IDs at 1.
func (c *Composition) LayerByID(id uint32) *Layer {
	if id == 0 {
		return nil
	}
	for _, l := range c.Layers {
		if l.ID == id {
			return l
		}
	}
	return nil
}

// LayerByName returns the first layer in this composition whose Name
// matches the given string, or nil if none does. Layer names aren't
// guaranteed unique within a comp; use LayerByID for precise identity.
func (c *Composition) LayerByName(name string) *Layer {
	for _, l := range c.Layers {
		if l.Name == name {
			return l
		}
	}
	return nil
}

// ActiveCamera returns the topmost enabled (Visible == true) camera
// layer in the comp, or nil when no enabled camera is present. AE
// uses this layer's transform + camera-options for the comp's render.
// Multiple cameras may coexist; AE picks the topmost enabled one.
func (c *Composition) ActiveCamera() *Layer {
	for _, l := range c.Layers {
		if l.Type == LayerTypeCamera && l.Visible {
			return l
		}
	}
	return nil
}

// Footage represents a source media file or solid/placeholder.
type Footage struct {
	ID        uint32
	Name      string
	Path      string
	Width     uint16
	Height    uint16
	FrameRate float64
	Duration  float64
	IsStill   bool
	IsSolid   bool
	// IsPlaceholder is true when opti tag = "Plac" (AE's placeholder
	// footage — name + dimensions only, no source file). Mutually
	// exclusive with IsSolid and with having a non-empty Path.
	IsPlaceholder bool

	// Comment / Label — Item-level metadata (project-panel comment +
	// timeline color chip). Populated by parseItem from the cmta + idta
	// chunks under the Item LIST.
	Comment string
	Label   uint8

	// Underlying chunks holding the source path. Set by the parser when
	// found; SetPath mutates these for write-back.
	aliasChunk *rifx.Chunk // Pin/Als2/alas — JSON with "fullpath"
	cpthChunk  *rifx.Chunk // legacy Cpth chunk, when present
	// sspcChunk is the source-settings chunk (~222 bytes). Holds width /
	// height (already on Footage) plus audio sample rate, start/end
	// frame, footage_missing flag, etc. — used by P1 1H convenience
	// helpers (FootageMissing / HasAudio / StartFrame / EndFrame).
	// Offsets per py-aep binary/footage_chunks.py::SspcChunk.
	sspcChunk *rifx.Chunk

	// Item-level write-back references (used by SetComment / SetLabel).
	itemCmtaChunk *rifx.Chunk
	itemIdtaChunk *rifx.Chunk
	itemLayrParent *rifx.Chunk
}

// Folder is a project panel folder.
type Folder struct {
	ID   uint32
	Name string
}

// Layer represents a layer within a composition.
type Layer struct {
	Index    int
	Name     string
	Type     LayerType
	ID                uint32 // own layer ID (ldta @0x00) — referenced by ParentID of children
	ParentID          uint32 // parent layer's ID (ldta @0x84); 0 = no parent
	SourceID          uint32 // item ID of the layer's source (footage or pre-comp), per ldta@0x28
	TrackMatteLayerID uint32 // explicit matte SOURCE layer ID (ldta @0xA0, AE 23+); 0 = no explicit matte. See Layer.TrackMatteLayer() to resolve to *Layer; SetTrackMatteLayer to assign.
	LightKind         LightKind // light type (Parallel / Spot / Point / Ambient) stored at ldta @0x88; only meaningful when Type == LayerTypeLight

	StartTime float64
	Duration  float64
	Stretch   float64

	Quality              LayerQuality   // ldta @0x04
	Label                uint8          // timeline label color index (0..16) @0x3D
	BlendingMode         BlendingMode   // ldta @0x63
	PreserveTransparency bool           // ldta @0x67
	TrackMatte           TrackMatteType // ldta @0x6B
	AutoOrient           AutoOrientType // 3 bits across ldta 0x25/0x26

	// Flag bits from ldta @0x25-0x27. See parse_layer.go's decoder for
	// exact bit positions.
	Is3D                 bool
	Solo                 bool
	Shy                  bool
	Locked               bool
	Visible              bool // bit0 of 0x27 — "video enabled"
	IsAdjust             bool
	IsNull               bool
	IsGuide              bool
	MarkersLocked        bool
	MotionBlur           bool
	EffectsEnabled       bool
	AudioEnabled         bool
	FrameBlendEnabled    bool
	CollapseTransform    bool
	SamplingBicubic      bool // false = Bilinear (default), true = Bicubic
	FrameBlendPixelMotion bool // false = Frame Mix, true = Pixel Motion

	Properties      []*Property
	Effects         []*Effect
	Markers         []*Marker
	Masks           []*Mask
	ShapePaths      []*ShapePath
	ShapePrimitives []*ShapePrimitive // Rect / Ellipse / Star parametric shapes

	// TextSourceRaw holds the opaque btds payload for text layers (CoolType
	// PostScript-style serialization of font, size, color, and the actual
	// text string). Nil for non-text layers; kept verbatim for round-trip
	// writeback and downstream tools that need fields beyond TextSource.
	TextSourceRaw []byte

	// TextSource holds the decoded view of TextSourceRaw — text content,
	// fonts, per-run styling, justification. Nil for non-text layers or
	// when the btdk PostScript dict failed to parse (in which case the
	// failure is collected on Project.Warnings).
	TextSource *TextSource

	// IsShapeLayer is true when the layer has an "ADBE Root Vectors Group"
	// property tree (= it's a Shape Layer with vector primitives).
	IsShapeLayer bool

	// Comment is the user-set layer comment (AE's "Comments" timeline
	// column / Layer Settings dialog). Decoded from the sibling cmta chunk;
	// "" when no cmta is present. CRLF in the source bytes is normalized to
	// LF for Go-friendly multi-line strings.
	Comment string

	// comp is the owning composition, set by parseComposition after the
	// layer is appended. Used by Parent() to resolve ParentID to a *Layer.
	// Unexported to keep the public API minimal; nil for layers built
	// without going through the parser.
	comp *Composition

	// ldta is the underlying ldta chunk reference, captured by parseLayer.
	// Used by SetVisible/SetBlendingMode/etc. for length-preserving
	// flag-bit and byte-field writes. Nil for layers built outside the
	// parser; setters refuse with an error in that case.
	ldta *rifx.Chunk

	// nameChunk is the layer's name Utf8 chunk. Used by SetName for
	// length-variable text replacement.
	nameChunk *rifx.Chunk

	// commentChunk is the layer's cmta chunk (may be nil when no
	// comment was set). SetComment replaces its data or creates one.
	commentChunk *rifx.Chunk

	// layrList is the owning Layr LIST itself — needed when SetComment
	// has to insert a fresh cmta chunk (no existing one to mutate).
	layrList *rifx.Chunk

	// shapeRootGroup is the runtime VectorGroup tree for LayerTypeShape
	// layers. Populated by parseLayer (via hydrateShapeNodes) when a Layr
	// is parsed; lazily initialized by WrapShapeLayer on first wrap of a
	// freshly-built layer. The wrapper does NOT own this — mutations
	// persist across wrap calls and feed the Phase 4 write-time sync.
	shapeRootGroup *VectorGroup

	// shapeTransform is the runtime Layer-level Transform for shape
	// layers. Same ownership rules as shapeRootGroup.
	shapeTransform *LayerTransform

	// shapeDirty gates write-time sync (syncShapeLayerChunks). True for
	// layers built via NewShapeLayer (the lowered chunk is initially a
	// placeholder; sync must rewrite it with the user's mutations).
	// False for parser-loaded layers (the on-disk chunks ARE the source
	// of truth; re-lowering would lose content our hydrators don't yet
	// understand — nested VectorGroup, ADBE Vector Transform Group, etc).
	shapeDirty bool

	// btdsChunk is the btds LIST holding the text source bytes
	// (TextSourceRaw is an alias of this chunk's Data). Length-variable
	// text writes (per-run setters) update this chunk's Data to point
	// at a fresh splice; WriteAEP recomputes parent LIST sizes.
	btdsChunk *rifx.Chunk

	// AlternateSourceID is the AVItem id overriding this layer's source via
	// the Essential Properties → Media Replacement workflow (AE 18+). 0
	// means no override is in effect (either the layer has no Essential
	// Property slot, or it has one but the slot is unset). The slot is
	// only persisted when AE has promoted the layer's source via
	// `AVLayer.addToMotionGraphicsTemplateAs()` (see the override chunk
	// pattern in `parse_layer.go::findAlternateSourceBlsi`). Use
	// AlternateSource() to resolve to the *Composition / *Footage item.
	AlternateSourceID uint32

	// alternateSourceBlsi is the underlying blsi chunk (4-byte BE uint32
	// holding the alt source AVItem id). nil for layers without an
	// Essential Properties media-replacement slot. SetAlternateSource
	// rewrites its first 4 data bytes in place (length-preserving).
	alternateSourceBlsi *rifx.Chunk
}

// AVItem is the polymorphic "AV item" set used by Media Replacement
// (Essential Properties → Property.setAlternateSource). AE allows either
// a Composition (precomp) or a Footage to be set as an alternate source;
// Folder items are not AV and are excluded.
//
// Implemented by *Composition and *Footage.
type AVItem interface {
	ItemID() uint32
	ItemName() string
}

// ItemID returns the composition's item id (mirrors the ID field).
// Provided so *Composition satisfies AVItem.
func (c *Composition) ItemID() uint32 { return c.ID }

// ItemName returns the composition's name. Provided so *Composition
// satisfies AVItem.
func (c *Composition) ItemName() string { return c.Name }

// ItemID returns the footage's item id (mirrors the ID field). Provided
// so *Footage satisfies AVItem.
func (f *Footage) ItemID() uint32 { return f.ID }

// ItemName returns the footage's name. Provided so *Footage satisfies
// AVItem.
func (f *Footage) ItemName() string { return f.Name }

// AVItemByID returns the *Composition or *Footage with the given id,
// or nil if no AV item in the project matches (or id == 0). Folders are
// not AV items and are not consulted.
func (p *Project) AVItemByID(id uint32) AVItem {
	if id == 0 {
		return nil
	}
	for _, c := range p.Compositions {
		if c.ID == id {
			return c
		}
	}
	for _, f := range p.Footage {
		if f.ID == id {
			return f
		}
	}
	return nil
}

// Parent returns the layer's parent layer, or nil if this layer has no
// parent (ParentID == 0), the parent ID does not match any layer in the
// owning composition, or the layer was built outside the parser (no
// owning comp wired up). AE enforces same-comp parenting; cross-comp
// lookup is not performed. Returns the immediate parent only — call
// Parent() on the result to walk the chain.
func (l *Layer) Parent() *Layer {
	if l.comp == nil {
		return nil
	}
	return l.comp.LayerByID(l.ParentID)
}

// SourceComposition returns the composition this layer references as its
// source (a pre-comp layer), or nil when the source is footage / a solid /
// any non-composition item, when the SourceID does not resolve to any
// project item, or when the layer was built outside the parser (no
// owning comp/project wired up). Use Project.CompositionByID directly for
// arbitrary lookups.
func (l *Layer) SourceComposition() *Composition {
	if l.comp == nil || l.comp.proj == nil {
		return nil
	}
	return l.comp.proj.CompositionByID(l.SourceID)
}

// SourceFootage returns the footage item this layer references as its
// source (a solid / file / placeholder footage), or nil when the source
// is a composition / no source / outside the parser. Use Project.FootageByName
// or iterate Project.Footage for arbitrary lookups.
func (l *Layer) SourceFootage() *Footage {
	if l.comp == nil || l.comp.proj == nil || l.SourceID == 0 {
		return nil
	}
	for _, f := range l.comp.proj.Footage {
		if f.ID == l.SourceID {
			return f
		}
	}
	return nil
}

// TrackMatteLayer returns the layer used as this layer's track matte
// source (AE 23+ explicit pointer at ldta @0xA0). Returns nil when:
//
//   - this layer has no explicit matte source (TrackMatteLayerID == 0;
//     AE <= 22 used implicit "layer immediately above" — for that mode
//     the caller should look at the sibling at Index-1 manually);
//   - the ID does not resolve to any layer in the owning composition;
//   - the layer was built outside the parser (no comp back-pointer).
//
// Pair with TrackMatte (the matte mode at ldta @0x6B). If TrackMatte
// is None, the result of this call is meaningless even when non-nil.
func (l *Layer) TrackMatteLayer() *Layer {
	if l.comp == nil || l.TrackMatteLayerID == 0 {
		return nil
	}
	return l.comp.LayerByID(l.TrackMatteLayerID)
}

// Property match-names for the five standard Transform properties.
// AE uses these exact identifiers for every layer that has a Transform group.
const (
	MatchNameAnchorPoint = "ADBE Anchor Point"
	MatchNamePosition    = "ADBE Position"
	MatchNameScale       = "ADBE Scale"
	MatchNameRotateZ     = "ADBE Rotate Z"
	MatchNameOpacity     = "ADBE Opacity"

	// ShapeLayer-canonical 6-axis Transform stream names (per iter 2 RE
	// of tolerance.aep): AE saves ShapeLayer Position split into Position_0
	// (X) + Position_1 (Y) and always emits Orientation / Rotate X / Rotate Y
	// / Envir Appear at default. See workshop/scars/v2-2-aelayer-structure.md
	// "iter 2 新 RE 发现" for the schema.
	MatchNamePosition0   = "ADBE Position_0"
	MatchNamePosition1   = "ADBE Position_1"
	MatchNameEnvirAppear = "ADBE Envir Appear in Reflect"
)

// PropertyByMatchName returns the first property on the layer whose
// MatchName equals name, or nil if no such property exists.
func (l *Layer) PropertyByMatchName(name string) *Property {
	for _, p := range l.Properties {
		if p.MatchName == name {
			return p
		}
	}
	return nil
}

// AnchorPoint returns the layer's Anchor Point property (3D), or nil.
func (l *Layer) AnchorPoint() *Property { return l.PropertyByMatchName(MatchNameAnchorPoint) }

// Position returns the layer's Position property (3D), or nil.
func (l *Layer) Position() *Property { return l.PropertyByMatchName(MatchNamePosition) }

// Scale returns the layer's Scale property (3D), or nil.
func (l *Layer) Scale() *Property { return l.PropertyByMatchName(MatchNameScale) }

// Rotation returns the layer's Z-axis Rotation property (1D, degrees), or nil.
// On 3D layers the X- and Y-rotation properties have different match-names
// (ADBE Rotate X / ADBE Rotate Y) — use PropertyByMatchName for those.
func (l *Layer) Rotation() *Property { return l.PropertyByMatchName(MatchNameRotateZ) }

// Opacity returns the layer's Opacity property (1D, 0..1), or nil.
func (l *Layer) Opacity() *Property { return l.PropertyByMatchName(MatchNameOpacity) }

// Property represents an animatable layer property.
//
// When the property has keyframes, Keyframes is populated and StaticValue
// is nil. When it has no keyframes, StaticValue holds the constant value
// (a float64 for 1D, []float64 for multi-component).
//
// Expression, if non-empty, holds the JavaScript expression source attached
// to the property. AE evaluates the expression at runtime to override the
// keyframed/static value. Presence of an expression does not preclude
// keyframes or a static value — they coexist in the file.
type Property struct {
	MatchName         string // ADBE identifier, e.g. "ADBE Position", "ADBE Opacity"
	Name              string // display name (often empty)
	Components        int    // 1 for scalar, 2 for 2D point, 3 for 3D point, etc.
	Keyframes         []*Keyframe
	StaticValue       any
	Expression        string // JS expression source, "" when no expression set
	ExpressionEnabled bool   // tdb4 @0x78 inverted: false = AE ignores expression at render time. Always true for properties without an expression (AE's default state)

	// Write-back references — non-nil when SetValue / SetKeyframes can
	// modify the underlying RIFX bytes in-place.
	cdat       *rifx.Chunk // current/static value chunk (no keyframes)
	ldat       *rifx.Chunk // keyframe stream chunk (with keyframes)
	lhd3       *rifx.Chunk // keyframe-list header chunk (count @0x08, bpk @0x10)
	bytesPerKF int         // bytes per keyframe block within ldat.Data

	// tdbs is the property's owning tdbs LIST — used by SetExpression
	// to insert/remove the Utf8 chunk holding the JS source.
	tdbs *rifx.Chunk
	// tdb4 is the property metadata chunk under tdbs (124 bytes); holds
	// the dimension, type flags, spatial / animated / no_value / color /
	// integer / vector bits. Populated by parseLeafProperty. Used by
	// tdb4 flag readers (IsSpatial, IsAnimated, etc.).
	tdb4 *rifx.Chunk
	// tdsb is the property subprop flags chunk (4 bytes); holds
	// locked_ratio (byte 2 bit 4), dimensions_separated (byte 3 bit 1),
	// enabled (byte 3 bit 0), roto_bezier (byte 0 bit 0).
	tdsb *rifx.Chunk
	// exprChunk is the Utf8 chunk holding the expression JS source.
	// Nil when the property has no expression; SetExpression creates
	// or removes it as needed.
	exprChunk *rifx.Chunk
}

// InterpType identifies a keyframe's interpolation mode on one side
// (in-side or out-side). AE stores these as single-byte enums in the
// keyframe block at offsets 0x04 (in-interp) and 0x05 (out-interp).
type InterpType uint8

const (
	InterpLinear InterpType = 1
	InterpBezier InterpType = 2
	InterpHold   InterpType = 3
)

func (it InterpType) String() string {
	switch it {
	case InterpLinear:
		return "linear"
	case InterpBezier:
		return "bezier"
	case InterpHold:
		return "hold"
	default:
		return fmt.Sprintf("unknown(%d)", uint8(it))
	}
}

// TemporalEase is one side's temporal ease for one component:
//
//   - Speed: the value's rate of change at the keyframe, in property-value
//     units per second. Default 0 = "stops" at the keyframe ("ease in/out").
//   - Influence: how far (0..1) the bezier handle extends along the time
//     axis. Default 0.333 (~1/3) matches AE's "Easy Ease" preset.
//
// For spatial properties (Position/Anchor) and mask paths, AE stores ONE
// TemporalEase per side (speed is along the motion path, not per axis), so
// Keyframe.InTemporalEase / OutTemporalEase have length 1.
// For non-spatial properties, AE stores one TemporalEase per component, so
// the slices have length 1 (1D) or length 3 (3D, e.g. Scale).
type TemporalEase struct {
	Speed     float64
	Influence float64
}

// Keyframe represents a single keyframe on a property.
//
// Time is in seconds (decoded from the comp's TickRate).
// Value is a float64 for 1D properties or []float64 for multi-component.
//
// Interpolation and easing data:
//   - InInterp / OutInterp: linear / bezier / hold per side
//   - InSpatialTangent / OutSpatialTangent: 3D vector for spatial
//     properties (Position, Anchor Point) only; nil for non-spatial
//   - InTemporalEase / OutTemporalEase: temporal ease (speed + influence)
//     per side, per component. See TemporalEase docs for length rules.
type Keyframe struct {
	Time  float64
	Value any

	InInterp  InterpType
	OutInterp InterpType

	InSpatialTangent  []float64 // length 3 when present; nil for non-spatial
	OutSpatialTangent []float64

	InTemporalEase  []TemporalEase // length 1 for spatial+1D, length N for non-spatial N-D
	OutTemporalEase []TemporalEase

	// Write-back references — set by the parser.
	ldat     *rifx.Chunk // owning ldat chunk
	offset   int         // start of this keyframe block within ldat.Data
	dims     int         // dimensionality (mirrors Property.Components)
	tickRate float64     // owning composition's TickRate (for SetTime)
	compFps  float64     // owning composition's FrameRate (for FrameTime / SetFrameTime)
}

// ShapeLayer is the V2.2 typed wrapper around *Layer (spec §2.1). V1 callers
// keep using *Layer directly; V2.2 creation / hydration paths return
// *ShapeLayer, exposing shape-specific API (RootGroup, Transform, shorthand
// transform accessors) on top of the embedded layer.
//
// The wrapper holds runtime state only — it does NOT carry rifx.Chunk refs
// (Inv-1). Lowering (Phase 2 `lower_layer.go`) consumes the runtime tree
// and produces chunks; hydration rebuilds the runtime tree from chunks.
type ShapeLayer struct {
	*Layer // embed: V1 setters/getters + private shape state live here
}

// WrapShapeLayer wraps a parsed/created *Layer as a ShapeLayer. Caller is
// responsible for ensuring layer.Type == LayerTypeShape (matches V1 contract
// pattern: typed wrappers trust the caller). Lazily initializes the
// runtime shape state on the Layer itself so all wrappers of the same
// Layer share the same state — the wrapper is a thin façade.
func WrapShapeLayer(layer *Layer) *ShapeLayer {
	if layer.shapeRootGroup == nil {
		layer.shapeRootGroup = NewVectorGroup()
	}
	if layer.shapeTransform == nil {
		layer.shapeTransform = newLayerTransform()
	}
	// WrapShapeLayer is the V2.2 opt-in: callers signal "I'm going to use
	// V2.2 mutation APIs (RootGroup / Transform)". Mark dirty so write-time
	// sync re-lowers from the runtime tree. V1-only code paths (Property /
	// ShapePrimitives) never call WrapShapeLayer and remain unaffected.
	// Caveat: re-lowering loses on-disk content V2.2 hydration doesn't
	// preserve (V2.2-unsupported shape kinds, per-group transforms with
	// non-default values, opaque material settings).
	layer.shapeDirty = true
	return &ShapeLayer{Layer: layer}
}

// RootGroup returns the default RootGroup. Newly attached nodes go to the
// end of `RootGroup().Children` (top of render stack; spec §3.2).
func (s *ShapeLayer) RootGroup() *VectorGroup { return s.shapeRootGroup }

// Transform returns the typed Layer-level Transform surface (spec §3.3a).
func (s *ShapeLayer) Transform() *LayerTransform { return s.shapeTransform }

// Position is shorthand for s.Transform().Position(). V2.2 ShapeLayer is
// 2D-only (3D ShapeLayer = V2.3+); returns the 2D stream.
func (s *ShapeLayer) Position() *PropertyStream[[2]float64] { return s.shapeTransform.position }

// Scale is shorthand for s.Transform().Scale().
func (s *ShapeLayer) Scale() *PropertyStream[[2]float64] { return s.shapeTransform.scale }

// Rotation is shorthand for s.Transform().Rotation().
func (s *ShapeLayer) Rotation() *PropertyStream[float64] { return s.shapeTransform.rotation }

// Opacity is shorthand for s.Transform().Opacity().
func (s *ShapeLayer) Opacity() *PropertyStream[float64] { return s.shapeTransform.opacity }

// LayerTransform is the typed wrapper for a layer's Transform property
// group (spec §3.3a). V2.2 ShapeLayer is 2D, so Position / Scale /
// AnchorPoint are 2D streams; 3D layers are V2.3+.
//
// Default values (runtime-facing; lowering elides defaults per RE-S2):
//
//	AnchorPoint = [0, 0]
//	Position    = [0, 0]
//	Scale       = [100, 100]
//	Rotation    = 0
//	Opacity     = 100
type LayerTransform struct {
	anchorPoint *PropertyStream[[2]float64]
	position    *PropertyStream[[2]float64]
	scale       *PropertyStream[[2]float64]
	rotation    *PropertyStream[float64]
	opacity     *PropertyStream[float64]
}

// newLayerTransform constructs a default-valued LayerTransform.
func newLayerTransform() *LayerTransform {
	lt := &LayerTransform{
		anchorPoint: NewPropertyStream[[2]float64](),
		position:    NewPropertyStream[[2]float64](),
		scale:       NewPropertyStream[[2]float64](),
		rotation:    NewPropertyStream[float64](),
		opacity:     NewPropertyStream[float64](),
	}
	// Defaults per spec §3.6 runtime-facing convention.
	_ = lt.scale.SetStaticValue([2]float64{100, 100})
	_ = lt.opacity.SetStaticValue(100)
	return lt
}

func (t *LayerTransform) AnchorPoint() *PropertyStream[[2]float64] { return t.anchorPoint }
func (t *LayerTransform) Position() *PropertyStream[[2]float64]    { return t.position }
func (t *LayerTransform) Scale() *PropertyStream[[2]float64]       { return t.scale }
func (t *LayerTransform) Rotation() *PropertyStream[float64]       { return t.rotation }
func (t *LayerTransform) Opacity() *PropertyStream[float64]        { return t.opacity }
