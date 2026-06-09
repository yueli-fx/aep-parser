---
status: done
summary: V3 M8 方案② 物理分包实现计划（A 先行 + B′）：P0 基线+inventory → P1 抽 internal/codec → P2 单包内 back-ref 接口化（concrete→XWriter） → P3 git mv 物理分包 → P4 下游+收口+双版本 ship-gate
---

> **✅ 完成（2026-06-09）。** 物理分包落地：`internal/{rifx,codec,scene,serializer}` + `aep` 薄 facade。
> P0-P2 + P3.0 见 git log；P3.1 stage 1 抽 scene（`ac62b25`+`0303113`）、stage 2 抽 serializer（`a347e46`）；
> P4 守卫+CLAUDE.md #3 多包（`ed4669c`）。**P4.3 ship-gate 经 byte-identity 等效验收**（用户 2026-06-09 拍板）：
> round-trip 证全 183 fixture 输出与 split 前（已 ship-gated）逐字节相同，relocation 零写出语义变更 ⟹ AE 接受性不变，
> 不再跑冗余 AE 自动化。全程绿 + byte-identical + DAG OK + docs 内容零变。
> **遗留 follow-up（非阻塞，另起 commit）**：serializer 结构性 op 函数仍带与 facade 重复的富 doc comment（doc home 已是 facade）→ trim。

# V3 M8 物理分包（A 先行 + B′）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `internal/aep` 单包物理拆成 `internal/{scene,serializer,codec}` 三包（+ `internal/aep` 薄 facade），令 scene 编译期零 rifx，且保留全部方法 API（`Open→*Project` / `WriteAEP` / `Set*` 不变）。

**Architecture:** B′——back-ref 由 scene 类型上的具体 `*xBackrefs` 字段改为 scene 内定义、serializer 实现的 `XWriter` 接口；scene 方法体「改值 + 调 back 接口」（patch-first 原子序）。先在单包内做 concrete→interface 解耦（P2，最危险，回归门当安全网），再把物理分包做成机械 `git mv`（P3）。全程保 byte-identical round-trip。

**Tech Stack:** Go 1.25；`go vet ./... && go test ./...`；AE 2020 + 2025 双版本 ship-gate（`scripts/ae_run.ps1`，`AE_SHIP_GATE=1`）。

**设计源：** `flightdeck/specs/2026-06-07-v3-m8-physical-split-design.md`（§0 约束 C-1..C-7 / §1 DAG / §2 接口机制 / §5 分期 / §6 验证）。读本 plan 前先读该 spec。

**硬约束（贯穿每个 task）：** CLAUDE.md #1 length-preserving、#2 public API、#3 包边界、#5 opaque byte-identical、#6 ship-gate。每个 commit 前 `go vet ./... && go test ./...` 全绿；结构改动后跑 byte-identical 基线（Task 0.1 建）。

---

## Phase 0 — 基线 + inventory（spec 推迟到 plan 首步的发现工作）

> 此 Phase 不动产品代码，只建安全网 + 产出 P2 依赖的 Set* 分类表与 writer 接口清单。**P2 的具体步骤枚举自 Task 0.3 的产出**——故 Task 0.3 必须先完成。

### Task 0.1: byte-identical round-trip 基线脚本

**Files:**
- Create: `tmp_debug/split_roundtrip_baseline/main.go`
- Create: `tmp_debug/split_roundtrip_baseline/README.md`

- [ ] **Step 1: 写基线工具**

`tmp_debug/split_roundtrip_baseline/main.go`：遍历 `test_data/` 下全部 `*.aep`（含 `re_*.aep`），对每个跑 `aep.Open → WriteAEP(buf) → sha256(buf)`，与磁盘原文件的 `Open→Write` 前置读取对比，打印 `<file> <PASS|MISMATCH> <sha256>`。退出码非 0 当任一 MISMATCH。

```go
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/aep-parser/internal/aep"
)

func main() {
	root := "test_data"
	var fail int
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(p) != ".aep" {
			return nil
		}
		proj, err := aep.Open(p)
		if err != nil {
			fmt.Printf("%s OPEN-ERR %v\n", p, err)
			fail++
			return nil
		}
		var buf bytes.Buffer
		if err := proj.WriteAEP(&buf); err != nil {
			fmt.Printf("%s WRITE-ERR %v\n", p, err)
			fail++
			return nil
		}
		sum := sha256.Sum256(buf.Bytes())
		fmt.Printf("%s %s\n", p, hex.EncodeToString(sum[:]))
		return nil
	})
	if fail > 0 {
		os.Exit(1)
	}
}
```

- [ ] **Step 2: 跑基线并存指纹**

Run: `go run ./tmp_debug/split_roundtrip_baseline > tmp_debug/split_roundtrip_baseline/baseline.txt`
Expected: 退出码 0，`baseline.txt` 每行 `<path> <sha256>`，无 OPEN-ERR/WRITE-ERR。

