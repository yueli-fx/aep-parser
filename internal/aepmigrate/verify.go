package aepmigrate

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/profilediff"
)

type VerifyOptions struct {
	SourcePath          string
	TargetPath          string
	TargetVersion       VersionLabel
	MigrationReportPath string
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

func loadVerifyMigrationReport(path string) (*Report, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("verify migration report: read: %w", err)
	}
	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("verify migration report: parse: %w", err)
	}
	if report.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("verify migration report: unsupported schema_version %d", report.SchemaVersion)
	}
	return &report, nil
}

func filterReportedProfileDiffs(diffs []profilediff.Diff, migrationReport *Report, target VersionLabel) ([]profilediff.Diff, int) {
	if migrationReport == nil {
		return diffs, 0
	}
	var unexpected []profilediff.Diff
	allowedCount := 0
	for _, diff := range diffs {
		if migrationReportAllowsProfileDiff(*migrationReport, diff, target) {
			allowedCount++
			continue
		}
		unexpected = append(unexpected, diff)
	}
	return unexpected, allowedCount
}

func migrationReportAllowsProfileDiff(report Report, diff profilediff.Diff, target VersionLabel) bool {
	for _, entry := range report.Entries {
		if entry.Path != diff.Path || entry.TargetVersion != target || entry.Reason == "" {
			continue
		}
		switch entry.Class {
		case ClassRetargeted, ClassTranslated, ClassApproximated, ClassDropped:
			return true
		}
	}
	return false
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
