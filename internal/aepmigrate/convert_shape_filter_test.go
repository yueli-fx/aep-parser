package aepmigrate

import "testing"

func TestConvertWritesRecipeRectRoundCornersShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-round-corners.json")
}

func TestConvertWritesRecipeRectOffsetPathsShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-offset-paths.json")
}

func TestConvertWritesRecipeRectTrimPathsShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-trim.json")
}
