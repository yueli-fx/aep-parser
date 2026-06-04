package main

import (
	"encoding/json"
	"testing"
)

func TestBuildIndex_SampleKeys(t *testing.T) {
	m, err := loadManifest("./testdata/manifest_sample.json")
	if err != nil {
		t.Fatal(err)
	}
	out, err := buildIndex(m)
	if err != nil {
		t.Fatal(err)
	}
	var idx docIndex
	if err := json.Unmarshal([]byte(out), &idx); err != nil {
		t.Fatalf("index is not valid JSON: %v", err)
	}

	// type entry → file + #type-object anchor
	w, ok := idx.Symbols["Widget"]
	if !ok || w.Kind != "type" || w.File != "out/sample_file.md" || w.Anchor != "#widget-object" {
		t.Fatalf("Widget type entry wrong: %+v", w)
	}
	// field with setter → read-write + json tag
	name, ok := idx.Symbols["Widget.Name"]
	if !ok || name.Kind != "field" || name.RW != "read-write" || name.JSON != "name" {
		t.Fatalf("Widget.Name entry wrong: %+v", name)
	}
	// field without setter → read-only
	if tags := idx.Symbols["Widget.Tags"]; tags.RW != "read-only" {
		t.Fatalf("Widget.Tags should be read-only: %+v", tags)
	}
	// method → #typemethod anchor + example flag
	sn, ok := idx.Symbols["Widget.SetName"]
	if !ok || sn.Kind != "method" || sn.Anchor != "#widgetsetname" || !sn.Example {
		t.Fatalf("Widget.SetName entry wrong: %+v", sn)
	}
}
