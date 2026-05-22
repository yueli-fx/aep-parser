# V2.1 Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让用户能 `aep.NewProject(target ...AETarget) *Project` + `proj.NewComposition(name, w, h, fps, duration) (*Composition, error)`，生成的 .aep 用 AE 2020 / 2022 / 2025 三个版本都能打开。

**Architecture:** 嵌入 3 份 AE-saved empty `.aep` 当 template；新 Composition 通过合成 chunks → append 到根 Fold LIST → **重 parseComposition 闭环** 出 `*Composition`（保证 New / Open 同构）。WriteAEP 不改（既有 RIFX 序列化自动重算父 LIST size）。

**Tech Stack:** Go 1.24+，`//go:embed`，既有 `internal/rifx` + `internal/aep`，AE ScriptingAPI (ExtendScript) for ship gate。

**Spec:** [`../specs/v2-1-foundation-design.md`](../specs/v2-1-foundation-design.md)

---

## 文件清单

**新建**：
- `internal/aep/new_project.go` — `aep.NewProject` + `AETarget` enum + embed
- `internal/aep/new_composition.go` — `Project.NewComposition` + chunk builders
- `internal/aep/cdta_layout.go` — cdta / idta byte 偏移常量集中表
- `internal/aep/framerate_canonical.go` — NTSC canonical fps 表
- `internal/aep/new_project_test.go` — NewProject tests
- `internal/aep/new_composition_test.go` — NewComposition tests
- `internal/aep/framerate_canonical_test.go` — canonical round-trip tests
- `test_data/verify_v2_1.jsx` — AE ship gate JSX
- `test_data/comp_item_children.golden.txt` — RE-3 产物 golden fixture
- `tmp_debug/dump_fdta/main.go` — RE-1 探针
- `workshop/scars/<topic>.md` — RE 阶段产生的 negative findings（按需）

**已存在（不动）**：
- `internal/aep/templates/{2020,2022,2025}.aep` ✅

**修改**：
- `internal/aep/types_core.go` — `Project` 加 `nextItemID uint32` + `rootFold *rifx.Chunk` 非导出字段
- `internal/aep/parse.go` — `parseProject` 末尾调 `initDerived(rifxRoot)`
- `internal/aep/write_composition.go` — 散落 hex 偏移迁到 `cdta_layout.go` 常量

---

# Phase 0 — RE Prerequisites（投入 code 前必须解决）

Phase 0 全部是 investigation tasks。产物是 **plan 内嵌的 findings** + （按需）`workshop/scars/*.md`。完成后 plan 自身更新（task 0.X 标注 `[finding]: ...`），后续 Phase 1+ 按 finding 实现。

### Task 0.1: RE-1 — `fdta` 字节语义 dump 矩阵

**Files:**
- Use: `internal/aep/templates/2020.aep`, `2022.aep`, `2025.aep`
- Create: `tmp_debug/dump_fdta/main.go`
- Create: `test_data/fdta_probe_AE2020.jsx`（用 AE 2020 saved 1/2 comp + delete comp + nested folder）
- Create: `test_data/fdta_probe_AE2025.jsx`（同上用 AE 2025）

- [ ] **Step 1: 写 dump_fdta 工具**

```go
// tmp_debug/dump_fdta/main.go
package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: dump_fdta file.aep")
		os.Exit(2)
	}
	f, _ := os.Open(os.Args[1])
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil { panic(err) }

	fold := root.FindFirstList(rifx.IDFold)
	if fold == nil { fmt.Println("no Fold LIST"); return }
	for _, ch := range fold.Children {
		if string(ch.ID[:]) == "fdta" {
			fmt.Printf("fdta (%d B) hex=%s\n", len(ch.Data), hex.EncodeToString(ch.Data))
			// item count under same Fold
			items := 0
			for _, sib := range fold.Children {
				if sib.IsList() && sib.FormType == rifx.IDItem { items++ }
			}
			fmt.Printf("  sibling Item count: %d\n", items)
			return
		}
	}
}
```

- [ ] **Step 2: 创建 dir 并跑 baseline dump**

```bash
mkdir -p tmp_debug/dump_fdta
# (创建上面的 main.go 后)
go run ./tmp_debug/dump_fdta internal/aep/templates/2020.aep
go run ./tmp_debug/dump_fdta internal/aep/templates/2022.aep
go run ./tmp_debug/dump_fdta internal/aep/templates/2025.aep
```

记录三个 hex 输出到下面 "RE-1 findings" 段。

- [ ] **Step 3: 写 fdta_probe_AE2025.jsx（5 个场景）**

```jsx
(function () {
    var outDir = "e:/projects/tools/aep-parser/test_data/fdta_probe/";
    var log = [];
    function save(name) {
        var f = new File(outDir + name);
        app.project.save(f);
        log.push("saved " + name);
    }
    function fresh() {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
    }

    fresh();
    save("AE2025_empty.aep");

    fresh();
    app.project.items.addComp("C1", 1920, 1080, 1, 5, 30);
    save("AE2025_1comp.aep");

    fresh();
    app.project.items.addComp("C1", 1920, 1080, 1, 5, 30);
    app.project.items.addComp("C2", 1920, 1080, 1, 5, 30);
    save("AE2025_2comp.aep");

    fresh();
    var c1 = app.project.items.addComp("C1", 1920, 1080, 1, 5, 30);
    var c2 = app.project.items.addComp("C2", 1920, 1080, 1, 5, 30);
    c1.remove();
    save("AE2025_1comp_after_delete.aep");

    fresh();
    var folder = app.project.items.addFolder("F1");
    app.project.items.addComp("InsideC", 1920, 1080, 1, 5, 30).parentFolder = folder;
    save("AE2025_nested.aep");

    var marker = new File("e:/projects/tools/aep-parser/test_data/fdta_probe_AE2025.done");
    marker.open("w"); marker.write(log.join("\n")); marker.close();
})();
```

- [ ] **Step 4: 跑 AE 2025 + AE 2020 各一遍**

```bash
mkdir -p test_data/fdta_probe
rm -f test_data/fdta_probe_AE2025.done
"E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe" -r "E:/projects/tools/aep-parser/test_data/fdta_probe_AE2025.jsx"
until [ -f test_data/fdta_probe_AE2025.done ]; do sleep 2; done

# 重复 AE 2020（改 outDir suffix 为 _AE2020 / 改 path / 改 marker）
```

- [ ] **Step 5: dump 五种状态 fdta + 逐字节 diff**

```bash
for v in AE2025 AE2020; do
  for s in empty 1comp 2comp 1comp_after_delete nested; do
    echo "=== ${v}_${s} ==="
    go run ./tmp_debug/dump_fdta "test_data/fdta_probe/${v}_${s}.aep"
  done
done
```

把每行 hex 抄到 plan 里下面的 "RE-1 findings" 段。

- [ ] **Step 6: 分类字段**

对照五种状态的 fdta 字节，按字节位置分：
- **monotonic**（每加 1 个 item 自增）：填 → 视为 item count / next-id allocator
- **topology-sensitive**（删 / nest 时变）：填 → 视为 layout pointer
- **static**（永远不变 / 不依赖 topology）：填 → 安全忽略

- [ ] **Step 7: 写 finding 到本 plan 文件**

把结论 + 证据加到本 Task 末尾的 "**[finding]:**" 段。如果有 monotonic / topology 字段，**追加** Phase 4 子任务（updateFdtaOnAppend）。如果无害，文档记录"builder 不动 fdta"。

- [ ] **Step 8: Commit**

```bash
git add tmp_debug/dump_fdta test_data/fdta_probe_AE*.jsx test_data/fdta_probe/ workshop/plans/v2-1-foundation-plan.md
git commit -m "re(v2): dump fdta semantics across AE 2020/2025 × 5 topologies"
```

**[finding] — fdta 完全无害，builder 不需要 updateFdtaOnAppend**

**Dump 矩阵**（14B fdta hex；2 AE 版本 × 5 拓扑 = 10 行 + 3 baseline templates）：

```
=== baseline templates (saved by AE during normal use) ===
templates/2020.aep   00 00 00 00 00 00 00 00 00 00 f4 22 00 00
templates/2022.aep   00 00 00 00 00 00 00 00 00 00 da 00 00 00
templates/2025.aep   00 00 00 00 00 00 00 00 00 00 00 00 00 00

=== AE 2020 probe (5 topologies; all identical to each other) ===
empty / 1comp / 2comp / 1comp_after_delete / nested
                      00 00 00 00 00 00 00 00 00 00 62 b0 00 00

=== AE 2025 probe (5 topologies; all identical, match template) ===
empty / 1comp / 2comp / 1comp_after_delete / nested
                      00 00 00 00 00 00 00 00 00 00 00 00 00 00
```

**字节级分类**：

| 范围 | 跨拓扑 | 跨版本 | 跨 session | 解释 |
| --- | --- | --- | --- | --- |
| @0x00..@0x09 (10B) | 不变 | 不变 | 不变 | reserved padding |
| @0x0A..@0x0B (2B) | **不变** | 不同（2020 `f4 22` 或 `62 b0` / 2022 `da 00` / 2025 `00 00`） | **不同**（template 2020 `f4 22` vs probe 2020 `62 b0`） | opaque session-state token |
| @0x0C..@0x0D (2B) | 不变 | 不变 | 不变 | reserved padding |

**关键**：
1. 同 AE 版本下 fdta 100% 静态 — empty / 1 comp / 2 comp / delete / nested 全 byte-identical
2. `@0x0A..@0x0B` 不是 child count / next-id / topology pointer，是跨 session 漂移的 opaque marker
3. AE 2025 干脆全写零，证明语义上 AE 自己也不依赖

**结论 — builder 不动 fdta**：
- 嵌入 template 的 fdta 字节原样保留
- NewComposition append 新 Item 后**不**重写 fdta
- 不追加 `updateFdtaOnAppend` 子任务，Phase 4/5 不受影响

---

### Task 0.2: RE-2 — `idpc` 是 per-item key？

**Files:**
- Use: `tmp_debug/list_item_chunks/`（已有）+ Task 0.1 生成的 fixtures
- Create: `test_data/idpc_collision.jsx`（故意写两 comp 同 idpc 看 AE 行为）

- [ ] **Step 1: dump 5 个 fixture 的所有 comp Item LIST 的 idpc**

```bash
for f in test_data/fdta_probe/AE2025_*.aep; do
  echo "=== $f ==="
  go run ./tmp_debug/list_item_chunks "$f" | grep -A 1 'chunk idpc'
done
```

记录 idpc 字节长度 + 值 + 是否全 0 / 是否唯一。

- [ ] **Step 2: 写 idpc finding**

预期可能：
- (a) 8 字节，AE 2025 默认全 0，但 1+ comp 时 AE 填唯一值 → idpc 是 UUID
- (b) 8 字节，永远全 0 → AE 重写时填，builder 可填 0
- (c) 16 字节，超出我们 hex dump 看到的 → check 用 `xxd` 验证

- [ ] **Step 3: 写 idpc_collision.jsx — 跨 fixture 验证唯一性**

