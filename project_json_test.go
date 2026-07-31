package aep_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/yueli-fx/aep-parser"
)

func TestProjectJSONIsDetachedVersionedAndPathSafe(t *testing.T) {
	path := filepath.Join("test_data", "fixtures", "v2_2_transform_kf_re.aep")
	document, err := aep.Open(path)
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
			Compositions []struct {
				Layers []struct {
					Properties []json.RawMessage `json:"properties"`
				} `json:"layers"`
			} `json:"compositions"`
			Footage []struct {
				Path string `json:"path"`
			} `json:"footage"`
			RenderQueue json.RawMessage `json:"render_queue"`
		} `json:"project"`
		PropertyTrees []struct {
			Children []json.RawMessage `json:"children"`
		} `json:"property_trees"`
	}
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		t.Fatalf("decode ProjectJSON: %v", err)
	}
	if snapshot.SchemaVersion != aep.ProjectJSONSchemaVersion {
		t.Fatalf("schema version = %d", snapshot.SchemaVersion)
	}
	if len(snapshot.Project.Compositions) != 1 ||
		len(snapshot.Project.Compositions[0].Layers) != 1 ||
		len(snapshot.Project.Compositions[0].Layers[0].Properties) == 0 ||
		len(snapshot.PropertyTrees) != 1 ||
		len(snapshot.PropertyTrees[0].Children) == 0 {
		t.Fatalf("unexpected detached project summary: %+v", snapshot)
	}
	for _, footage := range snapshot.Project.Footage {
		if footage.Path != "" {
			t.Fatalf("footage path leaked: %q", footage.Path)
		}
	}
	if len(snapshot.Project.RenderQueue) != 0 &&
		string(snapshot.Project.RenderQueue) != "null" {
		t.Fatalf("render queue leaked: %s", snapshot.Project.RenderQueue)
	}
	if strings.Contains(string(payload), filepath.Clean(path)) {
		t.Fatal("input path leaked into ProjectJSON")
	}

	var nilDocument *aep.Document
	if _, err := nilDocument.ProjectJSON(); err == nil {
		t.Fatal("ProjectJSON succeeded on nil document")
	}
}

func TestClassifyErrorReturnsStableCategories(t *testing.T) {
	path := filepath.Join("test_data", "fixtures", "v2_smoke_ae2020.aep")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	limits := aep.DefaultLimits()
	limits.MaxInputBytes = uint64(len(data) - 1)
	_, parseErr := aep.ParseWithLimits(bytes.NewReader(data), limits)
	info, ok := aep.ClassifyError(parseErr)
	if !ok || info.Code != "resource-limit" ||
		info.Resource != "input bytes" ||
		info.Actual != uint64(len(data)) {
		t.Fatalf("limit classification = %+v, %v", info, ok)
	}

	_, parseErr = aep.Parse(bytes.NewReader(nil))
	info, ok = aep.ClassifyError(parseErr)
	if !ok || info.Code != "invalid-format" {
		t.Fatalf("format classification = %+v, %v (%v)", info, ok, parseErr)
	}
	if _, ok := aep.ClassifyError(nil); ok {
		t.Fatal("nil error classified")
	}
}
