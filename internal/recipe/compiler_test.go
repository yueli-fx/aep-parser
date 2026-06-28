package recipe_test

import (
	"path/filepath"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/profile"
	"github.com/example/aep-parser/internal/recipe"
)

func TestCompileMinimalTextShapeRecipeBuildsProfile(t *testing.T) {
	rec := minimalRecipe()
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, recipe.StaticCapabilities{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid || report.OutputPath != outPath {
		t.Fatalf("report = %+v", report)
	}

	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	if prof.Fingerprint.CompCount != 1 {
		t.Fatalf("CompCount = %d, want 1", prof.Fingerprint.CompCount)
	}
	if prof.Fingerprint.LayerCount != 2 {
		t.Fatalf("LayerCount = %d, want 2", prof.Fingerprint.LayerCount)
	}
	var textLayers, shapeLayers int
	for _, layer := range prof.Comps[0].Layers {
		if layer.Text != nil {
			textLayers++
		}
		if len(layer.Shapes) > 0 {
			shapeLayers++
		}
	}
	if textLayers != 1 || shapeLayers != 1 {
		t.Fatalf("textLayers=%d shapeLayers=%d, want 1/1; layers=%+v", textLayers, shapeLayers, prof.Comps[0].Layers)
	}
}

func TestCompileToFileCreatesParentDirectory(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "nested", "recipe.aep")

	report, err := recipe.CompileToFile(minimalRecipe(), outPath, recipe.StaticCapabilities{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v", report)
	}
	if _, err := aep.Open(outPath); err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
}
