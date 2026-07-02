package aepmigrate

import "testing"

func TestConvertWritesRecipeEffectLayerParamProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-effect-layer-param.json")
}

func TestConvertWritesRecipeEffectParamExpressionProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-effect-param-expression.json")
}

func TestConvertWritesRecipeEffectParamKeyframesProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-effect-param-keyframes.json")
}

func TestConvertWritesRecipeEffectParamVectorKeyframesProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-effect-param-vector-keyframes.json")
}
