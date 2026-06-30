package recipedoc

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderRecipeIndexBuildsPathLookup(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}
	out, err := RenderRecipeIndex(doc)
	if err != nil {
		t.Fatal(err)
	}

	var parsed struct {
		SchemaVersion int                       `json:"schema_version"`
		Paths         []string                  `json:"paths"`
		Fields        map[string]recipeIndexRow `json:"fields"`
		BySourceType  map[string][]string       `json:"by_source_type"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed.SchemaVersion != doc.SchemaVersion {
		t.Fatalf("schema_version = %d, want %d", parsed.SchemaVersion, doc.SchemaVersion)
	}
	if len(parsed.Paths) != len(doc.Fields) {
		t.Fatalf("paths len = %d, want %d", len(parsed.Paths), len(doc.Fields))
	}
	field := parsed.Fields["comps[].background_color"]
	if field.Path != "comps[].background_color" {
		t.Fatalf("missing background color field: %+v", field)
	}
	if field.Type != "array<float64>" {
		t.Fatalf("type = %q", field.Type)
	}
	if !containsString(field.CapabilityKeys, "comp.set_background_color") {
		t.Fatalf("capability keys = %+v", field.CapabilityKeys)
	}
	if field.Example != "examples/recipes/minimal-comp-background-color.json" {
		t.Fatalf("example = %q", field.Example)
	}
	if !containsString(parsed.BySourceType["CompSpec"], "comps[].background_color") {
		t.Fatalf("CompSpec paths missing background color: %+v", parsed.BySourceType["CompSpec"])
	}
}

type recipeIndexRow struct {
	Path           string   `json:"path"`
	Type           string   `json:"type"`
	CapabilityKeys []string `json:"capability_keys"`
	Example        string   `json:"example"`
}

func TestRenderRecipeIndexIsDeterministic(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}
	a, err := RenderRecipeIndex(doc)
	if err != nil {
		t.Fatal(err)
	}
	b, err := RenderRecipeIndex(doc)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatal("recipe index output is not deterministic")
	}
}

func TestRenderRecipeIndexEscapesNewlines(t *testing.T) {
	doc := Document{Fields: []FieldModel{{
		Path:       "comps[].field",
		JSONName:   "field",
		SourceType: "CompSpec",
		Type:       TypeRef{Kind: "string", GoType: "string"},
		Summary:    "line one\nline two",
	}}}
	out, err := RenderRecipeIndex(doc)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "line one\nline two") {
		t.Fatal("raw newline should be JSON escaped")
	}
	if !strings.Contains(out, `line one\nline two`) {
		t.Fatalf("missing escaped newline in:\n%s", out)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