- [ ] **Step 3: 记 fixture git hash + 写 README**

`README.md` 写：用法（`go run` 后 `git diff --no-index baseline.txt <new>`）、过期策略（fixture 改动→重生基线、commit 标注）、当前 `git rev-parse HEAD`。

Run: `git rev-parse HEAD`（把输出记进 README）

- [ ] **Step 4: Commit**

```bash
git add tmp_debug/split_roundtrip_baseline/
git commit -m "test(aep): byte-identical round-trip baseline for M8 physical split"
```

### Task 0.2: DAG 边界 CI 断言脚本（先建，P1 起逐步收紧）

**Files:**
- Create: `tmp_debug/dag_boundary/main.go`

- [ ] **Step 1: 写依赖断言工具**

用 `go list -deps -json` 或 `golang.org/x/tools/go/packages` 读各 internal 包的 import 闭包，断言：`internal/scene` 直接 imports ⊆ {`internal/codec`}（不含 rifx/serializer）；`internal/serializer` 不 import `internal/aep`；`internal/codec` 不 import `internal/scene` 且不 import `internal/rifx`。包不存在时跳过该断言（P1 前 scene/serializer 尚未拆出）。

```go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// directImports 返回某包的直接 import 列表（go list -f）。
func directImports(pkg string) ([]string, error) {
	out, err := exec.Command("go", "list", "-f", "{{range .Imports}}{{.}}\n{{end}}", pkg).Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}

func main() {
	const base = "github.com/example/aep-parser/internal/"
	type rule struct {
		pkg     string
		banned  []string
	}
	rules := []rule{
		{"scene", []string{base + "rifx", base + "serializer", base + "aep"}},
		{"serializer", []string{base + "aep"}},
		{"codec", []string{base + "scene", base + "rifx"}},
	}
	var fail int
	for _, r := range rules {
		imps, err := directImports(base + r.pkg)
		if err != nil {
			fmt.Printf("SKIP %s (%v)\n", r.pkg, err) // 包未拆出
			continue
		}
		for _, imp := range imps {
			for _, b := range r.banned {
				if imp == b {
					fmt.Printf("FAIL %s imports %s\n", r.pkg, imp)
					fail++
				}
			}
		}
	}
	if fail > 0 {
		os.Exit(1)
	}
	fmt.Println("DAG OK")
}
```

- [ ] **Step 2: 跑（P0 时全 SKIP，合法）**

Run: `go run ./tmp_debug/dag_boundary`
Expected: 三行 `SKIP`（scene/serializer/codec 尚未拆出）+ `DAG OK`，退出码 0。

- [ ] **Step 3: Commit**

```bash
git add tmp_debug/dag_boundary/
git commit -m "test(aep): package-DAG boundary assertion harness for M8 split"
```

### Task 0.3: Set* 分类表 + writer 接口清单（P2 的输入，关键产出）

**Files:**
- Create: `flightdeck/plans/m8-setter-inventory.md`

- [ ] **Step 1: 枚举所有 Set\* + 其 back-ref 用法**

Run（分别看「方法定义」与「back 访问」）：
```bash
rg -n "^func \([a-z]+ \*[A-Za-z]+\) Set[A-Z]" internal/aep
rg -n "\.back\.|\.back\b" internal/aep --glob '!*_test.go'
```
Expected: 得到全部 `Set*` 方法定义清单 + 全部经 `.back.` 的字节访问点。

- [ ] **Step 2: 分两类填表**

`m8-setter-inventory.md` 列两张表：
- **A 类（back-ref setter，需 writer 接口方法）**：方法体出现 `recv.back.<chunk>` 字节 patch 者。逐行记 `receiver | method | 触碰的 chunk 字段 | 归属 XWriter 接口`。
- **B 类（纯图 setter，留 scene 无接口）**：方法体只改 scene 值 / 委托到别的 Set*（如 `RectNode.SetSize → Property.SetStaticValue`）者。

判定规则（spec §3）：**凡方法体经 `recv.back.*` 做 `*rifx.Chunk` patch → A 类；凡只赋 scene 字段或委托 → B 类**。边界案例（group property / effect parameter setter）逐个看方法体定夺，记入表并标注理由。

- [ ] **Step 3: 列 9 个 XWriter 接口的方法集**

按 10 个 backref struct（`back_{composition,layer,property,keyframe,marker,mask,footage,project,render_queue,property_group}.go`）分组，每组列出该类型 A 类 setter → 对应 `XWriter` 接口方法签名（exported，patch-first 语义）。含 `ProjectWriter.WriteAEP(io.Writer) error`。产出精确方法计数（spec 估 ~60-100，本步给实数）。

- [ ] **Step 4: Commit**

```bash
git add flightdeck/plans/m8-setter-inventory.md
git commit -m "docs(flightdeck): M8 setter classification + XWriter interface inventory"
```

