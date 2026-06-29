package serializer

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// keyframeBack returns the concrete back-refs behind a Keyframe's writer
// interface for serializer-stage raw chunk access (tick rate, block layout).
// Returns nil when the keyframe was built outside the parser. Free function
// (the receiver is a scene type, so it can't be a method post package-split).
func keyframeBack(k *Keyframe) *keyframeBackrefs {
	if kb, ok := scene.KeyframeBack(k).(*keyframeBackrefs); ok {
		return kb
	}
	return nil
}

// setKeyframeBack wires a concrete back-ref onto a scene Keyframe via the
// scene wiring API (serializer stage cannot set the unexported field directly).
func setKeyframeBack(k *Keyframe, b *keyframeBackrefs) {
	scene.SetKeyframeBack(k, b)
}

// keyframeBackrefs holds the rifx.Chunk reference plus the cached layout
// metadata (offset within ldat.Data, dimensionality, owning comp's TickRate
// and FrameRate) that powers Keyframe.SetTime / SetValue / SetFrameTime
// in-place writes.
//
// Cached fields (offset / dims / tickRate / compFps) duplicate state that
// could be re-derived (offset from Property.lhd3 bpk + keyframe index;
// dims from Property.Components; rates from owning composition). They live
// here because the parser already computed them once per keyframe block.
//
// Lifecycle:
//   - Populated by parseKeyframes when keyframes are decoded from an .aep file.
//   - Also populated by write_property.go::reparseKeyframes after a
//     length-variable splice rewrites the ldat.Data.
//   - Nil for keyframes built outside the parser.
//   - opaque is reserved for a future phase (per-keyframe-block ancillary
//     chunks if/when we identify any); currently nil.
type keyframeBackrefs struct {
	ldat     *rifx.Chunk
	offset   int
	dims     int
	tickRate float64
	compFps  float64

	opaque map[rifx.ChunkID]*rifx.Chunk
}

var _ KeyframeWriter = (*keyframeBackrefs)(nil)

func (b *keyframeBackrefs) FrameRateHz() float64 {
	if b == nil {
		return 0
	}
	return b.compFps
}

func (b *keyframeBackrefs) CompTickRate() float64 {
	if b == nil {
		return 0
	}
	return b.tickRate
}

func (b *keyframeBackrefs) SetTime(seconds float64) error {
	if b.ldat == nil {
		return fmt.Errorf("keyframe at offset %d: no underlying ldat chunk", b.offset)
	}
	if seconds < 0 {
		return fmt.Errorf("keyframe time %v: negative not supported", seconds)
	}
	if b.offset+4 > len(b.ldat.Data) {
		return fmt.Errorf("keyframe time write: offset %d out of bounds (ldat=%d bytes)", b.offset, len(b.ldat.Data))
	}
	rate := b.tickRate
	if rate == 0 {
		rate = aeLegacyTimeBase
	}
	ticks := uint32(math.Round(seconds * rate))
	binary.BigEndian.PutUint32(b.ldat.Data[b.offset:b.offset+4], ticks)
	return nil
}

