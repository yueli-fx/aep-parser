// internal/aep/property_stream.go
package codec

import "fmt"

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

// StreamMode is the PropertyStream state machine mode (Static ↔ Animated
// are mutually exclusive).
type StreamMode int

const (
	// StreamModeStatic — `static` field holds the current value; no keyframes.
	StreamModeStatic StreamMode = iota
	// StreamModeAnimated — `keyframes` is authoritative; `static` invalid.
	StreamModeAnimated
)

// PropertyStream[T] is the canonical animation primitive. T
// parameterizes the value shape:
//
//	float64, [2]float64, [3]float64, [4]float64, BezierPath
//
// Time domain is seconds. Serializer-side `lower_property_stream.go`
// converts to ticks using `lowerCtx.tickRate` (composition tick rate).
//
// State machine: a stream starts Static (zero `static` value), flips
// to Animated on first `AddKeyframe*`, and returns to Static on `Clear()`.
// `SetStaticValue` is an error in Animated mode — callers must `Clear()`
// first if they want to revert to a static value.
type PropertyStream[T any] struct {
	mode       StreamMode
	static     T
	keyframes  []StreamKeyframe[T]
	expression string // "" = no expression; reserved for V2.3+
}

// StreamKeyframe[T] is the typed keyframe — distinct from the older
// non-generic `Keyframe` struct (which carries serializer back-references
// and `any` value). Named with the `Stream` prefix to avoid Go's type-name
// collision between generic and non-generic declarations in one package.
type StreamKeyframe[T any] struct {
	Time            float64 // seconds
	Value           T
	InEase, OutEase TemporalEase // reuses V1 TemporalEase {Speed, Influence}
}

// NewPropertyStream returns a Static-mode stream holding the zero value of T.
func NewPropertyStream[T any]() *PropertyStream[T] {
	return &PropertyStream[T]{mode: StreamModeStatic}
}

// Mode reports the current state-machine mode.
func (ps *PropertyStream[T]) Mode() StreamMode { return ps.mode }

// StaticValue returns (current static value, true) in Static mode, or
// (zero, false) in Animated mode.
func (ps *PropertyStream[T]) StaticValue() (T, bool) {
	var zero T
	if ps.mode != StreamModeStatic {
		return zero, false
	}
	return ps.static, true
}

// Keyframes returns the underlying keyframe slice (not a copy). Empty in
// Static mode; non-empty in Animated mode.
func (ps *PropertyStream[T]) Keyframes() []StreamKeyframe[T] { return ps.keyframes }

// HasKeyframes reports whether the stream has at least one keyframe.
func (ps *PropertyStream[T]) HasKeyframes() bool { return len(ps.keyframes) > 0 }

// SetStaticValue updates the static value. Errors in Animated mode:
// callers must Clear() first.
func (ps *PropertyStream[T]) SetStaticValue(v T) error {
	if ps.mode == StreamModeAnimated {
		return fmt.Errorf("SetStaticValue: stream is in Animated mode; call Clear() first")
	}
	ps.static = v
	return nil
}

// AddKeyframeLinear adds a linear-interp keyframe (zero ease both sides).
// Triggers Static → Animated transition on first call.
//
// Constraints: time >= 0; duplicate `time` rejected.
func (ps *PropertyStream[T]) AddKeyframeLinear(time float64, value T) error {
	return ps.addKeyframe(StreamKeyframe[T]{Time: time, Value: value})
}

// AddKeyframeWithEase adds a keyframe with explicit in/out temporal ease.
// Triggers Static → Animated transition on first call.
//
// A side whose TemporalEase is the zero value stays linear on that side;
// an eased side requires Influence in (0, 1] (fraction of the keyframe
// interval, matching the binary encoding — AE's UI percent / 100). The
// serialized keyframe carries Bezier interpolation on each eased side.
func (ps *PropertyStream[T]) AddKeyframeWithEase(time float64, value T, in, out TemporalEase) error {
	for side, e := range map[string]TemporalEase{"in": in, "out": out} {
		if e == (TemporalEase{}) {
			continue
		}
		if e.Influence <= 0 || e.Influence > 1 {
			return fmt.Errorf("AddKeyframeWithEase: %s ease influence %g out of range (0, 1] (fraction, not percent)", side, e.Influence)
		}
	}
	return ps.addKeyframe(StreamKeyframe[T]{Time: time, Value: value, InEase: in, OutEase: out})
}

func (ps *PropertyStream[T]) addKeyframe(kf StreamKeyframe[T]) error {
	if kf.Time < 0 {
		return fmt.Errorf("AddKeyframe: time < 0 (got %g; seconds-only, non-negative)", kf.Time)
	}
	for _, existing := range ps.keyframes {
		if existing.Time == kf.Time {
			return fmt.Errorf("AddKeyframe: duplicate time %g (AE does not allow keyframes at the same time)", kf.Time)
		}
	}
	if ps.mode == StreamModeStatic {
		ps.mode = StreamModeAnimated // Static → Animated transition
	}
	ps.keyframes = append(ps.keyframes, kf)
	// Maintain ascending time order so lowering can iterate directly.
	for i := len(ps.keyframes) - 1; i > 0 && ps.keyframes[i].Time < ps.keyframes[i-1].Time; i-- {
		ps.keyframes[i], ps.keyframes[i-1] = ps.keyframes[i-1], ps.keyframes[i]
	}
	return nil
}

// Clear drops all keyframes and returns to Static mode. The retained
// static value is whatever the last SetStaticValue set (or zero if never set).
func (ps *PropertyStream[T]) Clear() error {
	ps.keyframes = nil
	ps.mode = StreamModeStatic
	return nil
}
