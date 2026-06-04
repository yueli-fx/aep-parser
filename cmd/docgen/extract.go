package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"reflect"
	"strings"
)

// loadedPackage 捆绑 go/doc 视图 + 共享 fset（签名打印 / 注释定位都要 fset）。
type loadedPackage struct {
	Doc  *doc.Package
	Fset *token.FileSet
	Pkg  *ast.Package
}

// loadPackage 解析 dir 下的 Go 包（含 _test.go，便于关联 Example），
// 返回 go/doc 视图。AllDecls 保证未导出符号也在 AST 里（按需过滤）。
func loadPackage(dir string) (*loadedPackage, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", dir, err)
	}
	for name, astPkg := range pkgs {
		// 跳过外部测试包（xxx_test），主包名优先。
		if len(name) > 5 && name[len(name)-5:] == "_test" {
			continue
		}
		files := make([]*ast.File, 0, len(astPkg.Files))
		for _, f := range astPkg.Files {
			files = append(files, f)
		}
		dpkg, err := doc.NewFromFiles(fset, files, "github.com/example/aep-parser/"+dir, doc.AllDecls)
		if err != nil {
			return nil, fmt.Errorf("doc.NewFromFiles: %w", err)
		}
		return &loadedPackage{Doc: dpkg, Fset: fset, Pkg: astPkg}, nil
	}
	return nil, fmt.Errorf("no buildable package in %s", dir)
}

// extractTypes 把 loadedPackage 转成 []*docType（仅字段 + 常量；方法在 Task 4 补）。
func extractTypes(lp *loadedPackage) []*docType {
	var out []*docType
	for _, ty := range lp.Doc.Types {
		dt := &docType{name: ty.Name, doc: ty.Doc}
		dt.attributes = append(dt.attributes, extractFields(lp, ty)...)
		for _, c := range ty.Consts {
			dt.consts = append(dt.consts, constBlock{
				doc:  c.Doc,
				code: printNode(lp.Fset, c.Decl),
			})
		}
		out = append(out, dt)
	}
	return out
}

// extractFields 走 type 的 GenDecl → StructType，取导出字段。
func extractFields(lp *loadedPackage, ty *doc.Type) []symbol {
	var out []symbol
	for _, spec := range ty.Decl.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok || st.Fields == nil {
			continue
		}
		for _, f := range st.Fields.List {
			for _, nm := range f.Names {
				if !nm.IsExported() {
					continue
				}
				out = append(out, symbol{
					name:      nm.Name,
					kind:      kindField,
					fieldDecl: nm.Name + " " + printNode(lp.Fset, f.Type),
					doc:       directiveStrippedText(f.Doc),
					jsonName:  jsonTag(f.Tag),
				})
			}
		}
	}
	return out
}

// printNode 用 go/printer 渲 AST 节点为源码字符串。
func printNode(fset *token.FileSet, node any) string {
	var b strings.Builder
	_ = printer.Fprint(&b, fset, node)
	return b.String()
}

// directiveStrippedText 返回注释 prose，剥离 //tool:directive 行（ast.Text() 已做）。
func directiveStrippedText(g *ast.CommentGroup) string {
	if g == nil {
		return ""
	}
	return strings.TrimSpace(g.Text())
}

// jsonTag 从 struct tag 取 json 名（去掉 ,omitempty 等）。
func jsonTag(tag *ast.BasicLit) string {
	if tag == nil {
		return ""
	}
	v := strings.Trim(tag.Value, "`")
	st := reflectStructTag(v)
	j := st["json"]
	if i := strings.Index(j, ","); i >= 0 {
		j = j[:i]
	}
	if j == "-" {
		return ""
	}
	return j
}

// reflectStructTag 用 reflect.StructTag 解析（避免手写 tag 解析）。
func reflectStructTag(raw string) map[string]string {
	st := reflect.StructTag(raw)
	out := map[string]string{}
	for _, k := range []string{"json"} {
		if v, ok := st.Lookup(k); ok {
			out[k] = v
		}
	}
	return out
}
