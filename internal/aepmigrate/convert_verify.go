package aepmigrate

import (
	"bytes"
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/profilediff"
)

func verifyConvertedProfileBytes(report *Report, source *profile.Profile, targetPath string, targetBytes []byte) error {
	targetProject, err := aep.FromReader(bytes.NewReader(targetBytes))
	if err != nil {
		return fmt.Errorf("profile diff target open: %w", err)
	}
	target, err := profile.Build(targetProject, profile.Options{Path: targetPath})
	if err != nil {
		return fmt.Errorf("profile diff target profile: %w", err)
	}
	diffReport, err := profilediff.Compare(source, target, profilediff.Options{})
	if err != nil {
		return fmt.Errorf("profile diff compare: %w", err)
	}
	unexpectedDiffs, allowedCount := filterReportedProfileDiffs(diffReport.Diffs, report, report.Target.Version)
	report.Verification.ProfileDiffCount = len(unexpectedDiffs)
	report.Verification.ProfileDiffIgnoredCount = diffReport.IgnoredCount + allowedCount
	report.Verification.ProfileDiffs = verificationDiffs(unexpectedDiffs)
	if len(unexpectedDiffs) == 0 {
		report.Verification.ProfileDiffStatus = "pass"
		return nil
	}
	report.Verification.ProfileDiffStatus = "fail"
	report.Entries = append(report.Entries, Entry{
		Path:          "verification.profile_diff",
		Class:         ClassBlocked,
		TargetVersion: report.Target.Version,
		Reason:        fmt.Sprintf("Profile diff verification found %d unexpected migrated output difference(s).", diffReport.DiffCount),
	})
	report.Summary = summarize(report.Entries)
	return nil
}

func verificationDiffs(diffs []profilediff.Diff) []VerificationDiff {
	out := make([]VerificationDiff, 0, len(diffs))
	for _, diff := range diffs {
		out = append(out, VerificationDiff{
			Path:       diff.Path,
			Kind:       string(diff.Kind),
			Severity:   string(diff.Severity),
			ActionType: string(diff.ActionType),
			Expected:   diff.Expected,
			Actual:     diff.Actual,
		})
	}
	return out
}
