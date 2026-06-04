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

// 测试辅助（findType 在 extract.go，生产代码亦用）。
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

func TestExtractMethods_Classification(t *testing.T) {
	lp, _ := loadPackage("./testdata/sample")
	types := withMethods(extractTypes(lp), lp)
	w := findType(types, "Widget")

	// Name 字段因 SetName 存在 → RW。
	if n := findSym(w.attributes, "Name"); n == nil || !n.readWrite {
		t.Fatalf("Name should be RW: %+v", n)
	}
	// Tags 无 setter → R。
	if tg := findSym(w.attributes, "Tags"); tg == nil || tg.readWrite {
		t.Fatalf("Tags should be R: %+v", tg)
	}
	// Size 无参单返回非 Set → getter(Attributes,R)。
	if s := findSym(w.attributes, "Size"); s == nil || s.kind != kindGetter {
		t.Fatalf("Size should be getter: %+v", s)
	}
	// SetName → 动作方法，签名规范化单行。
	sn := findSym(w.methods, "SetName")
	if sn == nil || sn.signature != "func (w *Widget) SetName(name string) error" {
		t.Fatalf("SetName sig wrong: %+v", sn)
	}
	// Clone 本是 getter-like，但 //docgen:method 强制进 Methods。
	if c := findSym(w.methods, "Clone"); c == nil {
		t.Fatal("Clone should be forced into methods by directive")
	}
	if c := findSym(w.attributes, "Clone"); c != nil {
		t.Fatal("Clone must NOT be an attribute")
	}
}
