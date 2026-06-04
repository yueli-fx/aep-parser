package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
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
