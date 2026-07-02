package aepmigrate

import "testing"

func TestConvertWritesRecipeLayerLabelProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-layer-label.json")
}

func TestConvertWritesRecipeLayerCommentProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-layer-comment.json")
}

func TestConvertWritesRecipeLayerParentProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-layer-parent.json")
}
