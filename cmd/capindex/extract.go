package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/apidoc"
)

// extractEntries walks the given Go package dirs (skipping _test.go) and returns
// one Entry per exported func/method/type/const, with the `aep:cap` directive
// parsed from each symbol's raw doc comment. The directive is written as a
// no-space `//aep:cap ...` line so go/doc strips it from rendered docs; capindex
// reads it back from the raw AST comment list.
//
// Dirs are scanned in order with facade-priority dedup: the public capability
// surface spans internal/aep (facade re-export funcs + type aliases) and
// internal/scene (the real types whose Set*/getter methods are the bulk of the
// API). When the same symbol appears in two packages (a facade alias type and
// its scene definition, or a facade wrapper func over a codec original), the
// first-scanned (facade) entry wins so the tag lives on the public surface.
func extractEntries(dirs ...string) ([]Entry, error) {
	var entries []Entry
	seen := map[string]bool{}
	for _, dir := range dirs {
		pkg := filepath.Base(dir)
		got, err := extractDir(dir)
		if err != nil {
			return nil, err
		}
		for _, e := range got {
			key := e.Recv + "\x00" + e.Symbol + "\x00" + e.Kind
			if seen[key] {
				continue // facade-priority: keep the first (earlier dir) definition
			}
			seen[key] = true
			e.Pkg = pkg
			entries = append(entries, e)
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

// extractDir parses one package dir and returns its exported entries (unsorted,
// Pkg unset — extractEntries stamps Pkg + dedups).
func extractDir(dir string) ([]Entry, error) {
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
	pos := fset.Position(d.Pos())
	e.Pos = fmt.Sprintf("%s:%d", filepath.Base(pos.Filename), pos.Line)
	e.Params = extractParamNames(d)
	e.ReturnsNonError = returnsNonError(d)
	attachCap(&e, d.Doc)
	return e, true
}

// extractParamNames returns the non-receiver parameter names of d, in order
// (unnamed params are skipped — they cannot carry an @param).
func extractParamNames(d *ast.FuncDecl) []string {
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

// returnsNonError reports whether d returns at least one result whose type is
// not the builtin error.
func returnsNonError(d *ast.FuncDecl) bool {
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
	raw := rawCommentText(cg)
	ann, err := apidoc.Parse(raw)
	if err != nil {
		e.HasCap, e.parseErr = true, err
		return
	}
	if ann.HasTags {
		if apidoc.HasLegacyCap(raw) {
			e.HasCap = true
			e.parseErr = fmt.Errorf("carries BOTH @tag and legacy aep:cap (single-format invariant)")
			return
		}
		e.HasCap, e.Ann, e.Cap = true, ann, capFromAnnotation(ann)
		return
	}
	c, ok, err := parseCapTag(raw)
	if err != nil {
		e.HasCap, e.parseErr = true, err
		return
	}
	if ok {
		e.HasCap, e.Cap = true, *c
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
