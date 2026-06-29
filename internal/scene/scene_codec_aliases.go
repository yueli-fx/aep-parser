package scene

import "github.com/yueli-fx/aep-parser/internal/codec"

// Codec type aliases used by scene runtime types. These mirror the public aep
// facade aliases (facade_codec.go) but live here so scene compiles importing
// only internal/codec. Both alias the same underlying codec type, so they are
// transparently identical across the package boundary.

// Gradient holds parsed gradient color data from a gradient fill/stroke
// property ("ADBE Vector Grad Colors").
type Gradient = codec.Gradient

// GradientColorStop represents a single color stop in a gradient.
type GradientColorStop = codec.GradientColorStop

// GradientAlphaStop represents a single alpha (opacity) stop in a gradient.
type GradientAlphaStop = codec.GradientAlphaStop

// PropertyStream[T] is the canonical animation primitive (time domain
// seconds). T parameterizes the value shape (float64, [N]float64, BezierPath).
type PropertyStream[T any] = codec.PropertyStream[T]
