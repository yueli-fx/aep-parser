package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// 3D-layer Orientation is stored in an otst wrapper (tdbs + otky), and its
// cdat value is little-endian (unlike normal big-endian cdat). It is a
// 3-component (X/Y/Z degrees) property even though its tdb4 dimension byte
// reads 0x01. Golden values cross-checked against py-aep
// samples/models/layer/orientation_*.json.
func TestProperty_Orientation3D(t *testing.T) {
	cases := []struct {
		fixture string
		want    []float64
	}{
		{"orientation_5_0_0", []float64{5, 0, 0}},
		{"orientation_0_279_0", []float64{0, 279, 0}},
	}
	for _, c := range cases {
		t.Run(c.fixture, func(t *testing.T) {
			proj, err := aep.Open("../../test_data/" + c.fixture + ".aep")
			if err != nil {
				t.Skipf("%s.aep not present: %v", c.fixture, err)
			}
			o := findProp(proj, "ADBE Orientation")
			if o == nil {
				t.Fatal("ADBE Orientation not found")
			}
			if o.Components != 3 {
				t.Errorf("Components = %d, want 3", o.Components)
			}
			got, ok := o.StaticValue.([]float64)
			if !ok {
				t.Fatalf("StaticValue type = %T (%v), want []float64", o.StaticValue, o.StaticValue)
			}
			if len(got) != 3 || got[0] != c.want[0] || got[1] != c.want[1] || got[2] != c.want[2] {
				t.Errorf("StaticValue = %v, want %v", got, c.want)
			}
		})
	}
}

// Animated orientation: keyframe X/Y/Z values come from otky/otda (the kfl
// ldat value slot is zero). Golden orientation_with_keyframes.json:
// kf0 = [5,0,0], kf1 = [0,0,0]. (Easing/tangents not asserted — not yet
// validated for orientation; see parseOrientationProperty.)
func TestProperty_OrientationKeyframes(t *testing.T) {
	proj, err := aep.Open("../../test_data/orientation_with_keyframes.aep")
	if err != nil {
		t.Skipf("orientation_with_keyframes.aep not present: %v", err)
	}
	o := findProp(proj, "ADBE Orientation")
	if o == nil {
		t.Fatal("ADBE Orientation not found")
	}
	if o.Components != 3 {
		t.Errorf("Components = %d, want 3", o.Components)
	}
	if len(o.Keyframes) != 2 {
		t.Fatalf("len(Keyframes) = %d, want 2", len(o.Keyframes))
	}
	want := [][]float64{{5, 0, 0}, {0, 0, 0}}
	for i, w := range want {
		got, ok := o.Keyframes[i].Value.([]float64)
		if !ok {
			t.Errorf("kf[%d].Value type = %T (%v), want []float64", i, o.Keyframes[i].Value, o.Keyframes[i].Value)
			continue
		}
		if len(got) != 3 || got[0] != w[0] || got[1] != w[1] || got[2] != w[2] {
			t.Errorf("kf[%d].Value = %v, want %v", i, got, w)
		}
	}
}
