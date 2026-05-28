# Board — aep-parser

**Last updated**: 2026-05-28 by claude (V3 Phase 3 Task 1 完 — DuplicateLayer RE done, 10 Findings 落 scar)
**Active focus**: V3 Phase 3 Task 2 —— 按 scar `ae-duplicatelayer-re.md` 10 Findings 写 strategy spec，然后 Task 3 实现 + Task 4 测试 + Task 5 ship-gate。Plan: `plans/2026-05-28-v3-phase3-duplicatelayer-plan.md`。

## Next session

**V3 Phase 3 Task 2 — strategy spec**：把 scar 10 个 Findings 翻译成决策矩阵，落 `specs/2026-05-28-v3-phase3-duplicatelayer-strategy.md`。关键决策已经被 RE 锁定，写起来快：

1. **Signature**: `func (c *Composition) DuplicateLayer(index int, name string) (*Layer, error)` —— 取 explicit name 参数（Finding 9 AE 不 auto-suffix；让 caller 控制更干净）
2. **Insertion**: new layer at `oldIndexOf(source)` (Finding 1)。Phase 3 MVP **refuse 当 source 有 trackMatteType 设置**（Finding 2 quirk —— defer 复杂特殊处理）
3. **ID**: `proj.allocItemID()` 同 NewShapeLayer (Finding 3)
4. **16-chunk verbatim clone** + mutate ldta @0x00..0x03 为新 ID (Finding 10)。Data slice 必须 fresh copy (concurrent-mutate scar)
5. **Refuse-cases**: Type≠AV / index 越界 / back==nil / source.TrackMatte≠None (Phase 3 conservative) / source 是 shape/text (Q8 deferred)
6. **不动**: children's ParentID (Finding 6 — 仍指 source)
7. **V2.1 atomic**: snapshot itemList children + Layers + Warnings + proj.nextItemID + 任何 ldta 写过的 neighbor (实际只动 clone 自己的 ldta @0x00；不会动 neighbor，但保 rollback 完整)

**Task 1 产物（已 ship）**:
- `test_data/re_duplicate_layer.jsx` (4 modes)
- 4 AE 2020 fixtures (gitignored)
- `tmp_debug/diff_dup_blocks/main.go` (clone vs source byte-diff helper) + `tmp_debug/probe_ldta/main.go` (ldta first-32B + key offset dump)
- `workshop/scars/ae-duplicatelayer-re.md` (10 Findings incl byte-level verification — clone 14 followers + ldta-body 全 byte-identical except @0x00..0x03)

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

