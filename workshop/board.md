# Board — aep-parser 项目看板

> 文档分工见 [../CLAUDE.md](../CLAUDE.md) 场景触发器表。本文件 = 现在在做啥 + 最近归档（≤ 2 周）+ PASS count 单一权威源。

**Last updated**: 2026-05-23 by claude (V2.2 Phase 5 ship gate iter 3 — fix B 6-axis Transform schema (min-viable over-emit + Orientation otst + Position_0/_1 split + hydrate combine)；AE 2025 仍同错 (14s reject)；3 fixes necessary but not yet sufficient；PASS = **173 / 0 FAIL**)
**Active focus**: 🔴 V2.2 ship gate iter 4 — fix C (tdum/tduM + cdat per-dim padding) 或 minimum-failing bisection。fix A+iter2+iter3 三组 structural 修都对了但 AE 还拒。需缩范围。

## Next session 进来先做

1. 确认 PASS = **173** + vet clean
2. **ship gate iter 4 — 二选一**:
   - **路线 A: fix C (tdum/tduM + cdat per-dim padding)**：详见 `workshop/scars/v2-2-aelayer-structure.md` "Phase 5 fix order 修正后" 第 2 步。Tolerance 详:
     - `ADBE Position_0` static: `cdat (40B) + tdum (8B) + tduM (8B)`
     - `ADBE Vector Rect Size` static: `cdat (80B)` ← 我们 emit 48B (per-dim padding 错)
     - 改 `lower_property_stream.go::makeCdat` + `canonicalCdatSize`，spatial property 后 append tdum + tduM
   - **路线 B: minimum-failing bisection**：写 `tmp_debug/gen_minimum_failing/main.go` 出 1 个 empty ShapeLayer (无 shape 节点 / 无 keyframe / default transform)。跑 AE → 看是否仍同错。若 fail 同错 = 问题在 Layer 骨架；若 PASS = 问题在 shape/keyframe emit。
3. iter 4 PASS → Phase 5 剩 Task 5.5 (opaque preservation) + 5.6 (AE reopen) → Phase 6 docs sync
4. Claude 可自跑 AE: `AE_SHIP_GATE=1 go test -count=1 ./internal/aep/ -run TestV2_2_AEShipGate_AE2025 -v -timeout 180s` (无需关 2020；2020 跑前要关 2025)
5. 诊断工具齐: `tmp_debug/gen_canonical_failing/` 重建；`tmp_debug/dump_chunks/<path>` dump；`tmp_debug/dump_failing.txt` / `tmp_debug/dump_tolerance.txt` baseline；`tmp_debug/test_hydrate/` 直 hydrate 验

## V2.2 进度地图（本会话快照）

| Phase | 状态 | PASS | Commits | 关键产出 |
|---|---|---|---|---|
| 0 RE | ✅ 14/14 | 122 不变 | `d78c3af`..`8f35a63` | 9 RE finding (RE-S1..S9) 进 spec §8；schema corrections (matchName 真名 = `ADBE Root Vectors Group`；AE elides defaults；Path 用 om-s/shap/f32 bbox-norm) |
| 1 Runtime | ✅ 6/6 | 122→141 (+19) | `9ad612b`..`e1ab3a6` | ldta_layout / capability_matrix / PropertyStream[T] / shape_graph (5 typed nodes) / ShapeLayer wrapper |
| 2 Serializer | ✅ 5/5 | 141→154 (+13) | `3a2321c`..`8324ff9` | 4 个 lower_*.go primitive + V2.1 item-siblings rename |
| 3 Public API | ✅ 4/4 | 154→167 (+13) | `80a5ac5 / 7627204 / 785f6b3 / 3e56a49` | NewShapeLayer + Add{Rect,Ellipse,Path,Fill,Stroke} + PropertyGroup escape hatch β |
| 4 Roundtrip | ✅ 5/5 | 167→172 (+5) | (本会话 4 commits + 本 docs commit) | hydrateShapeNodes + write-time sync + canonical 3-layer roundtrip + atomicity ×3 + mutate-existing (skip) |
| 5 Ship gate | 🔴 4/8 + fix-iter-3 | 173 不变 | (Phase 5 commits + iter-1/2/3 commits) | 5.1/5.2/5.3/5.4 ✅; **5.7 iter 1**: outer wrapper ✅; **5.7 iter 2**: 5 placeholder ✅; **5.7 iter 3**: Transform 6-axis schema + Position_0/_1 split + Orientation otst ✅；AE 2025 同错 (14s reject) — 3 structural 修必须但不充分；下一步 fix C (tdum/tduM + cdat padding) 或 minimum-failing bisection |
| 6 Docs | ⏳ 0/5 | — | — | docs/shape.md + board archive + coverage sync + spec §6.4a/§6.5/§8 finalize |

