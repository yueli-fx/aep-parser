# Board — aep-parser 项目看板

> 文档分工见 [../CLAUDE.md](../CLAUDE.md) 场景触发器表。本文件 = 现在在做啥 + 最近归档（≤ 2 周）+ PASS count 单一权威源。

**Last updated**: 2026-05-26 by claude (V2.2 Phase 6 docs sync **闭环** ✓ — workshop v0.8.0 hard-fail compliance migration (10 scars + 3 playbooks 全补 `when_to_read`/`applies_to`/`last_updated` frontmatter) + Phase 6 docs sync 3 文件全 update: `docs/shape.md` 加 "V2.2 alpha — Builder API" section (+83 行, 含 API surface + 限制声明 + RE 路线说明), `workshop/plans/coverage.md` 加 V2.2 alpha ShapeLayer 写路径段 (+14 行), `workshop/specs/v2-2-layer-creation-design.md` finalize: §6.4a 标 closed + §6.5 confirm AECapabilities 空 struct + 新 §10 实际 shipped iter-7/8 embed approach (+116 行)。V2.2 Phase 5 + Phase 6 完整闭环, alpha 可 ship。V2.2.1 候选 (Ellipse/Path/Stroke + Color encoding + keyframe persist) 留下次。PASS = **175 / 0 FAIL** 不变)
**Active focus**: 🟢 V2.2 alpha 完整闭环 (Phase 5 ship gate + Phase 6 docs sync)。V2.2.1 / V3 work 留下次会话。

## iter-8 (本会话) 总结

**iter-7 解 Layr Transform 后**, iter-8 用同样 transplant 法 isolate shape content silent drop:
- `swap_rect_body` (ours wrappers + tolerance Rect body) → layers=1 ✓ — Rect body 内部是触发器
- `swap_fill_body` (only Fill body swapped) → layers=0 — Rect body 还在 drop
- `swap_both_bodies` (both Rect + Fill swapped) → layers=1 ✓ — 各 shape body 独立校验

**Solution (iter-7 pattern, finer granularity)**:
- `tmp_debug/extract_shape_bodies/main.go` 抽 tolerance Rect + Fill body → `internal/aep/templates/v2_2_shape_rect_body.bin` (448B) + `v2_2_shape_fill_body.bin` (426B)
- `lower_shape_node.go::lowerRectNode` / `lowerFillNode` 重写: embed + clone + `overwriteShapeStreamCdat` 覆写 Size / Color cdat 用 runtime 值
- Ellipse / Path / Stroke / Position / Roundness / Opacity / Blend Mode etc 全标 runtime-only (V2.2.1 工作)

**AE bisect**:
| variant | iter-7 | iter-8 |
|---|---|---|
| #2 empty | layers=1 | layers=1 |
| #3 +Rect | drop | **layers=1** ✓ |
| #4 +SetSize | drop | **layers=1** ✓ |
| #5 +Fill | drop | **layers=1** ✓ |
| #6 size kf | drop | **layers=1** ✓ |
| #7 position kf | drop | **layers=1** ✓ |

## Next session 进来先做

1. 确认 PASS = **175** + vet clean
2. **V2.2.1 候选** (V2.2 alpha 闭环, 下一步选一条做):
   - **Ellipse/Path/Stroke embed bytes**: 用户用 AE create 各 shape fixture (`re_v2_2_ellipse.aep` / `re_v2_2_path.aep` / `re_v2_2_stroke.aep`), 抽 body, extend `tmp_debug/extract_shape_bodies/`, refactor `lowerEllipseNode` / `lowerPathNode` / `lowerStrokeNode` 用 embed approach
   - **Fill Color 编码 RE**: tolerance.aep cdat 跟 JSX 0..1 input 不对齐 (0.5 → 0x406fe0... ≈ 255); 写 RE fixture 不同 Color setValue → diff cdat 字节, 找出编码方式
   - **keyframe 持久化**: embed body 现只 static cdat slot; 加 LIST(list) lhd3/ldat keyframe encoding 注入路径
