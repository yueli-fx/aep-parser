package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/example/aep-parser/internal/apidoc"
)

func TestRunValidate_ConvertedSymbolStrict(t *testing.T) {
	surface := &publicSurface{funcs: map[string]bool{"AddThing": true}}
	good := Entry{
		Symbol: "AddThing", Kind: "func", Pkg: "aep", Pos: "f.go:1",
		Params: []string{"x"}, ReturnsNonError: true,
		Ann: &apidoc.Annotation{
			Summary: "Add a thing to the project", Domain: "structural", Stability: "stable",
			Verify: "roundtrip", Since: "AE2020", Returns: "the created thing",
			Params: []apidoc.Param{{Name: "x", Desc: "the thing to add"}}, HasTags: true,
		},
	}
	ctx := apidoc.Context{Gates: map[string]bool{}}
	if errs := runValidate([]Entry{good}, surface, ctx, apidoc.ModeWarn); len(errs) != 0 {
		t.Fatalf("clean converted symbol, got: %v", errs)
	}

	bad := good
	badAnn := *good.Ann
	badAnn.Summary = "Add a thing." // trailing period
	bad.Ann = &badAnn
	errs := runValidate([]Entry{bad}, surface, ctx, apidoc.ModeWarn)
	if len(errs) == 0 || !strings.Contains(errs[0].Error(), "period") {
		t.Fatalf("want period violation even in warn mode (symbol is converted), got: %v", errs)
	}
}

func TestRunValidate_UntaggedWarnVsStrict(t *testing.T) {
	surface := &publicSurface{funcs: map[string]bool{"AddThing": true}}
	untagged := Entry{Symbol: "AddThing", Kind: "func", Pkg: "aep", Pos: "f.go:1"}
	ctx := apidoc.Context{Gates: map[string]bool{}}
	if errs := runValidate([]Entry{untagged}, surface, ctx, apidoc.ModeWarn); len(errs) != 0 {
		t.Errorf("warn mode must ignore untagged surface symbol, got: %v", errs)
	}
	if errs := runValidate([]Entry{untagged}, surface, ctx, apidoc.ModeStrict); len(errs) == 0 {
		t.Error("strict mode must flag untagged surface symbol")
	}
}

func TestValidate_RealTree_WarnClean(t *testing.T) {
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
	gates, err := scanTests(filepath.Join(root, "internal"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := apidoc.Context{Gates: gates, IncidentsDir: filepath.Join(root, "flightdeck", "incidents")}
	if errs := runValidate(entries, surface, ctx, apidoc.ModeWarn); len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("validate: %v", e)
		}
	}
}
