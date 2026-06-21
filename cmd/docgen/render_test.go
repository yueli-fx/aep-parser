package main

import (
	"os"
	"strings"
	"testing"

	"github.com/example/aep-parser/internal/apidoc"
)

func TestRenderSymbol_ParamTable(t *testing.T) {
	s := symbol{
		name: "AddMask", kind: kindMethod, signature: "func (l *Layer) AddMask(path BezierPath) (*Mask, error)",
		annotated: true, summary: "Add a vector mask to a layer",
		doc:     "Appends a closed Bezier mask to the layer's Mask Parade.",
		params:  []apidoc.Param{{Name: "path", Desc: "outline in layer-pixel space"}},
		returns: "the created mask",
	}
	var b strings.Builder
	renderSymbol(&b, "Layer", s)
	out := b.String()
	for _, want := range []string{
		"Add a vector mask to a layer",              // summary line
		"Appends a closed Bezier mask",              // description prose
		"| Parameter | Description |",               // param table header
		"| `path` | outline in layer-pixel space |", // param row
		"**Returns:** the created mask",             // returns line
	} {
		if !strings.Contains(out, want) {
			t.Errorf("render missing %q in:\n%s", want, out)
		}
	}
}

func TestRenderType_Widget_Golden(t *testing.T) {
	lp, _ := loadPackage("./testdata/sample")
	types := withMethods(extractTypes(lp), lp)
	attachExamples(types, "./testdata/sample")
	w := findType(types, "Widget")

	got := renderType(w)
	want, err := os.ReadFile("./testdata/golden/widget.md")
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("render mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
