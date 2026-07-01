package aepmigrate

import (
	"os"
	"path/filepath"
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
