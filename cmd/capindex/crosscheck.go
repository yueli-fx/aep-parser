package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// validateEntries runs every tagged Entry through validateCap and cross-checks
// its gate references against the scanned test set and its incident slugs
// against the incidents directory. Returns all problems (anti-false-green:
// a verify=ae-accept/render-pixel claim whose gate is missing or disabled is an
// error; a dangling incident link is an error). Untagged entries are ignored
// here (coverage enforcement is separate).
func validateEntries(entries []Entry, gates map[string]bool, incidentsDir string) []error {
	var errs []error
	for _, e := range entries {
		if e.parseErr != nil {
			errs = append(errs, fmt.Errorf("%s: malformed aep:cap: %v", e.Symbol, e.parseErr))
			continue
		}
		if !e.HasCap {
			continue
		}
		c := e.Cap
		if err := validateCap(&c); err != nil {
			errs = append(errs, fmt.Errorf("%s: %v", e.Symbol, err))
			continue
		}
		for _, g := range c.Gate {
			skipped, ok := gates[g]
			if !ok {
				errs = append(errs, fmt.Errorf("%s: gate %q not found in any *_test.go", e.Symbol, g))
			} else if skipped {
				errs = append(errs, fmt.Errorf("%s: gate %q is unconditionally skipped (disabled)", e.Symbol, g))
			}
		}
		for _, slug := range c.Incident {
			if path, ok := incidentPath(incidentsDir, slug); !ok {
				errs = append(errs, fmt.Errorf("%s: incident %q not found under incidents/ or knowledge/ (%s)", e.Symbol, slug, path))
			}
		}
	}
	return errs
}

func incidentPath(incidentsDir, slug string) (string, bool) {
	candidates := []string{
		filepath.Join(incidentsDir, slug+".md"),
		filepath.Join(filepath.Dir(incidentsDir), "archive", "incidents", slug+".md"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
	}

	knowledgeRoot := filepath.Join(filepath.Dir(incidentsDir), "knowledge")
	found := ""
	_ = filepath.WalkDir(knowledgeRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		if d.Name() == slug+".md" {
			found = path
			return fs.SkipAll
		}
		return nil
	})
	if found != "" {
		return found, true
	}
	return filepath.Join(knowledgeRoot, "**", slug+".md"), false
}