## V2.2 永久知识（Phase 0 RE 已 freeze，不要再 RE）

- ShapeLayer 内容 root matchName = `ADBE Root Vectors Group`（plan 早期写错"Vector Materials Group"等都是虚构）
- AE 默认值会被 elide — `addProperty()` 无 setValue 时 sub-prop tdmn/cdat 完全不写盘。V2.2 builder 选"always emit"策略；ship gate 验证 AE 接受
- 嵌套 shape sub-groups (Dashes/Taper/Wave) 即便 default 也保留 3-child empty LIST tdgp — 不同 scalar/vector elision
- BezierPath 编码: `LIST(om-s)→omks→shap`，shph(24B) + lhd3(52B) + ldat(96B = 24×float32 BE = 4 verts × 6 f32 anchor.xy/in.xy/out.xy)。**bbox-normalized 0..1**，linear 和 tangent 同一格式
- keyframe lhd3 = 52 B；TickRate 不在 lhd3（来自 cdta @0x08）；numKeyframes @0x08；bytesPerKeyframe @0x10；2D 空间 stride = 128 B
- ldta 160B AE 2020 / 164B AE 2025 (后 4B 全 0 padding) → escape hatch AE 2020 minimum + length-preserving 跨版本 zero issue
- AECapabilities ship V2.2 时空 struct — 7 candidates 全 [no入 matrix]（escape hatch single canonical 全覆盖）
- ShapeLayer Layr 级 Transform Group 是 6-axis form (Position_0/_1, Orientation, RotateX/Y, Envir Appear) — 不是 2D user-facing 5-stream。V2.2 lowering 选 5-stream user-facing form（与 V1 parser convention 一致），与底层 6-axis 差异 Phase 4 roundtrip 校
- 12 fixtures + 5 RE scaffold tools 留在 `tmp_debug/{re_v22,dump_root,dump_kf,dump_cdat_seq,dump_ldat_f32,dump_path_bytes}`

## 永久知识（不要再 RE）

- `CompItem.dropFrame` / `TextDocument.fontLocation` / variable fonts axes **写** —— runtime-only 或 ScriptingAPI 不存在（详 `scars/runtime-only-fields.md` / `variable-fonts-write-noop.md`）
- 手动 kerning **首次启用** = 结构性（详 `scars/kerning-first-enable.md`）
- AE 脚本 `setAlternateSource(item)` 自动包 wrapper precomp，blsi 指 wrapper 不指 item（详 `scars/altsource-wrapper-precomp.md`）
- Shutter setter 触发 AE 端 work-area divisor 重编码到 24008 —— AE cosmetic optimization，我们 setter 单点写就行（详 `scars/shutter-side-effect-divisors.md`）
- Camera FilmSize 在 ldta `@0x98`（102.0472=36mm×72PPI），但 ScriptingAPI 不可写 → 不 ship setter（详 `scars/camera-filmsize-ldta-write-blocked.md`）
- cdta 总长 0xCC=204 字节，**无 tail**；layout AE 2020 ↔ 2025 完全一致

## 项目史（Wave 1-3 全收）

`Wave 1` 文本字段解封 / `Wave 2` AE 22+ 已有未 RE / `Wave 3` AE 24+ 新概念 + 文本扩展 —— 三个 Wave 在 2026-05-18 ~ 2026-05-22 期间全部 🟢 主体完工。暂搁项: `environmentLayer`（需 360° 素材）/ `ligature`（需 OT liga 字体）/ `maskFeatherFalloff`（位置未 RE）。

