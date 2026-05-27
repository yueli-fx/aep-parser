# Board — aep-parser

**Last updated**: 2026-05-28 by claude (V3 Phase 2 Task 1 完 — DeleteLayer RE 收口，scar 落地，playbook 刷新)
**Active focus**: V3 Phase 2 Task 2 —— 基于 scar `scars/ae-deletelayer-re.md` 6 个 Finding 写 DeleteLayer impl strategy 决策（refuse-cases + 是否在 V3 Phase 2 引入 follower-chunk per-layer 跟踪 vs 16-chunk splice 当作 atomic block）。

## Next session

**V3 Phase 2 Task 2 — 决策表**：基于 RE 结果（scar `ae-deletelayer-re.md`）决定:
1. **16-chunk splice 当 atomic block**（推荐）vs 先把 14 follower 拆到 `layerBackrefs.followers []` 再独立删 — 前者简单且足够 Phase 2，后者为 Phase 3 InsertLayer/DuplicateLayer 铺路但增加 Phase 2 风险面
2. **refuse-cases** 清单（plan Task 2.2 草稿）：
   - index 越界（idx < 0 || idx >= len(c.Layers)）
   - 单层 comp 删剩 0 个（AE 接不接？需 fixture 验，或直接保守拒）
   - 删 camera/light/audio (Type ≠ AV) — AE 行为未 RE，保守拒
   - back == nil（comp 不是从 parse 来的，没 itemList 可改）
3. **是否在 Phase 2 引入 backrefs.opaque shard** — Phase 1 留了占位字段，Q4 finding 显示 14 follower 是 opaque round-trip 的（parser 不读）。Phase 2 splice 不动它们就行，opaque 字段暂不启用。
4. **冒烟测**：是否 Phase 2 不实现 follower-chunk 拆分，留个 TODO 给 Phase 3 — 推荐 yes，scope 小
完后进 Task 3 实现，最后 Task 5 跑 8 次 AE 打开 ship-gate（4 mode × AE 2020+2025）。

**Task 1 产物**（已 ship）：
- `test_data/re_delete_layer.jsx` 4-mode RE harness
- `test_data/re_delete_layer_{baseline,middle,parent,matte}.aep` 4 fixtures（AE 2020 × 3 + AE 2025 × 1）
- `tmp_debug/dump_layers/main.go` chunk diff 工具
- `workshop/scars/ae-deletelayer-re.md` 6-finding RE doc
- `workshop/playbooks/re-fixture.md` 刷新 — agent 自己调 ae_run.ps1，AE 2020 是默认版本
- `workshop/playbooks/verify.md` 加 dump_layers 到 tmp_debug 工具表

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

