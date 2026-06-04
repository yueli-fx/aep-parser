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

func TestExtractType_FieldsAndConsts(t *testing.T) {
	lp, err := loadPackage("./testdata/sample")
	if err != nil {
		t.Fatal(err)
	}
	types := extractTypes(lp)
	w := findType(types, "Widget")
	if w == nil {
		t.Fatal("Widget missing")
	}
	if got := fieldNames(w); len(got) != 2 || got[0] != "Name" || got[1] != "Tags" {
		t.Fatalf("fields = %v, want [Name Tags]", got)
	}
	name := findSym(w.attributes, "Name")
	if name == nil || name.jsonName != "name" || name.fieldDecl != "Name string" {
		t.Fatalf("Name field wrong: %+v", name)
	}
	m := findType(types, "Mode")
	if m == nil || len(m.consts) != 1 {
		t.Fatalf("Mode consts = %v", m)
	}
}

// 测试辅助（后续 Task 也复用，定义在此一次）。
func findType(ts []*docType, name string) *docType {
	for _, t := range ts {
		if t.name == name {
			return t
		}
	}
	return nil
}
func findSym(ss []symbol, name string) *symbol {
	for i := range ss {
		if ss[i].name == name {
			return &ss[i]
		}
	}
	return nil
}
func fieldNames(t *docType) []string {
	var out []string
	for _, s := range t.attributes {
		if s.kind == kindField {
			out = append(out, s.name)
		}
	}
	return out
}
