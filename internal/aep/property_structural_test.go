package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// paradeChildNames returns the match-names of the layer's Effect Parade
// children in order, or nil if the layer has no parade.
func paradeChildNames(l *aep.Layer) []string {
	parade := l.EffectsParade()
	if parade == nil {
		return nil
	}
	out := make([]string, parade.NumProperties())
	for i := 0; i < parade.NumProperties(); i++ {
		out[i] = parade.ChildByIndex(i).PropertyMatchName()
	}
	return out
}

func effectMatchNames(l *aep.Layer) []string {
	out := make([]string, len(l.Effects))
	for i, e := range l.Effects {
		out[i] = e.MatchName
	}
	return out
}

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// layerWithEffects returns the first layer in the project that has an Effect
// Parade (AE-saved solids carry the name on the footage source, so the layer
// name parses empty — find by structure, not name).
func layerWithEffects(proj *aep.Project) *aep.Layer {
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			if p := l.EffectsParade(); p != nil && p.NumProperties() > 0 {
				return l
			}
		}
	}
	return nil
}

// aeParadeOrder reads the AE-saved fixture's Effect Parade child match-names
// for cross-checking our mutation result against AE's own output. Returns nil
// (and the test treats it as "no cross-check") when the fixture is absent.
func aeParadeOrder(t *testing.T, path string) []string {
	t.Helper()
	proj, err := aep.Open(path)
	if err != nil {
		return nil
	}
	l := layerWithEffects(proj)
	if l == nil {
		return nil
	}
	return paradeChildNames(l)
}

func loadFxLayer(t *testing.T) (*aep.Project, *aep.Layer) {
	t.Helper()
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("re_property_struct_baseline.aep not present; run test_data/re_property_struct.jsx (RE_PROP_MODE=baseline) in AE 2020")
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects found in baseline fixture")
	}
	return proj, l
}

func TestPropertyGroup_Remove_Effect(t *testing.T) {
	proj, l := loadFxLayer(t)

	want := []string{"ADBE Gaussian Blur 2", "ADBE Tint", "ADBE Fill"}
	if got := paradeChildNames(l); !eq(got, want) {
		t.Fatalf("baseline parade = %v, want %v", got, want)
	}
	if got := effectMatchNames(l); !eq(got, want) {
		t.Fatalf("baseline Layer.Effects = %v, want %v", got, want)
	}

	// Remove the middle effect (Tint).
	tint := l.EffectsParade().ChildByIndex(1).(*aep.AEPropertyGroup)
	if err := tint.Remove(); err != nil {
		t.Fatalf("Remove(Tint): %v", err)
	}

	wantAfter := []string{"ADBE Gaussian Blur 2", "ADBE Fill"}
	// Scene immediately consistent (tree + flat Effects).
	if got := paradeChildNames(l); !eq(got, wantAfter) {
		t.Errorf("post-remove parade (in-memory) = %v, want %v", got, wantAfter)
	}
	if got := effectMatchNames(l); !eq(got, wantAfter) {
		t.Errorf("post-remove Layer.Effects (in-memory) = %v, want %v", got, wantAfter)
	}

	// Round-trip: written bytes re-parse to the reduced parade.
	re := reopenWritten(t, proj)
	rl := layerWithEffects(re)
	if rl == nil {
		t.Fatal("FxLayer missing after round-trip")
	}
	if got := paradeChildNames(rl); !eq(got, wantAfter) {
		t.Errorf("round-trip parade = %v, want %v", got, wantAfter)
	}
	if got := effectMatchNames(rl); !eq(got, wantAfter) {
		t.Errorf("round-trip Layer.Effects = %v, want %v", got, wantAfter)
	}

	// Cross-check against AE's own remove output.
	if ae := aeParadeOrder(t, "../../test_data/re_property_struct_remove.aep"); ae != nil {
		if !eq(paradeChildNames(rl), ae) {
			t.Errorf("round-trip parade %v != AE-removed %v", paradeChildNames(rl), ae)
		}
	}
}

func TestPropertyGroup_MoveTo_Effect(t *testing.T) {
	proj, l := loadFxLayer(t)

	// Move the last effect (Fill, idx 2) to the front (idx 0).
	fill := l.EffectsParade().ChildByIndex(2).(*aep.AEPropertyGroup)
	if err := fill.MoveTo(0); err != nil {
		t.Fatalf("MoveTo(0): %v", err)
	}

	wantAfter := []string{"ADBE Fill", "ADBE Gaussian Blur 2", "ADBE Tint"}
	if got := paradeChildNames(l); !eq(got, wantAfter) {
		t.Errorf("post-move parade (in-memory) = %v, want %v", got, wantAfter)
	}
	if got := effectMatchNames(l); !eq(got, wantAfter) {
		t.Errorf("post-move Layer.Effects (in-memory) = %v, want %v", got, wantAfter)
	}

	re := reopenWritten(t, proj)
	rl := layerWithEffects(re)
	if got := paradeChildNames(rl); !eq(got, wantAfter) {
		t.Errorf("round-trip parade = %v, want %v", got, wantAfter)
	}

	if ae := aeParadeOrder(t, "../../test_data/re_property_struct_move.aep"); ae != nil {
		if !eq(paradeChildNames(rl), ae) {
			t.Errorf("round-trip parade %v != AE-moved %v", paradeChildNames(rl), ae)
		}
	}
}

