---
status: active
summary: docgen 生成器 pilot 实现计划 — cmd/docgen（go/doc+go/ast 抽取 → markdown），TDD against testdata/sample，终点产出 docs/property.gen.md + 验证清单
---

## 执行进度（2026-06-04）

**Task 1-8 全部完成**（生成器 = 可用软件，7 测试全绿、vet/build 干净、opus 最终 review 判 sound）。commit 链 `b0de225`(T1)→`4fc8521`(T2)→`04f2ef2`(T3)→`bd05afd`(T4)→`879bf0d`(T5)→`acec054`(findType 修)→`6886622`(T6)→`ea6491d`(T7)→`3320d1e`(T8)→`13d1aee`(review 清理)。

**执行中抓到并修掉的真实 plan bug：**
- T4：`doc.NewFromFiles` 就地清空 `FuncDecl.Doc` → 加 `RawFuncDocs`（doc 处理前预捕获）扫 directive；字段 `f.Doc` 存活无需此处理（实测确认）。
- T5：`findType` 被生产代码（attachExamples/generateFile）用，但原计划放在 `extract_test.go` → 搬到 `extract.go`（非测试构建才能 build）。

**最终 review 的 defer 项（pilot 已知、非阻塞）：**
- type-level Example（`ExampleWidget` 无方法名）当前被丢弃——需要时补。
- 小写 suffix 约定（`ExampleX_setName` vs `_SetName`）未严格区分——无害（未导出方法本就不文档化）。
- `attachExamples` 静默吞解析错误——可加 stderr warning。
- manifest 多文件时 package 被 2N 次重解析——效率项，真实 manifest 多文件时把 loadPackage 提到循环外。

**Task 9 = 人工决策门（未启动）。** 只读 demo 已跑：生成器在真实 `internal/aep` 上产出 461 行 `property.md`（结构/R·RW 全对，见 `tmp/property_probe.gen.md`）。**关键发现**：真实 `.go` 已有部分**英文** doc comment，而手写 `docs/property.md` 是**中文**——「prose 用英文还是把中文搬进注释」是 Task 9 启动前需用户拍板的核心决策（连带 CLAUDE.md「不写注释」铁律调和）。

# docgen Pilot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建 `cmd/docgen` 生成器——从 `internal/aep` 的 Go doc comment + AST 自动生成 `docs/*.md`，pilot 跑通 property 域并产出 `docs/property.gen.md` 供并列验证。

**Architecture:** 标准库 `go/doc`+`go/ast`+`go/doc/comment`+`go/printer` 自研。extractor 把一个包解析成中间 `docType` 模型（字段/getter/方法/常量 + JSON tag + R·RW + Example），renderer 把模型按稳定序渲成 markdown（go/doc/comment→markdown 渲 prose，go/printer 渲规范化单行签名）。全程 TDD against `cmd/docgen/testdata/sample` 固定小包 + golden 文件。

**Tech Stack:** Go 1.25.1，纯 stdlib（零新依赖）。manifest 用 JSON（`encoding/json`）。

**Scope（本 plan 边界）：** 只做**生成器 + 单 type 跑通 + property.gen.md 产出 + 验证清单**。完整 property prose 迁移（742 行搬进注释）、其余 12 文件推广、CI 接线 = 验证通过后的 follow-on，不在本 plan。

**约定：** 命令在仓库根 `E:\projects\tools\aep-parser` 跑。所有 `cmd/docgen` 包文件 `package main`；测试 `package main`（白盒）。

---

### Task 1: testdata 固定输入包（fixture）

后续所有抽取/渲染测试都打这个包，先建好且锁定不变。

**Files:**
- Create: `cmd/docgen/testdata/sample/sample.go`
- Create: `cmd/docgen/testdata/sample/example_test.go`

- [ ] **Step 1: 写 sample.go（被解析的源）**

```go
// Package sample 是 docgen 的测试固定输入，覆盖 字段/getter/方法/常量/JSON tag/directive/Example。
package sample

import "fmt"

// Widget 是一个示例 widget。
//
// 第二段：演示 docgen 抽取多段 prose。
type Widget struct {
	// Name 是显示名。
	Name string `json:"name"`

	// Tags 是标签。
	Tags []string `json:"tags"`

	hidden int // 未导出，应被忽略
}

// SetName 设置 Name。
//
// 当 name 为空时返回 error。
func (w *Widget) SetName(name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}
	w.Name = name
	return nil
}

// Size 返回标签数。
func (w *Widget) Size() int { return len(w.Tags) }

// Clone 返回副本。
//
//docgen:method
func (w *Widget) Clone() *Widget { return &Widget{Name: w.Name} }

// Mode 是 widget 模式。
type Mode int

// Widget 模式常量。
const (
	ModeOff Mode = iota
	ModeOn
)
```

