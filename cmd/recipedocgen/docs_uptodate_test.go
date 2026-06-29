package main

import (
	"os"
	"strings"
	"testing"
)

func TestRecipeDocsUpToDate(t *testing.T) {
	schema, markdown, err := generate("../../docs")
	if err != nil {
		t.Fatal(err)
	}
	assertFileCurrent(t, "../../docs/recipe_schema.json", schema)
	assertFileCurrent(t, "../../docs/recipe.md", markdown)
}

func assertFileCurrent(t *testing.T, path, got string) {
	t.Helper()
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	norm := func(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
	if norm(string(want)) != norm(got) {
		t.Fatalf("%s is stale; run go generate ./cmd/recipedocgen", path)
	}
}
