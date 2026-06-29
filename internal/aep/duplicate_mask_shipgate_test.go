// internal/aep/duplicate_mask_shipgate_test.go
//
// AE ship gate for DuplicateMask (triple-aware Mask-Atom clone). Adds one mask
// (Orig) on the AE-native baseline's effect layer, duplicates it via
// DuplicateMask, WriteAEP, and has AE open the mutated file: proves AE ACCEPTS
// a Mask Parade with a cloned (tdmn "ADBE Mask Atom", mkif, atom tdgp) triple
// (distinct internal index) next to a real Effect Parade, reads back both
// masks (name / closed / vertices) with the effects untouched, and keeps them
// across its own resave.
//
// Reuses runMaskShipGate (verify_mask.jsx). Gated by AE_SHIP_GATE.
package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runDuplicateMaskGate(t *testing.T, aeExe, ver string) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("re_property_struct_baseline.aep not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	src, err := aep.AddMask(l, "Orig", rectPath())
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	clone, err := aep.DuplicateMask(l, src)
	if err != nil {
		t.Fatalf("DuplicateMask: %v", err)
	}
	if clone.Index == src.Index {
		t.Fatalf("clone index %d collides with source %d", clone.Index, src.Index)
	}
	if len(l.Masks) != 2 {
		t.Fatalf("after duplicate masks = %d, want 2", len(l.Masks))
	}

	expect := []gateMaskExpect{
		{name: "Orig", closed: true, vertices: rectPath().Vertices},
		{name: "Orig", closed: true, vertices: rectPath().Vertices},
	}
	runMaskShipGate(t, aeExe, "dupmask-"+ver, proj, l.ID, 3, expect)
}

func TestDuplicateMask_AEShipGate_AE2020(t *testing.T) { runDuplicateMaskGate(t, ae2020(), "AE2020") }
func TestDuplicateMask_AEShipGate_AE2025(t *testing.T) { runDuplicateMaskGate(t, ae2025(), "AE2025") }
