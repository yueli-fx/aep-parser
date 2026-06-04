package main

import (
	"os"
	"strings"
	"testing"
)

// TestDocsUpToDate is the drift gate: it regenerates every file in the repo
// manifest (docs/docgen.json) and fails when a committed docs/*.md differs from
// what docgen would produce now — i.e. someone edited a doc comment but forgot
// `go generate ./cmd/docgen`. This stands in for a CI `git diff --exit-code
// docs/` step; the repo has no CI pipeline, so the check rides on `go test`.
//
// Comparison normalizes CRLF→LF: with core.autocrlf=true the working tree may
// carry CRLF while the generator emits LF, which is not real drift.
func TestDocsUpToDate(t *testing.T) {
	const manifestPath = "../../docs/docgen.json"
	if _, err := os.Stat(manifestPath); err != nil {
		t.Skipf("repo manifest not present (%v) — skipping drift gate", err)
	}
	m, err := loadManifest(manifestPath)
	if err != nil {
		t.Fatalf("loadManifest: %v", err)
	}
	norm := func(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
	for _, fm := range m.Files {
		got, err := generateFile(m, fm)
		if err != nil {
			t.Fatalf("generate %s: %v", fm.Out, err)
		}
		want, err := os.ReadFile(m.resolve(fm.Out))
		if err != nil {
			t.Fatalf("read %s: %v", fm.Out, err)
		}
		if norm(got) != norm(string(want)) {
			t.Errorf("%s is stale — run `go generate ./cmd/docgen` and commit", fm.Out)
		}
	}
	if m.Index != "" {
		got, err := buildIndex(m)
		if err != nil {
			t.Fatalf("buildIndex: %v", err)
		}
		want, err := os.ReadFile(m.resolve(m.Index))
		if err != nil {
			t.Fatalf("read %s: %v", m.Index, err)
		}
		if norm(got) != norm(string(want)) {
			t.Errorf("%s is stale — run `go generate ./cmd/docgen` and commit", m.Index)
		}
	}
}
