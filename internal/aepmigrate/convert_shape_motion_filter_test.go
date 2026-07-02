package aepmigrate

import "testing"

func TestConvertWritesRecipeRectWigglePathsShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-wiggle-paths.json")
}

func TestConvertWritesRecipeRectWiggleTransformShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-wiggle-transform.json")
}

func TestConvertWritesRecipeRectRepeaterShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-repeater.json")
}

func TestConvertWritesRecipeRectMergePathsShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-merge-paths.json")
}
