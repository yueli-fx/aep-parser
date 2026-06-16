// Command capindex generates the source-driven capability index for
// internal/aep from `aep:cap` doc-comment tags: docs/capabilities.json (the
// machine-queryable index) + docs/capabilities.md (the human table). It also
// serves queries (-q) and a CI drift/validation check (-check).
//
// Usage (from anywhere in the module tree):
//
//	go run ./cmd/capindex            # regenerate docs/capabilities.{json,md}
//	go run ./cmd/capindex -q solid   # print capabilities matching "solid"
//	go run ./cmd/capindex -check     # CI: validate tags + verify files up to date
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	q := flag.String("q", "", "query term; print matching capabilities and exit")
	check := flag.Bool("check", false, "CI mode: validate tags + verify committed files are current")
	flag.Parse()

	root, err := repoRoot()
	if err != nil {
		fatal(err)
	}
	pkgDir := filepath.Join(root, "internal", "aep")
	incidentsDir := filepath.Join(root, "flightdeck", "incidents")
	jsonPath := filepath.Join(root, "docs", "capabilities.json")
	mdPath := filepath.Join(root, "docs", "capabilities.md")

	entries, err := extractEntries(pkgDir)
	if err != nil {
		fatal(err)
	}

	if *q != "" {
		hits := query(entries, *q)
		for _, e := range hits {
			printEntry(e)
		}
		if len(hits) == 0 {
			fmt.Printf("no capability matches %q\n", *q)
		}
		return
	}

	gates, err := scanTests(filepath.Join(root, "internal"))
	if err != nil {
		fatal(err)
	}
	if errs := validateEntries(entries, gates, incidentsDir); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "capindex:", e)
		}
		os.Exit(1)
	}

	jsonOut, err := renderJSON(entries)
	if err != nil {
		fatal(err)
	}
	mdOut := renderMarkdown(entries)

	if *check {
		if !fileMatches(jsonPath, jsonOut) || !fileMatches(mdPath, mdOut) {
			fmt.Fprintln(os.Stderr, "capindex: docs/capabilities.{json,md} out of date — run `go generate ./...`")
			os.Exit(1)
		}
		return
	}

	if err := os.WriteFile(jsonPath, jsonOut, 0o644); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(mdPath, mdOut, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("capindex: wrote %d capabilities → %s + %s\n", len(taggedSorted(entries)), rel(root, jsonPath), rel(root, mdPath))
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

func fileMatches(path string, want []byte) bool {
	got, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return bytes.Equal(got, want)
}

func rel(root, p string) string {
	if r, err := filepath.Rel(root, p); err == nil {
		return r
	}
	return p
}

func printEntry(e Entry) {
	sym := e.Symbol
	if e.Recv != "" {
		sym = e.Recv + "." + e.Symbol
	}
	fmt.Printf("● %s  [%s · %s · verify=%s · AE≥%s]\n", sym, e.Cap.Domain, e.Cap.Tier, e.Cap.Verify, e.Cap.MinVer)
	if e.Signature != "" {
		fmt.Printf("    %s\n", e.Signature)
	}
	if e.Summary != "" {
		fmt.Printf("    %s\n", e.Summary)
	}
	if len(e.Cap.Gate) > 0 {
		fmt.Printf("    gate: %s\n", strings.Join(e.Cap.Gate, ", "))
	}
	if e.Cap.Boundary != "" {
		fmt.Printf("    ⚠ %s\n", e.Cap.Boundary)
	}
	if len(e.Cap.Incident) > 0 {
		fmt.Printf("    incident: %s\n", strings.Join(e.Cap.Incident, ", "))
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "capindex:", err)
	os.Exit(1)
}
