# Board — aep-parser

**Last updated**: 2026-05-28 by claude (V3 Phase 3 Task 5 完 — ship-gate 6/6 PASS; DuplicateLayer ready for Stable promotion)
**Active focus**: V3 Phase 3 Task 6 (godoc alpha → Stable + finish/ migration + commit)。Plan: `plans/2026-05-28-v3-phase3-duplicatelayer-plan.md` Task 6。

## Next session

**V3 Phase 3 Task 6 — Stable promotion + finish/ migration**：
1. `internal/aep/duplicate_layer.go` godoc: 去掉 "Alpha: AE 2020 + AE 2025 ship-gate pending" 一行，标记 Stable
2. `workshop/plans/coverage.md`: + DuplicateLayer row (Layer.Duplicate via Composition.DuplicateLayer)
3. 移 `plans/2026-05-28-v3-phase3-duplicatelayer-plan.md` → `plans/finish/`
4. 移 `specs/2026-05-28-v3-phase3-duplicatelayer-strategy.md` → `specs/finish/`
5. Commit: `feat(aep): V3 Phase 3 ship-gate green — DuplicateLayer Stable`

**Task 5 产物（已 ship 6/6 PASS）**:
- `tmp_debug/ge_duplicate_layer/main.go` — 3 mode ge_*.aep (solo/dup_parent/dup_child) producer; baseline = re_delete_layer_baseline.aep
- `test_data/verify_ge_duplicate_layer.jsx` — mode 切换 (`$.getenv("GE_DUP_MODE")`), 验 numLayers==4 + clone identity + parent expectations per mode
- AE 2020 × 3 + AE 2025 × 3 全 PASS。Key 行为验证：
  - **solo**: clone L2_clone @ AE idx 2, source L2_mid pushed to idx 3
  - **dup_parent**: F6 confirmed — L3 still parents to L2_mid (orig at idx 4), NOT clone at idx 2
  - **dup_child**: F7-inverse confirmed — both L1_clone (idx 1, clone) and L1_top (idx 2, source) show parent=L2_mid
- matte mode by-design refused per strategy spec §5 — no ge fixture, no AE verify (expected, not coverage gap)

**Task 3 产物（已 ship）**:
- `internal/aep/duplicate_layer.go` (~210 LOC) — `DuplicateLayer(int, string) (*Layer, error)`，按 spec §3 algorithm 12 步实现。reuse `deepCloneChunk` from new_composition.go（不要重写！）；reuse `findLayrIndexInItemList` from delete_layer.go (同 package 可直接调)。选项 a — 重跑 `parseLayer(clonedLayr, index, ctx)` 建 cloneLayer with backrefs into cloned chunks。
- `internal/aep/duplicate_layer_test.go` (9 tests, all PASS): RefuseEmptyName / RefuseOutOfRange / RefuseMissingBackref / RefuseNonAV / RefuseTrackMatte / HappyPath_Middle (验 ID alloc + name + 4 layers + delta=16 + nextItemID bumped) / FreshDataSlices (concurrent-mutate scar 防护) / RoundTrip (Open→Dup→Write→Reopen) / LdtaBodyVerbatim (F10 字节级验证：clone 仅 @0x00..0x03 变)
- PASS 267→277 (+10)；FAIL=0；vet clean (info-level 是 pre-existing 与 task 无关)
- **Stable promotion 在 Task 6** (godoc 现标 alpha；ship-gate green 后转 Stable)

**Task 2 产物（已 ship）**:
- `workshop/specs/2026-05-28-v3-phase3-duplicatelayer-strategy.md`（9 sections：signature / decision matrix / clone algorithm / no-ref-cleanup / refuse-cases / atomic / deferred / ship-gate / file map）
- 镜像 DeleteLayer strategy 结构；scar 10 Findings 全部翻译成 impl 决策；ship-gate target 6/6（matte 模式 by design 拒接，文档化为 expected）

