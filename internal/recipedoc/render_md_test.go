package recipedoc

import (
	"strings"
	"testing"
)

func TestRenderMarkdownIncludesFieldReference(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}
	out := RenderMarkdown(doc)
	requireSubstring(t, out, "# Recipe Field Reference")
	requireSubstring(t, out, "## CompSpec")
	requireSubstring(t, out, "`comps[].background_color`")
	requireSubstring(t, out, "`comp.set_background_color`")
	requireSubstring(t, out, "structural")
}

func TestRenderMarkdownMarksMissingSemanticSummary(t *testing.T) {
	doc := Document{Fields: []FieldModel{{
		Path:                   "comps[].unknown_future_field",
		JSONName:               "unknown_future_field",
		SourceType:             "CompSpec",
		SourceField:            "UnknownFutureField",
		Type:                   TypeRef{Kind: "number", GoType: "float64"},
		MissingSemanticSummary: true,
	}}}
	out := RenderMarkdown(doc)
	requireSubstring(t, out, "missing semantic summary")
}

func requireSubstring(t *testing.T, s, want string) {
	t.Helper()
	if !strings.Contains(s, want) {
		t.Fatalf("missing %q in:\n%s", want, s)
	}
}
