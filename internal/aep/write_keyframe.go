package aep

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Keyframe writers. All edits are length-preserving block writes
// into the parent ldat. Layout dispatch (spatial vs non-spatial)
// is owned by the Keyframe value itself.

// SetTime rewrites this keyframe's time in-place using its owning
// composition's TickRate (decoded at parse time from cdta). Chunk size
// does not change, so writing the project back never relocates anything.
func (k *Keyframe) SetTime(seconds float64) error {
	if k.back == nil {
		return fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	if k.back.ldat == nil {
		return fmt.Errorf("keyframe at offset %d: no underlying ldat chunk", k.back.offset)
	}
	if seconds < 0 {
		return fmt.Errorf("keyframe time %v: negative not supported", seconds)
	}
	if k.back.offset+4 > len(k.back.ldat.Data) {
		return fmt.Errorf("keyframe time write: offset %d out of bounds (ldat=%d bytes)", k.back.offset, len(k.back.ldat.Data))
	}
	rate := k.back.tickRate
	if rate == 0 {
		rate = aeLegacyTimeBase
	}
	ticks := uint32(math.Round(seconds * rate))
	binary.BigEndian.PutUint32(k.back.ldat.Data[k.back.offset:k.back.offset+4], ticks)
	k.Time = float64(ticks) / rate
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
	if k.back.ldat == nil {
		return fmt.Errorf("keyframe at offset %d: no underlying ldat chunk", k.back.offset)
	}
	valOffset := kfValueOffset(k.back.ldat.Data, k.back.offset, k.back.dims)
	switch x := v.(type) {
	case float64:
		if k.back.dims != 1 {
			return fmt.Errorf("keyframe: property is %dD, expected []float64 value", k.back.dims)
		}
		if err := writeFloat64(k.back.ldat.Data, k.back.offset+valOffset, x); err != nil {
			return err
		}
		k.Value = x
		return nil
	case []float64:
		if len(x) != k.back.dims {
			return fmt.Errorf("keyframe: got %d components, property is %dD", len(x), k.back.dims)
		}
		for i, f := range x {
			if err := writeFloat64(k.back.ldat.Data, k.back.offset+valOffset+i*8, f); err != nil {
				return err
			}
		}
		k.Value = append([]float64(nil), x...)
		return nil
	default:
		return fmt.Errorf("keyframe: unsupported value type %T", v)
	}
}

// SetInInterp / SetOutInterp write a new interpolation enum byte to
// the keyframe block @0x04 / @0x05.
// length-preserving (1 byte).
func (k *Keyframe) SetInInterp(t InterpType) error {
	return k.setInterpByte(0x04, t, func() { k.InInterp = t })
}

// SetOutInterp see SetInInterp.
func (k *Keyframe) SetOutInterp(t InterpType) error {
	return k.setInterpByte(0x05, t, func() { k.OutInterp = t })
}

func (k *Keyframe) setInterpByte(off int, t InterpType, sync func()) error {
	if k.back == nil {
		return fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	if k.back.ldat == nil {
		return fmt.Errorf("keyframe at offset %d: no underlying ldat chunk", k.back.offset)
	}
	if k.back.offset+off+1 > len(k.back.ldat.Data) {
		return fmt.Errorf("keyframe interp write: offset %d out of bounds", k.back.offset+off)
	}
	k.back.ldat.Data[k.back.offset+off] = byte(t)
	sync()
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
	return k.setTemporalEase(eases, true)
}

// SetOutTemporalEase mirrors SetInTemporalEase for the out side.
func (k *Keyframe) SetOutTemporalEase(eases []TemporalEase) error {
	return k.setTemporalEase(eases, false)
}

func (k *Keyframe) setTemporalEase(eases []TemporalEase, inSide bool) error {
	if k.back == nil {
		return fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	if k.back.ldat == nil {
		return fmt.Errorf("keyframe at offset %d: no underlying ldat chunk", k.back.offset)
	}
	blk, err := k.blockSlice()
	if err != nil {
		return err
	}
	dims := k.back.dims
	if dims <= 0 {
		dims = 1
	}
	l := layoutFor(blk[0x07], dims)
	if l.spatialStyle {
		if len(eases) != 1 {
			return fmt.Errorf("keyframe (spatial / 1D): expected 1 ease, got %d", len(eases))
		}
		speedOff, infOff := 0x18, 0x20
		if !inSide {
			speedOff, infOff = 0x28, 0x30
		}
		if err := writeFloat64(k.back.ldat.Data, k.back.offset+speedOff, eases[0].Speed); err != nil {
			return err
		}
		if err := writeFloat64(k.back.ldat.Data, k.back.offset+infOff, eases[0].Influence); err != nil {
			return err
		}
		dst := &k.InTemporalEase
		if !inSide {
			dst = &k.OutTemporalEase
		}
		*dst = []TemporalEase{eases[0]}
		return nil
	}
	if len(eases) != dims {
		return fmt.Errorf("keyframe (non-spatial %dD): expected %d eases, got %d", dims, dims, len(eases))
	}
	// Non-spatial layout: Speed/Influence interleaved per side per component.
	// in-speed   @ 0x08 + (dims+i)*8     out-speed @ 0x08 + (3*dims+i)*8
	// in-infl    @ 0x08 + (2*dims+i)*8   out-infl  @ 0x08 + (4*dims+i)*8
	speedBase, infBase := dims, 2*dims
	if !inSide {
		speedBase, infBase = 3*dims, 4*dims
	}
	for i := 0; i < dims; i++ {
		if err := writeFloat64(k.back.ldat.Data, k.back.offset+8+(speedBase+i)*8, eases[i].Speed); err != nil {
			return err
		}
		if err := writeFloat64(k.back.ldat.Data, k.back.offset+8+(infBase+i)*8, eases[i].Influence); err != nil {
			return err
		}
	}
	dst := &k.InTemporalEase
	if !inSide {
		dst = &k.OutTemporalEase
	}
	*dst = append((*dst)[:0], eases...)
	return nil
}

// SetInSpatialTangent / SetOutSpatialTangent write 3D Bezier tangent
// vectors. Only valid for spatial properties (Position / AnchorPoint
// with motion path enabled). Slice length must equal `Property.Components`.
// length-preserving (dims × 8 bytes).
func (k *Keyframe) SetInSpatialTangent(v []float64) error {
	return k.setSpatialTangent(v, true)
}

// SetOutSpatialTangent see SetInSpatialTangent.
func (k *Keyframe) SetOutSpatialTangent(v []float64) error {
	return k.setSpatialTangent(v, false)
}

func (k *Keyframe) setSpatialTangent(v []float64, inSide bool) error {
	if k.back == nil {
		return fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	if k.back.ldat == nil {
		return fmt.Errorf("keyframe at offset %d: no underlying ldat chunk", k.back.offset)
	}
	blk, err := k.blockSlice()
	if err != nil {
		return err
	}
	dims := k.back.dims
	if dims <= 0 {
		dims = 1
	}
	l := layoutFor(blk[0x07], dims)
	if !l.spatialStyle {
		return fmt.Errorf("keyframe is non-spatial (%dD); SpatialTangent not applicable", dims)
	}
	if len(v) != dims {
		return fmt.Errorf("keyframe spatial tangent: expected %d components, got %d", dims, len(v))
	}
	tanBase := l.valueOff + dims*8 // in tangents follow value
	if !inSide {
		tanBase += dims * 8 // out tangents follow in tangents
	}
	for i, x := range v {
		if err := writeFloat64(k.back.ldat.Data, k.back.offset+tanBase+i*8, x); err != nil {
			return err
		}
	}
	dst := &k.InSpatialTangent
	if !inSide {
		dst = &k.OutSpatialTangent
	}
	*dst = append((*dst)[:0], v...)
	return nil
}

// blockSlice returns the bpk-byte slice for this keyframe's block, or
// an error if bounds don't permit.
func (k *Keyframe) blockSlice() ([]byte, error) {
	if k.back == nil {
		return nil, fmt.Errorf("keyframe: no underlying ldat chunk")
	}
	if k.back.ldat == nil {
		return nil, fmt.Errorf("keyframe at offset %d: no underlying ldat chunk", k.back.offset)
	}
	end := k.back.offset + k.blockSize()
	if end > len(k.back.ldat.Data) {
		return nil, fmt.Errorf("keyframe block at offset %d exceeds ldat (%d bytes, end=%d)", k.back.offset, len(k.back.ldat.Data), end)
	}
	return k.back.ldat.Data[k.back.offset:end], nil
}

// blockSize returns the bpk for this keyframe. We don't have the
// parent Property's bytesPerKF on Keyframe directly (we'd need a back
// reference), so re-derive from the layout: spatial → 0x38 + 3*N*8,
// non-spatial → 0x08 + 5*N*8.
func (k *Keyframe) blockSize() int {
	if k.back == nil || k.back.ldat == nil || k.back.offset+8 > len(k.back.ldat.Data) {
		return 0
	}
	dims := k.back.dims
	if dims <= 0 {
		dims = 1
	}
	l := layoutFor(k.back.ldat.Data[k.back.offset+0x07], dims)
	if l.spatialStyle {
		return 0x38 + 3*dims*8
	}
	return 0x08 + 5*dims*8
}
