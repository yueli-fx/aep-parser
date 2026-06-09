package serializer

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/scene"
)

// Property-level keyframe insert/delete (length-variable ldat rebuild that
// re-parses the property's keyframe stream). These are structural serializer
// free-functions reaching the concrete property/keyframe back-refs; the pure
// scene setters (SetStaticValue / SetExpressionEnabled / SetExpression) live
// in internal/scene.

// InsertKeyframe builds a new bpk-byte keyframe block and inserts it
// into the property's ldat stream, then updates the lhd3 count header.
// Returns the new Keyframe and its index in Property.Keyframes
// (insertion is time-sorted; ties land after existing keys at the
// same time).
//
// Requires the property to already have ≥1 keyframe so the new block
// can clone the existing layout (header byte @0x07, bpk, etc.). For
// properties without keyframes, use SetStaticValue or build keyframes
// in AE first — synthesizing the lhd3/ldat chunks from scratch isn't
// supported yet.
//
// `value` follows the same rules as Keyframe.SetValue:
//   - 1D property: pass float64
//   - multi-component: pass []float64 (length == Property.Components)
//
// The new keyframe's interpolation is Linear/Linear; ease + tangents
// are zeroed. Call SetInInterp / SetInTemporalEase / SetInSpatialTangent
// on the returned Keyframe to refine.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. BREAKING vs the former Property.InsertKeyframe method form.
func InsertKeyframe(p *Property, time float64, value any) (*Keyframe, int, error) {
	pb := propertyBack(p)
	if pb == nil || pb.ldat == nil || pb.lhd3 == nil {
		return nil, -1, fmt.Errorf("property %q: no existing keyframes (insert from scratch not supported)", p.MatchName)
	}
	if pb.bytesPerKF <= 0 {
		return nil, -1, fmt.Errorf("property %q: bytesPerKF is zero", p.MatchName)
	}
	if time < 0 {
		return nil, -1, fmt.Errorf("property %q: negative time %v", p.MatchName, time)
	}

	// Decide insertion index (time-sorted; ties go after).
	insertIdx := len(p.Keyframes)
	for i, kf := range p.Keyframes {
		if time < kf.Time {
			insertIdx = i
			break
		}
	}

	tickRate := aeLegacyTimeBase
	if len(p.Keyframes) > 0 {
		if kb := keyframeBack(p.Keyframes[0]); kb != nil {
			tickRate = kb.tickRate
		}
	}
	if tickRate == 0 {
		tickRate = aeLegacyTimeBase
	}

	// Build the new block: clone header byte @0x07 from an existing
	// keyframe (sample[0]) so layout matches; zero everything else.
	bpk := pb.bytesPerKF
	if len(pb.ldat.Data) < bpk {
		return nil, -1, fmt.Errorf("property %q: ldat shorter than one block (%d bytes, bpk=%d)", p.MatchName, len(pb.ldat.Data), bpk)
	}
	headerByte := pb.ldat.Data[0x07]
	block := make([]byte, bpk)
	binary.BigEndian.PutUint32(block[0:4], uint32(math.Round(time*tickRate)))
	block[0x04] = byte(scene.InterpLinear)
	block[0x05] = byte(scene.InterpLinear)
	block[0x07] = headerByte

	// Write value at the layout-appropriate offset.
	l := layoutFor(headerByte, p.Components)
	switch v := value.(type) {
	case float64:
		if p.Components != 1 {
			return nil, -1, fmt.Errorf("property %q is %dD; expected []float64", p.MatchName, p.Components)
		}
		if err := writeFloat64(block, l.valueOff, v); err != nil {
			return nil, -1, err
		}
	case []float64:
		if len(v) != p.Components {
			return nil, -1, fmt.Errorf("property %q is %dD; got %d-component value", p.MatchName, p.Components, len(v))
		}
		for i, x := range v {
			if err := writeFloat64(block, l.valueOff+i*8, x); err != nil {
				return nil, -1, err
			}
		}
	default:
		return nil, -1, fmt.Errorf("property %q: unsupported value type %T", p.MatchName, value)
	}

	// Splice the block into ldat at insertIdx * bpk.
	splice := insertIdx * bpk
	old := pb.ldat.Data
	newData := make([]byte, 0, len(old)+bpk)
	newData = append(newData, old[:splice]...)
	newData = append(newData, block...)
	newData = append(newData, old[splice:]...)
	pb.ldat.Data = newData

	// Bump count in lhd3 @0x08.
	if len(pb.lhd3.Data) >= 0x0C {
		newCount := binary.BigEndian.Uint32(pb.lhd3.Data[0x08:0x0C]) + 1
		binary.BigEndian.PutUint32(pb.lhd3.Data[0x08:0x0C], newCount)
	}

	// Rebuild Property.Keyframes — the simplest correct path. Re-points
	// every Keyframe.offset to its new position in the resized ldat.
	if err := reparseKeyframes(p, tickRate); err != nil {
		return nil, -1, fmt.Errorf("property %q: reparse after InsertKeyframe: %w", p.MatchName, err)
	}
	if insertIdx >= len(p.Keyframes) {
		insertIdx = len(p.Keyframes) - 1
	}
	return p.Keyframes[insertIdx], insertIdx, nil
}

