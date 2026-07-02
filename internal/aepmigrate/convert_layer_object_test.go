package aepmigrate

import "testing"

func TestConvertWritesRecipeLayerNullFlagProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-layer-null-flag.json")
}

func TestConvertWritesRecipeLayerAdvancedSwitchesProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-layer-advanced-switches.json")
}

func TestConvertWritesRecipeLayerObjectProfile(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-layer-object-profile.json")
}
