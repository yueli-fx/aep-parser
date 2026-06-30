package aepmigrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestAssessReportsPassForEmptyCompatibleProject(t *testing.T) {
	source := writeTempProject(t, aep.TargetAE2020)

	report, err := Assess(Options{InputPath: source, Target: VersionAE2020})
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	if report.SchemaVersion != SchemaVersion {
		t.Fatalf("SchemaVersion = %d, want %d", report.SchemaVersion, SchemaVersion)
	}
	if report.Target.Version != VersionAE2020 {
		t.Fatalf("target = %q, want AE2020", report.Target.Version)
	}
	if report.Summary.Status != StatusPass {
		t.Fatalf("status = %q, entries=%+v", report.Summary.Status, report.Entries)
	}
	if report.Verification.ProfileDiffStatus != "not_run" {
		t.Fatalf("ProfileDiffStatus = %q, want not_run", report.Verification.ProfileDiffStatus)
	}
}

func TestAssessBlocksExplicitMatteDowngrade(t *testing.T) {
	source := writeTempProjectWithExplicitMatte(t)

	report, err := Assess(Options{InputPath: source, Target: VersionAE2020})
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	if report.Summary.Status != StatusBlocked {
		t.Fatalf("status = %q, want blocked; entries=%+v", report.Summary.Status, report.Entries)
	}
	if report.Summary.Blocked == 0 {
		t.Fatalf("blocked count = 0, entries=%+v", report.Entries)
	}
}

func TestAssessAllowsClassicMatteOnAE2020Target(t *testing.T) {
	source := writeTempProjectWithClassicMatte(t)

	report, err := Assess(Options{InputPath: source, Target: VersionAE2020})
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	if report.Summary.Status != StatusPass {
		t.Fatalf("status = %q, want pass; entries=%+v", report.Summary.Status, report.Entries)
	}
}

func writeTempProject(t *testing.T, target aep.AETarget) string {
	t.Helper()
	project := aep.NewProject(target)
	path := filepath.Join(t.TempDir(), "source.aep")
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

func writeTempProjectWithClassicMatte(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "Matte", 640, 360, [3]float64{1, 1, 1}); err != nil {
		t.Fatalf("NewSolidLayer matte: %v", err)
	}
	fill, err := aep.NewSolidLayer(comp, "Fill", 640, 360, [3]float64{1, 0, 0})
	if err != nil {
		t.Fatalf("NewSolidLayer fill: %v", err)
	}
	if err := fill.SetTrackMatte(aep.TrackMatteAlpha); err != nil {
		t.Fatalf("SetTrackMatte: %v", err)
	}
	path := filepath.Join(t.TempDir(), "classic-matte.aep")
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

func writeTempProjectWithExplicitMatte(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2025)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	matte, err := aep.NewSolidLayer(comp, "Matte", 640, 360, [3]float64{1, 1, 1})
	if err != nil {
		t.Fatalf("NewSolidLayer matte: %v", err)
	}
	fill, err := aep.NewSolidLayer(comp, "Fill", 640, 360, [3]float64{1, 0, 0})
	if err != nil {
		t.Fatalf("NewSolidLayer fill: %v", err)
	}
	if err := fill.SetTrackMatteSource(matte, aep.TrackMatteAlpha); err != nil {
		t.Fatalf("SetTrackMatteSource: %v", err)
	}
	path := filepath.Join(t.TempDir(), "explicit-matte.aep")
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
