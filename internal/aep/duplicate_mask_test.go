// internal/aep/duplicate_mask_test.go
//
// Go round-trip tests for DuplicateMask (no AE required). Proves the triple
// clone (tdmn "ADBE Mask Atom" + mkif + atom tdgp) is inserted right after the
// source with a bumped internal index, the clone's geometry/flags match the
// source, the result survives WriteAEP → re-parse, and refuse-cases leave the
// project untouched. AE acceptance is covered by duplicate_mask_shipgate_test.go.
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestDuplicateMask_RoundTrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	src, err := aep.AddMask(l, "Orig", rectPath())
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	if err := src.SetInverted(true); err != nil {
		t.Fatalf("SetInverted: %v", err)
	}

	clone, err := aep.DuplicateMask(l, src)
	if err != nil {
		t.Fatalf("DuplicateMask: %v", err)
	}
	if clone == nil {
		t.Fatal("DuplicateMask returned nil")
	}
	// Clone sits right after the source, carries the source's name/flags, and a
	// distinct internal index.
	if got, want := maskNames(l.Masks), []string{"Orig", "Orig"}; !eq(got, want) {
		t.Errorf("masks = %v, want %v", got, want)
	}
	if l.Masks[1] != clone {
		t.Error("clone should be inserted at index 1 (right after source)")
	}
	if clone.Index == src.Index {
		t.Errorf("clone index %d equals source index %d (should be bumped)", clone.Index, src.Index)
	}
	if !clone.Inverted {
		t.Error("clone should inherit Inverted=true from source")
	}

	// Clone's setters work immediately (fresh back-refs, not aliased).
	if err := clone.SetInverted(false); err != nil {
		t.Fatalf("SetInverted on clone: %v", err)
	}
	if !src.Inverted {
		t.Error("mutating the clone must not affect the source (chunk aliasing)")
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
	if rl == nil || len(rl.Masks) != 2 {
		t.Fatalf("re-parsed masks = %v, want 2", rl)
	}
	if rl.Masks[0].Index == rl.Masks[1].Index {
		t.Errorf("re-parsed mask indexes collide: %d == %d", rl.Masks[0].Index, rl.Masks[1].Index)
	}
	if !rl.Masks[0].Inverted || rl.Masks[1].Inverted {
		t.Errorf("re-parsed inverted = %v, %v; want true (source), false (clone)", rl.Masks[0].Inverted, rl.Masks[1].Inverted)
	}
	assertRectMask(t, rl.Masks[1]) // clone geometry intact
}

func TestDuplicateMask_Refuse(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	m, err := aep.AddMask(l, "M", rectPath())
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}

	if _, err := aep.DuplicateMask(nil, m); err == nil {
		t.Error("DuplicateMask(nil layer) should error")
	}
	if _, err := aep.DuplicateMask(l, nil); err == nil {
		t.Error("DuplicateMask(nil mask) should error")
	}
	// A removed mask is no longer in l.Masks → refuse.
	if err := aep.RemoveMask(l, m); err != nil {
		t.Fatalf("RemoveMask: %v", err)
	}
	if _, err := aep.DuplicateMask(l, m); err == nil {
		t.Error("DuplicateMask of a removed mask should error")
	}
}
