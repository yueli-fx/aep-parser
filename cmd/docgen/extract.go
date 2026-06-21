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

	"github.com/example/aep-parser/internal/apidoc"
)

// loadedPackage 捆绑 go/doc 视图 + 共享 fset（签名打印 / 注释定位都要 fset）。
// RawFuncDocs：go/doc 处理前从原始 AST 抓取的函数注释（key = FuncDecl.Pos()），
// 因 doc.NewFromFiles 会清空 Decl.Doc，directive 扫描须用此副本。
type loadedPackage struct {
	Doc         *doc.Package
	Fset        *token.FileSet
	RawFuncDocs map[token.Pos]*ast.CommentGroup
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
		// 捕获原始函数注释（go/doc.NewFromFiles 会清空 Decl.Doc）。
		rawDocs := map[token.Pos]*ast.CommentGroup{}
		for _, f := range astPkg.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				if fd, ok := n.(*ast.FuncDecl); ok && fd.Doc != nil {
					rawDocs[fd.Pos()] = fd.Doc
				}
				return true
			})
		}
		dpkg, err := doc.NewFromFiles(fset, files, "github.com/example/aep-parser/"+dir, doc.AllDecls)
		if err != nil {
			return nil, fmt.Errorf("doc.NewFromFiles: %w", err)
		}
		return &loadedPackage{Doc: dpkg, Fset: fset, RawFuncDocs: rawDocs}, nil
	}
	return nil, fmt.Errorf("no buildable package in %s", dir)
}

// lookupPackageFunc 在单个包里按名找包级函数（含 go/doc 归到某 type.Funcs 的
// 构造器形函数），找到则用该包 fset + RawFuncDocs 构 symbol；否则 nil。
func lookupPackageFunc(lp *loadedPackage, name string) *symbol {
	for _, fn := range lp.Doc.Funcs {
		if fn.Name == name {
			return funcSymbol(lp, fn)
		}
	}
	// go/doc files a constructor-shaped func (NewT returning *T) under that
	// type's Funcs rather than the package's, so scan both.
	for _, ty := range lp.Doc.Types {
		for _, fn := range ty.Funcs {
			if fn.Name == name {
				return funcSymbol(lp, fn)
			}
		}
	}
	return nil
}

func funcSymbol(lp *loadedPackage, fn *doc.Func) *symbol {
	rawDoc := lp.RawFuncDocs[fn.Decl.Pos()] // doc.NewFromFiles 会清空 Decl.Doc，用副本
	s := &symbol{
		name:      fn.Name,
		kind:      kindMethod,
		signature: normalizeSignature(lp.Fset, fn.Decl),
	}
	applyAnnotation(s, rawDoc)
	return s
}

// applyAnnotation parses a raw doc-comment group for @tag fields. When present,
// it fills the symbol's summary/description/params/returns from the annotation;
// otherwise it falls back to the legacy directive-stripped prose.
func applyAnnotation(s *symbol, rawDoc *ast.CommentGroup) {
	ann, err := apidoc.Parse(rawCommentOf(rawDoc))
	if err != nil || !ann.HasTags {
		s.doc = directiveStrippedText(rawDoc)
		return
	}
	s.annotated = true
	s.summary = ann.Summary
	s.doc = ann.Description
	s.params = ann.Params
	s.returns = ann.Returns
}

// rawCommentOf reconstructs comment text WITHOUT go/doc directive stripping, so
// `// @tag` lines survive for apidoc.Parse (mirrors capindex's rawCommentText).
func rawCommentOf(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	var b strings.Builder
	for _, c := range cg.List {
		t := c.Text
		if strings.HasPrefix(t, "//") {
			t = t[2:]
		} else {
			t = strings.TrimSuffix(strings.TrimPrefix(t, "/*"), "*/")
		}
		b.WriteString(strings.TrimPrefix(t, " "))
		b.WriteByte('\n')
	}
	return b.String()
}

