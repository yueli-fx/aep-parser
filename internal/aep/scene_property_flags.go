package aep

import (
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/rifx"
)

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
	if p.back == nil || p.back.tdb4 == nil || len(p.back.tdb4.Data) <= off {
		return false
	}
	return (p.back.tdb4.Data[off]>>uint(b))&1 == 1
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
	if p.back == nil || p.back.tdb4 == nil || len(p.back.tdb4.Data) <= 0x44 {
		return len(p.Keyframes) > 0
	}
	return p.back.tdb4.Data[0x44] != 0
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
	if p.back == nil || p.back.tdsb == nil || len(p.back.tdsb.Data) <= off {
		return false
	}
	return (p.back.tdsb.Data[off]>>uint(b))&1 == 1
}

// LockedRatio reports whether the property's locked ratio flag is set.
// Read from tdsb @0x02 bit 4 (py-aep: byte 2 bit 4 = locked_ratio).
func (p *Property) LockedRatio() bool {
	return p.tdsbBit(0x02, 4)
}

// SetLockedRatio sets the locked ratio flag on the property.
// Writes to tdsb @0x02 bit 4 (length-preserving).
func (p *Property) SetLockedRatio(v bool) error {
	if p.back == nil || p.back.tdsb == nil {
		return fmt.Errorf("property has no tdsb chunk (cannot set LockedRatio)")
	}
	if len(p.back.tdsb.Data) < 3 {
		return fmt.Errorf("tdsb chunk too short (len=%d)", len(p.back.tdsb.Data))
	}
	if v {
		p.back.tdsb.Data[0x02] |= 1 << 4
	} else {
		p.back.tdsb.Data[0x02] &^= 1 << 4
	}
	return nil
}

// determinePropertyTypes derives PropertyControlType and PropertyValueType
// from tdb4 flags. Port of py-aep's _determine_property_types().
func (p *Property) determinePropertyTypes() (PropertyControlType, PropertyValueType) {
	pct := PCTLUnknown
	pvt := PVTUnknown

	if p.IsNoValue() {
		pvt = PVTNoValue
	}
	if p.IsColor() {
		pct = PCTLColor
		pvt = PVTColor
	} else if p.IsInteger() && p.Components <= 1 {
		pct = PCTLBoolean
		pvt = PVTOneD
	} else if p.IsVector() || (p.IsInteger() && p.Components > 1) {
		switch p.Components {
		case 1:
			pct = PCTLScalar
			pvt = PVTOneD
		case 2:
			pct = PCTLTwoD
			if p.IsSpatial() {
				pvt = PVTTwoDSpatial
			} else {
				pvt = PVTTwoD
			}
		case 3:
			pct = PCTLThreeD
			if p.IsSpatial() {
				pvt = PVTThreeDSpatial
			} else {
				pvt = PVTThreeD
			}
		}
	}

	return pct, pvt
}

// ControlType returns the UI control type for the property (scalar slider,
// color picker, angle dial, etc.). Derived from tdb4 flags.
func (p *Property) ControlType() PropertyControlType {
	pct, _ := p.determinePropertyTypes()
	return pct
}

// PropertyValueType returns the type of value stored in the property
// (OneD, TwoD, ThreeD, Color, NoValue, etc.). Derived from tdb4 flags.
func (p *Property) ValuePropertyType() PropertyValueType {
	_, pvt := p.determinePropertyTypes()
	return pvt
}

// decodeTdumValue reads a tdum/tduM chunk's payload. Layout depends on
// tdb4 type flags: color → 4×float32 BE, integer → 1×uint32 BE,
// otherwise N×float64 BE (N = size/8).
func (p *Property) decodeTdumValue(c *rifx.Chunk) any {
	if c == nil || len(c.Data) == 0 {
		return nil
	}
	d := c.Data
	if p.IsColor() && len(d) >= 16 {
		// 4 × float32 BE
		vals := make([]float64, 4)
		for i := 0; i < 4; i++ {
			bits := uint32(d[i*4])<<24 | uint32(d[i*4+1])<<16 | uint32(d[i*4+2])<<8 | uint32(d[i*4+3])
			vals[i] = float64(math.Float32frombits(bits))
		}
		return vals
	}
	if p.IsInteger() && len(d) >= 4 {
		v := uint32(d[0])<<24 | uint32(d[1])<<16 | uint32(d[2])<<8 | uint32(d[3])
		return float64(v)
	}
	// Default: N × float64 BE
	count := len(d) / 8
	if count == 1 {
		v, _ := readFloat64BE(d, 0)
		return v
	}
	vals := make([]float64, count)
	for i := 0; i < count; i++ {
		v, _ := readFloat64BE(d, i*8)
		vals[i] = v
	}
	return vals
}

// MinValue returns the minimum permitted value for the property, or nil
// if no tdum chunk is present. Type depends on property kind: float64
// for scalars, []float64 for multi-component, float64 (from uint32) for
// integer properties.
func (p *Property) MinValue() any {
	if p.back == nil {
		return nil
	}
	return p.decodeTdumValue(p.back.tdum)
}