---

## 最近归档（≤ 2 周）

### 2026-05-23 V2.2 Phase 5 partial — JSX + Go shipgate + Tier 3 preservation (PASS 不变 172, +2 SKIP)

落 3 个 AE-independent task；剩 5 个 (5.3/5.5/5.6/5.7/5.8) 等用户 AE 机器或依赖 5.3 fixture。

- **Task 5.1 `verify_v2_2.jsx`** (commit `0fb6e9e`): JSX ship gate driver；per-check log；try/catch 包到 .done 必写防 Go 端 timeout-hang。Drives 3 canonical ShapeLayer (A animated / B static / C path)。
- **Task 5.2 `shape_layer_shipgate_test.go`** (commit `7854dd8`): `TestV2_2_AEShipGate_AE2020/2025`，仿 V2.1 runAEShipGate pattern。AE_SHIP_GATE env gate；CI / no-AE auto SKIP。
- **Task 5.4 `shape_preservation_test.go`** (commit `6dad000`): `TestV2_2_NestedGroup_Preservation`，skip-if-fixture-missing。Roundtrip 后 layer count / name / ShapePrimitive count unchanged。
- **`tmp_debug/gen_shape_tolerance.jsx`** (本 commit): Phase 5 Task 5.3 jsx — 用户跑 AE 2025 产 `test_data/v2_2_shape_tolerance.aep` fixture (nested VectorGroup + Rect + Fill)。
- **下一步 (用户)**: 跑 gen_shape_tolerance.jsx → commit fixture → 跑 ship gate (AE_SHIP_GATE=1)。
- **下一步 (Claude, 5.3 之后)**: Task 5.5 opaque chunk preservation + Project.RootChunk() + chunk-signature helper；Task 5.6 nested-group AE reopen test。

### 2026-05-23 V2.2 Phase 4 roundtrip + hydration (167 → 172 PASS, +5)

闭环 chunk ↔ runtime 双向。Phase 3 是 build-time only（NewShapeLayer 之后的 mutation 不到盘）；Phase 4 加 write-time sync + parse-time hydration 闭环。

- **Design 关键转向**: ShapeLayer 的 rootGroup/transform **从 wrapper 移到 Layer**（私字段 shapeRootGroup / shapeTransform）。ShapeLayer 变 thin façade — 所有 wrapper of 同一 Layer 共享 state。这是让 wrapper 上的 mutation 能 propagate 到 disk 的 enabler（write-time sync 只能从 Layer 找到 runtime tree）。
- **Task 4.1 hydrateShapeNodes** (commit `33e7db0`): chunk → VectorGroup tree (5 per-node hydrators)。复用 V1 `parseLeafProperty` per Inv-1。同 commit 落 `sync_shape_layers.go` — WriteAEP pre-pass 重 lower 任何 shapeRootGroup 非空的 layer，覆盖 `layrList.Children` in place。`base.layrList` back-ref 在 NewShapeLayer 也设上让 sync 找得到。
- **Task 4.2 canonical roundtrip** (commit `d050fa6`): 3-layer test (animated Rect+Fill+Position / static Ellipse+Stroke / static Path+Fill+Stroke)。一次过；逼出来的 hydration 扩展:
  - 3 个 generic stream hydrator (Float64/Vec2/Color4) 检 V1 prop.Keyframes，非空走 AddKeyframeLinear (flip→Animated)，空走 SetStaticValue。
  - hydratePathNode: om-s → omks → shap; 读 lhd3[0x0C] 拿 n; shph[3] 拿 closed flag; 用 shph bbox denorm f32 ldat 顶点。
  - hydrateLayerTransform: V1 layer.Properties (Anchor/Position/Scale/RotateZ/Opacity by matchName) → typed shapeTransform。parseLayer LayerTypeShape 分支 wire 上。
