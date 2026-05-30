# aep package 重组 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `internal/aep` 70 个源文件沿统一 `<stage>_<domain>` 命名轴重组、拆大文件、归并测试基建、并加 AST 边界守卫，零行为变更。

**Architecture:** 单包内重组（非物理分包，理由见 spec §0）。7 阶段前缀 `scene_/codec_/parse_/lower_/write_/back_/mutate_` + §2.1 依赖矩阵，由 `arch_boundary_test.go`（AST）强制。strangler 六阶段，每阶段 commit、suite 全绿、byte-identical round-trip 验证。

**Tech Stack:** Go 1.18+（泛型）、`go/parser`+`go/ast`（边界守卫）、`crypto/sha256`（round-trip 基线）、`git mv`（保 rename 链）。

**Spec:** `flightdeck/specs/2026-05-30-aep-package-reorg-design.md`（本计划逐条实现其 §3 映射 + §9 阶段）。

---

## 全局验证门（Gate）—— 每次移动/拆分文件后必跑

```powershell
# 1. 编译 + vet + 单测全绿
go build ./... ; if ($LASTEXITCODE) { throw } ; go vet ./internal/aep/ ; if ($LASTEXITCODE) { throw } ; go test -count=1 ./internal/aep/ ; if ($LASTEXITCODE) { throw }
# 2. 公共 API 零变更（基线 tmp/api_before.txt 是**排序归一化**指纹, ~3506 行; 实现者**禁止重新生成它**）
#    go doc -all 输出: ① 内嵌源文件名行(`internal/aep/foo.go`) ② 改名/拆分会重排声明顺序+移动空行 → 全是噪声。
#    指纹须: 滤文件名行 + 滤空行 + **Sort-Object**(集合比对, 消除重排假阳性)。base 与 after 都排序后比, 空 = exported API 集合不变。
(go doc -all ./internal/aep) -notmatch '^internal/aep/.*\.go$' -notmatch '^\s*$' | Sort-Object > tmp/api_after.txt ; (Compare-Object (Get-Content tmp/api_before.txt) (Get-Content tmp/api_after.txt))  # 输出须为空
# 3. round-trip 字节稳定（基线 Task 1 生成）
go test -count=1 ./internal/aep/ -run TestReorgRoundtripBaseline   # 须 PASS
```

> **改名纪律**：一律 `git mv`（保 rename 追踪）；同包内改名**不触碰任何引用处**（symbol 包内可见，引用与文件名无关）——故 Gate 第 1 步若编译失败，必是漏移了某个文件或拆分时漏带 import，不是「引用没改」。每批 `git mv` 独立 commit。

> **注释纪律**（写/改任何注释或拆分搬运注释时遵守 `flightdeck/checklists/comments.md`）：注释只写 why / 不变量 / 坑；**禁止** spec/plan 引用（`// §6`、`// 详见 spec`）、阶段码（`// Phase 5`、`// P4`、`// T7`）、reviewer 出处、日期/署名、历史考古（`// 之前用 X`）。本计划代码块里的标识符开头 doc 注释已合规；搬运既有注释时**不得**注入上述过程元信息。错误信息/日志文案同样**自描述**（写"scene model must stay chunk-free"而非"§6 violation"）。

---

## Task 0: 分支 + API 基线

**Files:**
- Create: `tmp/api_before.txt`（gitignored）

- [ ] **Step 1: 建分支**

```powershell
git checkout -b refactor/aep-package-reorg
```

- [ ] **Step 2: 捕获公共 API 基线（过滤内嵌文件名行）**

```powershell
New-Item -ItemType Directory -Force tmp | Out-Null
(go doc -all ./internal/aep) -notmatch '^internal/aep/.*\.go$' -notmatch '^\s*$' | Sort-Object > tmp/api_before.txt
```
`go doc -all` 每个文件声明前加源文件名行，且改名/拆分会重排声明顺序、移动空行。指纹须滤文件名行 + 滤空行 + **排序**（变成顺序无关的集合指纹）。一旦生成，**整个重组期间不可重新生成**（重排或新增 test 文件都会变；后续 Gate 只读它做比对）。

