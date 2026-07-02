package aepmigrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func TestConvertWritesDefaultNullLayerProject(t *testing.T) {
	source := writeTempProjectWithDefaultNullLayer(t)
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
	outProject, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open converted: %v", err)
	}
	prof, err := profile.Build(outProject, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile converted: %v", err)
	}
	if len(prof.Comps) != 1 || len(prof.Comps[0].Layers) != 1 {
		t.Fatalf("converted profile layers = %+v", prof.Comps)
	}
	if got := prof.Comps[0].Layers[0].Type; got != "null" {
		t.Fatalf("layer type = %q, want null", got)
	}
}

func TestConvertWritesRecipeDefaultNullLayerProject(t *testing.T) {
	source := writeTempRecipeDefaultNullLayer(t)
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

func TestConvertWritesDefaultSolidLayerProject(t *testing.T) {
	source := writeTempProjectWithOneSolidLayer(t)
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
	outProject, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open converted: %v", err)
	}
	prof, err := profile.Build(outProject, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile converted: %v", err)
	}
	if len(prof.Comps) != 1 || len(prof.Comps[0].Layers) != 1 {
		t.Fatalf("converted profile layers = %+v", prof.Comps)
	}
	layer := prof.Comps[0].Layers[0]
	if layer.Type != "av" || layer.SourceRef == nil || layer.SourceRef.Kind != "footage" {
		t.Fatalf("converted layer source = %+v, want av footage layer", layer)
	}
}

func TestConvertWritesRecipeDefaultSolidLayerProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-source-ref.json"))
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

func TestConvertWritesDefaultAdjustmentLayerProject(t *testing.T) {
	source := writeTempProjectWithOneAdjustmentLayer(t)
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

func TestConvertWritesRecipeDefaultAdjustmentLayerProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-default-adjustment-layer.json"))
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

func TestConvertWritesDefaultCameraLayerProject(t *testing.T) {
	source := writeTempProjectWithOneCameraLayer(t)
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

func TestConvertWritesRecipeDefaultCameraLayerProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-default-camera-layer.json"))
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

func TestConvertWritesDefaultLightLayerProject(t *testing.T) {
	source := writeTempProjectWithOneLightLayer(t)
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

func TestConvertWritesRecipeDefaultLightLayerProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-default-light-layer.json"))
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

func TestConvertWritesDefaultPrecompLayerProject(t *testing.T) {
	source := writeTempProjectWithOnePrecompLayer(t)
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

func TestConvertWritesRecipeDefaultPrecompLayerProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-default-precomp-layer.json"))
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

func TestConvertWritesRecipeLayerLabelProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-label.json"))
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

func TestConvertWritesRecipeLayerCommentProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-comment.json"))
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

func TestConvertWritesRecipeLayerCommonSwitchesProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-common-switches.json"))
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

func TestConvertWritesRecipeLayerShyProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-shy.json"))
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

func TestConvertWritesRecipeLayerMotionBlurProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-motion-blur.json"))
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

func TestConvertWritesRecipeLayerQualityBlendingProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-quality-blending.json"))
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

func TestConvertWritesRecipeLayerParentProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-parent.json"))
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

func TestConvertWritesRecipeLayerNullFlagProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-null-flag.json"))
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

func TestConvertWritesRecipeCameraObjectProfile(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-camera-object-profile.json"))
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

func TestConvertWritesRecipeLightObjectProfile(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-light-object-profile.json"))
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

func TestConvertWritesRecipeLightSourceProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-light-source.json"))
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

func TestConvertWritesRecipeLayerAdvancedSwitchesProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-advanced-switches.json"))
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

func TestConvertWritesRecipeLayerObjectProfile(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-object-profile.json"))
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

func TestConvertWritesRecipeLayerTimingProject(t *testing.T) {
	source := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-timing.json"))
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

func TestConvertWritesMovedNullLayerProject(t *testing.T) {
	source := writeTempProjectWithMovedNullLayer(t)
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

func TestConvertWritesKeyframedNullLayerProject(t *testing.T) {
	source := writeTempProjectWithKeyframedNullLayer(t)
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