3. V2.3 方向 (大改): 看 `workshop/specs/v3-direction.md`; embed boilerplate approach pattern 可能跟 V3 capability framework 怎么整合
4. iter-7/8 工具齐 (V2.2.1 复用):
   - `extract_transform_group/` / `extract_shape_bodies/` — embed 抽取流水
   - `swap_propgroup/` / `swap_rect_body/` / `swap_fill_body/` / `swap_both_bodies/` / `swap_rvg/` / `swap_reverse/` — transplant isolate
   - `verify_baseline/` / `bisect_v2_2/` (batch)
   - `gen_v2_only/` / `gen_v3_only/` / `gen_v5_only/` / `gen_iter5_check/`
   - `dump_root/` / `dump_project_chunks/` / `dump_cdta_full/` / `dump_gide/`
   - `find_layerid_refs/` / `diff_*` 系列

---

## 旧 next session (Phase 6 docs sync — 已 done)

1. 确认 PASS = **175** + vet clean ✓ (本会话开始)
2. **Phase 6 docs sync** (本会话已 done — 3 文件):
   - variant #2 (empty ShapeLayer) PASS layers=1 ✓
   - variant #3 (+AddRect) 仍 drop → 触发器在 OUR Root Vectors Group / Rect body emit
   - 跟 iter-7 同样的 transplant 法 isolate: 拿 iter-7 状态的 minfail_v3.aep + swap in tolerance's Root Vectors Group → AE 跑看 layers.length
   - 如果 swap 后 layers=1 → 跟 iter-7 同思路: embed tolerance Rect body 作 boilerplate
   - 如果 swap 后 layers=0 → 触发器不在 Root Vectors Group, 而在别处 (5-层 nesting wrapper / Item-level changes 加 Rect 后 chunks?)
3. iter-8 之前**不要再盲改** — 用 `tmp_debug/swap_propgroup/` 或类似 transplant 工具 isolate 真凶。
4. iter-7 工具齐:
   - `tmp_debug/verify_baseline/` (单变体 AE + dump 完整 metrics — items/comp/layer class)
   - `tmp_debug/transplant_layr/` / `transplant_tdgp/` (Layr / outer tdgp 整块 swap)
   - `tmp_debug/swap_propgroup/` (per-prop-group swap, batch AE 跑 6 个)
   - `tmp_debug/swap_reverse/` (反向 swap 验证)
   - `tmp_debug/extract_transform_group/` (出 templates/v2_2_transform_group_body.bin)
   - `tmp_debug/gen_v2_only/` (build empty ShapeLayer .aep)
   - `tmp_debug/dump_project_chunks/` (root-level chunks)
   - `tmp_debug/find_layerid_refs/` (search by byte sequence)
   - `tmp_debug/diff_stream/` / `diff_named_chunk/` / `diff_comp_chunks/` / `diff_ldta/` (multi-level diffs)
   - `tmp_debug/dump_cdta_full/` / `dump_gide/`
5. Phase 6 docs sync 等 iter-8 闭环后做。

## V2.2 进度地图（本会话快照）

