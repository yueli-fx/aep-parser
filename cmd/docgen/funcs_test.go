package main

import (
	"strings"
	"testing"
)

func TestExtractPackageFuncs(t *testing.T) {
	lp, err := loadPackage("./testdata/sample")
	if err != nil {
		t.Fatal(err)
	}
	funcs, err := extractPackageFuncs(lp, []string{"NewWidget"})
	if err != nil {
		t.Fatal(err)
	}
	if len(funcs) != 1 {
		t.Fatalf("want 1 func, got %d", len(funcs))
	}
	f := funcs[0]
	if f.name != "NewWidget" {
		t.Fatalf("name = %q", f.name)
	}
	if f.signature != "func NewWidget(name string) *Widget" {
		t.Fatalf("signature = %q", f.signature)
	}
	if !strings.Contains(f.doc, "构造一个 Widget") {
		t.Fatalf("doc = %q", f.doc)
	}
}

func TestExtractPackageFuncs_MissingErrors(t *testing.T) {
	lp, _ := loadPackage("./testdata/sample")
	if _, err := extractPackageFuncs(lp, []string{"NoSuchFunc"}); err == nil {
		t.Fatal("expected error for unknown func name")
	}
}

func TestRenderFuncs(t *testing.T) {
	lp, _ := loadPackage("./testdata/sample")
	funcs, _ := extractPackageFuncs(lp, []string{"NewWidget"})
	out := renderFuncs(funcs)
	if !strings.Contains(out, "## Functions") {
		t.Fatalf("missing section header: %q", out)
	}
	if !strings.Contains(out, "### NewWidget") {
		t.Fatalf("missing func heading: %q", out)
	}
	if !strings.Contains(out, "func NewWidget(name string) *Widget") {
		t.Fatalf("missing signature: %q", out)
	}
}