- [ ] **Step 3: 确认基线非空且无文件名残留**

```powershell
(Get-Content tmp/api_before.txt | Measure-Object -Line).Lines               # 应 ~3506
(Select-String -Path tmp/api_before.txt -Pattern '^internal/aep/.*\.go$').Count   # 应 = 0
```

此文件是后续每个 Gate 第 2 步的比对基准，**整个重组期间不可重新生成**。

> **注**：Task 0 已由 controller 执行（分支 `refactor/aep-package-reorg` 已建、`tmp/api_before.txt` 已归一化为 3506 行）。此任务留作记录与重跑依据。

---

## Task 1: Round-trip 字节基线（generator + test）

**Files:**
- Create: `tmp_debug/reorg_baseline/main.go`（生成器）
- Create: `tmp_debug/reorg_baseline/baseline.json`（生成产物，commit）
- Create: `internal/aep/reorg_baseline_test.go`（守卫测试）

不变量：重组是纯文件移动，故每个 fixture 的 `Open → WriteAEP` 输出字节**必须跨重组保持不变**。基线快照当前 main 的输出 sha256，测试断言其不变。

- [ ] **Step 1: 写生成器**

`tmp_debug/reorg_baseline/main.go`：
```go
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/example/aep-parser/internal/aep"
)

type entry struct {
	OpenOK bool   `json:"open_ok"`
	SHA    string `json:"sha,omitempty"`
	Err    string `json:"err,omitempty"`
}

func main() {
	roots := []string{"test_data", "data"}
	out := map[string]entry{}
	for _, root := range roots {
		_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || filepath.Ext(p) != ".aep" {
				return nil
			}
			key := filepath.ToSlash(p)
			proj, oerr := aep.Open(p)
			if oerr != nil {
				out[key] = entry{OpenOK: false, Err: oerr.Error()}
				return nil
			}
			var buf bytes.Buffer
			if werr := proj.WriteAEP(&buf); werr != nil {
				out[key] = entry{OpenOK: true, Err: "WriteAEP: " + werr.Error()}
				return nil
			}
			sum := sha256.Sum256(buf.Bytes())
			out[key] = entry{OpenOK: true, SHA: hex.EncodeToString(sum[:])}
			return nil
		})
	}
	keys := make([]string, 0, len(out))
	for k := range out {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ordered := map[string]entry{}
	for _, k := range keys {
		ordered[k] = out[k]
	}
	b, _ := json.MarshalIndent(ordered, "", "  ")
	if err := os.WriteFile("tmp_debug/reorg_baseline/baseline.json", b, 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("baseline: %d fixtures\n", len(out))
}
```

- [ ] **Step 2: 生成基线（在未改动的 main 代码上）**

```powershell
go run ./tmp_debug/reorg_baseline
```
预期：`baseline: 115 fixtures`（数量以实际 `*.aep` 为准）。

- [ ] **Step 3: 写守卫测试**

`internal/aep/reorg_baseline_test.go`：
```go
package aep_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/example/aep-parser/internal/aep"
)

type reorgEntry struct {
	OpenOK bool   `json:"open_ok"`
	SHA    string `json:"sha,omitempty"`
	Err    string `json:"err,omitempty"`
}

// TestReorgRoundtripBaseline 断言每个 fixture 的 Open→WriteAEP 输出跨重组不变。
// 基线由 tmp_debug/reorg_baseline 生成；缺失时 skip（仅重组窗口内启用）。
func TestReorgRoundtripBaseline(t *testing.T) {
	const baselinePath = "../../tmp_debug/reorg_baseline/baseline.json"
	raw, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Skipf("no reorg baseline (%v) — generate via `go run ./tmp_debug/reorg_baseline`", err)
	}
	var want map[string]reorgEntry
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("unmarshal baseline: %v", err)
	}
	for path, exp := range want {
		path, exp := path, exp
		t.Run(path, func(t *testing.T) {
			proj, oerr := aep.Open("../../" + path)
			if !exp.OpenOK {
				if oerr == nil {
					t.Fatalf("expected Open error %q, got nil", exp.Err)
				}
				return // Open 失败的 fixture：只要仍失败即可（不比错误文本）
			}
			if oerr != nil {
				t.Fatalf("Open: %v (baseline had OpenOK)", oerr)
			}
			var buf bytes.Buffer
			if werr := proj.WriteAEP(&buf); werr != nil {
				t.Fatalf("WriteAEP: %v", werr)
			}
			sum := sha256.Sum256(buf.Bytes())
			if got := hex.EncodeToString(sum[:]); got != exp.SHA {
				t.Fatalf("WriteAEP byte mismatch: baseline %s got %s", exp.SHA, got)
			}
		})
	}
}
```

