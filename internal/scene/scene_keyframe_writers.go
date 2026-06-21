package scene

import (
	"fmt"
	"math"
)

// Keyframe writers. All edits are length-preserving block writes
// into the parent ldat. Layout dispatch (spatial vs non-spatial)
// is owned by keyframeBackrefs.

// aeLegacyTimeBase is the fallback per-second keyframe tick rate (used when
// the owning composition's TickRate is unknown). Mirrors the parser-side
// constant of the same name; kept here so scene time accessors avoid a
// serializer dependency.
const aeLegacyTimeBase = 8000.0

// @summary     Set the keyframe's time
// @description Uses the owning composition's TickRate, decoded at parse
//   time from cdta.
// @param       seconds  the new keyframe time in seconds
// @domain      keyframe
// @stability   stable
// @verify      ae-accept
// @gate        TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving — the ldat chunk size never changes, so
//   writing the project back never relocates anything
// @alias       keyframe time,关键帧时间,kf time,时间点
func (k *Keyframe) SetTime(seconds float64) error {
	if k.back == nil {
		return fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	if err := k.back.SetTime(seconds); err != nil {
		return err
	}
	rate := k.back.CompTickRate()
	if rate == 0 {
		rate = aeLegacyTimeBase
	}
	ticks := uint32(math.Round(seconds * rate))
	k.Time = float64(ticks) / rate
	return nil
}

// @summary     Set the keyframe's numeric value
// @description For a 1D property pass a float64. For a multi-component
//   property pass []float64 with exact length equal to the property's
//   component count.
// @param       v  the new value: float64 for 1D, []float64 matching the
//   property's component count for N-D
// @domain      keyframe
// @stability   stable
// @verify      ae-accept
// @gate        TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving — chunk size is unchanged
// @alias       keyframe value,关键帧值,kf value,属性值
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

// @summary     Set the keyframe's in-side interpolation type
// @description See SetOutInterp for the out side.
// @param       t  the new in-interpolation type
// @domain      keyframe
// @stability   stable
// @verify      ae-accept
// @gate        TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving (1 byte at the keyframe block offset 0x04)
// @alias       keyframe interpolation,关键帧插值,in interpolation,缓入缓出
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

// @summary     Set the keyframe's out-side interpolation type
// @description See SetInInterp for the in side.
// @param       t  the new out-interpolation type
// @domain      keyframe
// @stability   stable
// @verify      ae-accept
// @gate        TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving (1 byte at the keyframe block offset 0x05)
// @alias       keyframe out interpolation,关键帧输出插值,out interpolation,贝塞尔线性
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

// @summary     Set the keyframe's in-side temporal ease
// @description The slice length must match the keyframe's current ease
//   shape: a spatial property (Position / Anchor) or 1D non-spatial
//   property takes length 1; a non-spatial N-D property (Scale 3D /
//   Feather 2D / 4D color) takes length N.
// @param       eases  the new per-component in-ease values
// @domain      keyframe
// @stability   stable
// @verify      ae-accept
// @gate        TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving (8 or 16 bytes per ease, at fixed offsets
//   determined by the property's layout)
// @alias       temporal ease,时间缓动,ease in,速度曲线,缓入
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

// @summary     Set the keyframe's out-side temporal ease
// @description Mirrors SetInTemporalEase for the out side; the slice
//   length must match the keyframe's current ease shape.
// @param       eases  the new per-component out-ease values
// @domain      keyframe
// @stability   stable
// @verify      ae-accept
// @gate        TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving
// @alias       temporal ease out,缓出,ease out,速度曲线出
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

// @summary     Set the keyframe's in-side spatial Bezier tangent
// @description Only valid for spatial properties (Position / AnchorPoint
//   with motion path enabled). See SetOutSpatialTangent for the out side.
// @param       v  the new tangent vector; length must equal
//   Property.Components
// @domain      keyframe
// @stability   stable
// @verify      ae-accept
// @gate        TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving (dims x 8 bytes)
// @alias       spatial tangent,空间切线,bezier tangent,运动路径切线,in tangent
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

// @summary     Set the keyframe's out-side spatial Bezier tangent
// @description Only valid for spatial properties (Position / AnchorPoint
//   with motion path enabled). See SetInSpatialTangent for the in side.
// @param       v  the new tangent vector; length must equal
//   Property.Components
// @domain      keyframe
// @stability   stable
// @verify      ae-accept
// @gate        TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving (dims x 8 bytes)
// @alias       out tangent,出切线,bezier out tangent,运动路径出切线
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
