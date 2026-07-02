package aepmigrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConvertWritesRecipeLayerExplicitMatteProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-explicit-matte.json"))
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
		t.Fatalf("status = %q, entries=%+v, diffs=%+v", report.Summary.Status, report.Entries, report.Verification.ProfileDiffs)
	}
	if report.Verification.ProfileDiffStatus != "pass" || report.Verification.ProfileDiffCount != 0 {
		t.Fatalf("profile diff verification = %+v, want pass with 0 diffs", report.Verification)
	}
}

func TestConvertWritesRecipeLayerTrackMatteProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-track-matte.json"))
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
		t.Fatalf("status = %q, entries=%+v, diffs=%+v", report.Summary.Status, report.Entries, report.Verification.ProfileDiffs)
	}
	if report.Verification.ProfileDiffStatus != "pass" || report.Verification.ProfileDiffCount != 0 {
		t.Fatalf("profile diff verification = %+v, want pass with 0 diffs", report.Verification)
	}
}

func TestConvertWritesAE2025ClassicTrackMatteDowngradeProject(t *testing.T) {
	source := writeTempRecipeWithTarget(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-track-matte.json"), "AE2025")
	outPath := filepath.Join(t.TempDir(), "converted.aep")

	report, err := Convert(ConvertOptions{
		InputPath:  source,
		OutputPath: outPath,
		Target:     VersionAE2020,
	})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if report.Summary.Status != StatusPass {
		t.Fatalf("status = %q, entries=%+v, diffs=%+v", report.Summary.Status, report.Entries, report.Verification.ProfileDiffs)
	}
	if report.Verification.ProfileDiffStatus != "pass" || report.Verification.ProfileDiffCount != 0 {
		t.Fatalf("profile diff verification = %+v, want pass with 0 diffs", report.Verification)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
}

func TestConvertWritesRecipeLayerMaskProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-mask.json"))
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
		t.Fatalf("status = %q, entries=%+v, diffs=%+v", report.Summary.Status, report.Entries, report.Verification.ProfileDiffs)
	}
	if report.Verification.ProfileDiffStatus != "pass" || report.Verification.ProfileDiffCount != 0 {
		t.Fatalf("profile diff verification = %+v, want pass with 0 diffs", report.Verification)
	}
	assertConvertedLayerMaskProfile(t, outPath)
}

func TestConvertRefusesExplicitMatteDowngradeProjects(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-explicit-matte.json"))

	for _, target := range []VersionLabel{VersionAE2020, VersionAE2024} {
		t.Run(string(target), func(t *testing.T) {
			outPath := filepath.Join(t.TempDir(), "converted.aep")

			report, err := Convert(ConvertOptions{
				InputPath:  source,
				OutputPath: outPath,
				Target:     target,
			})
			if err != nil {
				t.Fatalf("Convert: %v", err)
			}
			if report.Summary.Status != StatusBlocked {
				t.Fatalf("status = %q, want blocked; entries=%+v", report.Summary.Status, report.Entries)
			}
			if _, err := os.Stat(outPath); !os.IsNotExist(err) {
				t.Fatalf("output exists or stat failed unexpectedly: %v", err)
			}
		})
	}
}