- **Task 4.3 atomicity** (commit `6ee8bc8`): 3 个 Inv-10 校验（empty name / negative time / duplicate time）— 行为 Phase 3 已 in place，pure 校验。+3 PASS。
- **Task 4.4 mutate-existing** (commit `4ef0e6a`): skip-if-missing test；fixture `v2_2_shape_tolerance.aep` 在 Phase 5。
- **Plan 错字纠**: plan line 2796 写的是 "ADBE Vector Materials Group"，真名 "ADBE Root Vectors Group"（board 永久知识有；hydrate 用真名）。
- **下一步**: Phase 5 — AE ship gate (verify_v2_2.jsx + AE 2020/25 实测 + Tier 3 preservation)。Phase 2 deferred 的 byte-layout 漏可能在 Phase 5 跑 AE 时暴露。

### 2026-05-23 V2.2 Phase 3 public API 3/4 (154 → 167 PASS, +13)

Phase 3 wired 用户可见入口。Tasks 3.1/3.2/3.3 落，3.4 (board archive) 即本段。

- **Task 3.1 `new_layer.go`** (commit `80a5ac5`) — `(c *Composition) NewShapeLayer(name string) (*ShapeLayer, error)`：原子 mutation（empty name → error before any state change；warnings-as-failure rollback per V2.1 pattern）。allocItemID via existing V2.1 helper。lowerCtx 接 c.TickRate / target / nextLayerID。
- **Task 3.2 VectorGroup.AddX** (commit `7627204`) — 5 个 Add{Rect,Ellipse,Path,Fill,Stroke}：append-to-Children = render-order top（first add = bottom = Children[0]）。
- **Task 3.3 escape hatch β** (commit `785f6b3`) — `PropertyGroup.{Float64,Vec2,Vec3,Color,Path}Stream(name)` typed accessors + Child(name) 嵌套。每个 ShapeNode 的 Properties() 从 Phase 1 占位 nil 改返实际 PropertyGroup（streams map 持有 typed *PropertyStream[T]）。escape hatch 用 runtime name (`"Size"`, `"Color"`) — 不暴露 AE matchName。
- **commits**: `80a5ac5` (3.1) / `7627204` (3.2) / `785f6b3` (3.3) / 本 docs commit
- **下一步**: Phase 4 — roundtrip + hydration (Task 4.1 hydrateShapeNodes + Task 4.2 canonical 3-layer roundtrip + 4.3 atomicity tests + 4.4 mutate-existing + 4.5 collect)

### 2026-05-23 V2.2 Phase 2 serializer primitives complete (141 → 154 PASS, +13)

Phase 2 落 4 个 `lower_*.go` 序列化 primitive。每个 file = 一个责任 (Inv-2)。
Byte layouts 来自 Phase 0 RE finding (spec §8 RE-S1..S9)；非 Phase 0 已 freeze
的字段标 TODO + Phase 4 roundtrip 校。

- **`lower_property_stream.go`** (Task 2.1, +4 PASS) — 5 个 typed lowering func:
  `LowerFloat64Stream` / `LowerVec2Stream` / `LowerVec3Stream` / `LowerColorStream`
  / `LowerPathStream`。generic core `lowerStream[T]` 覆盖 scalar / vector /
  color；Path 走单独 om-s / omks / shap / lhd3 / ldat f32 path per RE-S5b/S8。
  chunk builders: `padMatchName` (40B NUL-pad) + `makeTd{mn,sb,sn,b4}` +
  `makeCdat`。keyframe encoding mirror parse_keyframe.go layout (spatial-style
  bpk=0x38+3*dim*8; non-spatial bpk=0x08+5*dim*8); time = round(seconds × tickRate).
  `encodeBezier` bbox-normalize over verts ∪ verts+inTan ∪ verts+outTan →
  24×f32 BE per N-vertex path.
- **`lower_shape_node.go`** (Task 2.2, +6 PASS) — `shapeMatchNames` table +
  dispatcher `lowerShapeNode` + 5 per-kind funcs。Hot path emits typed:
  Rect (Size/Position/Roundness), Ellipse (Size/Position), Fill (Color/Opacity),
  Stroke (Color/Opacity/Width), Path (Path)。非-hot props (Direction / Blend
  Mode / Composite Order / Line Cap / Dashes / Taper / Wave 等) 用
  `emptySubPropPlaceholder` (per RE-S5d 3-child header-only group pattern)。
  `lowerVectorGroup` 按 RE-S3: tdsb + tdsn + N × (tdmn + LIST tdgp) + Group End.