**为什么 DuplicateLayer 起手**: smallest scope of V3 Phase 3 候选；最大复用 Phase 2 工具链；反向验「16-chunk delete unit」对 follower chunks 的理解 —— Finding 10 已 nail down: 14 follower 全 byte-identical clone，AE 不 mutate 任何 follower 字段。

**Phase 2 完整产物**（reference，git ship）：
- 代码: `internal/aep/delete_layer.go` + `_test.go` (9 test, 8 PASS 1 SKIP)
- 工具: `tmp_debug/dump_layers/main.go` + `tmp_debug/ge_delete_layer/main.go`
- JSX: `test_data/re_delete_layer.jsx` (5 modes) + `test_data/verify_ge_delete_layer.jsx`
- 文档: `scars/ae-deletelayer-re.md` + `specs/finish/2026-05-28-v3-phase2-deletelayer-strategy.md` + `plans/finish/2026-05-28-v3-phase2-deletelayer-plan.md`
- Playbook 刷新: re-fixture.md + verify.md

**并行 R-only 仍 deferred**（不阻塞 V3）：
- **Gradient W**: XML 重序列化 / SetGradient / per-keyframe gradients — 需 fixture
- **DisplayColorSpace R**: separate chunk 位置未 RE
- **ValueText**: per-type formatter — P3

**并行候选**：V2.2.1 ShapeLayer 拓展（其实最好压到 V3 Phase 5 一起做，避免 alpha API 重复）。

**并行 R-only 仍 deferred**（不阻塞 V3）：
- **Gradient W**: XML 重序列化 / SetGradient / per-keyframe gradients — 需 fixture
- **DisplayColorSpace R**: separate chunk 位置未 RE
- **ValueText**: per-type formatter — P3

**并行候选**：V2.2.1 ShapeLayer 拓展（其实最好压到 V3 Phase 5 一起做，避免 alpha API 重复）。

## In flight

无。

## Blockers

无。

## Deferred

- **V2.2.1 ShapeLayer 拓展**（Ellipse/Path/Stroke embed bytes / Fill Color 编码 RE / keyframe 持久化）— 跟 py-aep parity 并行；用户用 AE create fixture 后可起
- **V3 capability framework** — Layer.Remove/Duplicate/Move/PropertyBase 结构性 ops 都靠这套；进 Phase 3
- `environmentLayer` 360° 素材 / `ligature` OT liga 字体 / `maskFeatherFalloff` 位置未 RE
- Composition.SetRenderer / ldta 零值区 probe / Footage proxy / Project nhed 扩展 — 见 [coverage.md](plans/coverage.md) "剩余可探方向"
- **`linearizeWorkingSpace` ScriptingAPI quirk** — chunk byte 跟 AE 自己写一致但 ScriptingAPI 读不到 true，归 OCIO/CMS-联动；详 [scars/project-flag-chunks-lnrb-lnrp.md](scars/project-flag-chunks-lnrb-lnrp.md)

## Recently finished