- **2026-05-28 V3 Phase 2 Task 1 完 — DeleteLayer RE + scar + playbook 刷新** — 走通了 RE harness → AE-saved fixture → chunk dump → scar 整圈。4 mode (`baseline/middle/parent/matte`) JSX 用 `$.getenv()` 切，agent 自己调 `ae_run.ps1`（baseline/middle/parent → AE 2020 17.7x45；matte → AE 2025 25.1x68 因 `TrackMatteLayerID` AE 23+）。`tmp_debug/dump_layers` 新写：parsed layers + raw Item LIST children + Layr/Ewst pairing。**6 Finding 进 scar `ae-deletelayer-re.md`**：(F1) splice 不 gap (237→221 children, Δ-16)；(F2) ParentID reset to 0；(F3) TrackMatteLayerID reset but trackMatteType byte 不动；(F4) **每 layer = 16-chunk atomic delete unit**（Layr + Ewst + 14 follower fvdv/fiop/ftts/foac/fiac/fipc/fifl ×2）—— 我们 parser 现在不 per-layer 跟踪 followers，但 V2.1 round-trip 靠 itemList children 原序保 byte-identical；(F5) DLay/SLay/CLay/SecL 同 16-chunk pattern 但要拒接 (FormType==IDLayr only)；(F6) 不复用 ID, head counter 不动。**playbook 刷新**：re-fixture.md 加"agent 自己调 ae_run.ps1, 不要 ask user" + "AE 2020 是默认 RE 版本, AE 23+ 字段才升 AE 25"；verify.md 加 dump_layers 到工具表。Next: Task 2 决策表（refuse-cases + 16-chunk splice atomic 推荐）。
- **2026-05-28 V3 Phase 1 完成 — backrefs shard 抽取** — 纯 refactor，按 `plans/2026-05-28-v3-phase1-backrefs-plan.md` 7 task 顺序执行（commits 1b7ff63→cfa9ff7）。6 个 scene 类型（Property/Keyframe/Layer/Composition/Footage/Project）顶层 `*rifx.Chunk` / `bytesPerKF` / `offset` / `dims` / `tickRate` / `compFps` / `rootFold` 等共 ~46 field 全部下沉到 per-type `back *<type>Backrefs` shard（`back_property.go` / `back_keyframe.go` / `back_layer.go` / `back_composition.go` / `back_footage.go` / `back_project.go`）。每个 backref struct 含 `opaque map[rifx.ChunkID]*rifx.Chunk` 占位字段（CLAUDE.md #5，Phase 2+ 启用）。验收：`go vet` clean、PASS=259 不变、FAIL=0、`TestRoundtripWrite*` 全过、public API 字段 / 签名零 delta（Stable API 不动；`CompItemListForTest` / `ItemListForTest` / `RootForTest` / `RootFoldForTest` 等 test-only helpers 加 nil-guard）。下一步 Phase 2 候选 DeleteLayer 走 scene + back 拆分的首个结构性 mutation。
- **2026-05-27 CLAUDE.md 硬约束 V3 时代刷新** — #1 length-preserving 加上"结构性 ops 走 V2.1 atomic invariants"补丁；#2 拆 Stable / Alpha 两级（撤销新加但已知 broken 的 API 不算违反 — 上次删 `ImportPlaceholder` 时的犹豫不再存在）；#3 单 package 理由更新为"文件名规约 + lint > 子包 re-export 噪音"；新加 #5 opaque preservation（V2.2 教训显化）+ #6 AE 接受 gate（结构性写必须双版本 ship-gate）。配套 `specs/2026-05-27-v3-deep-think.md` § 1.1 / § 2.5 / § 8 引用这些新硬约束。**这之前 V3 决策得绕这些约束的旧版本走，现在显化后 V3 Phase 1 plan 可以直接遵循**。
- **2026-05-27 V3 deep-think spec** — `specs/2026-05-27-v3-deep-think.md`（266 行）。验证 M1-M8（M3 ShapeGraph 延后到 Phase 5；M4 effect schema 出 V3 scope）+ 7 个 open question 含推荐答案 + 7 行风险表 + Phase 1 纯 refactor 起步建议（back-ref 抽取，公开 API 零 delta，1-2 session 完成）+ 4 个决策点等 user 拍。
- **2026-05-27 P2c followup#2 — P2 parity 收尾** — 一波拿下 spec 里 P2 scope 剩下的 quick wins，没碰 V3/P3 结构性。新加 `Property.IsModified() / Enabled() / Active() / Elided() / IsNameSet()`（Enabled 读 tdsb byte3 bit0，IsModified 走 animated/expression/value≠default 三轨）+ `AEPropertyGroup.IsModified()`（indexed group 有 children = modified；否则递归）+ `Layer.AVSource() / CanSetCollapseTransformation() / CanSetTimeRemapEnabled()`（py-aep `AVLayer.can_set_*` parity）+ `Project.XmpPacket() R`（trailing UTF-8 after RIFX，已经 round-trip 走 `root.Trailing`，只缺 accessor）。spec §2.1 XmpPacket / §2.3 LayerType+CanSet+IsModified / §2.4 IsModified+ValueType / §2.5 MainSource / §2.9 Application.Version 等 7 行刷新成 ✅ done。新加 21 个 PASS test。**PASS 259**（基线 238 → 259，+21）。剩余 ❌ 全部 V3/P3：Render Queue / Essential Graphics / MotionGraphics / Layer structural ops / DimensionsSeparated / ValueText。

## Hanging tasks

无。
