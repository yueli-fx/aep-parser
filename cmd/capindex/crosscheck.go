package main

import (
	"fmt"
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
			path := filepath.Join(incidentsDir, slug+".md")
			if _, err := os.Stat(path); err != nil {
				errs = append(errs, fmt.Errorf("%s: incident %q not found (%s)", e.Symbol, slug, path))
			}
		}
	}
	return errs
}