| Phase | 状态 | PASS | Commits | 关键产出 |
|---|---|---|---|---|
| 0 RE | ✅ 14/14 | 122 不变 | `d78c3af`..`8f35a63` | 9 RE finding (RE-S1..S9) 进 spec §8；schema corrections (matchName 真名 = `ADBE Root Vectors Group`；AE elides defaults；Path 用 om-s/shap/f32 bbox-norm) |
| 1 Runtime | ✅ 6/6 | 122→141 (+19) | `9ad612b`..`e1ab3a6` | ldta_layout / capability_matrix / PropertyStream[T] / shape_graph (5 typed nodes) / ShapeLayer wrapper |
| 2 Serializer | ✅ 5/5 | 141→154 (+13) | `3a2321c`..`8324ff9` | 4 个 lower_*.go primitive + V2.1 item-siblings rename |
| 3 Public API | ✅ 4/4 | 154→167 (+13) | `80a5ac5 / 7627204 / 785f6b3 / 3e56a49` | NewShapeLayer + Add{Rect,Ellipse,Path,Fill,Stroke} + PropertyGroup escape hatch β |
| 4 Roundtrip | ✅ 5/5 | 167→172 (+5) | (本会话 4 commits + 本 docs commit) | hydrateShapeNodes + write-time sync + canonical 3-layer roundtrip + atomicity ×3 + mutate-existing (skip) |
| 5 Ship gate | ✅ 4/8 + iter-1..8 | 173 → 175 | (Phase 5 + iter-1..8 commits) | 5.1/5.2/5.3/5.4 ✅; iter-1..5b layer-skel structural fixes ✅; iter-6a..f 6 轮盲改 silent drop 无果 (revert); **iter-7 transplant isolate + embed tolerance Transform Group bytes** ✅ variant #2 PASS; **iter-8 同思路 embed Rect + Fill body bytes** ✅ variants #2-7 全 PASS layers=1, layer class=ShapeLayer。Phase 5 ship gate **全闭环**。V2.2.1 留 Ellipse/Path/Stroke + Color encoding + keyframe 持久化 |
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

### 2026-05-26 V2.2 Phase 6 docs sync + workshop v0.8.0 frontmatter migration (PASS 175 不变)

**Workshop schema migration (v0.4.0 → v0.8.0)**: 10 scars + 3 playbooks 全补 `when_to_read` + `applies_to` + `last_updated` frontmatter (v0.8.0 hard-fail rule)。粗粒度可路由 not 完美 taxonomy. recheck 13/13 PASS.

**Phase 6 docs sync** — 3 文件:
- `docs/shape.md` (+83 行) — 新 "V2.2 alpha — Builder API" section: API surface table (`NewProject / NewComposition / NewShapeLayer / RootGroup / AddRect / AddFill / SetSize / SetColor`) + example (一个红色 200×200 矩形) + V2.2 alpha 限制 (runtime-only / 不支持 shape kind / Color 编码不准 / keyframe 不持久化) + RE 路线说明 (iter-1..8 总结 + embed pattern)
- `workshop/plans/coverage.md` (+14 行) — Shape 段新加 "V2.2 alpha ShapeLayer 写路径" 子段, 含 V2.2.1 候选清单
- `workshop/specs/v2-2-layer-creation-design.md` (+116 行) — finalize:
  - §6.4a 标 closed (原 pending RE items 改为 "iter-7/8 embed approach 闭环, byte-level 工作转给 V2.2.1+")
  - §6.5 confirm AECapabilities 仍空 struct ship
  - 新 §10 "Phase 5/6 实际 shipped — iter-7/8 embed approach": 原 design vs 实际 shipped 差距表 + Embed boilerplate pattern (3 个 container 大小 + 实现模式 + rifx.ReadChunk API) + transparent wrappers 清单 (still from-scratch) + V2.2 alpha 限制 (V2.2.1 候选) + RE methodology (transplant + embed 4-step) + Phase 5 ship gate 5 条永久教训

V2.2 alpha 现 **完整闭环**: Phase 5 (iter-1..8) ship gate PASS + Phase 6 docs sync done. 下次会话进 V2.2.1 (Ellipse/Path/Stroke + Color encoding + keyframe persist) 或 V3 方向。

### 2026-05-25 V2.2 Phase 5 **iter-8 完整闭环 — Phase 5 ship gate 全 PASS** (PASS 175 不变；variants #2-7 全 layers=1)

iter-7 解 Layr Transform 后, iter-8 用同样 transplant 法 isolate shape content silent drop. 三个 swap 测试 (`swap_rect_body` / `swap_fill_body` / `swap_both_bodies`) 证明每个 shape body 都被 AE 独立校验, 任一 broken 就 silent drop.

