package aep_test

import (
	"bytes"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// TestPropertySetExpression covers length-variable expression source
// writes — set / replace / clear, with WriteAEP round-trip.
func TestPropertySetExpression(t *testing.T) {
	textPayload := []byte("OPAQUE-PAYLOAD")
	data := buildExtendedAEP(textPayload, nil, nil)
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	layer := proj.Compositions[0].Layers[0]
	op := layer.Opacity()
	if op == nil {
		t.Fatal("Opacity property missing")
	}
	if op.Expression != "time*2" {
		t.Fatalf("baseline Expression = %q, want %q", op.Expression, "time*2")
	}

	const newExpr = "wiggle(2, 30)"
	if err := op.SetExpression(newExpr); err != nil {
		t.Fatalf("SetExpression replace: %v", err)
	}
	if op.Expression != newExpr {
		t.Errorf("in-memory after replace: Expression = %q", op.Expression)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	op2 := proj2.Compositions[0].Layers[0].Opacity()
	if op2 == nil || op2.Expression != newExpr {
		t.Errorf("after roundtrip: Expression = %q, want %q", op2.Expression, newExpr)
	}

	if err := op2.SetExpression(""); err != nil {
		t.Fatalf("SetExpression clear: %v", err)
	}
	if op2.Expression != "" {
		t.Errorf("after clear in-mem: Expression = %q", op2.Expression)
	}

	var buf2 bytes.Buffer
	if err := proj2.WriteAEP(&buf2); err != nil {
		t.Fatalf("WriteAEP after clear: %v", err)
	}
	proj3, err := aep.FromReader(bytes.NewReader(buf2.Bytes()))
	if err != nil {
		t.Fatalf("re-parse after clear: %v", err)
	}
	op3 := proj3.Compositions[0].Layers[0].Opacity()
	if op3 != nil && op3.Expression != "" {
		t.Errorf("after roundtrip-clear: Expression = %q", op3.Expression)
	}

	// Add an expression to a property that didn't have one — use a
	// fresh property from the synthetic keyframed AEP.
	freshData := buildKeyframedAEP()
	freshProj, err := aep.FromReader(bytes.NewReader(freshData))
	if err != nil {
		t.Fatalf("fresh FromReader: %v", err)
	}
	pos := freshProj.Compositions[0].Layers[0].Position()
	if pos == nil {
		t.Fatal("fresh Position missing")
	}
	if pos.Expression != "" {
		t.Fatalf("fresh Position expression should start empty; got %q", pos.Expression)
	}
	if err := pos.SetExpression("[value[0]+10, value[1], value[2]]"); err != nil {
		t.Fatalf("SetExpression insert: %v", err)
	}

	var bufFresh bytes.Buffer
	if err := freshProj.WriteAEP(&bufFresh); err != nil {
		t.Fatalf("WriteAEP fresh: %v", err)
	}
	freshProj2, err := aep.FromReader(bytes.NewReader(bufFresh.Bytes()))
	if err != nil {
		t.Fatalf("re-parse fresh: %v", err)
	}
	pos2 := freshProj2.Compositions[0].Layers[0].Position()
	if pos2 == nil || pos2.Expression != "[value[0]+10, value[1], value[2]]" {
		t.Errorf("fresh roundtrip: Position.Expression = %q", pos2.Expression)
	}
}

// TestPropertySetExpressionEnabled exercises the tdb4 @0x78 inverted
// disabled-flag toggle against the AE 2020 fixture re_batch2.aep.
// Skips when the fixture is absent.
func TestPropertySetExpressionEnabled(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_batch2.aep")
	if err != nil {
		t.Skipf("re_batch2.aep not present; rerun /tmp/re_batch2.jsx in AE")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_B2_A_exprEnabled" {
			comp = c
			break
		}
	}
	if comp == nil || len(comp.Layers) == 0 {
		t.Fatal("expected RE_B2_A_exprEnabled comp with at least one layer")
	}
	op := comp.Layers[0].Opacity()
	if op == nil || op.Expression == "" {
		t.Fatal("Opacity with expression missing")
	}
	// Toggle both ways and verify Go-field tracking; don't assume the
	// fixture's starting state since the JSX that built it could go either way.
	if err := op.SetExpressionEnabled(false); err != nil {
		t.Fatalf("SetExpressionEnabled(false): %v", err)
	}
	if op.ExpressionEnabled {
		t.Errorf("after SetExpressionEnabled(false): ExpressionEnabled = true")
	}
	if err := op.SetExpressionEnabled(true); err != nil {
		t.Fatalf("SetExpressionEnabled(true): %v", err)
	}
	if !op.ExpressionEnabled {
		t.Errorf("after SetExpressionEnabled(true): ExpressionEnabled = false")
	}
	// Final state for the roundtrip assertion: enabled.
	if err := op.SetExpressionEnabled(true); err != nil {
		t.Fatalf("SetExpressionEnabled(true) #2: %v", err)
	}
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	for _, c := range proj2.Compositions {
		if c.Name != "RE_B2_A_exprEnabled" || len(c.Layers) == 0 {
			continue
		}
		got := c.Layers[0].Opacity()
		if got == nil || got.Expression != "time * 50" {
			t.Errorf("expression source corrupted after toggle: %q", got.Expression)
		}
		if got != nil && !got.ExpressionEnabled {
			t.Errorf("roundtrip ExpressionEnabled = false, want true (final enabled state should persist)")
		}
		break
	}
}
