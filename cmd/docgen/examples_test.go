package main

import (
	"strings"
	"testing"
)

func TestAttachExamples_BindsToMethod(t *testing.T) {
	lp, _ := loadPackage("./testdata/sample")
	types := withMethods(extractTypes(lp), lp)
	attachExamples(types, "./testdata/sample")

	w := findType(types, "Widget")
	sn := findSym(w.methods, "SetName")
	if sn == nil || len(sn.examples) != 1 {
		t.Fatalf("SetName should have 1 example, got %+v", sn)
	}
	if !strings.Contains(sn.examples[0].code, "w.SetName") {
		t.Fatalf("example code wrong: %q", sn.examples[0].code)
	}
}
