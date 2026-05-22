# V2.2 ship gate FAIL — Layr 结构 + Transform schema + tdum/tduM 缺

2026-05-23 — Phase 5 Task 5.7 first run. AE 2025 + AE 2020 都拒收 V2.2 builder 产的 canonical 3-ShapeLayer .aep。

## 错误信号

- **AE 2025**: 文件损坏弹框 — "默认 / imager / 颜色管理 / 自定义 渲染设置" 可能无效（典型 parse-time hard reject）
- **AE 2020**: "unexpected error" 弹框（用户选不保存关掉）+ items.length 留 0 → verify_v2_2.jsx 的 `items[1]` 报 "值 1 不在 1..0 范围内"。也是 parse fail。

## Diff 方法

留下两个 tmp_debug 工具:
- `tmp_debug/gen_canonical_failing/` — 重建 AE 拒的 canonical aep 到 `tmp_debug/v2_2_canonical_failing.aep`
- `tmp_debug/dump_chunks/<path>` — chunk 树 dump

跑:
```bash
go run tmp_debug/gen_canonical_failing/main.go
go run tmp_debug/dump_chunks/main.go tmp_debug/v2_2_canonical_failing.aep > tmp_debug/dump_failing.txt
go run tmp_debug/dump_chunks/main.go test_data/v2_2_shape_tolerance.aep > tmp_debug/dump_tolerance.txt
```

tolerance.aep 是 AE 自己存的 nested-Rect+Fill ShapeLayer fixture (Task 5.3 出)，结构 AE 认。Diff 二者即看 V2.2 lowering 哪里跟 AE 不一致。

## 找到的 3 个结构差异（按优先级）

### A. Layr 子树缺外层 LIST(tdgp) 包裹

**Failing (ours)**:
```
[LIST Layr]
  ldta (160 B)
  Utf8 (19 B)
  tdmn = ADBE Root Vectors Group (40 B)       ← flat 在 Layr 直接 child
  [LIST tdgp]                                   ← root vectors body
    tdsb / tdsn / ...
  tdmn = ADBE Transform Group (40 B)           ← flat 在 Layr 直接 child
  [LIST tdgp]                                   ← transform body
    ...
```

**Tolerance (AE-accepted)**:
```
[LIST Layr]
  ldta (164 B)
  Utf8 (6 B)
  [LIST tdgp]                                   ← OUTER wrapper
    tdsb (4 B)
    tdsn (8 B)
    tdmn = ADBE Root Vectors Group (40 B)
    [LIST tdgp]                                 ← root vectors body
      ...
    tdmn = ADBE Transform Group (40 B)
    [LIST tdgp]                                 ← transform body
      ...
    tdmn = ADBE Layer Styles (40 B)
    [LIST tdgp]
      ...
    tdmn = ADBE Group End (40 B)
```

AE 把 **所有 property group 套在一个外层 LIST(tdgp)** 里。我们直接平铺给 Layr —— AE parse 时把 tdmn 跟后续 LIST 配对失败，整个 Layr 就废了。

**Fix**: `lower_layer.go::lowerShapeLayer` 加 outer `LIST(tdgp)` wrapper：
```go
outer := &rifx.Chunk{ID: IDList, FormType: IDTdgp}
outer.Children = append(outer.Children, makeTdsb(), makeTdsn(""))
outer.Children = append(outer.Children, makeTdmn("ADBE Root Vectors Group"), rootGroupTdgp)
outer.Children = append(outer.Children, makeTdmn("ADBE Transform Group"), transformBody)
// + ADBE Layer Styles? ADBE Effect Parade? ADBE Masks Group? 都要看 AE 实际有没有 emit
outer.Children = append(outer.Children, makeTdmn("ADBE Group End"))
layr.Children = append(layr.Children, ldta, utf8Name, outer)
```

需要 review tolerance.aep 看完整的 outer tdgp 子项清单 — Layer Styles / Effect Parade / Masks Group / Material Options / etc 都可能要补 empty placeholder。

### B. Transform Group schema 选错 — 必须 6-axis 不能 2D

board permanent knowledge line 41 早警告:
> ShapeLayer Layr 级 Transform Group 是 6-axis form (Position_0/_1, Orientation, RotateX/Y, Envir Appear) — 不是 2D user-facing 5-stream。V2.2 lowering 选 5-stream user-facing form（与 V1 parser convention 一致），与底层 6-axis 差异 Phase 4 roundtrip 校

Phase 4 Go roundtrip 通了（Go parser 容错），但 AE 实测**不接受 2D 5-stream form**。

**Failing**: `ADBE Anchor Point / ADBE Position / ADBE Scale / ADBE Rotate Z / ADBE Opacity`
**Tolerance**: `ADBE Position_0 / Position_1 / Orientation / Rotate X / Rotate Y / Envir Appear`

