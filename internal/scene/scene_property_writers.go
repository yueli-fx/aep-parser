package scene

import "fmt"

// Property-level scene writers: static value, expression source & enabled bit.
// Each delegates the byte-patch work to the PropertyWriter back-ref and then
// mirrors the matching scene field. The keyframe-stream ops (InsertKeyframe /
// DeleteKeyframe) are structural serializer free-functions and live in the
// serializer stage (internal/aep), not here.

// SetStaticValue rewrites a property's constant value in-place (only valid
// for properties without keyframes — those with a cdat chunk).
//
//aep:cap domain=keyframe tier=stable verify=roundtrip boundary="length-preserving 低风险;仅适用于无关键帧属性(cdat chunk);无专门 AE gate→round-trip" alias="static value,静态值,constant value,set value,属性值,cdat"
func (p *Property) SetStaticValue(v any) error {
	if p.back == nil {
		return fmt.Errorf("property %q: no static-value chunk (has keyframes?)", p.MatchName)
	}
	switch x := v.(type) {
	case float64:
		if p.Components != 1 {
			return fmt.Errorf("property %q is %dD, expected []float64", p.MatchName, p.Components)
		}
	case []float64:
		if len(x) != p.Components {
			return fmt.Errorf("property %q: got %d components, property is %dD", p.MatchName, len(x), p.Components)
		}
	default:
		return fmt.Errorf("property: unsupported value type %T", v)
	}
	if err := p.back.SetStaticValue(v); err != nil {
		return err
	}
	switch x := v.(type) {
	case float64:
		p.StaticValue = x
	case []float64:
		p.StaticValue = append([]float64(nil), x...)
	}
	return nil
}

// SetExpressionEnabled toggles whether AE evaluates the property's
// expression at render time (separate knob from `SetExpression` which
// writes the JS source itself).
//
// RE'd against an AE-2025-native enabled/disabled fixture pair (expr_re,
// 2026-06-12): tdb4 byte @0x77 is the disabled flag (0 = AE evaluates,
// 1 = expression kept but off); the neighbouring @0x78 is a has-expression
// marker kept in sync by SetExpression. (The historic reading of @0x78 as
// an inverted enabled byte conflated the two — it made every SetExpression
// output render-dead, and writing @0x78=0 for "disabled" made AE drop the
// expression text entirely.) length-preserving (1 byte).
//
// Requires the property's tdbs to contain a `tdb4` chunk (always
// present for properties parsed from real .aep files). Returns an
// error otherwise.
//
//aep:cap domain=expr tier=stable verify=roundtrip boundary="length-preserving 低风险;tdb4 @0x77 1 byte;无专门 AE gate→round-trip;仅对有 tdbs 的属性有效" alias="expression enabled,表达式启用,enable expression,disable expression,表达式开关,toggle expression"
func (p *Property) SetExpressionEnabled(enabled bool) error {
	if p.back == nil {
		return fmt.Errorf("property %q: no tdbs reference", p.MatchName)
	}
	if err := p.back.SetExpressionEnabled(enabled); err != nil {
		return err
	}
	p.ExpressionEnabled = enabled
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
//
//aep:cap domain=expr tier=stable verify=ae-accept gate=TestExpression_AEShipGate_AE2020,TestExpression_AEShipGate_AE2025,TestExprEffect_AEShipGate_AE2020,TestExprEffect_AEShipGate_AE2025 incident=expression-enable-byte-pair boundary="length-variable;Utf8 须插 cdat 后/tdum-tduM 前(曾是 AE2020 假绿坑,已修+gated);单值表达式 round-trip,复杂引用未逐一验" alias="expression,表达式,js 表达式,wiggle 表达式,linkexpr"
func (p *Property) SetExpression(source string) error {
	if p.back == nil {
		return fmt.Errorf("property %q: no tdbs reference (built outside parser?)", p.MatchName)
	}
	if err := p.back.SetExpression(source); err != nil {
		return err
	}
	p.Expression = source
	return nil
}