**Solution (iter-7 pattern, 更细 granularity)**:
- `tmp_debug/extract_shape_bodies/main.go` 抽 tolerance Rect + Fill body → `templates/v2_2_shape_rect_body.bin` (448B) + `v2_2_shape_fill_body.bin` (426B), 各 5-child
- `lower_shape_node.go::lowerRectNode` / `lowerFillNode` 重写: `//go:embed` + sync.Once cache + `cloneChunk` + `overwriteShapeStreamCdat`. Rect Size cdat 写 user value [w, h] f64 BE; Fill Color cdat 写 [r, g, b, a] f64 BE
- 删 dead code: 这俩 from-scratch path 不再需要 (Direction / Position / Roundness sub-prop placeholders + LowerColorStream / LowerFloat64Stream calls for Fill Opacity 等)
- 测试调整: `TestV2_2_CanonicalShapeGraph_Roundtrip` Rect Size 期望从 "Animated 2 keyframes" 改为 "Static fallback first kf value [50,50]" (跟 iter-7 Layr Position 同 pattern)

**AE bisect 验真**: 6 个 variants 全 PASS layers.length=1, layer[1]: name=L class=ShapeLayer enabled=true.

**V2.2 alpha 限制清单 (Phase 6 docs 要声明)**:
- ShapeLayer Layr-level Transform: 仅 Position static 持久化; Anchor/Scale/Rotation/Opacity 全 runtime-only
- ShapeLayer Position keyframes: 不持久化 (first kf 作 static fallback)
- Rect: Size static 持久化; Position / Roundness / Direction 全 runtime-only; Size keyframes 不持久化
- Fill: Color 持久化 (但编码不准 — JSX 0.5 → 0x406fe0... ≈ 255, 实际显示色可能不对); Opacity / Blend Mode / Composite Order / Fill Rule 全 runtime-only
- Ellipse / Path / Stroke: V2.2 不支持 (Go-side 能 emit + parse, 但 AE 不会显示)
- 全部 keyframe 不持久化 (lower 端只覆 cdat scalar)

**V2.2.1 候选 (subplan)**:
- Ellipse/Path/Stroke embed bytes: 需用户用 AE create 各 shape fixture (`re_v2_2_ellipse.aep` 等), 抽 body, embed
- Fill Color 编码 RE: tolerance bytes 跟 user 0-1 输入不对齐, 需 RE 编码方式
- keyframe 持久化: embed body 现只 static cdat slot, 加 LIST(list) lhd3/ldat 编码 = byte-level RE

**iter-8 永久教训**: iter-7 的 embed approach 是 generalizable. 任何 "complex multi-stream property container" 类 silent drop 都用同套法 (transplant isolate → extract bytes → embed → cdat 覆值). V2.2 ship gate 全程印证: byte-level RE 是 dead end (iter-6); semantic-level transplant isolation 是 right tool.

### 2026-05-25 V2.2 Phase 5 **iter-7 跨第一道 silent-drop 闸门** (PASS 175 不变；variant #2 empty ShapeLayer 首次 PASS layers=1)

iter-5b 后用户陪跑 7 轮 AE bisect 测 6 个候选 (iter-6a tdsb 0x03 stamping / 6b/c Anchor Vec3 / 6d Position tdum/tduM / 6e per-layer trailing chunks / 6f placeholder tdsb 0x03)，全没动 silent drop。GPT 写 `workshop/wip/gpt` pivot 建议: **停 byte-patching, 走 semantic RE**。三步:

1. **verify_baseline 排除 measurement bug** (新 jsx + go tool 跑单变体 AE, dump 完整 metrics: items / comp / layer class / activeItem / selection)。Tolerance.aep 跑出 PASS layers=1 + layer[1] class=ShapeLayer ✓ → measurement OK, silent drop 是真问题。
2. **Transplant 法 isolate**: 
   - `transplant_layr/`: tolerance.aep 整 Layr 换成 ours → layers=0 → 触发器在 Layr 内部
   - `transplant_tdgp/`: tolerance Layr 的 LIST(tdgp) 换成 ours → layers=0 → 触发器在 outer property tree
   - `swap_propgroup/`: batch 5 变体, 各 swap 一组 (RootVectors+Transform+LayerStyles / 4 placeholders / Transform only / RootVectors only / LayerStyles only) → 唯一 drop = 含 Transform 的 swap → **真凶 = Transform Group body**
   - `swap_reverse/`: 反向验证 (ours base + tolerance Transform) → PASS layers=1 ✓✓ 100% 确认