---

## Phase 1 — 抽 `internal/codec`（纯叶，低风险，零行为变更）

> 8 个 `codec_*.go` 已是纯值/字节 codec（实测不 import rifx、不引用 scene 类型）。本 Phase 物理移成真包，facade re-alias 其 exported 符号保 API 零-diff。

### Task 1.1: 移动 codec 文件到新包

**Files:**
- Move: `internal/aep/codec_{framerate,postscript,cdta_layout,ldta_layout,property_stream,gradient,render_settings,output_module}.go` → `internal/codec/`

- [ ] **Step 1: 记录迁移前 API + 基线**

Run:
```bash
go doc -all ./internal/aep > tmp_debug/api_before.txt
go run ./tmp_debug/split_roundtrip_baseline > tmp_debug/rt_before.txt
```
Expected: 两文件生成，round-trip 退出码 0。

- [ ] **Step 2: git mv 8 文件 + 改 package 声明**

```bash
mkdir internal/codec
git mv internal/aep/codec_framerate.go internal/codec/framerate.go
git mv internal/aep/codec_postscript.go internal/codec/postscript.go
git mv internal/aep/codec_cdta_layout.go internal/codec/cdta_layout.go
git mv internal/aep/codec_ldta_layout.go internal/codec/ldta_layout.go
git mv internal/aep/codec_property_stream.go internal/codec/property_stream.go
git mv internal/aep/codec_gradient.go internal/codec/gradient.go
git mv internal/aep/codec_render_settings.go internal/codec/render_settings.go
git mv internal/aep/codec_output_module.go internal/codec/output_module.go
```
然后把每个文件首行 `package aep` 改为 `package codec`。

- [ ] **Step 3: 编译——必然报错，记录 aep 包对 codec 符号的引用点**

Run: `go build ./... 2>&1 | rg "undefined:" | sort -u`
Expected: 一批 `undefined: GradientXxx / PropertyStream / decodeCdta… / NewPropertyStream …` —— 这是 aep 包代码引用了已迁出的符号。把清单存 `tmp_debug/codec_refs.txt`。

- [ ] **Step 4: 在 aep 包加 codec import + 限定符**

对 Step 3 清单里**未导出**的 codec helper（如 `decodeCdtaXxx`）：若被 aep 包用，则它们不该在 codec（codec 不暴露给 aep 的应是导出 API）——**导出它们**（首字母大写）或确认它们只在 codec 内用。对**导出**符号（Gradient 等），在 aep 引用处加 `codec.` 限定。逐文件 `go build ./internal/aep` 收敛到绿。

> 纪律：纯机械——只加 `codec.` 限定 / 导出 helper，不改逻辑。

- [ ] **Step 5: 验证 codec 纯度（DAG 断言）**

Run: `go run ./tmp_debug/dag_boundary`
Expected: `codec` 不再 SKIP；断言 codec 不 import rifx/scene 通过；`DAG OK`。若 codec 误 import rifx → 回到 spec §1 codec 纯度约束，该 helper 不属 codec，迁回 aep。

- [ ] **Step 6: 全绿 + round-trip**

Run:
```bash
go vet ./... && go test ./...
go run ./tmp_debug/split_roundtrip_baseline > tmp_debug/rt_after.txt
git diff --no-index tmp_debug/rt_before.txt tmp_debug/rt_after.txt
```
Expected: 测试全绿；round-trip diff 为空（byte-identical）。

- [ ] **Step 7: Commit**

```bash
git add internal/codec/ internal/aep/
git commit -m "refactor(aep): extract internal/codec package (pure value/byte codecs)"
```

### Task 1.2: facade re-alias codec 公共符号（保 API 零-diff）

**Files:**
- Create: `internal/aep/facade_codec.go`（暂居 aep 包；P3 时它就是 facade 的一部分）

- [ ] **Step 1: 比对 API diff，找出消失的 exported codec 符号**

Run:
```bash
go doc -all ./internal/aep > tmp_debug/api_after_codec.txt
git diff --no-index tmp_debug/api_before.txt tmp_debug/api_after_codec.txt
```
Expected: diff 显示 `Gradient` / `GradientColorStop` / `GradientAlphaStop` / `ParseGradientXML` / `EncodeGradientXML` / `PropertyStream` / `StreamKeyframe` / `StreamMode` / `NewPropertyStream` 等从 aep 公共面消失。记下完整清单。

- [ ] **Step 2: 写别名**