- [ ] **Step 4: 跑测试确认 PASS（代码未改，应全绿）**

```powershell
go test -count=1 ./internal/aep/ -run TestReorgRoundtripBaseline -v
```
预期：每个 fixture 一个 PASS 子测试。

- [ ] **Step 5: 写 README + commit**

`tmp_debug/reorg_baseline/README.md`：说明用法（`go run ./tmp_debug/reorg_baseline` 生成；测试 `TestReorgRoundtripBaseline` 消费）、**过期策略**（重组完成 + landing 后删除此目录 + 测试文件 + 移除 baseline.json）、基线对应的 fixture 取自当前工作树（记录 `git rev-parse HEAD` 于 README 顶部一行）。

```powershell
git rev-parse HEAD   # 记到 README 顶部
git add tmp_debug/reorg_baseline/ internal/aep/reorg_baseline_test.go
git commit -m "test(reorg): round-trip byte baseline harness (115 fixtures, sha256 snapshot)"
```

---

## Task 2: AST 边界守卫（TDD：先失败枚举白名单，再放行）

**Files:**
- Create: `internal/aep/arch_boundary_test.go`

实现 §6：`scene_` 禁 import rifx；`codec_` 禁引用 scene 类型。AST 扫描，**首跑空白名单 → 失败列出全部违例（= 权威白名单底数）**，再填白名单转 PASS。

- [ ] **Step 1: 写守卫（空白名单）**

`internal/aep/arch_boundary_test.go`：
```go
package aep

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

const rifxImportPath = "github.com/example/aep-parser/internal/rifx"

func archStageOf(name string) string {
	for _, s := range []string{"scene", "codec", "parse", "lower", "write", "back", "mutate"} {
		if strings.HasPrefix(name, s+"_") {
			return s
		}
	}
	return ""
}

// sceneRifxWhitelist names scene_ files still permitted to import rifx.
// Each entry is a known boundary violation pending removal; the guard fails
// for any scene_ file NOT listed here that imports rifx.
var sceneRifxWhitelist = map[string]bool{
	// filled from this guard's first run after the scene_ rename
}

// sceneTypeNames are the runtime types a codec_ file must never reference —
// codec_ handles only value objects, byte streams, and rifx structures.
var sceneTypeNames = map[string]bool{
	"Project": true, "Composition": true, "Layer": true, "ShapeLayer": true,
	"Property": true, "ShapeNode": true, "VectorGroup": true, "Footage": true,
	"LayerTransform": true, "RectNode": true, "EllipseNode": true, "PathNode": true,
	"FillNode": true, "StrokeNode": true,
}

func srcFiles(t *testing.T) []string {
	t.Helper()
	all, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, f := range all {
		if !strings.HasSuffix(f, "_test.go") {
			out = append(out, f)
		}
	}
	return out
}

func TestArchBoundary_SceneNoRifxImport(t *testing.T) {
	fset := token.NewFileSet()
	for _, f := range srcFiles(t) {
		if archStageOf(f) != "scene" {
			continue
		}
		af, err := parser.ParseFile(fset, f, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		for _, imp := range af.Imports {
			if strings.Trim(imp.Path.Value, `"`) == rifxImportPath {
				if sceneRifxWhitelist[f] {
					t.Logf("WHITELIST: %s imports rifx — pending removal", f)
					continue
				}
				t.Errorf("scene_ file %s imports rifx: scene model must stay chunk-free", f)
			}
		}
	}
}

