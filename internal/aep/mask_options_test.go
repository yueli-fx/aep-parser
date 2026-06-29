// internal/aep/mask_options_test.go
//
// Go round-trip tests for Mask Feather / Opacity / Expansion (SetFeather /
// SetOpacity / SetExpansion) — the default-elided option leaves materialized via
// synthesis-insert into the mask atom group. Proves the spliced leaves survive
// WriteAEP → re-parse with the requested values. AE acceptance + render are
// covered by the ship-gate (mg_mask_opacity_shipgate_test.go).
package aep_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestMaskOptions_RoundTrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}

	m, err := aep.AddMask(l, "Opt Mask", rectPath())
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	// Setters work right after AddMask (parseMasks wired the atom back-ref).
	if err := m.SetOpacity(0.5); err != nil {
		t.Fatalf("SetOpacity: %v", err)
	}
	if err := m.SetFeather([2]float64{20, 30}); err != nil {
		t.Fatalf("SetFeather: %v", err)
	}
	if err := m.SetExpansion(15); err != nil {
		t.Fatalf("SetExpansion: %v", err)
	}
	if m.Opacity != 0.5 || m.Feather != [2]float64{20, 30} || m.Expansion != 15 {
		t.Errorf("scene fields = op %v feather %v exp %v", m.Opacity, m.Feather, m.Expansion)
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
	if rl == nil || len(rl.Masks) != 1 {
		t.Fatalf("re-parsed masks = %v, want 1", rl)
	}
	rm := rl.Masks[0]
	if math.Abs(rm.Opacity-0.5) > 1e-9 {
		t.Errorf("re-parsed opacity = %v, want 0.5", rm.Opacity)
	}
	if math.Abs(rm.Feather[0]-20) > 1e-9 || math.Abs(rm.Feather[1]-30) > 1e-9 {
		t.Errorf("re-parsed feather = %v, want [20,30]", rm.Feather)
	}
	if math.Abs(rm.Expansion-15) > 1e-9 {
		t.Errorf("re-parsed expansion = %v, want 15", rm.Expansion)
	}
}

// Idempotent overwrite: setting opacity twice must not splice a second leaf.
func TestMaskOptions_OverwriteNoDuplicate(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	m, err := aep.AddMask(l, "Opt Mask", rectPath())
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	if err := m.SetOpacity(0.5); err != nil {
		t.Fatalf("SetOpacity #1: %v", err)
	}
	if err := m.SetOpacity(0.25); err != nil {
		t.Fatalf("SetOpacity #2: %v", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rm := maskParadeLayer(re, l.ID).Masks[0]
	if math.Abs(rm.Opacity-0.25) > 1e-9 {
		t.Errorf("re-parsed opacity = %v, want 0.25 (overwrite, not duplicate)", rm.Opacity)
	}
}
