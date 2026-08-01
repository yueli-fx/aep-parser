package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestNewCompositionUpdatesProjectPanelItems(t *testing.T) {
	project := aep.NewProject(aep.TargetAE2020)
	composition, err := aep.NewComposition(project, "Main", 1920, 1080, 30, 10)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	want := aep.ProjectItem{ID: composition.ID, Kind: aep.ItemTypeComposition, Order: 0}
	if len(project.Items) != 1 || project.Items[0] != want {
		t.Fatalf("project items = %+v, want %+v", project.Items, want)
	}
}

func TestDuplicateCompositionUpdatesMixedProjectPanelOrder(t *testing.T) {
	project, err := aep.Open("../../test_data/fixtures/renderer_ae2020_r0.aep")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	duplicate, err := aep.DuplicateComposition(project, project.Compositions[0], "Renderer Copy")
	if err != nil {
		t.Fatalf("DuplicateComposition: %v", err)
	}

	want := []aep.ProjectItem{
		{ID: 1, Kind: aep.ItemTypeComposition, Order: 0},
		{ID: duplicate.ID, Kind: aep.ItemTypeComposition, Order: 1},
		{ID: 13, Kind: aep.ItemTypeFolder, Order: 2},
		{ID: 14, Kind: aep.ItemTypeFootage, ParentID: 13, Order: 0},
	}
	if len(project.Items) != len(want) {
		t.Fatalf("project items = %+v, want %+v", project.Items, want)
	}
	for index := range want {
		if project.Items[index] != want[index] {
			t.Fatalf("project item %d = %+v, want %+v", index, project.Items[index], want[index])
		}
	}
}
