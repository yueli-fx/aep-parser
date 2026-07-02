package aepmigrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConvertWritesRecipeLayerExplicitMatteProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-layer-explicit-matte.json")
}

func TestConvertWritesRecipeLayerTrackMatteProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-layer-track-matte.json")
}

func TestConvertWritesAE2025ClassicTrackMatteDowngradeProject(t *testing.T) {
	source := writeTempRecipeWithTarget(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-track-matte.json"), "AE2025")
	assertSourceConvertsPassToTarget(t, source, VersionAE2020)
}

func TestConvertWritesRecipeLayerMaskProject(t *testing.T) {
	outPath := assertRecipeConvertsPass(t, "minimal-layer-mask.json")
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
