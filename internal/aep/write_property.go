package aep

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/rifx"
)

// Property-level writers: static value, expression source &
// enabled bit, and keyframe insert/delete (length-variable ldat
// rebuild that re-parses the property's keyframe stream).

// SetStaticValue rewrites a property's constant value in-place (only valid
// for properties without keyframes — those with a cdat chunk).
func (p *Property) SetStaticValue(v any) error {
	if p.cdat == nil {
		return fmt.Errorf("property %q: no static-value chunk (has keyframes?)", p.MatchName)
	}
	switch x := v.(type) {
	case float64:
		if p.Components != 1 {
			return fmt.Errorf("property %q is %dD, expected []float64", p.MatchName, p.Components)
		}
		if err := writeFloat64(p.cdat.Data, 0, x); err != nil {
			return err
		}
		p.StaticValue = x
		return nil
	case []float64:
		if len(x) != p.Components {
			return fmt.Errorf("property %q: got %d components, property is %dD", p.MatchName, len(x), p.Components)
		}
		for i, f := range x {
			if err := writeFloat64(p.cdat.Data, i*8, f); err != nil {
				return err
			}
		}
		p.StaticValue = append([]float64(nil), x...)
		return nil
	default:
		return fmt.Errorf("property: unsupported value type %T", v)
	}
}

