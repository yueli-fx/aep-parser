package aep

import "fmt"

// Property tdb4 flag readers — mirror of py-aep's `property.is_spatial`,
// `property.is_animated`, etc. All read from the 124-byte tdb4 metadata
// chunk under each tdbs LIST. Byte offsets sourced from py-aep's
// binary/property_chunks.py::Tdb4Chunk:
//
//	byte 0x05      spatial/static flags — bit 3 = is_spatial, bit 0 = static
//	byte 0x0B      can-vary-over-time flags — bit 1 = can_vary_over_time
//	byte 0x39      no_value flags — bit 0 = no_value
//	byte 0x3B      type flags — bit 0 = color, bit 2 = integer, bit 3 = vector
//	byte 0x44      animated (boolean)
//	byte 0x77      expression_disabled (bit 0); inverted form already
//	               exposed as ExpressionEnabled.
//
// Properties built without going through the parser (no tdb4 reference)
// fall back to safe defaults: animated derived from len(Keyframes),
// the others returning false.

// tdb4Bit returns bit `b` of tdb4 byte at offset `off`. Returns false
// when the tdb4 chunk is absent or too short.
func (p *Property) tdb4Bit(off, b int) bool {
	if p.tdb4 == nil || len(p.tdb4.Data) <= off {
		return false
	}
	return (p.tdb4.Data[off]>>uint(b))&1 == 1
}

// IsSpatial reports whether the property is spatial (motion-path) —
// Position and Anchor Point are spatial; Scale / Rotate / Opacity are
// not. Read from tdb4 @0x05 bit 3.
func (p *Property) IsSpatial() bool {
	return p.tdb4Bit(0x05, 3)
}

// IsAnimated reports whether the property has keyframes. When the
// tdb4 chunk is available it reads the explicit animated flag (byte
// 0x44 bool); otherwise it falls back to len(Keyframes) > 0.
func (p *Property) IsAnimated() bool {
	if p.tdb4 == nil || len(p.tdb4.Data) <= 0x44 {
		return len(p.Keyframes) > 0
	}
	return p.tdb4.Data[0x44] != 0
}

// IsColor reports whether the property is a color (RGBA). 4-channel
// color properties like Fill Color, Tritone Highlights/Midtones/Shadows.
// Read from tdb4 @0x3B bit 0.
func (p *Property) IsColor() bool {
	return p.tdb4Bit(0x3B, 0)
}

// IsInteger reports whether the property is integer-valued (some
// dropdowns / enum knobs). Read from tdb4 @0x3B bit 2.
func (p *Property) IsInteger() bool {
	return p.tdb4Bit(0x3B, 2)
}

// IsVector reports whether the property is a vector type — multi-axis
// point properties like Position, Scale. Read from tdb4 @0x3B bit 3.
func (p *Property) IsVector() bool {
	return p.tdb4Bit(0x3B, 3)
}

// IsNoValue reports whether the property has no stored value
// (group-like properties, dropdown markers). Read from tdb4 @0x39 bit 0.
func (p *Property) IsNoValue() bool {
	return p.tdb4Bit(0x39, 0)
}

// CanVaryOverTime reports whether AE will accept keyframes on this
// property. Most properties can; group containers and some metadata-only
// properties cannot. Read from tdb4 @0x0B bit 1.
func (p *Property) CanVaryOverTime() bool {
	return p.tdb4Bit(0x0B, 1)
}

// tdsbBit returns bit `b` of tdsb byte at offset `off`. Returns false
// when the tdsb chunk is absent or too short.
func (p *Property) tdsbBit(off, b int) bool {
	if p.tdsb == nil || len(p.tdsb.Data) <= off {
		return false
	}
	return (p.tdsb.Data[off]>>uint(b))&1 == 1
}

// LockedRatio reports whether the property's locked ratio flag is set.
// Read from tdsb @0x02 bit 4 (py-aep: byte 2 bit 4 = locked_ratio).
func (p *Property) LockedRatio() bool {
	return p.tdsbBit(0x02, 4)
}

// SetLockedRatio sets the locked ratio flag on the property.
// Writes to tdsb @0x02 bit 4 (length-preserving).
func (p *Property) SetLockedRatio(v bool) error {
	if p.tdsb == nil {
		return fmt.Errorf("property has no tdsb chunk (cannot set LockedRatio)")
	}
	if len(p.tdsb.Data) < 3 {
		return fmt.Errorf("tdsb chunk too short (len=%d)", len(p.tdsb.Data))
	}
	if v {
		p.tdsb.Data[0x02] |= 1 << 4
	} else {
		p.tdsb.Data[0x02] &^= 1 << 4
	}
	return nil
}
