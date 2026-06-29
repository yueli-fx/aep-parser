// internal/aep/move_mask_shipgate_test.go
//
// AE ship gate for MoveMask (triple-aware Mask-Atom reorder). Builds three
// masks (A/B/C) on the AE-native baseline's effect layer, moves C to the front
// via MoveMask, WriteAEP, and has AE open the mutated file: proves AE ACCEPTS a
// Mask Parade whose (tdmn "ADBE Mask Atom", mkif, atom tdgp) triples were
// re-emitted in a new order next to a real Effect Parade, reads the masks back
// as C/A/B with the effects untouched, and keeps the order across its resave.
//
// Reuses runMaskShipGate (verify_mask.jsx). Gated by AE_SHIP_GATE.
package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runMoveMaskGate(t *testing.T, aeExe, ver string) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("re_property_struct_baseline.aep not present: %v", err)
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
	// Move C (last) to the front; expected read-back order is C, A, B.
	if err := aep.MoveMask(l, l.Masks[2], 0); err != nil {
		t.Fatalf("MoveMask: %v", err)
	}
	if got := maskNames(l.Masks); got[0] != "C" {
		t.Fatalf("after move masks = %v, want C first", got)
	}

	expect := []gateMaskExpect{
		{name: "C", closed: true, vertices: rectPath().Vertices},
		{name: "A", closed: true, vertices: rectPath().Vertices},
		{name: "B", closed: true, vertices: rectPath().Vertices},
	}
	runMaskShipGate(t, aeExe, "movemask-"+ver, proj, l.ID, 3, expect)
}

func TestMoveMask_AEShipGate_AE2020(t *testing.T) { runMoveMaskGate(t, ae2020(), "AE2020") }
func TestMoveMask_AEShipGate_AE2025(t *testing.T) { runMoveMaskGate(t, ae2025(), "AE2025") }
