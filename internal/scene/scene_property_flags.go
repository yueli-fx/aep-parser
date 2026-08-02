package scene

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Property tdb4 flag readers — convenience accessors mirroring AE's
// property flag queries (is-spatial, is-animated, etc.). All read from the
// 124-byte tdb4 metadata chunk under each tdbs LIST. Byte offsets sourced
// from the reference parser's binary/property_chunks.py::Tdb4Chunk:
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
	if p.back == nil {
		return false
	}
	v, ok := p.back.Tdb4Byte(off)
	if !ok {
		return false
	}
	return (v>>uint(b))&1 == 1
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
	if p.back == nil {
		return len(p.Keyframes) > 0
	}
	v, ok := p.back.Tdb4Byte(0x44)
	if !ok {
		return len(p.Keyframes) > 0
	}
	return v != 0
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
	if p.back == nil {
		return false
	}
	v, ok := p.back.TdsbByte(off)
	if !ok {
		return false
	}
	return (v>>uint(b))&1 == 1
}

// LockedRatio reports whether the property's locked ratio flag is set.
// Read from tdsb @0x02 bit 4 (byte 2 bit 4).
func (p *Property) LockedRatio() bool {
	return p.tdsbBit(0x02, 4)
}

// @summary     Set the property's locked-ratio flag
// @param       v  true to lock the ratio, false to unlock it
// @domain      keyframe
// @stability   stable
// @verify      roundtrip
// @since       AE2020
// @boundary    length-preserving, low-risk write to tdsb @0x02 bit 4
// @alias       locked ratio,锁定比例,constrain proportions,aspect ratio lock,等比缩放,tdsb
func (p *Property) SetLockedRatio(v bool) error {
	if p.back == nil {
		return fmt.Errorf("property has no tdsb chunk (cannot set LockedRatio)")
	}
	return p.back.SetLockedRatio(v)
}

// DimensionsSeparated reports whether a multidimensional property has its
// dimensions split into separate per-axis followers (AE's "Separate
// Dimensions" on Position). Read from tdsb byte 3 (_enable_flags) bit 1;
// defaults to false when the tdsb chunk is absent. Only the leader (the
// "ADBE Position" property) carries this flag set. W (the structural
// separate/merge toggle) is deferred — it restructures the property group.
func (p *Property) DimensionsSeparated() bool {
	return p.tdsbBit(0x03, 1)
}

// separationFollowers are the per-axis component match-names AE creates when
// Position dimensions are separated (X / Y / Z). Position is the only
// property AE allows to separate; mirrors the reference parser's _SEPARATION_FOLLOWERS set.
var separationFollowers = [...]string{"ADBE Position_0", "ADBE Position_1", "ADBE Position_2"}

// IsSeparationLeader reports whether the property is the multidimensional
// leader that can be separated into per-axis followers — true for
// "ADBE Position" regardless of whether it is currently separated (use
// DimensionsSeparated for the actual state).
func (p *Property) IsSeparationLeader() bool {
	return p.MatchName == MatchNamePosition
}

// IsSeparationFollower reports whether the property is a per-axis component
// of a separated multidimensional property (X / Y / Z Position).
func (p *Property) IsSeparationFollower() bool {
	for _, mn := range separationFollowers {
		if p.MatchName == mn {
			return true
		}
	}
	return false
}

// SeparationDimension returns the axis a separation follower represents
// (0 = X, 1 = Y, 2 = Z), or -1 when the property is not a follower.
// (-1 means the property is not a follower.)
func (p *Property) SeparationDimension() int {
	for i, mn := range separationFollowers {
		if p.MatchName == mn {
			return i
		}
	}
	return -1
}

