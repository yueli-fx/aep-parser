package aep_test

import (
	"bytes"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// reopenWritten round-trips proj through WriteAEP → Open so tests can assert
// the serialized chunk tree re-parses to the expected scene state.
func reopenWritten(t *testing.T, proj *aep.Project) *aep.Project {
	t.Helper()
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse written bytes: %v", err)
	}
	return re
}

func scalarVal(v any) (float64, bool) {
	f, ok := v.(float64)
	return f, ok
}

// TestSetDimensionsSeparated_3D toggles Separate Dimensions on a 3D layer's
// Position leader and asserts the canonical AE 2020 byte mechanics REd in
// re_separate_dims_{before,after}.aep: leader keeps its DimensionsSeparated
// bit + resets to default [w/2,h/2,0]; the real value migrates into the
// Position_0/1/2 followers; a Position_2 follower is created for the 3D Z axis.
func TestSetDimensionsSeparated_3D(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_separate_dims_before.aep")
	if err != nil {
		t.Skipf("re_separate_dims_before.aep not present; run test_data/generators/re_separate_dims.jsx in AE 2020")
	}

	pos := findProp(proj, aep.MatchNamePosition)
	if pos == nil {
		t.Fatal("merged: ADBE Position leader not found")
	}
	if pos.DimensionsSeparated() {
		t.Fatal("precondition: merged fixture already separated")
	}
	xyz, ok := pos.StaticValue.([]float64)
	if !ok || len(xyz) < 3 {
		t.Fatalf("merged Position StaticValue = %v, want 3-component", pos.StaticValue)
	}
	wantX, wantY, wantZ := xyz[0], xyz[1], xyz[2]

	if err := aep.SetDimensionsSeparated(pos, true); err != nil {
		t.Fatalf("SetDimensionsSeparated(true): %v", err)
	}

	assertSeparated := func(t *testing.T, p *aep.Project, tag string) {
		t.Helper()
		leader := findProp(p, aep.MatchNamePosition)
		if leader == nil {
			t.Fatalf("%s: leader missing", tag)
		}
		if !leader.DimensionsSeparated() {
			t.Errorf("%s: leader DimensionsSeparated()=false, want true", tag)
		}
		def, ok := leader.DefaultValue.([]float64)
		if !ok {
			t.Fatalf("%s: leader DefaultValue not []float64: %v", tag, leader.DefaultValue)
		}
		lv, ok := leader.StaticValue.([]float64)
		if !ok || len(lv) < 3 {
			t.Fatalf("%s: leader StaticValue not 3D: %v", tag, leader.StaticValue)
		}
		for i := 0; i < 3; i++ {
			if lv[i] != def[i] {
				t.Errorf("%s: leader value[%d]=%g, want default %g", tag, i, lv[i], def[i])
			}
		}
		for mn, want := range map[string]float64{
			aep.MatchNamePosition0: wantX,
			aep.MatchNamePosition1: wantY,
			aep.MatchNamePosition2: wantZ,
		} {
			f := findProp(p, mn)
			if f == nil {
				t.Errorf("%s: follower %s not found", tag, mn)
				continue
			}
			if f.DimensionsSeparated() {
				t.Errorf("%s: follower %s reports separated", tag, mn)
			}
			got, ok := scalarVal(f.StaticValue)
			if !ok {
				t.Errorf("%s: follower %s value not scalar: %v", tag, mn, f.StaticValue)
				continue
			}
			if got != want {
				t.Errorf("%s: follower %s value=%g, want %g", tag, mn, got, want)
			}
		}
	}

	assertSeparated(t, proj, "in-memory")
	re := reopenWritten(t, proj)
	assertSeparated(t, re, "round-trip")
}