// DeleteKeyframe removes the keyframe at index i from the property's
// ldat stream and decrements the lhd3 count header. Returns an error
// when i is out of range or the property has no keyframe stream.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. BREAKING vs the former Property.DeleteKeyframe method form.
func DeleteKeyframe(p *Property, i int) error {
	pb := propertyBack(p)
	if pb == nil || pb.ldat == nil || pb.lhd3 == nil {
		return fmt.Errorf("property %q: no keyframe stream", p.MatchName)
	}
	if i < 0 || i >= len(p.Keyframes) {
		return fmt.Errorf("property %q: keyframe index %d out of range [0,%d)", p.MatchName, i, len(p.Keyframes))
	}
	bpk := pb.bytesPerKF
	splice := i * bpk
	old := pb.ldat.Data
	if splice+bpk > len(old) {
		return fmt.Errorf("property %q: ldat shorter than expected (off=%d, bpk=%d, len=%d)", p.MatchName, splice, bpk, len(old))
	}
	newData := make([]byte, 0, len(old)-bpk)
	newData = append(newData, old[:splice]...)
	newData = append(newData, old[splice+bpk:]...)
	pb.ldat.Data = newData

	if len(pb.lhd3.Data) >= 0x0C {
		newCount := binary.BigEndian.Uint32(pb.lhd3.Data[0x08:0x0C])
		if newCount > 0 {
			newCount--
		}
		binary.BigEndian.PutUint32(pb.lhd3.Data[0x08:0x0C], newCount)
	}

	tickRate := aeLegacyTimeBase
	if len(p.Keyframes) > 0 {
		if kb := keyframeBack(p.Keyframes[0]); kb != nil {
			tickRate = kb.tickRate
		}
	}
	if tickRate == 0 {
		tickRate = aeLegacyTimeBase
	}
	return reparseKeyframes(p, tickRate)
}

// reparseKeyframes rebuilds Property.Keyframes from the current
// ldat/lhd3 bytes. Used after InsertKeyframe / DeleteKeyframe so
// Keyframe.offset / Value / etc. reflect the new stream layout.
func reparseKeyframes(p *Property, tickRate float64) error {
	pb := propertyBack(p)
	if pb == nil || pb.ldat == nil || pb.lhd3 == nil {
		return fmt.Errorf("property %q: missing ldat/lhd3", p.MatchName)
	}
	if len(pb.lhd3.Data) < 0x14 {
		return fmt.Errorf("property %q: lhd3 too short (%d bytes)", p.MatchName, len(pb.lhd3.Data))
	}
	count := int(binary.BigEndian.Uint32(pb.lhd3.Data[0x08:0x0C]))
	bpk := int(binary.BigEndian.Uint32(pb.lhd3.Data[0x10:0x14]))
	if count < 0 || bpk <= 0 || count*bpk > len(pb.ldat.Data) {
		return fmt.Errorf("property %q: lhd3 says count=%d bpk=%d but ldat has %d bytes", p.MatchName, count, bpk, len(pb.ldat.Data))
	}
	pb.bytesPerKF = bpk
	p.Keyframes = make([]*Keyframe, 0, count)
	for i := 0; i < count; i++ {
		off := i * bpk
		kf := &Keyframe{}
		setKeyframeBack(kf, &keyframeBackrefs{
			ldat:     pb.ldat,
			offset:   off,
			dims:     p.Components,
			tickRate: tickRate,
		})
		kf.Time = float64(binary.BigEndian.Uint32(pb.ldat.Data[off:off+4])) / tickRate
		kf.Value = readKFValue(pb.ldat.Data, off, p.Components)
		decodeEasing(kf, pb.ldat.Data[off:off+bpk])
		p.Keyframes = append(p.Keyframes, kf)
	}
	return nil
}
