package aep

import (
	"fmt"
)

// Effect represents one effect instance applied to a layer (e.g. a Gaussian
// Blur applied to a footage layer). The MatchName identifies the effect
// type ("ADBE Gaussian Blur 2", "ADBE Gradient Wipe", etc.) and Parameters
// holds each tweakable knob as a Property. Parameter match-names are
// internal numeric suffixes like "ADBE Gaussian Blur 2-0001".
type Effect struct {
	MatchName  string      // effect type, e.g. "ADBE Gaussian Blur 2"
	Name       string      // user-set display name (currently mirrors MatchName)
	Parameters []*Property // effect knobs (Blurriness, Completion, …)
}

// Marker represents a single timeline marker on a layer (or a comp).
// Times are in seconds. The remaining string fields hold the user-
// editable annotation data (most are empty unless the user filled them
// in inside AE's "Marker Dialog").
type Marker struct {
	Time         float64 // marker position in seconds, decoded via the owning composition's TickRate
	Duration     float64 // seconds; 0 = point marker. Decoded from NmHd @0x08 / 600.
	Label        uint8   // timeline label color index (0..16); 0 = default. NmHd @0x10.
	Comment      string  // first Utf8 in the Nmrd block
	Chapter      string  // second Utf8 — chapter link
	URL          string  // third Utf8 — web target
	FrameTarget  string  // fourth Utf8 — frame target id
	CuePointName string  // fifth Utf8 — cue-point name (if used)

	// Write-back state. Populated by parseMarkers; zero/nil for markers
	// built outside the parser. ldatOffset / tickRate / compFps are plain
	// values; the chunk references live in back (see back_marker.go) so the
	// scene type stays chunk-free. SetTime / SetDuration / SetLabel /
	// SetComment refuse with an error when the reference they need is missing.
	ldatOffset int     // start of this marker's keyframe block in back.ldat.Data
	tickRate   float64 // owning composition's TickRate (for SetTime)
	compFps    float64 // owning composition's FrameRate (for FrameTime / SetFrameTime)

	back MarkerWriter // per-marker chunk references (nil outside parser)

	// list back-references the owning marker-set container, shared by every
	// marker in the same "ADBE Marker" set. Populated by parseMarkers; nil for
	// markers built outside the parser. Required by the structural ops
	// (Marker.Remove / Composition.AddMarker).
	list *markerList
}

// MaskMode is the compositing mode for a mask (matches AE C++ SDK values).
type MaskMode uint32

const (
	MaskModeNone       MaskMode = 0
	MaskModeAdd        MaskMode = 1
	MaskModeSubtract   MaskMode = 2
	MaskModeIntersect  MaskMode = 3
	MaskModeLighten    MaskMode = 4
	MaskModeDarken     MaskMode = 5
	MaskModeDifference MaskMode = 6
)

func (m MaskMode) String() string {
	switch m {
	case MaskModeNone:
		return "none"
	case MaskModeAdd:
		return "add"
	case MaskModeSubtract:
		return "subtract"
	case MaskModeIntersect:
		return "intersect"
	case MaskModeLighten:
		return "lighten"
	case MaskModeDarken:
		return "darken"
	case MaskModeDifference:
		return "difference"
	default:
		return fmt.Sprintf("unknown(%d)", uint32(m))
	}
}

// Mask represents a single vector mask on a layer. Each mask has a closed
// or open path made of Bezier vertices.
//
// Vertices holds the (first) path snapshot. For STATIC masks this is the
// only path. For ANIMATED masks (mask path keyframed in AE), every
// snapshot is in PathKeyframes; Vertices mirrors PathKeyframes[0].Vertices
// for convenience.
//
// mkif (48 bytes) decoded fields:
//   - Mode (uint32 BE @0x04)
//   - Inverted (uint8 @0x00)
//   - Index (uint32 BE @0x08) — 1-based mask index on the layer
//   - Color [3]uint8 (R/G/B @0x2D-0x2F; alpha at 0x2C is always 0xFF) —
//     the colored label shown next to the mask in AE's timeline.
//
// MkifRaw is kept around for round-trip preservation of any bytes we
// haven't decoded.
type Mask struct {
	Name          string             // from omtn ("" when unnamed)
	Closed        bool               // from shph[0x14] == 0x01
	Mode          MaskMode           // from mkif @0x04
	Inverted      bool               // from mkif @0x00
	Locked        bool               // from mkif @0x01 (== 1 when AE timeline UI lock is on)
	MotionBlur    MaskMotionBlurMode // from mkif @0x02 (0=SameAsLayer, 2=On, 3=Off)
	Index         uint32             // from mkif @0x08 (AE internal mask ID, not always sequential)
	Color         [3]uint8           // R, G, B from mkif @0x2D-0x2F (timeline label color)
	Vertices      []MaskVertex       // first snapshot (or only one when static)
	PathKeyframes []MaskPathKeyframe // nil when path is not animated

	// Other mask-level properties, populated when their cdat values exist
	// in the mask atom's property tree. Feather is 2D (X, Y) in pixels;
	// Opacity is 0..1; Expansion (AE's "Mask Expansion", internally
	// "ADBE Mask Offset") is in pixels.
	Feather   [2]float64
	Opacity   float64 // default 1.0 when no cdat present
	Expansion float64

	// Properties is every other tdmn+tdbs leaf inside the mask atom's
	// property group (anything beyond Shape/Feather/Opacity/Offset, or those
	// when keyframed). Useful for keyframed mask properties (e.g. animated
	// feather).
	Properties []*Property

	MkifRaw []byte // 48 bytes raw mask info (preserved for write-back)
	ShphRaw []byte // 24 bytes, path header

	// back holds the chunk references behind SetMode / SetInverted / SetColor
	// / SetClosed etc. (see back_mask.go). nil for masks built outside the
	// parser; the setters refuse in that case.
	back MaskWriter
}

