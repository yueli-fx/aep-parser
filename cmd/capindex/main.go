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

//go:generate go run .

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/example/aep-parser/internal/apidoc"
)

func main() {
	q := flag.String("q", "", "query term; print matching capabilities and exit")
	check := flag.Bool("check", false, "CI mode: validate tags + verify committed files are current")
	coverage := flag.Bool("coverage", false, "report public-surface tag coverage (P2 progress meter) and exit")
	validate := flag.Bool("validate", false, "validate @tag annotations against the schema and exit")
	flag.Parse()

	root, err := repoRoot()
	if err != nil {
		fatal(err)
	}
	pkgDirs := capindexPkgDirs(root)
	incidentsDir := filepath.Join(root, "flightdeck", "incidents")
	docgenPath := filepath.Join(root, "docs", "docgen.json")
	jsonPath := filepath.Join(root, "docs", "capabilities.json")
	mdPath := filepath.Join(root, "docs", "capabilities.md")

	entries, err := extractEntries(pkgDirs...)
	if err != nil {
		fatal(err)
	}

	if *validate {
		surface, err := loadSurface(docgenPath)
		if err != nil {
			fatal(err)
		}
		gates, err := scanTests(filepath.Join(root, "internal"))
		if err != nil {
			fatal(err)
		}
		ctx := apidoc.Context{Gates: gates, IncidentsDir: incidentsDir}
		errs := runValidate(entries, surface, ctx, apidoc.ModeWarn)
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "capindex:", e)
		}
		if len(errs) > 0 {
			os.Exit(1)
		}
		fmt.Println("capindex: @tag validation clean")
		return
	}

	if *coverage {
		surface, err := loadSurface(docgenPath)
		if err != nil {
			fatal(err)
		}
		reportCoverage(surface, entries)
		return
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

// capindexPkgDirs is the package set scanned for the capability surface, in
// facade-priority order. internal/aep holds the public re-export funcs + type
// aliases; internal/scene holds the real types whose Set*/getter methods are the
// bulk of the API. (codec capability is reached via facade_codec.go re-exports,
// so codec itself is not scanned.)
func capindexPkgDirs(root string) []string {
	return []string{
		filepath.Join(root, "internal", "aep"),
		filepath.Join(root, "internal", "scene"),
	}
}

// reportCoverage prints the P2 progress meter: how much of the documented public
// surface carries an aep:cap tag, and the still-untagged symbols.
func reportCoverage(s *publicSurface, entries []Entry) {
	tagged, untagged := s.coverage(entries)
	total := len(tagged) + len(untagged)
	pct := 0.0
	if total > 0 {
		pct = 100 * float64(len(tagged)) / float64(total)
	}
	fmt.Printf("capindex coverage: %d/%d public-surface symbols tagged (%.1f%%) — %d getter/reader methods exempt (write-surface-first policy)\n", len(tagged), total, pct, s.getterExempt(entries))
	if len(untagged) == 0 {
		return
	}
	fmt.Printf("\nuntagged (%d):\n", len(untagged))
	for _, e := range untagged {
		sym := e.Symbol
		if e.Recv != "" {
			sym = strings.TrimPrefix(e.Recv, "*") + "." + e.Symbol
		}
		fmt.Printf("  %s\t(%s %s)\n", sym, e.Kind, e.Pkg)
	}
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
