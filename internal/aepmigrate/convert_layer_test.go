package aepmigrate

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func TestConvertWritesDefaultNullLayerProject(t *testing.T) {
	source := writeTempProjectWithDefaultNullLayer(t)
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
	if got := prof.Comps[0].Layers[0].Type; got != "null" {
		t.Fatalf("layer type = %q, want null", got)
	}
}

func TestConvertWritesRecipeDefaultNullLayerProject(t *testing.T) {
	source := writeTempRecipeDefaultNullLayer(t)
	assertSourceConvertsPass(t, source)
}
