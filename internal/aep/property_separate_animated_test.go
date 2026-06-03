package aep_test

import (
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// kfExp is one expected per-axis follower keyframe after an animated Position
// is separated (values RE'd from re_sepdim_anim_after.aep; see
// incidents/separate-dimensions-write-mechanics.md §animated).
type kfExp struct {
	time     float64
	value    float64
	inSpeed  float64
	outSpeed float64
	inInf    float64
	outInf   float64
}

// TestSetDimensionsSeparated_Animated separates an animated 3D Position and
// asserts the leader collapses to its static default while each per-axis
// follower becomes an animated 1D temporal stream whose keyframes match AE's
// own split (out_speed = outSpatTan×100, in_speed = −inSpatTan×100, influence
// 0.01 on adjacent sides / 0 at the boundary, bezier both sides).
func TestSetDimensionsSeparated_Animated(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_sepdim_anim_before.aep")
	if err != nil {
		t.Skipf("re_sepdim_anim_before.aep not present; run test_data/re_separate_dims_anim.jsx in AE 2020")
	}

	pos := findProp(proj, aep.MatchNamePosition)
	if pos == nil {
		t.Fatal("merged: ADBE Position leader not found")
	}
	if pos.DimensionsSeparated() {
		t.Fatal("precondition: animated merged fixture already separated")
	}
	if !pos.IsAnimated() || len(pos.Keyframes) == 0 {
		t.Fatalf("precondition: leader not animated (kf=%d)", len(pos.Keyframes))
	}

	if err := pos.SetDimensionsSeparated(true); err != nil {
		t.Fatalf("SetDimensionsSeparated(true) on animated Position: %v", err)
	}

	// Expected per-axis follower keyframe tables (from the after fixture).
	want := map[string][]kfExp{
		aep.MatchNamePosition0: {
			{0, 100, 3333.3333334, 3333.3333334, 0, 0.01},
			{1, 300, 6666.6666668, 6666.6666668, 0.01, 0.01},
			{2, 500, 3333.3333334, 3333.3333334, 0.01, 0},
		},
		aep.MatchNamePosition1: {
			{0, 200, 3333.3333334, 3333.3333334, 0, 0.01},
			{1, 400, -1666.6666667, -1666.6666667, 0.01, 0.01},
			{2, 100, -5000.0000001, -5000.0000001, 0.01, 0},
		},
		aep.MatchNamePosition2: {
			{0, 50, 1666.6666667, 1666.6666667, 0, 0.01},
			{1, 150, 3333.3333334, 3333.3333334, 0.01, 0.01},
			{2, 250, 1666.6666667, 1666.6666667, 0.01, 0},
		},
	}

	assertSep := func(t *testing.T, p *aep.Project, tag string) {
		t.Helper()
		leader := findProp(p, aep.MatchNamePosition)
		if leader == nil {
			t.Fatalf("%s: leader missing", tag)
		}
		if !leader.DimensionsSeparated() {
			t.Errorf("%s: leader DimensionsSeparated()=false, want true", tag)
		}
		if leader.IsAnimated() || len(leader.Keyframes) != 0 {
			t.Errorf("%s: leader still animated (kf=%d), want static default", tag, len(leader.Keyframes))
		}
		def := leader.DefaultValue.([]float64)
		lv, ok := leader.StaticValue.([]float64)
		if !ok || len(lv) < 3 {
			t.Fatalf("%s: leader StaticValue not 3D: %v", tag, leader.StaticValue)
		}
		for i := 0; i < 3; i++ {
			if lv[i] != def[i] {
				t.Errorf("%s: leader value[%d]=%g, want default %g", tag, i, lv[i], def[i])
			}
		}

		for mn, kfs := range want {
			f := findProp(p, mn)
			if f == nil {
				t.Errorf("%s: follower %s not found", tag, mn)
				continue
			}
			if !f.IsAnimated() {
				t.Errorf("%s: follower %s not animated", tag, mn)
			}
			if f.DimensionsSeparated() {
				t.Errorf("%s: follower %s reports separated", tag, mn)
			}
			if len(f.Keyframes) != len(kfs) {
				t.Errorf("%s: follower %s has %d kf, want %d", tag, mn, len(f.Keyframes), len(kfs))
				continue
			}
			for i, w := range kfs {
				kf := f.Keyframes[i]
				if kf.InInterp != aep.InterpBezier || kf.OutInterp != aep.InterpBezier {
					t.Errorf("%s: %s kf%d interp=%s/%s, want bezier/bezier", tag, mn, i, kf.InInterp, kf.OutInterp)
				}
				gv, _ := kf.Value.(float64)
				if math.Abs(kf.Time-w.time) > 1e-9 || math.Abs(gv-w.value) > 1e-6 {
					t.Errorf("%s: %s kf%d time/value=%g/%g, want %g/%g", tag, mn, i, kf.Time, gv, w.time, w.value)
				}
				if len(kf.InTemporalEase) != 1 || len(kf.OutTemporalEase) != 1 {
					t.Errorf("%s: %s kf%d temporal-ease lengths in=%d out=%d, want 1/1", tag, mn, i, len(kf.InTemporalEase), len(kf.OutTemporalEase))
					continue
				}
				ie, oe := kf.InTemporalEase[0], kf.OutTemporalEase[0]
				if math.Abs(ie.Speed-w.inSpeed) > 1e-3 || math.Abs(oe.Speed-w.outSpeed) > 1e-3 {
					t.Errorf("%s: %s kf%d speed in/out=%g/%g, want %g/%g", tag, mn, i, ie.Speed, oe.Speed, w.inSpeed, w.outSpeed)
				}
				if math.Abs(ie.Influence-w.inInf) > 1e-6 || math.Abs(oe.Influence-w.outInf) > 1e-6 {
					t.Errorf("%s: %s kf%d influence in/out=%g/%g, want %g/%g", tag, mn, i, ie.Influence, oe.Influence, w.inInf, w.outInf)
				}
			}
		}
	}

	assertSep(t, proj, "in-memory")
	assertSep(t, reopenWritten(t, proj), "round-trip")
}
