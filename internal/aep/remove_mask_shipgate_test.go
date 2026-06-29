// internal/aep/remove_mask_shipgate_test.go
//
// AE ship gate for RemoveMask (triple-aware Mask-Atom splice). Builds three
// masks (Keep1 / DropMe / Keep2) on the AE-native baseline's effect layer via
// AddMask, removes the MIDDLE one via RemoveMask, WriteAEP, and has AE open the
// mutated file: proves AE ACCEPTS a project whose Mask Parade had a (tdmn
// "ADBE Mask Atom", mkif, atom tdgp) triple spliced out from among siblings
// next to a real Effect Parade, reads back the two survivors (names / closed /
// vertices) with the effects untouched, and keeps them across its own resave.
//
// Reuses runMaskShipGate (verify_mask.jsx). Gated by AE_SHIP_GATE.
package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runRemoveMaskGate(t *testing.T, aeExe, ver string) {
	proj, err := aep.Open("../../test_data/generated/fixtures/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("re_property_struct_baseline.aep not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	for _, name := range []string{"Keep1", "DropMe", "Keep2"} {
		if _, err := aep.AddMask(l, name, rectPath()); err != nil {
			t.Fatalf("AddMask %q: %v", name, err)
		}
	}
	// Remove the middle mask; survivors must be Keep1, Keep2 in order.
	if l.Masks[1].Name != "DropMe" {
		t.Fatalf("expected middle mask DropMe, got %q", l.Masks[1].Name)
	}
	if err := aep.RemoveMask(l, l.Masks[1]); err != nil {
		t.Fatalf("RemoveMask: %v", err)
	}
	if len(l.Masks) != 2 {
		t.Fatalf("after remove masks = %d, want 2", len(l.Masks))
	}

	expect := []gateMaskExpect{
		{name: "Keep1", closed: true, vertices: rectPath().Vertices},
		{name: "Keep2", closed: true, vertices: rectPath().Vertices},
	}
	runMaskShipGate(t, aeExe, "removemask-"+ver, proj, l.ID, 3, expect)
}

func TestRemoveMask_AEShipGate_AE2020(t *testing.T) { runRemoveMaskGate(t, ae2020(), "AE2020") }
func TestRemoveMask_AEShipGate_AE2025(t *testing.T) { runRemoveMaskGate(t, ae2025(), "AE2025") }
