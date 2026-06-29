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
		Dashes: &recipe.StrokeDashesSpec{
			Dash: ptr(18),
			Gap:  ptr(7),
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
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Dash 1", 18.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Gap 1", 7.0)
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

func TestCompileToFileSetsShapeStar(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Kind = "polygon"
	rec.Comps[0].Layers[1].Shape.Points = ptr(6)
	rec.Comps[0].Layers[1].Shape.Position = []float64{20, -10}
	rec.Comps[0].Layers[1].Shape.Rotation = ptr(30)
	rec.Comps[0].Layers[1].Shape.InnerRadius = ptr(45)
	rec.Comps[0].Layers[1].Shape.OuterRadius = ptr(120)
	rec.Comps[0].Layers[1].Shape.InnerRoundness = ptr(10)
	rec.Comps[0].Layers[1].Shape.OuterRoundness = ptr(20)
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
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Type", 2.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Points", 6.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Position", []float64{20, -10})
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Rotation", 30.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Inner Radius", 45.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Outer Radius", 120.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Inner Roundess", 10.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Outer Roundess", 20.0)
}

func TestCompileToFileSetsShapeTrim(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Trim = &recipe.TrimSpec{
		Start:  ptr(10),
		End:    ptr(85),
		Offset: ptr(15),
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
	assertLayerPropertyValue(t, layer, "ADBE Vector Trim Start", 10.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Trim End", 85.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Trim Offset", 15.0)
}

func TestCompileToFileSetsShapeRoundCorners(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.RoundCorners = &recipe.RoundCornersSpec{
		Radius: ptr(18),
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
	assertLayerPropertyValue(t, layer, "ADBE Vector RoundCorner Radius", 18.0)
}

func TestCompileToFileSetsShapeOffsetPaths(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.OffsetPaths = &recipe.OffsetPathsSpec{
		Amount:     ptr(24),
		LineJoin:   "bevel",
		MiterLimit: ptr(2),
		Copies:     ptr(3),
		CopyOffset: ptr(1.5),
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
	assertLayerPropertyValue(t, layer, "ADBE Vector Offset Amount", 24.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Offset Line Join", 3.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Offset Miter Limit", 2.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Offset Copies", 3.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Offset Copy Offset", 1.5)
}

func TestCompileToFileSetsShapeRepeater(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Repeater = &recipe.RepeaterSpec{
		Copies:       ptr(5),
		Offset:       ptr(2),
		Order:        "above",
		Anchor:       []float64{15, 25},
		Position:     []float64{120, 0},
		Scale:        []float64{80, 90},
		Rotation:     ptr(30),
		StartOpacity: ptr(100),
		EndOpacity:   ptr(25),
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
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Copies", 5.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Offset", 2.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Order", 2.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Anchor", []float64{15, 25})
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Position", []float64{120, 0})
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Scale", []float64{80, 90})
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Rotation", 30.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Opacity 1", 100.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Opacity 2", 25.0)
}

func TestCompileToFileSetsShapeMergePaths(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.MergePaths = &recipe.MergePathsSpec{
		Type: "subtract",
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
	assertLayerPropertyValue(t, layer, "ADBE Vector Merge Type", 3.0)
}

func TestCompileToFileSetsShapeZigZag(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.ZigZag = &recipe.ZigZagSpec{
		Size:   ptr(40),
		Detail: ptr(8),
		Points: "smooth",
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
	assertLayerPropertyValue(t, layer, "ADBE Vector Zigzag Size", 40.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Zigzag Detail", 8.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Zigzag Points", 2.0)
}

func TestCompileToFileSetsShapePuckerBloat(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.PuckerBloat = &recipe.PuckerBloatSpec{
		Amount: ptr(100),
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
	assertLayerPropertyValue(t, layer, "ADBE Vector PuckerBloat Amount", 100.0)
}

func TestCompileToFileSetsShapeTwist(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Twist = &recipe.TwistSpec{
		Angle:  ptr(150),
		Center: []float64{24, -12},
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
	assertLayerPropertyValue(t, layer, "ADBE Vector Twist Angle", 150.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Twist Center", []float64{24, -12})
}

func TestCompileToFileSetsShapeWigglePaths(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.WigglePaths = &recipe.WigglePathsSpec{
		Size:             ptr(60),
		Detail:           ptr(30),
		WigglesPerSecond: ptr(4),
		RandomSeed:       ptr(9),
		Points:           "smooth",
		Correlation:      ptr(80),
		TemporalPhase:    ptr(45),
		SpatialPhase:     ptr(20),
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
	assertLayerPropertyValue(t, layer, "ADBE Vector Roughen Size", 60.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Roughen Detail", 30.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Temporal Freq", 4.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Random Seed", 9.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Roughen Points", 2.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Correlation", 80.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Temporal Phase", 45.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Spatial Phase", 20.0)
}

func TestCompileToFileSetsShapeWiggleTransform(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.WiggleTransform = &recipe.WiggleTransformSpec{
		Anchor:           []float64{10, 12},
		Position:         []float64{80, 60},
		Scale:            []float64{20, 30},
		Rotation:         ptr(25),
		WigglesPerSecond: ptr(4),
		RandomSeed:       ptr(9),
		Correlation:      ptr(80),
		TemporalPhase:    ptr(45),
		SpatialPhase:     ptr(20),
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
	assertLayerPropertyValue(t, layer, "ADBE Vector Wiggler Anchor", []float64{10, 12})
	assertLayerPropertyValue(t, layer, "ADBE Vector Wiggler Position", []float64{80, 60})
	assertLayerPropertyValue(t, layer, "ADBE Vector Wiggler Scale", []float64{20, 30})
	assertLayerPropertyValue(t, layer, "ADBE Vector Wiggler Rotation", 25.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Xform Temporal Freq", 4.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Random Seed", 9.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Correlation", 80.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Temporal Phase", 45.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Spatial Phase", 20.0)
}

func TestCompileToFileSetsTextStyle(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].TextStyle = &recipe.TextStyleSpec{
		FontSize:      ptr(96),
		FillColor:     []float64{64, 128, 255, 255},
		Tracking:      ptr(120),
		FauxBold:      boolPtr(true),
		FauxItalic:    boolPtr(true),
		ApplyStroke:   boolPtr(true),
		StrokeColor:   []float64{255, 32, 64, 255},
		StrokeWidth:   ptr(8),
		Justification: "center",
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
	layer := findProfileLayer(t, prof, "Title")
	if layer.Text == nil || len(layer.Text.Runs) == 0 || len(layer.Text.Paragraphs) == 0 {
		t.Fatalf("text profile missing runs/paragraphs: %+v", layer.Text)
	}
	if got := layer.Text.Runs[0].FontSize; got != 96 {
		t.Fatalf("FontSize = %v, want 96", got)
	}
	assertFloatArray(t, "FillColor", layer.Text.Runs[0].FillColor[:], []float64{64.0 / 255.0, 128.0 / 255.0, 1, 1})
	if got := layer.Text.Runs[0].Tracking; got != 120 {
		t.Fatalf("Tracking = %v, want 120", got)
	}
	if !layer.Text.Runs[0].FauxBold {
		t.Fatal("FauxBold = false, want true")
	}
	if !layer.Text.Runs[0].FauxItalic {
		t.Fatal("FauxItalic = false, want true")
	}
	if !layer.Text.Runs[0].ApplyStroke {
		t.Fatal("ApplyStroke = false, want true")
	}
	assertFloatArray(t, "StrokeColor", layer.Text.Runs[0].StrokeColor[:], []float64{1, 32.0 / 255.0, 64.0 / 255.0, 1})
	if got := layer.Text.Runs[0].StrokeWidth; got != 8 {
		t.Fatalf("StrokeWidth = %v, want 8", got)
	}
	if got := layer.Text.Paragraphs[0].Justification; got != "Center" {
		t.Fatalf("Justification = %q, want Center", got)
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

func TestCompileToFileChecksExpectedTextStyleProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].TextStyle = &recipe.TextStyleSpec{
		FontSize:      ptr(96),
		FillColor:     []float64{64, 128, 255, 255},
		Tracking:      ptr(120),
		FauxBold:      boolPtr(true),
		FauxItalic:    boolPtr(true),
		ApplyStroke:   boolPtr(true),
		StrokeColor:   []float64{255, 32, 64, 255},
		StrokeWidth:   ptr(8),
		Justification: "center",
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		TextStyles: []recipe.ExpectedTextStyle{{
			LayerName:      "Title",
			RunIndex:       0,
			ParagraphIndex: 0,
			FontSize:       ptr(96),
			FillColor:      []float64{64.0 / 255.0, 128.0 / 255.0, 1, 1},
			Tracking:       ptr(120),
			FauxBold:       boolPtr(true),
			FauxItalic:     boolPtr(true),
			ApplyStroke:    boolPtr(true),
			StrokeColor:    []float64{1, 32.0 / 255.0, 64.0 / 255.0, 1},
			StrokeWidth:    ptr(8),
			Justification:  "center",
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
	assertProfileCheck(t, report, "expected_profile.text_styles[0].font_size", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].fill_color", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].tracking", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].faux_bold", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].faux_italic", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].apply_stroke", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].stroke_color", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].stroke_width", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].justification", true)
}

func TestCompileToFileChecksExpectedKeyframeProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.Position = nil
	rec.Comps[0].Layers[0].Transform.PositionKeyframes = []recipe.VectorKeyframe{
		{Time: 0, Value: []float64{900, 540}},
		{Time: 1, Value: []float64{1020, 540}},
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Keyframes: []recipe.ExpectedKeyframedProperty{{
			LayerName: "Title",
			MatchName: "ADBE Position",
			Keyframes: []recipe.ExpectedKeyframe{
				{Time: 0, Value: []float64{900, 540, 0}},
				{Time: 1, Value: []float64{1020, 540, 0}},
			},
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
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
}

func TestCompileToFileChecksExpectedOpacityKeyframeProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.Opacity = nil
	rec.Comps[0].Layers[0].Transform.OpacityKeyframes = []recipe.ScalarKeyframe{
		{Time: 0, Value: 20},
		{Time: 1, Value: 100},
		{Time: 2, Value: 40},
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Keyframes: []recipe.ExpectedKeyframedProperty{{
			LayerName: "Title",
			MatchName: "ADBE Opacity",
			Keyframes: []recipe.ExpectedKeyframe{
				{Time: 0, Value: 0.2},
				{Time: 1, Value: 1.0},
				{Time: 2, Value: 0.4},
			},
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
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[2]", true)
}

func TestCompileToFileChecksExpectedScaleKeyframeProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.Scale = nil
	rec.Comps[0].Layers[0].Transform.ScaleKeyframes = []recipe.VectorKeyframe{
		{Time: 0, Value: []float64{80, 80}},
		{Time: 1, Value: []float64{100, 120}},
		{Time: 2, Value: []float64{130, 90}},
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Keyframes: []recipe.ExpectedKeyframedProperty{{
			LayerName: "Title",
			MatchName: "ADBE Scale",
			Keyframes: []recipe.ExpectedKeyframe{
				{Time: 0, Value: []float64{0.8, 0.8, 1}},
				{Time: 1, Value: []float64{1, 1.2, 1}},
				{Time: 2, Value: []float64{1.3, 0.9, 1}},
			},
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
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[2]", true)
}

func TestCompileToFileChecksExpectedRotationKeyframeProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.Rotation = nil
	rec.Comps[0].Layers[0].Transform.RotationKeyframes = []recipe.ScalarKeyframe{
		{Time: 0, Value: 0},
		{Time: 1, Value: 45},
		{Time: 2, Value: -30},
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Keyframes: []recipe.ExpectedKeyframedProperty{{
			LayerName: "Title",
			MatchName: "ADBE Rotate Z",
			Keyframes: []recipe.ExpectedKeyframe{
				{Time: 0, Value: 0.0},
				{Time: 1, Value: 45.0},
				{Time: 2, Value: -30.0},
			},
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
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[2]", true)
}

func TestCompileToFileChecksExpectedAnchorPointKeyframeProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.AnchorPoint = nil
	rec.Comps[0].Layers[0].Transform.AnchorPointKeyframes = []recipe.VectorKeyframe{
		{Time: 0, Value: []float64{0, 0}},
		{Time: 1, Value: []float64{120, -40}},
		{Time: 2, Value: []float64{-60, 30}},
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Keyframes: []recipe.ExpectedKeyframedProperty{{
			LayerName: "Title",
			MatchName: "ADBE Anchor Point",
			Keyframes: []recipe.ExpectedKeyframe{
				{Time: 0, Value: []float64{0, 0, 0}},
				{Time: 1, Value: []float64{120, -40, 0}},
				{Time: 2, Value: []float64{-60, 30, 0}},
			},
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
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[2]", true)
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

func assertFloatArray(t *testing.T, label string, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s = %v, want %v", label, got, want)
		}
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
