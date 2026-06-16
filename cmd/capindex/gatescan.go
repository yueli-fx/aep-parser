package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
)

// scanTests walks the given roots for *_test.go files and returns a map of
// top-level `func Test*` name → whether it is UNCONDITIONALLY skipped.
//
// Ship-gate tests legitimately call t.Skip when AE_SHIP_GATE is unset (a
// conditional skip inside an `if`), so only a skip that is a direct top-level
// statement of the function body is treated as "disabled".
func scanTests(roots ...string) (map[string]bool, error) {
	out := map[string]bool{}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, de fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if de.IsDir() || !strings.HasSuffix(path, "_test.go") {
				return nil
			}
			fset := token.NewFileSet()
			f, perr := parser.ParseFile(fset, path, nil, 0)
			if perr != nil {
				return perr
			}
			for _, decl := range f.Decls {
				fd, ok := decl.(*ast.FuncDecl)
				if !ok || fd.Recv != nil {
					continue
				}
				if !strings.HasPrefix(fd.Name.Name, "Test") {
					continue
				}
				out[fd.Name.Name] = hasUnconditionalSkip(fd)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func hasUnconditionalSkip(fd *ast.FuncDecl) bool {
	if fd.Body == nil {
		return false
	}
	for _, stmt := range fd.Body.List {
		es, ok := stmt.(*ast.ExprStmt)
		if !ok {
			continue
		}
		call, ok := es.X.(*ast.CallExpr)
		if !ok {
			continue
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		switch sel.Sel.Name {
		case "Skip", "Skipf", "SkipNow":
			return true
		}
	}
	return false
}