3. **iter-7 解法 — embed tolerance bytes**:
   - `tmp_debug/extract_transform_group/main.go` 抽 tolerance "Nested" Layr Transform Group body LIST(tdgp) (15 children: tdsb + tdsn + 6 stream tdmn-LIST pairs + Group End) → `internal/aep/templates/v2_2_transform_group_body.bin` (1842 B)
   - `lower_layer.go` `//go:embed` + `rifx.ReadChunk` (new public Parse-single-chunk API) + `sync.Once` cache + `cloneChunk()` deep clone helper
   - `lowerLayerTransform` 重写: 不再 from-scratch 构造, 直接 clone embedded body + `overwriteScalarCdat(body, "ADBE Position_0/1", val)` 覆写 cdat scalar 用 runtime 值
   - V2.2 限制声明: Anchor Point / Scale / Rotate Z / Opacity 是 runtime-only (不持久化)，Layr Position 仅 static 持久化 (keyframes 用 first kf 作 static fallback)；V2.3 RE byte 布局做全 Transform 持久化
   - `TestV2_2_CanonicalShapeGraph_Roundtrip` 调整: Layr Position 期望从 "2 keyframes" 改为 "static fallback = first kf value [0,0]"
   - 删 iter-5b/6 的 dead code: `lowerOrientationDefault` / `splitVec2Stream` (boilerplate 包含 otst + Position_0/_1 已 ready-formed)

**结果**:
| variant | iter-7 前 | iter-7 后 |
|---|---|---|
| #2 empty ShapeLayer | drop | **PASS layers=1, layer[1] name=L class=ShapeLayer enabled=true** ✓ |
| #3-7 (with Rect/Fill/kf) | drop | drop (shape content 级 silent drop) |

**剩 iter-8 work**: variant #3+ silent drop 触发器 = shape content. 用同样的 transplant + embed 法 isolate Root Vectors Group / Rect body。**不要再盲改**。

**iter-6 教训留 scar**: byte-level patching 阶段已结束。下次 ship-gate-class 问题第一步用 transplant 法 isolate 真凶到具体 chunk 后再决定怎么修。GPT 在 `workshop/wip/gpt` 的 pivot 文档救了 ~3 iter 弯路。

### 2026-05-25 V2.2 Phase 5 iter-5b — Gide + Ewst layer-skel boilerplate (174 → 175 PASS, +1)

iter 5 ship 后用户跑 bisect_v2_2，**6 个变体都 PASS 但 `comp.layers.length=0` 全 0** —— 包括 variant #2 (empty ShapeLayer, 零 shape kid)。GPT 看完 bisect 输出指出："connect variant #2 也被丢弃，说明问题已经不在 shape payload (cdat/tdum/tduM) 层，而是在 **layer instantiation / comp membership** 层"。iter 5 的 Root Vectors Group 5-层嵌套虽然结构对了 (跟 tolerance.aep 同形)，但不是 silent-drop 的真因 —— AE 还没走到 shape-content validation。

立刻 pivot 到 layer-root 级 byte-diff `tmp_debug/minfail_v2.aep` (empty ShapeLayer, AE silent-drop) vs `test_data/v2_2_shape_tolerance.aep` (AE accepted)，找出 **2 个 layer-skel 真正缺的 chunk**：

- **`LIST Gide` (Layr 第 4 child)** — 每个 AE-saved Layr 必有的 boilerplate。内容跨 11 层 (1 user "Nested" + 1 DLay + 6 SLay + 3 CLay + 1 SecL) **byte-identical**:
  ```
  LIST Gide (2 children)
    chunk gdta (8 B) = 00 × 8
    LIST list (1 child)
      chunk lhd3 (52 B) = 00d00bee 00000000 00000000 00000001 00000010 00000002 00000001 00000002 00000000 × 16
  ```
  "Gide" 推测 = layer-side guide/handle；52B lhd3 跟 keyframe-list 同 chunk ID 但不同语义，treat opaque。
