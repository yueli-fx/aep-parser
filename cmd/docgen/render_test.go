package main

import (
	"os"
	"testing"
)

func TestRenderType_Widget_Golden(t *testing.T) {
	lp, _ := loadPackage("./testdata/sample")
	types := withMethods(extractTypes(lp), lp)
	attachExamples(types, "./testdata/sample")
	w := findType(types, "Widget")

	got := renderType(w)
	want, err := os.ReadFile("./testdata/golden/widget.md")
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("render mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
