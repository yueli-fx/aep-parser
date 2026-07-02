package aepmigrate

import "testing"

func TestConvertWritesDefaultPrecompLayerProject(t *testing.T) {
	source := writeTempProjectWithOnePrecompLayer(t)
	assertSourceConvertsPass(t, source)
}

func TestConvertWritesRecipeDefaultPrecompLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-default-precomp-layer.json")
}
