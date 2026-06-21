package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

// convertSource rewrites every exported func whose doc comment carries an
// aep:cap directive into a @tag block with placeholders. It edits the raw bytes
// (not the AST) to preserve all unrelated formatting; edits are applied
// back-to-front so byte offsets stay valid. The output is intentionally lossy
// (everything human is TODO) — the converter removes boilerplate, the manual
// pass finishes the prose.
func convertSource(src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "src.go", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	type edit struct {
		start, end int
		text       string
	}
	var edits []edit
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || !fd.Name.IsExported() || fd.Doc == nil {
			continue
		}
		legacy := capLine(fd.Doc)
		if legacy == "" {
			continue
		}
		block := buildTagBlock(fd, legacy)
		start := fset.Position(fd.Doc.Pos()).Offset
		end := fset.Position(fd.Doc.End()).Offset
		edits = append(edits, edit{start, end, block})
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	out := append([]byte(nil), src...)
	for _, e := range edits {
		tail := append([]byte(e.text), out[e.end:]...)
		out = append(out[:e.start], tail...)
	}
	return out, nil
}

// capLine returns the aep:cap directive body (after "aep:cap"), or "".
func capLine(cg *ast.CommentGroup) string {
	for _, c := range cg.List {
		t := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
		if strings.HasPrefix(t, "aep:cap") {
			return strings.TrimSpace(strings.TrimPrefix(t, "aep:cap"))
		}
	}
	return ""
}

// buildTagBlock renders the new // @tag block. Machine fields come from the
// parsed aep:cap k/v; prose/param/returns are TODO placeholders for the manual
// pass.
func buildTagBlock(fd *ast.FuncDecl, legacy string) string {
	kv := kvParse(legacy)
	var b bytes.Buffer
	fmt.Fprintf(&b, "// @summary    TODO summary for %s\n", fd.Name.Name)
	for _, p := range paramNames(fd) {
		fmt.Fprintf(&b, "// @param   %s TODO describe %s\n", p, p)
	}
	if hasNonErrReturn(fd) {
		b.WriteString("// @returns TODO describe the return value\n")
	}
	emit := func(tag, key string) {
		if v := kv[key]; v != "" {
			fmt.Fprintf(&b, "// @%-10s %s\n", tag, v)
		}
	}
	emit("domain", "domain")
	if t := kv["tier"]; t == "stable" || t == "alpha" {
		fmt.Fprintf(&b, "// @%-10s %s\n", "stability", t)
	} else if t != "" {
		fmt.Fprintf(&b, "// @%-10s TODO map tier=%s\n", "stability", t)
	}
	emit("verify", "verify")
	emit("gate", "gate")
	minver := kv["minver"]
	if minver == "" {
		minver = "2020"
	}
	fmt.Fprintf(&b, "// @%-10s AE%s\n", "since", minver)
	if v := kv["boundary"]; v != "" {
		if isASCII(v) {
			fmt.Fprintf(&b, "// @%-10s %s\n", "boundary", v)
		} else {
			fmt.Fprintf(&b, "// @%-10s TODO translate: %s\n", "boundary", v)
		}
	}
	emit("incident", "incident")
	emit("alias", "alias")
	return strings.TrimRight(b.String(), "\n")
}

// kvParse parses space-separated key=value tokens; a value may be double-quoted
// to contain spaces (mirrors capindex's tokenizeKV, minus error reporting — the
// converter is best-effort over already-valid directives).
func kvParse(s string) map[string]string {
	out := map[string]string{}
	i, n := 0, len(s)
	for i < n {
		for i < n && s[i] == ' ' {
			i++
		}
		if i >= n {
			break
		}
		ks := i
		for i < n && s[i] != '=' && s[i] != ' ' {
			i++
		}
		if i >= n || s[i] != '=' {
			break
		}
		key := s[ks:i]
		i++ // skip '='
		var val string
		if i < n && s[i] == '"' {
			i++
			vs := i
			for i < n && s[i] != '"' {
				i++
			}
			val = s[vs:i]
			if i < n {
				i++ // skip closing quote
			}
		} else {
			vs := i
			for i < n && s[i] != ' ' {
				i++
			}
			val = s[vs:i]
		}
		out[key] = val
	}
	return out
}

func paramNames(fd *ast.FuncDecl) []string {
	var out []string
	if fd.Type.Params == nil {
		return out
	}
	for _, f := range fd.Type.Params.List {
		for _, nm := range f.Names {
			if nm.Name != "_" {
				out = append(out, nm.Name)
			}
		}
	}
	return out
}

func hasNonErrReturn(fd *ast.FuncDecl) bool {
	if fd.Type.Results == nil {
		return false
	}
	for _, f := range fd.Type.Results.List {
		if id, ok := f.Type.(*ast.Ident); ok && id.Name == "error" {
			continue
		}
		return true
	}
	return false
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}