- **`lower_layer.go`** (Task 2.3, +3 PASS) — `lowerShapeLayer` 出 LIST(Layr)。
  Children per RE-S1: ldta (160B AE 2020) + Utf8 + 可选 tdmn(Root Vectors Group)
  + LIST(tdgp) + tdmn(Transform Group) + LIST(tdgp, transform body)。
  `buildLdtaBytes` 填 160B canonical 经 ldta_layout.go constants (LayerID @0x00 /
  Quality=Best @0x04 / LayerSubtype=4 @0x80 / Visible bit @0x27 / etc.)。
  `lowerLayerTransform` 出 V2.2 user-facing 2D 5-stream form (Anchor/Position/
  Scale/Rotate Z/Opacity) — 不是 RE-S2 内部 6-axis schema；matches V1 parser
  convention。
- **`lower_item_siblings.go`** (Task 2.4, +0 PASS, pure refactor) — V2.1
  `Project.NewComposition` 内 inline siblingChunks loop 提到 `lowerItemSiblings(_
  *lowerCtx) []*rifx.Chunk`。行为不变（仍 deep-clone from AE 2020 dummy 模板），
  只是给 V3 brainstorm 一个 named primitive。
- **V2.2 strategy**: 始终 emit cdat (即使 value == default)。AE 自己 elide
  defaults，我们不复制 — AE 接受 non-elided form。Phase 4 roundtrip 是 byte-exact
  闸门。
- **Deferred to Phase 4 roundtrip**: tdb4 byte 0x08..0x0B observed 0x0 vs 0xffffffff
  (RE-S2 vs S5a)；cdat per-dim padding (40/48/56/96)；lhd3 bytes @0x14..0x1F
  常量。可能需调整。
- **commits**: `3a2321c` (2.1) / `0b74ef8` (2.2) / `4a8b151` (2.3) / `00f5b1d` (2.4) /
  本 docs commit
- **下一步**: Phase 3 public API entry — `comp.NewShapeLayer(name)` + Add*
  attach API + escape hatch β (generic property tree)

### 2026-05-23 V2.2 Phase 1 runtime types complete (122 → 141 PASS, +19)

Phase 1 落 spec §2 所有 runtime concept 的 Go 类型骨架。纯类型 + 内部状态机；无 serializer / 无 NewShapeLayer 入口。

- **5 个新文件 + 3 个 test**:
  - `ldta_layout.go` — ldta offset 常量 (160 AE 2020/22 / 164 AE 2025 zero-pad tail)
  - `capability_matrix.go` — `AECapabilities` 空 struct + `Capabilities(target)` 纯函数；7 个 RE-S9 候选全不入 matrix (admission rule 不满足)
  - `property_stream.go` — `PropertyStream[T]` 泛型 + `StreamMode` enum + `StreamKeyframe[T]` (rename 避免跟 V1 非泛型 Keyframe 冲突)；reuse 既有 V1 `TemporalEase`
  - `shape_graph.go` — `ShapeNodeKind` enum / `ShapeNode` interface / `VectorGroup` / `BezierPath` / 5 个 typed node (Rect/Ellipse/Path/Fill/Stroke) + `PropertyGroup` placeholder
  - `types_core.go` 追加 — `ShapeLayer` (embed `*Layer`) + `LayerTransform` typed wrapper + 4 shorthand (Position/Scale/Rotation/Opacity)
- **TDD 全走**: 每个文件先写测试 → 跑见 compile fail → 实现 → PASS。
- **Defaults 校准 per spec §3.6**: RectNode Size=[100,100] / EllipseNode Size=[100,100] / FillNode Color=[1,1,1,1] white / StrokeNode Color=[0,0,0,1] black + Width=2 / LayerTransform Scale=[100,100] + Opacity=100.
- **2 个 spec ambiguity 已记**: (1) spec §3.6 表格 FillNode Color 写"[1,0,0,1] red"但 elide 段说 Color 是 elided default，user prompt + plan test 均要求 white — 取 white；(2) plan code 用 `type Keyframe[T]` 名跟 V1 非泛型 `Keyframe` Go 编译冲突，rename `StreamKeyframe[T]`。
- **commits**: `9ad612b / 63a6f8c / bfc0293 / 40dd34f / e858d29` — 5 个 Phase 1 commits + 本 docs commit
- **下一步**: Phase 2 serializer primitives (lower_property_stream + lower_shape_node + lower_layer + 5 个 typed lowering function)

