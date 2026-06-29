package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestManifestExampleDirsIncludesExtraPackages(t *testing.T) {
	m := &manifest{
		Pkg:         "./sample",
		Pkgs:        []string{"./extra"},
		ExamplePkgs: []string{"./examples", "./sample"},
		baseDir:     "testdata",
	}

	got := cleanPaths(m.exampleDirs())
	want := cleanPaths([]string{
		filepath.Join("testdata", "sample"),
		filepath.Join("testdata", "extra"),
		filepath.Join("testdata", "examples"),
	})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("example dirs = %v, want %v", got, want)
	}
}

func cleanPaths(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		out = append(out, filepath.Clean(p))
	}
	return out
}
