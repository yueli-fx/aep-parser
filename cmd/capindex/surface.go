package main

import (
	"encoding/json"
	"os"
	"strings"
)

// docgenManifest is the subset of docs/docgen.json capindex needs: docgen already
// curates the documented public API (per-file root types + free functions), so it
// doubles as the single source of "what is public" — capindex layers capability
// metadata on top instead of inventing a second notion of the surface.
type docgenManifest struct {
	Files []struct {
		Roots []string `json:"roots"`
		Funcs []string `json:"funcs"`
	} `json:"files"`
}

// publicSurface is the set of documented capability symbols that must carry an
// aep:cap tag: every facade (pkg=aep) exported func, plus every exported method
// whose receiver type is a docgen "root" type (Layer/Composition/Project/…). Enum
// consts and type aliases are the meta lane, gated separately.
type publicSurface struct {
	roots map[string]bool // root type names → their methods are public
	funcs map[string]bool // explicitly documented free-function names
}

func loadSurface(docgenPath string) (*publicSurface, error) {
	b, err := os.ReadFile(docgenPath)
	if err != nil {
		return nil, err
	}
	var m docgenManifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	s := &publicSurface{roots: map[string]bool{}, funcs: map[string]bool{}}
	for _, f := range m.Files {
		for _, r := range f.Roots {
			s.roots[r] = true
		}
		for _, fn := range f.Funcs {
			s.funcs[fn] = true
		}
	}
	return s, nil
}

// writeVerbs are the action-verb prefixes that mark a root-type method as a
// write/do capability (the audit surface). Methods without one of these are
// getters/readers/navigation — exempt from the must-tag requirement per the
// 2026-06-16 P2 policy (write-surface-first; getters may still be tagged
// domain=meta but are not CI-required). Adding a getter tag later is additive.
var writeVerbs = []string{
	"Set", "Add", "Remove", "Delete", "Insert", "Move", "Duplicate", "Animate",
	"Replace", "Clear", "Enable", "Disable", "Toggle", "Apply", "Reset", "Make",
}

func isWriteMethod(name string) bool {
	for _, v := range writeVerbs {
		if strings.HasPrefix(name, v) {
			return true
		}
	}
	return false
}

// requires reports whether entry e must carry an aep:cap tag.
func (s *publicSurface) requires(e Entry) bool {
	switch e.Kind {
	case "func":
		return e.Pkg == "aep" || s.funcs[e.Symbol]
	case "method":
		// root-type write/do methods only; getters/readers are exempt.
		return s.roots[strings.TrimPrefix(e.Recv, "*")] && isWriteMethod(e.Symbol)
	}
	return false // type + const = meta lane (Wave 9)
}

// getterExempt counts root-type getter/reader methods skipped by the
// write-surface policy — surfaced in the coverage report so the exemption is
// never a silent cap.
func (s *publicSurface) getterExempt(entries []Entry) int {
	n := 0
	for _, e := range entries {
		if e.Kind == "method" && s.roots[strings.TrimPrefix(e.Recv, "*")] && !isWriteMethod(e.Symbol) {
			n++
		}
	}
	return n
}

// coverage splits the public surface into tagged vs untagged for the P2 progress
// meter (not yet a CI gate — full-coverage enforcement flips on in Wave 9).
func (s *publicSurface) coverage(entries []Entry) (tagged, untagged []Entry) {
	for _, e := range entries {
		if !s.requires(e) {
			continue
		}
		if e.HasCap {
			tagged = append(tagged, e)
		} else {
			untagged = append(untagged, e)
		}
	}
	return
}
