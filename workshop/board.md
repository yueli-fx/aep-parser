# Board — aep-parser 项目看板

> 文档分工见 [../CLAUDE.md](../CLAUDE.md) 场景触发器表。本文件 = 现在在做啥 + 最近归档（≤ 2 周）+ PASS count 单一权威源。

**Last updated**: 2026-05-23 by claude (V2.2 Phase 1 runtime types 落地, ldta_layout + capability_matrix skeleton + PropertyStream[T] + shape graph + ShapeLayer wrapper; PASS = **141 / 0 FAIL**)
**Active focus**: 🟡 V2.2 ShapeLayer creation Phase 1 已 close (5 个新 .go + 3 个 test file, +19 PASS)。下一步: Phase 2 serializer primitives (`lower_property_stream.go` / `lower_shape_node.go` / `lower_layer.go` + rename `lower_item_siblings.go`)。

## Next session 进来先做

1. 确认 `go test ./internal/aep/... -count=1 -v | grep -c '^--- PASS'` = **141** + `go vet ./...` clean
2. 按 `workshop/plans/v2-2-layer-creation-plan.md` Phase 2 task 2.x 顺序开搞 serializer lowering
3. 走 `superpowers:executing-plans` 流程，每 phase 完一段 update board.md 归档
4. 改 public API 必同步 `docs/`、`coverage.md`、`coverage-detail.md`、本文件最近归档段

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
