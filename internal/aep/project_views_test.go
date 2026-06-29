package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestProjectViews_ReBatch(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_batch.aep")
	if err != nil {
		t.Skipf("re_batch.aep not present: %v", err)
	}
	// Footages() == .Footage slice (alias).
	if &proj.Footages()[0] != &proj.Footage[0] {
		t.Errorf("Footages() does not alias .Footage slice")
	}

	// RootFolder is the first Folder when any folders exist.
	if len(proj.Folders) > 0 && proj.RootFolder() != proj.Folders[0] {
		t.Errorf("RootFolder != Folders[0]")
	}

	// LayerByID: pick the first layer of the first comp; cross-comp
	// lookup should find it.
	var firstLayer *aep.Layer
	for _, c := range proj.Compositions {
		if len(c.Layers) > 0 && c.Layers[0].ID != 0 {
			firstLayer = c.Layers[0]
			break
		}
	}
	if firstLayer == nil {
		t.Skip("no layer with ID > 0")
	}
	got := proj.LayerByID(firstLayer.ID)
	if got != firstLayer {
		t.Errorf("LayerByID(%d) returned wrong layer; got %v want %v", firstLayer.ID, got, firstLayer)
	}
	if proj.LayerByID(0) != nil {
		t.Errorf("LayerByID(0) should return nil")
	}
	if proj.LayerByID(0xFFFFFFFF) != nil {
		t.Errorf("LayerByID(huge) should return nil")
	}

	// EffectNames: re_batch.aep has Gaussian Blur, Tritone, etc.
	names := proj.EffectNames()
	t.Logf("EffectNames: %v", names)
	hasGB, hasTritone := false, false
	for _, n := range names {
		if n == "ADBE Gaussian Blur 2" {
			hasGB = true
		}
		if n == "ADBE Tritone" {
			hasTritone = true
		}
	}
	if !hasGB && !hasTritone {
		t.Errorf("EffectNames=%v: expected at least Gaussian Blur or Tritone", names)
	}
}

func TestProjectViews_Empty(t *testing.T) {
	p := &aep.Project{}
	if got := p.Footages(); len(got) != 0 {
		t.Errorf("empty Footages: got %v, want empty", got)
	}
	if p.RootFolder() != nil {
		t.Error("empty RootFolder: should be nil")
	}
	if p.LayerByID(1) != nil {
		t.Error("empty LayerByID(1): should be nil")
	}
	if p.EffectNames() != nil {
		t.Error("empty EffectNames: should be nil")
	}
}
