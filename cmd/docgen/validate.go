package main

import (
	"fmt"
	"go/ast"
	"go/doc"

	"github.com/example/aep-parser/internal/apidoc"
)

// validateAnnotations runs apidoc.Validate over every annotated package-level
// func + type method in the loaded packages, in warn mode with WriteSurface
// false: docgen checks that what it is about to RENDER is well-formed
// (@param matches the signature, @summary length/format, enum legality), while
// capindex --validate owns capability-completeness. Un-annotated symbols are
// skipped (legacy prose path).
func validateAnnotations(lps []*loadedPackage) []error {
	var errs []error
	ctx := apidoc.Context{} // no gate/incident crosscheck here (capindex owns it)
	check := func(lp *loadedPackage, fn *doc.Func) {
		raw := rawCommentOf(lp.RawFuncDocs[fn.Decl.Pos()])
		ann, err := apidoc.Parse(raw)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %v", fn.Name, err))
			return
		}
		if !ann.HasTags {
			return
		}
		sym := apidoc.Symbol{
			Name:            fn.Name,
			Pos:             posOf(lp, fn.Decl),
			Params:          declParamNames(fn.Decl),
			ReturnsNonError: declReturnsNonError(fn.Decl),
			WriteSurface:    false,
		}
		errs = append(errs, apidoc.Validate(ann, sym, ctx, apidoc.ModeWarn)...)
	}
	for _, lp := range lps {
		for _, fn := range lp.Doc.Funcs {
			check(lp, fn)
		}
		for _, ty := range lp.Doc.Types {
			for _, fn := range ty.Funcs {
				check(lp, fn)
			}
			for _, fn := range ty.Methods {
				check(lp, fn)
			}
		}
	}
	return errs
}

func posOf(lp *loadedPackage, d *ast.FuncDecl) string {
	p := lp.Fset.Position(d.Pos())
	return fmt.Sprintf("%s:%d", p.Filename, p.Line)
}

func declParamNames(d *ast.FuncDecl) []string {
	var out []string
	if d.Type.Params == nil {
		return out
	}
	for _, f := range d.Type.Params.List {
		for _, n := range f.Names {
			if n.Name != "_" {
				out = append(out, n.Name)
			}
		}
	}
	return out
}

func declReturnsNonError(d *ast.FuncDecl) bool {
	if d.Type.Results == nil {
		return false
	}
	for _, f := range d.Type.Results.List {
		if id, ok := f.Type.(*ast.Ident); ok && id.Name == "error" {
			continue
		}
		return true
	}
	return false
}
