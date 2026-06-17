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

// SetTime rewrites this keyframe's time in-place using its owning
// composition's TickRate (decoded at parse time from cdta). Chunk size
// does not change, so writing the project back never relocates anything.
//
//aep:cap domain=keyframe tier=stable verify=ae-accept gate=TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025 boundary="length-preserving(ldat tick write);chunk size 不变;双版本 AE gated(keyframe_mutate)" alias="keyframe time,关键帧时间,kf time,时间点"
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

// SetValue rewrites this keyframe's numeric value in-place.
//
// For a 1D property pass a float64. For a multi-component property pass
// []float64 of exact length equal to the property's component count.
// Length-preserving: chunk size is unchanged.
//
//aep:cap domain=keyframe tier=stable verify=ae-accept gate=TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025 boundary="length-preserving;1D 传 float64,N-D 传 []float64 且长度须匹配;双版本 AE gated(keyframe_mutate)" alias="keyframe value,关键帧值,kf value,属性值"
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
//
//aep:cap domain=keyframe tier=stable verify=ae-accept gate=TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025 boundary="length-preserving(1B @0x04);双版本 AE gated(keyframe_mutate)" alias="keyframe interpolation,关键帧插值,in interpolation,缓入缓出"
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
//
//aep:cap domain=keyframe tier=stable verify=ae-accept gate=TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025 boundary="length-preserving(1B @0x05);双版本 AE gated(keyframe_mutate)" alias="keyframe out interpolation,关键帧输出插值,out interpolation,贝塞尔线性"
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
//
//aep:cap domain=keyframe tier=stable verify=ae-accept gate=TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025 boundary="length-preserving;slice 长度须匹配 ease shape(空间属性 1 / N-D 为 N);双版本 AE gated(keyframe_mutate)" alias="temporal ease,时间缓动,ease in,速度曲线,缓入"
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
//
//aep:cap domain=keyframe tier=stable verify=ae-accept gate=TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025 boundary="length-preserving;slice 长度须匹配 ease shape;双版本 AE gated(keyframe_mutate)" alias="temporal ease out,缓出,ease out,速度曲线出"
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
//
//aep:cap domain=keyframe tier=stable verify=ae-accept gate=TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025 boundary="length-preserving(dims×8B);仅空间属性有效(Position/AnchorPoint);slice 长度须等于 Property.Components;双版本 AE gated(keyframe_mutate)" alias="spatial tangent,空间切线,bezier tangent,运动路径切线,in tangent"
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
//
//aep:cap domain=keyframe tier=stable verify=ae-accept gate=TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025 boundary="length-preserving(dims×8B);仅空间属性有效;slice 长度须等于 Property.Components;双版本 AE gated(keyframe_mutate)" alias="out tangent,出切线,bezier out tangent,运动路径出切线"
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
