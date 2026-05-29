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

// FeetFramesFilmType represents the film type for feet+frames timecode display.
type FeetFramesFilmType uint8

const (
	FeetFramesFilmTypeMM35 FeetFramesFilmType = 0 // 35mm film
	FeetFramesFilmTypeMM16 FeetFramesFilmType = 1 // 16mm film
)

func (f FeetFramesFilmType) String() string {
	switch f {
	case FeetFramesFilmTypeMM35:
		return "MM35 (35mm)"
	case FeetFramesFilmTypeMM16:
		return "MM16 (16mm)"
	default:
		return fmt.Sprintf("unknown(%d)", f)
	}
}

// FootageTimecodeDisplayStartType represents how timecode is displayed for footage.
type FootageTimecodeDisplayStartType uint8

const (
	FootageTimecodeDisplayStartTypeStart0 FootageTimecodeDisplayStartType = 0 // Start at 0
	FootageTimecodeDisplayStartTypeUseSourceMedia FootageTimecodeDisplayStartType = 1 // Use source media
)

func (f FootageTimecodeDisplayStartType) String() string {
	switch f {
	case FootageTimecodeDisplayStartTypeStart0:
		return "Start at 0"
	case FootageTimecodeDisplayStartTypeUseSourceMedia:
		return "Use source media"
	default:
		return fmt.Sprintf("unknown(%d)", f)
	}
}

// FramesCountType represents how frames are counted in the project.
type FramesCountType uint8

const (
	FramesCountTypeStart0 FramesCountType = 0 // Start at 0
	FramesCountTypeStart1 FramesCountType = 1 // Start at 1
	FramesCountTypeTimecodeConversion FramesCountType = 2 // Timecode conversion
)

func (f FramesCountType) String() string {
	switch f {
	case FramesCountTypeStart0:
		return "Start at 0"
	case FramesCountTypeStart1:
		return "Start at 1"
	case FramesCountTypeTimecodeConversion:
		return "Timecode conversion"
	default:
		return fmt.Sprintf("unknown(%d)", f)
	}
}

// TimeDisplayType represents how time is displayed in the project.
type TimeDisplayType uint8

const (
	TimeDisplayTypeTimecode TimeDisplayType = 0 // Timecode
	TimeDisplayTypeFrames TimeDisplayType = 1   // Frames
)