由于我们暂无能力直接 hex-edit binary aep 然后让 AE 加载，这步通过对比已 saved fixture 间 idpc 是否不同 来验证唯一性。

```bash
go run ./tmp_debug/list_item_chunks test_data/fdta_probe/AE2025_2comp.aep | grep 'chunk idpc'
# 应输出 2 个不同的 idpc hex
```

- [ ] **Step 4: 决策 — `crypto/rand` 默认**

按 spec RE-2，default = `crypto/rand` v4 UUID（safest）。即使 dump 显示 AE 默认全 0，我们也用随机值（zero downside）。

写 finding 到下面 "[finding]:" 段：
- idpc 长度 = N bytes（实测 dump 数字）
- 默认实现 = `crypto/rand.Read(n)`
- AE 2020 / 2025 是否接受不同 idpc 行为（待 ship gate 实测）

- [ ] **Step 5: Commit**

```bash
git add test_data/idpc_collision.jsx workshop/plans/v2-1-foundation-plan.md
git commit -m "re(v2): idpc semantics — use crypto/rand for safety"
```

**[finding output slot]** _(populated by engineer after running this Task) — Record dumped bytes, classify monotonic / topology-sensitive / static fields, and decide whether builder needs `updateFdtaOnAppend` step (RE-1) / which idpc length + generation strategy (RE-2) / final idta 84-byte layout (RE-3) / WorkArea sentinel + frame rate canonical table values (RE-4) per task scope above._

---

### Task 0.3: RE-3 — `idta` 完整 84-byte layout + golden fixture

**Files:**
- Use: `tmp_debug/list_item_chunks/`
- Create: `test_data/comp_item_children.golden.txt`

- [ ] **Step 1: dump AE 2025 saved `2comp.aep` 的 Item LIST 全 children**

```bash
go run ./tmp_debug/list_item_chunks test_data/fdta_probe/AE2025_1comp.aep > /tmp/item_dump.txt
cat /tmp/item_dump.txt
```

记录：iide / idpc / idta / Utf8 / LIST(dats) / cdta / cdrp / comr / LIST(PRin) / LIST(DLay) ... LIST(Layr) 的**完整顺序 + 每个 chunk 字节长度**。

- [ ] **Step 2: 落 golden fixture**

```
# test_data/comp_item_children.golden.txt
# Output of `go run ./tmp_debug/list_item_chunks test_data/fdta_probe/AE2025_1comp.aep`
# (the first comp's Item LIST). Used by TestBuildCompItemGoldenMatch
# to assert builder output matches exact AE-saved structure.

Item LIST children order (AE 2025, 1 comp):
  [0] iide (4 B)
  [1] idpc (8 B)
  [2] idta (84 B)
  [3] Utf8 (N B, name)
  [4] LIST dats (1 child: numS)
  [5] cdta (204 B)
  [6] cdrp (1 B)
  [7] comr (1 B)
  [8] LIST PRin (2 children: prin + prda)
  [9] LIST DLay (...)
  ...
  [N] LIST Layr (M children)
```

实际填数据从 step 1 的 dump 输出抄。

- [ ] **Step 3: dump idta 84 字节 hex 完整内容**

```bash
go run ./tmp_debug/list_item_chunks test_data/fdta_probe/AE2025_1comp.aep | grep -A 1 'chunk idta'
```

把 hex 抄进 finding。识别哪些字节是 type / ID / label，哪些是固定默认值。

- [ ] **Step 4: 完成 idta layout 表**

把结论写到下面 "[finding]:" 段：
- `@0x00` uint16 BE type code (0x04 = comp，0x07 = footage etc.)
- `@0x14` uint32 BE Item ID（待验证：dump 多个 comp 看 idta @0x14 vs 实际 ID 是否一致）
- `@0x3A` uint8 label color（0 = none）
- 其余 75 字节：建议直接抄 AE 2025 fixture 的 idta 字节作 builder 默认（template-copy 策略），不一字节一字节 RE

- [ ] **Step 5: Commit**

```bash
git add test_data/comp_item_children.golden.txt workshop/plans/v2-1-foundation-plan.md
git commit -m "re(v2): idta layout + Item children golden fixture"
```

**[finding output slot]** _(populated by engineer after running this Task) — Record dumped bytes, classify monotonic / topology-sensitive / static fields, and decide whether builder needs `updateFdtaOnAppend` step (RE-1) / which idpc length + generation strategy (RE-2) / final idta 84-byte layout (RE-3) / WorkArea sentinel + frame rate canonical table values (RE-4) per task scope above._

---

### Task 0.4: RE-4 — WorkArea sentinel + frame rate canonical

**Files:**
- Use: `tmp_debug/dump_cdta/`（已有）

- [ ] **Step 1: dump 5 个 fixture 的 cdta @0x1C..@0x2C（work area 区段）**

```bash
for f in test_data/fdta_probe/AE2025_*.aep; do
  echo "=== $f ==="
  go run ./tmp_debug/dump_cdta "$f" | head -10
done
```

观察 `@0x1C..@0x1F`（WorkAreaStart dividend）/ `@0x20..@0x23`（divisor）/ `@0x24..@0x27`（end dividend）/ `@0x28..@0x2B`（end divisor）。

特别注意 end dividend 是 `0xFFFFFFFF` sentinel 还是 duration 真值。

- [ ] **Step 2: 写 WorkArea finding**

- AE 默认 work area = (sentinel / duration / 其它)
- 我们 `buildCompCdta` 写：
  - WorkAreaStart = 0/600（duration unit = 600 ticks/sec）
  - WorkAreaEnd = ? （结论 sentinel `0xFFFFFFFF` 还是 duration_frames * tickbase / fps）

- [ ] **Step 3: 用 既有 fixture 测 frame rate round-trip**

```bash
# 现有 re_tickrate.aep 含多 fps comps
go run ./tmp_debug/dump_cdta test_data/re_tickrate.aep | grep -E '^00000090|^=== '
```

观察 29.97 / 23.976 / 59.94 三个 comp 的 cdta `@0x9C..@0x9F`（whole + frac/65536）。

预期：
- 29.97 → whole=29, frac=0xF852 (= 63570/65536 = 0.97)
- 23.976 → whole=23, frac=0xF9DB (= 63963/65536 ≈ 0.976)
- 59.94 → whole=59, frac=0xF0A4 (= 61604/65536 ≈ 0.94)

记录精确 frac 值。

- [ ] **Step 4: 决策 canonical mapping**

写 finding：
- WorkArea：(builder 写 sentinel `0xFFFFFFFF` / 写真值 duration_frames)
- Canonical table：
  - `23.976 → whole=23, frac=<dumped frac>`
  - `29.97  → whole=29, frac=<dumped frac>`
  - `59.94  → whole=59, frac=<dumped frac>`
  - 用户输入容差 < 1e-3 内的 fps 都映射到这些 canonical 值
  - 其它 fps（24 / 25 / 30 / 60 etc.）走通用 `whole = floor(fps), frac = round((fps - whole) * 65536)`

- [ ] **Step 5: Commit**

```bash
git add workshop/plans/v2-1-foundation-plan.md
git commit -m "re(v2): WorkArea sentinel + NTSC frame rate canonical table"
```

**[finding output slot]** _(populated by engineer after running this Task) — Record dumped bytes, classify monotonic / topology-sensitive / static fields, and decide whether builder needs `updateFdtaOnAppend` step (RE-1) / which idpc length + generation strategy (RE-2) / final idta 84-byte layout (RE-3) / WorkArea sentinel + frame rate canonical table values (RE-4) per task scope above._

---

# Phase 1 — Foundation refactors（cdta_layout.go + framerate_canonical.go）

### Task 1.1: 建 `cdta_layout.go` 常量

**Files:**
- Create: `internal/aep/cdta_layout.go`

- [ ] **Step 1: 写常量文件**

```go
// internal/aep/cdta_layout.go
package aep

// Composition cdta byte offsets — single source of truth shared by
// parser (parse_composition.go), writer (write_composition.go), and
// builder (new_composition.go).
//
// cdta total size is 204 bytes (0xCC); layout is stable AE 2020 → AE 2025
// (bit-for-bit verified, see workshop/scars/).
const (
	cdtaResolutionFactorX = 0x00 // uint16 BE
	cdtaResolutionFactorY = 0x02 // uint16 BE
	cdtaTickRate          = 0x08 // uint32 BE
	cdtaWorkAreaStart     = 0x1C // uint32 BE dividend
	cdtaWorkAreaStartDiv  = 0x20 // uint32 BE divisor
	cdtaWorkAreaEnd       = 0x24 // uint32 BE dividend; 0xFFFFFFFF = sentinel
	cdtaWorkAreaEndDiv    = 0x28 // uint32 BE divisor
	cdtaBGColorR          = 0x34 // uint8
	cdtaBGColorG          = 0x35 // uint8
	cdtaBGColorB          = 0x36 // uint8
	cdtaFlagsByte8A       = 0x8A // bit 0 = Draft3D
	cdtaFlagsByte8B       = 0x8B // bit 0/3/4/5/7 = various comp flags
	cdtaWidth             = 0x8C // uint16 BE
	cdtaHeight            = 0x8E // uint16 BE
	cdtaPixelAspectNum    = 0x90 // uint32 BE numerator
	cdtaPixelAspectDen    = 0x94 // uint32 BE denominator
	cdtaFrameRateWhole    = 0x9C // uint16 BE whole part
	cdtaFrameRateFrac     = 0x9E // uint16 BE fractional part (frac / 65536)
	cdtaDisplayStartTime  = 0xA4 // uint32 BE dividend
	cdtaDisplayStartDiv   = 0xA8 // uint32 BE divisor
	cdtaShutterAngle      = 0xAE // uint16 BE
	cdtaDuration          = 0xB0 // uint32 BE frames
	cdtaShutterPhase      = 0xB4 // int32 BE
	cdtaMotionBlurAdaptive = 0xC4 // int32 BE
	cdtaMotionBlurSamples  = 0xC8 // int32 BE
	cdtaSize              = 0xCC // 总长 = 204
)

// Item idta byte offsets (84 bytes total).
// idta layout RE'd partially in workshop/board.md archives; remaining
// 75 bytes are template-copied as opaque defaults (see new_composition.go).
const (
	idtaTypeCode = 0x00 // uint16 BE; 0x04 = Composition
	idtaItemID   = 0x14 // uint32 BE; per-item ID (matches Composition.ID)
	idtaLabel    = 0x3A // uint8; label color index 0..16 (0 = none)
	idtaSize     = 84
)
```

- [ ] **Step 2: Build check**

```bash
go build ./internal/aep/
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/aep/cdta_layout.go
git commit -m "refactor(aep): extract cdta/idta byte offsets to cdta_layout.go"
```

---

### Task 1.2: 把 `write_composition.go` 现有 hex 偏移迁到常量

**Files:**
- Modify: `internal/aep/write_composition.go`

- [ ] **Step 1: 列现有 hex 偏移**

```bash
grep -nE '\b0x[0-9A-F]{2,4}\b' internal/aep/write_composition.go | head -30
```

记录每个偏移 + 对应的 cdta_layout.go 常量。

