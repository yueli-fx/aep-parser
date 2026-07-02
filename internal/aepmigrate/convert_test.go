package aepmigrate

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func TestConvertWritesTargetVersionNoLayerProject(t *testing.T) {
	source := writeTempProjectWithOneComp(t, aep.TargetAE2020)
	outPath := filepath.Join(t.TempDir(), "converted.aep")

	report, err := Convert(ConvertOptions{
		InputPath:  source,
		OutputPath: outPath,
		Target:     VersionAE2025,
	})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if report.Summary.Status != StatusPass {
		t.Fatalf("status = %q, entries=%+v", report.Summary.Status, report.Entries)
	}
	if report.Summary.Preserved != 0 || report.Summary.Retargeted == 0 {
		t.Fatalf("summary preserved=%d retargeted=%d, want no preserved placeholder and retargeted entries", report.Summary.Preserved, report.Summary.Retargeted)
	}
	converted, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open converted: %v", err)
	}
	version := NormalizeSourceVersion((&aep.Application{Project: converted}).Version())
	if version.Label != VersionAE2025 {
		t.Fatalf("converted version = %+v, want AE2025", version)
	}
	prof, err := profile.Build(converted, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build converted: %v", err)
	}
	if len(prof.Comps) != 1 {
		t.Fatalf("converted comps = %d, want 1", len(prof.Comps))
	}
	comp := prof.Comps[0]
	if comp.Name != "Main" || comp.Width != 640 || comp.Height != 360 {
		t.Fatalf("converted comp identity = %+v", comp)
	}
	if math.Abs(comp.FrameRate-24) > 0.001 || math.Abs(comp.Duration-2.5) > 0.001 {
		t.Fatalf("converted timing frameRate=%f duration=%f, want 24/2.5", comp.FrameRate, comp.Duration)
	}
}

func TestConvertRunsProfileDiffVerification(t *testing.T) {
	source := writeTempProjectWithConfiguredNoLayerComp(t)
	outPath := filepath.Join(t.TempDir(), "converted.aep")

	report, err := Convert(ConvertOptions{
		InputPath:  source,
		OutputPath: outPath,
		Target:     VersionAE2025,
	})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if report.Verification.ProfileDiffStatus != "pass" {
		t.Fatalf("profile diff status = %q, want pass; verification=%+v", report.Verification.ProfileDiffStatus, report.Verification)
	}
	if report.Verification.ProfileDiffCount != 0 {
		t.Fatalf("profile diff count = %d, want 0; verification=%+v", report.Verification.ProfileDiffCount, report.Verification)
	}
}
