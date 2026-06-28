package recipe_test

import (
	"testing"

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

func ptr(v float64) *float64 {
	return &v
}
