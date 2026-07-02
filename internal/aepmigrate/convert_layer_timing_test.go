package aepmigrate

import "testing"

func TestConvertWritesRecipeLayerTimingProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-layer-timing.json")
}

func TestConvertWritesMovedNullLayerProject(t *testing.T) {
	source := writeTempProjectWithMovedNullLayer(t)
	assertSourceConvertsPass(t, source)
}

func TestConvertWritesKeyframedNullLayerProject(t *testing.T) {
	source := writeTempProjectWithKeyframedNullLayer(t)
	assertSourceConvertsPass(t, source)
}