// extractPackageFuncs 抽取 names 指定的包级函数（非方法）为 symbol（按 names 顺序）。
// 未找到的名字报错（防 manifest 拼写）。
func extractPackageFuncs(lp *loadedPackage, names []string) ([]symbol, error) {
	out := make([]symbol, 0, len(names))
	for _, n := range names {
		s := lookupPackageFunc(lp, n)
		if s == nil {
			return nil, fmt.Errorf("package func %q not found", n)
		}
		out = append(out, *s)
	}
	return out, nil
}

// loadPackages 加载多个包目录的 go/doc 视图（顺序保留）。
func loadPackages(dirs []string) ([]*loadedPackage, error) {
	lps := make([]*loadedPackage, 0, len(dirs))
	for _, d := range dirs {
		lp, err := loadPackage(d)
		if err != nil {
			return nil, err
		}
		lps = append(lps, lp)
	}
	return lps, nil
}

// extractTypesMulti 跨多个包抽类型并合并：同名类型偏好真定义（非别名），
// 解决 facade 包里 `type X = scene.X` 别名壳遮蔽真类型字段/方法的问题
// （M8 物理分包后类型住 scene/codec、facade 只剩别名）。首见序稳定。
func extractTypesMulti(lps []*loadedPackage) []*docType {
	byName := map[string]*docType{}
	var order []string
	for _, lp := range lps {
		for _, dt := range extractTypes(lp) {
			existing, ok := byName[dt.name]
			if !ok {
				byName[dt.name] = dt
				order = append(order, dt.name)
				continue
			}
			if existing.isAlias && !dt.isAlias { // 真定义胜别名壳
				byName[dt.name] = dt
			}
		}
	}
	out := make([]*docType, 0, len(order))
	for _, n := range order {
		out = append(out, byName[n])
	}
	return out
}

// withMethodsMulti 把每个包的方法/getter 累加到已合并的 types（按包序应用，
// 真定义所在包通常在 facade 之后 → 其 setter 决定的字段 R/RW 最终生效）。
func withMethodsMulti(types []*docType, lps []*loadedPackage) []*docType {
	for _, lp := range lps {
		types = withMethods(types, lp)
	}
	return types
}

// extractPackageFuncsMulti 跨多个包按名找包级函数，按包序首个命中胜出
// （Pkg=facade 在前 → re-export 形函数优先）。任一名字全包未命中即报错。
func extractPackageFuncsMulti(lps []*loadedPackage, names []string) ([]symbol, error) {
	out := make([]symbol, 0, len(names))
	for _, n := range names {
		var s *symbol
		for _, lp := range lps {
			if s = lookupPackageFunc(lp, n); s != nil {
				break
			}
		}
		if s == nil {
			return nil, fmt.Errorf("package func %q not found in any scanned package", n)
		}
		out = append(out, *s)
	}
	return out, nil
}

// findType 在 []*docType 中按名查找（生产 + 测试共用）。
func findType(ts []*docType, name string) *docType {
	for _, t := range ts {
		if t.name == name {
			return t
		}
	}
	return nil
}

