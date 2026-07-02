package aepmigrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConvertWritesRecipeTextStyleProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-text-style.json"))
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

func TestConvertWritesRecipeTextShapeProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-text-shape.json"))
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

func TestConvertWritesRecipeTextAnimatorOpacityProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-opacity.json"))
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

func TestConvertWritesRecipeTextAnimatorPositionProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-position.json"))
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

func TestConvertWritesRecipeTextAnimatorRangeOffsetProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-range-offset.json"))
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

func TestConvertWritesRecipeTextAnimatorColorValueKeyframesProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-color-value-keyframes.json"))
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

func TestConvertWritesRecipeTextAnimatorRotationXYProjects(t *testing.T) {
	for _, recipeName := range []string{
		"minimal-text-animator-rotation-x.json",
		"minimal-text-animator-rotation-y.json",
	} {
		t.Run(recipeName, func(t *testing.T) {
			source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", recipeName))
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
		})
	}
}

func TestConvertWritesRecipeTextStaticTransformProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-default-text-static-transform.json"))
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

func TestConvertWritesDefaultTextLayerProject(t *testing.T) {
	source := writeTempProjectWithDefaultTextLayer(t)
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

func TestConvertWritesRecipeDefaultTextLayerProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-default-text-layer.json"))
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
