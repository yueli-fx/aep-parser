package gapledger_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aeoracle"
	"github.com/yueli-fx/aep-parser/internal/gapledger"
	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/profilediff"
)

func TestFromDiffReportMapsActionTypes(t *testing.T) {
	report := &profilediff.Report{
		SchemaVersion: profilediff.SchemaVersion,
		Diffs: []profilediff.Diff{
			testDiff("comps.by_id[1].layers.by_id[10]", profilediff.ActionWrite, profilediff.SeverityFidelity),
			testDiff("comps.by_id[1].layers.by_id[10].flags.motion_blur", profilediff.ActionSemantics, profilediff.SeverityFidelity),
			testDiff(`comps.by_id[1].layers.by_id[10].effects.by_match_name["THIRD Party"]#0`, profilediff.ActionPlugin, profilediff.SeverityBlocker),
			testDiff("comps.by_id[1].layers.by_id[10].unknown", profilediff.ActionParse, profilediff.SeverityUnknown),
			testDiff("comps.by_id[1].layers.by_id[10].name", profilediff.ActionInvestigate, profilediff.SeverityUnknown),
		},
	}

	gaps := gapledger.FromDiffReport(report, gapledger.Context{
		SourceProject: "original.aep",
		ObservedIn:    "clone.aep",
	})

	wantTypes := []gapledger.GapType{
		gapledger.TypeWriteGap,
		gapledger.TypeSemanticGap,
		gapledger.TypePluginGap,
		gapledger.TypeParseGap,
		gapledger.TypeInvestigateGap,
	}
	if len(gaps.Gaps) != len(wantTypes) {
		t.Fatalf("len(Gaps) = %d, want %d", len(gaps.Gaps), len(wantTypes))
	}
	for i, want := range wantTypes {
		got := gaps.Gaps[i]
		if got.Type != want {
			t.Fatalf("gap[%d].Type = %q, want %q", i, got.Type, want)
		}
		if got.ID == "" || got.ProfilePath == "" || got.Evidence.Source != "internal/profilediff" {
			t.Fatalf("gap[%d] missing contract fields: %+v", i, got)
		}
		if got.SourceProject != "original.aep" || got.ObservedIn != "clone.aep" {
			t.Fatalf("gap[%d] context = %+v", i, got)
		}
		if got.HumanNotes == "" {
			t.Fatalf("gap[%d].HumanNotes empty", i)
		}
	}
}

func TestFromRenderCompareSkipsExactMatch(t *testing.T) {
	report := aeoracle.CompareReport{
		SchemaVersion: aeoracle.SchemaVersion,
		TotalPixels:   4,
	}

	gaps := gapledger.FromRenderCompare(report, gapledger.Context{SourceProject: "original.png", ObservedIn: "clone.png"})
	if len(gaps.Gaps) != 0 {
		t.Fatalf("len(Gaps) = %d, want 0: %+v", len(gaps.Gaps), gaps.Gaps)
	}
}

func TestFromRenderCompareEmitsRenderGapWithMetrics(t *testing.T) {
	report := aeoracle.CompareReport{
		SchemaVersion:    aeoracle.SchemaVersion,
		ExpectedPath:     "original.png",
		ActualPath:       "clone.png",
		Width:            2,
		Height:           2,
		TotalPixels:      4,
		DifferentPixels:  1,
		DifferentPercent: 25,
		MaxChannelDelta:  9,
		ChannelThreshold: 1,
	}

	gaps := gapledger.FromRenderCompare(report, gapledger.Context{SourceProject: "original.aep", ObservedIn: "clone.aep"})
	if len(gaps.Gaps) != 1 {
		t.Fatalf("len(Gaps) = %d, want 1", len(gaps.Gaps))
	}
	gap := gaps.Gaps[0]
	if gap.Type != gapledger.TypeRenderGap || gap.ActionType != gapledger.ActionRender {
		t.Fatalf("gap type/action = %q/%q", gap.Type, gap.ActionType)
	}
	if gap.Details["different_pixels"] != 1 || gap.Details["max_channel_delta"] != uint8(9) {
		t.Fatalf("Details = %+v", gap.Details)
	}
}

func TestFromFrameSetCompareEmitsFrameGaps(t *testing.T) {
	report := aeoracle.FrameSetCompareReport{
		SchemaVersion:   aeoracle.SchemaVersion,
		ExpectedAEPPath: "original.aep",
		ActualAEPPath:   "clone.aep",
		Frames: []aeoracle.FrameCompareRecord{
			{
				Tag:          "f000024",
				Frame:        24,
				Seconds:      1,
				Reason:       "keyframe",
				ExpectedPath: "expected.png",
				ActualPath:   "actual.png",
				Status:       aeoracle.FrameStatusDifferent,
				Compare: &aeoracle.CompareReport{
					SchemaVersion:    aeoracle.SchemaVersion,
					ExpectedPath:     "expected.png",
					ActualPath:       "actual.png",
					TotalPixels:      4,
					DifferentPixels:  2,
					DifferentPercent: 50,
					MaxChannelDelta:  7,
				},
			},
			{
				Tag:          "f000048",
				Frame:        48,
				Seconds:      2,
				Reason:       "comp_end",
				ExpectedPath: "expected_end.png",
				Status:       aeoracle.FrameStatusMissingActual,
			},
		},
	}

	gaps := gapledger.FromFrameSetCompare(report, gapledger.Context{SourceProject: "original.aep", ObservedIn: "clone.aep"})
	if len(gaps.Gaps) != 2 {
		t.Fatalf("len(Gaps) = %d, want 2", len(gaps.Gaps))
	}
	if gaps.Gaps[0].Type != gapledger.TypeSemanticGap || gaps.Gaps[0].ActionType != gapledger.ActionRender {
		t.Fatalf("gap[0] type/action = %q/%q", gaps.Gaps[0].Type, gaps.Gaps[0].ActionType)
	}
	if gaps.Gaps[1].Type != gapledger.TypeWriteGap {
		t.Fatalf("gap[1].Type = %q, want write-gap", gaps.Gaps[1].Type)
	}
	if gaps.Gaps[0].Details["frame_tag"] != "f000024" || gaps.Gaps[0].Details["different_pixels"] != 2 {
		t.Fatalf("gap[0].Details = %+v", gaps.Gaps[0].Details)
	}
}

func testDiff(path string, action profilediff.ActionType, severity profilediff.Severity) profilediff.Diff {
	return profilediff.Diff{
		Path:       path,
		Kind:       profilediff.KindWrongValue,
		Severity:   severity,
		ActionType: action,
		Expected:   "expected",
		Actual:     "actual",
		Evidence: profile.Evidence{
			Level:      profile.EvidenceL1Parsed,
			Source:     "internal/profilediff",
			Confidence: "high",
		},
	}
}
