# Board — aep-parser

**Last updated**: 2026-05-28 by claude (V3 Phase 2 完 — DeleteLayer ship-gate 8/8 PASS, alpha → Stable)
**Active focus**: V3 Phase 3 候选评估（InsertLayer / DuplicateLayer / Layer.Remove on Project / capability matrix start）—— Phase 2 跑通了 "RE → strategy spec → impl → ship-gate" 完整流程，可复用。

## Next session

**V3 Phase 3 起步**：从下列候选选 1-2 个起 plan：

1. **`Composition.DuplicateLayer(idx) (*Layer, error)`** —— 反过来加，比 InsertLayer 简单（克隆已有 layer 而非合成）。但要 RE：AE duplicate 后 Layr 的 ID 怎么分（一定是新 ID）？fvdv/fiop/ftts/etc 14 follower chunks 也得复制？bytecode 完全 byte-clone 还是要 mutate 某些 ID/timestamp 字段？
2. **`Composition.InsertLayer(srcLayer, atIdx)`** —— 跟 DuplicateLayer 接近但 source layer 来自外部（跨 comp 移图层场景）。
3. **`Project.RemoveItem(itemID)`** —— Composition / Footage / Folder 级别 delete。Item delete 的引用清理面更大（Composition 删了之后所有引用它的 Layer.SourceID 要清；Footage 删了之后所有 Layer.SourceID 要清）。
4. **Capability matrix start** —— spec V3 deep-think § M2 提到「不同 chunk 类型 capability 不同」的元数据表。Phase 3 不一定要做但 Phase 4+ 必需。

**推荐**: DuplicateLayer 起手 —— smallest scope，最大复用 Phase 2 经验，把 "16-chunk delete unit" 反向变 "16-chunk clone unit" 验证我们对 follower chunks 的理解。

**Phase 2 完整产物**（git ship）：
- 代码: `internal/aep/delete_layer.go` + `_test.go` (9 test, 8 PASS 1 SKIP)
- 工具: `tmp_debug/dump_layers/main.go` (Item LIST diff) + `tmp_debug/ge_delete_layer/main.go` (ship-gate fixture producer)
- JSX: `test_data/re_delete_layer.jsx` (5 modes incl matte_predelete) + `test_data/verify_ge_delete_layer.jsx` (ship-gate verifier)
- 文档: `scars/ae-deletelayer-re.md` (6 findings) + `specs/finish/2026-05-28-v3-phase2-deletelayer-strategy.md` (9 sections) + `plans/finish/2026-05-28-v3-phase2-deletelayer-plan.md`
- Playbook 刷新: re-fixture.md (agent 自跑 AE / AE 2020 默认) + verify.md (+ dump_layers)

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