- [ ] **Step 2: 写 example_test.go（Example 函数）**

```go
package sample_test

import "github.com/example/aep-parser/cmd/docgen/testdata/sample"

func ExampleWidget_SetName() {
	w := &sample.Widget{}
	_ = w.SetName("hi")
	// Output:
}
```

- [ ] **Step 3: 验证固定包能编译**

Run: `go build ./cmd/docgen/testdata/sample/`
Expected: 无输出，exit 0。

- [ ] **Step 4: Commit**

```bash
git add cmd/docgen/testdata/sample/
git commit -m "test(docgen): add sample fixture package for extractor/renderer"
```

---

### Task 2: 包加载 loadPackage

**Files:**
- Create: `cmd/docgen/extract.go`
- Test: `cmd/docgen/extract_test.go`

- [ ] **Step 1: 写失败测试**

```go
package main

import "testing"

func TestLoadPackage_FindsType(t *testing.T) {
	pkg, err := loadPackage("./testdata/sample")
	if err != nil {
		t.Fatalf("loadPackage: %v", err)
	}
	if pkg == nil || pkg.Doc == nil {
		t.Fatal("nil package/doc")
	}
	var found bool
	for _, ty := range pkg.Doc.Types {
		if ty.Name == "Widget" {
			found = true
		}
	}
	if !found {
		t.Fatal("type Widget not found")
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./cmd/docgen/ -run TestLoadPackage_FindsType -v`
Expected: FAIL（`loadPackage` 未定义，编译错）。

- [ ] **Step 3: 写最小实现**

```go
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
```

> 注：`doc.NewFromFiles` 不含 `_test.go` 里的 Example —— Example 在 Task 5 单独从测试文件解析。此处 import path 串仅作 go/doc 内部用，不影响抽取。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./cmd/docgen/ -run TestLoadPackage_FindsType -v`
Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add cmd/docgen/extract.go cmd/docgen/extract_test.go
git commit -m "feat(docgen): loadPackage via go/doc"
```

---

### Task 3: 模型 + 字段/常量抽取

**Files:**
- Create: `cmd/docgen/model.go`
- Modify: `cmd/docgen/extract.go`
- Test: `cmd/docgen/extract_test.go`

- [ ] **Step 1: 写模型类型（model.go）**

```go
package main

type symKind int

const (
	kindField  symKind = iota // struct 字段 → Attributes
	kindGetter                // getter 方法 → Attributes
	kindMethod                // 动作方法 → Methods
)

type example struct {
	suffix string // "" 或 "merge"/"separate"
	code   string // 去掉 // Output 后的函数体源码
}

type symbol struct {
	name      string
	kind      symKind
	signature string // 规范化单行（方法/getter）；字段为 ""
	fieldDecl string // "Name string"（字段）；其它为 ""
	doc       string // directive 已剥离的 prose
	jsonName  string // struct tag json 名；无则 ""
	readWrite bool   // Attributes 专用：RW=true / R=false
	examples  []example
}

type constBlock struct {
	doc  string
	code string // const (...) 块源码（go/printer 渲）
}

type docType struct {
	name       string
	doc        string
	attributes []symbol // 字段（结构声明序）后接 getter（源序）
	methods    []symbol // 动作方法（源序）
	consts     []constBlock
}
```

- [ ] **Step 2: 写失败测试（字段 + 常量）**