- [ ] **Step 2: 替换**

对照表（参考 spec Internal Design 的 cdta_layout.go 常量段）。例如：

```go
// before
binary.BigEndian.PutUint16(c.cdta.Data[0x8C:0x8E], width)
binary.BigEndian.PutUint16(c.cdta.Data[0x8E:0x90], height)

// after
binary.BigEndian.PutUint16(c.cdta.Data[cdtaWidth:cdtaWidth+2], width)
binary.BigEndian.PutUint16(c.cdta.Data[cdtaHeight:cdtaHeight+2], height)
```

- [ ] **Step 3: Build + run all tests**

```bash
go vet ./...
go test -count=1 ./internal/aep/...
```
Expected: vet clean, all existing tests PASS (109+).

- [ ] **Step 4: Commit**

```bash
git add internal/aep/write_composition.go
git commit -m "refactor(write_composition): migrate hex offsets to cdta_layout.go"
```

---

### Task 1.3: 建 `framerate_canonical.go` + canonical table

**Files:**
- Create: `internal/aep/framerate_canonical.go`

- [ ] **Step 1: 写 canonical 表（按 RE-4 finding 填具体 frac 值）**

```go
// internal/aep/framerate_canonical.go
package aep

import "math"

// frameRateEncoding 表示 cdta @0x9C..@0x9F 的 (whole, frac) 对编码。
// AE 存 frame rate 为 whole + frac/65536，但 NTSC fractions（23.976/29.97/59.94）
// 必须用 AE 内部 canonical 值，否则 AE 重新保存时 rewrite + 测试 flaky。
type frameRateEncoding struct {
	whole uint16
	frac  uint16
}

// ntscCanonical: 容差 < ntscTolerance 内的用户 fps 都映射到这些 canonical 值。
// frac 值来自 RE-4 dump（test_data/re_tickrate.aep 各 fps comp 的 cdta @0x9E..@0x9F）。
const ntscTolerance = 1e-3

var ntscCanonical = map[float64]frameRateEncoding{
	23.976: {whole: 23, frac: 0xF9DB}, // RE-4 finding placeholder, fill from dump
	29.97:  {whole: 29, frac: 0xF852}, // RE-4 finding placeholder
	59.94:  {whole: 59, frac: 0xF0A4}, // RE-4 finding placeholder
}

// encodeFrameRate 把用户输入 fps（如 29.97 或 30000/1001）normalize 为
// canonical (whole, frac) 二元组写入 cdta。
//
// 行为：
//   - 容差内匹配 ntscCanonical 表 → 用 canonical 值
//   - 否则走通用 round：whole = floor(fps), frac = round((fps - whole) * 65536)
//
// 返回的 (whole, frac) 写入 cdta @cdtaFrameRateWhole / @cdtaFrameRateFrac。
func encodeFrameRate(fps float64) frameRateEncoding {
	for canonicalFps, enc := range ntscCanonical {
		if math.Abs(fps-canonicalFps) < ntscTolerance {
			return enc
		}
	}
	whole := uint16(math.Floor(fps))
	frac := uint16(math.Round((fps - float64(whole)) * 65536))
	return frameRateEncoding{whole: whole, frac: frac}
}

// decodeFrameRate 反向解（仅给单测验证 round-trip）。
func decodeFrameRate(enc frameRateEncoding) float64 {
	return float64(enc.whole) + float64(enc.frac)/65536.0
}
```

- [ ] **Step 2: Build**

```bash
go build ./internal/aep/
```

- [ ] **Step 3: Commit**

```bash
git add internal/aep/framerate_canonical.go
git commit -m "feat(aep): NTSC canonical fps encoding table (RE-4)"
```

---

### Task 1.4: framerate_canonical round-trip tests

**Files:**
- Create: `internal/aep/framerate_canonical_test.go`

- [ ] **Step 1: 写失败测试**

```go
// internal/aep/framerate_canonical_test.go
package aep

import (
	"math"
	"testing"
)

func TestNTSCCanonicalRoundtrip(t *testing.T) {
	cases := []struct {
		name string
		in   float64
		want float64 // 期望 decode 后精确等于 (canonical)
	}{
		{"29.97 exact", 29.97, 29.97},
		{"30000/1001 NTSC", 30000.0 / 1001.0, 29.97},
		{"23.976 exact", 23.976, 23.976},
		{"24000/1001 NTSC", 24000.0 / 1001.0, 23.976},
		{"59.94 exact", 59.94, 59.94},
		{"30 fps integer", 30.0, 30.0},
		{"24 fps integer", 24.0, 24.0},
		{"25 fps PAL", 25.0, 25.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc := encodeFrameRate(tc.in)
			got := decodeFrameRate(enc)
			if math.Abs(got-tc.want) > 1e-5 {
				t.Errorf("encode→decode %g: got %g, want %g (enc=%+v)", tc.in, got, tc.want, enc)
			}
		})
	}
}

func TestNTSCCanonicalIdempotent(t *testing.T) {
	// 容差内任何输入都映射到 canonical 值
	v1 := encodeFrameRate(29.97)
	v2 := encodeFrameRate(30000.0 / 1001.0)
	if v1 != v2 {
		t.Errorf("29.97 (%+v) != 30000/1001 (%+v): canonicalization failed", v1, v2)
	}
}
```

- [ ] **Step 2: Run + verify pass**

```bash
go test -count=1 ./internal/aep/ -run 'TestNTSC' -v
```
Expected: PASS (frac values may need adjustment based on RE-4 finding).

- [ ] **Step 3: Commit**

```bash
git add internal/aep/framerate_canonical_test.go
git commit -m "test(framerate): NTSC canonical round-trip + idempotency"
```

---

# Phase 2 — `Project.nextItemID` + `Project.rootFold`

### Task 2.1: 加 Project 非导出字段

**Files:**
- Modify: `internal/aep/types_core.go`

- [ ] **Step 1: 在 Project struct 加字段**

定位 `type Project struct {` 内任意 unexported 字段附近，加：

```go
// (in types_core.go, inside Project struct)
nextItemID uint32      // monotonic Item ID counter; never reused
rootFold   *rifx.Chunk // cached root Fold LIST reference; never owned
```

- [ ] **Step 2: Build**

```bash
go build ./internal/aep/
```

- [ ] **Step 3: Commit**

```bash
git add internal/aep/types_core.go
git commit -m "feat(types): Project.nextItemID + Project.rootFold fields"
```

---

### Task 2.2: 加 `initDerived(rifxRoot)` helper + wire 到 parseProject

**Files:**
- Modify: `internal/aep/parse.go`

- [ ] **Step 1: 加 helper method**

在 parse.go 末尾或合适位置加：

```go
// initDerived 在 parseProject 收尾时调用，初始化 Project 的 derived state：
//   - nextItemID = max(已有所有 item IDs) + 1
//   - rootFold = root Egg! 下第一个 formType=Fold 的 LIST
func (p *Project) initDerived(rifxRoot *rifx.Chunk) {
	var maxID uint32
	for _, c := range p.Compositions {
		if c.ID > maxID {
			maxID = c.ID
		}
	}
	for _, f := range p.Footage {
		if f.ID > maxID {
			maxID = f.ID
		}
	}
	for _, fo := range p.Folders {
		if fo.ID > maxID {
			maxID = fo.ID
		}
	}
	p.nextItemID = maxID + 1

	for _, c := range rifxRoot.Children {
		if c.IsList() && c.FormType == rifx.IDFold {
			p.rootFold = c
			break
		}
	}
}

// allocItemID 返回下一个可用 Item ID 并递增计数器。Monotonic，不 reuse。
func (p *Project) allocItemID() uint32 {
	id := p.nextItemID
	p.nextItemID++
	return id
}
```

- [ ] **Step 2: 找 parseProject 函数末尾，加 initDerived 调用**

```bash
grep -n 'func parseProject\|return proj' internal/aep/parse.go | head -10
```

定位 parseProject 返回 *Project 前的位置。在那里加：

```go
proj.initDerived(rifxRoot) // rifxRoot 是当前 scope 内的 root *rifx.Chunk
return proj, nil
```

具体变量名照 parse.go 实际命名调整。

- [ ] **Step 3: Build + run all tests**

```bash
go vet ./...
go test -count=1 ./internal/aep/...
```
Expected: vet clean, 109+ PASS（无破坏）。

- [ ] **Step 4: Commit**

```bash
git add internal/aep/parse.go
git commit -m "feat(parse): initDerived populates Project.nextItemID + rootFold"
```

---

### Task 2.3: 测试 nextItemID + rootFold 初始化

**Files:**
- Modify: `internal/aep/aep_test.go` 加新测试

- [ ] **Step 1: 写测试**

```go
// 加到 aep_test.go (package aep_test)

func TestProjectInitDerived_EmptyProject(t *testing.T) {
	data := buildMinimalAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	// buildMinimalAEP 含若干 comp/footage，验证 nextItemID > max existing
	var max uint32
	for _, c := range proj.Compositions {
		if c.ID > max {
			max = c.ID
		}
	}
	if got := proj.NextItemIDForTest(); got <= max {
		t.Errorf("nextItemID = %d, want > %d", got, max)
	}
	if proj.RootFoldForTest() == nil {
		t.Error("rootFold not cached")
	}
}
```

由于 `nextItemID` / `rootFold` 是非导出字段，test 通过 test-only accessor 拿。在 `internal/aep/parse.go` 末尾加（test build only via const file）：

```go
// 加到 internal/aep/testhelpers_test.go 新文件（package aep）
package aep

// 仅给同 package 的 white-box 测试用；不导出
func (p *Project) NextItemIDForTest() uint32 { return p.nextItemID }
func (p *Project) RootFoldForTest() *rifx.Chunk { return p.rootFold }
```

或更优雅：把 test 写成 white-box（package aep 而非 aep_test）。本仓既有 `*_test.go` 都是 `package aep_test`，所以走 accessor 路线。

**命名约定**：所有 white-box test accessor 用 `<Name>ForTest()` 后缀（大写 export-style 但仅 test build 可见因为定义在 `_test.go`）。

- [ ] **Step 2: 写 white-box test helper**

```go
// internal/aep/testhelpers_test.go (package aep — white box)
package aep

import "github.com/example/aep-parser/internal/rifx"

// White-box accessors for unexported Project fields.
// _test.go 后缀使这些方法仅在 test build 时编译，不污染 production binary。
func (p *Project) NextItemIDForTest() uint32     { return p.nextItemID }
func (p *Project) RootFoldForTest() *rifx.Chunk  { return p.rootFold }
```

后续 NewProject/NewComposition test 大多走 `package aep_test`（black-box）；需要白盒访问的几个 test（例如 `TestProjectInitDerived_EmptyProject`）单独走 `package aep`，调用 `ForTest` 方法。两套 test 文件混 OK，Go 允许同 dir 内既有 `package foo` test 也有 `package foo_test` test。

- [ ] **Step 3: Run test**

