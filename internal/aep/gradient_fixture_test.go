package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// TestGradient_FixturePyAep verifies Property.Gradient is populated when
// parsing AE's "ADBE Vector Grad Colors" GCst LIST → GCky → Utf8 (XML)
// chunk path, using py-aep's gradient.aep sample as the reference
// fixture. py-aep's gradient.json gives the expected schema (color
// stops + alpha stops; version "4").
func TestGradient_FixturePyAep(t *testing.T) {
	path := "../../workshop/reference/py-aep/samples/models/property/gradient.aep"
	proj, err := aep.Open(path)
	if err != nil {
		t.Skipf("py-aep gradient.aep not present: %v", err)
	}

	var gradProps []*aep.Property
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			for _, p := range l.Properties {
				if p.MatchName == "ADBE Vector Grad Colors" {
					gradProps = append(gradProps, p)
				}
			}
		}
	}
	if len(gradProps) == 0 {
		t.Fatal("no 'ADBE Vector Grad Colors' property found in fixture")
	}

	// py-aep dump shows two G-* properties (one stroke, one fill).
	if len(gradProps) < 2 {
		t.Errorf("expected ≥ 2 gradient props, got %d", len(gradProps))
	}

	for i, p := range gradProps {
		if p.Gradient == nil {
			t.Errorf("gradient[%d]: prop.Gradient = nil; expected populated from GCky Utf8 XML", i)
			continue
		}
		if p.Gradient.Version == "" {
			t.Errorf("gradient[%d]: empty version", i)
		}
		if len(p.Gradient.ColorStops) == 0 {
			t.Errorf("gradient[%d]: no color stops decoded", i)
		}
		// Every color stop has offset in [0,1] and a 3-component color.
		for j, cs := range p.Gradient.ColorStops {
			if cs.Offset < 0 || cs.Offset > 1 {
				t.Errorf("gradient[%d].ColorStops[%d].Offset = %v, want 0..1", i, j, cs.Offset)
			}
		}
		// Alpha stops are optional but the fixture has them.
		if len(p.Gradient.AlphaStops) == 0 {
			t.Errorf("gradient[%d]: no alpha stops decoded (py-aep sample has them)", i)
		}
	}
}