// MaskPathKeyframe is one keyframe of an animated mask path. Time is in
// seconds. Vertices is the full path snapshot at that time. Easing follows
// the same scalar (one TemporalEase per side) layout as spatial keyframes.
type MaskPathKeyframe struct {
	Time            float64
	Vertices        []MaskVertex
	InInterp        InterpType
	OutInterp       InterpType
	InTemporalEase  TemporalEase
	OutTemporalEase TemporalEase
}

// MaskVertex is one Bezier control point along a mask path. Coordinates
// are stored as ABSOLUTE positions (not offsets from the anchor):
//
//   - Anchor: the vertex position
//   - InTangent: incoming Bezier control point. For a straight incoming
//     segment InTangent equals Anchor (control coincides with vertex →
//     degenerate Bezier = straight line).
//   - OutTangent: outgoing Bezier control point. For a straight outgoing
//     segment OutTangent equals the NEXT vertex's Anchor (the control
//     points sit on the line between the two anchors → straight line).
//
// So an all-straight closed polygon has InTangent==Anchor for every vertex
// and OutTangent==anchorOf(next vertex). Curved segments have tangents
// offset away from the anchors.
type MaskVertex struct {
	Anchor     [2]float64
	InTangent  [2]float64
	OutTangent [2]float64
}

// BlendingMode is the layer's compositing blend mode (ldta @0x63).
// Values match py-aep's BlendingMode enum / AE's internal numbering.
type BlendingMode uint8

const (
	BlendingModeNormalCamera      BlendingMode = 0 // null/camera/light default
	BlendingModeNormal            BlendingMode = 2
	BlendingModeDissolve          BlendingMode = 3
	BlendingModeAdd               BlendingMode = 4
	BlendingModeMultiply          BlendingMode = 5
	BlendingModeScreen            BlendingMode = 6
	BlendingModeOverlay           BlendingMode = 7
	BlendingModeSoftLight         BlendingMode = 8
	BlendingModeHardLight         BlendingMode = 9
	BlendingModeDarken            BlendingMode = 10
	BlendingModeLighten           BlendingMode = 11
	BlendingModeClassicDifference BlendingMode = 12
	BlendingModeHue               BlendingMode = 13
	BlendingModeSaturation        BlendingMode = 14
	BlendingModeColor             BlendingMode = 15
	BlendingModeLuminosity        BlendingMode = 16
	BlendingModeStencilAlpha      BlendingMode = 17
	BlendingModeStencilLuma       BlendingMode = 18
	BlendingModeSilhouetteAlpha   BlendingMode = 19
	BlendingModeSilhouetteLuma    BlendingMode = 20
	BlendingModeLuminescentPremul BlendingMode = 21
	BlendingModeAlphaAdd          BlendingMode = 22
	BlendingModeClassicColorDodge BlendingMode = 23
	BlendingModeClassicColorBurn  BlendingMode = 24
	BlendingModeExclusion         BlendingMode = 25
	BlendingModeDifference        BlendingMode = 26
	BlendingModeColorDodge        BlendingMode = 27
	BlendingModeColorBurn         BlendingMode = 28
	BlendingModeLinearDodge       BlendingMode = 29
	BlendingModeLinearBurn        BlendingMode = 30
	BlendingModeLinearLight       BlendingMode = 31
	BlendingModeVividLight        BlendingMode = 32
	BlendingModePinLight          BlendingMode = 33
	BlendingModeHardMix           BlendingMode = 34
	BlendingModeLighterColor      BlendingMode = 35
	BlendingModeDarkerColor       BlendingMode = 36
	BlendingModeSubtract          BlendingMode = 37
	BlendingModeDivide            BlendingMode = 38
)