```bash
go test -count=1 ./internal/aep/ -run 'TestProjectInitDerived' -v
```
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/aep/aep_test.go internal/aep/testhelpers_test.go
git commit -m "test(aep): Project.nextItemID + rootFold init from FromReader"
```

---

# Phase 3 — `aep.NewProject(target ...AETarget)`

### Task 3.1: AETarget enum

**Files:**
- Create: `internal/aep/new_project.go`

- [ ] **Step 1: 写 AETarget enum**

```go
// internal/aep/new_project.go
package aep

import (
	"bytes"
	_ "embed"
	"fmt"
)

// AETarget selects which AE version's empty-project skeleton NewProject
// uses as the base. Output .aep files identify themselves as that AE
// version's format (via the `svap` chunk). Newer AE versions may silently
// upgrade them; older AE versions refuse to open files claiming a newer
// version.
//
// Underlying values are AE marketing years (2020/2022/2025/…) for
// debuggable panic messages and natural `target >= 2025` comparisons.
//
// AETarget values are NOT forward-compatible — unknown values panic.
// Upgrade the library when targeting a newer AE version.
type AETarget int

const (
	TargetAE2020 AETarget = 2020 // default; max compatibility (any AE 2020+ opens)
	TargetAE2022 AETarget = 2022 // 30 chunks; adds AE 24+ color-mgmt prefs
	TargetAE2025 AETarget = 2025 // 30 chunks; latest tested
)

//go:embed templates/2020.aep
var embeddedTemplate2020 []byte

//go:embed templates/2022.aep
var embeddedTemplate2022 []byte

//go:embed templates/2025.aep
var embeddedTemplate2025 []byte
```

- [ ] **Step 2: Build (no Go func defined yet)**

```bash
go build ./internal/aep/
```
Expected: builds; `_ = embeddedTemplate*` not yet used so might warn about unused — that's fine, the // go:embed loads it.

- [ ] **Step 3: Commit**

```bash
git add internal/aep/new_project.go
git commit -m "feat(aep): AETarget enum + embed AE 2020/2022/2025 templates"
```

---

### Task 3.2: NewProject 实现

**Files:**
- Modify: `internal/aep/new_project.go`

- [ ] **Step 1: 加 NewProject 函数**

```go
// 接续 new_project.go
//
// NewProject returns a fresh empty Project parsed from the embedded
// AE skeleton matching the requested target.
//
// Optional target arg: zero args = TargetAE2020 (max compatibility). Pass
// at most one target. Subsequent NewComposition calls populate it.
//
// Never returns an error: the embedded templates are build-time trusted;
// parser bugs panic with a "build bug" message (not user-facing).
// Panics on: multiple target args, or unknown AETarget value (forward-incompat).
func NewProject(target ...AETarget) *Project {
	t := TargetAE2020 // default
	switch len(target) {
	case 0:
		// use default
	case 1:
		t = target[0]
	default:
		panic(fmt.Sprintf("aep: NewProject accepts at most one target, got %d", len(target)))
	}

	var tmpl []byte
	switch t {
	case TargetAE2020:
		tmpl = embeddedTemplate2020
	case TargetAE2022:
		tmpl = embeddedTemplate2022
	case TargetAE2025:
		tmpl = embeddedTemplate2025
	default:
		panic(fmt.Sprintf("aep: unknown AETarget %d (forward-incompat; upgrade library)", int(t)))
	}

	p, err := FromReader(bytes.NewReader(tmpl))
	if err != nil {
		panic(fmt.Sprintf("aep: corrupt embedded template for AE %d (build bug): %v", int(t), err))
	}
	return p
}
```

- [ ] **Step 2: Build + smoke test**

```bash
go build ./internal/aep/
cat > /tmp/v2_smoke.go <<'EOF'
package main

import (
	"fmt"
	aep "github.com/example/aep-parser/internal/aep"
)

func main() {
	p := aep.NewProject()
	fmt.Printf("comps=%d footage=%d folders=%d\n", len(p.Compositions), len(p.Footage), len(p.Folders))
}
EOF
go run /tmp/v2_smoke.go
```
Expected: `comps=0 footage=0 folders=0`

- [ ] **Step 3: Commit**

```bash
git add internal/aep/new_project.go
git commit -m "feat(aep): NewProject(target ...AETarget) *Project"
```

---

### Task 3.3: NewProject tests + svap expected constants

**Files:**
- Create: `internal/aep/new_project_test.go`

- [ ] **Step 1: 写 svap expected constants + tests**

```go
// internal/aep/new_project_test.go
package aep_test

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

// svap expected bytes — frozen from template files. If templates are
// silently replaced, these constants will mismatch and tests fail loud.
// 来源: go run ./tmp_debug/dump_root templates/{2020,2022,2025}.aep | grep svap
const (
	svapExpectedAE2020 = "0b0b862d"
	svapExpectedAE2022 = "0b330640"
	svapExpectedAE2025 = "0f088644"
)

func TestNewProject_DefaultTargetAE2020(t *testing.T) {
	p := aep.NewProject()
	if len(p.Compositions) != 0 {
		t.Errorf("Compositions = %d, want 0", len(p.Compositions))
	}
	if len(p.Footage) != 0 {
		t.Errorf("Footage = %d, want 0", len(p.Footage))
	}
	if len(p.Folders) != 0 {
		t.Errorf("Folders = %d, want 0", len(p.Folders))
	}
	if p.BitsPerChannel != aep.BPC8 {
		t.Errorf("BitsPerChannel = %v, want BPC8", p.BitsPerChannel)
	}
	if len(p.Warnings) != 0 {
		t.Errorf("template produced %d warnings: %v", len(p.Warnings), p.Warnings)
	}
}

func TestNewProject_AllTargetsParseAndSvap(t *testing.T) {
	cases := []struct {
		target       aep.AETarget
		wantSvap     string
		wantChunkN   int // root Egg! 直接 children 数（24 for 2020, 30 for 2022/2025）
	}{
		{aep.TargetAE2020, svapExpectedAE2020, 24},
		{aep.TargetAE2022, svapExpectedAE2022, 30},
		{aep.TargetAE2025, svapExpectedAE2025, 30},
	}
	for _, tc := range cases {
		t.Run(target_name(tc.target), func(t *testing.T) {
			p := aep.NewProject(tc.target)
			if len(p.Warnings) != 0 {
				t.Errorf("template AE %d produced %d warnings: %v", int(tc.target), len(p.Warnings), p.Warnings)
			}

			// roundtrip 重 serialize 看 svap byte
			var buf bytes.Buffer
			if err := p.WriteAEP(&buf); err != nil {
				t.Fatalf("WriteAEP: %v", err)
			}
			root, err := rifx.Parse(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("re-parse: %v", err)
			}
			if len(root.Children) != tc.wantChunkN {
				t.Errorf("AE %d: root children = %d, want %d", int(tc.target), len(root.Children), tc.wantChunkN)
			}
			for _, c := range root.Children {
				if string(c.ID[:]) == "svap" {
					got := hex.EncodeToString(c.Data)
					if !strings.EqualFold(got, tc.wantSvap) {
						t.Errorf("AE %d svap = %s, want %s", int(tc.target), got, tc.wantSvap)
					}
					return
				}
			}
			t.Errorf("AE %d: svap chunk not found in output", int(tc.target))
		})
	}
}

func TestNewProject_RejectsMultipleTargets(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic, got none")
			return
		}
		msg := r.(string)
		if !strings.Contains(msg, "accepts at most one target") {
			t.Errorf("panic msg = %q, want contains 'accepts at most one target'", msg)
		}
	}()
	aep.NewProject(aep.TargetAE2020, aep.TargetAE2025)
}

func TestNewProject_RejectsUnknownTarget(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic, got none")
			return
		}
		msg := r.(string)
		if !strings.Contains(msg, "forward-incompat") {
			t.Errorf("panic msg = %q, want contains 'forward-incompat'", msg)
		}
	}()
	aep.NewProject(aep.AETarget(9999))
}

func TestNewProject_IndependentInstances(t *testing.T) {
	p1 := aep.NewProject()
	p2 := aep.NewProject()
	// mutating p1 should not affect p2's Warnings slice
	p1.Warnings = append(p1.Warnings, "fake")
	if len(p2.Warnings) != 0 {
		t.Errorf("p2.Warnings polluted by p1 mutation: %v", p2.Warnings)
	}
}

func target_name(t aep.AETarget) string {
	switch t {
	case aep.TargetAE2020:
		return "AE2020"
	case aep.TargetAE2022:
		return "AE2022"
	case aep.TargetAE2025:
		return "AE2025"
	}
	return "unknown"
}
```

- [ ] **Step 2: Run tests**

```bash
go test -count=1 ./internal/aep/ -run 'TestNewProject' -v
```
Expected: 5 tests PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/aep/new_project_test.go
git commit -m "test(new_project): default target / all targets / svap expected / reject paths"
```

---

# Phase 4 — Chunk builders

### Task 4.1: `buildCompIide`

**Files:**
- Create: `internal/aep/new_composition.go`

- [ ] **Step 1: dump 既有 fixture 的 iide 字节**

```bash
go run ./tmp_debug/list_item_chunks test_data/fdta_probe/AE2025_1comp.aep | grep -A 1 'chunk iide'
```

记录 4 字节 hex。

- [ ] **Step 2: 写 builder**

```go
// internal/aep/new_composition.go
package aep

import (
	"encoding/binary"
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// buildCompIide 构造 4-byte iide chunk（Item index entry header）。
// 内容来自 AE 2025 saved fixture dump（new_composition.go 注释段记录的 hex）。
// 语义未 RE，但 AE 跨 fixture 都接受相同字节。
func buildCompIide() *rifx.Chunk {
	return &rifx.Chunk{
		ID:   rifx.ChunkID{'i', 'i', 'd', 'e'},
		Data: []byte{0x00, 0x00, 0x00, 0x01}, // <-- 用 Task 4.1 Step 1 dump 的实际值替换
	}
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/aep/new_composition.go
git commit -m "feat(new_comp): buildCompIide (4 B fixed from AE 25 fixture)"
```

---

### Task 4.2: `buildCompIdpc`（按 RE-2 finding）

**Files:**
- Modify: `internal/aep/new_composition.go`

- [ ] **Step 1: 添加 import + builder**

```go
// 加 import
import (
	"crypto/rand"
	// ... existing
)

// buildCompIdpc 构造 idpc chunk（per-item persistent key）。
// RE-2 finding: 长度 = 8 bytes（或 16，按实测调整）。AE 接受 zero default
// 但跨项目引用 / Essential Graphics 可能依赖唯一性 — 默认走 crypto/rand。
//
// 若未来确认 AE 不要求唯一，可改用全 0；当前 safest = random（zero downside）。
func buildCompIdpc() *rifx.Chunk {
	buf := make([]byte, 8) // <-- 按 RE-2 实测调整长度
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand 失败极罕见；panic 不是 user-recoverable error
		panic(fmt.Sprintf("aep: crypto/rand failed: %v", err))
	}
	return &rifx.Chunk{
		ID:   rifx.ChunkID{'i', 'd', 'p', 'c'},
		Data: buf,
	}
}
```

- [ ] **Step 2: Build**

```bash
go build ./internal/aep/
```

- [ ] **Step 3: Commit**

