// internal/aep/property_stream.go
package aep

import "fmt"

// StreamMode is the PropertyStream state machine mode (Inv-8: Static ↔
// Animated mutually exclusive).
type StreamMode int

const (
	// StreamModeStatic — `static` field holds the current value; no keyframes.
	StreamModeStatic StreamMode = iota
	// StreamModeAnimated — `keyframes` is authoritative; `static` invalid.
	StreamModeAnimated
)

// PropertyStream[T] is the canonical V2.2 animation primitive (Inv-4 /
// spec §2.4). T parameterizes the value shape:
//
//	float64, [2]float64, [3]float64, [4]float64, BezierPath
//
// Time domain is seconds (Inv-7). Serializer-side `lower_property_stream.go`
// converts to ticks using `lowerCtx.tickRate` (composition tick rate).
//
// State machine (Inv-8): a stream starts Static (zero `static` value), flips
// to Animated on first `AddKeyframe*`, and returns to Static on `Clear()`.
// `SetStaticValue` is an error in Animated mode — callers must `Clear()`
// first if they want to revert to a static value.
type PropertyStream[T any] struct {
	mode       StreamMode
	static     T
	keyframes  []StreamKeyframe[T]
	expression string // "" = no expression; reserved for V2.3+
}

// StreamKeyframe[T] is the V2.2 typed keyframe — distinct from the V1
// non-generic `Keyframe` struct (which carries serializer back-references
// and `any` value). Named with the `Stream` prefix to avoid Go's type-name
// collision between generic and non-generic declarations in one package.
type StreamKeyframe[T any] struct {
	Time            float64 // seconds (Inv-7)
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

// SetStaticValue updates the static value. Errors in Animated mode (Inv-8):
// callers must Clear() first.
func (ps *PropertyStream[T]) SetStaticValue(v T) error {
	if ps.mode == StreamModeAnimated {
		return fmt.Errorf("SetStaticValue: stream is in Animated mode; call Clear() first (Inv-8)")
	}
	ps.static = v
	return nil
}

// AddKeyframeLinear adds a linear-interp keyframe (zero ease both sides).
// Triggers Static → Animated transition on first call (Inv-8).
//
// Constraints: time >= 0 (Inv-7); duplicate `time` rejected.
func (ps *PropertyStream[T]) AddKeyframeLinear(time float64, value T) error {
	return ps.addKeyframe(StreamKeyframe[T]{Time: time, Value: value})
}

// AddKeyframeWithEase adds a keyframe with explicit in/out temporal ease.
// Triggers Static → Animated transition on first call (Inv-8).
func (ps *PropertyStream[T]) AddKeyframeWithEase(time float64, value T, in, out TemporalEase) error {
	return ps.addKeyframe(StreamKeyframe[T]{Time: time, Value: value, InEase: in, OutEase: out})
}

func (ps *PropertyStream[T]) addKeyframe(kf StreamKeyframe[T]) error {
	if kf.Time < 0 {
		return fmt.Errorf("AddKeyframe: time < 0 (got %g; Inv-7: seconds-only, non-negative)", kf.Time)
	}
	for _, existing := range ps.keyframes {
		if existing.Time == kf.Time {
			return fmt.Errorf("AddKeyframe: duplicate time %g (AE does not allow keyframes at the same time)", kf.Time)
		}
	}
	if ps.mode == StreamModeStatic {
		ps.mode = StreamModeAnimated // Inv-8 transition
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