### 2026-05-23 V2.2 Phase 0 RE 完 (122 PASS 不变)

V2.2 实施前 RE：14 个 fixture (12 AE 2020 + 4 AE 2025) → 9 个 RE finding (RE-S1..S9) 填 spec §8。

- 2 个改变 spec 的大发现:
  1. matchName "Vector Materials/Vectors/Vector Transform Group" 全是虚构 → 真名 ADBE Root Vectors Group
  2. AE 默认值 elide — empty shape body = 3-child tdgp (tdsb+tdsn+Group End); cdat 只在 explicit setValue 时出现
- 其它发现: Path 用 shap/shph/lhd3/ldat 族 (非 cdat float64) + bbox-normalized float32; keyframe lhd3=52B (TickRate 不在 lhd3); Stroke nested-group 子树 (Dashes/Taper/Wave) 即便 default 也保留 3-child tdgp; ldta size 160 (2020/22) vs 164 (2025) — tail 4B 零填充, 不入 capability matrix
- AECapabilities ship V2.2 时空 struct (7 candidates 全 [no 入 matrix])
- 工具: gen_shape_dummy.jsx 12 scenarios; dump_root/dump_kf/dump_cdat_seq/dump_ldat_f32/dump_path_bytes 5 RE scaffolds
- spec §3.6 defaults table + §4.0/4.2/4.3 schema 校准 + §6.4 substrate source attribution 全 freeze (本 task 0.14)
- next: Phase 1 runtime types (PASS ≥ 130 target)

### 2026-05-22 V2.2 brainstorm 完，spec ship (PASS 不变)

- **Spec**: `workshop/specs/v2-2-layer-creation-design.md` (1213 行；commit `5b5ba6f`)
- **范围**: ShapeLayer end-to-end — 空 ShapeLayer + Rect/Ellipse/Path/Fill/Stroke 节点 + PropertyStream 静态+keyframe + AE 2020/25 ship gate
- **架构**: 10 条 Architecture Invariant + 4 条 Serializer Invariant + Capability admission rule（3 条件）+ 5 个 reusable serializer primitive 给 V3 inherit
- **Classification deliverable**: 17 runtime concepts / 19 serialization artifacts / 5 substrates / 7 capability candidates 显式归类 —— V3 brainstorm 直接 input
- **Dual-track 战略**（per GPT feedback）: V2 RE/prototyping continues + V3 abstractions 增量提取 from V2 work
- **3 路 LLM 反馈**绕了 6 轮（每 section 一轮），全部 absorb 进 spec
- **下一步**: writing-plans 产 `workshop/plans/v2-2-layer-creation-plan.md`，按 7 phase 拆 task 执行

### 2026-05-22 V2.1 Foundation ship — NewProject + NewComposition (109 → 122 PASS, +13)

V2 第一个 sub-project：从零创建 .aep 通过 AE 2020 / AE 2025 ship gate。

- **Public API**:
  - `aep.NewProject(target ...AETarget) *Project` (零参 = TargetAE2020)
  - `proj.NewComposition(name, w, h, fps, duration) (*Composition, error)`
  - `AETarget` enum: `TargetAE2020 / TargetAE2022 / TargetAE2025`
- **AE 接受 gate 找到 5 道梯度症状** —— 详 `scars/ae25-acceptance-gate.md`:
  1. cdta 二级 timing 字段空 → AE 崩
  2. head counter < itemID → "数据丢失"
  3. Item Fold-level siblings 缺 → "数据丢失"
  4. cdta masterTicks 错值 → duration 显示错
  5. per-version Item internals 差异 → AE 2020 拒开 AE 25 写的 items
