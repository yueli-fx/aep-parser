package aeoracle_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aeoracle"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func TestSelectFramesIncludesEndpointsKeyframesAndMidpoints(t *testing.T) {
	prof := frameProfile([]profile.Keyframe{
		{Time: 1.0, Value: 10},
		{Time: 3.0, Value: 20},
	})

	frames, err := aeoracle.SelectFrames(prof, aeoracle.FrameOptions{CompName: "Main", MaxFrames: 8})
	if err != nil {
		t.Fatalf("SelectFrames: %v", err)
	}

	want := []int{0, 15, 30, 60, 90, 105, 120}
	if got := frameIndexes(frames); !equalInts(got, want) {
		t.Fatalf("frames = %v, want %v; full=%+v", got, want, frames)
	}
	assertReason(t, frames, 0, "comp_start")
	assertReason(t, frames, 30, "keyframe")
	assertReason(t, frames, 60, "quiet_midpoint")
	assertReason(t, frames, 120, "comp_end")
}

func TestSelectFramesCapsDeterministicallyAndKeepsEndpoints(t *testing.T) {
	prof := frameProfile([]profile.Keyframe{
		{Time: 0.5, Value: 1},
		{Time: 1.0, Value: 2},
		{Time: 2.0, Value: 3},
		{Time: 3.0, Value: 4},
	})

	frames, err := aeoracle.SelectFrames(prof, aeoracle.FrameOptions{CompName: "Main", MaxFrames: 4})
	if err != nil {
		t.Fatalf("SelectFrames: %v", err)
	}

	want := []int{0, 15, 30, 120}
	if got := frameIndexes(frames); !equalInts(got, want) {
		t.Fatalf("frames = %v, want %v; full=%+v", got, want, frames)
	}
	assertReason(t, frames, 0, "comp_start")
	assertReason(t, frames, 120, "comp_end")
}

func TestSelectFramesDefaultsToFirstComp(t *testing.T) {
	prof := frameProfile(nil)

	frames, err := aeoracle.SelectFrames(prof, aeoracle.FrameOptions{MaxFrames: 3})
	if err != nil {
		t.Fatalf("SelectFrames: %v", err)
	}
	if got := frameIndexes(frames); !equalInts(got, []int{0, 60, 120}) {
		t.Fatalf("frames = %v, want [0 60 120]", got)
	}
}

func frameProfile(keyframes []profile.Keyframe) *profile.Profile {
	return &profile.Profile{
		SchemaVersion: profile.SchemaVersion,
		Comps: []profile.Composition{{
			ID:        1,
			Name:      "Main",
			FrameRate: 30,
			Duration:  4.05,
			Layers: []profile.Layer{{
				ID:   10,
				Name: "Layer",
				Type: "av",
				Timing: profile.LayerTiming{
					InPoint:  0,
					OutPoint: 4,
				},
				Flags: profile.LayerFlags{Visible: true},
				Properties: []profile.Property{{
					MatchName: "ADBE Position",
					Keyframes: keyframes,
				}},
			}},
		}},
	}
}

func frameIndexes(frames []aeoracle.FrameTarget) []int {
	out := make([]int, 0, len(frames))
	for _, frame := range frames {
		out = append(out, frame.Frame)
	}
	return out
}

func assertReason(t *testing.T, frames []aeoracle.FrameTarget, frame int, reason string) {
	t.Helper()
	for _, got := range frames {
		if got.Frame == frame && got.Reason == reason {
			return
		}
	}
	t.Fatalf("frame %d reason %q not found in %+v", frame, reason, frames)
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
