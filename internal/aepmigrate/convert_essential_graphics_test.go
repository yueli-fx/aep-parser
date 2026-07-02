package aepmigrate

import "testing"

func TestConvertWritesRecipeMotionGraphicsTemplateNameProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-motion-graphics-template-name.json")
}

func TestConvertWritesRecipeEssentialGraphicsSliderProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-essential-graphics-controller.json")
}

func TestConvertWritesRecipeEssentialGraphicsColorProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-essential-graphics-color-controller.json")
}

func TestConvertWritesRecipeEssentialGraphicsCheckboxProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-essential-graphics-checkbox-controller.json")
}

func TestConvertWritesRecipeEssentialGraphicsPointProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-essential-graphics-point-controller.json")
}
