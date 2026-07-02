package aepmigrate

import "testing"

func TestConvertWritesDefaultShapeLayerProject(t *testing.T) {
	source := writeTempProjectWithDefaultShapeLayer(t)
	assertSourceConvertsPass(t, source)
}

func TestConvertWritesRecipeDefaultShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-default-shape-layer.json")
}

func TestConvertWritesRectFillShapeLayerProject(t *testing.T) {
	source := writeTempProjectWithRectFillShapeLayer(t)
	assertSourceConvertsPass(t, source)
}

func TestConvertWritesRecipeRectFillShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-rect-fill-default-transform.json")
}
