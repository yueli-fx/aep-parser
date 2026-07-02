package aepmigrate

import "testing"

func TestConvertWritesDefaultLightLayerProject(t *testing.T) {
	source := writeTempProjectWithOneLightLayer(t)
	assertSourceConvertsPass(t, source)
}

func TestConvertWritesRecipeDefaultLightLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-default-light-layer.json")
}

func TestConvertWritesRecipeLightObjectProfile(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-light-object-profile.json")
}

func TestConvertWritesRecipeLightSourceProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-light-source.json")
}
