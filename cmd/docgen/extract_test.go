package main

import "testing"

func TestLoadPackage_FindsType(t *testing.T) {
	pkg, err := loadPackage("./testdata/sample")
	if err != nil {
		t.Fatalf("loadPackage: %v", err)
	}
	if pkg == nil || pkg.Doc == nil {
		t.Fatal("nil package/doc")
	}
	var found bool
	for _, ty := range pkg.Doc.Types {
		if ty.Name == "Widget" {
			found = true
		}
	}
	if !found {
		t.Fatal("type Widget not found")
	}
}
