package serializer

import "github.com/yueli-fx/aep-parser/internal/codec"

// Gradient holds parsed gradient color data from a gradient fill/stroke
// property ("ADBE Vector Grad Colors"). The XML is stored in the cdat chunk as
// a prop.map/prop.list/prop.pair structure.
type Gradient = codec.Gradient

// GradientColorStop represents a single color stop in a gradient.
type GradientColorStop = codec.GradientColorStop

// GradientAlphaStop represents a single alpha (opacity) stop in a gradient.
type GradientAlphaStop = codec.GradientAlphaStop

// ParseGradientXML parses AE gradient XML (prop.map format) into a *Gradient.
// Returns nil when parsing fails or the XML is empty. Exposed so tests can
// exercise the XML decoder without building a full RIFX tree.
func ParseGradientXML(xmlText string) *Gradient { return codec.ParseGradientXML(xmlText) }

// EncodeGradientXML renders a *Gradient back into AE's prop.map XML form (the
// inverse of ParseGradientXML). The byte layout — element order (Alpha Stops
// before Color Stops), the 6-float color array [offset, midpoint, r, g, b, 1],
// the 3-float alpha array [offset, midpoint, alpha], the trailing "Gradient
// Colors" = "1.0" marker, and flat newline-separated lines with no indentation
// — mirrors an AE 25.6-saved gradient fill verbatim so AE re-parses it without
// complaint.
//
// Exposed so the serializer (lowerGradientFillNode / lowerGradientStrokeNode) and tests can produce the
// Utf8 chunk payload. The output round-trips through ParseGradientXML.
func EncodeGradientXML(g *Gradient) string { return codec.EncodeGradientXML(g) }

// PropertyStream[T] is the canonical generic animation primitive. T parameterizes
// the value shape:
//
//	float64, [2]float64, [3]float64, [4]float64, BezierPath
//
// Time domain is seconds. Serializer-side `lower_property_stream.go` converts
// to ticks using `lowerCtx.tickRate` (composition tick rate).
//
// State machine: a stream starts Static (zero `static` value), flips to
// Animated on first `AddKeyframe*`, and returns to Static on `Clear()`.
// `SetStaticValue` is an error in Animated mode — callers must `Clear()` first
// if they want to revert to a static value.
type PropertyStream[T any] = codec.PropertyStream[T]

// StreamKeyframe[T] is the generic typed keyframe — distinct from the older
// non-generic `Keyframe` struct (which carries serializer back-references
// and `any` value). Named with the `Stream` prefix to avoid Go's type-name
// collision between generic and non-generic declarations in one package.
type StreamKeyframe[T any] = codec.StreamKeyframe[T]

// StreamMode is the PropertyStream state machine mode (Static ↔ Animated are
// mutually exclusive).
type StreamMode = codec.StreamMode

const (
	// StreamModeStatic — `static` field holds the current value; no keyframes.
	StreamModeStatic StreamMode = codec.StreamModeStatic
	// StreamModeAnimated — `keyframes` is authoritative; `static` invalid.
	StreamModeAnimated StreamMode = codec.StreamModeAnimated
)

// NewPropertyStream returns a Static-mode stream holding the zero value of T.
func NewPropertyStream[T any]() *PropertyStream[T] { return codec.NewPropertyStream[T]() }