```go
func TestExtractType_FieldsAndConsts(t *testing.T) {
	lp, err := loadPackage("./testdata/sample")
	if err != nil {
		t.Fatal(err)
	}
	types := extractTypes(lp)
	w := findType(types, "Widget")
	if w == nil {
		t.Fatal("Widget missing")
	}
	// 两个导出字段，hidden 被忽略。
	if got := fieldNames(w); len(got) != 2 || got[0] != "Name" || got[1] != "Tags" {
		t.Fatalf("fields = %v, want [Name Tags]", got)
	}
	name := findSym(w.attributes, "Name")
	if name == nil || name.jsonName != "name" || name.fieldDecl != "Name string" {
		t.Fatalf("Name field wrong: %+v", name)
	}
	// 常量挂在 Mode 上。
	m := findType(types, "Mode")
	if m == nil || len(m.consts) != 1 {
		t.Fatalf("Mode consts = %v", m)
	}
}

// 测试辅助。
func findType(ts []*docType, name string) *docType {
	for _, t := range ts {
		if t.name == name {
			return t
		}
	}
	return nil
}
func findSym(ss []symbol, name string) *symbol {
	for i := range ss {
		if ss[i].name == name {
			return &ss[i]
		}
	}
	return nil
}
func fieldNames(t *docType) []string {
	var out []string
	for _, s := range t.attributes {
		if s.kind == kindField {
			out = append(out, s.name)
		}
	}
	return out
}
```

- [ ] **Step 3: 跑测试确认失败**

Run: `go test ./cmd/docgen/ -run TestExtractType_FieldsAndConsts -v`
Expected: FAIL（`extractTypes` 未定义）。

- [ ] **Step 4: 实现 extractTypes（字段 + 常量，方法留 Task 4）**

在 `extract.go` 追加：

```go
import (
	"go/printer"
	"strings"
	// 保留 Task 2 已有的 import
)

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
```

在 `extract.go` import 块加 `"reflect"`。

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./cmd/docgen/ -run TestExtractType_FieldsAndConsts -v`
Expected: PASS。

- [ ] **Step 6: Commit**

```bash
git add cmd/docgen/model.go cmd/docgen/extract.go cmd/docgen/extract_test.go
git commit -m "feat(docgen): extract types, fields, consts"
```

---

### Task 4: 方法抽取 — setter/getter/directive + R·RW + 规范化签名

**Files:**
- Modify: `cmd/docgen/extract.go`
- Test: `cmd/docgen/extract_test.go`

分类规则（spec §R/RW）：
- 字段 `Foo` → RW 当存在方法 `SetFoo`，否则 R。
- 方法 `Set*` → 动作方法（Methods）。
- 其它方法：默认 getter 启发式 = 无参 + 单返回值 + 非 Set* → Attributes(R)；否则 Methods。
- **显式覆盖优先**：注释含 `//docgen:method` → 强制 Methods；`//docgen:attribute` → 强制 Attributes(R)；`//docgen:rw`/`//docgen:ro` 覆盖字段 R/RW。

- [ ] **Step 1: 写失败测试**

```go
func TestExtractMethods_Classification(t *testing.T) {
	lp, _ := loadPackage("./testdata/sample")
	types := withMethods(extractTypes(lp), lp)
	w := findType(types, "Widget")

	// Name 字段因 SetName 存在 → RW。
	if n := findSym(w.attributes, "Name"); n == nil || !n.readWrite {
		t.Fatalf("Name should be RW: %+v", n)
	}
	// Tags 无 setter → R。
	if tg := findSym(w.attributes, "Tags"); tg == nil || tg.readWrite {
		t.Fatalf("Tags should be R: %+v", tg)
	}
	// Size 无参单返回非 Set → getter(Attributes,R)。
	if s := findSym(w.attributes, "Size"); s == nil || s.kind != kindGetter {
		t.Fatalf("Size should be getter: %+v", s)
	}
	// SetName → 动作方法，签名规范化单行。
	sn := findSym(w.methods, "SetName")
	if sn == nil || sn.signature != "func (w *Widget) SetName(name string) error" {
		t.Fatalf("SetName sig wrong: %+v", sn)
	}
	// Clone 本是 getter-like，但 //docgen:method 强制进 Methods。
	if c := findSym(w.methods, "Clone"); c == nil {
		t.Fatal("Clone should be forced into methods by directive")
	}
	if c := findSym(w.attributes, "Clone"); c != nil {
		t.Fatal("Clone must NOT be an attribute")
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./cmd/docgen/ -run TestExtractMethods_Classification -v`
Expected: FAIL（`withMethods` 未定义）。

- [ ] **Step 3: 实现 withMethods**

在 `extract.go` 追加：

