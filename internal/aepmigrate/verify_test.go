package aepmigrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestVerifyPassesMatchingProfiles(t *testing.T) {
	source := writeTempVerifyProject(t, "Main")
	target := source

	report, err := Verify(VerifyOptions{
		SourcePath:    source,
		TargetPath:    target,
		TargetVersion: VersionAE2020,
	})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if report.Summary.Status != StatusPass {
		t.Fatalf("status = %q, entries=%+v", report.Summary.Status, report.Entries)
	}
	if report.Verification.ProfileDiffStatus != "pass" || report.Verification.ProfileDiffCount != 0 {
		t.Fatalf("profile diff verification = %+v, want pass with 0 diffs", report.Verification)
	}
	if report.Source.Path != source || report.Target.Path != target || report.Target.Version != VersionAE2020 {
		t.Fatalf("source/target metadata = %+v %+v", report.Source, report.Target)
	}
}

func TestVerifyBlocksDifferentProfiles(t *testing.T) {
	source := writeTempVerifyProject(t, "Source")
	target := writeTempVerifyProject(t, "Target")

	report, err := Verify(VerifyOptions{
		SourcePath:    source,
		TargetPath:    target,
		TargetVersion: VersionAE2020,
	})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if report.Summary.Status != StatusBlocked {
		t.Fatalf("status = %q, want blocked; entries=%+v", report.Summary.Status, report.Entries)
	}
	if report.Verification.ProfileDiffStatus != "fail" || report.Verification.ProfileDiffCount == 0 {
		t.Fatalf("profile diff verification = %+v, want fail with diffs", report.Verification)
	}
}

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

func TestVerifyBlocksLedgerBoundaryEvenWhenProfilesMatch(t *testing.T) {
	source := writeTempProjectWithExplicitMatte(t)

	report, err := Verify(VerifyOptions{
		SourcePath:    source,
		TargetPath:    source,
		TargetVersion: VersionAE2020,
	})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if report.Summary.Status != StatusBlocked {
		t.Fatalf("status = %q, want blocked; entries=%+v", report.Summary.Status, report.Entries)
	}
	if report.Verification.ProfileDiffStatus != "pass" || report.Verification.ProfileDiffCount != 0 {
		t.Fatalf("profile diff verification = %+v, want pass with 0 diffs", report.Verification)
	}
	for _, entry := range report.Entries {
		if entry.CapabilityKey == "layer.set_track_matte_source" && entry.Class == ClassBlocked {
			return
		}
	}
	t.Fatalf("missing explicit matte ledger boundary entry: %+v", report.Entries)
}

func writeTempVerifyProject(t *testing.T, compName string) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	if _, err := aep.NewComposition(project, compName, 640, 360, 24, 2); err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	path := filepath.Join(t.TempDir(), compName+".aep")
	out, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer out.Close()
	if err := project.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	return path
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