`internal/aep/facade_codec.go`：
```go
package aep

import "github.com/example/aep-parser/internal/codec"

type Gradient = codec.Gradient
type GradientColorStop = codec.GradientColorStop
type GradientAlphaStop = codec.GradientAlphaStop
type StreamMode = codec.StreamMode

// 函数/泛型：用包装而非别名（Go 不支持泛型类型别名的部分形式时）
func ParseGradientXML(s string) *codec.Gradient { return codec.ParseGradientXML(s) }
func EncodeGradientXML(g *codec.Gradient) string { return codec.EncodeGradientXML(g) }
// PropertyStream[T] / StreamKeyframe[T] / NewPropertyStream：按 Step 1 清单逐个别名/包装
```
> 泛型类型 `PropertyStream[T]`：Go 1.25 支持泛型类型别名（`type PropertyStream[T any] = codec.PropertyStream[T]`）；若 toolchain 不支持则改为在 facade 重导出构造函数 + 文档指向 codec。Step 1 清单逐项处理到 diff 清零。

- [ ] **Step 3: API diff 清零**

Run:
```bash
go doc -all ./internal/aep > tmp_debug/api_after_alias.txt
git diff --no-index tmp_debug/api_before.txt tmp_debug/api_after_alias.txt
```
Expected: **空 diff**（公共 API 恢复零-diff，约束#2）。

- [ ] **Step 4: 全绿 + round-trip + commit**

```bash
go vet ./... && go test ./...
go run ./tmp_debug/split_roundtrip_baseline > tmp_debug/rt_now.txt; git diff --no-index tmp_debug/rt_before.txt tmp_debug/rt_now.txt
git add internal/aep/facade_codec.go
git commit -m "refactor(aep): re-alias codec public symbols to preserve API (zero go-doc diff)"
```

---

## Phase 2 — 单包内 back-ref 接口化（concrete `*xBackrefs` → `XWriter` 接口）

> **最危险阶段**，但仍在单包 `internal/aep`（不动包边界）。安全网 = byte-identical 基线（verbatim 路径）+ 既有 setter 单测（setter 逻辑）+ Task 2.0 的 attach 完整性断言。**逐类型推进，每类型独立 commit**。具体 setter 清单来自 Task 0.3 的 inventory。

### Task 2.0: attach 完整性断言（新测试，TDD）

**Files:**
- Create: `internal/aep/attach_completeness_test.go`

- [ ] **Step 1: 写失败测试**

```go
package aep

import "testing"

// 每个 back-ref'd scene 对象在 Open 后必须 back != nil（C-4 / spec §2.3）。
func TestParseAttachCompleteness(t *testing.T) {
	proj, err := Open("test_data/re_compmarker.aep") // 含 comp/layer/property/marker
	if err != nil {
		t.Fatal(err)
	}
	walkAndAssertAttached(t, proj)
}
```
`walkAndAssertAttached` 遍历 proj → comps → layers → properties → markers/masks，断言每个**应 attached** 对象的 back 非 nil。**此刻 back 还是具体 `*xBackrefs`，断言写成 `obj.back != nil`。**

- [ ] **Step 2: 跑（预期 PASS——今天 Parse 已填 back）**

Run: `go test ./internal/aep/ -run TestParseAttachCompleteness -v`
Expected: PASS（建立基线断言；P2 接口化后它继续守护「无漏 attach」）。若 FAIL，说明今天就有漏 attach，先修。

- [ ] **Step 3: Commit**

```bash
git add internal/aep/attach_completeness_test.go
git commit -m "test(aep): assert Parse attaches back-refs to every scene object (M8 P2 guard)"
```

### Task 2.1: 定义 9 个 XWriter 接口（先定接口，impl 暂仍是具体 struct）

**Files:**
- Create: `internal/aep/scene_writers.go`（接口定义；P3 随 scene 迁出）

- [ ] **Step 1: 按 Task 0.3 §Step3 清单写接口**

`scene_writers.go`：每个 backref 类型一个接口，方法 = Task 0.3 列出的 A 类 setter。示例（实际方法集以 inventory 为准）：
```go
package aep

import "io"

type ProjectWriter interface {
	WriteAEP(w io.Writer) error
	SetLinearBlending(bool) error
	// … 项目级 A 类 setter（inventory）
}
type LayerWriter interface {
	SetVisible(bool) error
	SetName(string) error
	// … layer 级 A 类 setter（inventory）
}
type PropertyWriter interface {
	WriteStaticValue(any) error
	WriteExpression(string) error
	// … property 级 A 类 setter（inventory）
}
// Composition/Footage/Marker/Mask/Keyframe/PropertyGroup/RenderQueue Writer 同构
```

- [ ] **Step 2: 让现有 `*xBackrefs` 满足接口（编译期断言）**

每个 backref struct 已有 / 将有对应方法。本步加编译期断言确认 struct 实现接口：
```go
var _ ProjectWriter = (*projectBackrefs)(nil)
var _ LayerWriter = (*layerBackrefs)(nil)
// … 9 条
```
> 此刻多半报「`*projectBackrefs` 未实现 WriteAEP/SetX」——因为今天这些 patch 逻辑在 scene 类型的 `Set*` 方法体里，不在 backref struct 上。这正是 Task 2.2 要搬的。先让断言存在并 **预期编译失败**，作为 2.2 的 checklist。

