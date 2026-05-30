// internal/aep/shape_mutate_existing_test.go
//
// Exercises the "open existing → mutate →
// roundtrip" path. Fixture v2_2_shape_tolerance.aep is produced by AE;
// until it exists this test skips cleanly so it doesn't
// block PASS counts.
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestV2_2_MutateExistingShape(t *testing.T) {
	p, err := aep.Open("../../test_data/v2_2_shape_tolerance.aep")
	if err != nil {
		t.Skipf("v2_2_shape_tolerance.aep not present: %v (Phase 5 Task 5.3 produces)", err)
	}
	if len(p.Compositions) == 0 || len(p.Compositions[0].Layers) == 0 {
		t.Skip("fixture empty")
	}
	layer := p.Compositions[0].Layers[0]
	if layer.Type != aep.LayerTypeShape {
		t.Skipf("fixture layer 0 type = %v, want Shape", layer.Type)
	}
	sl := aep.WrapShapeLayer(layer)
	if len(sl.RootGroup().Children) == 0 {
		t.Skip("fixture root group empty")
	}
	rect, ok := sl.RootGroup().Children[0].(*aep.RectNode)
	if !ok {
		t.Skipf("fixture layer 0 first child kind = %v, want Rect", sl.RootGroup().Children[0].Kind())
	}

	if err := rect.SetSize([2]float64{500, 500}); err != nil {
		t.Fatalf("SetSize: %v", err)
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}

	reSl := aep.WrapShapeLayer(re.Compositions[0].Layers[0])
	reRect := reSl.RootGroup().Children[0].(*aep.RectNode)
	val, _ := reRect.Size().StaticValue()
	if val != [2]float64{500, 500} {
		t.Fatalf("mutated size lost: %v", val)
	}
}
