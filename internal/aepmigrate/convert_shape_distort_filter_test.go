package aepmigrate

import "testing"

func TestConvertWritesRecipeRectZigZagShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-zigzag.json")
}

func TestConvertWritesRecipeRectPuckerBloatShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-pucker-bloat.json")
}

func TestConvertWritesRecipeRectTwistShapeLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-shape-twist.json")
}
