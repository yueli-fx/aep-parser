package aep_test

import (
	"encoding/json"
	"reflect"
	"testing"

	aep "github.com/yueli-fx/aep-parser"
)

type projectItemSnapshot struct {
	ID       uint32 `json:"id"`
	Kind     string `json:"kind"`
	ParentID uint32 `json:"parent_id"`
	Order    int    `json:"order"`
}

func TestProjectJSONPreservesProjectPanelHierarchyAndOrder(t *testing.T) {
	document, err := aep.Open("test_data/fixtures/renderer_ae2020_r0.aep")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	payload, err := document.ProjectJSON()
	if err != nil {
		t.Fatalf("ProjectJSON: %v", err)
	}
	var snapshot struct {
		SchemaVersion int `json:"schema_version"`
		Project       struct {
			Items []projectItemSnapshot `json:"items"`
		} `json:"project"`
	}
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		t.Fatalf("decode ProjectJSON: %v", err)
	}
	if snapshot.SchemaVersion != aep.ProjectJSONSchemaVersion {
		t.Fatalf("schema_version = %d, want %d", snapshot.SchemaVersion, aep.ProjectJSONSchemaVersion)
	}

	want := []projectItemSnapshot{
		{ID: 1, Kind: "composition", ParentID: 0, Order: 0},
		{ID: 13, Kind: "folder", ParentID: 0, Order: 1},
		{ID: 14, Kind: "footage", ParentID: 13, Order: 0},
	}
	if !reflect.DeepEqual(snapshot.Project.Items, want) {
		t.Fatalf("project items = %+v, want %+v", snapshot.Project.Items, want)
	}
}

func TestProjectJSONPreservesSolidsFolderMembership(t *testing.T) {
	document, err := aep.Open("test_data/fixtures/re_shapes.aep")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	payload, err := document.ProjectJSON()
	if err != nil {
		t.Fatalf("ProjectJSON: %v", err)
	}
	var snapshot struct {
		Project struct {
			Items []projectItemSnapshot `json:"items"`
		} `json:"project"`
	}
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		t.Fatalf("decode ProjectJSON: %v", err)
	}

	childCount := 0
	for _, item := range snapshot.Project.Items {
		if item.ParentID != 19 {
			continue
		}
		if item.Kind != "footage" || item.Order != childCount {
			t.Fatalf("Solids child %d = %+v", childCount, item)
		}
		childCount++
	}
	if childCount != 13 {
		t.Fatalf("Solids folder children = %d, want 13", childCount)
	}
}