func (b *keyframeBackrefs) SetValue(v any) error {
	if b.ldat == nil {
		return fmt.Errorf("keyframe at offset %d: no underlying ldat chunk", b.offset)
	}
	valOffset := kfValueOffset(b.ldat.Data, b.offset, b.dims)
	switch x := v.(type) {
	case float64:
		if b.dims != 1 {
			return fmt.Errorf("keyframe: property is %dD, expected []float64 value", b.dims)
		}
		return writeFloat64(b.ldat.Data, b.offset+valOffset, x)
	case []float64:
		if len(x) != b.dims {
			return fmt.Errorf("keyframe: got %d components, property is %dD", len(x), b.dims)
		}
		for i, f := range x {
			if err := writeFloat64(b.ldat.Data, b.offset+valOffset+i*8, f); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("keyframe: unsupported value type %T", v)
	}
}

func (b *keyframeBackrefs) SetInInterp(t InterpType) error {
	return b.setInterpByte(0x04, t)
}

func (b *keyframeBackrefs) SetOutInterp(t InterpType) error {
	return b.setInterpByte(0x05, t)
}

func (b *keyframeBackrefs) setInterpByte(off int, t InterpType) error {
	if b.ldat == nil {
		return fmt.Errorf("keyframe at offset %d: no underlying ldat chunk", b.offset)
	}
	if b.offset+off+1 > len(b.ldat.Data) {
		return fmt.Errorf("keyframe interp write: offset %d out of bounds", b.offset+off)
	}
	b.ldat.Data[b.offset+off] = byte(t)
	return nil
}

func (b *keyframeBackrefs) SetInTemporalEase(eases []TemporalEase) error {
	return b.setTemporalEase(eases, true)
}

func (b *keyframeBackrefs) SetOutTemporalEase(eases []TemporalEase) error {
	return b.setTemporalEase(eases, false)
}

func (b *keyframeBackrefs) setTemporalEase(eases []TemporalEase, inSide bool) error {
	if b.ldat == nil {
		return fmt.Errorf("keyframe at offset %d: no underlying ldat chunk", b.offset)
	}
	blk, err := b.blockSlice()
	if err != nil {
		return err
	}
	dims := b.dims
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
		if err := writeFloat64(b.ldat.Data, b.offset+speedOff, eases[0].Speed); err != nil {
			return err
		}
		if err := writeFloat64(b.ldat.Data, b.offset+infOff, eases[0].Influence); err != nil {
			return err
		}
		return nil
	}
	if len(eases) != dims {
		return fmt.Errorf("keyframe (non-spatial %dD): expected %d eases, got %d", dims, dims, len(eases))
	}
	speedBase, infBase := dims, 2*dims
	if !inSide {
		speedBase, infBase = 3*dims, 4*dims
	}
	for i := 0; i < dims; i++ {
		if err := writeFloat64(b.ldat.Data, b.offset+8+(speedBase+i)*8, eases[i].Speed); err != nil {
			return err
		}
		if err := writeFloat64(b.ldat.Data, b.offset+8+(infBase+i)*8, eases[i].Influence); err != nil {
			return err
		}
	}
	return nil
}

func (b *keyframeBackrefs) SetInSpatialTangent(v []float64) error {
	return b.setSpatialTangent(v, true)
}

func (b *keyframeBackrefs) SetOutSpatialTangent(v []float64) error {
	return b.setSpatialTangent(v, false)
}

func (b *keyframeBackrefs) setSpatialTangent(v []float64, inSide bool) error {
	if b.ldat == nil {
		return fmt.Errorf("keyframe at offset %d: no underlying ldat chunk", b.offset)
	}
	blk, err := b.blockSlice()
	if err != nil {
		return err
	}
	dims := b.dims
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
	tanBase := l.valueOff + dims*8
	if !inSide {
		tanBase += dims * 8
	}
	for i, x := range v {
		if err := writeFloat64(b.ldat.Data, b.offset+tanBase+i*8, x); err != nil {
			return err
		}
	}
	return nil
}

func (b *keyframeBackrefs) blockSlice() ([]byte, error) {
	if b.ldat == nil {
		return nil, fmt.Errorf("keyframe at offset %d: no underlying ldat chunk", b.offset)
	}
	end := b.offset + b.blockSize()
	if end > len(b.ldat.Data) {
		return nil, fmt.Errorf("keyframe block at offset %d exceeds ldat (%d bytes, end=%d)", b.offset, len(b.ldat.Data), end)
	}
	return b.ldat.Data[b.offset:end], nil
}

func (b *keyframeBackrefs) blockSize() int {
	if b.ldat == nil || b.offset+8 > len(b.ldat.Data) {
		return 0
	}
	dims := b.dims
	if dims <= 0 {
		dims = 1
	}
	l := layoutFor(b.ldat.Data[b.offset+0x07], dims)
	if l.spatialStyle {
		return 0x38 + 3*dims*8
	}
	return 0x08 + 5*dims*8
}
