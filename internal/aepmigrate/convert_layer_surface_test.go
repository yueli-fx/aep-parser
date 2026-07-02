package aepmigrate

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func TestConvertWritesDefaultSolidLayerProject(t *testing.T) {
	source := writeTempProjectWithOneSolidLayer(t)
	_, outPath := assertSourceConvertsPass(t, source)
	outProject, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open converted: %v", err)
	}
	prof, err := profile.Build(outProject, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile converted: %v", err)
	}
	if len(prof.Comps) != 1 || len(prof.Comps[0].Layers) != 1 {
		t.Fatalf("converted profile layers = %+v", prof.Comps)
	}
	layer := prof.Comps[0].Layers[0]
	if layer.Type != "av" || layer.SourceRef == nil || layer.SourceRef.Kind != "footage" {
		t.Fatalf("converted layer source = %+v, want av footage layer", layer)
	}
}

func TestConvertWritesRecipeDefaultSolidLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-layer-source-ref.json")
}

func TestConvertWritesDefaultAdjustmentLayerProject(t *testing.T) {
	source := writeTempProjectWithOneAdjustmentLayer(t)
	assertSourceConvertsPass(t, source)
}

func TestConvertWritesRecipeDefaultAdjustmentLayerProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-default-adjustment-layer.json")
}