- **`LIST Ewst` (Item-level sibling, 0 children)** — 紧跟每个 Layr 之后。Item LIST 里布局是 `Layr → Ewst → Layr → Ewst → ...`，12 个 Layr 各带一个 Ewst。template service layers 自带 Ewst (template 里就有)；只有 V2.2 NewShapeLayer 加的 user Layr 缺这俩。

修改:
- `rifx.go`: 加 3 个 chunk ID 常量 `IDGide / IDGdta / IDEwst`
- `lower_layer.go::lowerShapeLayer`: 把 Gide boilerplate (新 `makeGideBoilerplate()` helper + `gideLhd3Boilerplate` 52B 常量) 加到 Layr 第 4 child
- `new_layer.go::NewShapeLayer`: insert Layr 时同时 insert Ewst sibling (slice grow by 2 / shift by 2)
- 新测试 `TestNewShapeLayer_EmitsEwstSibling` (+1 PASS) + 给 `TestLowerShapeLayer_EmptyHasLayrChunk` 加 Gide 断言
- `tmp_debug/dump_gide/` 新工具 — dump 每个 Layr 的 Gide 内容用作 byte-diff baseline

验证:
- `tmp_debug/iter5_check.aep` (gen_iter5_check 重跑) 的 Item-level structural counts 跟 tolerance.aep **完全一致** (1 Layr/4-child + 1 DLay + 6 SLay + 3 CLay + 1 SecL + 12 Ewst)
- dump_gide 跑 iter5_check vs tolerance — Gide content byte-identical

**下一步 (用户)**: 重跑 `go run tmp_debug/bisect_v2_2/main.go`，关键看 variant #2 的 `comp=Main layers.length=` 是否从 0 变 1。若变 1 = Phase 5 ship gate 真闭环；若仍 0 = 还有其它 layer-skel chunk 缺。

**永久教训**: 接 silent-drop 类 ship gate 第一步先确定 "是 hard reject 还是 silent drop" + 用最简变体 (empty layer) bisect 定位 — drop 在 instantiation 还是在 content。本来 iter 5 想做 shape-content fix，但 bisect 显示 variant #2 (零 content) 也 drop，立刻能 pivot。GPT 反馈 + bisect 数据 catch 到这个，否则 iter 6/7 可能继续走 cdat padding 的弯路。

### 2026-05-25 V2.2 Phase 5 iter 5 — Root Vectors Group 5-层嵌套 emit (173 → 174 PASS, +1)

iter 4 留下"AE 2025 opens file but layers.length=0"问题；iter 5 假设 = Root Vectors Group 嵌套深度。byte-diff tolerance.aep vs re_shapes.aep 验证：**所有 AE-saved fixture 的 ShapeLayer 都是 5-层嵌套**（Root Vectors Group → Vector Group → Vectors Group → 用户 shapes / + Vector Transform Group(empty) / + Vector Materials Group(empty)）。我们 iter-4 emit `Root → [shapes 直挂]` 跳过两层 wrapper，AE 视为 malformed → silent drop。

- **`lower_shape_node.go::lowerVectorGroup`** 重写出 5-层嵌套：
  - **Vectors Group body** (innermost): tdsb(**0x00000401**) + tdsn + N×(tdmn shape + LIST tdgp body) + Group End
  - **Vector Group body** (middle, 9 child): tdsb(0x00000001) + tdsn + Vectors Group + Vector Transform Group(empty) + Vector Materials Group(empty) + Group End
  - **Root Vectors Group body** (outermost): tdsb(**0x00000401**) + tdsn + tdmn(Vector Group) + LIST + Group End