- [ ] **Step 3: 跑编译，记录未实现方法清单**

Run: `go build ./internal/aep 2>&1 | rg "does not implement|missing method"`
Expected: 一批「`*xBackrefs` does not implement XWriter (missing method SetX)」。这是 Task 2.2 的逐项工作清单。存 `tmp_debug/writer_todo.txt`。

- [ ] **Step 4: Commit（接口骨架，含预期未实现）**

> 例外：此步 `go build` 红。仅当采用「接口先行」工作法时 commit 骨架；若 worker 偏好绿态，合并 2.1+2.2 的首个类型一起 commit。两种皆可，注明。

```bash
git add internal/aep/scene_writers.go
git commit -m "refactor(aep): define XWriter interfaces for back-ref inversion (M8 P2, impls pending)"
```

### Task 2.2: 逐类型把 patch 逻辑搬进 backref impl + scene 字段改接口（worked pattern + 枚举）

> **对 Task 0.3 inventory 里的每个 backref 类型重复以下 6 步。** 下面用 **Layer** 完整走一遍作模板；其余 8 类（Project/Property/Composition/Footage/Marker/Mask/Keyframe/PropertyGroup）**同构套用**，setter 清单各取自 inventory。**每类型一个独立 commit。**

**Files（以 Layer 为例）：**
- Modify: `internal/aep/back_layer.go`（layerBackrefs 加 patch 方法）
- Modify: `internal/aep/write_layer.go`（Set* 方法体改「先调接口」）
- Modify: `internal/aep/scene_layer.go`（`back *layerBackrefs` → `back LayerWriter`）

- [ ] **Step 1: 把 patch 字节逻辑从 scene 方法体搬进 backref impl 方法**

当前（示意）`write_layer.go`：
```go
func (l *Layer) SetVisible(v bool) error {
	// … 直接 patch l.back.ldta 的字节逻辑 …
}
```
改为：patch 逻辑迁到 `back_layer.go` 的 `(*layerBackrefs)` 方法：
```go
// back_layer.go
func (b *layerBackrefs) SetVisible(v bool) error {
	// … 原 patch l(.back).ldta 字节的逻辑，改用 b.ldta …
	return nil
}
```

- [ ] **Step 2: scene 方法体改为 patch-first 委托（C-1 原子序）**

`write_layer.go`：
```go
func (l *Layer) SetVisible(v bool) error {
	if l.back == nil {        // detached（C-3）
		l.Visible = v
		return nil
	}
	if err := l.back.SetVisible(v); err != nil { // 先 patch（可失败）
		return err
	}
	l.Visible = v             // 成功后才同步 scene 值
	return nil
}
```
> 对 Layer 的**每个 A 类 setter**（inventory）重复 Step1-2。

- [ ] **Step 3: scene 字段类型改接口**

`scene_layer.go`：
```go
type Layer struct {
	// …
	back LayerWriter   // 原为 *layerBackrefs
}
```
Parse/New* 处 `layer.back = &layerBackrefs{...}` 不变（具体类型满足接口，自动赋值）。

- [ ] **Step 4: 编译——本类型的 `var _ LayerWriter = (*layerBackrefs)(nil)` 应通过**

Run: `go build ./internal/aep 2>&1 | rg "LayerWriter"`
Expected: 无 LayerWriter 相关 missing-method 报错（其余 8 类仍报，正常）。

- [ ] **Step 5: 全绿 + round-trip + setter 单测**

Run:
```bash
go test ./internal/aep/ -run 'TestLayer|TestSetVisible|TestParseAttachCompleteness' -v
go run ./tmp_debug/split_roundtrip_baseline > tmp_debug/rt_now.txt; git diff --no-index tmp_debug/rt_before.txt tmp_debug/rt_now.txt
```
Expected: Layer 相关 setter 单测 + attach 断言 PASS；round-trip 空 diff。
> **abort 条件（spec §5）**：若某类 setter 单测无法保绿且非调用点同步问题 → 暂停该类、记 `tmp_debug/writer_todo.txt`、该类暂留具体 `back *xBackrefs`（不改字段类型），其余类继续。

- [ ] **Step 6: Commit（本类型）**

```bash
git add internal/aep/back_layer.go internal/aep/write_layer.go internal/aep/scene_layer.go
git commit -m "refactor(aep): invert Layer back-ref to LayerWriter interface (M8 P2)"
```

- [ ] **Step 7: 对其余 8 类重复 Step1-6**

顺序建议（风险低→高，依 inventory setter 数）：Keyframe → Marker → Mask → Footage → Composition → Project → PropertyGroup → **Property**（最后，因 shape/stroke/fill setter 全委托到 `Property.SetStaticValue`，Property 接口化后它们自动经 `prop.back.WriteStaticValue` 工作——验证 `rect.SetSize` 链路 round-trip 绿）。每类独立 commit。