func (t TimeDisplayType) String() string {
	switch t {
	case TimeDisplayTypeTimecode:
		return "Timecode"
	case TimeDisplayTypeFrames:
		return "Frames"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

// ColorManagementSystem represents the color management system used by the project.
type ColorManagementSystem uint8

const (
	ColorManagementSystemAdobe ColorManagementSystem = 0 // Adobe color management
	ColorManagementSystemOCIO  ColorManagementSystem = 1 // OCIO color management
)

func (c ColorManagementSystem) String() string {
	switch c {
	case ColorManagementSystemAdobe:
		return "Adobe"
	case ColorManagementSystemOCIO:
		return "OCIO"
	default:
		return fmt.Sprintf("unknown(%d)", c)
	}
}

// LutInterpolationMethod represents the LUT interpolation method for the project.
type LutInterpolationMethod uint8

const (
	LutInterpolationMethodTrilinear    LutInterpolationMethod = 0 // Trilinear
	LutInterpolationMethodTetrahedral  LutInterpolationMethod = 1 // Tetrahedral
)

func (l LutInterpolationMethod) String() string {
	switch l {
	case LutInterpolationMethodTrilinear:
		return "Trilinear"
	case LutInterpolationMethodTetrahedral:
		return "Tetrahedral"
	default:
		return fmt.Sprintf("unknown(%d)", l)
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

	// back holds the underlying RIFX root + project-level single-field chunk
	// refs that power length-preserving writes. Nil for projects built outside
	// the parser. See back_project.go.
	back *projectBackrefs

	// V2: derived state for structural mutation (NewComposition / 未来 NewFootage etc.)
	nextItemID uint32   // monotonic Item ID counter; never reused (see Invariants #9)
	target     AETarget // which AE-version template NewProject loaded; drives per-target builder chunk selection
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

	// Item-level metadata shared with Footage / Folder. Populated by
	// parseItem from the surrounding Item LIST (cmta child + idta byte).
	Comment string // Item.comment — AE's project-panel comment column
	Label   uint8  // Item.label — project-panel color index 0..16 (idta @0x3A)

	// back holds the underlying RIFX chunk refs that power length-preserving
	// writes. Nil for comps built outside the parser. See back_composition.go.
	back *compositionBackrefs
}

// CdtaRawBytes returns the comp's cdta chunk Data slice, or nil if the
// comp has no cdta. Read-only access for debugging / RE tools — the
// underlying byte slice is the live chunk data; do not mutate.
func (c *Composition) CdtaRawBytes() []byte {
	if c.back == nil || c.back.cdta == nil {
		return nil
	}
	return c.back.cdta.Data
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

	// back holds the underlying RIFX chunk refs that power length-preserving
	// writes. Nil for footage items built outside the parser. See back_footage.go.
	back *footageBackrefs
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

	// back holds the underlying RIFX chunk refs that power length-preserving
	// writes. Nil for layers built outside the parser. See back_layer.go.
	back *layerBackrefs

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

	// AlternateSourceID is the AVItem id overriding this layer's source via
	// the Essential Properties → Media Replacement workflow (AE 18+). 0
	// means no override is in effect (either the layer has no Essential
	// Property slot, or it has one but the slot is unset). The slot is
	// only persisted when AE has promoted the layer's source via
	// `AVLayer.addToMotionGraphicsTemplateAs()` (see the override chunk
	// pattern in `parse_layer.go::findAlternateSourceBlsi`). Use
	// AlternateSource() to resolve to the *Composition / *Footage item.
	AlternateSourceID uint32

	// propertyTree is the hierarchical mirror of the layer's tdgp property
	// tree (P2c PropertyGroup hierarchy). Built by buildPropertyGroupTree
	// alongside the flat Layer.Properties slice; nil for layers built
	// outside the parser. See AEPropertyGroup in property_group.go.
	propertyTree *AEPropertyGroup
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
	// / Envir Appear at default. See flightdeck/incident-reports/v2-2-aelayer-structure.md
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

	// DefaultValue is the property's default value (what AE considers
	// the "unmodified" state). For transform properties, set from
	// hardcoded tables during parse; for effect parameters, from pard
	// chunks. nil when unknown (non-transform, non-effect properties).
	DefaultValue any
	// LastValue is the property's last-set value from the pard chunk
	// (effect parameters only). nil for non-effect properties.
	LastValue any
	// NbOptions is the number of options for enum/dropdown effect
	// parameters (pard chunk). 0 for non-enum properties.
	NbOptions int

	// Gradient holds the parsed gradient data for "ADBE Vector Grad
	// Colors" properties. The XML is stored in the cdat chunk and
	// parsed during property initialization. nil for non-gradient
	// properties.
	Gradient *Gradient

	// back holds the underlying RIFX chunk refs that power length-preserving
	// writes. nil for properties built outside the parser. See back_property.go.
	back *propertyBackrefs

	// parentTreeGroup is the AEPropertyGroup that contains this leaf in
	// the layer's hierarchical property tree (P2c). Populated by
	// wirePropertyTreeLeaves; nil for properties built outside the
	// parser or not reachable from the tree (e.g. effect parameters,
	// mask sub-properties — those live in Effect.Parameters / Mask, not
	// in the layer's top-level tdgp tree).
	parentTreeGroup *AEPropertyGroup
}

// PropertyControlType identifies the UI control type for a property
// (scalar slider, color picker, angle dial, checkbox, dropdown, etc.).
// Derived from tdb4 flags; mirrors py-aep PropertyControlType enum.
type PropertyControlType uint8

const (
	PCTLLayer      PropertyControlType = 0
	PCTLInteger    PropertyControlType = 1
	PCTLScalar     PropertyControlType = 2
	PCTLAngle      PropertyControlType = 3
	PCTLBoolean    PropertyControlType = 4
	PCTLColor      PropertyControlType = 5
	PCTLTwoD       PropertyControlType = 6
	PCTLEnum       PropertyControlType = 7
	PCTLPaintGroup PropertyControlType = 9
	PCTLSlider     PropertyControlType = 10
	PCTLCurve      PropertyControlType = 11
	PCTLMask       PropertyControlType = 12
	PCTLGroup      PropertyControlType = 13
	PCTLThreeD     PropertyControlType = 18
	PCTLUnknown    PropertyControlType = 15
)

func (p PropertyControlType) String() string {
	switch p {
	case PCTLLayer:
		return "layer"
	case PCTLInteger:
		return "integer"
	case PCTLScalar:
		return "scalar"
	case PCTLAngle:
		return "angle"
	case PCTLBoolean:
		return "boolean"
	case PCTLColor:
		return "color"
	case PCTLTwoD:
		return "two_d"
	case PCTLEnum:
		return "enum"
	case PCTLPaintGroup:
		return "paint_group"
	case PCTLSlider:
		return "slider"
	case PCTLCurve:
		return "curve"
	case PCTLMask:
		return "mask"
	case PCTLGroup:
		return "group"
	case PCTLThreeD:
		return "three_d"
	default:
		return fmt.Sprintf("unknown(%d)", uint8(p))
	}
}

// PropertyValueType identifies the type of value stored in a property.
// Mirrors py-aep PropertyValueType enum (ExtendScript constants).
type PropertyValueType uint16

const (
	PVTUnknown       PropertyValueType = 0
	PVTNoValue       PropertyValueType = 6412
	PVTThreeDSpatial PropertyValueType = 6413
	PVTThreeD        PropertyValueType = 6414
	PVTTwoDSpatial   PropertyValueType = 6415
	PVTTwoD          PropertyValueType = 6416
	PVTOneD          PropertyValueType = 6417
	PVTColor         PropertyValueType = 6418
)

func (p PropertyValueType) String() string {
	switch p {
	case PVTUnknown:
		return "unknown"
	case PVTNoValue:
		return "no_value"
	case PVTThreeDSpatial:
		return "three_d_spatial"
	case PVTThreeD:
		return "three_d"
	case PVTTwoDSpatial:
		return "two_d_spatial"
	case PVTTwoD:
		return "two_d"
	case PVTOneD:
		return "one_d"
	case PVTColor:
		return "color"
	default:
		return fmt.Sprintf("unknown(%d)", uint16(p))
	}
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

	// back holds the underlying RIFX chunk ref + cached layout metadata that
	// power length-preserving keyframe writes. Nil for keyframes built outside
	// the parser. See back_keyframe.go.
	back *keyframeBackrefs
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

// AnchorPoint is shorthand for s.Transform().AnchorPoint().
func (s *ShapeLayer) AnchorPoint() *PropertyStream[[2]float64] {
	return s.shapeTransform.anchorPoint
}

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
