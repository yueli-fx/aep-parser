package aepmigrate

import "testing"

func TestConvertWritesRecipeLayerCommonSwitchesProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-layer-common-switches.json")
}

func TestConvertWritesRecipeLayerShyProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-layer-shy.json")
}

func TestConvertWritesRecipeLayerMotionBlurProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-layer-motion-blur.json")
}

func TestConvertWritesRecipeLayerQualityBlendingProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-layer-quality-blending.json")
}
