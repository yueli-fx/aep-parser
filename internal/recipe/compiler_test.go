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

func TestCompileToFileMaterializesSupportedEffects(t *testing.T) {
	effects := aep.SupportedEffects()
	if len(effects) == 0 {
		t.Fatal("SupportedEffects is empty")
	}
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{MatchName: effects[0]}}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	if got := prof.Comps[0].Layers[0].Effects; len(got) != 1 || got[0].MatchName != effects[0] {
		t.Fatalf("Effects = %+v, want %q", got, effects[0])
	}
}

func TestCompileToFileSetsEffectParams(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{
		MatchName: "ADBE Gaussian Blur 2",
		Params: []recipe.EffectParam{
			{MatchName: "ADBE Gaussian Blur 2-0001", Value: 25.0},
			{MatchName: "ADBE Gaussian Blur 2-0002", Value: 2.0},
			{MatchName: "ADBE Gaussian Blur 2-0003", Value: true},
		},
	}}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	params := prof.Comps[0].Layers[0].Effects[0].Params
	assertParamValue(t, params, "ADBE Gaussian Blur 2-0001", 25.0)
	assertParamValue(t, params, "ADBE Gaussian Blur 2-0002", 2.0)
	assertParamValue(t, params, "ADBE Gaussian Blur 2-0003", 1.0)
}

func assertParamValue(t *testing.T, params []profile.Property, matchName string, want float64) {
	t.Helper()
	for _, param := range params {
		if param.MatchName != matchName {
			continue
		}
		got, ok := param.StaticValue.(float64)
		if !ok || got != want {
			t.Fatalf("%s StaticValue = %v, want %v", matchName, param.StaticValue, want)
		}
		return
	}
	t.Fatalf("param %q not found in %+v", matchName, params)
}
