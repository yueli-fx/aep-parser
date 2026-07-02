package aepmigrate

import "testing"

func TestConvertWritesRecipeTextStyleProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-text-style.json")
}

func TestConvertWritesRecipeTextShapeProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-text-shape.json")
}

func TestConvertWritesRecipeTextStaticTransformProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-default-text-static-transform.json")
}

func TestConvertWritesDefaultTextLayerProject(t *testing.T) {
	source := writeTempProjectWithDefaultTextLayer(t)
	assertSourceConvertsPass(t, source)
}

func TestConvertWritesRecipeDefaultTextLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-default-text-layer.json")
}