func TestPropertyGroup_Duplicate_Effect(t *testing.T) {
	proj, l := loadFxLayer(t)

	want := []string{"ADBE Gaussian Blur 2", "ADBE Tint", "ADBE Fill"}
	if got := paradeChildNames(l); !eq(got, want) {
		t.Fatalf("baseline parade = %v, want %v", got, want)
	}

	// Duplicate the first effect (Gaussian Blur). AE inserts the clone
	// immediately after the source.
	src := l.EffectsParade().ChildByIndex(0).(*aep.AEPropertyGroup)
	clone, err := src.Duplicate()
	if err != nil {
		t.Fatalf("Duplicate(GaussianBlur): %v", err)
	}
	if clone == nil || clone == src {
		t.Fatalf("Duplicate returned %v (want a fresh, distinct clone)", clone)
	}
	if clone.MatchName != "ADBE Gaussian Blur 2" {
		t.Errorf("clone.MatchName = %q, want ADBE Gaussian Blur 2", clone.MatchName)
	}

	wantAfter := []string{"ADBE Gaussian Blur 2", "ADBE Gaussian Blur 2", "ADBE Tint", "ADBE Fill"}
	if got := paradeChildNames(l); !eq(got, wantAfter) {
		t.Errorf("post-duplicate parade (in-memory) = %v, want %v", got, wantAfter)
	}
	if got := effectMatchNames(l); !eq(got, wantAfter) {
		t.Errorf("post-duplicate Layer.Effects (in-memory) = %v, want %v", got, wantAfter)
	}
	// The clone effect must be a distinct *Effect from the source (no aliasing).
	if l.Effects[0] == l.Effects[1] {
		t.Error("clone Layer.Effects entry aliases the source effect pointer")
	}

	// Round-trip: written bytes re-parse to the grown parade.
	re := reopenWritten(t, proj)
	rl := layerWithEffects(re)
	if rl == nil {
		t.Fatal("FxLayer missing after round-trip")
	}
	if got := paradeChildNames(rl); !eq(got, wantAfter) {
		t.Errorf("round-trip parade = %v, want %v", got, wantAfter)
	}
	if got := effectMatchNames(rl); !eq(got, wantAfter) {
		t.Errorf("round-trip Layer.Effects = %v, want %v", got, wantAfter)
	}

	// Cross-check effect ORDER against AE's own duplicate output. AE persists a
	// localized display-name suffix on the clone (we don't — see
	// mutate_property_structural.go Duplicate doc); the match-name order is
	// identical.
	if ae := aeParadeOrder(t, "../../test_data/re_property_struct_duplicate.aep"); ae != nil {
		if !eq(paradeChildNames(rl), ae) {
			t.Errorf("round-trip parade %v != AE-duplicated %v", paradeChildNames(rl), ae)
		}
	}
}

func TestPropertyGroup_Duplicate_RefuseNonIndexed(t *testing.T) {
	_, l := loadFxLayer(t)

	// The Effect Parade's own parent is the layer root (NAMED), so duplicating
	// the parade itself must error — same predicate as Remove/MoveTo.
	if _, err := l.EffectsParade().Duplicate(); err == nil {
		t.Error("Duplicate() on the Effect Parade (parent not indexed) should error, got nil")
	}
}

// TestPropertyGroup_Remove_RefuseNonIndexed verifies AE's refuse contract: a
// node whose parent is not an INDEXED_GROUP cannot be removed. The Effect
// Parade's own parent is the layer root (NAMED), so removing the parade itself
// must error.
func TestPropertyGroup_Remove_RefuseNonIndexed(t *testing.T) {
	_, l := loadFxLayer(t)

	parade := l.EffectsParade()
	if !parade.IsIndexedGroup() {
		t.Fatal("Effect Parade should report IsIndexedGroup()=true")
	}
	if err := parade.Remove(); err == nil {
		t.Error("Remove() on the Effect Parade (parent not indexed) should error, got nil")
	}

	tg := l.TransformGroup()
	if tg == nil {
		t.Fatal("TransformGroup missing")
	}
	if tg.IsIndexedGroup() {
		t.Error("Transform Group should report IsIndexedGroup()=false")
	}
}