- **2026-05-28 V3 Phase 2 完 — Composition.DeleteLayer ship-gate green (Stable)** — 完整 RE→spec→impl→ship-gate 闭环。`delete_layer.go` (~200 行, V2.1 atomic invariant pattern, adaptive splice) + 9 test (8 PASS 1 SKIP); PASS 259→267. **Ship-gate 8/8 PASS**: AE 2020 (baseline/middle/parent/matte_ae20) + AE 2025 (baseline/middle/parent/matte_ae25); matte 拆双轨因为 AE 2020 拒接 AE 25 saved files 的版本 policy（与 DeleteLayer 无关）。matte_ae25 .done log 自验 F3：post-delete L2 trackMatteType=5013 保留 matteSrc 清，跟我们 impl 严格一致。AE 2020 matte_ae20 (implicit 无 @0xA0) cleanup branch 走 length guard skip 路径，文件完整保。Tooling: `tmp_debug/ge_delete_layer` 产 5 ge_*.aep + `verify_ge_delete_layer.jsx` 8 模式 verifier。Plan + strategy spec 移 finish/。godoc 从 alpha → Stable。Next: Phase 3 候选 DuplicateLayer（"16-chunk delete unit" 反向验我们对 follower chunks 的理解）。
- **2026-05-28 V3 Phase 2 Task 2 完 — DeleteLayer 策略矩阵落 spec** — `specs/2026-05-28-v3-phase2-deletelayer-strategy.md` (9 sections)。基于 scar 6 findings 全面决策。**关键修正 F4 → adaptive splice**：发现 NewShapeLayer 仅插 Layr+Ewst (2 chunks) 即 ship-gate 过，说明 14 follower 不是 layer-bound 必需；DeleteLayer 必须 adapt 不能 hardcode 16（否则会过删 Go-built layer 的下游 service block）。算法 locate Layr → assert Ewst → 消 leaf 直到下一个 LIST/EOF —— AE-saved (16) 和 Go-built (2) 两种状态都正确。其他决策：ParentID/TrackMatteLayerID 孤儿 reset to 0（match AE），TrackMatte byte @0x6B 不动（match AE），head counter 不动 (Inv-9)，opaque shard Phase 2 不启用（deferred Phase 3 InsertLayer/Duplicate）。refuse-cases 表 5 项（idx 越界 / back nil / Type≠AV / 单层删剩 0 / 后两者保守，AE 行为未 RE）。signature 0-based 索引匹配 Go slice。V2.1 atomic invariant 全套（snapshot itemList children + Layers + Warnings + 受影响 ldta bytes / rollback on warning）。Next: Task 3 实现 + Task 4 round-trip + Task 5 user-跑 8 次 AE 开盘 ship-gate（agent 现在可自跑 per playbook update）。
- **2026-05-28 V3 Phase 2 Task 1 完 — DeleteLayer RE + scar + playbook 刷新** — 走通了 RE harness → AE-saved fixture → chunk dump → scar 整圈。4 mode (`baseline/middle/parent/matte`) JSX 用 `$.getenv()` 切，agent 自己调 `ae_run.ps1`（baseline/middle/parent → AE 2020 17.7x45；matte → AE 2025 25.1x68 因 `TrackMatteLayerID` AE 23+）。`tmp_debug/dump_layers` 新写：parsed layers + raw Item LIST children + Layr/Ewst pairing。**6 Finding 进 scar `ae-deletelayer-re.md`**：(F1) splice 不 gap (237→221 children, Δ-16)；(F2) ParentID reset to 0；(F3) TrackMatteLayerID reset but trackMatteType byte 不动；(F4) **每 layer = 16-chunk atomic delete unit**（Layr + Ewst + 14 follower fvdv/fiop/ftts/foac/fiac/fipc/fifl ×2）—— 我们 parser 现在不 per-layer 跟踪 followers，但 V2.1 round-trip 靠 itemList children 原序保 byte-identical；(F5) DLay/SLay/CLay/SecL 同 16-chunk pattern 但要拒接 (FormType==IDLayr only)；(F6) 不复用 ID, head counter 不动。**playbook 刷新**：re-fixture.md 加"agent 自己调 ae_run.ps1, 不要 ask user" + "AE 2020 是默认 RE 版本, AE 23+ 字段才升 AE 25"；verify.md 加 dump_layers 到工具表。Next: Task 2 决策表（refuse-cases + 16-chunk splice atomic 推荐）。
- **2026-05-28 V3 Phase 1 完成 — backrefs shard 抽取** — 纯 refactor，按 `plans/2026-05-28-v3-phase1-backrefs-plan.md` 7 task 顺序执行（commits 1b7ff63→cfa9ff7）。6 个 scene 类型（Property/Keyframe/Layer/Composition/Footage/Project）顶层 `*rifx.Chunk` / `bytesPerKF` / `offset` / `dims` / `tickRate` / `compFps` / `rootFold` 等共 ~46 field 全部下沉到 per-type `back *<type>Backrefs` shard（`back_property.go` / `back_keyframe.go` / `back_layer.go` / `back_composition.go` / `back_footage.go` / `back_project.go`）。每个 backref struct 含 `opaque map[rifx.ChunkID]*rifx.Chunk` 占位字段（CLAUDE.md #5，Phase 2+ 启用）。验收：`go vet` clean、PASS=259 不变、FAIL=0、`TestRoundtripWrite*` 全过、public API 字段 / 签名零 delta（Stable API 不动；`CompItemListForTest` / `ItemListForTest` / `RootForTest` / `RootFoldForTest` 等 test-only helpers 加 nil-guard）。下一步 Phase 2 候选 DeleteLayer 走 scene + back 拆分的首个结构性 mutation。
- **2026-05-27 CLAUDE.md 硬约束 V3 时代刷新** — #1 length-preserving 加上"结构性 ops 走 V2.1 atomic invariants"补丁；#2 拆 Stable / Alpha 两级（撤销新加但已知 broken 的 API 不算违反 — 上次删 `ImportPlaceholder` 时的犹豫不再存在）；#3 单 package 理由更新为"文件名规约 + lint > 子包 re-export 噪音"；新加 #5 opaque preservation（V2.2 教训显化）+ #6 AE 接受 gate（结构性写必须双版本 ship-gate）。配套 `specs/2026-05-27-v3-deep-think.md` § 1.1 / § 2.5 / § 8 引用这些新硬约束。**这之前 V3 决策得绕这些约束的旧版本走，现在显化后 V3 Phase 1 plan 可以直接遵循**。

## Hanging tasks

无。
