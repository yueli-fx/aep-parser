package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// TestCanSetCollapseTransformation_Fixture verifies py-aep capability
// query semantics against fixtures with comp / solid / file sources.
func TestCanSetCollapseTransformation_Fixture(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			// Only checking that the function returns without panicking +
			// matches expected semantics for known layer types.
			got := l.CanSetCollapseTransformation()
			src := l.AVSource()
			if src == nil {
				if got {
					t.Errorf("layer %q: no source but CanSetCollapseTransformation=true", l.Name)
				}
				continue
			}
			if _, isComp := src.(*aep.Composition); isComp && !got {
				t.Errorf("layer %q: precomp source but CanSetCollapseTransformation=false", l.Name)
			}
			if f, isFootage := src.(*aep.Footage); isFootage {
				want := f.IsSolid
				if got != want {
					t.Errorf("layer %q footage IsSolid=%v but CanSetCollapseTransformation=%v", l.Name, want, got)
				}
			}
		}
	}
}

func TestCanSetTimeRemapEnabled_Fixture(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			got := l.CanSetTimeRemapEnabled()
			src := l.AVSource()
			if src == nil {
				if got {
					t.Errorf("layer %q: no source but CanSetTimeRemapEnabled=true", l.Name)
				}
				continue
			}
			if c2, isComp := src.(*aep.Composition); isComp {
				want := c2.Duration > 0
				if got != want {
					t.Errorf("layer %q precomp Duration=%v but CanSetTimeRemapEnabled=%v", l.Name, c2.Duration, got)
				}
			}
			if f, isFootage := src.(*aep.Footage); isFootage {
				want := f.Duration > 0 && !f.IsStill
				if got != want {
					t.Errorf("layer %q footage Duration=%v IsStill=%v but CanSetTimeRemapEnabled=%v", l.Name, f.Duration, f.IsStill, got)
				}
			}
		}
	}
}

func TestCanSetCollapseTransformation_Standalone(t *testing.T) {
	l := &aep.Layer{Name: "synth"}
	if l.CanSetCollapseTransformation() {
		t.Error("standalone Layer.CanSetCollapseTransformation() = true; want false")
	}
	if l.CanSetTimeRemapEnabled() {
		t.Error("standalone Layer.CanSetTimeRemapEnabled() = true; want false")
	}
	if l.AVSource() != nil {
		t.Error("standalone Layer.AVSource() != nil")
	}
}