- **新 `makeTdsbContainer()`** (lower_property_stream.go) 出 0x00000401 — user-extensible container flag。AE 在 Root Vectors Group + Vectors Group 这两个 "user-addProperty 入口" 设此位；Vector Group routing body + empty placeholders + leaf tdbs 保持 0x00000001。
- **`hydrate_shape.go::hydrateVectorGroup`** 镜像改写为 `collectShapeKids` 递归 helper — 透明穿透 Vector Group / Vectors Group wrapper，把所有 typed shape 节点 (Rect/Ellipse/Path/Fill/Stroke) flat 收集到 `shapeRootGroup.Children`。Vector Transform Group / Vector Materials Group 忽略 (V2.2 不暴露 per-group transform / materials)。
- **`WrapShapeLayer` 语义变更**: 现在 always 标记 `layer.shapeDirty = true`。原 contract "shapeDirty=true 仅 NewShapeLayer 后" 是 iter-4 之前 hydrate 不完整的临时防御 (怕 sync 把 hydrate 没还原的内容洗掉)。iter 5 hydrate 完整后 contract 改：**WrapShapeLayer 是 V2.2 opt-in 信号 — 调它 = 承诺用 V2.2 mutation API + write-sync 从 runtime tree re-lower**。V1-only path (Property.SetStaticValue) 不经 WrapShapeLayer 不受影响。
- **测试**: `TestLowerVectorGroup_EmitsRootVectorsGroupChildren` 改写校验新结构 (root body 5 child + 必须含 Vector Group / Vectors Group / Vector Transform Group / Vector Materials Group 四 tdmn)。`TestV2_2_MutateExistingShape` 之前因 hydrate 找不到 shape 节点 silent skip，iter 5 后实际跑通 → +1 PASS。
- **byte-diff 验证**: `tmp_debug/gen_iter5_check/main.go` 出 iter5_check.aep (1 ShapeLayer with Rect 200×100 + Fill gray, 仿 tolerance.aep 内容)。dump matchName 序列 + tdsb 标志位跟 tolerance.aep **完全一致** (49-132 行 Root Vectors Group 子树 byte-for-byte 结构同形)。
- **残留差异 vs tolerance.aep (V2.2 always-emit 策略 + fix C 未启动)**:
  - 我们 emit Rect 11-child body (Direction + Size + Position + Roundness)；tolerance elide-default 只 emit Size
  - 我们 emit Fill 13-child body (Blend Mode + Composite Order + Fill Rule + Color + Opacity)；tolerance 只 Color
  - 我们 tdsn 带 display name ("Rectangle Path" / "Size" / "Fill" / 等)；tolerance 用 "" 或 "-_0_/-" 占位串
  - 我们 cdat 48B (Vec2 spatial)；tolerance 80B + tdum(8B) + tduM(8B) — **fix C 候选**: spatial property per-dim padding + bbox bound chunks emit
- **commits**: 1 commit 本 iter — `lower_shape_node.go` (refactor lowerVectorGroup) + `lower_property_stream.go` (makeTdsbContainer) + `hydrate_shape.go` (collectShapeKids) + `types_core.go` (WrapShapeLayer 标 dirty) + 2 test 调整 + `tmp_debug/dump_root` (tdmn/tdsn 解码升级)
- **下一步**: 用户跑 AE 2025 bisect 验真。若 AE layers.length=1 → Phase 5 闸门通过 → Task 5.5/5.6。若仍 =0 → fix C (tdum/tduM + cdat padding)。

### 在此之前（≤ 2 周前里程碑摘要）

- Wave 3 主体完工：文本扩展 7 字段 / alternateSource (Media Replacement) / manual kerning R/W / variable fonts axes R / `fontLocation` negative finding
- Wave 2 主体完工：trackMatteLayer / TimeRemap / LightKind / DisplayStartTime；negative: `dropFrame` runtime-only / `timeRemapEnabled toggle` structural / ldta 第三长度未发现
- Wave 1 主体完工：CapsOption / BaselineOption / StrokeOverFill / 段缩进×5 / AutoHyphenate
- 文档骨架重构：CLAUDE.md / COORDINATOR.md / REFERENCE.md 三件套（后整合到 `workshop/`）
- 之前：文本 16 setter + AddFont / Composition 标志位 + pixelAspect / Keyframe 增删 / Item-level comment + label / Project bitsPerChannel
