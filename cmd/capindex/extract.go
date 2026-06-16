package main

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"sort"
	"strings"
)

// extractEntries walks the Go package in dir (skipping _test.go) and returns one
// Entry per exported func/method/type/const, with the `aep:cap` directive parsed
// from each symbol's raw doc comment. The directive is written as a no-space
// `//aep:cap ...` line so go/doc strips it from rendered docs; capindex reads it
// back from the raw AST comment list.
func extractEntries(dir string) ([]Entry, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	for name, pkg := range pkgs {
		if strings.HasSuffix(name, "_test") {
			continue
		}
		for _, f := range pkg.Files {
			for _, decl := range f.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					if e, ok := funcEntry(fset, d); ok {
						entries = append(entries, e)
					}
				case *ast.GenDecl:
					entries = append(entries, genEntries(d)...)
				}
			}
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Symbol != entries[j].Symbol {
			return entries[i].Symbol < entries[j].Symbol
		}
		return entries[i].Recv < entries[j].Recv
	})
	return entries, nil
}

func funcEntry(fset *token.FileSet, d *ast.FuncDecl) (Entry, bool) {
	if !d.Name.IsExported() {
		return Entry{}, false
	}
	recv, kind := "", "func"
	if d.Recv != nil && len(d.Recv.List) > 0 {
		recv = recvTypeName(d.Recv.List[0].Type)
		if !ast.IsExported(strings.TrimPrefix(recv, "*")) {
			return Entry{}, false
		}
		kind = "method"
	}
	e := Entry{
		Symbol:    d.Name.Name,
		Kind:      kind,
		Recv:      recv,
		Signature: funcSignature(fset, d),
		Summary:   firstSentence(d.Doc),
	}
	attachCap(&e, d.Doc)
	return e, true
}

func genEntries(d *ast.GenDecl) []Entry {
	var out []Entry
	switch d.Tok {
	case token.TYPE:
		for _, spec := range d.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || !ts.Name.IsExported() {
				continue
			}
			doc := ts.Doc
			if doc == nil {
				doc = d.Doc // single-spec block keeps the doc on the GenDecl
			}
			e := Entry{Symbol: ts.Name.Name, Kind: "type", Summary: firstSentence(doc)}
			attachCap(&e, doc)
			out = append(out, e)
		}
	case token.CONST:
		for _, spec := range d.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			doc := vs.Doc
			if doc == nil {
				doc = d.Doc
			}
			for _, n := range vs.Names {
				if !n.IsExported() {
					continue
				}
				e := Entry{Symbol: n.Name, Kind: "const", Summary: firstSentence(doc)}
				attachCap(&e, doc)
				out = append(out, e)
			}
		}
	}
	return out
}

func attachCap(e *Entry, cg *ast.CommentGroup) {
	c, ok, err := parseCapTag(rawCommentText(cg))
	if err != nil {
		e.HasCap = true
		e.parseErr = err
		return
	}
	if ok {
		e.HasCap = true
		e.Cap = *c
	}
}

func recvTypeName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		return "*" + recvTypeName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return recvTypeName(t.X)
	case *ast.IndexListExpr:
		return recvTypeName(t.X)
	}
	return ""
}

func funcSignature(fset *token.FileSet, d *ast.FuncDecl) string {
	cp := *d
	cp.Doc = nil
	cp.Body = nil
	var b strings.Builder
	_ = printer.Fprint(&b, fset, &cp)
	return strings.TrimSpace(b.String())
}

// rawCommentText reconstructs the comment text WITHOUT go/doc's directive
// stripping, so the `//aep:cap` line survives for parseCapTag.
func rawCommentText(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	var b strings.Builder
	for _, c := range cg.List {
		t := c.Text
		if strings.HasPrefix(t, "//") {
			t = t[2:]
		} else {
			t = strings.TrimPrefix(t, "/*")
			t = strings.TrimSuffix(t, "*/")
		}
		b.WriteString(strings.TrimPrefix(t, " "))
		b.WriteByte('\n')
	}
	return b.String()
}

// firstSentence returns the first sentence of the cleaned doc text. cg.Text()
// already strips the `//aep:cap` directive, so the summary stays clean.
func firstSentence(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	txt := strings.TrimSpace(cg.Text())
	if i := strings.Index(txt, "\n\n"); i >= 0 {
		txt = txt[:i]
	}
	txt = strings.ReplaceAll(txt, "\n", " ")
	if i := strings.Index(txt, ". "); i >= 0 {
		txt = txt[:i+1]
	}
	return strings.TrimSpace(txt)
}
