package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
)

// docIndex is the machine-readable companion to the markdown docs: a flat map
// keyed by qualified symbol name ("Property.SetDimensionsSeparated") so a tool
// (or an agent) can KV-locate a symbol's doc file + anchor + signature + one-
// line summary without scanning the generated markdown. json.Marshal sorts map
// keys, so output is deterministic (drift-gate friendly).
type docIndex struct {
	GoVersion string                `json:"go_version"`
	Symbols   map[string]indexEntry `json:"symbols"`
}

type indexEntry struct {
	Kind      string `json:"kind"`                // type | field | getter | method
	File      string `json:"file"`                // output markdown filename
	Anchor    string `json:"anchor"`              // GitHub-style anchor within File
	Signature string `json:"signature,omitempty"` // method/getter signature or field decl
	RW        string `json:"rw,omitempty"`        // read-only | read-write (attributes only)
	JSON      string `json:"json,omitempty"`      // struct tag json name (fields only)
	Example   bool   `json:"example,omitempty"`   // a doc Example is attached
	Summary   string `json:"summary,omitempty"`   // first paragraph of prose, collapsed
}

// buildIndex walks the same files/roots the markdown generator does and emits
// the JSON index. It loads each file's package independently (cheap, mirrors
// generateFile) so the index always reflects the exact symbols rendered.
func buildIndex(m *manifest) (string, error) {
	idx := docIndex{GoVersion: runtime.Version(), Symbols: map[string]indexEntry{}}
	dirs := m.pkgDirs()
	lps, err := loadPackages(dirs)
	if err != nil {
		return "", err
	}
	types := withMethodsMulti(extractTypesMulti(lps), lps)
	for _, d := range dirs {
		attachExamples(types, d)
	}
	for _, fm := range m.Files {
		for _, root := range fm.Roots {
			dt := findType(types, root)
			if dt == nil {
				return "", fmt.Errorf("root type %q not found in %v", root, dirs)
			}
			idx.Symbols[dt.name] = indexEntry{
				Kind:    "type",
				File:    fm.Out,
				Anchor:  "#" + anchorFor(dt.name+" object"),
				Summary: summaryOf(dt.doc),
			}
			for _, s := range dt.attributes {
				idx.Symbols[dt.name+"."+s.name] = symbolEntry(fm.Out, dt.name, s)
			}
			for _, s := range dt.methods {
				idx.Symbols[dt.name+"."+s.name] = symbolEntry(fm.Out, dt.name, s)
			}
		}
		if len(fm.Funcs) > 0 {
			funcs, err := extractPackageFuncsMulti(lps, fm.Funcs)
			if err != nil {
				return "", err
			}
			for _, f := range funcs {
				idx.Symbols[f.name] = indexEntry{
					Kind:      "func",
					File:      fm.Out,
					Anchor:    "#" + anchorFor(f.name),
					Signature: f.signature,
					Summary:   summaryOf(f.doc),
				}
			}
		}
	}
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

func symbolEntry(file, typeName string, s symbol) indexEntry {
	e := indexEntry{
		File:    file,
		Anchor:  "#" + anchorFor(typeName+"."+s.name),
		Summary: summaryOf(s.doc),
		Example: len(s.examples) > 0,
	}
	switch s.kind {
	case kindField:
		e.Kind = "field"
		e.Signature = s.fieldDecl
		e.JSON = s.jsonName
		e.RW = rwLabel(s.readWrite)
	case kindGetter:
		e.Kind = "getter"
		e.Signature = s.signature
		e.RW = rwLabel(s.readWrite)
	default:
		e.Kind = "method"
		e.Signature = s.signature
	}
	return e
}

func rwLabel(rw bool) string {
	if rw {
		return "read-write"
	}
	return "read-only"
}

// anchorFor reproduces GitHub's heading-anchor slug: lowercase, drop
// punctuation other than spaces/hyphens/underscores, spaces→hyphens. So
// "Property.SetExpression" → "propertysetexpression", "Property object" →
// "property-object" — matching the anchors renderType/renderSymbol headings get.
func anchorFor(heading string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(heading) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
		case r == ' ', r == '-':
			b.WriteByte('-')
		}
	}
	return b.String()
}

// summaryOf collapses a symbol's first prose paragraph into a single line.
func summaryOf(doc string) string {
	doc = strings.TrimSpace(doc)
	if doc == "" {
		return ""
	}
	if i := strings.Index(doc, "\n\n"); i >= 0 {
		doc = doc[:i]
	}
	return strings.Join(strings.Fields(doc), " ")
}