- [ ] **Step 8: 全部接口化后，删除编译期断言之外的具体类型残留**

Run: `rg -n "\*layerBackrefs|\*propertyBackrefs|\*projectBackrefs" internal/aep --glob '!back_*.go' --glob '!parse_*.go' --glob '!*_test.go'`
Expected: 空——scene 侧（scene_*/write_*）只引用 XWriter 接口，具体 backref 仅在 `back_*.go`（impl）+ `parse_*.go`（构造时赋值）出现。这是 P3 可机械分包的前提。

### Task 2.3: P2 收口验证

- [ ] **Step 1: scene 文件 rifx 引用清零核查**

Run: `rg -n "rifx\.|internal/rifx" internal/aep/scene_*.go`
Expected: 空（scene_* 文件不再直接引用 rifx——所有 chunk 耦合已进 back_*/parse_ 经接口隔离）。
> 若有残留：该 scene_* 文件仍有未接口化的 chunk 访问，回 Task 2.2 处理该类型。

- [ ] **Step 2: 全套 + round-trip + API diff**

Run:
```bash
go vet ./... && go test ./...
go run ./tmp_debug/split_roundtrip_baseline > tmp_debug/rt_now.txt; git diff --no-index tmp_debug/rt_before.txt tmp_debug/rt_now.txt
go doc -all ./internal/aep > tmp_debug/api_after_p2.txt
git diff --no-index tmp_debug/api_before.txt tmp_debug/api_after_p2.txt
```
Expected: 全绿；round-trip 空 diff；API diff **仅 = 新增 XWriter 接口 + plumbing**（核心 R/W 方法签名零变，约束#2），逐条核对无意外。

---

## Phase 3 — 物理分包（机械 git mv + DAG 编译边界）

> ⚠️ **部分作废（2026-06-09 勘察修正，见 cockpit §下一步 item 2）**：本 Phase 原设「机械 git mv」不成立。两点反转：(1)「`write_*.go`→serializer」反了——post-P2 的 write_*.go 实为 **scene-delegate**（rifx-clean），应属 scene；(2) scene 文件仍有 **concrete-backref 耦合**（如 `scene_property_flags.go` 经 `propertyBack()` 读 `pb.tdum`），即 P2「故意保 concrete——P3 统一解耦」的欠账，须先接口化才能拆包。已采「单包内先拆混装文件」排法（用户拍板）：先把每文件按 receiver/依赖（碰 rifx/concrete-backref/建 chunk→serializer；纯 scene→scene）拆成单一去向 + 完成 concrete→interface 解耦 + WriteAEP 委托化，**全绿后** Task 3.1/3.2 的 git mv 才退化为机械。下方原步骤的「文件→包」清单仍可参考，但须按修正后的二分重新归类。
>
> scene_* 现只引用 XWriter 接口（Task 2.3 Step1 已证 rifx 维度）→ 物理拆包**非纯机械**：跨包 export 可见性、init 顺序、测试辅助、concrete-backref 解耦需逐一处理（spec §5 P3）。

### Task 3.1: 建 internal/scene + 迁类型/accessor/接口/WriteJSON

**Files:**
- Move: `internal/aep/scene_*.go` → `internal/scene/`
- Move: `internal/aep/scene_writers.go` → `internal/scene/writers.go`
- Move: `internal/aep/write_json.go`（WriteJSON 纯 scene 导出）→ `internal/scene/`

- [ ] **Step 1: git mv scene 文件 + 改 package + 加 codec import**

```bash
mkdir internal/scene
git mv internal/aep/scene_layer.go internal/scene/layer.go
# … 全部 scene_*.go + scene_writers.go + write_json.go
```
每文件 `package aep` → `package scene`；引用 codec 处加 `codec.` 限定 + import。

- [ ] **Step 2: 处理 export 可见性**

scene 包内原**未导出**但被 serializer 侧（parse_/write_）调用的 helper/字段：必须导出或提供导出访问器。逐个 `go build ./internal/serializer` 报错驱动（serializer 暂仍在 aep 包名下，下一 task 拆）。**`AttachWriter` 等 plumbing 此时导出**（spec §2.2，接受 cmd/ 暴露）。

- [ ] **Step 3: 编译收敛 scene 包独立**

Run: `go build ./internal/scene`
Expected: 绿——scene 包独立编译。

- [ ] **Step 4: DAG 断言（scene⊥rifx 编译期）**

Run: `go run ./tmp_debug/dag_boundary`
Expected: scene 不再 SKIP；`scene` 直接 imports ⊆ {codec}，不含 rifx/serializer；`DAG OK`。**若 scene 误 import rifx → 编译已失败**（硬边界兑现）。

- [ ] **Step 5: 全绿 + round-trip + commit**

