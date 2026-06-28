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

func TestCompileToFileSetsShapeStroke(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Stroke = &recipe.StrokeSpec{
		Color:   []float64{255, 0, 0, 255},
		Width:   ptr(6),
		Opacity: ptr(80),
	}
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
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Color", []float64{255, 255, 0, 0})
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Width", 6.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Opacity", 80.0)
}

func TestCompileToFileSetsShapeDetail(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Position = []float64{12, -6}
	rec.Comps[0].Layers[1].Shape.Roundness = ptr(18)
	rec.Comps[0].Layers[1].Shape.FillOpacity = ptr(45)
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
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Rect Position", []float64{12, -6})
	assertLayerPropertyValue(t, layer, "ADBE Vector Rect Roundness", 18.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Fill Opacity", 45.0)
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

func TestCompileToFileChecksExpectedProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.ExpectedProfile = recipe.ExpectedProfile{
		CompCount:       intPtr(1),
		LayerCount:      intPtr(2),
		TextLayerCount:  intPtr(1),
		ShapeLayerCount: intPtr(1),
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layer_count", true)
}

func TestCompileToFileRefusesExpectedProfileMismatch(t *testing.T) {
	rec := minimalRecipe()
	rec.ExpectedProfile = recipe.ExpectedProfile{
		LayerCount: intPtr(3),
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if report.Valid {
		t.Fatalf("report = %+v, want invalid", report)
	}
	assertRefusal(t, report, "profile_contract_mismatch")
	assertProfileCheck(t, report, "expected_profile.layer_count", false)
	if _, err := aep.Open(outPath); err == nil {
		t.Fatal("AEP was written despite expected-profile mismatch")
	}
}

func TestCompileToFileChecksExpectedEffectParamProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{
		MatchName: "ADBE Gaussian Blur 2",
		Params: []recipe.EffectParam{
			{MatchName: "ADBE Gaussian Blur 2-0001", Value: 25.0},
		},
	}}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Effects: []recipe.ExpectedEffect{{
			LayerName: "Title",
			MatchName: "ADBE Gaussian Blur 2",
			Params: []recipe.ExpectedEffectParam{{
				MatchName: "ADBE Gaussian Blur 2-0001",
				Value:     25.0,
			}},
		}},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.effects[0].params[0]", true)
}

func TestCompileToFileChecksExpectedLayerPropertyProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Stroke = &recipe.StrokeSpec{
		Color: []float64{255, 0, 0, 255},
		Width: ptr(6),
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Properties: []recipe.ExpectedProperty{
			{LayerName: "Underline", MatchName: "ADBE Vector Stroke Color", Value: []float64{255, 255, 0, 0}},
			{LayerName: "Underline", MatchName: "ADBE Vector Stroke Width", Value: 6.0},
		},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
	assertProfileCheck(t, report, "expected_profile.properties[1]", true)
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

func findProfileLayer(t *testing.T, prof *profile.Profile, name string) profile.Layer {
	t.Helper()
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if layer.Name == name {
				return layer
			}
		}
	}
	t.Fatalf("layer %q not found in %+v", name, prof.Comps)
	return profile.Layer{}
}

func assertLayerPropertyValue(t *testing.T, layer profile.Layer, matchName string, want any) {
	t.Helper()
	for _, prop := range layer.Properties {
		if prop.MatchName == matchName {
			assertProfileValue(t, matchName, prop.StaticValue, want)
			return
		}
	}
	for _, shape := range layer.Shapes {
		for _, prop := range shape.Properties {
			if prop.MatchName == matchName {
				assertProfileValue(t, matchName, prop.StaticValue, want)
				return
			}
		}
	}
	t.Fatalf("property %q not found on layer %+v", matchName, layer)
}

func assertProfileValue(t *testing.T, label string, got, want any) {
	t.Helper()
	switch want := want.(type) {
	case float64:
		gotFloat, ok := got.(float64)
		if !ok || gotFloat != want {
			t.Fatalf("%s StaticValue = %v, want %v", label, got, want)
		}
	case []float64:
		gotSlice, ok := got.([]float64)
		if !ok || len(gotSlice) != len(want) {
			t.Fatalf("%s StaticValue = %v, want %v", label, got, want)
		}
		for i := range want {
			if gotSlice[i] != want[i] {
				t.Fatalf("%s StaticValue = %v, want %v", label, got, want)
			}
		}
	default:
		t.Fatalf("unsupported want type %T", want)
	}
}

func assertProfileCheck(t *testing.T, report recipe.Report, path string, passed bool) {
	t.Helper()
	for _, check := range report.ProfileChecks {
		if check.Path == path && check.Passed == passed {
			return
		}
	}
	t.Fatalf("profile check %q passed=%v not found in %+v", path, passed, report.ProfileChecks)
}

func intPtr(v int) *int {
	return &v
}
