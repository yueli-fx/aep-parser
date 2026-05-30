// internal/aep/shape_preservation_test.go
//
// Tier 3 nested-group semantic preservation.
//
// Fixture `test_data/v2_2_shape_tolerance.aep` is produced
// via tmp_debug/gen_shape_tolerance.jsx (requires AE). Until
// the fixture lands the test SKIPs cleanly per V1 pattern.
//
// Semantic-level assertions (minimum preservation set):
//   - layer count unchanged across roundtrip
//   - layer name unchanged across roundtrip
//   - V1 ShapePrimitive count unchanged (proxy for "nested group topology
//     intact" without needing per-byte chunk-tree walk — opaque chunk
//     fidelity is covered separately)
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestV2_2_NestedGroup_Preservation(t *testing.T) {
	p, err := aep.Open("../../test_data/v2_2_shape_tolerance.aep")
	if err != nil {
		t.Skipf("fixture v2_2_shape_tolerance.aep missing: %v (produced via AE)", err)
	}
	if len(p.Compositions) == 0 || len(p.Compositions[0].Layers) == 0 {
		t.Skip("fixture empty")
	}

	preLayerCount := len(p.Compositions[0].Layers)
	preLayerName := p.Compositions[0].Layers[0].Name
	prePrimCount := len(p.Compositions[0].Layers[0].ShapePrimitives)

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}

	if got := len(re.Compositions[0].Layers); got != preLayerCount {
		t.Fatalf("layer count changed: %d → %d", preLayerCount, got)
	}
	if got := re.Compositions[0].Layers[0].Name; got != preLayerName {
		t.Fatalf("layer name changed: %q → %q", preLayerName, got)
	}
	if got := len(re.Compositions[0].Layers[0].ShapePrimitives); got != prePrimCount {
		t.Fatalf("ShapePrimitive count changed: %d → %d (nested group topology damaged)", prePrimCount, got)
	}
}
