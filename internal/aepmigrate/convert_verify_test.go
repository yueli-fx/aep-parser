package aepmigrate

import (
	"bytes"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func TestVerifyConvertedProfileBytesAllowsDiffsReportedByConvertReport(t *testing.T) {
	sourceProject := aep.NewProject(aep.TargetAE2020)
	if _, err := aep.NewComposition(sourceProject, "Source", 640, 360, 24, 2); err != nil {
		t.Fatalf("NewComposition source: %v", err)
	}
	sourceProfile, err := profile.Build(sourceProject, profile.Options{Path: "source.aep"})
	if err != nil {
		t.Fatalf("Build source profile: %v", err)
	}

	targetProject := aep.NewProject(aep.TargetAE2020)
	if _, err := aep.NewComposition(targetProject, "Target", 640, 360, 24, 2); err != nil {
		t.Fatalf("NewComposition target: %v", err)
	}
	var targetBytes bytes.Buffer
	if err := targetProject.WriteAEP(&targetBytes); err != nil {
		t.Fatalf("WriteAEP target: %v", err)
	}

	report := Report{
		SchemaVersion: SchemaVersion,
		Target:        TargetInfo{Version: VersionAE2020},
		Entries: []Entry{{
			Path:          "comps.by_id[1].name",
			Class:         ClassTranslated,
			TargetVersion: VersionAE2020,
			Reason:        "Comp rename is an intentional migration mapping.",
		}},
	}
	report.Summary = summarize(report.Entries)

	if err := verifyConvertedProfileBytes(&report, sourceProfile, "target.aep", targetBytes.Bytes()); err != nil {
		t.Fatalf("verifyConvertedProfileBytes: %v", err)
	}
	if report.Verification.ProfileDiffStatus != "pass" || report.Verification.ProfileDiffCount != 0 || report.Verification.ProfileDiffIgnoredCount != 1 {
		t.Fatalf("profile diff verification = %+v, want pass with 1 ignored diff", report.Verification)
	}
	if report.Summary.Status != StatusPass {
		t.Fatalf("status = %q, entries=%+v", report.Summary.Status, report.Entries)
	}
}
