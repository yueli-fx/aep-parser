package main

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
)

// attachExamples 解析 dir 下 *_test.go 的 Example 函数，按命名关联到 type.method。
func attachExamples(types []*docType, dir string) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return
	}
	var files []*ast.File
	for _, p := range pkgs {
		for _, f := range p.Files {
			files = append(files, f)
		}
	}
	for _, ex := range doc.Examples(files...) {
		typ, meth, suf := splitExampleName(ex.Name)
		dt := findType(types, typ)
		if dt == nil {
			continue
		}
		code := formatExampleBody(fset, ex)
		bindExample(dt, meth, example{suffix: suf, code: code})
	}
}

// splitExampleName: "Widget_SetName_merge" → (Widget, SetName, merge)。
// doc.Example.Name 已去掉前缀 "Example"。
func splitExampleName(name string) (typ, meth, suffix string) {
	parts := strings.SplitN(name, "_", 3)
	switch len(parts) {
	case 1:
		return parts[0], "", ""
	case 2:
		return parts[0], parts[1], ""
	default:
		return parts[0], parts[1], parts[2]
	}
}

func bindExample(dt *docType, meth string, ex example) {
	for _, bucket := range [][]symbol{dt.methods, dt.attributes} {
		for i := range bucket {
			if bucket[i].name == meth {
				bucket[i].examples = append(bucket[i].examples, ex)
				return
			}
		}
	}
}

// formatExampleBody 渲 Example 函数体（不含大括号 / Output 注释）。
func formatExampleBody(fset *token.FileSet, ex *doc.Example) string {
	var b strings.Builder
	_ = printer.Fprint(&b, fset, ex.Code)
	s := b.String()
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	var lines []string
	for _, ln := range strings.Split(s, "\n") {
		lines = append(lines, strings.TrimPrefix(strings.TrimRight(ln, " \t"), "\t"))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