// TrackMatteType is the layer's track-matte mode (ldta @0x6B).
type TrackMatteType uint8

const (
	TrackMatteNone         TrackMatteType = 0
	TrackMatteAlpha        TrackMatteType = 1
	TrackMatteAlphaInverse TrackMatteType = 2
	TrackMatteLuma         TrackMatteType = 3
	TrackMatteLumaInverse  TrackMatteType = 4
)

// AutoOrientType is the layer's auto-orientation mode (Layer > Transform >
// Auto-Orient in AE). Encoded as 3 mutually-exclusive bits across ldta
// bytes 0x25 and 0x26 — see parse_layer.go's flag-bit doc. AE collapses
// the bits into one of four UI choices, which we mirror here.
type AutoOrientType uint8

const (
	AutoOrientNone                    AutoOrientType = 0
	AutoOrientAlongPath               AutoOrientType = 1 // motion along its Position path
	AutoOrientCameraOrPointOfInterest AutoOrientType = 2 // face camera / its POI
	AutoOrientCharactersTowardCamera  AutoOrientType = 3 // 3D text per-character billboard
)

func (a AutoOrientType) String() string {
	switch a {
	case AutoOrientNone:
		return "none"
	case AutoOrientAlongPath:
		return "along-path"
	case AutoOrientCameraOrPointOfInterest:
		return "camera-or-point-of-interest"
	case AutoOrientCharactersTowardCamera:
		return "characters-toward-camera"
	default:
		return fmt.Sprintf("unknown(%d)", uint8(a))
	}
}

// LayerQuality is the layer's render quality (ldta @0x04, uint16).
type LayerQuality uint16

const (
	LayerQualityWireframe LayerQuality = 0
	LayerQualityDraft     LayerQuality = 1
	LayerQualityBest      LayerQuality = 2
)

// ShapePath is one custom Bezier path on a Shape Layer (the result of
// drawing with the Pen tool inside an "ADBE Vector Shape - Group"). It
// reuses MaskVertex for the per-vertex Anchor + InTangent + OutTangent
// triplet; the storage format is identical to mask paths.
//
// Parametric shapes (Rect, Ellipse, Star) get their own ShapePrimitive
// entries — see Layer.ShapePrimitives().
// ShapePath is a freeform Bezier path inside a shape layer. Vertices reuse the
// mask-vertex model (absolute coordinates; see MaskVertex). Editing the
// geometry of a parsed path is not yet supported.
type ShapePath struct {
	Name     string       // from omtn ("" when unnamed)
	Closed   bool         // true = closed path; false = open
	Vertices []MaskVertex // path control points (absolute coords; see MaskVertex)
	ShphRaw  []byte       // raw shph path-header bytes (preserved for write-back)
}

// ShapePrimitiveKind identifies which AE parametric primitive a
// ShapePrimitive represents. The values are stable strings rather than
// integers so they round-trip cleanly through JSON.
type ShapePrimitiveKind string

const (
	ShapePrimitiveRect    ShapePrimitiveKind = "rect"
	ShapePrimitiveEllipse ShapePrimitiveKind = "ellipse"
	ShapePrimitiveStar    ShapePrimitiveKind = "star"
)

// ShapePrimitive is one parametric primitive (Rectangle / Ellipse /
// Star Polygon) inside a Shape Layer's vector tree. Sub-property
// fields are nil when AE didn't write a cdat for them (typically
// happens for defaulted values like Rect Position [0, 0]).
//
// Each primitive sits inside a Vector Group; multiple primitives can
// coexist in one shape layer. Each `*Property` is a regular Property
// (same chunk references as the flat Layer.Properties entry, so
// SetStaticValue / Keyframe.SetX work the same way).
type ShapePrimitive struct {
	Kind      ShapePrimitiveKind
	GroupName string // owning Vector Group's display name ("RectGroup" etc.); empty when unnamed

	// Common fields (per-kind). Nil for primitives that don't carry the
	// concept (e.g. Star.Size doesn't exist; Star has Inner/OuterRadius).
	Size      *Property // Rect / Ellipse: 2D [w, h]
	Position  *Property // 2D [x, y] — local offset within the owning group
	Roundness *Property // Rect: 1D corner radius

	// Star-only fields.
	StarType       *Property // 1D enum (1=Star, 2=Polygon)
	Points         *Property // 1D integer-valued
	Rotation       *Property // 1D degrees
	InnerRadius    *Property
	OuterRadius    *Property
	InnerRoundness *Property // AE's own internal name uses "Roundess" (typo); we expose the corrected spelling
	OuterRoundness *Property
}
