package aep

import "github.com/example/aep-parser/internal/codec"

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

// Layer represents a layer within a composition.
type Layer struct {
	Index             int       // 1-based timeline position (top of stack = 1)
	Name              string    // layer display name (writable via SetName)
	Type              LayerType // layer kind (AV / shape / text / camera / light / null / adjustment)
	ID                uint32    // own layer ID (ldta @0x00) — referenced by ParentID of children
	ParentID          uint32    // parent layer's ID (ldta @0x84); 0 = no parent
	SourceID          uint32    // item ID of the layer's source (footage or pre-comp), per ldta@0x28
	TrackMatteLayerID uint32    // explicit matte SOURCE layer ID (ldta @0xA0, AE 23+); 0 = no explicit matte. See Layer.TrackMatteLayer() to resolve to *Layer; SetTrackMatteLayer to assign.
	LightKind         LightKind // light type (Parallel / Spot / Point / Ambient) stored at ldta @0x88; only meaningful when Type == LayerTypeLight

	StartTime float64 // in-point in seconds (timeline start of the layer)
	Duration  float64 // layer duration in seconds
	Stretch   float64 // time-stretch percentage (100 = normal speed)

	Quality              LayerQuality   // ldta @0x04
	Label                uint8          // timeline label color index (0..16) @0x3D
	BlendingMode         BlendingMode   // ldta @0x63
	PreserveTransparency bool           // ldta @0x67
	TrackMatte           TrackMatteType // ldta @0x6B
	AutoOrient           AutoOrientType // 3 bits across ldta 0x25/0x26

	// Flag bits from ldta @0x25-0x27. See parse_layer.go's decoder for
	// exact bit positions.
	Is3D                  bool
	Solo                  bool
	Shy                   bool
	Locked                bool
	Visible               bool // bit0 of 0x27 — "video enabled"
	IsAdjust              bool
	IsNull                bool
	IsGuide               bool
	MarkersLocked         bool
	MotionBlur            bool
	EffectsEnabled        bool
	AudioEnabled          bool
	FrameBlendEnabled     bool
	CollapseTransform     bool
	SamplingBicubic       bool // false = Bilinear (default), true = Bicubic
	FrameBlendPixelMotion bool // false = Frame Mix, true = Pixel Motion

	Properties      []*Property       // top-level animatable properties (Transform group, etc.)
	Effects         []*Effect         // applied effects (see effect.md)
	Markers         []*Marker         // layer markers (see marker.md)
	Masks           []*Mask           // vector masks (see mask.md)
	ShapePaths      []*ShapePath      // shape-layer Bezier paths (see shape.md)
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
	back LayerWriter

	// shapeRootGroup is the runtime VectorGroup tree for LayerTypeShape
	// layers. Populated by parseLayer (via hydrateShapeNodes) when a Layr
	// is parsed; lazily initialized by WrapShapeLayer on first wrap of a
	// freshly-built layer. The wrapper does NOT own this — mutations
	// persist across wrap calls and feed the write-time sync.
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

// layerBack returns the concrete backrefs for read-side raw chunk access
// during M8 P2 (the back field now holds the LayerWriter interface; reads
// type-assert until P3 splits scene/serializer).
func (l *Layer) layerBack() *layerBackrefs {
	if lb, ok := l.back.(*layerBackrefs); ok {
		return lb
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

	// ShapeLayer-canonical 6-axis Transform stream names: AE saves ShapeLayer
	// Position split into Position_0 (X) + Position_1 (Y) and always emits
	// Orientation / Rotate X / Rotate Y / Envir Appear at default.
	MatchNamePosition0   = "ADBE Position_0"
	MatchNamePosition1   = "ADBE Position_1"
	MatchNamePosition2   = "ADBE Position_2"
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

// ShapeLayer is the V2.2 typed wrapper around *Layer. V1 callers
// keep using *Layer directly; V2.2 creation / hydration paths return
// *ShapeLayer, exposing shape-specific API (RootGroup, Transform, shorthand
// transform accessors) on top of the embedded layer.
//
// The wrapper holds runtime state only — it does NOT carry rifx.Chunk refs.
// Lowering (`lower_layer.go`) consumes the runtime tree
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
// end of `RootGroup().Children` (top of render stack).
func (s *ShapeLayer) RootGroup() *VectorGroup { return s.shapeRootGroup }

// Transform returns the typed Layer-level Transform surface.
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
// group. V2.2 ShapeLayer is 2D, so Position / Scale /
// AnchorPoint are 2D streams; 3D layers are V2.3+.
//
// Default values (runtime-facing; lowering elides defaults):
//
//	AnchorPoint = [0, 0]
//	Position    = [0, 0]
//	Scale       = [100, 100]
//	Rotation    = 0
//	Opacity     = 100
type LayerTransform struct {
	anchorPoint *codec.PropertyStream[[2]float64]
	position    *codec.PropertyStream[[2]float64]
	scale       *codec.PropertyStream[[2]float64]
	rotation    *codec.PropertyStream[float64]
	opacity     *codec.PropertyStream[float64]
}

// newLayerTransform constructs a default-valued LayerTransform.
func newLayerTransform() *LayerTransform {
	lt := &LayerTransform{
		anchorPoint: codec.NewPropertyStream[[2]float64](),
		position:    codec.NewPropertyStream[[2]float64](),
		scale:       codec.NewPropertyStream[[2]float64](),
		rotation:    codec.NewPropertyStream[float64](),
		opacity:     codec.NewPropertyStream[float64](),
	}
	// Defaults per runtime-facing convention.
	_ = lt.scale.SetStaticValue([2]float64{100, 100})
	_ = lt.opacity.SetStaticValue(100)
	return lt
}

func (t *LayerTransform) AnchorPoint() *PropertyStream[[2]float64] { return t.anchorPoint }
func (t *LayerTransform) Position() *PropertyStream[[2]float64]    { return t.position }
func (t *LayerTransform) Scale() *PropertyStream[[2]float64]       { return t.scale }
func (t *LayerTransform) Rotation() *PropertyStream[float64]       { return t.rotation }
func (t *LayerTransform) Opacity() *PropertyStream[float64]        { return t.opacity }
