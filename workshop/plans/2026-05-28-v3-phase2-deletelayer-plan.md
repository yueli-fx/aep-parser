# V3 Phase 2 — Composition.DeleteLayer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Spec**: [`../specs/2026-05-27-v3-deep-think.md`](../specs/2026-05-27-v3-deep-think.md) § 4 Phase 2 candidate
**Predecessor**: V3 Phase 1 (2026-05-28, commits `1b7ff63`…`2bebbd7`) — `back *<type>Backrefs` shard 全部就位，PASS 259。

**Goal**: 实现 `Composition.DeleteLayer(idx int) error` —— 第一个走 V3 scene+back 拆分的结构性 mutation API。**必须通过 AE 2020 + AE 2025 双版本 ship-gate**（CLAUDE.md 硬约束 #6）才算 ship。

**Architecture**: V2.1 atomic invariants 全套复用（snapshot → mutate → warnings-as-failure → rollback；详 [scars/ae25-acceptance-gate.md](../scars/ae25-acceptance-gate.md)）。 Scene 侧 `c.Layers[idx]` 切片 splice，back 侧 `c.back.itemList.Children` 删 Layr + 紧邻的 Ewst sibling。引用清理（ParentID / TrackMatteLayerID 孤儿）走 RE-driven 策略 —— **不要瞎猜 AE 行为，先用 fixture diff 决定**。