```go
import "strings" // 已有

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
		// 字段 R/RW：默认按 setter 存在，directive 覆盖。
		for i := range dt.attributes {
			a := &dt.attributes[i]
			if a.kind != kindField {
				continue
			}
			a.readWrite = setterTargets[a.name]
			applyFieldDirective(a)
		}
		// 方法分类。
		for _, fn := range ty.Methods {
			if !ast.IsExported(fn.Name) {
				continue
			}
			sym := symbol{
				name:      fn.Name,
				doc:       directiveStrippedText(fn.Decl.Doc),
				signature: normalizeSignature(lp.Fset, fn.Decl),
			}
			switch classifyMethod(fn) {
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

// classifyMethod：directive 优先，否则启发式。
func classifyMethod(fn *doc.Func) symKind {
	switch directiveOf(fn.Decl.Doc) {
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

// applyFieldDirective：字段上的 //docgen:rw / :ro 覆盖。
func applyFieldDirective(a *symbol) {
	// 字段 directive 已在 extractFields 时丢失（ast.Text 剥离）→ 这里从 doc 文本不可得；
	// 改由 directiveOf 在抽取期读取并存。见 Step 4 的字段 directive 透传。
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
	// 折叠内部换行/制表为单空格。
	s = strings.Join(strings.Fields(s), " ")
	return s
}
```

- [ ] **Step 4: 字段 directive 透传修正**

`extractFields` 内每个字段补存 directive，改 `applyFieldDirective` 真正生效。把 `extractFields` 里 `out = append(out, symbol{...})` 改为先建 `sym` 再按字段注释 directive 设标记：

```go
sym := symbol{
	name:      nm.Name,
	kind:      kindField,
	fieldDecl: nm.Name + " " + printNode(lp.Fset, f.Type),
	doc:       directiveStrippedText(f.Doc),
	jsonName:  jsonTag(f.Tag),
}
switch directiveOf(f.Doc) {
case "rw":
	sym.fieldRWForced, sym.readWrite = true, true
case "ro":
	sym.fieldRWForced, sym.readWrite = true, false
}
out = append(out, sym)
```

在 `model.go` 的 `symbol` 加字段 `fieldRWForced bool`。并把 `withMethods` 里字段 R/RW 那段改为尊重强制：

```go
for i := range dt.attributes {
	a := &dt.attributes[i]
	if a.kind != kindField || a.fieldRWForced {
		continue
	}
	a.readWrite = setterTargets[a.name]
}
```

删掉空的 `applyFieldDirective`（连同 Step 3 里对它的调用）。

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./cmd/docgen/ -run TestExtractMethods_Classification -v`
Expected: PASS。

- [ ] **Step 6: Commit**

```bash
git add cmd/docgen/extract.go cmd/docgen/model.go cmd/docgen/extract_test.go
git commit -m "feat(docgen): method classification, R/RW, normalized signatures, directives"
```

---

### Task 5: Example 自建索引

**Files:**
- Create: `cmd/docgen/examples.go`
- Modify: `cmd/docgen/extract.go`（在 withMethods 后调用 attachExamples）
- Test: `cmd/docgen/examples_test.go`

关联规则：`Example<Type>_<Method>[_<suffix>]` → 该 type 的该 method。多 suffix 按声明序。

- [ ] **Step 1: 写失败测试**

```go
package main

import "testing"

