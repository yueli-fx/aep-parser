package gapledger

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"github.com/example/aep-parser/internal/aeoracle"
	"github.com/example/aep-parser/internal/profile"
	"github.com/example/aep-parser/internal/profilediff"
)

const SchemaVersion = 1

type GapType string

const (
	TypeParseGap       GapType = "parse-gap"
	TypeWriteGap       GapType = "write-gap"
	TypeSemanticGap    GapType = "semantic-gap"
	TypeRenderGap      GapType = "render-gap"
	TypePluginGap      GapType = "plugin-gap"
	TypeAssetGap       GapType = "asset-gap"
	TypeLibBlocked     GapType = "lib-blocked"
	TypeInvestigateGap GapType = "investigate-gap"
)

type ActionType string

const (
	ActionParse       ActionType = "parse"
	ActionWrite       ActionType = "write"
	ActionSemantics   ActionType = "semantics"
	ActionRender      ActionType = "render"
	ActionAsset       ActionType = "asset"
	ActionPlugin      ActionType = "plugin"
	ActionDocs        ActionType = "docs"
	ActionKnowledge   ActionType = "knowledge"
	ActionInvestigate ActionType = "investigate"
)

type Severity string

const (
	SeverityBlocker    Severity = "blocker"
	SeverityFidelity   Severity = "fidelity"
	SeverityPolish     Severity = "polish"
	SeverityAcceptable Severity = "acceptable"
	SeverityUnknown    Severity = "unknown"
)

type Context struct {
	SourceProject string
	ObservedIn    string
}

type Report struct {
	SchemaVersion int    `json:"schema_version"`
	SourceProject string `json:"source_project,omitempty"`
	ObservedIn    string `json:"observed_in,omitempty"`
	GapCount      int    `json:"gap_count"`
	Gaps          []Gap  `json:"gaps,omitempty"`
}

type Gap struct {
	ID               string           `json:"id"`
	Type             GapType          `json:"type"`
	ProfilePath      string           `json:"profile_path,omitempty"`
	SourceProject    string           `json:"source_project,omitempty"`
	ObservedIn       string           `json:"observed_in,omitempty"`
	Evidence         profile.Evidence `json:"evidence"`
	Severity         Severity         `json:"severity"`
	ActionType       ActionType       `json:"action_type"`
	HumanNotes       string           `json:"human_notes,omitempty"`
	LinkedCapability string           `json:"linked_capability,omitempty"`
	LinkedKnowledge  string           `json:"linked_knowledge,omitempty"`
	Details          map[string]any   `json:"details,omitempty"`
}

func FromDiffReport(diffReport *profilediff.Report, ctx Context) Report {
	report := Report{
		SchemaVersion: SchemaVersion,
		SourceProject: ctx.SourceProject,
		ObservedIn:    ctx.ObservedIn,
	}
	if diffReport == nil {
		return report
	}
	for _, diff := range diffReport.Diffs {
		gapType, action := mapDiffAction(diff.ActionType)
		gap := Gap{
			ID:            gapID(gapType, diff.Path),
			Type:          gapType,
			ProfilePath:   diff.Path,
			SourceProject: ctx.SourceProject,
			ObservedIn:    ctx.ObservedIn,
			Evidence:      diff.Evidence,
			Severity:      Severity(diff.Severity),
			ActionType:    action,
			HumanNotes:    diffNote(diff),
			Details: map[string]any{
				"diff_kind": diff.Kind,
				"expected":  diff.Expected,
				"actual":    diff.Actual,
			},
		}
		report.Gaps = append(report.Gaps, gap)
	}
	report.GapCount = len(report.Gaps)
	return report
}

func FromRenderCompare(compare aeoracle.CompareReport, ctx Context) Report {
	report := Report{
		SchemaVersion: SchemaVersion,
		SourceProject: ctx.SourceProject,
		ObservedIn:    ctx.ObservedIn,
	}
	if compare.DifferentPixels == 0 {
		return report
	}
	path := fmt.Sprintf("render[%s->%s]", compare.ExpectedPath, compare.ActualPath)
	report.Gaps = append(report.Gaps, Gap{
		ID:            gapID(TypeRenderGap, path),
		Type:          TypeRenderGap,
		ProfilePath:   path,
		SourceProject: ctx.SourceProject,
		ObservedIn:    ctx.ObservedIn,
		Evidence: profile.Evidence{
			Level:      profile.EvidenceL4Render,
			Source:     "internal/aeoracle",
			Confidence: "high",
		},
		Severity:   SeverityFidelity,
		ActionType: ActionRender,
		HumanNotes: fmt.Sprintf("render differs: %d/%d pixels (%.4f%%), max channel delta %d", compare.DifferentPixels, compare.TotalPixels, compare.DifferentPercent, compare.MaxChannelDelta),
		Details: map[string]any{
			"expected_path":     compare.ExpectedPath,
			"actual_path":       compare.ActualPath,
			"width":             compare.Width,
			"height":            compare.Height,
			"total_pixels":      compare.TotalPixels,
			"different_pixels":  compare.DifferentPixels,
			"different_percent": compare.DifferentPercent,
			"max_channel_delta": compare.MaxChannelDelta,
			"channel_threshold": compare.ChannelThreshold,
		},
	})
	report.GapCount = len(report.Gaps)
	return report
}

