package aepmigrate

import "testing"

func TestConvertWritesDefaultCameraLayerProject(t *testing.T) {
	source := writeTempProjectWithOneCameraLayer(t)
	assertSourceConvertsPass(t, source)
}

func TestConvertWritesRecipeDefaultCameraLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-default-camera-layer.json")
}

func TestConvertWritesRecipeCameraObjectProfile(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-camera-object-profile.json")
}
