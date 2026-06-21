package scene

import "fmt"

// Property-level scene writers: static value, expression source & enabled bit.
// Each delegates the byte-patch work to the PropertyWriter back-ref and then
// mirrors the matching scene field. The keyframe-stream ops (InsertKeyframe /
// DeleteKeyframe) are structural serializer free-functions and live in the
// serializer stage (internal/aep), not here.

// @summary    Rewrite a property's constant value in place
// @param      v  the new value (float64 for 1D, []float64 matching Components for multi-D)
// @domain     keyframe
// @stability  stable
// @verify     ae-accept
// @gate       TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk. Only valid for properties
//   without keyframes (those holding a cdat chunk).
// @alias      static value,静态值,constant value,set value,属性值,cdat
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

// @summary    Toggle whether AE evaluates the property's expression at render time
// @param      enabled  true to let AE evaluate the expression, false to keep it attached but off
// @domain     expr
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @boundary   length-preserving, 1 byte at tdb4 offset 0x77 (the disabled
//   flag; 0 = AE evaluates, 1 = expression kept but off). The neighboring
//   offset 0x78 is a separate has-expression marker kept in sync by
//   SetExpression — the two bytes are distinct and must not be conflated.
//   Requires the property's tdbs to contain a tdb4 chunk, which is always
//   present for properties parsed from real files; returns an error
//   otherwise.
// @alias      expression enabled,表达式启用,enable expression,disable expression,表达式开关,toggle expression
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

// @summary    Rewrite the JavaScript expression source attached to a property
// @param      source  the expression source text; empty string removes the expression entirely
// @domain     expr
// @stability  stable
// @verify     ae-accept
// @gate       TestExpression_AEShipGate_AE2020,TestExpression_AEShipGate_AE2025,TestExprEffect_AEShipGate_AE2020,TestExprEffect_AEShipGate_AE2025
// @since      AE2020
// @incident   expression-enable-byte-pair
// @boundary   length-variable: the underlying Utf8 chunk's data is replaced,
//   or a new Utf8 chunk is inserted into the property's tdbs LIST when none
//   existed previously. The new chunk must be inserted after the cdat chunk
//   and before the tdum/tduM chunks — getting this wrong produced a
//   false-positive pass on AE 2020 that has since been fixed and gated.
//   AE has a separate enable/disable expression toggle in addition to the
//   source text; passing "" removes the Utf8 chunk and yields the
//   no-expression state. The write path has been exercised across static,
//   cross-layer, keyframed, time-varying, and effect-parameter-reading
//   expressions; helper APIs like linear/ease/valueAtTime are not gated
//   individually since they exercise the same underlying mechanism.
// @alias      expression,表达式,js 表达式,wiggle 表达式,linkexpr
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