- **Canonical seed 收敛**: AE 高版本 back-compat 读低版本 → builder 永远输出 AE 2020 最小子集，一个 `templates/2020_dummy_comp.aep` 即可跨 AE 2020/22/25。`Project.target` 字段管 empty-project skeleton (svap/nhed)，不管 Item internals。
- **新文件**:
  - `internal/aep/new_project.go` / `new_composition.go` / `cdta_layout.go` / `framerate_canonical.go`
  - `internal/aep/templates/2020_dummy_comp.aep` (唯一 canonical seed)
  - `test_data/verify_v2_1.jsx` + `test_data/v2_smoke{,_ae2020}.aep`
  - `tmp_debug/gen_dummy_comp.jsx` + `tmp_debug/v2_smoke/main.go`
- **修改**: `Project.target / nextItemID / rootFold` (types_core) / `parseProject.initDerived` (parse) / `write.go` `syncHeadCounters`
- **教训**: builder ≠ parser 反向（parser 容错 / AE 严格）。多信号区分根因（崩溃 / 数据丢失 / 显示错值 / ScriptingAPI 偏移）。ScriptingAPI 至少 NTSC `shutterAngle` 返回 stored × 1.2 quirk —— 储存字节跟 AE-saved 一致即可。
- **commits**: `e4e5b00 / 9dab01f / 9d6a938 / 728e5af / 270bdae` —— 共 5 个 Phase 6 commits + Phase 7 docs sync。

### 2026-05-22 文档大重组 (109 PASS 不变)

- **目的**: workshop/ 内多文档重复 + 过时（PASS count 在 4 处显示不同值；INDEX/CLAUDE/coverage 各自有"文档地图"；architecture.md 文件地图早就过时；refactor.md Phase 1-3 全完成但还挂着）。
- **删除**: `workshop/INDEX.md`（与 CLAUDE.md 场景触发器表重复）、`workshop/plans/refactor.md`（Phase 1-3 全完成，历史归档到本文件）。
- **重写**: `CLAUDE.md`（场景触发器表设为**唯一权威文档地图**）/ `workshop/specs/architecture.md`（文件地图刷到当前 25 production + 18 test 文件现状）/ 本 `board.md`（修内部 PASS count 矛盾）。
- **更新**: `coverage.md` 删 "下一步候选" 全做完段 + "文档分工提醒" 重复段；`coverage-detail.md` 删 "优先级聚合 P0-P3" + "Wave Roadmap" 历史段（已归档）；`playbooks/verify.md` PASS=109。
- **PASS count 单一权威源规则**：以后 PASS 数字只在本文件 `Last updated` 一处出现，其他文档说 "见 board.md" 或不提具体数字（防漂移）。

### 2026-05-22 自由探索批次 (100 → 109 PASS, +9)

整体收尾：cdta tail / effect param setter / 跨版本 / Material / Geometry / Renderer / typed accessor cleanup。

