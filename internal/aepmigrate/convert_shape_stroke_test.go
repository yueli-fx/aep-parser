package aepmigrate

import "testing"

func TestConvertWritesRecipeRectStrokeShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-stroke-style.json")
}

func TestConvertWritesRecipeRectStrokeDashesShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-stroke-dashes.json")
}

func TestConvertWritesRecipeEllipseStrokeTaperShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-stroke-taper.json")
}

func TestConvertWritesRecipeEllipseStrokeWaveShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-stroke-wave.json")
}