```bash
git add internal/aep/new_composition.go
git commit -m "feat(new_comp): buildCompIdpc (crypto/rand per-item key, RE-2)"
```

---

### Task 4.3: `buildCompIdta`（按 RE-3 finding）

**Files:**
- Modify: `internal/aep/new_composition.go`

- [ ] **Step 1: 准备 idta default bytes（template-copy 策略）**

从 AE 2025 saved fixture 的 idta 84 字节 dump，作为 builder 默认值。把 hex 抄进代码：

```go
// idtaCompDefaultBytes：从 AE 25 saved 1-comp fixture dump 的 idta 84 字节。
// 我们 RE 出来的字段（type / id / label）覆写，其余 75 字节保持 AE-default 不动。
var idtaCompDefaultBytes = [idtaSize]byte{
	// 抄 AE 2025 fixture 的 idta hex（按 RE-3 Step 3 dump 填）
	// 例如：
	// 0x00, 0x04, ... ...
}
```

- [ ] **Step 2: 写 builder**

```go
// buildCompIdta 构造 84-byte idta（comp item 元数据）。
// 用 idtaCompDefaultBytes 作底，覆写 type / itemID / label。
func buildCompIdta(itemID uint32) *rifx.Chunk {
	data := make([]byte, idtaSize)
	copy(data, idtaCompDefaultBytes[:])
	// type code @0x00 已在 default 里写好（0x0004）；冗余覆写防 default 漂移
	binary.BigEndian.PutUint16(data[idtaTypeCode:], 0x0004)
	binary.BigEndian.PutUint32(data[idtaItemID:], itemID)
	data[idtaLabel] = 0 // label = none
	return &rifx.Chunk{
		ID:   rifx.ChunkID{'i', 'd', 't', 'a'},
		Data: data,
	}
}
```

- [ ] **Step 3: Build**

```bash
go build ./internal/aep/
```

- [ ] **Step 4: Commit**

```bash
git add internal/aep/new_composition.go
git commit -m "feat(new_comp): buildCompIdta (template-copy 75 B + RE'd type/id/label)"
```

---

### Task 4.4: `buildCompCdta`

**Files:**
- Modify: `internal/aep/new_composition.go`

- [ ] **Step 1: 写 builder**

```go
// buildCompCdta 构造 204-byte cdta，5 个必填 + AE 默认值。
// Frame rate 经 encodeFrameRate 走 canonical 表（RE-4）。
// WorkAreaEnd 写 sentinel 0xFFFFFFFF（per RE-4 finding；若 RE 显示真值，改 hardcoded）。
func buildCompCdta(w, h uint16, fps, duration float64) []byte {
	d := make([]byte, cdtaSize)

	// ResolutionFactor @0x00/0x02 default [1,1]
	binary.BigEndian.PutUint16(d[cdtaResolutionFactorX:], 1)
	binary.BigEndian.PutUint16(d[cdtaResolutionFactorY:], 1)

	// TickRate @0x08（modern AE 25 写 1024 = 0x400；按 fixture 实测调整）
	binary.BigEndian.PutUint32(d[cdtaTickRate:], 1024)

	// WorkArea @0x1C..@0x2B
	binary.BigEndian.PutUint32(d[cdtaWorkAreaStart:], 0)        // start dividend = 0
	binary.BigEndian.PutUint32(d[cdtaWorkAreaStartDiv:], 600)   // divisor = 600
	binary.BigEndian.PutUint32(d[cdtaWorkAreaEnd:], 0xFFFFFFFF) // sentinel = "use duration"
	binary.BigEndian.PutUint32(d[cdtaWorkAreaEndDiv:], 600)

	// BGColor @0x34..@0x36 default {0,0,0}
	// （bytes default 已 0）

	// Width @0x8C / Height @0x8E
	binary.BigEndian.PutUint16(d[cdtaWidth:], w)
	binary.BigEndian.PutUint16(d[cdtaHeight:], h)

	// PixelAspect @0x90/0x94 default 1/1
	binary.BigEndian.PutUint32(d[cdtaPixelAspectNum:], 1)
	binary.BigEndian.PutUint32(d[cdtaPixelAspectDen:], 1)

	// FrameRate @0x9C/0x9E — canonical encoding
	enc := encodeFrameRate(fps)
	binary.BigEndian.PutUint16(d[cdtaFrameRateWhole:], enc.whole)
	binary.BigEndian.PutUint16(d[cdtaFrameRateFrac:], enc.frac)

	// DisplayStartTime @0xA4/0xA8 default 0/1
	binary.BigEndian.PutUint32(d[cdtaDisplayStartTime:], 0)
	binary.BigEndian.PutUint32(d[cdtaDisplayStartDiv:], 1)

	// ShutterAngle @0xAE default 180
	binary.BigEndian.PutUint16(d[cdtaShutterAngle:], 180)

	// Duration @0xB0 = round(duration_seconds * fps) frames
	durationFrames := uint32(math.Round(duration * fps))
	binary.BigEndian.PutUint32(d[cdtaDuration:], durationFrames)

	// ShutterPhase @0xB4 default 0
	// MotionBlurAdaptive @0xC4 default 128
	binary.BigEndian.PutUint32(d[cdtaMotionBlurAdaptive:], 128)
	// MotionBlurSamples @0xC8 default 16
	binary.BigEndian.PutUint32(d[cdtaMotionBlurSamples:], 16)

	return d
}
```

加 import:

```go
import "math"
```

- [ ] **Step 2: Build**

```bash
go build ./internal/aep/
```

- [ ] **Step 3: Commit**

```bash
git add internal/aep/new_composition.go
git commit -m "feat(new_comp): buildCompCdta (5 required + AE defaults + canonical fps)"
```

---

### Task 4.5: `buildEmptyLayrList` + template-copy helpers

**Files:**
- Modify: `internal/aep/new_composition.go`

- [ ] **Step 1: 写 builders**

```go
// buildEmptyLayrList 构造空 LIST formType=Layr（0 children）。
func buildEmptyLayrList() *rifx.Chunk {
	return &rifx.Chunk{
		ID:       rifx.IDList,
		FormType: rifx.IDLayr,
		Children: nil,
	}
}

// templateCompItemChunks holds non-builder chunks copied from AE 25 fixture's
// dummy comp Item LIST. These are chunks we haven't RE'd semantically but
// AE expects them present (LIST(dats) / cdrp / comr / LIST(PRin) / LIST(DLay)
// / etc.). Loaded on first NewComposition call from the AE 2025 template's
// (only) comp — if template has no dummy comp, populated from a separate
// embedded fixture.
//
// 策略：从 AE 25 临时 saved-with-1-comp fixture 提取这些 chunk 原始 *rifx.Chunk
// 引用，每次 NewComposition deep-copy 一份给新 Item LIST。
var templateCompItemChunks []*rifx.Chunk

// initTemplateCompItemChunks 从一个 saved-with-1-comp fixture（test_data 内或
// 第 4 个 embed）解析出 Item LIST，剔除 iide/idpc/idta/Utf8/cdta/Layr 这 6 个
// builder-managed children，剩下的（dats/cdrp/comr/PRin/DLay 等）作 template。
//
// 实现细节见 plan Task 4.6。当前 NewComposition 流程会在第一次调用时 lazy init。
func initTemplateCompItemChunks() { /* impl in 4.6 */ }
```

- [ ] **Step 2: Build**

```bash
go build ./internal/aep/
```

- [ ] **Step 3: Commit**

```bash
git add internal/aep/new_composition.go
git commit -m "feat(new_comp): empty Layr LIST builder + template-chunks placeholder"
```

---

### Task 4.6: `buildCompItem` + template chunks 初始化

**Files:**
- Modify: `internal/aep/new_composition.go`
- Need: 一份 AE 25 saved-with-1-comp fixture，可复用 `test_data/fdta_probe/AE2025_1comp.aep`，或新嵌入 `internal/aep/templates/2025_1comp.aep`

- [ ] **Step 1: 决策 fixture 来源**

Option A: 把 `test_data/fdta_probe/AE2025_1comp.aep` copy 到 `internal/aep/templates/2025_dummy_comp.aep` + embed。
Option B: 用 RE-3 Step 5 已经生成的 fixture，硬编码"template copy"路径。

走 Option A（embed-friendly，spec invariant "build-time trusted"）。

```bash
cp test_data/fdta_probe/AE2025_1comp.aep internal/aep/templates/2025_dummy_comp.aep
```

- [ ] **Step 2: embed + parse + 提取 template chunks**

```go
// 加 import
import "sync"

//go:embed templates/2025_dummy_comp.aep
var embeddedDummyCompTemplate []byte

var templateInit sync.Once

func ensureTemplateCompItemChunks() {
	templateInit.Do(func() {
		p, err := FromReader(bytes.NewReader(embeddedDummyCompTemplate))
		if err != nil {
			panic(fmt.Sprintf("aep: corrupt dummy-comp template (build bug): %v", err))
		}
		if len(p.Compositions) == 0 {
			panic("aep: dummy-comp template missing comp (build bug)")
		}
		// 找到 comp Item LIST，把 builder-managed chunk 之外的 children 存起来
		comp := p.Compositions[0]
		itemList := comp.ItemListChunkForTest() // 需要 white-box accessor，详见下面 Step 3
		if itemList == nil {
			panic("aep: dummy-comp Item LIST not accessible (build bug)")
		}
		for _, ch := range itemList.Children {
			if isBuilderManagedChunk(ch) {
				continue
			}
			templateCompItemChunks = append(templateCompItemChunks, deepCloneChunk(ch))
		}
	})
}

// isBuilderManagedChunk: builder 自己合成的 chunk，不从 template 复制。
func isBuilderManagedChunk(c *rifx.Chunk) bool {
	tag := string(c.ID[:])
	switch tag {
	case "iide", "idpc", "idta", "Utf8", "cdta":
		return true
	}
	if c.IsList() && string(c.FormType[:]) == "Layr" {
		return true
	}
	return false
}

// deepCloneChunk: 递归深拷贝 *rifx.Chunk（防止 NewComposition 间共享 mutation）。
func deepCloneChunk(c *rifx.Chunk) *rifx.Chunk {
	clone := &rifx.Chunk{
		ID:       c.ID,
		FormType: c.FormType,
		Data:     append([]byte(nil), c.Data...),
	}
	for _, ch := range c.Children {
		clone.Children = append(clone.Children, deepCloneChunk(ch))
	}
	return clone
}
```

- [ ] **Step 3: 加 white-box accessor**

由于 `Composition.itemListChunk` 或类似字段是 unexported，需要 test-build accessor。或者直接在 parse.go 找到 itemListChunk 暴露 unexported 方法（同 package 内部用，不破 public API）。

实际更简单：parse_composition.go 已经把 Item LIST 作为参数传给 parseComposition，对 *Composition 加 unexported `itemList *rifx.Chunk` 字段持引用（如已有则用，否则加）。

```bash
grep -n 'itemList\|Item LIST\|parseComposition' internal/aep/types_core.go internal/aep/parse_composition.go | head -10
```

如未有，加：