// SetExpressionEnabled toggles whether AE evaluates the property's
// expression at render time (separate knob from `SetExpression` which
// writes the JS source itself).
//
// RE'd against AE 2020 fixture: byte at tdb4 payload offset 0x78 acts
// as a "disabled" flag — value 0 = enabled (AE applies expression),
// value 1 = disabled (expression source preserved but ignored).
// length-preserving (1 byte).
//
// Requires the property's tdbs to contain a `tdb4` chunk (always
// present for properties parsed from real .aep files). Returns an
// error otherwise.
func (p *Property) SetExpressionEnabled(enabled bool) error {
	if p.tdbs == nil {
		return fmt.Errorf("property %q: no tdbs reference", p.MatchName)
	}
	tdb4 := p.tdbs.FindFirst(rifx.IDtdb4)
	if tdb4 == nil {
		tdb4 = p.tdbs.FindFirst(rifx.IDTdb4)
	}
	if tdb4 == nil {
		return fmt.Errorf("property %q: tdb4 chunk missing", p.MatchName)
	}
	if len(tdb4.Data) <= 0x78 {
		return fmt.Errorf("property %q: tdb4 too short (%d bytes) for expressionEnabled write", p.MatchName, len(tdb4.Data))
	}
	if enabled {
		tdb4.Data[0x78] = 0
	} else {
		tdb4.Data[0x78] = 1
	}
	p.ExpressionEnabled = enabled
	return nil
}

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
func (p *Property) InsertKeyframe(time float64, value any) (*Keyframe, int, error) {
	if p.ldat == nil || p.lhd3 == nil {
		return nil, -1, fmt.Errorf("property %q: no existing keyframes (insert from scratch not supported)", p.MatchName)
	}
	if p.bytesPerKF <= 0 {
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
		tickRate = p.Keyframes[0].tickRate
	}
	if tickRate == 0 {
		tickRate = aeLegacyTimeBase
	}

	// Build the new block: clone header byte @0x07 from an existing
	// keyframe (sample[0]) so layout matches; zero everything else.
	bpk := p.bytesPerKF
	if len(p.ldat.Data) < bpk {
		return nil, -1, fmt.Errorf("property %q: ldat shorter than one block (%d bytes, bpk=%d)", p.MatchName, len(p.ldat.Data), bpk)
	}
	headerByte := p.ldat.Data[0x07]
	block := make([]byte, bpk)
	binary.BigEndian.PutUint32(block[0:4], uint32(math.Round(time*tickRate)))
	block[0x04] = byte(InterpLinear)
	block[0x05] = byte(InterpLinear)
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
	old := p.ldat.Data
	newData := make([]byte, 0, len(old)+bpk)
	newData = append(newData, old[:splice]...)
	newData = append(newData, block...)
	newData = append(newData, old[splice:]...)
	p.ldat.Data = newData

	// Bump count in lhd3 @0x08.
	if len(p.lhd3.Data) >= 0x0C {
		newCount := binary.BigEndian.Uint32(p.lhd3.Data[0x08:0x0C]) + 1
		binary.BigEndian.PutUint32(p.lhd3.Data[0x08:0x0C], newCount)
	}

	// Rebuild Property.Keyframes — the simplest correct path. Re-points
	// every Keyframe.offset to its new position in the resized ldat.
	if err := p.reparseKeyframes(tickRate); err != nil {
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
func (p *Property) DeleteKeyframe(i int) error {
	if p.ldat == nil || p.lhd3 == nil {
		return fmt.Errorf("property %q: no keyframe stream", p.MatchName)
	}
	if i < 0 || i >= len(p.Keyframes) {
		return fmt.Errorf("property %q: keyframe index %d out of range [0,%d)", p.MatchName, i, len(p.Keyframes))
	}
	bpk := p.bytesPerKF
	splice := i * bpk
	old := p.ldat.Data
	if splice+bpk > len(old) {
		return fmt.Errorf("property %q: ldat shorter than expected (off=%d, bpk=%d, len=%d)", p.MatchName, splice, bpk, len(old))
	}
	newData := make([]byte, 0, len(old)-bpk)
	newData = append(newData, old[:splice]...)
	newData = append(newData, old[splice+bpk:]...)
	p.ldat.Data = newData

	if len(p.lhd3.Data) >= 0x0C {
		newCount := binary.BigEndian.Uint32(p.lhd3.Data[0x08:0x0C])
		if newCount > 0 {
			newCount--
		}
		binary.BigEndian.PutUint32(p.lhd3.Data[0x08:0x0C], newCount)
	}

	tickRate := aeLegacyTimeBase
	if len(p.Keyframes) > 0 {
		tickRate = p.Keyframes[0].tickRate
	}
	if tickRate == 0 {
		tickRate = aeLegacyTimeBase
	}
	return p.reparseKeyframes(tickRate)
}

// reparseKeyframes rebuilds Property.Keyframes from the current
// ldat/lhd3 bytes. Used after InsertKeyframe / DeleteKeyframe so
// Keyframe.offset / Value / etc. reflect the new stream layout.
func (p *Property) reparseKeyframes(tickRate float64) error {
	if p.ldat == nil || p.lhd3 == nil {
		return fmt.Errorf("property %q: missing ldat/lhd3", p.MatchName)
	}
	if len(p.lhd3.Data) < 0x14 {
		return fmt.Errorf("property %q: lhd3 too short (%d bytes)", p.MatchName, len(p.lhd3.Data))
	}
	count := int(binary.BigEndian.Uint32(p.lhd3.Data[0x08:0x0C]))
	bpk := int(binary.BigEndian.Uint32(p.lhd3.Data[0x10:0x14]))
	if count < 0 || bpk <= 0 || count*bpk > len(p.ldat.Data) {
		return fmt.Errorf("property %q: lhd3 says count=%d bpk=%d but ldat has %d bytes", p.MatchName, count, bpk, len(p.ldat.Data))
	}
	p.bytesPerKF = bpk
	p.Keyframes = make([]*Keyframe, 0, count)
	for i := 0; i < count; i++ {
		off := i * bpk
		kf := &Keyframe{
			ldat:     p.ldat,
			offset:   off,
			dims:     p.Components,
			tickRate: tickRate,
		}
		kf.Time = float64(binary.BigEndian.Uint32(p.ldat.Data[off:off+4])) / tickRate
		kf.Value = readKFValue(p.ldat.Data, off, p.Components)
		decodeEasing(kf, p.ldat.Data[off:off+bpk])
		p.Keyframes = append(p.Keyframes, kf)
	}
	return nil
}

// SetExpression rewrites the JavaScript expression source attached to
// this property.
//
// length-variable — the underlying Utf8 chunk's data is replaced
// (or a new Utf8 chunk is inserted into the property's tdbs LIST when
// none existed previously; conversely, passing "" removes the chunk
// entirely, leaving the property with no expression).
//
// Note: AE has a separate Enable/Disable Expression toggle (in addition
// to the source). Removing the Utf8 chunk via SetExpression("") gets
// you the "no expression at all" state.
//
// Returns an error if the property is one built outside the parser
// (no owning tdbs LIST reference).
func (p *Property) SetExpression(source string) error {
	if p.tdbs == nil {
		return fmt.Errorf("property %q: no tdbs reference (built outside parser?)", p.MatchName)
	}
	if source == "" {
		// Remove existing expression chunk if present.
		if p.exprChunk != nil {
			out := p.tdbs.Children[:0]
			for _, ch := range p.tdbs.Children {
				if ch == p.exprChunk {
					continue
				}
				out = append(out, ch)
			}
			p.tdbs.Children = out
			p.exprChunk = nil
		}
		p.Expression = ""
		return nil
	}
	if p.exprChunk != nil {
		p.exprChunk.Data = []byte(source)
	} else {
		newUtf8 := &rifx.Chunk{ID: rifx.IDUtf8, Data: []byte(source)}
		p.tdbs.Children = append(p.tdbs.Children, newUtf8)
		p.exprChunk = newUtf8
	}
	p.Expression = source
	return nil
}
