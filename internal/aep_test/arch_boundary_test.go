package aep_test

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// Package-level import-DAG guard for the M8 physical split. Pre-split this file
// enforced per-file naming-axis rules inside the single internal/aep package;
// post-split (internal/{rifx,codec,scene,serializer} + aep facade) the real
// boundaries are between packages, so the guard now AST-scans each package's
// production imports and asserts the layering:
//
//	rifx        (leaf)            imports none of ours
//	codec       (pure value/byte) -X-> scene, serializer, rifx, aep
//	scene       (runtime model)   -X-> rifx, serializer, aep   (may import codec)
//	serializer  (chunk codec)     -X-> aep                     (imports scene+codec+rifx)
//	aep         (facade)          imports scene+serializer+codec (top — no ban)
//
// scene's chunk-freeness (no rifx) is the load-bearing invariant from the V3
// arc; it is now a package boundary rather than a per-file whitelist.
const modBase = "github.com/yueli-fx/aep-parser/internal/"

// bannedImports maps each package (by dir, relative to this test's internal/aep
// working dir) to the internal packages it must not import.
var bannedImports = map[string][]string{
	"../codec":      {"scene", "serializer", "rifx", "aep"},
	"../scene":      {"rifx", "serializer", "aep"},
	"../serializer": {"aep"},
}

func TestArchBoundary_PackageDAG(t *testing.T) {
	fset := token.NewFileSet()
	for dir, banned := range bannedImports {
		files, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err != nil {
			t.Fatalf("glob %s: %v", dir, err)
		}
		bannedSet := make(map[string]bool, len(banned))
		for _, b := range banned {
			bannedSet[modBase+b] = true
		}
		scanned := 0
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			scanned++
			af, err := parser.ParseFile(fset, f, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("parse %s: %v", f, err)
			}
			for _, imp := range af.Imports {
				p := strings.Trim(imp.Path.Value, `"`)
				if bannedSet[p] {
					t.Errorf("%s imports %s — forbidden by package DAG (%s must not depend on it)",
						f, p, filepath.Base(dir))
				}
			}
		}
		if scanned == 0 {
			t.Errorf("package dir %s: no source files scanned — guard would be vacuous", dir)
		}
	}
}
