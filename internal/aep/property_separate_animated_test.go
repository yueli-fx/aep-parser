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
// own split (speed = central-difference of values × 100, same in/out; influence
// 0.01/segmentDuration = 0.01 here for 1.0s segments / 0 at the boundary,
// bezier both sides). See TestSetDimensionsSeparated_Animated_NonUniform for
// the timing-dependent influence.
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

// TestSetDimensionsSeparated_Animated_NonUniform separates an animated 3D
// Position with NON-uniform keyframe spacing (0.5 / 1.0 / 1.5s) and asserts the
// per-axis follower influence follows 0.01/segmentDuration (0 at boundaries) —
// the timing-dependent rule that a constant-0.01 mapping would get wrong. Guards
// the central-difference + per-side-influence generalization fix.
func TestSetDimensionsSeparated_Animated_NonUniform(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_sepdim_anim2_before.aep")
	if err != nil {
		t.Skipf("re_sepdim_anim2_before.aep not present; run test_data/re_sepdim_anim2.jsx in AE 2020")
	}
	pos := findProp(proj, aep.MatchNamePosition)
	if pos == nil || pos.DimensionsSeparated() || !pos.IsAnimated() {
		t.Fatal("precondition: animated merged non-uniform fixture missing/wrong state")
	}
	if err := pos.SetDimensionsSeparated(true); err != nil {
		t.Fatalf("SetDimensionsSeparated(true): %v", err)
	}

	// segments 0.5/1.0/1.5 → influence 0.02/0.01/0.006667 on the adjacent side.
	type inf struct{ in, out float64 }
	wantInf := []inf{
		{0, 0.02},           // KF1: boundary in, seg1 out
		{0.02, 0.01},        // KF2: seg1 in, seg2 out
		{0.01, 1.0 / 150.0}, // KF3: seg2 in, seg3 out (0.01/1.5)
		{1.0 / 150.0, 0},    // KF4: seg3 in, boundary out
	}
	wantVal := map[string][]float64{
		aep.MatchNamePosition0: {0, 200, -50, 600},
		aep.MatchNamePosition1: {0, -100, 400, 50},
		aep.MatchNamePosition2: {0, 300, 100, -200},
	}

	check := func(t *testing.T, p *aep.Project, tag string) {
		t.Helper()
		for mn, vals := range wantVal {
			f := findProp(p, mn)
			if f == nil || !f.IsAnimated() {
				t.Errorf("%s: %s missing/not animated", tag, mn)
				continue
			}
			if len(f.Keyframes) != len(vals) {
				t.Errorf("%s: %s has %d kf, want %d", tag, mn, len(f.Keyframes), len(vals))
				continue
			}
			for i, kf := range f.Keyframes {
				if gv, _ := kf.Value.(float64); math.Abs(gv-vals[i]) > 1e-6 {
					t.Errorf("%s: %s kf%d value=%g want %g", tag, mn, i, gv, vals[i])
				}
				if math.Abs(kf.InTemporalEase[0].Influence-wantInf[i].in) > 1e-6 ||
					math.Abs(kf.OutTemporalEase[0].Influence-wantInf[i].out) > 1e-6 {
					t.Errorf("%s: %s kf%d influence in/out=%g/%g want %g/%g", tag, mn, i,
						kf.InTemporalEase[0].Influence, kf.OutTemporalEase[0].Influence, wantInf[i].in, wantInf[i].out)
				}
			}
		}
	}
	check(t, proj, "in-memory")
	check(t, reopenWritten(t, proj), "round-trip")
}

// TestSetDimensionsSeparated_Merge_Animated collapses the separated animated
// fixture back into a single 3D spatial Position stream and asserts the leader
// recovers its keyframe values + spatial tangents (in/out tangent = ∓follower
// speed / 100) and the per-axis followers are gone.
func TestSetDimensionsSeparated_Merge_Animated(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_sepdim_anim_after.aep")
	if err != nil {
		t.Skipf("re_sepdim_anim_after.aep not present; run test_data/re_separate_dims_anim.jsx in AE 2020")
	}
	pos := findProp(proj, aep.MatchNamePosition)
	if pos == nil {
		t.Fatal("separated: ADBE Position leader not found")
	}
	if !pos.DimensionsSeparated() {
		t.Fatal("precondition: animated fixture not separated")
	}

	if err := pos.SetDimensionsSeparated(false); err != nil {
		t.Fatalf("SetDimensionsSeparated(false) on animated: %v", err)
	}

	type leaderKF struct {
		time   float64
		value  [3]float64
		inTan  [3]float64
		outTan [3]float64
	}
	want := []leaderKF{
		{0, [3]float64{100, 200, 50}, [3]float64{-33.3333333, -33.3333333, -16.6666667}, [3]float64{33.3333333, 33.3333333, 16.6666667}},
		{1, [3]float64{300, 400, 150}, [3]float64{-66.6666667, 16.6666667, -33.3333333}, [3]float64{66.6666667, -16.6666667, 33.3333333}},
		{2, [3]float64{500, 100, 250}, [3]float64{-33.3333333, 50, -16.6666667}, [3]float64{33.3333333, -50, 16.6666667}},
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
		if !leader.IsAnimated() {
			t.Errorf("%s: merged leader not animated", tag)
		}
		for _, mn := range []string{aep.MatchNamePosition0, aep.MatchNamePosition1, aep.MatchNamePosition2} {
			if findProp(p, mn) != nil {
				t.Errorf("%s: follower %s still present after merge", tag, mn)
			}
		}
		if len(leader.Keyframes) != len(want) {
			t.Fatalf("%s: leader has %d kf, want %d", tag, len(leader.Keyframes), len(want))
		}
		for i, w := range want {
			kf := leader.Keyframes[i]
			v, ok := kf.Value.([]float64)
			if !ok || len(v) < 3 {
				t.Errorf("%s: kf%d value not 3D: %v", tag, i, kf.Value)
				continue
			}
			if math.Abs(kf.Time-w.time) > 1e-9 {
				t.Errorf("%s: kf%d time=%g, want %g", tag, i, kf.Time, w.time)
			}
			for a := 0; a < 3; a++ {
				if math.Abs(v[a]-w.value[a]) > 1e-6 {
					t.Errorf("%s: kf%d value[%d]=%g, want %g", tag, i, a, v[a], w.value[a])
				}
				if len(kf.InSpatialTangent) < 3 || math.Abs(kf.InSpatialTangent[a]-w.inTan[a]) > 1e-4 {
					t.Errorf("%s: kf%d inTan[%d]=%v, want %g", tag, i, a, kf.InSpatialTangent, w.inTan[a])
				}
				if len(kf.OutSpatialTangent) < 3 || math.Abs(kf.OutSpatialTangent[a]-w.outTan[a]) > 1e-4 {
					t.Errorf("%s: kf%d outTan[%d]=%v, want %g", tag, i, a, kf.OutSpatialTangent, w.outTan[a])
				}
			}
		}
	}

	assertMerged(t, proj, "in-memory")
	assertMerged(t, reopenWritten(t, proj), "round-trip")
}