// extractTypes 把 loadedPackage 转成 []*docType（仅字段 + 常量；方法在 withMethods 补）。
func extractTypes(lp *loadedPackage) []*docType {
	var out []*docType
	for _, ty := range lp.Doc.Types {
		dt := &docType{name: ty.Name, doc: ty.Doc, isAlias: isAliasType(ty)}
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

// isAliasType reports whether ty is declared as a type alias (`type X = Y`)
// rather than a defining declaration (`type X Y` / `type X struct{…}`). In the
// aep facade package the public types are aliases to the real scene/codec
// types, so their go/doc view carries no fields/methods — extractTypesMulti
// prefers the real (non-alias) definition when merging across packages.
func isAliasType(ty *doc.Type) bool {
	for _, spec := range ty.Decl.Specs {
		if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.Name == ty.Name {
			return ts.Assign.IsValid()
		}
	}
	return false
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
			// Prefer the leading doc comment; fall back to the trailing
			// inline comment (`Field T // ...`), the common style for short
			// field docs.
			fdoc := f.Doc
			if fdoc == nil {
				fdoc = f.Comment
			}
			for _, nm := range f.Names {
				if !nm.IsExported() {
					continue
				}
				sym := symbol{
					name:      nm.Name,
					kind:      kindField,
					fieldDecl: nm.Name + " " + printNode(lp.Fset, f.Type),
					doc:       directiveStrippedText(fdoc),
					jsonName:  jsonTag(f.Tag),
				}
				switch directiveOf(fdoc) {
				case "rw":
					sym.fieldRWForced, sym.readWrite = true, true
				case "ro":
					sym.fieldRWForced, sym.readWrite = true, false
				}
				out = append(out, sym)
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

// withMethods 给已抽取的 types 填 methods/getters，并据 setter 存在性定字段 R/RW。
func withMethods(types []*docType, lp *loadedPackage) []*docType {
	byName := map[string]*docType{}
	for _, t := range types {
		byName[t.name] = t
	}
	for _, ty := range lp.Doc.Types {
		dt := byName[ty.Name]
		if dt == nil {
			continue
		}
		setterTargets := map[string]bool{} // "Name" ← SetName
		for _, fn := range ty.Methods {
			if t := strings.TrimPrefix(fn.Name, "Set"); t != fn.Name && t != "" {
				setterTargets[t] = true
			}
		}
		// 字段 R/RW：默认按 setter 存在；被 directive 锁定者跳过。
		for i := range dt.attributes {
			a := &dt.attributes[i]
			if a.kind != kindField || a.fieldRWForced {
				continue
			}
			a.readWrite = setterTargets[a.name]
		}
		// 方法分类。
		for _, fn := range ty.Methods {
			if !ast.IsExported(fn.Name) {
				continue
			}
			rawDoc := lp.RawFuncDocs[fn.Decl.Pos()]
			sym := symbol{
				name:      fn.Name,
				signature: normalizeSignature(lp.Fset, fn.Decl),
			}
			applyAnnotation(&sym, rawDoc)
			switch classifyMethod(fn, rawDoc) {
			case kindGetter:
				sym.kind = kindGetter
				sym.readWrite = false
				dt.attributes = append(dt.attributes, sym)
			default:
				sym.kind = kindMethod
				dt.methods = append(dt.methods, sym)
			}
		}
	}
	return types
}

// classifyMethod：directive 优先，否则启发式。rawDoc 是 go/doc 处理前的原始注释。
func classifyMethod(fn *doc.Func, rawDoc *ast.CommentGroup) symKind {
	switch directiveOf(rawDoc) {
	case "method":
		return kindMethod
	case "attribute":
		return kindGetter
	}
	if strings.HasPrefix(fn.Name, "Set") {
		return kindMethod
	}
	ft := fn.Decl.Type
	noParams := ft.Params == nil || len(ft.Params.List) == 0
	oneResult := ft.Results != nil && len(ft.Results.List) == 1
	if noParams && oneResult {
		return kindGetter
	}
	return kindMethod
}

// directiveOf 扫注释组找首个 //docgen:<x>，返回 <x>（无则 ""）。
func directiveOf(g *ast.CommentGroup) string {
	if g == nil {
		return ""
	}
	for _, c := range g.List {
		line := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
		if rest := strings.TrimPrefix(line, "docgen:"); rest != line {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

// normalizeSignature 去函数体 + 折叠空白为单行。
func normalizeSignature(fset *token.FileSet, decl *ast.FuncDecl) string {
	cp := *decl
	cp.Body = nil
	cp.Doc = nil
	s := printNode(fset, &cp)
	s = strings.Join(strings.Fields(s), " ")
	return s
}