```go
// internal/aep/types_core.go inside Composition struct
itemList *rifx.Chunk // owning Item LIST; populated by parseComposition

// internal/aep/parse_composition.go inside parseComposition
comp.itemList = item // 在 parseComposition 入参 item *rifx.Chunk 后立即赋值
```

- [ ] **Step 4: 写 buildCompItem**

```go
// buildCompItem 包成完整 LIST formType=Item。
// Children 顺序（按 AE 2025 saved fixture，见 test_data/comp_item_children.golden.txt）：
//   iide / idpc / idta / Utf8 / LIST(dats) / cdta / cdrp / comr / LIST(PRin) / LIST(DLay) / ... / 空 LIST(Layr)
//
// builder-managed: iide / idpc / idta / Utf8 / cdta / Layr
// template-copy:   dats / cdrp / comr / PRin / DLay / others
func buildCompItem(itemID uint32, name string, cdta []byte) *rifx.Chunk {
	ensureTemplateCompItemChunks()

	// 4 个 builder-managed leading chunks（按 fixture 顺序：iide, idpc, idta, Utf8）
	iide := buildCompIide()
	idpc := buildCompIdpc()
	idta := buildCompIdta(itemID)
	utf8 := &rifx.Chunk{
		ID:   rifx.IDUtf8,
		Data: []byte(name),
	}
	cdtaChunk := &rifx.Chunk{
		ID:   rifx.IDCdta,
		Data: cdta,
	}
	layrList := buildEmptyLayrList()

	// 组装：iide / idpc / idta / Utf8 / <template chunks before cdta> / cdta / <template chunks after cdta> / Layr
	// 具体插入点按 golden fixture 的 children index
	children := []*rifx.Chunk{iide, idpc, idta, utf8}
	// dats LIST 在 cdta 之前（index 4 in golden）
	for _, tc := range templateCompItemChunks {
		// 简化：把 template chunks 全部插到 cdta 之前；layr 是最后一个；具体顺序按 golden fixture 调
		children = append(children, deepCloneChunk(tc))
	}
	children = append(children, cdtaChunk, layrList)

	return &rifx.Chunk{
		ID:       rifx.IDList,
		FormType: rifx.IDItem,
		Children: children,
	}
}
```

实际顺序细节（template chunks 的相对位置）须按 golden fixture (RE-3) 调。本 step 后跑 golden fixture 测试（Task 4.7）验证顺序。

- [ ] **Step 5: Build**

```bash
go build ./internal/aep/
```

- [ ] **Step 6: Commit**

```bash
git add internal/aep/new_composition.go internal/aep/templates/2025_dummy_comp.aep internal/aep/types_core.go internal/aep/parse_composition.go
git commit -m "feat(new_comp): buildCompItem with template-copy strategy + dummy-comp embed"
```

---

### Task 4.7: Golden fixture 测试

**Files:**
- Modify: `internal/aep/new_composition_test.go`（创建）

- [ ] **Step 1: 写 builder 输出 vs golden 对比测试**

```go
// internal/aep/new_composition_test.go
package aep_test

import (
	"fmt"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// TestBuildCompItemMatchesGoldenStructure 验证 buildCompItem 产出的 children
// 顺序 / 类型 / 大致大小跟 AE 25 saved 1-comp fixture 一致。
//
// 跑这个测试 = 验证 Adobe binary format invariant："builder output structurally
// indistinguishable from AE-saved comp Item LIST"。任何 chunk 顺序漂移 or
// 缺失 immediately fail.
func TestBuildCompItemMatchesGoldenStructure(t *testing.T) {
	p := aep.NewProject() // 默认 AE 2020 — 注意：AE 2020 vs AE 2025 dummy comp 结构可能差异；此测试以 AE 2025 为 golden
	comp, err := p.NewComposition("GoldenTest", 1920, 1080, 29.97, 5)
	if err != nil {
		t.Fatal(err)
	}

	// expected order from test_data/comp_item_children.golden.txt
	expected := []string{
		"iide", "idpc", "idta", "Utf8", "LIST(dats)", "cdta", "cdrp", "comr",
		"LIST(PRin)", "LIST(DLay)", // ... 完整顺序从 golden 抄
		"LIST(Layr)", // 最后一个
	}

	got := chunkTagSequence(comp)
	if !sliceEqual(got, expected) {
		t.Errorf("Item LIST children order mismatch:\n got: %v\nwant: %v", got, expected)
	}
}

func chunkTagSequence(c *aep.Composition) []string {
	itemList := c.ItemListChunkForTest() // unexported test-helper
	if itemList == nil {
		return nil
	}
	var out []string
	for _, ch := range itemList.Children {
		tag := string(ch.ID[:])
		if ch.IsList() {
			out = append(out, fmt.Sprintf("LIST(%s)", string(ch.FormType[:])))
		} else {
			out = append(out, tag)
		}
	}
	return out
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) { return false }
	for i := range a { if a[i] != b[i] { return false } }
	return true
}

// 用于 debug fail：dump 实际 vs 期望对比
var _ = strings.Join
```

`Composition.ItemListChunkForTest()` 白盒 accessor 加到 `internal/aep/testhelpers_test.go`：

```go
// (testhelpers_test.go - package aep)
func (c *Composition) ItemListChunkForTest() *rifx.Chunk { return c.itemList }
```

- [ ] **Step 2: 跑测试 — 这里 expected 顺序可能要按实测调**

```bash
go test -count=1 ./internal/aep/ -run 'TestBuildCompItemMatchesGoldenStructure' -v
```
Expected: 第一次 likely FAIL，对比输出调整 `expected` 数组到 golden fixture 实际顺序，再跑。

- [ ] **Step 3: 把最终 expected 顺序也写到 `test_data/comp_item_children.golden.txt`**

确保 plan 中的 golden fixture 文档跟测试代码常量一致。

- [ ] **Step 4: Commit**

```bash
git add internal/aep/new_composition_test.go internal/aep/testhelpers_test.go test_data/comp_item_children.golden.txt
git commit -m "test(new_comp): buildCompItem matches AE 25 golden chunk order"
```

---

# Phase 5 — `Project.NewComposition` + atomic mutation

### Task 5.1: 输入 validation

**Files:**
- Modify: `internal/aep/new_composition.go`

- [ ] **Step 1: 加 validation 函数**

```go
// validateNewCompositionInputs returns nil if all inputs are valid, or
// an error naming the offending field + value.
func validateNewCompositionInputs(name string, w, h uint16, fps, duration float64) error {
	if name == "" {
		return fmt.Errorf("composition name cannot be empty")
	}
	if w == 0 || h == 0 {
		return fmt.Errorf("composition size must be > 0 (got %dx%d)", w, h)
	}
	if fps <= 0 {
		return fmt.Errorf("frame rate must be > 0 (got %g)", fps)
	}
	if duration <= 0 {
		return fmt.Errorf("duration must be > 0 (got %g)", duration)
	}
	return nil
}
```

- [ ] **Step 2: Build**

```bash
go build ./internal/aep/
```

- [ ] **Step 3: Commit**

```bash
git add internal/aep/new_composition.go
git commit -m "feat(new_comp): validateNewCompositionInputs"
```

---

### Task 5.2: `Project.NewComposition` 主流程（含 atomic + warnings-as-failure）

**Files:**
- Modify: `internal/aep/new_composition.go`

- [ ] **Step 1: 写主函数**

```go
// NewComposition adds an empty composition to the project's root folder.
// 详 spec: workshop/specs/v2-1-foundation-design.md §Public API
func (p *Project) NewComposition(
	name string,
	width, height uint16,
	frameRate, duration float64,
) (*Composition, error) {
	// 1. Validate
	if err := validateNewCompositionInputs(name, width, height, frameRate, duration); err != nil {
		return nil, err
	}

	// 2. Allocate ID (monotonic, never reuses)
	id := p.allocItemID()

	// 3. Build chunks
	cdtaBytes := buildCompCdta(width, height, frameRate, duration)
	itemList := buildCompItem(id, name, cdtaBytes)

	// 4. Atomic mutation prep
	if p.rootFold == nil {
		return nil, fmt.Errorf("internal: project missing root Fold (template malformed?)")
	}
	oldChildLen := len(p.rootFold.Children)
	oldCompsLen := len(p.Compositions)
	oldWarningsLen := len(p.Warnings)

	// 5. Append to rootFold + reparse closed loop
	// parseComposition signature: (item *rifx.Chunk, id uint32, name string, warnings *[]string) (*Composition, error)
	p.rootFold.Children = append(p.rootFold.Children, itemList)
	comp, err := parseComposition(itemList, id, name, &p.Warnings)
	if err != nil {
		// Rollback
		p.rootFold.Children = p.rootFold.Children[:oldChildLen]
		return nil, fmt.Errorf("internal: re-parsing new composition: %w", err)
	}

	// 6. Warnings-as-failure: builder must produce zero parser warnings
	if len(p.Warnings) != oldWarningsLen {
		newWarnings := p.Warnings[oldWarningsLen:]
		p.rootFold.Children = p.rootFold.Children[:oldChildLen]
		p.Warnings = p.Warnings[:oldWarningsLen]
		return nil, fmt.Errorf("internal: builder produced %d parser warning(s): %v", len(newWarnings), newWarnings)
	}

	// 7. Wire back-pointer + register in typed index
	comp.proj = p
	comp.itemList = itemList
	p.Compositions = append(p.Compositions, comp)

	// 8. Sanity: oldCompsLen + 1 == len(p.Compositions)
	_ = oldCompsLen
	return comp, nil
}
```

注意：`parseComposition` 现有签名可能跟上面不完全一致；按实际 grep 调整：

```bash
grep -n 'func parseComposition' internal/aep/parse_composition.go
```

- [ ] **Step 2: Build**

```bash
go build ./internal/aep/
```
Expected: compiles. 如果 parseComposition 签名不同，按实际改 step 1 的 call site。

- [ ] **Step 3: Smoke test**

```bash
cat > /tmp/v2_newcomp_smoke.go <<'EOF'
package main

import (
	"fmt"
	aep "github.com/example/aep-parser/internal/aep"
)

func main() {
	p := aep.NewProject()
	c, err := p.NewComposition("Main", 1920, 1080, 29.97, 10)
	if err != nil { panic(err) }
	fmt.Printf("created comp: id=%d name=%q size=%dx%d fps=%.2f dur=%.1f\n",
		c.ID, c.Name, c.Width, c.Height, c.FrameRate, c.Duration)
	fmt.Printf("project comps: %d\n", len(p.Compositions))
}
EOF
go run /tmp/v2_newcomp_smoke.go
```
Expected: `created comp: id=N name="Main" size=1920x1080 fps=29.97 dur=10.0` + `project comps: 1`

- [ ] **Step 4: Commit**

```bash
git add internal/aep/new_composition.go
git commit -m "feat(new_comp): Project.NewComposition with atomic mutation + warnings-as-failure"
```

---

### Task 5.3: NewComposition unit tests（字段 + roundtrip）

**Files:**
- Modify: `internal/aep/new_composition_test.go`

- [ ] **Step 1: 写测试**

