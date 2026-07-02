package aepmigrate

import "testing"

func TestConvertWritesRecipePolystarShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-polystar.json")
}

func TestConvertWritesRecipeGradientFillShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-gradient-fill.json")
}

func TestConvertWritesRecipeGradientStrokeShapeLayerProjects(t *testing.T) {
	recipes := []string{
		"minimal-shape-gradient-stroke.json",
		"minimal-shape-gradient-stroke-alpha-stops.json",
		"minimal-shape-gradient-stroke-highlight.json",
		"minimal-shape-gradient-stroke-style.json",
	}
	for _, recipeName := range recipes {
		t.Run(recipeName, func(t *testing.T) {
			assertRecipeConvertsPass(t, recipeName)
		})
	}
}
