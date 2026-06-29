package recipedoc

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderJSONSchemaUsesJSONSchemaShape(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}
	out, err := RenderJSONSchema(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"$schema"`) {
		t.Fatal("missing $schema")
	}
	if !strings.Contains(out, `"x-aep-path": "comps[].background_color"`) {
		t.Fatal("missing x-aep-path for background color")
	}
	if !strings.Contains(out, `"x-aep-capabilities"`) {
		t.Fatal("missing capability extension")
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestRenderJSONSchemaIsDeterministic(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}
	a, err := RenderJSONSchema(doc)
	if err != nil {
		t.Fatal(err)
	}
	b, err := RenderJSONSchema(doc)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatal("schema output is not deterministic")
	}
}
