package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/profilediff"
)

type VerifyOptions struct {
	SourcePath    string
	TargetPath    string
	TargetVersion VersionLabel
}

func Verify(opts VerifyOptions) (Report, error) {
	sourceProject, err := aep.Open(opts.SourcePath)
	if err != nil {
		return Report{}, err
	}
	targetProject, err := aep.Open(opts.TargetPath)
	if err != nil {
		return Report{}, err
	}
	sourceProfile, err := profile.Build(sourceProject, profile.Options{Path: opts.SourcePath})
	if err != nil {
		return Report{}, err
	}
	targetProfile, err := profile.Build(targetProject, profile.Options{Path: opts.TargetPath})
	if err != nil {
		return Report{}, err
	}
	sourceVersion := NormalizeSourceVersion((&aep.Application{Project: sourceProject}).Version())
	report := Report{
		SchemaVersion: SchemaVersion,
		Source: SourceInfo{
			Path:       opts.SourcePath,
			Version:    sourceVersion.Label,
			VersionRaw: sourceVersion.Raw,
		},
		Target: TargetInfo{
			Version: opts.TargetVersion,
			Path:    opts.TargetPath,
		},
		Verification: Verification{
			AEOpenStatus: "not_run",
			RenderStatus: "not_run",
		},
	}

	diffReport, err := profilediff.Compare(sourceProfile, targetProfile, profilediff.Options{})
	if err != nil {
		return Report{}, fmt.Errorf("profile diff compare: %w", err)
	}
	report.Verification.ProfileDiffCount = diffReport.DiffCount
	report.Verification.ProfileDiffIgnoredCount = diffReport.IgnoredCount
	report.Verification.ProfileDiffs = verificationDiffs(diffReport.Diffs)
	if diffReport.DiffCount == 0 {
		report.Verification.ProfileDiffStatus = "pass"
		report.Summary = Summary{Status: StatusPass}
		return report, nil
	}

	report.Verification.ProfileDiffStatus = "fail"
	report.Entries = append(report.Entries, Entry{
		Path:          "verification.profile_diff",
		Class:         ClassBlocked,
		TargetVersion: opts.TargetVersion,
		Reason:        fmt.Sprintf("Profile diff verification found %d unexpected target difference(s).", diffReport.DiffCount),
	})
	report.Summary = summarize(report.Entries)
	return report, nil
}
