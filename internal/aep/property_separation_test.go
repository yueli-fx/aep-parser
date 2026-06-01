package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// findProp walks every comp/layer for the first property with matchName.
func findProp(proj *aep.Project, matchName string) *aep.Property {
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			if p := l.PropertyByMatchName(matchName); p != nil {
				return p
			}
		}
	}
	return nil
}

// Golden cross-checked against py-aep transform_separated.json: the
// "ADBE Position" leader carries dimensionsSeparated=true and the
// ADBE Position_0/1/2 followers carry separationDimension 0/1/2.
//
// NOTE: this fixture is a 3D layer; our parser currently truncates its
// Transform Group after Position_1 (Position_2 / Scale / Rotate Z / Opacity
// dropped, Orientation mis-parsed as a group) — see
// incidents/transform-group-3d-truncation.md. The assertions below only
// touch the properties our parser surfaces (Position leader + Position_0/1).
func TestProperty_DimensionsSeparated(t *testing.T) {
	sep, err := aep.Open("../../test_data/transform_separated.aep")
	if err != nil {
		t.Skipf("transform_separated.aep not present: %v", err)
	}

	pos := findProp(sep, aep.MatchNamePosition)
	if pos == nil {
		t.Fatal("separated: ADBE Position not found")
	}
	if !pos.DimensionsSeparated() {
		t.Error("separated Position DimensionsSeparated() = false, want true")
	}
	if !pos.IsSeparationLeader() {
		t.Error("Position IsSeparationLeader() = false, want true")
	}
	if pos.IsSeparationFollower() {
		t.Error("Position IsSeparationFollower() = true, want false")
	}

	// Followers (X/Y) carry the per-axis dimension and are not leaders; their
	// own dimensionsSeparated bit is clear.
	for i, mn := range []string{"ADBE Position_0", "ADBE Position_1"} {
		f := findProp(sep, mn)
		if f == nil {
			t.Errorf("follower %s not found", mn)
			continue
		}
		if !f.IsSeparationFollower() {
			t.Errorf("%s IsSeparationFollower() = false, want true", mn)
		}
		if got := f.SeparationDimension(); got != i {
			t.Errorf("%s SeparationDimension() = %d, want %d", mn, got, i)
		}
		if f.IsSeparationLeader() {
			t.Errorf("%s IsSeparationLeader() = true, want false", mn)
		}
		if f.DimensionsSeparated() {
			t.Errorf("%s DimensionsSeparated() = true, want false", mn)
		}
	}
}

func TestProperty_SeparationFallback(t *testing.T) {
	// Bare property (no tdsb, ordinary match-name): all separation
	// accessors return their false/-1 defaults.
	p := &aep.Property{MatchName: "ADBE Opacity"}
	if p.DimensionsSeparated() {
		t.Error("DimensionsSeparated on bare Property = true, want false")
	}
	if p.IsSeparationLeader() || p.IsSeparationFollower() {
		t.Error("non-Position property reported as leader/follower")
	}
	if p.SeparationDimension() != -1 {
		t.Errorf("SeparationDimension = %d, want -1", p.SeparationDimension())
	}

	// Position_2 is a valid follower (dimension 2) even though the current
	// parser drops it from the 3D fixture — verify the pure match-name logic.
	p2 := &aep.Property{MatchName: "ADBE Position_2"}
	if !p2.IsSeparationFollower() || p2.SeparationDimension() != 2 {
		t.Errorf("Position_2 follower=%v dim=%d, want true 2", p2.IsSeparationFollower(), p2.SeparationDimension())
	}
}
