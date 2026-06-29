package aep_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// findRECM returns the "RE_CM" composition from a parsed re_compmarker
// project, or fails the test.
func findRECM(t *testing.T, proj *aep.Project) *aep.Composition {
	t.Helper()
	for _, c := range proj.Compositions {
		if c.Name == "RE_CM" {
			return c
		}
	}
	t.Fatal("RE_CM comp missing from fixture")
	return nil
}

// TestMarkerRemoveRoundtrip removes one of the two composition markers in
// re_compmarker.aep and verifies the survivor round-trips through WriteAEP
// with its fields intact and the marker count dropped to one.
func TestMarkerRemoveRoundtrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_compmarker.aep")
	if err != nil {
		t.Skipf("re_compmarker.aep not present; rerun the JSX in AE")
	}
	comp := findRECM(t, proj)
	if len(comp.Markers) != 2 {
		t.Fatalf("expected 2 comp markers; got %d", len(comp.Markers))
	}

	// Remove the first marker ("comp marker A", t=1.0). The survivor is the
	// point marker "second marker" (t=2.5, chapter "chap-X").
	if err := aep.RemoveMarker(comp.Markers[0]); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if len(comp.Markers) != 1 {
		t.Fatalf("after Remove: Markers = %d, want 1", len(comp.Markers))
	}
	surv := comp.Markers[0]
	if surv.Comment != "second marker" {
		t.Errorf("survivor.Comment = %q, want %q", surv.Comment, "second marker")
	}
	if math.Abs(surv.Time-2.5) > 1e-3 {
		t.Errorf("survivor.Time = %g, want 2.5", surv.Time)
	}
	if surv.Chapter != "chap-X" {
		t.Errorf("survivor.Chapter = %q, want chap-X", surv.Chapter)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	comp2 := findRECM(t, proj2)
	if len(comp2.Markers) != 1 {
		t.Fatalf("roundtrip Markers = %d, want 1", len(comp2.Markers))
	}
	r := comp2.Markers[0]
	if r.Comment != "second marker" {
		t.Errorf("roundtrip survivor.Comment = %q", r.Comment)
	}
	if math.Abs(r.Time-2.5) > 1e-3 {
		t.Errorf("roundtrip survivor.Time = %g, want 2.5", r.Time)
	}
	if r.Chapter != "chap-X" {
		t.Errorf("roundtrip survivor.Chapter = %q", r.Chapter)
	}
}

// TestMarkerRemoveAllRoundtrip removes every composition marker and verifies
// the empty marker set round-trips to zero markers (count decremented to 0).
func TestMarkerRemoveAllRoundtrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_compmarker.aep")
	if err != nil {
		t.Skipf("re_compmarker.aep not present; rerun the JSX in AE")
	}
	comp := findRECM(t, proj)
	if len(comp.Markers) != 2 {
		t.Fatalf("expected 2 comp markers; got %d", len(comp.Markers))
	}
	// Remove both (always remove index 0 — the slice shrinks under us).
	for len(comp.Markers) > 0 {
		if err := aep.RemoveMarker(comp.Markers[0]); err != nil {
			t.Fatalf("Remove: %v", err)
		}
	}
	if len(comp.Markers) != 0 {
		t.Fatalf("after removing all: Markers = %d, want 0", len(comp.Markers))
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	comp2 := findRECM(t, proj2)
	if len(comp2.Markers) != 0 {
		t.Fatalf("roundtrip Markers = %d, want 0", len(comp2.Markers))
	}
}

// TestMarkerAddRoundtrip appends a third composition marker via clone-template,
// sets its comment, and verifies all three round-trip with the new marker a
// clean point marker (duration 0, label 0) at the requested time.
func TestMarkerAddRoundtrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_compmarker.aep")
	if err != nil {
		t.Skipf("re_compmarker.aep not present; rerun the JSX in AE")
	}
	comp := findRECM(t, proj)
	if len(comp.Markers) != 2 {
		t.Fatalf("expected 2 comp markers; got %d", len(comp.Markers))
	}

	nm, err := aep.AddMarker(comp, 4.0)
	if err != nil {
		t.Fatalf("AddMarker: %v", err)
	}
	if len(comp.Markers) != 3 {
		t.Fatalf("after AddMarker: Markers = %d, want 3", len(comp.Markers))
	}
	if nm.Duration != 0 || nm.Label != 0 || nm.Comment != "" {
		t.Errorf("new marker not clean: dur=%g label=%d comment=%q", nm.Duration, nm.Label, nm.Comment)
	}
	if err := nm.SetComment("added marker"); err != nil {
		t.Fatalf("SetComment on new marker: %v", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	comp2 := findRECM(t, proj2)
	if len(comp2.Markers) != 3 {
		t.Fatalf("roundtrip Markers = %d, want 3", len(comp2.Markers))
	}
	// Originals intact.
	if comp2.Markers[0].Comment != "comp marker A" {
		t.Errorf("roundtrip Markers[0].Comment = %q", comp2.Markers[0].Comment)
	}
	if comp2.Markers[1].Comment != "second marker" {
		t.Errorf("roundtrip Markers[1].Comment = %q", comp2.Markers[1].Comment)
	}
	// New marker (appended at the tail).
	added := comp2.Markers[2]
	if math.Abs(added.Time-4.0) > 1e-3 {
		t.Errorf("added.Time = %g, want 4.0", added.Time)
	}
	if added.Comment != "added marker" {
		t.Errorf("added.Comment = %q, want %q", added.Comment, "added marker")
	}
	if added.Duration != 0 {
		t.Errorf("added.Duration = %g, want 0 (point marker)", added.Duration)
	}
	if added.Label != 0 {
		t.Errorf("added.Label = %d, want 0", added.Label)
	}
}

// TestMarkerAddIntoEmptyRejects confirms AddMarker refuses an empty marker set
// (the canonical-seed path is a separate, not-yet-implemented slice).
func TestMarkerAddIntoEmptyRejects(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_compmarker.aep")
	if err != nil {
		t.Skipf("re_compmarker.aep not present; rerun the JSX in AE")
	}
	comp := findRECM(t, proj)
	for len(comp.Markers) > 0 {
		if err := aep.RemoveMarker(comp.Markers[0]); err != nil {
			t.Fatalf("Remove: %v", err)
		}
	}
	if _, err := aep.AddMarker(comp, 1.0); err == nil {
		t.Error("AddMarker into emptied comp: expected error")
	}
}

// TestMarkerRemoveRejectsStandalone ensures a Marker built outside the parser
// (no list back-ref) refuses Remove rather than panicking.
func TestMarkerRemoveRejectsStandalone(t *testing.T) {
	m := &aep.Marker{}
	if err := aep.RemoveMarker(m); err == nil {
		t.Error("Remove on standalone marker: expected error")
	}
}

// TestMarkerRemoveTwiceRejects ensures a marker already detached from its list
// cannot be removed a second time.
func TestMarkerRemoveTwiceRejects(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_compmarker.aep")
	if err != nil {
		t.Skipf("re_compmarker.aep not present; rerun the JSX in AE")
	}
	comp := findRECM(t, proj)
	m := comp.Markers[0]
	if err := aep.RemoveMarker(m); err != nil {
		t.Fatalf("first Remove: %v", err)
	}
	if err := aep.RemoveMarker(m); err == nil {
		t.Error("second Remove on detached marker: expected error")
	}
}
