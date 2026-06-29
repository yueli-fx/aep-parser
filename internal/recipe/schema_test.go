package recipe_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/recipe"
)

func TestValidateAcceptsMinimalTextShapeRecipe(t *testing.T) {
	rec := minimalRecipe()

	report := recipe.Validate(rec)

	if report.Valid != true {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	if len(report.Refusals) != 0 {
		t.Fatalf("Refusals = %+v, want none", report.Refusals)
	}
}

func TestValidateRejectsUnsupportedLayerType(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers = append(rec.Comps[0].Layers, recipe.Layer{
		Type: "camera",
		Name: "Camera 1",
	})

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "unsupported_layer_type")
}

func TestValidateRejectsOutOfRangeKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.PositionKeyframes = []recipe.VectorKeyframe{
		{Time: 5, Value: []float64{960, 540}},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframe_time_out_of_range")
}

func TestValidateRejectsUnsortedKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.PositionKeyframes = []recipe.VectorKeyframe{
		{Time: 1, Value: []float64{960, 540}},
		{Time: 0, Value: []float64{900, 540}},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframes_not_sorted")
}

func TestValidateRejectsOutOfRangeOpacityKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.OpacityKeyframes = []recipe.ScalarKeyframe{
		{Time: 5, Value: 50},
		{Time: 1, Value: 101},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframe_time_out_of_range")
	assertRefusal(t, report, "invalid_opacity_keyframe_value")
}

func TestValidateRejectsUnsortedOpacityKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.OpacityKeyframes = []recipe.ScalarKeyframe{
		{Time: 1, Value: 100},
		{Time: 0, Value: 50},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframes_not_sorted")
}

func TestValidateRejectsInvalidScaleKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.ScaleKeyframes = []recipe.VectorKeyframe{
		{Time: 5, Value: []float64{100, 100}},
		{Time: 1, Value: []float64{80}},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframe_time_out_of_range")
	assertRefusal(t, report, "invalid_vector_size")
}

func TestValidateRejectsUnsortedScaleKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.ScaleKeyframes = []recipe.VectorKeyframe{
		{Time: 1, Value: []float64{100, 100}},
		{Time: 0, Value: []float64{80, 80}},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "keyframes_not_sorted")
}

func TestValidateAcceptsExpectedKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Keyframes: []recipe.ExpectedKeyframedProperty{{
			LayerName: "Title",
			MatchName: "ADBE Position",
			Keyframes: []recipe.ExpectedKeyframe{
				{Time: 0, Value: []float64{900, 540}},
				{Time: 1, Value: []float64{1020, 540}},
			},
		}},
	}

	report := recipe.Validate(rec)

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
}

func TestValidateRejectsInvalidExpectedKeyframes(t *testing.T) {
	rec := minimalRecipe()
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Keyframes: []recipe.ExpectedKeyframedProperty{{
			LayerName: "",
			MatchName: "",
			Keyframes: []recipe.ExpectedKeyframe{
				{Time: -1, Value: []float64{}},
			},
		}},
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_expected_profile")
}

func TestValidateRefusesUnsupportedEffects(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{MatchName: "Third Party Magic"}}

	report := recipe.ValidateWithCapabilities(rec, recipe.StaticCapabilities{
		"Third Party Magic": recipe.CapabilityUnsupported,
	})

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "unsupported_effect")
}

func TestValidateReportsUsedCapabilities(t *testing.T) {
	report := recipe.ValidateWithCapabilities(minimalRecipe(), stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "NewComposition")
	assertCapability(t, report, "NewTextLayer")
	assertCapability(t, report, "Layer.SetText")
	assertCapability(t, report, "NewShapeLayer")
	assertCapability(t, report, "RectNode.SetSize")
	assertCapability(t, report, "FillNode.SetColor")
	assertCapability(t, report, "SetLayerTransform")
	if len(report.Downgrades) != 0 {
		t.Fatalf("Downgrades = %+v, want none", report.Downgrades)
	}
}

func TestValidateReportsShapeStrokeCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Stroke = &recipe.StrokeSpec{
		Color:   []float64{255, 0, 0, 255},
		Width:   ptr(6),
		Opacity: ptr(80),
	}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "VectorGroup.AddStroke")
	assertCapability(t, report, "StrokeNode.SetColor")
	assertCapability(t, report, "StrokeNode.SetWidth")
	assertCapability(t, report, "StrokeNode.SetOpacity")
}

func TestValidateReportsShapeDetailCapabilities(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Position = []float64{12, -6}
	rec.Comps[0].Layers[1].Shape.Roundness = ptr(18)
	rec.Comps[0].Layers[1].Shape.FillOpacity = ptr(45)

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "RectNode.SetPosition")
	assertCapability(t, report, "RectNode.SetRoundness")
	assertCapability(t, report, "FillNode.SetOpacity")
}

func TestValidateReportsTextStyleCapabilities(t *testing.T) {
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

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "Layer.SetRunFontSize")
	assertCapability(t, report, "Layer.SetRunFillColor")
	assertCapability(t, report, "Layer.SetRunTracking")
	assertCapability(t, report, "Layer.SetRunFauxBold")
	assertCapability(t, report, "Layer.SetRunFauxItalic")
	assertCapability(t, report, "Layer.SetRunApplyStroke")
	assertCapability(t, report, "Layer.SetRunStrokeColor")
	assertCapability(t, report, "Layer.SetRunStrokeWidth")
	assertCapability(t, report, "Layer.SetParagraphJustification")
}