```go
// （继续在 new_composition_test.go）

func TestNewComposition_Fields(t *testing.T) {
	p := aep.NewProject()
	c, err := p.NewComposition("Main", 1920, 1080, 29.97, 10)
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "Main" {
		t.Errorf("Name = %q, want Main", c.Name)
	}
	if c.Width != 1920 || c.Height != 1080 {
		t.Errorf("Size = %dx%d, want 1920x1080", c.Width, c.Height)
	}
	if math.Abs(c.FrameRate-29.97) > 1e-5 {
		t.Errorf("FrameRate = %g, want 29.97", c.FrameRate)
	}
	if math.Abs(c.Duration-10) > 1e-3 {
		t.Errorf("Duration = %g, want 10", c.Duration)
	}
	if c.PixelAspect != 1.0 {
		t.Errorf("PixelAspect = %g, want 1.0", c.PixelAspect)
	}
	if c.ResolutionFactor != [2]uint16{1, 1} {
		t.Errorf("ResolutionFactor = %v, want [1 1]", c.ResolutionFactor)
	}
	if c.BGColor != [3]uint8{0, 0, 0} {
		t.Errorf("BGColor = %v, want [0 0 0]", c.BGColor)
	}
	if c.ID == 0 {
		t.Error("ID not allocated")
	}
	if len(p.Compositions) != 1 || p.Compositions[0] != c {
		t.Errorf("Project.Compositions not updated correctly")
	}
}

func TestNewComposition_Roundtrip(t *testing.T) {
	p := aep.NewProject()
	c1, _ := p.NewComposition("Main", 1920, 1080, 29.97, 10)
	c1.SetBGColor([3]uint8{20, 30, 40})
	c1.SetResolutionFactor(2, 2)

	c2, _ := p.NewComposition("BG", 1280, 720, 30, 5)
	if c2.ID == c1.ID {
		t.Errorf("ID collision: c1=%d c2=%d", c1.ID, c2.ID)
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatal(err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if len(re.Compositions) != 2 {
		t.Fatalf("re.Compositions = %d, want 2", len(re.Compositions))
	}
	main := findCompByName(re, "Main")
	if main == nil {
		t.Fatal("Main comp not found post-roundtrip")
	}
	if main.BGColor != [3]uint8{20, 30, 40} {
		t.Errorf("post-rt BGColor = %v, want [20 30 40]", main.BGColor)
	}
	if main.ResolutionFactor != [2]uint16{2, 2} {
		t.Errorf("post-rt ResolutionFactor = %v, want [2 2]", main.ResolutionFactor)
	}
	if math.Abs(main.FrameRate-29.97) > 1e-5 {
		t.Errorf("post-rt FrameRate = %g, want 29.97 (canonical)", main.FrameRate)
	}
}

func findCompByName(p *aep.Project, name string) *aep.Composition {
	for _, c := range p.Compositions {
		if c.Name == name {
			return c
		}
	}
	return nil
}
```

加 import `math`（如果还没）。

- [ ] **Step 2: Run**

```bash
go test -count=1 ./internal/aep/ -run 'TestNewComposition' -v
```
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/aep/new_composition_test.go
git commit -m "test(new_comp): Fields + Roundtrip (canonical fps + SetX after New)"
```

---

### Task 5.4: NewComposition rejection tests

**Files:**
- Modify: `internal/aep/new_composition_test.go`

- [ ] **Step 1: 写 reject 测试**

```go
func TestNewComposition_RejectsInvalid(t *testing.T) {
	cases := []struct {
		name        string
		fn          func(p *aep.Project) error
		wantInError string
	}{
		{"empty name", func(p *aep.Project) error {
			_, e := p.NewComposition("", 1920, 1080, 30, 5)
			return e
		}, "name cannot be empty"},
		{"zero width", func(p *aep.Project) error {
			_, e := p.NewComposition("x", 0, 1080, 30, 5)
			return e
		}, "size must be > 0"},
		{"zero height", func(p *aep.Project) error {
			_, e := p.NewComposition("x", 1920, 0, 30, 5)
			return e
		}, "size must be > 0"},
		{"zero fps", func(p *aep.Project) error {
			_, e := p.NewComposition("x", 1920, 1080, 0, 5)
			return e
		}, "frame rate must be > 0"},
		{"negative fps", func(p *aep.Project) error {
			_, e := p.NewComposition("x", 1920, 1080, -29.97, 5)
			return e
		}, "frame rate must be > 0"},
		{"zero duration", func(p *aep.Project) error {
			_, e := p.NewComposition("x", 1920, 1080, 30, 0)
			return e
		}, "duration must be > 0"},
		{"negative duration", func(p *aep.Project) error {
			_, e := p.NewComposition("x", 1920, 1080, 30, -1)
			return e
		}, "duration must be > 0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := aep.NewProject()
			err := tc.fn(p)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantInError) {
				t.Errorf("err = %q, want contains %q", err.Error(), tc.wantInError)
			}
			if len(p.Compositions) != 0 {
				t.Errorf("project polluted on failure: %d comps", len(p.Compositions))
			}
		})
	}
}
```

- [ ] **Step 2: Run**

```bash
go test -count=1 ./internal/aep/ -run 'TestNewComposition_RejectsInvalid' -v
```
Expected: 7 sub-tests PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/aep/new_composition_test.go
git commit -m "test(new_comp): rejection paths (7 invalid inputs)"
```

---

### Task 5.5: 混合 Open + NewComposition ID 测试 + empty Layr LIST invariant

**Files:**
- Modify: `internal/aep/new_composition_test.go`

- [ ] **Step 1: 写测试**

```go
func TestNewComposition_OnOpenedProject_NoIDCollision(t *testing.T) {
	p, err := aep.Open("../../test_data/re_batch.aep")
	if err != nil {
		t.Skipf("re_batch.aep not present: %v", err)
	}
	var maxOld uint32
	for _, c := range p.Compositions {
		if c.ID > maxOld {
			maxOld = c.ID
		}
	}
	for _, f := range p.Footage {
		if f.ID > maxOld {
			maxOld = f.ID
		}
	}

	c, err := p.NewComposition("Added", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if c.ID <= maxOld {
		t.Errorf("new comp ID %d should be > existing max %d", c.ID, maxOld)
	}
}

func TestNewComposition_EmptyLayrListPreserved(t *testing.T) {
	p := aep.NewProject()
	c, _ := p.NewComposition("EmptyL", 1920, 1080, 30, 5)
	if len(c.Layers) != 0 {
		t.Errorf("fresh NewComposition has %d layers, want 0", len(c.Layers))
	}

	// Roundtrip: AE should not auto-insert sentinel layer
	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatal(err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	for _, recomp := range re.Compositions {
		if recomp.Name == "EmptyL" && len(recomp.Layers) != 0 {
			t.Errorf("post-rt EmptyL has %d layers, want 0", len(recomp.Layers))
		}
	}
}
```

- [ ] **Step 2: Run**

```bash
go test -count=1 ./internal/aep/ -run 'TestNewComposition_OnOpenedProject|TestNewComposition_EmptyLayrList' -v
```
Expected: PASS.

- [ ] **Step 3: Run full test suite (regression check)**

```bash
go test -count=1 ./internal/aep/...
```
Expected: no FAIL（既有 109+ tests + 新加的 ~15 个 NewProject/NewComposition tests）。

- [ ] **Step 4: Commit**

```bash
git add internal/aep/new_composition_test.go
git commit -m "test(new_comp): Open+New ID monotonic; empty Layr preserved roundtrip"
```

---

# Phase 6 — AE 25 ship gate

### Task 6.1: `verify_v2_1.jsx`

**Files:**
- Create: `test_data/verify_v2_1.jsx`

- [ ] **Step 1: 写 JSX**

```jsx
// test_data/verify_v2_1.jsx
//
// Args (passed via AfterFX.exe -r script.jsx <inputAEP> <doneFile> <resavedAEP>):
//   $.global.v21Args = { input: ..., done: ..., resaved: ... }
// 由 Go test 在调用前用 BridgeTalk 或临时 .json file 注入。
// 简化路径：Go test 先写一个 args.json，JSX 从约定位置读：
//   e:/projects/tools/aep-parser/test_data/v2_1_args.json
//
// 内容：{"input": "...", "done": "...", "resaved": "..."}
//
// 这种"约定位置 args 文件"的取参方式比 AfterFX -r 命令行 args 更跨版本稳定。

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/v2_1_args.json");
    var args = {};
    var doneFile = null;
    var log = [];
    var ok = false;

    try {
        // 读 args
        argsFile.open("r");
        var s = argsFile.read();
        argsFile.close();
        // ExtendScript: 无 JSON.parse 在所有版本；用 eval 退路
        args = eval("(" + s + ")");
        doneFile = new File(args.done);

        var inFile = new File(args.input);
        app.openProject(inFile);
        log.push("opened " + inFile.fsName);
        log.push("items.length=" + app.project.items.length);
        ok = (app.project.items.length === 2);

        if (ok) {
            var c = app.project.items[1]; // 1-indexed; assumes first item is Main
            ok = (c.name === "Main"
                  && c.width === 1920 && c.height === 1080
                  && Math.abs(c.frameRate - 29.97) < 1e-5
                  && Math.abs(c.duration - 10) < 1e-3
                  && Math.abs(c.shutterAngle - 180) < 1e-3
                  && Math.abs(c.pixelAspect - 1.0) < 1e-5
                  && c.bgColor[0] === 20 && c.bgColor[1] === 30 && c.bgColor[2] === 40);
            log.push("Main: name=" + c.name + " " + c.width + "x" + c.height
                     + " fps=" + c.frameRate + " dur=" + c.duration
                     + " bg=[" + c.bgColor.join(",") + "]");
        }

        // Reopen-save-resave-reopen
        if (ok && args.resaved) {
            var resavedFile = new File(args.resaved);
            app.project.save(resavedFile);
            app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
            app.openProject(resavedFile);
            ok = (app.project.items.length === 2);
            var c2 = app.project.items[1];
            ok = ok && c2.name === "Main"
                 && Math.abs(c2.frameRate - 29.97) < 1e-5
                 && Math.abs(c2.duration - 10) < 1e-3
                 && c2.bgColor[0] === 20;
            log.push("resaved items.length=" + app.project.items.length
                     + " resaved-fps=" + c2.frameRate);
        }
    } catch (e) {
        ok = false;
        log.push("ERROR: " + e.toString());
    }

    // .done 必写出；try/catch 防 Go 端 hang
    try {
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_1_test.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {
        // 写不出 .done 算 Go 端 timeout，无法救
    }
})();
```

- [ ] **Step 2: Commit**

```bash
git add test_data/verify_v2_1.jsx
git commit -m "test(ship_gate): AE-side JSX driver (try/catch + reopen-save-reopen)"
```

---

### Task 6.2: Go test driver

**Files:**
- Modify: `internal/aep/new_composition_test.go`

- [ ] **Step 1: 写 ship-gate test**

