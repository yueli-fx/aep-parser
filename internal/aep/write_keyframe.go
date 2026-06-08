package aep

import (
	"fmt"
	"math"
)

// Keyframe writers. All edits are length-preserving block writes
// into the parent ldat. Layout dispatch (spatial vs non-spatial)
// is owned by keyframeBackrefs.

// SetTime rewrites this keyframe's time in-place using its owning
// composition's TickRate (decoded at parse time from cdta). Chunk size
// does not change, so writing the project back never relocates anything.
func (k *Keyframe) SetTime(seconds float64) error {
	if k.back == nil {
		return fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	if err := k.back.SetTime(seconds); err != nil {
		return err
	}
	if kb, ok := k.back.(*keyframeBackrefs); ok {
		rate := kb.tickRate
		if rate == 0 {
			rate = aeLegacyTimeBase
		}
		ticks := uint32(math.Round(seconds * rate))
		k.Time = float64(ticks) / rate
	}
	return nil
}

// SetValue rewrites this keyframe's numeric value in-place.
//
// For a 1D property pass a float64. For a multi-component property pass
// []float64 of exact length equal to the property's component count.
// Length-preserving: chunk size is unchanged.
func (k *Keyframe) SetValue(v any) error {
	if k.back == nil {
		return fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	if err := k.back.SetValue(v); err != nil {
		return err
	}
	switch x := v.(type) {
	case float64:
		k.Value = x
	case []float64:
		k.Value = append([]float64(nil), x...)
	}
	return nil
}

// SetInInterp / SetOutInterp write a new interpolation enum byte to
// the keyframe block @0x04 / @0x05.
// length-preserving (1 byte).
func (k *Keyframe) SetInInterp(t InterpType) error {
	if k.back == nil {
		return fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	if err := k.back.SetInInterp(t); err != nil {
		return err
	}
	k.InInterp = t
	return nil
}

// SetOutInterp see SetInInterp.
func (k *Keyframe) SetOutInterp(t InterpType) error {
	if k.back == nil {
		return fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	if err := k.back.SetOutInterp(t); err != nil {
		return err
	}
	k.OutInterp = t
	return nil
}

// SetInTemporalEase writes a new per-side temporal ease list to the
// keyframe. The slice length must match the keyframe's current ease
// shape:
//
//   - spatial property (Position / Anchor) or 1D non-spatial: length 1
//   - non-spatial N-D (Scale 3D / Feather 2D / 4D color): length N
//
// length-preserving (8 or 16 bytes per ease, fixed offsets per layout).
func (k *Keyframe) SetInTemporalEase(eases []TemporalEase) error {
	if k.back == nil {
		return fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	if err := k.back.SetInTemporalEase(eases); err != nil {
		return err
	}
	k.InTemporalEase = append(k.InTemporalEase[:0], eases...)
	return nil
}

// SetOutTemporalEase mirrors SetInTemporalEase for the out side.
func (k *Keyframe) SetOutTemporalEase(eases []TemporalEase) error {
	if k.back == nil {
		return fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	if err := k.back.SetOutTemporalEase(eases); err != nil {
		return err
	}
	k.OutTemporalEase = append(k.OutTemporalEase[:0], eases...)
	return nil
}

// SetInSpatialTangent / SetOutSpatialTangent write 3D Bezier tangent
// vectors. Only valid for spatial properties (Position / AnchorPoint
// with motion path enabled). Slice length must equal `Property.Components`.
// length-preserving (dims × 8 bytes).
func (k *Keyframe) SetInSpatialTangent(v []float64) error {
	if k.back == nil {
		return fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	if err := k.back.SetInSpatialTangent(v); err != nil {
		return err
	}
	k.InSpatialTangent = append(k.InSpatialTangent[:0], v...)
	return nil
}

// SetOutSpatialTangent see SetInSpatialTangent.
func (k *Keyframe) SetOutSpatialTangent(v []float64) error {
	if k.back == nil {
		return fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	if err := k.back.SetOutSpatialTangent(v); err != nil {
		return err
	}
	k.OutSpatialTangent = append(k.OutSpatialTangent[:0], v...)
	return nil
}

// blockSlice returns the bpk-byte slice for this keyframe's block, or
// an error if bounds don't permit.
func (k *Keyframe) blockSlice() ([]byte, error) {
	kb, ok := k.back.(*keyframeBackrefs)
	if !ok || kb == nil {
		return nil, fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	return kb.blockSlice()
}

// blockSize returns the bpk for this keyframe.
func (k *Keyframe) blockSize() int {
	kb, ok := k.back.(*keyframeBackrefs)
	if !ok || kb == nil {
		return 0
	}
	return kb.blockSize()
}