func TestValidateRejectsInvalidTextStyle(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].TextStyle = &recipe.TextStyleSpec{
		RunIndex:       -1,
		ParagraphIndex: -1,
		FontSize:       ptr(0),
		FillColor:      []float64{255, 0},
		StrokeColor:    []float64{255, 0, 300},
		StrokeWidth:    ptr(-1),
		Justification:  "middle",
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_text_style_run_index")
	assertRefusal(t, report, "invalid_text_style_paragraph_index")
	assertRefusal(t, report, "invalid_text_font_size")
	assertRefusal(t, report, "invalid_text_fill_color")
	assertRefusal(t, report, "invalid_text_stroke_color")
	assertRefusal(t, report, "invalid_text_stroke_width")
	assertRefusal(t, report, "invalid_text_justification")
}

func TestValidateRejectsInvalidShapeDetail(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Position = []float64{12}
	rec.Comps[0].Layers[1].Shape.Roundness = ptr(-1)
	rec.Comps[0].Layers[1].Shape.FillOpacity = ptr(101)

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_vector_size")
	assertRefusal(t, report, "invalid_shape_roundness")
	assertRefusal(t, report, "invalid_shape_fill_opacity")
}

func TestValidateRejectsInvalidShapeStroke(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Stroke = &recipe.StrokeSpec{
		Color:   []float64{255, 0},
		Width:   ptr(-1),
		Opacity: ptr(101),
	}

	report := recipe.Validate(rec)

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "invalid_shape_stroke_color")
	assertRefusal(t, report, "invalid_shape_stroke_width")
	assertRefusal(t, report, "invalid_shape_stroke_opacity")
}

func TestValidateAcceptsSupportedEffects(t *testing.T) {
	effects := aep.SupportedEffects()
	if len(effects) == 0 {
		t.Fatal("SupportedEffects is empty")
	}
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{MatchName: effects[0]}}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "AddEffect")
}

func TestValidateAcceptsSupportedEffectParams(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{
		MatchName: "ADBE Gaussian Blur 2",
		Params: []recipe.EffectParam{
			{MatchName: "ADBE Gaussian Blur 2-0001", Value: 25.0},
			{MatchName: "ADBE Gaussian Blur 2-0003", Value: true},
		},
	}}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if !report.Valid {
		t.Fatalf("Valid = false, report=%+v", report)
	}
	assertCapability(t, report, "SetEffectParam")
}

func TestValidateRejectsUnsupportedEffectParamValue(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{
		MatchName: "ADBE Gaussian Blur 2",
		Params: []recipe.EffectParam{
			{MatchName: "ADBE Gaussian Blur 2-0001", Value: "high"},
		},
	}}

	report := recipe.ValidateWithCapabilities(rec, stableCapabilityIndex{})

	if report.Valid {
		t.Fatal("Valid = true, want false")
	}
	assertRefusal(t, report, "unsupported_effect_param_value")
}

func minimalRecipe() recipe.Recipe {
	return recipe.Recipe{
		SchemaVersion: recipe.SchemaVersion,
		Project:       recipe.ProjectSpec{Name: "Minimal title card"},
		Comps: []recipe.CompSpec{{
			Name:            "Main",
			Width:           1920,
			Height:          1080,
			FrameRate:       30,
			Duration:        4,
			BackgroundColor: []float64{0, 0, 0},
			Layers: []recipe.Layer{
				{
					Type: "text",
					Name: "Title",
					Text: "BOOYAH",
					Transform: recipe.Transform{
						Position: []float64{960, 540},
						Scale:    []float64{100, 100},
						Opacity:  ptr(100),
					},
				},
				{
					Type: "shape",
					Name: "Underline",
					Shape: &recipe.ShapeSpec{
						Kind:      "rect",
						Size:      []float64{640, 12},
						FillColor: []float64{255, 255, 255},
					},
					Transform: recipe.Transform{
						Position: []float64{960, 650},
					},
				},
			},
		}},
	}
}

func assertRefusal(t *testing.T, report recipe.Report, code string) {
	t.Helper()
	for _, refusal := range report.Refusals {
		if refusal.Code == code {
			return
		}
	}
	t.Fatalf("refusal %q not found in %+v", code, report.Refusals)
}

func assertCapability(t *testing.T, report recipe.Report, query string) {
	t.Helper()
	for _, use := range report.Capabilities {
		if use.Query == query {
			return
		}
	}
	t.Fatalf("capability %q not found in %+v", query, report.Capabilities)
}

func ptr(v float64) *float64 {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

type stableCapabilityIndex struct{}

func (stableCapabilityIndex) Lookup(query string) recipe.CapabilityLookup {
	return recipe.CapabilityLookup{
		Query:  query,
		Status: recipe.CapabilitySupported,
		Symbol: query,
		Domain: "test",
		Tier:   "stable",
		Verify: "ae-accept",
	}
}