```bash
go vet ./... && go test ./...
go run ./tmp_debug/split_roundtrip_baseline > tmp_debug/rt_now.txt; git diff --no-index tmp_debug/rt_before.txt tmp_debug/rt_now.txt
git add internal/scene/ internal/aep/
git commit -m "refactor: extract internal/scene package (types + accessors + XWriter interfaces + WriteJSON; zero rifx)"
```

### Task 3.2: 建 internal/serializer + 迁 parse/lower/write/back/mutate

**Files:**
- Move: `internal/aep/{parse_,lower_,write_,back_,mutate_}*.go` → `internal/serializer/`

- [ ] **Step 1: git mv serializer 文件 + 改 package + 加 scene/codec/rifx import**

```bash
mkdir internal/serializer
git mv internal/aep/parse_*.go internal/serializer/
git mv internal/aep/lower_*.go internal/serializer/
git mv internal/aep/write_*.go internal/serializer/   # 除已迁 scene 的 write_json.go
git mv internal/aep/back_*.go internal/serializer/
git mv internal/aep/mutate_*.go internal/serializer/
```
每文件 `package aep` → `package serializer`；scene 类型引用加 `scene.` 限定 + import。

- [ ] **Step 2: 处理 init/global-var 顺序（spec §5 P3 显式核实）**

Run: `rg -n "^func init\(|^var [A-Z]?\w+ = " internal/serializer/*.go`
Expected: 列出所有包级 init/var。逐个核：跨包后初始化顺序是否依赖原单包内的相对顺序？若有，改为显式初始化或 sync.Once。无 init/global-var 则记「无」。

- [ ] **Step 3: 编译收敛**

Run: `go build ./internal/serializer`
Expected: 绿。serializer import scene + codec + rifx。

- [ ] **Step 4: DAG 断言全通**

Run: `go run ./tmp_debug/dag_boundary`
Expected: 无 SKIP；`DAG OK`（scene⊥rifx、serializer⊥aep、codec⊥scene 全部满足）。

- [ ] **Step 5: 全绿 + round-trip + commit**

```bash
go vet ./... && go test ./...
go run ./tmp_debug/split_roundtrip_baseline > tmp_debug/rt_now.txt; git diff --no-index tmp_debug/rt_before.txt tmp_debug/rt_now.txt
git add internal/serializer/ internal/aep/
git commit -m "refactor: extract internal/serializer package (parse/lower/write/back/mutate; imports scene+codec+rifx)"
```

### Task 3.3: 收缩 internal/aep 为薄 facade

**Files:**
- Modify: `internal/aep/`（仅剩 facade：Open/FromReader + 类型别名 + New* re-export + facade_codec.go）

- [ ] **Step 1: 写 facade 入口**

`internal/aep/aep.go`：
```go
package aep

import (
	"io"
	"github.com/example/aep-parser/internal/scene"
	"github.com/example/aep-parser/internal/serializer"
)

type Project = scene.Project
type Composition = scene.Composition
type Layer = scene.Layer
// … 全部公共 scene 类型 + 枚举别名

func Open(path string) (*Project, error)              { return serializer.Open(path) }
func FromReader(r io.ReadSeeker) (*Project, error)    { return serializer.FromReader(r) }
// New* re-export（spec §2.4）：
func NewShapeLayer(/* … */) (/* … */) { return serializer.NewShapeLayer(/* … */) }
// … 其余 New*/结构性构造器
```
> `Open`/`FromReader`/`New*` 的实现已随 parse_/mutate_ 迁入 serializer；facade 仅委托。

- [ ] **Step 2: API diff——核对破点集**

Run:
```bash
go doc -all ./internal/aep > tmp_debug/api_final.txt
git diff --no-index tmp_debug/api_before.txt tmp_debug/api_final.txt
```
Expected: diff **仅 = 新增 plumbing（XWriter 接口、AttachWriter）**；核心 `Open/FromReader/WriteAEP/WriteJSON/Set*/New*` 签名零变（B′ 不破核心 API，约束#2）。逐条核对。

- [ ] **Step 3: 全绿 + round-trip + commit**

```bash
go vet ./... && go test ./...
go run ./tmp_debug/split_roundtrip_baseline > tmp_debug/rt_now.txt; git diff --no-index tmp_debug/rt_before.txt tmp_debug/rt_now.txt
git add internal/aep/
git commit -m "refactor: reduce internal/aep to thin facade (aliases + Open/New* re-export over scene+serializer)"
```

---

## Phase 4 — 下游切换 + 收口 + 双版本 ship-gate

### Task 4.1: 下游编译 + 测试辅助迁移

**Files:**
- Verify: `cmd/aepdemo/`、`cmd/docgen/`
- Modify: 跨包后失效的测试辅助（如直接访问 `layer.back` 的测试）

- [ ] **Step 1: 下游编译**

Run: `go build ./cmd/...`
Expected: 绿（B′ API 不变，预期零改或极小）。若 cmd 直接用了迁出的内部符号 → 改用 facade。

