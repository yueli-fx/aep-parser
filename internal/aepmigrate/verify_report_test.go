package aepmigrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyAllowsDiffsReportedByMigrationReport(t *testing.T) {
	source := writeTempVerifyProject(t, "Source")
	target := writeTempVerifyProject(t, "Target")
	migrationReportPath := writeTempVerifyMigrationReport(t, Report{
		SchemaVersion: SchemaVersion,
		Target:        TargetInfo{Version: VersionAE2020},
		Entries: []Entry{{
			Path:          `comps.by_id[1].name`,
			Class:         ClassTranslated,
			TargetVersion: VersionAE2020,
			Reason:        "Comp rename is an intentional migration mapping.",
		}},
	})

	report, err := Verify(VerifyOptions{
		SourcePath:          source,
		TargetPath:          target,
		TargetVersion:       VersionAE2020,
		MigrationReportPath: migrationReportPath,
	})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if report.Summary.Status != StatusPass {
		t.Fatalf("status = %q, want pass; entries=%+v verification=%+v", report.Summary.Status, report.Entries, report.Verification)
	}
	if report.Verification.ProfileDiffStatus != "pass" || report.Verification.ProfileDiffCount != 0 || report.Verification.ProfileDiffIgnoredCount != 1 {
		t.Fatalf("profile diff verification = %+v, want pass with 1 ignored diff", report.Verification)
	}
}

func TestVerifyRejectsMigrationReportTargetVersionMismatch(t *testing.T) {
	source := writeTempVerifyProject(t, "Source")
	target := writeTempVerifyProject(t, "Target")
	migrationReportPath := writeTempVerifyMigrationReport(t, Report{
		SchemaVersion: SchemaVersion,
		Target:        TargetInfo{Version: VersionAE2025},
		Entries: []Entry{{
			Path:          `comps.by_id[1].name`,
			Class:         ClassTranslated,
			TargetVersion: VersionAE2025,
			Reason:        "Comp rename is an intentional migration mapping.",
		}},
	})

	_, err := Verify(VerifyOptions{
		SourcePath:          source,
		TargetPath:          target,
		TargetVersion:       VersionAE2020,
		MigrationReportPath: migrationReportPath,
	})
	if err == nil {
		t.Fatal("Verify succeeded, want target version mismatch error")
	}
	if !strings.Contains(err.Error(), "target version mismatch") {
		t.Fatalf("error = %q, want target version mismatch", err)
	}
}

func writeTempVerifyMigrationReport(t *testing.T, report Report) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "migration-report.json")
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}
