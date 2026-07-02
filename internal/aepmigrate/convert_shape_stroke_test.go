package aepmigrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConvertWritesRecipeRectStrokeShapeLayerProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-shape-stroke-style.json"))
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
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
}

func TestConvertWritesRecipeRectStrokeDashesShapeLayerProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-shape-stroke-dashes.json"))
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
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
}

func TestConvertWritesRecipeEllipseStrokeTaperShapeLayerProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-shape-stroke-taper.json"))
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
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
}

func TestConvertWritesRecipeEllipseStrokeWaveShapeLayerProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-shape-stroke-wave.json"))
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
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
}