**Tech Stack**: Go 1.x, `internal/aep` single-package convention (CLAUDE.md #3), V2.1 atomic invariant pattern (Inv-10/-11)。

**Non-goals (Phase 2)**:
- 不实现 `InsertLayer` / `DuplicateLayer` / `MoveAfter` / `Reparent` —— 这些是 Phase 3+。
- 不实现 `RemoveEffect` / `DeleteProperty` —— 不同层级，Phase 4+。
- Phase 1 留的 `opaque` 字段在 Phase 2 *可能* 启用（取决于 RE 结果 § Task 2）。

---

## 关键 RE 未知

写代码前必须用 AE-saved fixture 回答（Task 1-3 解决）：

| Q | 问题 | 影响代码路径 |
|---|---|---|
| RE-Q1 | AE delete 后 itemList 是「splice (后面层全前移)」还是「gap (位置 hold)」？ | 决定 `c.back.itemList.Children` 删法 |
| RE-Q2 | 被删层是某层的 parent — AE 把 child.ParentID 改 0 还是改 child.parent.parent.ID（祖父级 reparent）？ | child layer reference 清理 |
| RE-Q3 | 被删层是某层的 track matte source — AE 怎么处理孤儿 TrackMatteLayerID？ | 同上 |
| RE-Q4 | 被删层的 Ewst sibling 也被删掉？（极大概率 yes per Phase 1 RE 注释）| 确认双删 |
| RE-Q5 | AE 删层后 head chunk 的 nextItemID counter 变不变？| `syncHeadCounters` 是否需要 down-adjust（Invariant #9 monotonic 说不能，但 RE 验一下）|
| RE-Q6 | 还有别处引用被删层 ID 吗（render queue / essential graphics / 表达式 string）？| 是否需要 string-level scrub（成本极高，先看是否真有）|

**RE 不到的话**：保守策略 = 拒绝带 children / matte refs 的 delete，逼调用方先 reparent 自己。先看 Q1-Q4，Q5-Q6 看时间。

---

## Pre-flight check

从 `e:/projects/tools/aep-parser` 跑：

```powershell
go vet ./...                                                              # clean
go test -count=1 ./internal/aep/... 2>&1 | Select-String '^--- PASS' | Measure-Object | Select-Object -ExpandProperty Count   # expect 259
git log --oneline -1                                                      # expect docs(workshop): V3 Phase 1 complete
git status --short                                                        # expect clean
```

不匹配 → STOP，先 reconcile board.md。

---

## Task 1: RE harness — JSX driver + AE fixture spec

**Goal**: 产出 4 个 AE-saved fixture，覆盖 RE-Q1..Q4。User 跑 AE，Go 这边 parse_btdk 工具拿 diff。

**Files (create):**
- `test_data/re_delete_layer.jsx` — fixture 生成器
- `test_data/re_delete_layer_baseline.aep` — pre-deletion (AE 2025 saved)
- `test_data/re_delete_layer_middle.aep` — middle layer deleted
- `test_data/re_delete_layer_parent.aep` — has-children parent deleted（trigger Q2）
- `test_data/re_delete_layer_matte.aep` — track-matte source deleted（trigger Q3）

**JSX 内容大纲**（`re_delete_layer.jsx`）：

```javascript
// 单 JSX，多 baseline mode 切换 via $.evalFile arg or env
var mode = $.global.RE_DELETE_MODE || "baseline";
var proj = app.newProject();
var comp = proj.items.addComp("Test", 1920, 1080, 1, 5, 24);
// 三个 solid layer，layer1 (顶层) parent layer2 layer3 (底层)
var l1 = comp.layers.addSolid([1,0,0], "L1_top", 100, 100, 1);
var l2 = comp.layers.addSolid([0,1,0], "L2_mid", 100, 100, 1);
var l3 = comp.layers.addSolid([0,0,1], "L3_bot", 100, 100, 1);
l1.parent = l2;
l2.trackMatteType = TrackMatteType.ALPHA;  // l2 上的 track matte source = l3 (下一层)
if (mode === "baseline") {
    // 不动，存
} else if (mode === "middle") {
    l2.remove();  // 删中间，看 Layr gap 还是 splice + 看 l1.parent 怎么变
} else if (mode === "parent") {
    l2.remove();  // 删 parent（l1 失去 parent）
} else if (mode === "matte") {
    l3.remove();  // 删 matte source（l2 失去 track matte ref）
}
proj.save(new File("test_data/re_delete_layer_" + mode + ".aep"));
app.quit();
```

- [x] **Step 1.1**: 写 `test_data/re_delete_layer.jsx`（多 mode via `$.getenv("RE_DELETE_MODE")`；per-mode setup 隔离 RE-Q1..Q4；teardown close + quit；done marker）。layout: `l1`=index 1, `l2`=index 2, `l3`=index 3 (add 反序)；matte 走 implicit "layer above"（AE 2020 兼容），AE 23+ 自动写显式 ID。
- [x] **Step 1.2**: Agent 自己跑 — baseline/middle/parent 用 **AE 2020 17.7x45**，matte 用 **AE 2025 25.1x68**（因 `TrackMatteLayerID` 是 AE 23+ 字段）。所有 4 fixture exit=0 PASS。
- [x] **Step 1.3 (tool ready)**: `tmp_debug/dump_layers/main.go` 已写好 + 4 fixture dump 过 — Item LIST baseline=237 children, middle/parent/matte=221 (Δ-16)。
- [x] **Step 1.4**: scar `workshop/scars/ae-deletelayer-re.md` 落地（含 6 个 Finding + impl rules）。**关键结果**：(F1) splice 不 gap；(F2) ParentID reset to 0；(F3) TrackMatteLayerID reset 但 trackMatteType byte 不动；(F4) 每 layer = 16 chunk delete unit（Layr + Ewst + 14 follower fvdv/fiop/ftts/foac/fiac/fipc/fifl ×2）；(F5) DLay/SLay/CLay/SecL 同样 16-chunk pattern 但 DeleteLayer 拒接；(F6) 不复用 ID（head counter 不动）。

**完成判定**: scar 文件包含 RE-Q1..Q4 的 AE-验证答案，每条带 fixture name + Go-端 dump 引用。

---

## Task 2: 选 implementation strategy

**Goal**: 基于 Task 1 的 RE 答案，定下 DeleteLayer 的策略矩阵。**不写代码**，写决策。

- [ ] **Step 2.1**: 把 RE-Q1..Q6 答案塞进决策表（写在 plan 末尾，或新建 spec 文件 `specs/2026-05-28-v3-phase2-deletelayer-strategy.md`）。表至少含：
    - `itemList.Children` 删法（splice / gap）
    - Ewst sibling 同删 / 留
    - Layer.ParentID 清理：reset to 0 / lift to grandparent / refuse delete
    - TrackMatteLayerID 清理：reset to 0 / re-pick neighbor / refuse delete
    - head chunk counter 调整：none / down-adjust
    - opaque shard 启用：no / yes（若 RE 发现孤儿 chunks）
- [ ] **Step 2.2**: 决策"refuse delete"的所有场景表（unsupported 状态），落到 error message：
    - 比如 layer 0 (camera/light/audio)? 索引越界？ 单层 comp 删剩 0 个 AE 接受吗？
- [ ] **Step 2.3**: 决策被 commit/board 接受后，进 Task 3

---

## Task 3: 实现 DeleteLayer

**Files:**
- Create: `internal/aep/delete_layer.go`
- Modify: `internal/aep/new_layer.go` 如需共享 helper (e.g. `findLayrIndexInItemList`)
- Tests: `internal/aep/delete_layer_test.go`（白盒 + fixture round-trip）

**Signature**（确认 in Task 2，可微调）：

```go
// DeleteLayer removes the layer at the given 1-based index (matching AE
// Scripting's Layer.index 1-based convention; **note** Go-side
// Composition.Layers is 0-based slice — we accept 1-based for API
// symmetry with NewShapeLayer-future return values).
//
// Atomic invariants (V2.1 pattern):
//   - Snapshot pre-mutation state
//   - Mutate itemList.Children + comp.Layers + (any reference fixups)
//   - If any parser warnings appeared → full rollback + return error
//   - If no warnings → commit (return nil)
//
// AE acceptance: this is a structural write, **requires AE 2020 + AE
// 2025 ship-gate** (CLAUDE.md hard constraint #6) before public API
// promotion to Stable. Until ship-gate green, this is alpha.
func (c *Composition) DeleteLayer(index int) error
```

- [ ] **Step 3.1**: Snapshot helper（slice copy of `c.back.itemList.Children` + `c.Layers` + `proj.Warnings`）
- [ ] **Step 3.2**: Locate Layr LIST + Ewst sibling in itemList.Children（per Task 2 decision; share helper with NewShapeLayer if applicable）
- [ ] **Step 3.3**: Reference cleanup per Task 2 decision matrix（ParentID / TrackMatteLayerID / SourceID if matte）
- [ ] **Step 3.4**: Remove Layr (+ Ewst) from itemList; remove from `c.Layers`
- [ ] **Step 3.5**: Re-parse closed loop? Optional — `NewShapeLayer` 不 re-parse 但走 `syncShapeLayerChunks`。DeleteLayer 不动生成内容，可能不需要 re-parse。Task 2 决定。
- [ ] **Step 3.6**: Warnings-as-failure check：若 `len(proj.Warnings) > oldWarningsLen` → rollback + return error
- [ ] **Step 3.7**: 加 unit tests（覆盖 Task 2 决策的 refuse-cases，覆盖 happy path round-trip）

---

## Task 4: Round-trip + AE round-trip 测试

**Goal**: 证明 DeleteLayer 后 WriteAEP 产物是 valid AEP。

- [ ] **Step 4.1**: Go-端 round-trip: `Open → DeleteLayer(mid) → WriteAEP → Reopen → 校验 layers 数量 + 残留 layer 字段不变`
- [ ] **Step 4.2**: Byte-equivalence: `Open(re_delete_layer_baseline.aep) → DeleteLayer 删中间 → WriteAEP` 跟 `re_delete_layer_middle.aep` 字节级别 ≈ 一致（容忍头部 timestamp / random UUID 不一致；其余必须一样）。**这是 RE-driven validation：Go 端模仿 AE delete 输出。**
- [ ] **Step 4.3**: 类似步骤 4.2 跑 parent / matte 三种 mode
- [ ] **Step 4.4**: PASS count 增加 (PASS 259 + N)；FAIL=0；vet clean

---

## Task 5: AE 2020 + 2025 ship-gate (USER-SIDE)

**这是 user-action，不是 Go-side。** 走 `scripts/ae_run.ps1` wrapper。

- [ ] **Step 5.1**: Go-端写产 4 个 modified-by-Go fixtures（baseline AEP → DeleteLayer → WriteAEP → save as `test_data/ge_delete_layer_<mode>.aep`，`ge_` 前缀 = "Go-emitted"）
- [ ] **Step 5.2**: User 用 `scripts/ae_run.ps1` 让 AE 2025 打开每个 ge_ fixture，检验 (a) AE 不崩；(b) AE 不报"文件数据丢失"；(c) layer 数量/类型符合预期
- [ ] **Step 5.3**: 重复 5.2 但用 AE 2020
- [ ] **Step 5.4**: 若任一步骤失败，记 mode 信号到 [scars/ae25-acceptance-gate.md](../scars/ae25-acceptance-gate.md)，回 Task 2 重新决策

**完成判定**: 4 个 mode × 2 个 AE 版本 = 8 次打开全 PASS。

---

## Task 6: Docs sync + 升级 API 等级

- [ ] **Step 6.1**: 把 `DeleteLayer` 从 alpha 升 Stable（已通过双版本 ship-gate per CLAUDE.md #2）。更新 source code 注释，去掉 alpha 标记。
- [ ] **Step 6.2**: 更新 `workshop/board.md`:
    - Recently finished 加 V3 Phase 2 完成行
    - Next session 改为 V3 Phase 3 候选（InsertLayer / DuplicateLayer / Layer.Remove on Project / capability matrix start）
    - Active focus 改 Phase 3 ready
- [ ] **Step 6.3**: 更新 `workshop/plans/coverage.md` — DeleteLayer 在 Composition mutations 表加一行 ✅
- [ ] **Step 6.4**: 若 Task 1 产生了 scar `ae-deletelayer-re.md`，确认 frontmatter 完整（`when_to_read` / `applies_to` / `last_updated`）
- [ ] **Step 6.5**: 把本 plan 移到 `workshop/plans/finish/`
- [ ] **Step 6.6**: 总 commit message：`feat(aep): V3 Phase 2 — Composition.DeleteLayer (AE 2020 + 2025 ship-gate green)`

---

## Acceptance criteria

全部同时满足：

1. `go vet ./...` 清 + `go test -count=1 ./internal/aep/...` PASS = 259 + N (N ≥ 4, Task 4 加的)
2. AE 2020 + AE 2025 各打开 4 mode 的 ge_*.aep 全 PASS（Task 5）
3. Public API: `Composition.DeleteLayer(int) error` 出现在 godoc，文档完整
4. board.md `Last updated` 翻到 Phase 2 完成日期，Recently finished 加新行

---

## Open questions / risk register

| Risk | Mitigation |
|---|---|
| RE-Q1..Q4 RE 不出来（AE behavior 反直觉）| Task 2 选保守策略：refuse delete on parent/matte refs |
| AE silent-drop (Mode 3) 触发 — Go-emitted fixture AE 报"文件数据丢失" | bisect feedback loop — 削减改动直到 AE 接受，再加回（[scars/ae25-acceptance-gate.md](../scars/ae25-acceptance-gate.md) Stage 4 pattern）|
| nextItemID counter 必须 down-adjust（Q5 yes）— 但 CLAUDE.md Invariant #9 说 monotonic 不 reuse —— 矛盾 | 优先 invariant #9（不 down-adjust）；若 AE 拒绝就在 scar 里记 invariant 例外条件 |
| 表达式 string 里硬编码引用被删 layer ID — round-trip 后 Go-emitted 还含孤儿 ID | 接受 (out-of-scope) — 用户自己负责清理，加 doc warning |
| `Project.Warnings` rollback 漏 partial mutation | Snapshot 全场快照 (proj.Warnings + comp.Layers + itemList.Children + 任何 reference 修改对象的旧值)，rollback 时全恢复 |

---

## Related

- Spec: [`../specs/2026-05-27-v3-deep-think.md`](../specs/2026-05-27-v3-deep-think.md) (§ 4 Phase 2 candidate; § 3 risk register)
- Phase 1 plan: [`finish/2026-05-28-v3-phase1-backrefs-plan.md`](finish/2026-05-28-v3-phase1-backrefs-plan.md)
- V2.1 atomic invariant pattern: `internal/aep/new_layer.go::NewShapeLayer` (lines 110-211)
- Ship-gate playbook: [`../playbooks/re-fixture.md`](../playbooks/re-fixture.md) § "何时启用 ship-gate" + § "GDI 自动化 — `scripts/ae_run.ps1`"
- CLAUDE.md hard constraints: #1 (length-preserving + V2.1 atomic), #2 (Stable vs Alpha API), #5 (opaque preservation — 取决于 Task 2 决策), #6 (AE 接受 gate — 强制双版本)
- AE acceptance gate scar: [`../scars/ae25-acceptance-gate.md`](../scars/ae25-acceptance-gate.md)
