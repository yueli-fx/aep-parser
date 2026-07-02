package aepmigrate

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/profilediff"
)

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