```go
// （加到 new_composition_test.go）

func TestV2_1_AEShipGate_AE2025(t *testing.T) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE 2025 installed to run")
	}
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}

	// 1. New + WriteAEP → tempDir/v2_1_test.aep
	tempDir, err := filepath.Abs(t.TempDir())
	if err != nil { t.Fatal(err) }
	inputAEP := filepath.Join(tempDir, "v2_1_test.aep")
	resavedAEP := filepath.Join(tempDir, "v2_1_test.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_1_test.done")

	p := aep.NewProject(aep.TargetAE2025)
	main, err := p.NewComposition("Main", 1920, 1080, 29.97, 10)
	if err != nil { t.Fatal(err) }
	main.SetBGColor([3]uint8{20, 30, 40})

	_, err = p.NewComposition("BG_loop", 1920, 1080, 30, 5)
	if err != nil { t.Fatal(err) }

	out, err := os.Create(inputAEP)
	if err != nil { t.Fatal(err) }
	if err := p.WriteAEP(out); err != nil { t.Fatal(err) }
	out.Close()

	// 2. 写 args.json (固定路径 e:/projects/tools/aep-parser/test_data/v2_1_args.json)
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_1_args.json`
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		inputAEP, doneFile, resavedAEP)
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile) // 防上轮残留

	// 3. AfterFX -r verify_v2_1.jsx
	jsxPath := `e:/projects/tools/aep-parser/test_data/verify_v2_1.jsx`
	cmd := exec.Command(aeExe, "-r", jsxPath)
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting AE: %v", err)
	}

	// 4. 等 .done with timeout
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			t.Fatalf("timeout waiting for %s", doneFile)
		case <-ticker.C:
			if _, err := os.Stat(doneFile); err == nil {
				goto done
			}
		}
	}
done:
	content, err := os.ReadFile(doneFile)
	if err != nil { t.Fatal(err) }
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || lines[0] != "PASS" {
		t.Errorf("ship gate FAIL:\n%s", string(content))
	}
}
```

加 imports：

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
)
```

- [ ] **Step 2: Build (skip-run sanity)**

```bash
go test -count=1 ./internal/aep/ -run 'TestV2_1_AEShipGate_AE2025' -v
```
Expected: SKIP (because AE_SHIP_GATE not set).

- [ ] **Step 3: Commit**

```bash
git add internal/aep/new_composition_test.go
git commit -m "test(ship_gate): Go driver for AE 2025 ship gate"
```

---

### Task 6.3: 跑 AE 2025 ship gate（手动 / 半自动）

**Files:** none

- [ ] **Step 1: 跑 ship gate**

```bash
AE_SHIP_GATE=1 go test -count=1 ./internal/aep/ -run 'TestV2_1_AEShipGate_AE2025' -v
```
Expected: PASS。若 FAIL，看 `.done` 内容 + 调整 builder（fdta / idpc / idta / cdta 等 RE finding 没落实的地方）。

- [ ] **Step 2: 添加 AE 2020 版本**

复制 step 1 的 test，改 `aeExe` 默认 path 到 AE 2020 + 改 target 到 `TargetAE2020`。

```go
func TestV2_1_AEShipGate_AE2020(t *testing.T) {
	// 同上，但：
	//   aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe` (or AE2020_EXE env)
	//   p := aep.NewProject(aep.TargetAE2020)
	// 其它一致
}
```

- [ ] **Step 3: 跑 AE 2020**

```bash
AE_SHIP_GATE=1 go test -count=1 ./internal/aep/ -run 'TestV2_1_AEShipGate_AE2020' -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/aep/new_composition_test.go
git commit -m "test(ship_gate): AE 2020 + AE 2025 cross-version gates pass"
```

---

# Phase 7 — Docs sync

### Task 7.1: 用户文档 (docs/)

**Files:**
- Modify: `docs/composition.md`
- Modify: `docs/project.md`

- [ ] **Step 1: docs/composition.md 加 NewComposition 段**

```bash
grep -n '^### \|^## ' docs/composition.md | head -10
```

在合适位置（"Attributes" 段之前或"创建 / 加载"段，按既有结构）加：

```markdown
### Project.NewComposition

```go
func (p *Project) NewComposition(name string, width, height uint16, frameRate, duration float64) (*Composition, error)
```

在 project 根目录新建空 composition（无图层）。Required 参数：name / width / height / frameRate / duration，任一不合法返 error。可选字段（BGColor / PixelAspect / ResolutionFactor / Shutter*）走既有 `Set*` 方法。

ID 自动分配（monotonic，never reuse）。新 comp append 到 root Fold，跟 `Open(...)` 出来的 comp 同构 — 所有 `Set*` 方法立即可用。

```go
proj := aep.NewProject()
main, err := proj.NewComposition("Main", 1920, 1080, 29.97, 10)
if err != nil { log.Fatal(err) }
main.SetBGColor([3]uint8{20, 30, 40})
main.SetResolutionFactor(2, 2)
```
```

- [ ] **Step 2: docs/project.md 加 NewProject 段**

```markdown
### aep.NewProject

```go
func NewProject(target ...AETarget) *Project
```

返回一个全新 empty Project，可以接续调 `NewComposition` 等添加内容。零参数 = `TargetAE2020` 默认（最大兼容）。支持显式指定 `TargetAE2020 / TargetAE2022 / TargetAE2025`。

输出 .aep 用对应 AE 版本格式标识；newer AE 自动 upgrade，older AE 拒开。`AETarget` 值不向前兼容 —— 升库时旧二进制传 unknown target 会 panic。

**Never returns error** —— 嵌入 template 是 build-time trusted；如果 panic 出 "build bug" 信息，那是库自己的 bug，不是用户输入问题。

```go
proj := aep.NewProject()                      // AE 2020 兼容
proj25 := aep.NewProject(aep.TargetAE2025)    // AE 25 格式
```
```

- [ ] **Step 3: Commit**

```bash
git add docs/composition.md docs/project.md
git commit -m "docs: NewProject + NewComposition API reference"
```

---

### Task 7.2: workshop/board.md 归档

**Files:**
- Modify: `workshop/board.md`

- [ ] **Step 1: 在最近归档段加一条**

```markdown
### 2026-05-22 V2.1 Foundation ship — NewProject + NewComposition (109 → 109+N PASS)

V2 第一个 sub-project：从零创建 .aep 通过 AE 2020 / 2025 ship gate。

- **Public API**: `aep.NewProject(target ...AETarget) *Project` + `proj.NewComposition(name, w, h, fps, duration) (*Composition, error)` + `AETarget` enum (2020/2022/2025)
- **架构**: hybrid template embed（3 份 AE-saved empty 各 ~8 KB）+ parse 闭环（New 出 comp 走 parseComposition 跟 Open 同构）
- **Plan**: workshop/plans/v2-1-foundation-plan.md（8 phases，~30+ tasks，含 4 个 Day-1 RE prerequisites）
- **新文件**: new_project.go / new_composition.go / cdta_layout.go / framerate_canonical.go / templates/{2020,2022,2025,2025_dummy_comp}.aep + 测试 / verify_v2_1.jsx / golden fixture
- **修改**: Project.nextItemID + rootFold (types_core) / parseProject 加 initDerived (parse) / write_composition.go 迁 hex offsets 到 cdta_layout.go
- **教训**: （从 RE-1/2/3/4 findings 提炼，按实际情况补）
```

- [ ] **Step 2: 更新 Last updated + PASS count**

```markdown
**Last updated**: 2026-05-22 by claude (V2.1 ship — NewProject + NewComposition + AE 2020/2025 ship gates PASS；总 PASS = <实际数>)
**Active focus**: 🟢 V2.1 完工 —— 下个 sub-project V2.2 (Solid/Camera/Light layer 创建) 待开
```

- [ ] **Step 3: Commit**

```bash
git add workshop/board.md
git commit -m "docs(board): archive V2.1 Foundation ship"
```

---

### Task 7.3: coverage.md / coverage-detail.md

**Files:**
- Modify: `workshop/plans/coverage.md`
- Modify: `workshop/plans/coverage-detail.md`

- [ ] **Step 1: coverage.md 加 V2 段**

```markdown
## V2 进度（结构性写）

- ✅ V2.1 Foundation (2026-05-22): `aep.NewProject` + `Project.NewComposition`
- ⚪ V2.2 Layer creation (Solid/Camera/Light)
- ⚪ V2.3 Footage import + 引用型 Layer
- ⚪ V2.4 属性树结构性扩展 (AddEffect / AddMask)
- ⚪ V2.5 Text + Shape Layer 合成
- ⚪ V2.6 Marker / Folder / 收尾
```

- [ ] **Step 2: coverage-detail.md 在 CompItem / Project 段加 New* API**

按既有表格结构添加：

```markdown
| `app.project` （new from scratch）| `aep.NewProject(target ...AETarget)` | ✅ R/W | V2.1，三 target AE 2020/2022/2025 |
| `app.project.items.addComp` | `Project.NewComposition(name, w, h, fps, dur)` | ✅ R/W | V2.1，atomic mutation |
```

- [ ] **Step 3: Commit**

```bash
git add workshop/plans/coverage.md workshop/plans/coverage-detail.md
git commit -m "docs(coverage): V2.1 NewProject + NewComposition ship"
```

---

### Task 7.4: 最终验证 + 状态确认

**Files:** none

- [ ] **Step 1: 跑全套 + vet**

```bash
go vet ./...
go test -count=1 ./internal/aep/...
echo "---PASS---"
go test -count=1 ./internal/aep/... -v | grep -c '^--- PASS'
echo "---FAIL---"
go test -count=1 ./internal/aep/... -v | grep -c '^--- FAIL'
```
Expected: vet clean / 109+N PASS / 0 FAIL（N = 新加的测试数，~20+）。

- [ ] **Step 2: AE ship gate (按需，需要本地 AE)**

```bash
AE_SHIP_GATE=1 go test -count=1 ./internal/aep/ -run 'TestV2_1_AEShipGate' -v
```
Expected: 2 ship gate tests PASS（AE 2020 + AE 2025）。

- [ ] **Step 3: 整理 tmp_debug 工具**

```bash
ls tmp_debug/dump_fdta/  # 留作长期 RE 工具
# 不删，跟既有 dump_cdta / list_props 等同列
```

- [ ] **Step 4: 写 V2.1 完成 commit**

```bash
git log --oneline | head -30  # 看 V2.1 整体提交链
# 如果需要 squash 或加 tag，按 git playbook 决
```

---

## 自审 Checklist

实现 V2.1 plan 全部 task 后核：

- [ ] `aep.NewProject` + `Project.NewComposition` 两个 public API ship 完
- [ ] AETarget 三个值（2020/2022/2025）都有对应 embed template
- [ ] AE 2020 + AE 2025 ship gate 都 PASS
- [ ] 12 条 Invariants 全部体现在代码：单调 ID / atomic rollback / warnings-as-failure / unknown chunks pass-through / etc.
- [ ] 4 个 RE prerequisites 都有 finding 记录（plan 内 + 必要时 scars/）
- [ ] 总测试 PASS = 109 (baseline) + ~20 (new V2.1 tests) = ~130
- [ ] 文档同步：CLAUDE / board / coverage / coverage-detail / docs/composition / docs/project
- [ ] golden fixture (`test_data/comp_item_children.golden.txt`) 跟实际 builder 输出一致