func FromFrameSetCompare(frameSet aeoracle.FrameSetCompareReport, ctx Context) Report {
	report := Report{
		SchemaVersion: SchemaVersion,
		SourceProject: ctx.SourceProject,
		ObservedIn:    ctx.ObservedIn,
	}
	for _, frame := range frameSet.Frames {
		if frame.Status == aeoracle.FrameStatusOK {
			continue
		}
		gapType, evidenceKind := frameSetGapType(frame.Status)
		path := fmt.Sprintf("render.frames[%q]", frame.Tag)
		details := map[string]any{
			"evidence_kind": evidenceKind,
			"frame_tag":     frame.Tag,
			"frame":         frame.Frame,
			"seconds":       frame.Seconds,
			"reason":        frame.Reason,
			"expected_path": frame.ExpectedPath,
			"actual_path":   frame.ActualPath,
			"status":        frame.Status,
		}
		if frame.Compare != nil {
			details["width"] = frame.Compare.Width
			details["height"] = frame.Compare.Height
			details["total_pixels"] = frame.Compare.TotalPixels
			details["different_pixels"] = frame.Compare.DifferentPixels
			details["different_percent"] = frame.Compare.DifferentPercent
			details["max_channel_delta"] = frame.Compare.MaxChannelDelta
			details["channel_threshold"] = frame.Compare.ChannelThreshold
		}
		report.Gaps = append(report.Gaps, Gap{
			ID:            gapID(gapType, path),
			Type:          gapType,
			ProfilePath:   path,
			SourceProject: ctx.SourceProject,
			ObservedIn:    ctx.ObservedIn,
			Evidence: profile.Evidence{
				Level:      profile.EvidenceL4Render,
				Source:     "internal/aeoracle",
				Confidence: "high",
			},
			Severity:   SeverityFidelity,
			ActionType: ActionRender,
			HumanNotes: frameSetNote(frame),
			Details:    details,
		})
	}
	report.GapCount = len(report.Gaps)
	return report
}

func frameSetGapType(status string) (GapType, string) {
	switch status {
	case aeoracle.FrameStatusDifferent:
		return TypeSemanticGap, "render_frame_delta"
	case aeoracle.FrameStatusMissingActual:
		return TypeWriteGap, "render_frame_missing_actual"
	case aeoracle.FrameStatusMissingExpected:
		return TypeInvestigateGap, "render_frame_missing_expected"
	default:
		return TypeInvestigateGap, "render_frame_status"
	}
}

func frameSetNote(frame aeoracle.FrameCompareRecord) string {
	if frame.Compare != nil {
		return fmt.Sprintf("render frame %s differs: %d/%d pixels (%.4f%%), max channel delta %d", frame.Tag, frame.Compare.DifferentPixels, frame.Compare.TotalPixels, frame.Compare.DifferentPercent, frame.Compare.MaxChannelDelta)
	}
	return fmt.Sprintf("render frame %s status %s", frame.Tag, frame.Status)
}

func mapDiffAction(action profilediff.ActionType) (GapType, ActionType) {
	switch action {
	case profilediff.ActionParse:
		return TypeParseGap, ActionParse
	case profilediff.ActionWrite:
		return TypeWriteGap, ActionWrite
	case profilediff.ActionSemantics:
		return TypeSemanticGap, ActionSemantics
	case profilediff.ActionRender:
		return TypeRenderGap, ActionRender
	case profilediff.ActionAsset:
		return TypeAssetGap, ActionAsset
	case profilediff.ActionPlugin:
		return TypePluginGap, ActionPlugin
	default:
		return TypeInvestigateGap, ActionInvestigate
	}
}

func diffNote(diff profilediff.Diff) string {
	return fmt.Sprintf("%s %s at %s", diff.Kind, diff.ActionType, diff.Path)
}

func gapID(typ GapType, path string) string {
	sum := sha1.Sum([]byte(string(typ) + "\x00" + path))
	return string(typ) + "-" + hex.EncodeToString(sum[:])[:12]
}
