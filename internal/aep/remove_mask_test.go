// internal/aep/remove_mask_test.go
//
// Go round-trip tests for RemoveMask (no AE required). Proves the triple-aware
// splice drops exactly the targeted mask atom (tdmn "ADBE Mask Atom" + mkif +
// atom tdgp), survives WriteAEP → re-parse with the surviving masks intact and
// their geometry/back-refs unaffected, and that the refuse-cases leave the
// project untouched. AE acceptance is covered separately by the ship-gate
// (remove_mask_shipgate_test.go).
package aep_test

import (
	"bytes"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// addThreeMasks adds masks A/B/C to the baseline's effect layer and returns it.
func addThreeMasks(t *testing.T) (*aep.Project, *aep.Layer) {
	t.Helper()
	proj, err := aep.Open("../../test_data/generated/fixtures/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	for _, name := range []string{"A", "B", "C"} {
		if _, err := aep.AddMask(l, name, rectPath()); err != nil {
			t.Fatalf("AddMask %q: %v", name, err)
		}
	}
	return proj, l
}

func maskNames(masks []*aep.Mask) []string {
	out := make([]string, len(masks))
	for i, m := range masks {
		out[i] = m.Name
	}
	return out
}

func TestRemoveMask_Middle_RoundTrip(t *testing.T) {
	proj, l := addThreeMasks(t)
	if got, want := maskNames(l.Masks), []string{"A", "B", "C"}; !eq(got, want) {
		t.Fatalf("setup masks = %v, want %v", got, want)
	}
	mid := l.Masks[1] // "B"

	if err := aep.RemoveMask(l, mid); err != nil {
		t.Fatalf("RemoveMask: %v", err)
	}
	if got, want := maskNames(l.Masks), []string{"A", "C"}; !eq(got, want) {
		t.Errorf("after remove masks = %v, want %v", got, want)
	}

	// The surviving masks' setters still work (back-refs untouched).
	if err := l.Masks[0].SetInverted(true); err != nil {
		t.Fatalf("SetInverted on survivor A: %v", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rl := maskParadeLayer(re, l.ID)
	if rl == nil {
		t.Fatal("re-parsed: layer not found")
	}
	if got, want := maskNames(rl.Masks), []string{"A", "C"}; !eq(got, want) {
		t.Fatalf("re-parsed masks = %v, want %v", got, want)
	}
	if !rl.Masks[0].Inverted {
		t.Error("re-parsed survivor A lost Inverted=true")
	}
	assertRectMask(t, rl.Masks[1]) // C's geometry intact

	// Effects are untouched by a mask removal.
	if got, want := paradeChildNames(rl), []string{"ADBE Gaussian Blur 2", "ADBE Tint", "ADBE Fill"}; !eq(got, want) {
		t.Errorf("effect parade after mask remove = %v, want %v", got, want)
	}
}

func TestRemoveMask_LastMask_LeavesEmptyParade(t *testing.T) {
	proj, err := aep.Open("../../test_data/generated/fixtures/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	m, err := aep.AddMask(l, "Only", rectPath())
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	if err := aep.RemoveMask(l, m); err != nil {
		t.Fatalf("RemoveMask: %v", err)
	}
	if len(l.Masks) != 0 {
		t.Fatalf("after removing only mask, layer.Masks = %d, want 0", len(l.Masks))
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rl := maskParadeLayer(re, l.ID)
	if rl == nil {
		t.Fatal("re-parsed: layer not found")
	}
	if len(rl.Masks) != 0 {
		t.Fatalf("re-parsed masks = %d, want 0", len(rl.Masks))
	}
}

func TestRemoveMask_Refuse(t *testing.T) {
	proj, l := addThreeMasks(t)
	m := l.Masks[0]

	if err := aep.RemoveMask(nil, m); err == nil {
		t.Error("RemoveMask(nil layer) should error")
	}
	if err := aep.RemoveMask(l, nil); err == nil {
		t.Error("RemoveMask(nil mask) should error")
	}

	// A mask from a different layer is not in l.Masks → refuse.
	other := layerWithEffectsExcluding(proj, l)
	if other != nil {
		om, err := aep.AddMask(other, "Foreign", rectPath())
		if err != nil {
			t.Fatalf("AddMask on other layer: %v", err)
		}
		if err := aep.RemoveMask(l, om); err == nil {
			t.Error("RemoveMask with a mask from another layer should error")
		}
	}

	// Double-remove: the second call must refuse (mask already gone).
	if err := aep.RemoveMask(l, m); err != nil {
		t.Fatalf("first RemoveMask: %v", err)
	}
	if err := aep.RemoveMask(l, m); err == nil {
		t.Error("second RemoveMask of the same mask should error")
	}
	// The other two masks are still present and intact.
	if got, want := maskNames(l.Masks), []string{"B", "C"}; !eq(got, want) {
		t.Errorf("after one remove masks = %v, want %v", got, want)
	}
}

// layerWithEffectsExcluding returns any effect-bearing layer other than skip,
// or nil if the baseline has only one.
func layerWithEffectsExcluding(proj *aep.Project, skip *aep.Layer) *aep.Layer {
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			if l != skip && len(l.Effects) > 0 {
				return l
			}
		}
	}
	return nil
}
