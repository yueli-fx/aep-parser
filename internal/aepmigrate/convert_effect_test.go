package aepmigrate

import "testing"

func TestConvertWritesRecipeTextEffectProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-text-effect.json")
}

func TestConvertWritesRecipeAdjustmentEffectProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-adjustment-layer.json")
}