- **2026-05-28 V3 Phase 3 Task 5 完 — ship-gate 6/6 PASS** — `tmp_debug/ge_duplicate_layer/main.go` 产 3 mode ge_*.aep (solo/dup_parent/dup_child)，baseline 复用 Phase 2 的 `re_delete_layer_baseline.aep`（3-solid pre-mutation；dup_parent/dup_child 用 Go SetParent 设引用再 dup，省一套 RE fixture）。`test_data/verify_ge_duplicate_layer.jsx` mode 切换验 (a) AE 不拒接 (b) numLayers==4 (c) clone 在 expected AE idx + 名字 (d) parent 关系 per mode。**6/6 PASS** (AE 2020 × 3 + AE 2025 × 3): solo 验 clone L2_clone @ idx 2 + 推 source L2_mid 到 idx 3；dup_parent 验 F6 — L3 仍 parents to L2_mid (orig) 不是 clone；dup_child 验 F7-inverse — L1_clone + L1_top 都 parents to L2_mid。matte mode 按 spec §5 by-design refuse，不产 ge fixture（expected absence，非 coverage gap）。Next: Task 6 godoc alpha → Stable + plan/spec → finish/ + commit。
- **2026-05-28 V3 Phase 3 Task 4 完 — Structural equivalence test PASS first try** — `TestDuplicateLayer_StructuralEquivalence_Solo` in `duplicate_layer_test.go`: opens `re_delete_layer_baseline.aep` (3-solid baseline) → `DuplicateLayer(1, "L2_mid")` → compares Composition[0].itemList.Children (ID + IsList + FormType, in order) vs AE's `re_duplicate_layer_solo.aep`. **PASS 一次过** —— Go-side dup 产的 itemList chunk shape 与 AE solo post-dup 完全对齐（chunk 数 + 每个 child 的 ID/IsList/FormType）。这是 ship-gate 前最强的"Go 实现行为对的"信号——AE 接受的 layout pattern (F1/F4/F5/F10) 我们 byte 不一致但 shape 一致，符合 length-preserving + verbatim deep-clone 预期。helper 加 `openDupFixture(t, mode)` 复用其他 mode。PASS 277→278；FAIL=0；vet clean。Next: Task 5 ship-gate (agent-side `tmp_debug/ge_duplicate_layer` + `verify_ge_duplicate_layer.jsx` × 3 mode × 2 version = 6 次 AE 开盘验证)。
- **2026-05-28 V3 Phase 3 Task 3 完 — DuplicateLayer 实现 + 9 tests PASS (alpha)** — `internal/aep/duplicate_layer.go` (~210 LOC) 按 spec §3 algorithm 12 步直翻：validate → snapshot (含 `proj.nextItemID`) → locate source Layr → defensive Ewst assert → adaptive block scan → deep clone block → alloc newID + 写 ldta @0x00..0x03 → rewrite Utf8 name → splice itemList @ srcLayrIdx → re-parse clonedLayr (选项 a — backrefs auto-bound to cloned chunks) → splice c.Layers @ index → warnings-as-failure rollback。**关键复用**：`deepCloneChunk` 已存在于 `new_composition.go`（不要重写！我第一次写时撞了 DuplicateDecl）；`findLayrIndexInItemList` 已存在于 `delete_layer.go` 同 package 可直接调用。**9 tests all PASS**: 5 refuse (EmptyName/OutOfRange/MissingBackref/NonAV/TrackMatte) + 4 happy (HappyPath_Middle 验 ID+name+count+delta=16+nextItemID bump / FreshDataSlices 验 concurrent-mutate scar 防护 / RoundTrip Open→Dup→Write→Reopen / LdtaBodyVerbatim F10 字节级验 clone 仅 @0x00..0x03 变)。**PASS 267→277** (+10 含子测试)；FAIL=0；vet clean。Alpha 标在 godoc — ship-gate (Task 5) green 后 Task 6 转 Stable。Next: Task 4 structural equivalence (vs `re_duplicate_layer_solo.aep`) + Task 5 ship-gate (`tmp_debug/ge_duplicate_layer` + `verify_ge_duplicate_layer.jsx` 跑 AE 2020/2025 × 3 mode = 6 次).
- **2026-05-28 V3 Phase 3 Task 2 完 — DuplicateLayer 策略矩阵落 spec** — `specs/2026-05-28-v3-phase3-duplicatelayer-strategy.md` (9 sections，镜像 DeleteLayer strategy 结构). 决策表把 scar 10 个 Finding 全翻译成 impl 规则：signature `(c *Composition) DuplicateLayer(index int, name string) (*Layer, error)` (F9 explicit name)；插入位置 = `index`（push source down，F1 的 3/4 简单情况）；ID = `proj.allocItemID()` (F3)；clone = adaptive 16-chunk deep copy (F5+F10)，每个 Data slice fresh `append([]byte(nil), src.Data...)` (concurrent-mutate scar)；mutate ONLY ldta @0x00..0x03 = newID (F10)；SourceID/ParentID/TrackMatteLayerID/TrackMatte 全 copy from source (F4/F7/F8)；name 用 caller-supplied 走 length-variable Utf8 写入。**No-reference-cleanup pass**：spec §4 显式拒绝走 neighbor refs 重写（F6 — AE 不动子层 ParentID，clone 是 sibling shadow）。**Refuse-cases 7 项** (spec §5): index 越界 / back==nil / Type≠AV / **TrackMatte≠None (F2 quirk Phase 3 拒接，~30 LOC 特殊定位逻辑 defer 到 3.1)** / name="" / corruption / shape-text deferred。Atomic invariants 复用 V2.1 pattern (spec §6) — key diff vs DeleteLayer：snapshot 必须含 `proj.nextItemID`（clone bumps；rollback 要 un-bump）；不需要 snapshot neighbor ldta bytes（无 cleanup pass）。Ship-gate target 6/6（3 modes × AE 2020+2025；`dup_matted` by design 不产 ge fixture，refuse error 是预期）。**buildLayerFromChunks 留给 Task 3 自决**（方案 a 重跑 parse vs b struct-clone+rebind backrefs，建议先试 b 更短）。
- **2026-05-28 V3 Phase 2 完 — Composition.DeleteLayer ship-gate green (Stable)** — 完整 RE→spec→impl→ship-gate 闭环。`delete_layer.go` (~200 行, V2.1 atomic invariant pattern, adaptive splice) + 9 test (8 PASS 1 SKIP); PASS 259→267. **Ship-gate 8/8 PASS**: AE 2020 (baseline/middle/parent/matte_ae20) + AE 2025 (baseline/middle/parent/matte_ae25); matte 拆双轨因为 AE 2020 拒接 AE 25 saved files 的版本 policy（与 DeleteLayer 无关）。matte_ae25 .done log 自验 F3：post-delete L2 trackMatteType=5013 保留 matteSrc 清，跟我们 impl 严格一致。AE 2020 matte_ae20 (implicit 无 @0xA0) cleanup branch 走 length guard skip 路径，文件完整保。Tooling: `tmp_debug/ge_delete_layer` 产 5 ge_*.aep + `verify_ge_delete_layer.jsx` 8 模式 verifier。Plan + strategy spec 移 finish/。godoc 从 alpha → Stable。Next: Phase 3 候选 DuplicateLayer（"16-chunk delete unit" 反向验我们对 follower chunks 的理解）。
- **2026-05-28 V3 Phase 2 Task 2 完 — DeleteLayer 策略矩阵落 spec** — `specs/2026-05-28-v3-phase2-deletelayer-strategy.md` (9 sections)。基于 scar 6 findings 全面决策。**关键修正 F4 → adaptive splice**：发现 NewShapeLayer 仅插 Layr+Ewst (2 chunks) 即 ship-gate 过，说明 14 follower 不是 layer-bound 必需；DeleteLayer 必须 adapt 不能 hardcode 16（否则会过删 Go-built layer 的下游 service block）。算法 locate Layr → assert Ewst → 消 leaf 直到下一个 LIST/EOF —— AE-saved (16) 和 Go-built (2) 两种状态都正确。其他决策：ParentID/TrackMatteLayerID 孤儿 reset to 0（match AE），TrackMatte byte @0x6B 不动（match AE），head counter 不动 (Inv-9)，opaque shard Phase 2 不启用（deferred Phase 3 InsertLayer/Duplicate）。refuse-cases 表 5 项（idx 越界 / back nil / Type≠AV / 单层删剩 0 / 后两者保守，AE 行为未 RE）。signature 0-based 索引匹配 Go slice。V2.1 atomic invariant 全套（snapshot itemList children + Layers + Warnings + 受影响 ldta bytes / rollback on warning）。Next: Task 3 实现 + Task 4 round-trip + Task 5 user-跑 8 次 AE 开盘 ship-gate（agent 现在可自跑 per playbook update）。
- **2026-05-28 V3 Phase 2 Task 1 完 — DeleteLayer RE + scar + playbook 刷新** — 走通了 RE harness → AE-saved fixture → chunk dump → scar 整圈。4 mode (`baseline/middle/parent/matte`) JSX 用 `$.getenv()` 切，agent 自己调 `ae_run.ps1`（baseline/middle/parent → AE 2020 17.7x45；matte → AE 2025 25.1x68 因 `TrackMatteLayerID` AE 23+）。`tmp_debug/dump_layers` 新写：parsed layers + raw Item LIST children + Layr/Ewst pairing。**6 Finding 进 scar `ae-deletelayer-re.md`**：(F1) splice 不 gap (237→221 children, Δ-16)；(F2) ParentID reset to 0；(F3) TrackMatteLayerID reset but trackMatteType byte 不动；(F4) **每 layer = 16-chunk atomic delete unit**（Layr + Ewst + 14 follower fvdv/fiop/ftts/foac/fiac/fipc/fifl ×2）—— 我们 parser 现在不 per-layer 跟踪 followers，但 V2.1 round-trip 靠 itemList children 原序保 byte-identical；(F5) DLay/SLay/CLay/SecL 同 16-chunk pattern 但要拒接 (FormType==IDLayr only)；(F6) 不复用 ID, head counter 不动。**playbook 刷新**：re-fixture.md 加"agent 自己调 ae_run.ps1, 不要 ask user" + "AE 2020 是默认 RE 版本, AE 23+ 字段才升 AE 25"；verify.md 加 dump_layers 到工具表。Next: Task 2 决策表（refuse-cases + 16-chunk splice atomic 推荐）。
- **2026-05-28 V3 Phase 1 完成 — backrefs shard 抽取** — 纯 refactor，按 `plans/2026-05-28-v3-phase1-backrefs-plan.md` 7 task 顺序执行（commits 1b7ff63→cfa9ff7）。6 个 scene 类型（Property/Keyframe/Layer/Composition/Footage/Project）顶层 `*rifx.Chunk` / `bytesPerKF` / `offset` / `dims` / `tickRate` / `compFps` / `rootFold` 等共 ~46 field 全部下沉到 per-type `back *<type>Backrefs` shard（`back_property.go` / `back_keyframe.go` / `back_layer.go` / `back_composition.go` / `back_footage.go` / `back_project.go`）。每个 backref struct 含 `opaque map[rifx.ChunkID]*rifx.Chunk` 占位字段（CLAUDE.md #5，Phase 2+ 启用）。验收：`go vet` clean、PASS=259 不变、FAIL=0、`TestRoundtripWrite*` 全过、public API 字段 / 签名零 delta（Stable API 不动；`CompItemListForTest` / `ItemListForTest` / `RootForTest` / `RootFoldForTest` 等 test-only helpers 加 nil-guard）。下一步 Phase 2 候选 DeleteLayer 走 scene + back 拆分的首个结构性 mutation。
- **2026-05-27 CLAUDE.md 硬约束 V3 时代刷新** — #1 length-preserving 加上"结构性 ops 走 V2.1 atomic invariants"补丁；#2 拆 Stable / Alpha 两级（撤销新加但已知 broken 的 API 不算违反 — 上次删 `ImportPlaceholder` 时的犹豫不再存在）；#3 单 package 理由更新为"文件名规约 + lint > 子包 re-export 噪音"；新加 #5 opaque preservation（V2.2 教训显化）+ #6 AE 接受 gate（结构性写必须双版本 ship-gate）。配套 `specs/2026-05-27-v3-deep-think.md` § 1.1 / § 2.5 / § 8 引用这些新硬约束。**这之前 V3 决策得绕这些约束的旧版本走，现在显化后 V3 Phase 1 plan 可以直接遵循**。

## Hanging tasks

无。
