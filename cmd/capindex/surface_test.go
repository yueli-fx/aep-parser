package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSurface(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	s, err := loadSurface(filepath.Join(root, "docs", "docgen.json"))
	if err != nil {
		t.Fatal(err)
	}
	// docgen roots include the big capability types whose methods are public.
	for _, r := range []string{"Layer", "Composition", "Project", "Property", "Mask"} {
		if !s.roots[r] {
			t.Errorf("expected %q in docgen roots", r)
		}
	}
	// docgen funcs include curated free functions.
	for _, fn := range []string{"AddEffect", "InsertKeyframe", "NewShapeLayer"} {
		if !s.funcs[fn] {
			t.Errorf("expected %q in docgen funcs", fn)
		}
	}
}

// TestWriteSurfaceFullyTagged is the P2 CI drift guard: every public write/do
// symbol (facade func + write-verb method on an exported type) MUST carry an
// aep:cap tag. Getters/readers are exempt by policy. A new untagged write
// method fails this — keeping the capability index complete over time.
func TestWriteSurfaceFullyTagged(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	entries, err := extractEntries(capindexPkgDirs(root)...)
	if err != nil {
		t.Fatal(err)
	}
	surface, err := loadSurface(filepath.Join(root, "docs", "docgen.json"))
	if err != nil {
		t.Fatal(err)
	}
	var missing []string
	for _, e := range entries {
		if surface.requires(e) && !e.HasCap {
			sym := e.Symbol
			if e.Recv != "" {
				sym = strings.TrimPrefix(e.Recv, "*") + "." + e.Symbol
			}
			missing = append(missing, sym)
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d public write-surface symbol(s) missing an aep:cap tag (add a tag, or it's a getter that should be exempt): %v", len(missing), missing)
	}
}

func TestPublicSurfaceRequires(t *testing.T) {
	s := &publicSurface{
		roots: map[string]bool{"Layer": true},
		funcs: map[string]bool{"AddEffect": true},
	}
	cases := []struct {
		e    Entry
		want bool
	}{
		{Entry{Symbol: "SetOpacity", Kind: "method", Recv: "*Layer", Pkg: "scene"}, true},   // write method on a root type
		{Entry{Symbol: "SetColor", Kind: "method", Recv: "*FillNode", Pkg: "scene"}, true},  // write method on a non-root capability type
		{Entry{Symbol: "AddMask", Kind: "method", Recv: "*Layer", Pkg: "scene"}, true},      // Add* write method
		{Entry{Symbol: "Width", Kind: "method", Recv: "*Layer", Pkg: "scene"}, false},       // getter — exempt (write-surface-first)
		{Entry{Symbol: "HasAudio", Kind: "method", Recv: "*Layer", Pkg: "scene"}, false},    // Has* reader — exempt
		{Entry{Symbol: "PropertyByPath", Kind: "method", Recv: "*Layer", Pkg: "scene"}, false}, // navigation — exempt
		{Entry{Symbol: "NewShapeLayer", Kind: "func", Pkg: "aep"}, true},                    // facade func
		{Entry{Symbol: "AddEffect", Kind: "func", Pkg: "scene"}, true},                      // documented func anywhere
		{Entry{Symbol: "setOpaque", Kind: "method", Recv: "*shard", Pkg: "scene"}, false},   // non-root receiver
		{Entry{Symbol: "BlendingModeAdd", Kind: "const", Pkg: "aep"}, false},                // enum const = meta lane
		{Entry{Symbol: "Layer", Kind: "type", Pkg: "aep"}, false},                           // type alias = meta lane
	}
	for _, tc := range cases {
		if got := s.requires(tc.e); got != tc.want {
			t.Errorf("requires(%s.%s kind=%s) = %v, want %v", tc.e.Recv, tc.e.Symbol, tc.e.Kind, got, tc.want)
		}
	}
}

// TestExtractEntries_MultiPackage proves the P2 scope expansion: scene methods
// (the Set*/getter bulk) are now extracted, not just facade funcs.
func TestExtractEntries_MultiPackage(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	entries, err := extractEntries(capindexPkgDirs(root)...)
	if err != nil {
		t.Fatal(err)
	}
	var setOpacity *Entry
	for i := range entries {
		if entries[i].Symbol == "SetOpacity" && entries[i].Recv == "*Layer" {
			setOpacity = &entries[i]
			break
		}
	}
	if setOpacity == nil {
		t.Fatal("SetOpacity on *Layer not extracted from scene package")
	}
	if setOpacity.Kind != "method" || setOpacity.Pkg != "scene" {
		t.Errorf("SetOpacity: kind=%q pkg=%q, want method/scene", setOpacity.Kind, setOpacity.Pkg)
	}
}