// TestSetDimensionsSeparated_Merge collapses a separated 3D Position back to
// merged: leader takes the [X,Y,Z] value, all Position_0/1/2 followers vanish.
func TestSetDimensionsSeparated_Merge(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_sepdim_merge_before.aep")
	if err != nil {
		t.Skipf("re_sepdim_merge_before.aep not present; run test_data/generators/re_separate_dims_ext.jsx in AE 2020")
	}
	pos := findProp(proj, aep.MatchNamePosition)
	if pos == nil || !pos.DimensionsSeparated() {
		t.Fatal("precondition: merge fixture not separated")
	}
	// Capture migrated follower values to predict the merged leader value.
	wantX, _ := scalarVal(findProp(proj, aep.MatchNamePosition0).StaticValue)
	wantY, _ := scalarVal(findProp(proj, aep.MatchNamePosition1).StaticValue)
	wantZ, _ := scalarVal(findProp(proj, aep.MatchNamePosition2).StaticValue)

	if err := aep.SetDimensionsSeparated(pos, false); err != nil {
		t.Fatalf("SetDimensionsSeparated(false): %v", err)
	}

	assertMerged := func(t *testing.T, p *aep.Project, tag string) {
		t.Helper()
		leader := findProp(p, aep.MatchNamePosition)
		if leader == nil {
			t.Fatalf("%s: leader missing", tag)
		}
		if leader.DimensionsSeparated() {
			t.Errorf("%s: leader still separated", tag)
		}
		lv, ok := leader.StaticValue.([]float64)
		if !ok || len(lv) < 3 {
			t.Fatalf("%s: leader value not 3D: %v", tag, leader.StaticValue)
		}
		if lv[0] != wantX || lv[1] != wantY || lv[2] != wantZ {
			t.Errorf("%s: merged leader value = %v, want [%g %g %g]", tag, lv, wantX, wantY, wantZ)
		}
		for _, mn := range []string{aep.MatchNamePosition0, aep.MatchNamePosition1, aep.MatchNamePosition2} {
			if findProp(p, mn) != nil {
				t.Errorf("%s: follower %s still present after merge", tag, mn)
			}
		}
	}
	assertMerged(t, proj, "in-memory")
	assertMerged(t, reopenWritten(t, proj), "round-trip")
}

// TestSetDimensionsSeparated_2D separates a 2D layer's Position: only X/Y
// followers appear (no Position_2 — that is 3D-only).
func TestSetDimensionsSeparated_2D(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_sepdim_2d_before.aep")
	if err != nil {
		t.Skipf("re_sepdim_2d_before.aep not present; run test_data/generators/re_separate_dims_ext.jsx in AE 2020")
	}
	pos := findProp(proj, aep.MatchNamePosition)
	if pos == nil || pos.DimensionsSeparated() {
		t.Fatal("precondition: 2D fixture missing/already separated")
	}
	xy := pos.StaticValue.([]float64)
	wantX, wantY := xy[0], xy[1]

	if err := aep.SetDimensionsSeparated(pos, true); err != nil {
		t.Fatalf("SetDimensionsSeparated(true) on 2D: %v", err)
	}

	assert2D := func(t *testing.T, p *aep.Project, tag string) {
		t.Helper()
		leader := findProp(p, aep.MatchNamePosition)
		if !leader.DimensionsSeparated() {
			t.Errorf("%s: 2D leader not separated", tag)
		}
		if x, _ := scalarVal(findProp(p, aep.MatchNamePosition0).StaticValue); x != wantX {
			t.Errorf("%s: Position_0 = %g, want %g", tag, x, wantX)
		}
		if y, _ := scalarVal(findProp(p, aep.MatchNamePosition1).StaticValue); y != wantY {
			t.Errorf("%s: Position_1 = %g, want %g", tag, y, wantY)
		}
		if findProp(p, aep.MatchNamePosition2) != nil {
			t.Errorf("%s: 2D separation must NOT create Position_2", tag)
		}
	}
	assert2D(t, proj, "in-memory")
	assert2D(t, reopenWritten(t, proj), "round-trip")
}

// TestSetDimensionsSeparated_IdempotentRefuse verifies double-separate and
// non-leader calls are refused without mutating state.
func TestSetDimensionsSeparated_Refuse(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_separate_dims_before.aep")
	if err != nil {
		t.Skipf("re_separate_dims_before.aep not present")
	}
	// Non-leader property refuses.
	op := findProp(proj, aep.MatchNameOpacity)
	if op != nil {
		if err := aep.SetDimensionsSeparated(op, true); err == nil {
			t.Error("SetDimensionsSeparated on Opacity should refuse")
		}
	}
	// Double-separate refuses.
	pos := findProp(proj, aep.MatchNamePosition)
	if err := aep.SetDimensionsSeparated(pos, true); err != nil {
		t.Fatalf("first separate: %v", err)
	}
	if err := aep.SetDimensionsSeparated(pos, true); err == nil {
		t.Error("second SetDimensionsSeparated(true) should refuse (already separated)")
	}
}
