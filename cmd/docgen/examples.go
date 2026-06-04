package main

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"sort"
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
	var comments []*ast.CommentGroup
	for _, p := range pkgs {
		for _, f := range p.Files {
			files = append(files, f)
			comments = append(comments, f.Comments...)
		}
	}
	// printer.CommentedNode expects comments in ascending position order; map
	// iteration above mixes files, so sort the merged set.
	sort.Slice(comments, func(i, j int) bool { return comments[i].Pos() < comments[j].Pos() })
	for _, ex := range doc.Examples(files...) {
		typ, meth, suf := splitExampleName(ex.Name)
		dt := findType(types, typ)
		if dt == nil {
			continue
		}
		code := formatExampleBody(fset, ex, comments)
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

// formatExampleBody 渲 Example 函数体（不含大括号 / Output 注释）。comments 是
// 全部源文件的注释组（按位置排序）；printer 只打印落在 ex.Code 范围内的，从而
// 保留示例体里的注释（否则裸打印 AST 子树会丢注释、留下空行）。
func formatExampleBody(fset *token.FileSet, ex *doc.Example, comments []*ast.CommentGroup) string {
	var b strings.Builder
	_ = printer.Fprint(&b, fset, &printer.CommentedNode{Node: ex.Code, Comments: comments})
	s := b.String()
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	var lines []string
	for ln := range strings.SplitSeq(s, "\n") {
		// Stop at the testing "// Output:" marker — printing with comments
		// now includes it (it lives inside the example block), but it is not
		// part of the documented snippet.
		if isOutputMarker(ln) {
			break
		}
		lines = append(lines, strings.TrimPrefix(strings.TrimRight(ln, " \t"), "\t"))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// isOutputMarker reports whether a line is the go-test Output directive
// ("// Output:" / "// Unordered output:").
func isOutputMarker(line string) bool {
	low := strings.ToLower(strings.TrimSpace(line))
	low = strings.TrimPrefix(low, "//")
	low = strings.TrimSpace(low)
	return strings.HasPrefix(low, "output:") || strings.HasPrefix(low, "unordered output:")
}