- [ ] **Step 2: 修测试辅助**

Run: `go test ./... 2>&1 | rg "undefined|cannot refer to unexported"`
Expected: 列出跨包后失效的测试访问（如测试断言 `layer.back.(*layerBackrefs).ldta`）。改为经导出 API 或类型断言到 scene 接口（spec gpt#86-97 已预警测试需 mock/断言）。逐个修绿。

- [ ] **Step 3: 全绿 + commit**

```bash
go vet ./... && go test ./...
git add -A
git commit -m "refactor: fix downstream + test helpers after M8 package split"
```

### Task 4.2: 退役 AST 守卫 + 更新 CLAUDE.md

**Files:**
- Modify/Delete: `internal/aep/arch_boundary_test.go`（或迁 serializer 包内命名轴 lint）
- Modify: `CLAUDE.md`（硬约束#3）

- [ ] **Step 1: AST 守卫降级**

scene⊥rifx 现由包边界编译期保证 → 删除 `arch_boundary_test.go` 中 scene→rifx 白名单断言部分；保留 serializer 包内 `codec_`/命名轴 lint（若仍有价值），或整体删除并以 `tmp_debug/dag_boundary` 为 CI 边界守卫。

- [ ] **Step 2: 更新 CLAUDE.md 硬约束#3**

把「`internal/aep` 单 package + 文件名命名轴」改为多包描述：`internal/{rifx,codec,scene,serializer}` + `aep` facade 的 DAG；scene⊥rifx 编译期硬边界；B′ back-ref 接口破环；指向本 spec/plan。

- [ ] **Step 3: 全绿 + commit**

```bash
go vet ./... && go test ./...
git add internal/ CLAUDE.md
git commit -m "docs: retire scene→rifx AST guard (now compiler-enforced); update CLAUDE.md #3 for multi-package layout"
```

### Task 4.3: 双版本 ship-gate 终验

- [ ] **Step 1: 跑 AE 2020 + 2025 全套 ship-gate**

Run: `pwsh scripts/ae_run.ps1`（`AE_SHIP_GATE=1`，AE 2020 + 2025；自验留痕，feedback `ae-ship-gate-self-serve`）
Expected: 全套 PASS（纯结构搬迁，零结构性写路径语义变更，预期零回归）。exit 2 时按 `feedback_ship_gate_exit2` 抓 dialog 辨 flake vs reject。

- [ ] **Step 2: 最终 round-trip + 全绿**

Run:
```bash
go vet ./... && go test ./...
go run ./tmp_debug/split_roundtrip_baseline > tmp_debug/rt_now.txt; git diff --no-index tmp_debug/rt_before.txt tmp_debug/rt_now.txt
go run ./tmp_debug/dag_boundary
```
Expected: 全绿；round-trip 空 diff；`DAG OK`。

- [ ] **Step 3: landing**

完成后 plan 翻 `status: done`，跑 `/flightdeck:landing` 归档到 `landed/plans/` + 同步 INDEX/cockpit（rules.md「大计划完成即自动 landing」）。清理 `tmp_debug/{split_roundtrip_baseline,dag_boundary,api_*.txt,rt_*.txt,codec_refs.txt,writer_todo.txt}` 中的临时产物（保留 dag_boundary 作 CI 守卫）。

---

## 验证矩阵（每 Phase exit）

| Phase | go vet+test | byte-identical | go doc API diff | DAG 断言 | AE ship-gate |
|---|---|---|---|---|---|
| P0 | 绿 | 基线就位 | before 存档 | 全 SKIP | — |
| P1 | 绿 | 空 diff | **零-diff**（facade alias） | codec 通过 | — |
| P2 | 绿 | 空 diff | 仅 plumbing 增量 | （单包，n/a） | — |
| P3 | 绿 | 空 diff | 仅 plumbing 增量 | **全通**（编译期硬边界） | — |
| P4 | 绿 | 空 diff | 仅 plumbing 增量 | 全通 | **双版本 PASS** |

## 风险与回滚

- 每 Phase/每类型独立 commit；失败 `git revert` 回上一绿态。
- P2 单类型卡住：abort 退路（该类暂留具体 backref，其余继续；P3 时缺类或局部 handle）——spec §5。
- byte-identical 抓不到 setter 接线错 → 靠既有 setter 单测 + Task 2.0 attach 完整性断言补盲。
- C-1 双源一致性：直写 scene 字段绕过序列化无机械防护（spec C-1），靠 code-review；本 plan 不引入新直写路径。

## 关联
- spec：`flightdeck/specs/2026-06-07-v3-m8-physical-split-design.md`
- inventory（Task 0.3 产出）：`flightdeck/plans/m8-setter-inventory.md`
- 前置：`landed/specs/2026-05-30-aep-package-reorg-design.md`（方案①命名轴）