func TestArchBoundary_CodecNoSceneRef(t *testing.T) {
	fset := token.NewFileSet()
	for _, f := range srcFiles(t) {
		if archStageOf(f) != "codec" {
			continue
		}
		af, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		ast.Inspect(af, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if ok && sceneTypeNames[id.Name] {
				t.Errorf("codec_ file %s references scene type %q: codec must stay scene-free", f, id.Name)
			}
			return true
		})
	}
}
```

- [ ] **Step 2: 首跑 —— 预期 scene 测试失败、列出违例**

```powershell
go test -count=1 ./internal/aep/ -run TestArchBoundary -v
```
预期：`TestArchBoundary_SceneNoRifxImport` **此时无 scene_ 文件**（还没改名）→ 0 违例 PASS；`TestArchBoundary_CodecNoSceneRef` 同理 0 文件 PASS。**真正的枚举发生在 P3/P1 改名后**——届时重跑此测试，失败输出即权威白名单。先确保守卫本身编译/运行正确。

- [ ] **Step 3: commit**

```powershell
git add internal/aep/arch_boundary_test.go
git commit -m "test(reorg): AST arch-boundary guard (scene_ no-rifx-import, codec_ no-scene-ref); empty whitelist"
```

---

## Task 3: 改名 `codec_*`（5 文件）+ property_stream 纯度定夺

**Files (git mv):**
- `framerate_canonical.go` → `codec_framerate.go`
- `postscript.go` → `codec_postscript.go`
- `gradient.go` → `codec_gradient.go`
- `cdta_layout.go` → `codec_cdta_layout.go`
- `ldta_layout.go` → `codec_ldta_layout.go`

- [ ] **Step 1: property_stream 纯度扫描**

```powershell
go run ./tmp_debug/reorg_baseline > $null   # 确保基线在
Select-String -Path internal/aep/property_stream.go -Pattern 'Project|Composition|\bLayer\b|\bProperty\b|ShapeNode|Footage|rifx'
```
判定：**无匹配 → 归 `codec_property_stream.go`**（纯泛型 IR，方案②可搬）；**有匹配 → 归 `scene_property_stream.go`**（P3 处理）。把结论记入本任务 commit message。

- [ ] **Step 2: git mv 5 个确定的 codec_ 文件**（+ 若 Step1 判纯则 property_stream）

```powershell
cd internal/aep
git mv framerate_canonical.go codec_framerate.go
git mv postscript.go          codec_postscript.go
git mv gradient.go            codec_gradient.go
git mv cdta_layout.go         codec_cdta_layout.go
git mv ldta_layout.go         codec_ldta_layout.go
cd ../..
```

- [ ] **Step 3: 跑边界守卫（codec 现有文件，须仍 0 违例）**

```powershell
go test -count=1 ./internal/aep/ -run TestArchBoundary_CodecNoSceneRef -v
```
预期 PASS。**若失败**：某 codec_ 文件引用了 scene 类型 → 回退该文件为 `scene_`（spec §3 注），并记录。

- [ ] **Step 4: 全局 Gate**

跑「全局验证门」三步。预期全绿、API 零 diff、baseline PASS。

- [ ] **Step 5: commit**

```powershell
git commit -m "refactor(aep): rename pure leaf encoders to codec_* (framerate/postscript/gradient/cdta/ldta layout)"
```

---

## Task 4: 改名 `mutate_*`（10 文件，含 move 合并）

**Files (git mv):**
```
new_project.go            → mutate_project_new.go
new_composition.go        → mutate_composition_new.go
new_layer.go              → mutate_layer_new.go
delete_layer.go           → mutate_layer_delete.go
insert_layer.go           → mutate_layer_insert.go
duplicate_layer.go        → mutate_layer_duplicate.go
duplicate_composition.go  → mutate_composition_duplicate.go
sync_shape_layers.go      → mutate_shape_sync.go
import_closure.go         → mutate_import_closure.go
move_layer.go             → mutate_layer_move.go   (主)
layer_move.go             → 内容并入 mutate_layer_move.go 后删除
```

- [ ] **Step 1: git mv 9 个 + move 主文件**

```powershell
cd internal/aep
git mv new_project.go mutate_project_new.go
git mv new_composition.go mutate_composition_new.go
git mv new_layer.go mutate_layer_new.go
git mv delete_layer.go mutate_layer_delete.go
git mv insert_layer.go mutate_layer_insert.go
git mv duplicate_layer.go mutate_layer_duplicate.go
git mv duplicate_composition.go mutate_composition_duplicate.go
git mv sync_shape_layers.go mutate_shape_sync.go
git mv import_closure.go mutate_import_closure.go
git mv move_layer.go mutate_layer_move.go
cd ../..
```

- [ ] **Step 2: 合并 layer_move.go 进 mutate_layer_move.go**

把 `layer_move.go` 的全部函数（`(l *Layer) MoveToBeginning/MoveToEnd/MoveAfter/MoveBefore`、`locateInComp`、`locatePair`）逐字剪切到 `mutate_layer_move.go` 末尾（已确认与 `MoveLayer`/`indexOfChunk` 零函数名冲突，spec §3），合并 import 块去重，删除 `layer_move.go`：

```powershell
git rm internal/aep/layer_move.go   # 内容已手工并入 mutate_layer_move.go
```

- [ ] **Step 3: 全局 Gate**

跑三步。预期全绿、API 零 diff、baseline PASS。

- [ ] **Step 4: commit**

```powershell
git commit -m "refactor(aep): unify structural ops under mutate_* (new/delete/insert/move/duplicate/sync/import); merge layer_move into mutate_layer_move"
```

---

## Task 5: 其它阶段改名（json → write_, hydrate → parse_）

**Files (git mv):**
- `json.go` → `write_json.go`
- `hydrate_shape.go` → `parse_shape_hydrate.go`

- [ ] **Step 1: git mv**

```powershell
cd internal/aep
git mv json.go write_json.go
git mv hydrate_shape.go parse_shape_hydrate.go
cd ../..
```

- [ ] **Step 2: 全局 Gate**

跑三步。预期全绿、API 零 diff、baseline PASS。

- [ ] **Step 3: commit**

```powershell
git commit -m "refactor(aep): json->write_json (serialize stage), hydrate_shape->parse_shape_hydrate (read stage)"
```

---

## Task 6: 改名 `scene_*`（accessor/model 批量，~20 文件）

**Files (git mv):**
```
types_features.go         → scene_features.go
layer_accessors.go        → scene_layer_accessors.go      (Task 7 再拆)
layer_convenience.go      → scene_layer_convenience.go
layer_matte.go            → scene_layer_matte.go
layer_property_groups.go  → scene_layer_property_groups.go
frame_time_accessors.go   → scene_frame_time.go
composition_convenience.go→ scene_composition_convenience.go
composition_views.go      → scene_composition_views.go
project_views.go          → scene_project_views.go
project_settings.go       → scene_project_settings.go
footage_convenience.go    → scene_footage_convenience.go
footage_source.go         → scene_footage_source.go
property_group.go         → scene_property_group.go
property_defaults.go      → scene_property_defaults.go
property_flags.go         → scene_property_flags.go
property_units.go         → scene_property_units.go
shape_graph.go            → scene_shape_graph.go
text_types.go             → scene_text.go
application.go            → scene_application.go
capability_matrix.go      → scene_capability.go
```
（`property_stream.go` 若 Task 3 判非纯，此处一并 → `scene_property_stream.go`；`types_core.go` 留待 Task 7 拆分时改名。）

- [ ] **Step 1: git mv 全部**

```powershell
cd internal/aep
git mv types_features.go scene_features.go
git mv layer_accessors.go scene_layer_accessors.go
git mv layer_convenience.go scene_layer_convenience.go
git mv layer_matte.go scene_layer_matte.go
git mv layer_property_groups.go scene_layer_property_groups.go
git mv frame_time_accessors.go scene_frame_time.go
git mv composition_convenience.go scene_composition_convenience.go
git mv composition_views.go scene_composition_views.go
git mv project_views.go scene_project_views.go
git mv project_settings.go scene_project_settings.go
git mv footage_convenience.go scene_footage_convenience.go
git mv footage_source.go scene_footage_source.go
git mv property_group.go scene_property_group.go
git mv property_defaults.go scene_property_defaults.go
git mv property_flags.go scene_property_flags.go
git mv property_units.go scene_property_units.go
git mv shape_graph.go scene_shape_graph.go
git mv text_types.go scene_text.go
git mv application.go scene_application.go
git mv capability_matrix.go scene_capability.go
cd ../..
```

- [ ] **Step 2: 跑边界守卫 —— 此时枚举 scene_ 的 rifx-import 违例（权威白名单）**

```powershell
go test -count=1 ./internal/aep/ -run TestArchBoundary_SceneNoRifxImport -v
```
预期：**失败**，列出真正 import rifx 的 scene_ 文件（已知候选 `scene_features` / `scene_project_settings` / `scene_property_group` / `scene_property_flags`，可能含 `scene_project_views`）。**把这些文件名填进 `sceneRifxWhitelist`**（Task 2 的 map），再跑一次 → PASS（带 WHITELIST 日志）。

- [ ] **Step 3: 提交白名单**

```powershell
# 编辑 arch_boundary_test.go 的 sceneRifxWhitelist 填入 Step2 输出的文件
go test -count=1 ./internal/aep/ -run TestArchBoundary_SceneNoRifxImport -v   # PASS
```

- [ ] **Step 4: 全局 Gate + commit**

```powershell
git add -A
git commit -m "refactor(aep): rename model/accessor files to scene_*; populate arch-guard rifx whitelist (cleared in P6)"
```

---

## Task 7: 拆 `types_core.go` → 4 个 scene_ 文件（按类型名）

**Files:**
- Create: `internal/aep/scene_project.go`（`type Project` + 其方法/相关常量）
- Create: `internal/aep/scene_composition.go`（`type Composition` + 方法）
- Create: `internal/aep/scene_layer.go`（`type Layer` + 方法）
- Create: `internal/aep/scene_property.go`（`type Property` + 方法）
- Delete: `internal/aep/types_core.go`

边界**按类型名**（非行号）：`type Project` / `type Composition` / `type Layer` / `type Property` 各自连同紧随其后、明显属于该类型的方法与私有常量。

- [ ] **Step 1: 确认四个类型起止**

```powershell
Select-String -Path internal/aep/types_core.go -Pattern '^type (Project|Composition|Layer|Property) struct'
```

- [ ] **Step 2: 建 4 个新文件，逐类型剪切**

每个新文件以 `package aep` 开头 + **仅该文件实际用到的 import**（编译器会报多余/缺失 import，据此修正）。把 `type X struct {...}` 及其后属于 X 的 `func (x *X) ...`、X 专属私有 helper/常量从 `types_core.go` 剪到对应文件。文件级共享的 import/常量（如多个类型共用的）留一份在引用最集中的文件或保留一个 `scene_core.go` 承载共享件。

- [ ] **Step 3: 删除空壳 types_core.go**

```powershell
git rm internal/aep/types_core.go   # 内容已全部迁出
```

- [ ] **Step 4: 全局 Gate**

跑三步。**重点**：API 零 diff（类型/方法没变，只是换文件）、baseline PASS（行为不变）。守卫 `TestArchBoundary_SceneNoRifxImport` 须 PASS（`scene_layer.go` 持有 `back *layerBackrefs` 字段不触发 rifx import；若触发说明误带了 rifx 引用，移回 back_ 或入白名单并记 P6）。

- [ ] **Step 5: commit**

```powershell
git add -A
git commit -m "refactor(aep): split types_core.go into scene_{project,composition,layer,property}.go (by type)"
```

---

## Task 8: 拆 `scene_layer_accessors.go`（946 行 → 2 文件）

**Files:**
- Keep: `internal/aep/scene_layer_accessors.go`（transform / timing accessors）
- Create: `internal/aep/scene_layer_property_access.go`（property 树导航 accessors）
- Test: 无新测试（纯移动）

- [ ] **Step 1: 分组识别**

```powershell
Select-String -Path internal/aep/scene_layer_accessors.go -Pattern '^func '
```
把「property 树导航」类方法（按名称/职责：返回 `*Property`/`*PropertyGroup`、遍历 property stream 的）划入新文件，其余（transform/timing/几何 accessor）留原文件。目标各 < ~450 行。

- [ ] **Step 2: 建新文件、剪切对应方法**

`scene_layer_property_access.go`：`package aep` + 实际 import + 剪入的方法。

- [ ] **Step 3: 全局 Gate**

跑三步。API 零 diff、baseline PASS。

- [ ] **Step 4: commit**

```powershell
git add -A
git commit -m "refactor(aep): split scene_layer_accessors.go (transform/timing | property-tree navigation)"
```

---

## Task 9: 测试基建归并 + helper 收敛

**Files:**
- Create: `internal/aep/testutil_fixtures_test.go`（fixture 路径常量 + loader）
- Modify: `internal/aep/testhelpers_test.go` → 内容并入 `testutil_*`，删除
- Modify: `internal/aep/ship_gate_helpers_test.go` → 内容并入 `testutil_*`，删除

- [ ] **Step 1: 盘点重复 helper**

```powershell
Select-String -Path internal/aep/testhelpers_test.go,internal/aep/ship_gate_helpers_test.go,internal/aep/testutil_*_test.go -Pattern '^func '
```
找出重复定义 / 同名 helper，确定唯一归属。

- [ ] **Step 2: 集中 fixture loader**

`testutil_fixtures_test.go`：把散落的 `test_data/...` 路径字面量收敛为常量 + `loadFixture(t, name)` helper。用例改引常量（**只改测试，不动源码**）。

- [ ] **Step 3: 合并 helper 文件**

把 `testhelpers_test.go` / `ship_gate_helpers_test.go` 的 helper 迁入对应 `testutil_<domain>_test.go`，删两文件。

- [ ] **Step 4: 全局 Gate（含全测试编译）**

```powershell
go test -count=1 ./internal/aep/   # 全绿（测试重组不改源，API/baseline 不受影响）
```

- [ ] **Step 5: commit**

```powershell
git add -A
git commit -m "test(aep): converge test helpers into testutil_*; central fixture loader"
```

---

## Task 10: 拆测试巨文件（feature-first 命名）

**Files (拆分，feature-first 允许，spec §5):**
- `layer_test.go`(1789) → 按 feature：如 `layer_transform_test.go` / `layer_matte_test.go` / `layer_blend_test.go` …
- `shape_layer_shipgate_test.go`(1376) → 按 primitive/子项：`shape_rect_shipgate_test.go` / `shape_ellipse_shipgate_test.go` / `shape_stroke_shipgate_test.go` …
- `text_test.go`(1105) → `text_runs_test.go` / `text_animator_test.go` …
- `composition_test.go`(777) / `keyframe_test.go`(645) → 按 feature 拆，目标 < ~600 行

- [ ] **Step 1: 按 `func Test*` 分组**

```powershell
Select-String -Path internal/aep/layer_test.go -Pattern '^func Test'
```
对每个巨文件，按测试主题归组到 feature-first 新文件。

- [ ] **Step 2: 逐文件剪切（一次一个巨文件，单独验证）**

每拆完一个巨文件即 `go test -count=1 ./internal/aep/` 全绿后再拆下一个。ship-gate 测试保持 `AE_SHIP_GATE` 门控 + `-run` pattern 不变。

- [ ] **Step 3: commit（每个巨文件一个 commit）**

```powershell
git add -A
git commit -m "test(aep): split layer_test.go by feature (transform/matte/blend/...)"
# 重复 for shape/text/composition/keyframe
```

---

## Task 11: P6 收口 —— 白名单清零

**Files:**
- Modify: `internal/aep/arch_boundary_test.go`（清空/缩小 `sceneRifxWhitelist`）
- Modify: 残留 import rifx 的 `scene_*.go`（把 rifx 引用迁出，或确认无法迁出）

- [ ] **Step 1: 逐个白名单文件尝试去 rifx**

对 `sceneRifxWhitelist` 每个文件：若其 rifx 用法是 `rifx.ChunkID`/`IDxxx` 常量比较等纯导航 → 评估能否移到 `parse_`/`write_`/`back_` 文件（按 §2.1 矩阵）。能移则移、删白名单项；移完跑守卫确认。

- [ ] **Step 2: 无法清零项 → 记录 + 转 V3 spec**

清不掉的文件：在 `arch_boundary_test.go` 白名单项旁加注释写明技术原因；在 `specs/2026-05-22-v3-direction.md` 追加「方案②前置解耦项」一节列出。

- [ ] **Step 3: 守卫 + 全局 Gate**

```powershell
go test -count=1 ./internal/aep/ -run TestArchBoundary -v   # PASS（白名单尽量空）
```

- [ ] **Step 4: commit**

```powershell
git add -A
git commit -m "refactor(aep): clear scene_ rifx-import whitelist where feasible; escalate residuals to V3 spec"
```

---

## Task 12: P6 收口 —— CLAUDE.md + 终验 + landing

**Files:**
- Modify: `CLAUDE.md`（硬约束 #3，按 spec §10）
- Modify: `flightdeck/cockpit.md`、`flightdeck/manifest.md`

- [ ] **Step 1: 更新 CLAUDE.md 硬约束 #3**

按 spec §10：保留「单 package」结论但理由改为 Go 方法同包 + Stable API（非 re-export 噪音）；命名轴改 §2 的 7 前缀；注明 serialize = lower_+write_（附触发文案）、codec_ / back_ 含义 + 守卫；指向本 spec + V3 spec。

- [ ] **Step 2: 终验 —— AE 双版本 ship-gate（自验留痕，无需人工）**

```powershell
$env:AE_SHIP_GATE=1
go test -count=1 ./internal/aep/ -run 'TestV2_2_.*AEShipGate' -v
```
预期全 PASS（纯文件移动，零回归）。AE 2025 冷启偶发 exit 2 → warm retry（先 kill AfterFX + 跑 known-good fixture 热身），见 cockpit 自验留痕。

- [ ] **Step 3: 删除重组期工件**

```powershell
Remove-Item -Recurse -Force tmp_debug/reorg_baseline
git rm internal/aep/reorg_baseline_test.go
```
（`arch_boundary_test.go` **保留**——它是 §6 的常驻护栏。）

- [ ] **Step 4: 更新 cockpit + manifest，commit**

cockpit：`Last updated` + Active focus 改为「aep package 重组已 ship」；manifest 无 in-flight。

```powershell
git add -A
git commit -m "docs: CLAUDE.md #3 update (Go-method rationale + 7-prefix axis); reorg landed; remove reorg baseline harness"
```

- [ ] **Step 5: finishing-a-development-branch**

调用 `superpowers:finishing-a-development-branch` 决定合并/PR。

---

## Self-Review（写完核对 spec）

- **§3 映射全覆盖**：codec_(T3)、mutate_(T4)、json/hydrate(T5)、scene_(T6)、types_core 拆(T7)、layer_accessors 拆(T8) —— 70 文件去向均有任务。✓
- **§4 大文件**：types_core(T7)、layer_accessors(T8)、测试巨文件(T10)。✓
- **§5 测试基建**：T9（helper 收敛 + fixture 集中）+ T10（feature-first 拆分）。✓
- **§6 守卫**：T2 建（AST、空白名单）→ T6 枚举填白名单 → T11 清零。✓
- **§8 验证**：T0 API 基线 + T1 round-trip 基线 + 每任务 Gate。✓
- **§9 阶段**：P1(T1-T3)、P2(T4)、P3(T5-T6)、P4(T7-T8)、P5(T9-T10)、P6(T11-T12) 对齐。✓
- **§10 CLAUDE.md**：T12 Step1。✓
- **类型一致性**：`sceneRifxWhitelist`/`sceneTypeNames`/`archStageOf` 跨 T2/T6/T11 同名一致。`TestReorgRoundtripBaseline` 跨 T1/Gate 一致。✓
- **占位符扫描**：无 TBD；两处「实现时定」均给判据（property_stream 纯度扫描 T3S1；白名单由守卫枚举 T6S2）。✓
