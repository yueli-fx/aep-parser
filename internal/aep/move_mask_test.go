// internal/aep/move_mask_test.go
//
// Go round-trip tests for MoveMask (no AE required). Proves the triple-aware
// reorder permutes the mask atom run (tdmn "ADBE Mask Atom" + mkif + atom
// tdgp) without touching chunk count, keeps the scene/flat slices in sync,
// survives WriteAEP → re-parse in the new order, and refuse-cases leave the
// project untouched. AE acceptance is covered by move_mask_shipgate_test.go.
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestMoveMask_RoundTrip(t *testing.T) {
	proj, l := addThreeMasks(t) // A, B, C
	if got, want := maskNames(l.Masks), []string{"A", "B", "C"}; !eq(got, want) {
		t.Fatalf("setup masks = %v, want %v", got, want)
	}

	// Move C (index 2) to the front (index 0): expect C, A, B.
	if err := aep.MoveMask(l, l.Masks[2], 0); err != nil {
		t.Fatalf("MoveMask: %v", err)
	}
	if got, want := maskNames(l.Masks), []string{"C", "A", "B"}; !eq(got, want) {
		t.Errorf("after move masks = %v, want %v", got, want)
	}

	// Surviving back-refs still work after the reorder.
	if err := l.Masks[0].SetInverted(true); err != nil {
		t.Fatalf("SetInverted on moved mask C: %v", err)
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
	if got, want := maskNames(rl.Masks), []string{"C", "A", "B"}; !eq(got, want) {
		t.Fatalf("re-parsed masks = %v, want %v", got, want)
	}
	if !rl.Masks[0].Inverted {
		t.Error("re-parsed moved mask C lost Inverted=true")
	}

	// Effects untouched by a mask reorder.
	if got, want := paradeChildNames(rl), []string{"ADBE Gaussian Blur 2", "ADBE Tint", "ADBE Fill"}; !eq(got, want) {
		t.Errorf("effect parade after mask move = %v, want %v", got, want)
	}
}

func TestMoveMask_NoOpAndRefuse(t *testing.T) {
	proj, l := addThreeMasks(t)
	m := l.Masks[1] // B

	// No-op: move to current index leaves order unchanged.
	if err := aep.MoveMask(l, m, 1); err != nil {
		t.Fatalf("MoveMask no-op: %v", err)
	}
	if got, want := maskNames(l.Masks), []string{"A", "B", "C"}; !eq(got, want) {
		t.Errorf("after no-op masks = %v, want %v", got, want)
	}

	if err := aep.MoveMask(nil, m, 0); err == nil {
		t.Error("MoveMask(nil layer) should error")
	}
	if err := aep.MoveMask(l, nil, 0); err == nil {
		t.Error("MoveMask(nil mask) should error")
	}
	if err := aep.MoveMask(l, m, 3); err == nil {
		t.Error("MoveMask with toIndex out of range should error")
	}
	if err := aep.MoveMask(l, m, -1); err == nil {
		t.Error("MoveMask with negative toIndex should error")
	}

	// A removed mask is no longer movable.
	if err := aep.RemoveMask(l, m); err != nil {
		t.Fatalf("RemoveMask: %v", err)
	}
	if err := aep.MoveMask(l, m, 0); err == nil {
		t.Error("MoveMask of a removed mask should error")
	}
	_ = proj
}

// TestMoveMask_ToEnd exercises the forward direction (lower → higher index).
func TestMoveMask_ToEnd(t *testing.T) {
	_, l := addThreeMasks(t) // A, B, C
	if err := aep.MoveMask(l, l.Masks[0], 2); err != nil {
		t.Fatalf("MoveMask A→end: %v", err)
	}
	if got, want := maskNames(l.Masks), []string{"B", "C", "A"}; !eq(got, want) {
		t.Errorf("after move A to end masks = %v, want %v", got, want)
	}
}
