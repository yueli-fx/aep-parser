package aepmigrate

import "testing"

func TestConvertWritesRecipeTransformKeyframesProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-transform-keyframes.json")
}

func TestConvertWritesRecipeTransformKeyframeEaseProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-transform-keyframe-ease.json")
}

func TestConvertWritesRecipeTransformExpressionProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-transform-expression.json")
}