// MaxValue returns the maximum permitted value for the property, or nil
// if no tduM chunk is present.
func (p *Property) MaxValue() any {
	if p.back == nil {
		return nil
	}
	return p.decodeTdumValue(p.back.tduM)
}

// UnitsText returns the text description of the units for the property
// (e.g. "pixels", "degrees", "percent", "seconds", "dB"). Returns ""
// when the property has no known unit.
func (p *Property) UnitsText() string {
	if u, ok := unitsTextMap[p.MatchName]; ok {
		return u
	}
	if p.ControlType() == PCTLAngle {
		return "degrees"
	}
	return ""
}

// PropertyIndex returns the 0-based position of this property within its
// parent AEPropertyGroup, or -1 when the property has no parent group
// (i.e. was built outside the parser or is a top-level item).
func (p *Property) PropertyIndex() int {
	if p.parentTreeGroup == nil {
		return -1
	}
	return p.parentTreeGroup.PropertyIndex(p)
}

// PropertyDepth returns the number of levels of parent groups between
// this property and the containing layer. Returns 0 for top-level
// property groups (Transform, Effects, etc.), 1 for their direct
// children, and so on. Returns -1 when the property has no parent tree
// group wired up.
func (p *Property) PropertyDepth() int {
	if p.parentTreeGroup == nil {
		return -1
	}
	return p.parentTreeGroup.Depth()
}

// Enabled reports whether the property is enabled (UI toggle next to
// the property name in AE's timeline). Read from tdsb byte 3 bit 0;
// defaults to true when the tdsb chunk is absent — matches py-aep's
// `TdsbChunk._enable_flags` default of 1.
func (p *Property) Enabled() bool {
	if p.back == nil || p.back.tdsb == nil || len(p.back.tdsb.Data) < 4 {
		return true
	}
	return p.back.tdsb.Data[3]&0x01 != 0
}

// Active is an alias for Enabled — mirrors py-aep's `property.active`
// which returns `self.enabled`. Kept as a separate accessor for parity
// even though it's a thin wrapper.
func (p *Property) Active() bool { return p.Enabled() }

// IsModified reports whether the property has been changed from its
// default state. A property is considered modified when it has keyframes,
// has an expression (regardless of enabled state), or its StaticValue
// differs from DefaultValue. Mirrors py-aep's `Property.is_modified`
// minus the always-modified special-cases (Source Text, mask-index,
// effect-NoValue) which depend on parser context we don't yet track.
func (p *Property) IsModified() bool {
	if p.IsAnimated() {
		return true
	}
	if p.Expression != "" {
		return true
	}
	if p.DefaultValue == nil {
		return false
	}
	return !valuesEqualForModified(p.StaticValue, p.DefaultValue)
}

// valuesEqualForModified compares two property value slots (any). Equal
// when both nil, or both same kind and value(s) match. []float64 are
// compared element-wise with a small epsilon to absorb float64↔float32
// round-trip noise from tdum/tduM color decoding.
func valuesEqualForModified(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch av := a.(type) {
	case float64:
		bv, ok := b.(float64)
		if !ok {
			return false
		}
		return math.Abs(av-bv) < 1e-6
	case []float64:
		bv, ok := b.([]float64)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if math.Abs(av[i]-bv[i]) > 1e-6 {
				return false
			}
		}
		return true
	}
	return false
}

// Elided reports whether the property is hidden from the AE UI. py-aep
// sets this for parser-synthesized placeholder groups; we don't
// synthesize anything yet, so this always returns false. Reserved for
// future implementations that materialize missing slots.
func (p *Property) Elided() bool { return false }

// IsNameSet reports whether the property has an explicit display name
// set (different from its match-name). Currently a thin proxy:
// `Name != "" && Name != MatchName`. py-aep checks the underlying tdsn
// Utf8 chunk; our parser doesn't yet decode tdsn into Property.Name
// (always defaults to MatchName), so this returns false for parsed
// properties until tdsn decode lands.
func (p *Property) IsNameSet() bool {
	return p.Name != "" && p.Name != p.MatchName
}

// IsModified reports whether any child of this group has been modified.
// For indexed groups (Effects Parade, Mask Parade), the group is
// considered modified when it has any children. Mirrors py-aep's
// `PropertyGroup.is_modified`.
func (g *AEPropertyGroup) IsModified() bool {
	if g == nil {
		return false
	}
	// Indexed groups: presence of children = modification.
	switch g.MatchName {
	case MatchNameGroupEffectParade, MatchNameGroupMaskParade, MatchNameGroupShapeContents:
		return len(g.Children) > 0
	}
	for _, c := range g.Children {
		switch v := c.(type) {
		case *Property:
			if v.IsModified() {
				return true
			}
		case *AEPropertyGroup:
			if v.IsModified() {
				return true
			}
		}
	}
	return false
}
