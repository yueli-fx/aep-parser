package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGeneratedUpToDate is the CI drift guard: the committed
// docs/capabilities.{json,md} must equal a fresh regeneration, and every tagged
// capability must pass validation (gate exists & live, incident link resolves,
// tier×verify consistent). Mirrors cmd/docgen's docs_uptodate_test.go.
func TestGeneratedUpToDate(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	entries, err := extractEntries(filepath.Join(root, "internal", "aep"))
	if err != nil {
		t.Fatal(err)
	}
	gates, err := scanTests(filepath.Join(root, "internal"))
	if err != nil {
		t.Fatal(err)
	}
	if errs := validateEntries(entries, gates, filepath.Join(root, "flightdeck", "incidents")); len(errs) > 0 {
		for _, e := range errs {
			t.Error(e)
		}
		t.Fatal("aep:cap validation failed")
	}
	jsonOut, err := renderJSON(entries)
	if err != nil {
		t.Fatal(err)
	}
	mustMatch(t, filepath.Join(root, "docs", "capabilities.json"), jsonOut)
	mustMatch(t, filepath.Join(root, "docs", "capabilities.md"), renderMarkdown(entries))
}

func mustMatch(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != string(want) {
		t.Errorf("%s is out of date — run `go generate ./...`", path)
	}
}