- **`Composition.ResolutionFactor [2]uint16` R/W** (新发现) — cdta `@0x00`/`@0x02`（之前完全没 RE 过的头 4 字节）。`[1,1]`=Full / `[2,2]`=Half / `[3,4]`=非方形合法。新 fixture `re_cdta_probe.{jsx,aep}`。
- **`Composition.Renderer string` R only** (新发现) — `PRin` LIST → `prin` chunk @offset 4，NUL-sep 双段 ASCII（match-name + locale 名）。surface match-name (`ADBE Escher`=Advanced 3D / `ADBE Ernst`=Cinema 4D / `ADBE Standard`=Classic 3D)。新 rifx chunk IDs `IDPRin / IDPrin / IDPrda`。setter 是 P3（prda 长度随 renderer 变）。新 fixture `re_renderer.{jsx,aep}`。
- **Material Options 3D AV layer**: 17 getter + 17 setter + `MaterialCastsShadowsMode` 三态 enum（Off/On/Only）。`ADBE Casts Shadows` 同 match name 在 light 是双态、AV 3D 是三态 —— light 用 `SetLightCastsShadows(bool)`、AV 3D 用 `SetMaterialCastsShadows(mode)`。**AE 序列化陷阱**：设回默认值时 AE 会裁掉 property 不写盘；getter 可能 nil。新 fixture `re_material_options.{jsx,aep}`。
- **Geometry Options**: 3 getter + 3 setter（PlaneCurvature / PlaneSubdivision / BevelDirection）。新 fixture `re_geometry_options.{jsx,aep}`。
- **Camera/Light typed setter 全集**: 24 setter（13 Camera + 11 Light）+ 共享 `setScalarProperty` helper（nil → "property not present" 错误）。`SetCameraDepthOfField(bool) / SetLightCastsShadows(bool)` 用 bool 自动 → 1.0/0.0。`SetLightColor([]float64)` dim 校验由 SetStaticValue 处理。
- **Transform + AudioLevels typed setter**: 9 setter（AnchorPoint / Position / Scale / Rotation / RotateX / RotateY / Orientation / Opacity / AudioLevels）+ 3 新 getter (RotateX/Y/Orientation, 3D-only)。
- **Effect param 用 Property.SetStaticValue 即可** (验证) — `Layer.Effects[i].Parameters[j]` 就是普通 `*Property`，复用既有 setter。无新代码。
- **跨 AE 版本 cdta diff** (negative) — AE 2020 跑 `re_cdta_ae2020.jsx` 跟 AE 2025 跑 `re_cdta_probe.jsx`，cdta bit-for-bit 完全一致（5 年 stable）。不需要版本-条件分支。
- **副发现**: ldta `@0x80..0x83` = LayerSubtype enum (0=AV / 1=Light / 2=Camera / 3=Text / 4=Shape / 5=3DModel)，parser 已用。ldta `@0x98..0x9F` 疑似 camera FilmSize 但写路径堵死 (scar)。

### 2026-05-22 重构 Phase 1+2+3 全收 (100 PASS 不变)

文件级拆分，纯移位 0 改逻辑 0 改 public API。脚本 `scripts/split_{tests,parse_text,write}.py`。

- **Phase 1**: `aep_test.go` 6667 → 307；拆 12 个 test 文件 + 6 个 testutil。`split_tests.py` 一次性扫完，**比手动拆省 ~10× token**。教训：纯机械批量移动直接写脚本。
- **Phase 2**: `parse_text.go` 1016 → 354 + 新 `text_types.go` (326) + `postscript.go` (347)。`split_parse_text.py` 按行号 slice。
- **Phase 3**: `write.go` 805 → 171 + 新 `write_keyframe.go` (252) + `write_property.go` (306) + `Layer.SetText` 并入 `write_layer.go` (729→830)。`split_write.py` 含 append 模式。**教训**: import 检测 regex 用 `\bpkg\.\w` 强制要求 `.` 后跟标识符首字符，否则注释里 "ldat bytes. Used" 句号误中。
- **Phase 4 不做**: `write_layer.go (830)` 单一类型偏大但可接受；`types_core.go (705)` 几乎纯类型；`json.go (536)` cohesive。不在 budget。

### 2026-05-21 Phase 1.1 testutil 拆 5 文件 (100 PASS 不变)

aep_test.go 6667 → 5780。`testutil_aep24/mask/text/keyframe/layer_test.go` 出。后续 1.2 用脚本批量。

### 在此之前（≤ 2 周前里程碑摘要）

- Wave 3 主体完工：文本扩展 7 字段 / alternateSource (Media Replacement) / manual kerning R/W / variable fonts axes R / `fontLocation` negative finding
- Wave 2 主体完工：trackMatteLayer / TimeRemap / LightKind / DisplayStartTime；negative: `dropFrame` runtime-only / `timeRemapEnabled toggle` structural / ldta 第三长度未发现
- Wave 1 主体完工：CapsOption / BaselineOption / StrokeOverFill / 段缩进×5 / AutoHyphenate
- 文档骨架重构：CLAUDE.md / COORDINATOR.md / REFERENCE.md 三件套（后整合到 `workshop/`）
- 之前：文本 16 setter + AddFont / Composition 标志位 + pixelAspect / Keyframe 增删 / Item-level comment + label / Project bitsPerChannel