// determinePropertyTypes derives PropertyControlType and PropertyValueType
// from tdb4 flags. Port of the reference parser's _determine_property_types().
func (p *Property) determinePropertyTypes() (PropertyControlType, PropertyValueType) {
	pct := PCTLUnknown
	pvt := PVTUnknown

	if p.HasDeclaredControlType {
		pct = p.DeclaredControlType
		switch pct {
		case PCTLLayer:
			return pct, PVTLayerIndex
		case PCTLMask:
			return pct, PVTMaskIndex
		case PCTLCurve, PCTLPaintGroup:
			return pct, PVTCustomValue
		case PCTLGroup:
			return pct, PVTNoValue
		case PCTLColor:
			return pct, PVTColor
		case PCTLTwoD:
			if p.IsSpatial() {
				return pct, PVTTwoDSpatial
			}
			return pct, PVTTwoD
		case PCTLThreeD:
			if p.IsSpatial() {
				return pct, PVTThreeDSpatial
			}
			return pct, PVTThreeD
		case PCTLInteger, PCTLScalar, PCTLAngle, PCTLBoolean, PCTLEnum, PCTLSlider:
			return pct, PVTOneD
		}
	}
	if p.LayerRefPresent {
		return PCTLLayer, PVTLayerIndex
	}
	if p.Gradient != nil {
		return PCTLCurve, PVTCustomValue
	}
	if p.IsNoValue() {
		pvt = PVTNoValue
	}
	if p.IsColor() {
		if !p.HasDeclaredControlType {
			pct = PCTLColor
		}
		pvt = PVTColor
	} else if p.IsInteger() && p.Components <= 1 {
		if !p.HasDeclaredControlType {
			pct = PCTLBoolean
		}
		pvt = PVTOneD
	} else if p.IsVector() || (p.IsInteger() && p.Components > 1) {
		switch p.Components {
		case 1:
			if !p.HasDeclaredControlType {
				pct = PCTLScalar
			}
			pvt = PVTOneD
		case 2:
			if !p.HasDeclaredControlType {
				pct = PCTLTwoD
			}
			if p.IsSpatial() {
				pvt = PVTTwoDSpatial
			} else {
				pvt = PVTTwoD
			}
		case 3:
			if !p.HasDeclaredControlType {
				pct = PCTLThreeD
			}
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

// MinValue returns the minimum permitted value for the property, or nil
// if no tdum chunk is present. Type depends on property kind: float64
// for scalars, []float64 for multi-component, float64 (from uint32) for
// integer properties.
func (p *Property) MinValue() any {
	if p.back == nil {
		return nil
	}
	return p.decodeTdumValue(p.back.MinValueBytes())
}

// MaxValue returns the maximum permitted value for the property, or nil
// if no tduM chunk is present.
func (p *Property) MaxValue() any {
	if p.back == nil {
		return nil
	}
	return p.decodeTdumValue(p.back.MaxValueBytes())
}

// decodeTdumValue decodes a tdum / tduM min/max value payload into the
// property's value shape: []float64 (4 × float32) for colors, float64 (from
// uint32) for integers, else N × float64 BE.
func (p *Property) decodeTdumValue(d []byte) any {
	if len(d) == 0 {
		return nil
	}
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
		v, _ := tdumFloat64BE(d, 0)
		return v
	}
	vals := make([]float64, count)
	for i := 0; i < count; i++ {
		v, _ := tdumFloat64BE(d, i*8)
		vals[i] = v
	}
	return vals
}

func tdumFloat64BE(b []byte, offset int) (float64, bool) {
	if offset < 0 || offset+8 > len(b) {
		return 0, false
	}
	return math.Float64frombits(binary.BigEndian.Uint64(b[offset:])), true
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
// defaults to true when the tdsb chunk is absent.
func (p *Property) Enabled() bool {
	if p.back == nil {
		return true
	}
	v, ok := p.back.TdsbByte(3)
	if !ok {
		return true
	}
	return v&0x01 != 0
}

// Active is an alias for Enabled, provided as a convenience. Kept as a
// separate accessor even though it's a thin wrapper.
func (p *Property) Active() bool { return p.Enabled() }

// IsModified reports whether the property has been changed from its
// default state. A property is considered modified when it has keyframes,
// has an expression (regardless of enabled state), or its StaticValue
// differs from DefaultValue. Does not yet cover the always-modified
// special-cases (Source Text, mask-index, effect-NoValue) which depend
// on parser context we don't yet track.
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

// Elided reports whether the property is hidden from the AE UI. This is
// set for parser-synthesized placeholder groups; we don't synthesize
// anything yet, so this always returns false. Reserved for future
// implementations that materialize missing slots.
func (p *Property) Elided() bool { return false }

// IsNameSet reports whether the property has an explicit display name
// set (different from its match-name). Currently a thin proxy:
// `Name != "" && Name != MatchName`. The display name lives in the
// underlying tdsn Utf8 chunk; our parser doesn't yet decode tdsn into
// Property.Name (always defaults to MatchName), so this returns false
// for parsed properties until tdsn decode lands.
func (p *Property) IsNameSet() bool {
	return p.Name != "" && p.Name != p.MatchName
}

// IsModified reports whether any child of this group has been modified.
// For indexed groups (Effects Parade, Mask Parade), the group is
// considered modified when it has any children.
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
