package scene

import (
	"fmt"

	"github.com/example/aep-parser/internal/codec"
)

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
	ExpressionEnabled bool   // tdb4 @0x77 disabled byte (0 = AE evaluates; @0x78 is the has-expression marker). Always true for properties without an expression (AE's default state)

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

	// back holds the writer interface for length-preserving / expression
	// setters. Concrete chunk access (keyframe-stream ops, separate-dimensions,
	// flag readers, parse) goes through propertyBack. nil for properties built
	// outside the parser. See back_property.go.
	back PropertyWriter

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
//   - Influence: how far (0..1) the bezier handle extends along the time axis.
//     Default 0.333 (~1/3) matches AE's "Easy Ease" preset.
//
// For spatial properties (Position/Anchor) and mask paths, AE stores ONE
// TemporalEase per side (speed is along the motion path, not per axis),
// so Keyframe.InTemporalEase / OutTemporalEase have length 1. For non-spatial
// properties, AE stores one TemporalEase per component, so the slices have
// length 1 (1D) or length 3 (3D, e.g. Scale).
type TemporalEase = codec.TemporalEase

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
	back KeyframeWriter
}