func TestAttachExamples_BindsToMethod(t *testing.T) {
	lp, _ := loadPackage("./testdata/sample")
	types := withMethods(extractTypes(lp), lp)
	attachExamples(types, "./testdata/sample")

	w := findType(types, "Widget")
	sn := findSym(w.methods, "SetName")
	if sn == nil || len(sn.examples) != 1 {
		t.Fatalf("SetName should have 1 example, got %+v", sn)
	}
	if !contains(sn.examples[0].code, "w.SetName") {
		t.Fatalf("example code wrong: %q", sn.examples[0].code)
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && stringIndex(s, sub) >= 0 }
func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./cmd/docgen/ -run TestAttachExamples_BindsToMethod -v`
Expected: FAIL（`attachExamples` 未定义）。

- [ ] **Step 3: 实现 attachExamples（examples.go）**

```go
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
	// ex.Code 是 *ast.BlockStmt（含 {}），去掉首尾大括号 + 缩进。
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	var lines []string
	for _, ln := range strings.Split(s, "\n") {
		lines = append(lines, strings.TrimPrefix(strings.TrimRight(ln, " \t"), "\t"))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
```

> `bindExample` 遍历 dt.methods 取地址有效，因为 `dt.methods` 是 slice，`&bucket[i]` 指向底层数组；但 `bucket := dt.methods` 是 slice 头拷贝、底层数组共享，append 到 examples 是改元素内部 slice 字段，有效。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./cmd/docgen/ -run TestAttachExamples_BindsToMethod -v`
Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add cmd/docgen/examples.go cmd/docgen/examples_test.go
git commit -m "feat(docgen): self-built Example index, bind to method by name"
```

---

### Task 6: prose 渲染（go/doc/comment → markdown）

**Files:**
- Create: `cmd/docgen/prose.go`
- Test: `cmd/docgen/prose_test.go`

- [ ] **Step 1: 写失败测试**

```go
package main

import (
	"strings"
	"testing"
)

func TestRenderProse_Markdown(t *testing.T) {
	in := "Widget 是一个示例 widget。\n\n第二段。"
	out := renderProse(in)
	if !strings.Contains(out, "Widget 是一个示例 widget。") {
		t.Fatalf("missing para1: %q", out)
	}
	if !strings.Contains(out, "第二段。") {
		t.Fatalf("missing para2: %q", out)
	}
	// 两段之间空行分隔。
	if !strings.Contains(out, "\n\n") {
		t.Fatalf("paragraphs not separated: %q", out)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./cmd/docgen/ -run TestRenderProse_Markdown -v`
Expected: FAIL（`renderProse` 未定义）。

- [ ] **Step 3: 实现 renderProse**

```go
package main

import (
	"go/doc/comment"
	"strings"
)

// renderProse 把 Go doc-comment 文本解析成结构再渲成 markdown。
// doc comment 不是 markdown：必须经 comment.Parser/Printer 转换（无原生表格）。
func renderProse(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	var p comment.Parser
	parsed := p.Parse(text)
	var pr comment.Printer
	return strings.TrimSpace(string(pr.Markdown(parsed)))
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./cmd/docgen/ -run TestRenderProse_Markdown -v`
Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add cmd/docgen/prose.go cmd/docgen/prose_test.go
git commit -m "feat(docgen): prose render via go/doc/comment markdown printer"
```

---

### Task 7: 渲染器 model → markdown（golden）

**Files:**
- Create: `cmd/docgen/render.go`
- Create: `cmd/docgen/testdata/golden/widget.md`
- Test: `cmd/docgen/render_test.go`

- [ ] **Step 1: 写期望 golden（widget.md）**

```markdown
# Widget object

Widget 是一个示例 widget。

第二段：演示 docgen 抽取多段 prose。

## Attributes

### Widget.Name

```go
Name string
```

Name 是显示名。

JSON: `name` · read-write

### Widget.Tags

```go
Tags []string
```

Tags 是标签。

JSON: `tags` · read-only

### Widget.Size

```go
func (w *Widget) Size() int
```

Size 返回标签数。

read-only

## Methods

### Widget.SetName

```go
func (w *Widget) SetName(name string) error
```

SetName 设置 Name。

当 name 为空时返回 error。

**Example:**

```go
w := &sample.Widget{}
_ = w.SetName("hi")
```

### Widget.Clone

```go
func (w *Widget) Clone() *Widget
```

Clone 返回副本。
```

> 渲染顺序约定：Attributes = 字段(结构序 Name,Tags) 后接 getter(源序 Size)；Methods = 动作方法源序(SetName, Clone)。

- [ ] **Step 2: 写失败测试**

```go
package main

import (
	"os"
	"testing"
)

func TestRenderType_Widget_Golden(t *testing.T) {
	lp, _ := loadPackage("./testdata/sample")
	types := withMethods(extractTypes(lp), lp)
	attachExamples(types, "./testdata/sample")
	w := findType(types, "Widget")

	got := renderType(w)
	want, err := os.ReadFile("./testdata/golden/widget.md")
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("render mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
```

- [ ] **Step 3: 跑测试确认失败**

Run: `go test ./cmd/docgen/ -run TestRenderType_Widget_Golden -v`
Expected: FAIL（`renderType` 未定义）。

- [ ] **Step 4: 实现 renderType**

```go
package main

import (
	"fmt"
	"strings"
)

// renderType 渲单个 docType 为 markdown（不含文件级 version header / includes）。
func renderType(t *docType) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s object\n", t.name)
	if p := renderProse(t.doc); p != "" {
		fmt.Fprintf(&b, "\n%s\n", p)
	}

	if len(t.attributes) > 0 {
		b.WriteString("\n## Attributes\n")
		for _, a := range t.attributes {
			renderSymbol(&b, t.name, a)
		}
	}
	if len(t.methods) > 0 {
		b.WriteString("\n## Methods\n")
		for _, m := range t.methods {
			renderSymbol(&b, t.name, m)
		}
	}
	if len(t.consts) > 0 {
		b.WriteString("\n## Constants\n")
		for _, c := range t.consts {
			if p := renderProse(c.doc); p != "" {
				fmt.Fprintf(&b, "\n%s\n", p)
			}
			fmt.Fprintf(&b, "\n```go\n%s\n```\n", strings.TrimSpace(c.code))
		}
	}
	return b.String()
}

func renderSymbol(b *strings.Builder, typeName string, s symbol) {
	fmt.Fprintf(b, "\n### %s.%s\n\n", typeName, s.name)
	if s.kind == kindField {
		fmt.Fprintf(b, "```go\n%s\n```\n", s.fieldDecl)
	} else {
		fmt.Fprintf(b, "```go\n%s\n```\n", s.signature)
	}
	if p := renderProse(s.doc); p != "" {
		fmt.Fprintf(b, "\n%s\n", p)
	}
	// Attributes 标 R/RW（含 JSON）。
	if s.kind == kindField || s.kind == kindGetter {
		rw := "read-only"
		if s.readWrite {
			rw = "read-write"
		}
		if s.jsonName != "" {
			fmt.Fprintf(b, "\nJSON: `%s` · %s\n", s.jsonName, rw)
		} else {
			fmt.Fprintf(b, "\n%s\n", rw)
		}
	}
	for _, ex := range s.examples {
		label := "Example"
		if ex.suffix != "" {
			label = "Example (" + ex.suffix + ")"
		}
		fmt.Fprintf(b, "\n**%s:**\n\n```go\n%s\n```\n", label, ex.code)
	}
}
```

- [ ] **Step 5: 跑测试，diff 调 golden**

Run: `go test ./cmd/docgen/ -run TestRenderType_Widget_Golden -v`
Expected: 首跑可能 FAIL（空白/换行差异）。比对 got vs want，**以渲染器输出为准修正 golden 文件的空白**（渲染器逻辑正确即可），直到 PASS。

- [ ] **Step 6: Commit**

```bash
git add cmd/docgen/render.go cmd/docgen/render_test.go cmd/docgen/testdata/golden/widget.md
git commit -m "feat(docgen): renderType model→markdown with golden test"
```

---

### Task 8: manifest + main 编排 + _includes + version header（集成 golden）

**Files:**
- Create: `cmd/docgen/manifest.go`
- Create: `cmd/docgen/main.go`
- Create: `cmd/docgen/testdata/manifest_sample.json`
- Create: `cmd/docgen/testdata/golden/sample_file.md`
- Test: `cmd/docgen/main_test.go`

- [ ] **Step 1: 写 manifest 类型 + 解析（manifest.go）**

```go
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// manifest：每个输出文件一节。所有相对路径相对 manifest 文件自身目录解析（CWD 无关），
// 这样 `go run`（CWD=根）和 `go generate`（CWD=cmd/docgen）行为一致。
type manifest struct {
	Pkg     string         `json:"pkg"` // 被解析的包目录
	Files   []fileManifest `json:"files"`
	baseDir string         // = filepath.Dir(manifestPath)，不序列化
}

type fileManifest struct {
	Out   string   `json:"out"`   // 输出 md 路径
	Roots []string `json:"roots"` // 根类型，顺序即渲染序
	Head  string   `json:"head"`  // 可选 _includes 路径
	Tail  string   `json:"tail"`  // 可选
}

func loadManifest(path string) (*manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	m.baseDir = filepath.Dir(path)
	return &m, nil
}

// resolve 把 manifest 内相对路径锚到 baseDir。
func (m *manifest) resolve(p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(m.baseDir, p)
}
```

- [ ] **Step 2: 写测试 manifest + golden**

`cmd/docgen/testdata/manifest_sample.json`：

```json
{
  "pkg": "./sample",
  "files": [
    { "out": "out/sample_file.md", "roots": ["Widget", "Mode"] }
  ]
}
```

> 路径相对 manifest 自身目录（`cmd/docgen/testdata/`）：`./sample` → `testdata/sample`。

`cmd/docgen/testdata/golden/sample_file.md`：先留空文件，Step 5 用生成结果回填（与 Task 7 同法）。

- [ ] **Step 3: 写失败测试**

```go
package main

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateFile_Golden(t *testing.T) {
	m, err := loadManifest("./testdata/manifest_sample.json")
	if err != nil {
		t.Fatal(err)
	}
	got, err := generateFile(m, m.Files[0])
	if err != nil {
		t.Fatal(err)
	}
	// version header 存在。
	if !strings.HasPrefix(got, "<!-- Code generated by docgen") {
		t.Fatalf("missing version header: %q", got[:60])
	}
	want, err := os.ReadFile("./testdata/golden/sample_file.md")
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("mismatch:\n--- got ---\n%s", got)
	}
}
```

- [ ] **Step 4: 实现 generateFile + main**

`main.go`：

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func main() {
	mf := flag.String("manifest", "docs/docgen.json", "manifest path")
	flag.Parse()
	m, err := loadManifest(*mf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "docgen:", err)
		os.Exit(1)
	}
	for _, fm := range m.Files {
		out, err := generateFile(m, fm)
		if err != nil {
			fmt.Fprintln(os.Stderr, "docgen:", err)
			os.Exit(1)
		}
		dst := m.resolve(fm.Out)
		if err := os.WriteFile(dst, []byte(out), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "docgen:", err)
			os.Exit(1)
		}
		fmt.Println("docgen: wrote", dst)
	}
}

// generateFile 渲一个输出文件：version header + head include + 各 root type + tail include。
func generateFile(m *manifest, fm fileManifest) (string, error) {
	pkgDir := m.resolve(m.Pkg)
	lp, err := loadPackage(pkgDir)
	if err != nil {
		return "", err
	}
	types := withMethods(extractTypes(lp), lp)
	attachExamples(types, pkgDir)

	var b strings.Builder
	fmt.Fprintf(&b, "<!-- Code generated by docgen; DO NOT EDIT. %s -->\n", runtime.Version())
	if fm.Head != "" {
		h, err := os.ReadFile(m.resolve(fm.Head))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "\n%s\n", strings.TrimSpace(string(h)))
	}
	for _, root := range fm.Roots {
		dt := findType(types, root)
		if dt == nil {
			return "", fmt.Errorf("root type %q not found in %s", root, pkgDir)
		}
		b.WriteString("\n")
		b.WriteString(renderType(dt))
	}
	if fm.Tail != "" {
		tl, err := os.ReadFile(m.resolve(fm.Tail))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "\n%s\n", strings.TrimSpace(string(tl)))
	}
	return b.String(), nil
}
```

- [ ] **Step 5: 跑测试，回填 golden**

Run: `go test ./cmd/docgen/ -run TestGenerateFile_Golden -v`
首跑会 FAIL 并打印 got。把 got 内容写进 `testdata/golden/sample_file.md`（核对结构正确：header + Widget + Mode(含 Constants)），再跑至 PASS。

- [ ] **Step 6: 全包测试回归**

Run: `go test ./cmd/docgen/ -v`
Expected: 全 PASS。

Run: `go vet ./cmd/docgen/`
Expected: 无输出。

- [ ] **Step 7: Commit**

```bash
git add cmd/docgen/manifest.go cmd/docgen/main.go cmd/docgen/main_test.go cmd/docgen/testdata/
git commit -m "feat(docgen): manifest + main orchestration + includes + version header"
```

---

### Task 9: 接 property 域 → 产出 property.gen.md + 验证清单

生成器已可用。此 Task 把它指向真实的 `internal/aep` property 域，产出 `docs/property.gen.md` 并按 spec 验证清单核对。**本 Task 不写新测试**——它是 pilot 运行 + 人工验证门。

**Files:**
- Create: `docs/docgen.json`
- Modify: `internal/aep/*`（给 Property 域导出符号补 doc comment prose —— 内容迁移，量大，见 Step 2）
- Create: `docs/property.gen.md`（生成产物）
- Modify: 仓库根加 `//go:generate`（落在 `cmd/docgen/main.go` 顶或新建 `doc.go`）

- [ ] **Step 1: 写 docs/docgen.json（先只挂 Property 类型）**

```json
{
  "pkg": "../internal/aep",
  "files": [
    {
      "out": "property.gen.md",
      "roots": ["Property"],
      "head": "_includes/property.head.md"
    }
  ]
}
```

> 路径相对 manifest 自身目录（`docs/`）：`../internal/aep` → repo `internal/aep`，`property.gen.md` → `docs/property.gen.md`。先只填 `Property`（spec：先验 Property 类型符号跑顺，再纳入 Keyframe/TemporalEase/InterpType）。`property.head.md` 放原 property.md 的类型间叙事/Example 引言（手写逃生舱）。

- [ ] **Step 2: 迁移 Property 域 doc comment**

把现有 `docs/property.md` 里 Property 各字段/方法的中文 prose，搬进 `internal/aep` 对应导出符号上方的 doc comment（`scene_property.go` 的 `Property` 类型与字段、`scene_property_flags.go` / `scene_property_defaults.go` 的 getter、`mutate_property_*.go` 的 setter）。复用本会话工作树里已写好的 property.md DimSep 增补作为 `SetDimensionsSeparated` 的 prose 源。建几个 `Example*` 测试函数（至少含一个 `_suffix` 多 Example 场景验证）。

> 这是内容迁移，无单测；逐符号搬，搬完即进 Step 3 生成对比。

- [ ] **Step 3: 加 //go:generate 并生成**

在 `cmd/docgen/main.go` 顶部（package 行下）加：

```go
//go:generate go run . -manifest ../../docs/docgen.json
```

因路径相对 manifest 目录解析，两种入口等价：`go generate ./cmd/docgen`（CWD=cmd/docgen，manifest=../../docs/docgen.json）与从根 `go run ./cmd/docgen -manifest docs/docgen.json` 产出相同。

Run（从仓库根）: `go run ./cmd/docgen -manifest docs/docgen.json`
Expected: `docgen: wrote docs/property.gen.md`（或等价绝对/相对路径）。

- [ ] **Step 4: 验证清单（spec §Pilot 验证闭环）**

逐项核对，结果记进 commit body 或 incident：

```bash
# 覆盖率：H3 标题数 old vs gen
grep -c '^### ' docs/property.md
grep -c '^### ' docs/property.gen.md
# 人工 diff 核对 prose 无丢失
git diff --no-index docs/property.md docs/property.gen.md
# Example 编译/跑通
go test ./internal/aep/ -run Example -v
```

- [ ] 覆盖率：H3 数量相当（缺口逐一解释：合并/拆分/确属丢失）。
- [ ] 表格场景：Property 域若有概念表（Components 值表等）→ 判定走 `_includes/` 还是渲染器简单管道表，结论记进 spec。
- [ ] 多 Example：`_suffix` 关联渲染正确。
- [ ] 中英混排：prose + 代码块 + 列表渲染可接受。
- [ ] directive：对一个歧义方法验证 `//docgen:method|attribute` 覆盖生效且不泄漏进 `go doc ./internal/aep`。

- [ ] **Step 5: Commit pilot 产物 + 决策**

```bash
git add docs/docgen.json docs/_includes/ docs/property.gen.md internal/aep/ cmd/docgen/main.go
git commit -m "feat(docgen): pilot — Property scope → docs/property.gen.md + validation findings"
```

> 验证**可接受** → follow-on：扩 roots 到 Keyframe/TemporalEase/InterpType、`.gen.md` 替换手写 `property.md`、推广其余 12 文件、接 CI `git diff --exit-code docs/`、改 CLAUDE.md/rules House rule。
> **不可接受** → 调 render 规则重生成，结论回写 spec。

---

## 自审（spec 覆盖）

- 唯一源=注释 / AST 推导结构 → Task 2-4。
- R/RW 推断+显式覆盖 → Task 4（setter 探测 + `//docgen:rw/ro/method/attribute`）。
- go/doc/comment 渲 prose（非 markdown、无表格）→ Task 6。
- 规范化单行签名 → Task 4 normalizeSignature。
- Example 自建索引 + 多 Example → Task 5。
- 稳定排序（roots 序 + 声明序 + suffix 序）→ Task 7/8 渲染序 + Task 5。
- version header + Go 版本 → Task 8 generateFile。
- manifest + _includes head/tail 一等公民 → Task 8。
- pilot property.gen.md 并列 + 验证清单 + 先 Property 后 Keyframe 族 → Task 9。
- CI 门禁 / 全量推广 / CLAUDE.md 调和 → Task 9 follow-on（本 plan 边界外，已注明）。
