package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/profilediff"
)

type VerifyOptions struct {
	SourcePath          string
	TargetPath          string
	TargetVersion       VersionLabel
	MigrationReportPath string
	AEOpen              *AEOpenOptions
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
	report.Entries = append(report.Entries, verificationCapabilityEntries(classifyProfile(sourceVersion.Label, opts.TargetVersion, sourceProfile))...)
	migrationReport, err := loadVerifyMigrationReport(opts.MigrationReportPath)
	if err != nil {
		return Report{}, err
	}
	if migrationReport != nil {
		if migrationReport.Target.Version != opts.TargetVersion {
			return Report{}, fmt.Errorf("verify migration report: target version mismatch: report=%s verify=%s", migrationReport.Target.Version, opts.TargetVersion)
		}
		report.Entries = append(report.Entries, migrationReport.Entries...)
	}

	diffReport, err := profilediff.Compare(sourceProfile, targetProfile, profilediff.Options{})
	if err != nil {
		return Report{}, fmt.Errorf("profile diff compare: %w", err)
	}
	unexpectedDiffs, allowedCount := filterReportedProfileDiffs(diffReport.Diffs, migrationReport, opts.TargetVersion)
	report.Verification.ProfileDiffCount = len(unexpectedDiffs)
	report.Verification.ProfileDiffIgnoredCount = diffReport.IgnoredCount + allowedCount
	report.Verification.ProfileDiffs = verificationDiffs(unexpectedDiffs)
	if len(unexpectedDiffs) == 0 {
		report.Verification.ProfileDiffStatus = "pass"
		report.Summary = summarize(report.Entries)
		if report.Summary.Status == StatusBlocked {
			return report, nil
		}
		if err := verifyAEOpen(&report, ConvertOptions{AEOpen: opts.AEOpen}); err != nil {
			return Report{}, err
		}
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

func verificationCapabilityEntries(entries []Entry) []Entry {
	out := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		if entry.Path == "project" && entry.Class == ClassPreserved && entry.CapabilityKey == "" {
			continue
		}
		out = append(out, entry)
	}
	return out
}