**Fix**: `lower_layer.go::lowerLayerTransform` 改 emit 6-axis schema。runtime `LayerTransform` 仍暴露 2D API (Position/Scale/Rotation/Opacity)，但 lowering 把 2D Position[x,y] 拆到 Position_0 (X) + Position_1 (Y); Scale 类似？或 Scale 走另一路？需 RE 一次实际 AE-saved shape layer 的 Scale 在哪 stream 里。

可能 Scale + Rotate Z + Opacity 在另一个 sibling group（Layer Styles? Material Options?）。Phase 0 RE-S2 只 dump 了 Transform Group 内 6 axis，没看 Scale 在哪。 **Phase 5 fix 前先补 RE**。

### C. tdum/tduM 在 spatial property tdbs 里 缺

**Tolerance (Vector Rect Size)**:
```
[LIST tdbs]
  tdsb / tdsn / tdb4 / cdat (80 B) / tdum (8 B) / tduM (8 B)
```

**Failing (Vector Rect Size, animated → no cdat)**:
```
[LIST tdbs]
  tdsb / tdsn / tdb4 / [LIST list](lhd3 + ldat)
```

tolerance cdat 是 80B 不是我们的 48B → 也涉及 cdat per-dim padding 差异。

tdum/tduM 估计是 min/max spatial bound（uniform pair? 64-bit float each）。AE 用来加速 spatial-bound 计算。spatial property (Position 类) 都有；scalar (Opacity / Rotation) 没有。

**Fix**: `lower_property_stream.go` 给 spatial-style stream (`spatial: true` valueLayout) 在 cdat 后 append tdum (8 B) + tduM (8 B)。内容初步推断 = current value 的 min/max bound（静态时 min==max==current; 动画时 = keyframe min/max envelope）。需 hex-dump tolerance.aep 验。

## Phase 5 fix order

**注**: fix A 已落 (commit 见 board)。AE 2025 仍同错 — 单 A 不够。需 + B + C 至少。

~~1. **C 先**（最 local）: tdum/tduM emit。cdat padding 算清楚再写。~~
~~2. **B**: Transform 6-axis schema。~~
~~3. **A 最后**（最 invasive）: Layr outer tdgp wrapper。~~

修正后顺序（基于 fix A 已落 + AE 仍拒）:

1. **A done** — Layr 外层 LIST(tdgp) wrapper 已加。outer 现含 Root Vectors Group + Transform Group + Group End。Test 全过，AE 仍拒同错。**结论**: wrapper 是结构必需，但 outer 内容还不够。
2. **下一步候选** (按 AE 错误信号"默认色彩管理无效"猜):
   - **Layer Styles placeholder**: outer 加 `tdmn(ADBE Layer Styles) + LIST(tdgp, empty with Blend Options + 10 fx/enabled subprops + Group End)`。tolerance dump line 144-211 是完整列表。AE 也许严格要求这个 placeholder。
   - **Material Options placeholder**: outer 加 `tdmn(ADBE Material Options Group) + LIST(tdgp, 16+ default scalar props)`。tolerance dump line 222-345。
   - **Extrsn Options / Audio Group / Layer Sets**: 三个 empty placeholder。
3. **B Transform schema** (6-axis): 单独验。可能 AE 2D 5-stream 也接受（runtime API 角度），但 tolerance 显示 AE 自己存的是 6-axis。先 RE 一个 user-created shape layer Transform（不是 nested，是平 Shape） — 看 AE 存 Scale/Opacity 到哪。
4. **C tdum/tduM**: 看是否仅 placeholder 不够 / 还得加这些细节。

诊断方法 (token 用完前留): 每次加一组 placeholder → rebuild canonical → AE 试 → 错误变了说明那组是 critical → 继续；错误同说明无关 → revert 那组。

## fix A 实施记 (本 commit)

- `lower_layer.go::lowerShapeLayer`: outer `LIST(tdgp)` wrapper 加，含 tdsb + tdsn("") + (Root Vectors Group / Transform Group pair) + Group End。Layr direct children 现仅 `[ldta, Utf8, outer LIST(tdgp)]`。
- `hydrate_shape.go::hydrateShapeNodes`: 递归 visit 子树找 Root Vectors Group（不再只看 Layr 直 child）。
- `types_core.go` + `new_layer.go` + `sync_shape_layers.go`: 加 `Layer.shapeDirty bool`；NewShapeLayer 设 true；syncShapeLayerChunks 只对 dirty layer re-lower（防 parser-loaded layer 被 hydrate-roundtrip 损坏 — hydrate 还没支持 nested VectorGroup 等）。
- `lower_layer_test.go`: `TestLowerShapeLayer_WithShape_HasRootVectorsGroup` 测改用递归 walk 找 Root Vectors Group tdmn（不再 assume Layr 直 child）。

PASS 173 不变，0 FAIL，vet clean。

## 永久教训 → CLAUDE.md / scars

V2.2 Phase 4 Go roundtrip PASS **不代表 AE 接受**。Go parser 写 tolerant，AE parse 严格。下次类似 "writer + ship gate" 流程 phase 顺序要把 ship gate 提前。
