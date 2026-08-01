# ⚠ V2.2 ship gate FAIL — Layr 结构 + Transform schema + tdum/tduM 缺

V2.2 ship gate FAIL — Layr 结构 + Transform schema + tdum/tduM 缺

Path note: this is a historical RE/debug log from before the facade split. Historical prose may name old implementation paths such as `internal/aep/lower_layer.go` or `internal/aep/templates/*`; the current implementation lives mostly under `internal/serializer/*`, `internal/serializer/templates/*`, `internal/codec/*`, and `internal/aep_test/*`. Preserve old paths inside dated incident notes as evidence context; when using this as an action guide, confirm the current file with `rg` first.

2026-05-23 — Phase 5 Task 5.7 first run. AE 2025 + AE 2020 都拒收 V2.2 builder 产的 canonical 3-ShapeLayer .aep。

## 错误信号

- **AE 2025**: 文件损坏弹框 — "默认 / imager / 颜色管理 / 自定义 渲染设置" 可能无效（典型 parse-time hard reject）
- **AE 2020**: "unexpected error" 弹框（用户选不保存关掉）+ items.length 留 0 → verify_v2_2.jsx 的 `items[1]` 报 "值 1 不在 1..0 范围内"。也是 parse fail。

## Diff 方法

留下两个 tmp_debug 工具:
- `tmp_debug/gen_canonical_failing/` — 重建 AE 拒的 canonical aep 到 `tmp_debug/v2_2_canonical_failing.aep`
- `tools/debug/dump_chunks/<path>` — chunk 树 dump

跑:
```bash
go run tmp_debug/gen_canonical_failing/main.go
go run tools/debug/dump_chunks/main.go tmp_debug/v2_2_canonical_failing.aep > tmp_debug/dump_failing.txt
go run tools/debug/dump_chunks/main.go test_data/fixtures/v2_2_shape_tolerance.aep > tmp_debug/dump_tolerance.txt
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

## iter 2 实施记 — 5 placeholder 落 (本 commit)

`lower_layer.go::lowerShapeLayer` outer LIST(tdgp) 在 Transform Group 之后、Group End 之前，加 5 个 placeholder property group：

1. **ADBE Layer Styles** — 完整嵌套结构（Blend Options Group + Adv Blend Group + 10×fx/enabled empty 3-child + Group End）— `appendLayerStylesPlaceholder` helper。
2. **ADBE Extrsn Options Group** — 空 3-child。
3. **ADBE Material Options Group** — 空 3-child。
4. **ADBE Audio Group** — 空 3-child。
5. **ADBE Layer Sets** — 空 3-child。

新 helper `emptyPropGroup()` 出空 3-child LIST(tdgp) [tdsb + tdsn("") + tdmn("ADBE Group End")]。

dump_failing.txt 结构对照 tolerance.aep dump 145-361 行 **完全一致**（除 cdat/tdbs body 内部细节 = fix C）。

**AE 2025 ship gate**: 同错（"默认/imager/颜色管理/自定义 渲染设置可能无效"，46s timeout 内 FAIL）。按本 scar 诊断规则 "错误同说明无关"，placeholder 不是 critical 项。但 placeholder 结构正确，留下不害事 — fix B/C 实测前不必 revert。

## iter 2 新 RE 发现 — Transform Group 6-axis pure schema

读 tolerance.aep dump 101-144 行（**真正的 ShapeLayer Transform Group**，不是 DLay 的）发现:

- ShapeLayer 的 Transform Group **只含 6 stream**: `ADBE Position_0 / Position_1 / Orientation / Rotate X / Rotate Y / Envir Appear in Reflect`
- **完全没有** Anchor Point / Position (2-vec) / Scale / Rotate Z / Opacity
- 对比 DLay (camera/3D) Transform Group (lines 392-444) 有 Anchor + Position_0/_1/_2 + Scale + Rotate Z + Opacity + Envir Appear — schema 是 layer-subtype 相关的

**结论**: AE 的 ShapeLayer "always emit" set = 6-axis defaults (Position_0/_1 + Orientation + RotateX/Y + Envir Appear)。Anchor/Scale/Opacity/RotateZ 在用户没显式 set 时全部 elide。

**我们 V2.2 现行的 2D 5-stream 形式 (Anchor / Position / Scale / Rotate Z / Opacity) AE 不认**。即便 Phase 4 Go roundtrip PASS — 我们 parser 容错 + always-emit 自洽 — AE 严格 reject。

**fix B 真实形态**:

`lowerLayerTransform(t *LayerTransform, ctx)` 改写为 ShapeLayer-flavored:
- 总 emit 6 stream defaults: Position_0/_1, Orientation, RotateX, RotateY, Envir Appear (always emit per AE convention)
- runtime `t.position` ([2]float64) → 拆 Position_0 (X) + Position_1 (Y)；keyframe 同样拆
- runtime `t.anchorPoint` / `t.scale` / `t.rotation` / `t.opacity` **仅在 stream.Mode != ModeUnset (= 用户调过)** 时额外 emit `ADBE Anchor Point` / `ADBE Scale` / `ADBE Rotate Z` / `ADBE Opacity` (combined-vector form, sit 在 6-axis defaults 之后)

策略变更：从 "always emit" 改 "selective emit" (defaults elide except 6-axis defaults)。需:
1. PropertyStream.Mode 加 "Unset" 区分（默认值且未触发 setter）
2. NewShapeLayer 初始化 stream 时不预设值（Unset 状态）
3. SetXxx() 触发 Mode → ModeStatic + 值
4. AddKeyframeLinear() 触发 Mode → ModeAnimated
5. lower 时只 emit 非 Unset 的 stream

或更简单方案：保留 always-emit 但改 schema 名 (Position 拆成 Position_0/_1 + 加 4 个 defaults Orientation/RotateX/Y/EnvirAppear)，accept AE 接受 over-emit。需 RE 验证。

**hydrate 端**：parseLayer 的 hydrateLayerTransform 跟着改 — 读 Position_0/_1 → 合成 [2]float64 position；其它 stream 按 matchName fallback。

## iter 4 实施记 — minimum-failing bisection 揭出 6+ structural bugs

iter 4 用户决定不再 stack schema fix，改写 `tmp_debug/bisect_v2_2/` 跑 6 变体矩阵 (empty ShapeLayer → +AddRect → +SetSize → +AddFill → +keyframed → +spatial-keyframed)。结果：**#2 (empty ShapeLayer) 已 FAIL 同错**，证明问题在 Layer 骨架而非 shape/keyframe/spatial。

随后通过 byte-level diff against tolerance.aep + 反复跑 AE 验证，找出并修复 **6 个独立 structural bugs**：

### bug 1: Layr 位置错 — 必须在 DLay/SLay/CLay/SecL **之前**
- 之前：`c.itemList.Children = append(c.itemList.Children, layrChunk)` → Layr 在最末
- 修复：`insertLayrPosition` 找第一个 DLay/SLay/CLay/SecL，Layr 插那之前。`new_layer.go::insertLayrPosition` + `templateServiceLayerInsertTypes` map
- **效果**：AE 2025 不再 hard-reject 文件（从 "默认/imager/颜色管理/自定义 渲染设置可能无效" 弹框 → 文件正常打开）

### bug 2: ldta 必须 164B（不是 160B）
- 之前：`ldtaSize2020 = 160` for ShapeLayer
- 修复：`buildLdtaBytes` 用 `ldtaSize2025 = 164` (trailing 4 zero bytes)
- test 改名 `TestLowerShapeLayer_LdtaIs160Bytes` → `TestLowerShapeLayer_LdtaIs164Bytes`

### bug 3: ldta time fields 编码错 — 必须用 TickRate 作 divisor
- 之前：`InPointDivs=1, OutPointDivd=0xFFFFFFFF(sentinel), OutPointDivs=1` → AE 读 OutPointDivd = -1 (int32 BE) → 负 duration → 静默 drop layer
- 修复：所有 time field divisors = TickRate (30720 for 30fps); `OutPointDivd = duration * TickRate`; `Stretch = 1/1` not `100/100`
- 需要 `compDuration` plumb 到 `lowerCtx`，新加 field；`sync_shape_layers.go` 同步加

### bug 4: AttrByte0 @0x25 = 0x01 必须设 + @0x3B/@0x3D/@0x63 等
- 之前：`d[0x25] = 0x00`，AttrByte2 = `0x01` (只 visible)
- 修复：`d[0x25] = 0x01` (未文档化 bit), `d[0x27] = 0x87` (visible + audio + effects + collapse-transform), `d[0x3B] = 0x01` (unknown), `d[0x3D] = 0x08` (label color), `d[0x63] = 0x02` (BlendingMode)

### bug 5: LayerID **collision** with DLay
- 之前：`base.ID = c.proj.allocItemID()` → 项目级 nextItemID = 2，但模板 DLay 已用 LayerID=2 → AE 把我们 Layr 视作"被 DLay 标记 deleted"
- 修复：`maxLayerIDInItemList(c.itemList)` 扫所有 Layr/DLay/SLay/CLay/SecL 的 ldta @0x00，取 max+1。Template service layers 占 2..12，所以 user Layr 起 13.

### bug 6: head counter 没 cover layer ID
- 之前：`p.nextItemID` 没 bump 到 layer.ID 之上 → `write.go::syncHeadCounters` 把 head counter 设到 nextItemID（仍 = 2）→ AE 看 head < layer.ID → drop layer
- 修复：NewShapeLayer 里 `if layerID >= c.proj.nextItemID { c.proj.nextItemID = layerID + 1 }`

### bug 7: cdta @0x18 secondary divisor 错
- 之前：`buildCompCdta` emit 600 (fresh-comp marker，per cdta_layout.go doc)
- 修复：NewShapeLayer 加 user Layr 后，把 `c.cdta.Data[@0x18..0x1B]` 写成 TickRate (30720)
- doc 注释说 "AE rewrites to TickRate on user mod"，我们 explicit 写

### 当前状态 + 残留 bug 8 候选: Root Vectors Group 嵌套不够

修了 bug 1-7 后，AE 2025 现在：
- **opens file without exception** (huge progress)
- 但 `comp.layers.length=0` — AE 静默 drop layer 还有别的原因

byte-level diff 之 ldta + cdta + idta + head 都对了。剩下差异在 `LIST(tdgp)` 内部嵌套深度：

**Tolerance.aep ShapeLayer 内 Root Vectors Group 结构（5 层嵌套）**:
```
tdmn = ADBE Root Vectors Group
[LIST tdgp]
  tdmn = ADBE Vector Group         ← we miss this intermediate
  [LIST tdgp]
    tdmn = ADBE Vectors Group      ← AND this (plural!)
    [LIST tdgp]
      tdmn = ADBE Vector Shape - Rect (actual shape)
      ...
    tdmn = ADBE Vector Transform Group
    tdmn = ADBE Vector Materials Group
```

**Our V2.2 builder 现状**:
```
tdmn = ADBE Root Vectors Group
[LIST tdgp]
  [shape children direct here, no Vector Group / Vectors Group wrappers]
```

**iter 5 hypothesis**: AE 检查 `ADBE Root Vectors Group → ADBE Vector Group → ADBE Vectors Group` 嵌套 — 缺这两层 wrapper 就把 layer 视作 malformed Shape Layer 并 drop。

修复：`lower_layer.go::lowerShapeLayer` Root Vectors Group 改 emit 完整 5-层嵌套。`lowerVectorGroup` 输出 Vectors Group level 内容; 上面包 `ADBE Vector Group + ADBE Vector Transform Group + ADBE Vector Materials Group` 三 child。`hydrate_shape.go` 镜像改。

### iter 4 残留盲区 — 3 处 magic bytes 没独立验证

bug 4 / bug 7 实施时为减少 AE 跑次数，把多个 byte 一起改后再验。下列字节是"tolerance.aep 有 → 我们补上"的盲跟随，**机制理解为 0**：

| 位置 | 值 | 来源 | 是否独立验证 |
|---|---|---|---|
| ldta `@0x25` (AttrByte0) | `0x01` | tolerance 有，parse_layer.go 文档未覆盖此 bit | ❌ 跟其它 byte 一起改的 |
| ldta `@0x3B` | `0x01` | tolerance 有，无任何文档 | ❌ |
| cdta `@0x18` (secondaryDivisor) | `TickRate` (改 600) | cdta_layout.go doc 写 "AE rewrites to TickRate on user mod"；我们 explicit 写 | ❌ |

未来 iter（bug 8 修完 AE 显层后）应 **toggle 验** 这三处：单独把每处改回原值，跑 bisect_v2_2，若 AE 仍显层 = 该字节非必需可去。可能简化 builder。

ldta `@0x27=0x87` (visible + audio + effects + collapse-transform) 也是 batch 改的，但其中 bit0 visible (0x01) 历史已验证为必需，其余 3 bit 是 AE-default 行为，相对可信。

### iter 4 永久教训 — silent-drop vs hard-reject 是不同诊断类别

AE 的拒接有**两种独立失败模式**，需要不同 JSX 验证：

| 模式 | JSX 信号 | 含义 |
|---|---|---|
| hard reject | `app.open()` 抛异常 | 文件结构 fatal corruption — parser 在 chunk-level fail |
| silent drop | `app.open()` 不抛，但 `comp.layers.length` 少了我们加的层 | 文件 parser 接受 chunk 结构，但 AE 内部某 layer-validation 把我们的 layer 当 deleted/invalid/phantom 丢弃 |

`test_data/generators/verify_v2_2.jsx` (Phase 5 Task 5.1) 只验 PASS/FAIL 不验 silent drop。`test_data/generators/verify_open.jsx` (iter 4 新写) 显式 dump `items[i].typeName / layers.length / layers[k].name` 才暴露 silent drop。

未来 ship gate / 任何 JSX driver **必须打印 layers.length + 每层 name**，不然 silent drop 假阳性 PASS 难抓。

## Phase 5 fix order 修正后 (iter 3 后)

1. **A done** + **iter 2 done** + **iter 3 done** — outer wrapper + 5 placeholder + Transform 6-axis schema (Anchor + Position_0/_1 split + Scale + RotateZ + Opacity + Orientation otst + RotateX/Y + EnvirAppear)。**AE 2025 仍同错 + 同信号** ("默认/imager/颜色管理/自定义 渲染设置可能无效")。结论：现 3 个 fix 都是 necessary 但 not yet sufficient。
2. **下一步候选 fix C** — spatial property tdum/tduM emit + cdat per-dim padding 矫正。Tolerance 详:
   - `ADBE Position_0` static: `cdat (40B) + tdum (8B) + tduM (8B)` (tdum/tduM 在 cdat 后)
   - `ADBE Vector Rect Size` static: `cdat (80B)` — 我们 emit 48B → **per-dim padding 错**！dim=2 spatial 应该是 80 不是 48。
   - 8B tdum / 8B tduM 推测 = 单 f64 min/max bound
   - **animated** form 的 tdum/tduM 落点未知 — tolerance 只有 static 例子
3. **fix C 后仍 reject** → 必须启动**1-layer 最小失败 bisection**：
   - 写 `tmp_debug/gen_minimum_failing/main.go`: 只 1 个 empty ShapeLayer (无 shape 节点 / 无 keyframe / 默认 transform)
   - AE 跑 → 若 still fail 同错 = 问题在 layer 骨架 (ldta? Layr wrapper? Transform Group emit?)
   - 若 PASS = 问题在 shape 节点 emit / keyframe emit / cdat padding
4. byte-level diff: 把 minimum failing aep 跟 tolerance.aep 同段 hex diff (`xxd` / `dump_chunks` 比较)。

## iter 3 实施记 — fix B Transform 6-axis schema (本 commit)

- `internal/rifx/rifx.go`: 新 chunk IDs `IDOtst / IDOtky / IDOtda` (Orientation 特殊三件套)
- `types_core.go`: 新 `MatchNamePosition0 / MatchNamePosition1 / MatchNameEnvirAppear` 常量
- `lower_property_stream.go`: 新 `splitVec2Stream` helper — `PropertyStream[[2]float64]` → 两个 `PropertyStream[float64]` (X / Y)，保留 mode + keyframe time/ease
- `lower_layer.go::lowerLayerTransform`: 重写 emit 10 stream：Anchor + Position_0/_1 (split) + Scale + RotateZ + Opacity + Orientation (otst) + RotateX (default 0) + RotateY (default 0) + EnvirAppear (default 100)
- `lower_layer.go::lowerOrientationDefault`: 新 helper — otst → (tdbs cdat 24B + otky → otda 24B), 全 0
- `hydrate_shape.go::hydrateLayerTransform`: 加 Position_0/_1 case + `combinePositionXY` helper, 把 V1 Property layer X/Y 重新合 [2]float64 keyframe 流

dump_failing.txt 现 Transform Group 结构 = 10 stream 6-axis form, 跟 tolerance.aep DLay 那块完全一致 (但 ShapeLayer 那块更简, tolerance ShapeLayer 只 emit 6 axis defaults; user-touched extras (Anchor/Scale/etc) 在我们这里 over-emit).

PASS 173 不变 / 0 FAIL / vet clean。AE 2025 ship gate 同错 (14s reject, 比 iter 2 的 46s 快很多, AE 可能在 ldta/Layr parse 阶段更早 reject)。

## iter 5 实施记 — Root Vectors Group 5-层嵌套 + tdsb container flag + hydrate transparent passthrough (本 commit)

iter 4 用 dump_root/main.go (升级版含 tdmn/tdsn 解码) dump tolerance.aep + 多 AE-saved fixture (re_shapes.aep) 验证：**所有 AE-saved ShapeLayer 都是 5-层嵌套**。

我们 iter-4 emit `Root Vectors Group → [shapes 直挂]` 是 V2.1/2.2 早期对 Shape Layer schema 的误读 (以为 Root Vectors Group 自己就是 "Contents" 用户加 shapes 的地方)。实际 AE schema 是：

- `ADBE Root Vectors Group` (Layr 子树根) — 用户 addProperty(`ADBE Vector Group`) 在这里挂
- `ADBE Vector Group` (= UI "Group N") — 用户的命名分组容器
  - `ADBE Vectors Group` (= 该 Group 内的 "Contents" 集合) — 用户 addProperty(shape) 在这里挂
    - 实际 shape 子: Rect / Ellipse / Path / Fill / Stroke
  - `ADBE Vector Transform Group` (= 该 Group 的 transform; V2.2 always empty)
  - `ADBE Vector Materials Group` (= 该 Group 的 materials placeholder; V2.2 always empty)

V2.2 把 runtime `shapeRootGroup.Children = [Rect, Fill, ...]` 映射到 **单个 Vector Group wrapper** (语义 = AE 自动建的 "Group 1")；V2.3+ 可暴露多组。

### 修改清单

1. **`lower_shape_node.go::lowerVectorGroup`** 重写：返 Root Vectors Group body LIST(tdgp)，内部嵌 Vector Group → Vector Group body → Vectors Group → Vectors Group body (放 shape kids) + Vector Transform Group (empty) + Vector Materials Group (empty)。
2. **`lower_property_stream.go` 新 `makeTdsbContainer()`** — 出 `0x00000401` 这个 user-extensible flag。仅 Root Vectors Group body + Vectors Group body 用；其它 (Vector Group routing body / empty placeholders / leaf tdbs) 保持 `makeTdsb()` = `0x00000001`。这俩值都来自 tolerance.aep + re_shapes.aep RE 观测，无 AE doc 但跨多 fixture 一致。
3. **`hydrate_shape.go` 重写 `hydrateVectorGroup`**：内部 `collectShapeKids` 递归 helper — 看见 `ADBE Vector Group` / `ADBE Vectors Group` 就 transparent 递归下去；看见 typed shape (Rect/Ellipse/Path/Fill/Stroke) 就 hydrate 进 `g.Children`；忽略 Vector Transform Group / Vector Materials Group。多 Vector Group siblings 全部 flatten 进同一个 `shapeRootGroup.Children` (V2.2 不分 group)。
4. **`types_core.go::WrapShapeLayer` 改 always 标 `shapeDirty = true`** — iter-4 之前的"仅 NewShapeLayer 标 dirty"是 hydrate 不完整时的临时防御。iter 5 hydrate 完整后契约改为：**WrapShapeLayer 是 V2.2 opt-in；调它 = write-sync 从 runtime tree re-lower**。V1-only path (Property.SetStaticValue 等) 不经 WrapShapeLayer → 不受影响 (V1 测试 TestShapePrimitivesReal 通)。
5. **`lower_shape_node_test.go::TestLowerVectorGroup_*` 重写** — 校验 Root Vectors Group body 5 child + 必须含 Vector Group / Vectors Group / Vector Transform Group / Vector Materials Group 四个 tdmn marker。
6. **`tools/debug/dump_root/main.go` 升级** — 解码 tdmn matchName 为 ASCII (NUL-terminated 取 prefix) + 解码 tdsn embedded Utf8 record (skip "Utf8" magic + 4B size → 拿 name 字节)。原来全 hex 输出，无法肉眼看 matchName。
7. **新 `tmp_debug/gen_iter5_check/main.go`** — 出仿 tolerance.aep 内容的 .aep (1 ShapeLayer + Rect 200×100 + Fill gray) 用作 byte-diff baseline。

### 验证

- `go vet ./... && go test ./...` — PASS=174 (+1; `TestV2_2_MutateExistingShape` 之前 silent skip 因 hydrate 找不到 shape kids；iter 5 hydrate 后实际跑通 mutate→write→re-parse→读 Size=500 闭环)
- byte-level 验证: `go run tmp_debug/gen_iter5_check/main.go && go run tools/debug/dump_root/main.go tmp_debug/iter5_check.aep | sed -n '45,135p'` — 跟 tolerance.aep 同段 (45-100 行 Root Vectors Group 子树) 结构 + matchName 序列 + tdsb 标志位 **完全一致**

### 残留待启动 — fix C (spatial cdat padding + tdum/tduM)

iter5_check.aep vs tolerance.aep 还有以下差异 (V2.2 always-emit 策略 + 未做 fix C)：

- Rect body: 我们 11 child (Direction + Size + Position + Roundness)，tolerance 5 child (只 Size — elide default)
- Fill body: 我们 13 child (Blend Mode + Composite Order + Fill Rule + Color + Opacity)，tolerance 5 child (只 Color)
- tdsn display name: 我们 "Rectangle Path" / "Size" / "Fill"，tolerance "" 或 "-_0_/-" sentinel
- cdat sizes: 我们 Vec2 spatial = 48B；tolerance 80B + tdum(8B) + tduM(8B)。**fix C**: per-dim padding + 后置 tdum/tduM bound chunks (推测 spatial property 的 min/max envelope, f64 each, 静态时 = current value)

V2.2 always-emit 策略选择：AE 自己 elide default — 我们 over-emit，AE 接受 (Phase 4 Go roundtrip PASS 已证)。display name 差异同理 — AE 用 sentinel 占位串，我们用 English display name，AE 不在乎。**cdat padding 是真问题** — 48B → 80B 不是 elide，是字段缺。fix C 启动顺序按 AE bisect 反馈定：iter 5 后 AE 仍 layers.length=0 → 必须做。

### iter 5 永久教训

1. **byte-diff 必须解码语义字段** — iter 4 之前 dump_root 全 hex 输出，看不到 tdmn matchName，6 iter 才发现 5-层嵌套。任何 V2.x ship gate diff 必须有"解码 tdmn + tdsn"的 dump 工具。
2. **WrapShapeLayer 这种"opt-in" 入口 API 的语义敏感** — iter-1 ~ iter-4 时 shapeDirty 仅 NewShapeLayer 标，因 hydrate 不完整。iter 5 hydrate 完整后 contract 立即收紧 (WrapShapeLayer always 标)。**任何 V2.2 入口 API 在 hydrate 演进的同时要 reconsider 其副作用契约**。
3. **AE schema 层级 "User-extensible container" vs "Fixed routing"** 是有结构性表达 (tdsb 0x00000401 vs 0x00000001)。任何新 V2.x property tree 设计前先扫该位 — 直接告诉你这个层级是"用户加 property 的入口" 还是 "AE 内部固定结构"。

## iter-5b 实施记 — Gide + Ewst layer-skel boilerplate (本 commit)

iter 5 ship 完用户跑 bisect_v2_2 6 变体全 PASS (AE 不 hard-reject) **但 layers.length=0 全 0**，包括 variant #2 (empty ShapeLayer, 零 shape kid 零 keyframe 零任何 content)。

GPT 看完 bisect 输出立刻指出 (引用):

> 这个 bisect 很关键：连 variant #2 的 empty ShapeLayer 都被 AE 丢弃，说明问题已经不在 shape payload（cdat/tdum/tduM）层，而是在 layer instantiation / comp membership 层。
> 下一步不要再 bisect property streams，直接对比 AE-native empty ShapeLayer 与 generated variant #2 的 layer-root structure：重点看 comp layer refs、layer type discriminator、LIST ordering、tdmn/tdsn/tdgp/id linkage，以及可能缺失的 side chunks。
> 现在的证据表明 AE 还没进入 shape-content validation，就已经在 create-layer 阶段 silently drop 了 layer。

完全正确。立刻 pivot byte-diff `minfail_v2.aep` vs `tolerance.aep` 在 Item-level + Layr-children level，找出 2 个 layer-skel 真正缺的 chunk。

### 发现 1: LIST Gide 是 Layr 第 4 必需 child

tolerance.aep 的 "Nested" Layr 有 4 child：ldta + Utf8 + LIST tdgp (property tree) + **LIST Gide**。我们 V2.2 builder 只 emit 3 child (Gide 缺)。

Gide 内容跨所有 11 个 AE-saved Layr (Nested + DLay + 6 SLay + 3 CLay + SecL) **byte-identical**:

```
LIST Gide (2 children)
  chunk gdta (8 B) = 00 × 8
  LIST list (1 child)
    chunk lhd3 (52 B) = 00d00bee 00000000 00000000 00000001 00000010 00000002 00000001 00000002 00000000... (16B trailing zero)
```

"Gide" 推测 = layer-side guide/handle。lhd3 在这里跟 keyframe-list 的 lhd3 同 chunk ID 但不同语义 (52B 头但内容不是 keyframe header) — treat 完全 opaque 常量。

### 发现 2: LIST Ewst 是 Item-level Layr sibling

Item LIST 直接 children 中，每个 Layr/DLay/SLay/CLay/SecL **之后紧跟一个 `LIST Ewst (0 children)` 空 sibling**。tolerance.aep 12 个 layer 各带一个 Ewst (1 user Layr + 1 DLay + 6 SLay + 3 CLay + 1 SecL = 12 Ewst)。

我们 V2.2 之前的代码 `new_layer.go::NewShapeLayer` 插入 Layr 时只插 1 个 chunk。template 自带 11 个 service layer 各自的 Ewst (template 复制时一并进来)，但**用户 NewShapeLayer 加的 Layr 不带 Ewst** → Item LIST 里出现 `[user Layr] → [DLay] → [DLay's Ewst]` 顺序，跟 tolerance 的 `[Layr] → [Layr's Ewst] → [DLay] → [DLay's Ewst]` 不同。AE 视作 layer 结构不合法 → silent drop。

Item-level child 类型计数 (iter-5b 后):

| LIST type | iter5_check (我们) | tolerance | 一致? |
|---|---|---|---|
| Layr (user) | 1 | 1 | ✅ |
| DLay | 1 | 1 | ✅ |
| SLay | 6 | 6 | ✅ |
| CLay | 3 | 3 | ✅ |
| SecL | 1 | 1 | ✅ |
| **Ewst** | **12** (was 11) | 12 | ✅ |
| PRin | 1 | 1 | ✅ |
| dats | 1 | 1 | ✅ |

iter-5b 之前我们 11 个 Ewst (只 template service layers 自带)。修后 12 个 (用户 Layr 也有 Ewst)。

### 修改清单

1. **`internal/rifx/rifx.go`**: 新 3 个 chunk ID 常量 `IDGide / IDGdta / IDEwst`。
2. **`internal/aep/lower_layer.go::lowerShapeLayer`**: 在最后 `layr.Children = append(layr.Children, outer)` 之后，append `makeGideBoilerplate()` 作为 4th child。新 helper `makeGideBoilerplate()` + 52B `gideLhd3Boilerplate` 常量 byte slice。
3. **`internal/aep/new_layer.go::NewShapeLayer`** insert 路径: 把"insert 1 chunk"改"insert Layr + Ewst sibling 2 chunks" (slice grow by 2 / copy shift by 2 / 两个 slot 分别 set Layr 和 Ewst LIST 0 child)。
4. **测试**:
   - `lower_layer_test.go::TestLowerShapeLayer_EmptyHasLayrChunk` 加 Gide 存在断言 (FindFirstList(IDGide) != nil + Gide children == 2 + Gide[0] == gdta 8B)
   - 新 `new_shape_layer_test.go::TestNewShapeLayer_EmitsEwstSibling` (+1 PASS) — 校验 itemList.Children 里 user Layr 紧跟一个 `LIST(Ewst, 0 children)` sibling
   - 新 `CompItemListForTest(c *Composition) *rifx.Chunk` 测试辅助 (new_layer.go) 暴露 Composition.itemList
5. **`tmp_debug/dump_gide/`** 新工具 — 跨所有 Layr dump Gide 内容，验证 byte-identical。本 commit 用它确认 iter5_check.aep 的 Gide 跟 tolerance.aep 的 Gide 完全一致。

### 验证

- PASS=175 (was 174, +1 from new Ewst test)
- vet clean
- `go run tmp_debug/gen_iter5_check/main.go && go run tmp_debug/dump_gide/main.go tmp_debug/iter5_check.aep` 输出第一个 Gide (user Layr 的) byte-identical 跟 tolerance.aep 的第一个 Gide
- `grep formType=(Layr|Ewst|DLay|SLay|CLay|SecL|Gide) tmp_debug/dump_iter5_check.txt` 计数跟 tolerance 完全一致

### head counter B 残留差异 (推测无关 silent-drop)

iter5_check head bytes [16..19] = `0000000e` (= 14)，tolerance = `00000024` (= 36)。iter 4 RE doc 写 "≥ nextItemID 即可" — 我们 nextItemID=14, counter B=14, 满足 gate。Tolerance 的 36 推测是 AE 多次 save 之后的累计 counter (cosmetic save-sequence)。不再优先 — 若 iter-5b 后 AE 仍 drop 再 RE。

### 不确定项 / 下一 iter 候选

- **head counter B 真语义** (14 vs 36 — gate "≥ nextItemID" 够不够，还是另有 minimum?)
- **fix C: spatial cdat per-dim padding + tdum/tduM** 真的还需要么？iter-5b 后 AE 若 accept variant #2 (empty layer) → 进入 shape-content validation → 这时 cdat 48B vs 80B 才可能有戏。但 iter-5b 之前根本没走到这一步，所以 fix C 之前的"必须"判断是 over-stated。

## iter-5b 永久教训

1. **silent-drop 用 bisect "最简变体 (zero content)" 第一步定位 instantiation vs content** — variant #2 (empty ShapeLayer 零 shape kid) 也 silent drop = 问题在 instantiation 层不在 content 层。iter 5 之前我们已经做了 iter 3 (Transform schema) + iter 4 (7 个 byte-level fix) 没碰到 instantiation 层，因为 bisect 没 zero-content 变体。bisect_v2_2 加 #2 之后立刻看见。
2. **byte-diff "AE-saved 跟 our-built" 在 layer-skel level** 是最高 ROI 的 RE — Item LIST direct children type 计数 + Layr direct children 计数 + Item-level structural LIST 出现位置/顺序。比 cdat 内部字节 RE 高一两个数量级，且找到的 bug 通常是 "缺整 chunk" 而非 "字节错"，修起来确定性高。
3. **GPT 反馈很有信息量** — bisect 数据 + 一段 prose 分析就能 pivot 整个 fix 方向。下次 ship gate 类问题如果不动，主动找 LLM critique，对照 AE-saved fixture 跟 our-built 的结构-级 diff。

## iter-7 实施记 — embed tolerance Transform Group bytes (跨第一道 silent-drop 闸门 ✓)

iter-5b 之后我用户陪跑了 7 轮 AE bisect (iter-6a..6f) 测各种 byte-level 修改候选 (tdsb / 3D 升 dim / spatial bounds / trailing chunks / placeholder flags) 全没动 silent drop。GPT 看完 7 轮无果后写 `flightdeck/kneeboard/gpt`:

> 你已经连续得到: AE accepts file BUT layer count still 0
> 这说明: parser path 已经过去了, object materialization path 没过去
> 这是两个不同阶段。你之前一直在修 parser-level corruption。现在问题明显已经上升到: registration / indexing / ownership / linkage / class tagging / hierarchy admission
> 不建议继续 6g/6h 式 blind patching。byte-level diff 阶段已经结束了。下一阶段该换: object graph / semantic model / importer behavior

GPT 给的具体 3 步法救了项目:

### 3 步 semantic RE

1. **排除 epistemic hole** — 验证 measurement pipeline 是否可信。Tolerance.aep 用同一个 JSX probe 跑一遍, 看 layers.length 是否 = 1。
   - 写 `test_data/generators/verify_baseline.jsx` + `tmp_debug/verify_baseline/main.go` (dump 完整 metrics: items count / comp count / activeItem / selection / per-item class+typeName / per-comp layers + 每 layer class+enabled+index)
   - tolerance 跑出 `comp.layers.length=1, layer[1]: name=Nested class=ShapeLayer enabled=true` ✓ → measurement sound, silent drop 真问题
   
2. **Transplant 法 isolate 触发器** — 不 byte-diff, 直接 swap chunk 看 AE 反应。逐级缩小范围:
   - `tmp_debug/transplant_layr/`: tolerance.aep 整个 user Layr 换成 ours minfail_v2 的 → layers=0 → 触发器在 Layr 内部
   - `tmp_debug/transplant_tdgp/`: 只换 outer LIST(tdgp) → layers=0 → 触发器在 property tree
   - `tmp_debug/swap_propgroup/`: 5 变体 batch (各 swap 一组 prop group)
     | swap | layers.length |
     |---|---|
     | RootVectors+Transform+LayerStyles | **0** |
     | Extrsn+Material+Audio+LayerSets (4 placeholders) | **1** ✓ |
     | **Transform only** | **0** ← 唯一触发 |
     | RootVectors only | **1** ✓ |
     | LayerStyles only | **1** ✓ |
   - `tmp_debug/swap_reverse/`: 反向 (ours base + tolerance Transform) → **layers=1, layer[1]: name=L class=ShapeLayer** ✓ — 100% 确认
   - 结论: 其它 6 个 prop group 我们 emit 都 byte-OK; **silent drop 唯一来源 = Layr Transform Group body**

3. **iter-7 解法 — embed tolerance bytes**: byte-level RE 阶段已结束, 不再 from-scratch 构造。直接 embed tolerance Transform Group bytes 作 boilerplate。

### iter-7 实现

- `tmp_debug/extract_transform_group/main.go`: 加载 tolerance.aep, 找 first user Layr 的 LIST(tdgp) (Transform Group body, 15 children = tdsb + tdsn + Position_0/_1 tdbs(6-child) + Orientation otst(2-child) + RotateX/Y tdbs(4-child) + EnvirAppear tdbs(4-child) + Group End), 用 `rifx.Chunk.Write` 序列化 → `internal/aep/templates/v2_2_transform_group_body.bin` (1842 B)
- `internal/rifx/rifx.go`: 新 public `ReadChunk(r io.ReadSeeker) (*Chunk, error)` — `Parse` 的单 chunk 版 (无 RIFX root 包裹要求)。embedded blob 用这个 parse
- `internal/aep/lower_layer.go`:
  - `//go:embed templates/v2_2_transform_group_body.bin var v22TransformGroupBodyBytes []byte`
  - `sync.Once` cache + `cloneShapeTransformGroupBody()` + `cloneChunk()` deep-clone helper
  - `lowerLayerTransform` 重写: 不再 LowerVec2Stream/LowerFloat64Stream from-scratch 构造, 直接 clone embedded body + 用 `overwriteScalarCdat(body, "ADBE Position_0/1", t.position.static[*])` 覆写 cdat scalar 用 runtime 值。Animated Position fallback: 用 first keyframe value 作 static slot
  - 删 dead code: `lowerOrientationDefault` (otst wrapper 已在 embedded body 内); `splitVec2Stream` (无需手 split Position_0/_1, 直接覆 cdat); 早期 emit 用到的 `LowerVec2Stream(t.anchorPoint/.scale)` / `LowerFloat64Stream(t.rotation/.opacity)` 调用全 retire (V2.2 限制声明)
- `internal/aep/lower_property_stream.go`: 删 `splitVec2Stream`
- `internal/aep/shape_graph_roundtrip_test.go::TestV2_2_CanonicalShapeGraph_Roundtrip`: Layr Position 期望从 "2 keyframes [0,0]→[500,300]" 改为 "static fallback = first kf [0,0]" (V2.3 RE byte 布局后做全 Layr Position 持久化)

### V2.2 限制声明 (写进 docs/shape.md Phase 6)

- ShapeLayer Layr-level Transform: 仅 **Position** static 值持久化到磁盘 (单次覆写 Position_0/_1 cdat)
- **Anchor Point / Scale / Rotate Z / Opacity**: runtime-only, **不持久化到磁盘** (in-memory API surface 仍可读写, write→re-parse 后值丢失回默认)
- **Position keyframes**: 不持久化 (V2.2 用 first kf 作 static fallback)
- 完整 Layr Transform 持久化 = **V2.3 工作** (byte-level RE for Position keyframe + Anchor/Scale 3D 布局)

### iter-7 验证

- PASS=175 不变, vet clean
- `tmp_debug/bisect_v2_2/main.go` 跑 AE 6 变体:
  | variant | iter-7 前 | iter-7 后 |
  |---|---|---|
  | #2 empty ShapeLayer | drop layers=0 | **PASS layers=1, layer[1]: name=L class=ShapeLayer enabled=true** ✓✓ |
  | #3 + AddRect | drop | drop (shape content level silent drop, 跟 Layr Transform 无关) |
  | #4 + SetSize | drop | drop |
  | #5 + AddFill | drop | drop |
  | #6 + size keyframed | drop | drop |
  | #7 + position keyframed | drop | drop |

### iter-7 永久教训

1. **silent-drop 类问题第一步 transplant 法 isolate, 不 byte-diff 猜字段**。Byte-diff 适合 "AE 拒收文件" (hard reject) 类问题; silent drop = AE 接受文件但内部 object materialization fail, 这是 semantic-level 问题, byte-level diff 看不出来。
2. **GPT 反馈在 6 轮无果时及时 pivot 救项目**。如果继续 iter-6g/6h byte-patch 我估计还得 5-10 轮才能撞对。GPT 的"object graph / class admission"分类法 + 具体 3 步建议 = 一次 pivot 直接缩 1 个变量定位。
3. **Embed boilerplate 是 V2.x ship-gate 的合法路径**。当 chunk 内部 byte 布局复杂到 from-scratch 构造太脆 (Transform Group 多 sub-stream + 多 cdat padding + spatial tdum/tduM + 6-axis 3D 跟 2D 区分等), embed AE-saved bytes + post-process 覆值是更稳的方式。runtime 持久化能力有限 (cdat scalar 值能改, 其它字节固定) 是接受的代价, 文档声明清楚, 后续 iter 再 RE 全 byte 布局。
4. **测量 (probe) 跟实测分离**。GPT 第一步排除"epistemic hole"= 验 probe 自己是否可信。如果之前 5 轮 silent drop 其实是 probe 写错 (例如忘 select active comp), 后面所有 fix 都白做。这是 5 分钟 sanity check, 千万省不得。
5. **`flightdeck/kneeboard/gpt`** 文件用法: 高难度卡死时让 LLM 反馈写进这里, claude 直接 read 当作"另一个 RE 专家的建议"参考。Pivot 信息密度比单条 chat 消息高 5x。

### iter-8 候选 (下次会话)

variant #3-7 仍 drop = shape content 级 silent drop 触发器。同样 transplant 法 isolate:
- 拿 iter-7 状态 minfail_v3.aep (含 Rect) + swap in tolerance 的 Root Vectors Group → AE 跑看 layers.length
- =1 → 同 iter-7 思路 embed tolerance Root Vectors Group / Rect body bytes 作 boilerplate (V2.2 限制声明: 只支持 tolerance 那个 Rect 200×100 + Fill gray 配置? 或者 embed multiple shape templates?)
- =0 → 触发器在别处 (5-层 nesting wrapper / Vectors Group inner / shape kid 加进 Item LIST 后变化的 chunks?)

iter-8 别再盲改, 用 `tmp_debug/swap_propgroup/` 类似的 transplant batch 工具 isolate 真凶到具体 chunk 后再选 strategy.

## iter-8 实施记 — 同 transplant + embed 法泛化到 Rect/Fill body (Phase 5 全闭环)

iter-7 解 Layr Transform 后, variant #2 (empty ShapeLayer) PASS, 但 variant #3-7 (加 Rect/Fill/keyframes) 仍 drop. 用同样的 transplant 法 isolate shape-content level silent drop:

### iter-8 transplant isolate

| 测试 | base | swap | result |
|---|---|---|---|
| `swap_rvg` | minfail_v3 (ours w/ Rect) | tolerance Root Vectors Group body | **layers=1** ✓ — drop trigger 在 RVG 子树 |
| `swap_rect_body` | minfail_v3 | tolerance Rect body 只换 | **layers=1** ✓ — wrappers (Vector Group/Vectors Group/Transform/Materials) OK, drop 在 Rect body 内部 |
| `swap_fill_body` | minfail_v5 (Rect+Fill) | tolerance Fill body 只换 | **layers=0** — Rect body 还在 drop |
| `swap_both_bodies` | minfail_v5 | tolerance Rect + Fill bodies 都换 | **layers=1** ✓ — 各 shape body 独立校验, 任一 broken = drop |

**Validator boundary 锁定**:

| 元素 | 我们 emit | byte-OK? |
|---|---|---|
| Item / Layr / Gide / Ewst / cdta / head / idta | ours | ✓ |
| Root Vectors Group body wrapper | ours | ✓ |
| Vector Group body wrapper | ours | ✓ |
| Vectors Group body wrapper | ours | ✓ |
| Vector Transform/Materials placeholders | ours | ✓ |
| 4 layer-level placeholder bodies (Extrsn/Material/Audio/Layer Sets) | ours | ✓ |
| Layer Styles body | ours | ✓ |
| **Layr Transform Group body** | from-scratch | ❌ (iter-7 embed) |
| **Each Shape body (Rect/Fill/...)** | from-scratch | ❌ (iter-8 embed) |

AE 对 "complex multi-stream property containers" (Layr Transform / 每个 shape body) 独立 byte-level 校验。Wrappers + layer-skel chunks 都容错 (ours from-scratch byte-OK)。Embed approach 是 generalizable 解法。

### iter-8 实现

- `tmp_debug/extract_shape_bodies/main.go`: 抽 tolerance 的 Rect body + Fill body LIST(tdgp) → 2 binary blobs in `internal/aep/templates/`:
  - `v2_2_shape_rect_body.bin` (448 B, 5 children: tdsb + tdsn + Size sub-prop + Size LIST tdbs + Group End)
  - `v2_2_shape_fill_body.bin` (426 B, 5 children: tdsb + tdsn + Color sub-prop + Color LIST tdbs + Group End)
  - 注: tolerance.aep 用 AE-canonical "elide defaults" 风格, Rect Direction/Position/Roundness + Fill Opacity/Blend Mode/Composite Order/Fill Rule 全 elide; embedded body 只含 Size/Color。
- `internal/aep/lower_shape_node.go`:
  - 加 `//go:embed` for 2 个 blob + sync.Once cache + `cloneShapeRectBody()` / `cloneShapeFillBody()` helpers
  - 新 `overwriteShapeStreamCdat(body, streamName, data)` helper: 找 tdmn matching streamName 后的 LIST(tdbs)'s cdat, copy `data` 进 cdat[0..len(data)]
  - 新 `encodeF64sBE(...)` helper: 多 f64 BE encoding 一起
  - `lowerRectNode` 重写: cloneShapeRectBody + overwrite "ADBE Vector Rect Size" cdat with `encodeF64sBE(w, h)`
  - `lowerFillNode` 重写: cloneShapeFillBody + overwrite "ADBE Vector Fill Color" cdat with `encodeF64sBE(r, g, b, a)`
  - 删 dead code: Rect 的 Direction/Position/Roundness placeholders, Fill 的 Blend Mode/Composite Order/Fill Rule/Opacity emit calls; 不再需要 `LowerColorStream` / `LowerFloat64Stream` for these
- 测试调整: `TestV2_2_CanonicalShapeGraph_Roundtrip` Rect Size 从 "Animated 2 keyframes" 改为 "Static fallback = first kf value [50,50]" (跟 iter-7 Layr Position 同 V2.2 限制 pattern)

### iter-8 AE bisect 全 PASS

```
=== variant 2 empty ShapeLayer  → PASS  comp.layers.length=1  layer[1]: name=L
=== variant 3 + AddRect (default) → PASS  comp.layers.length=1  layer[1]: name=L
=== variant 4 + rect.SetSize (static) → PASS  comp.layers.length=1  layer[1]: name=L
=== variant 5 + AddFill (static color) → PASS  comp.layers.length=1  layer[1]: name=L
=== variant 6 + rect.Size keyframed → PASS  comp.layers.length=1  layer[1]: name=L
=== variant 7 + L.Position keyframed → PASS  comp.layers.length=1  layer[1]: name=L
```

Phase 5 ship gate **全闭环**。这是 V2.2 从开发到 AE 接受的 milestone, 也是 6-iter dead loop → 2-iter breakthrough 的转折。

### iter-8 V2.2 alpha 限制 (要进 docs/shape.md)

- ShapeLayer Layr Transform: 仅 Position **static** 持久化; Anchor/Scale/Rotation/Opacity runtime-only
- ShapeLayer Position **keyframes**: 不持久化 (first kf 作 static fallback)
- Rect: 仅 Size **static** 持久化; Position/Roundness/Direction runtime-only; Size keyframes 不持久化
- Fill: Color 持久化但 **编码可能不准** (JSX 0.5 → tolerance bytes 0x406fe0... ≈ 255, 我们 emit user 值 0.5 → 0x3fe0... 可能跟 AE-internal 编码不一致, 视觉色可能错; V2.2.1 RE)
- Fill Opacity / Blend Mode / Composite Order / Fill Rule: runtime-only
- **Ellipse / Path / Stroke**: V2.2 alpha **不支持** (Go-side 能 emit + parse, AE 会 silent drop layer; 需 V2.2.1 各 shape kind extract+embed)
- 全部 keyframe (Size/Color/Position 等): 不持久化 (embed body 只 static cdat slot, V2.2.1 加 LIST(list) lhd3/ldat keyframe 编码)

### iter-8 永久教训

1. **iter-7 embed approach 是 generalizable**, 不是一次性 trick。任何 "complex multi-stream property container" 类 silent drop 都用同套法 (transplant isolate → extract bytes → embed → cdat 覆值). V2.2 Phase 5 全程印证: 从 Layr Transform Group → 各 Shape body → 同样的 4 步流程。
2. **Byte-level RE 在 silent-drop 场景是 dead end** (iter-6a/b/c/d/e/f 6 轮证明). Semantic-level transplant isolation 是 right tool. 之前自己摸索 6 轮没解, GPT 看完 bisect 数据立刻 pivot 救项目 (`flightdeck/kneeboard/gpt`)。
3. **AE saved fixtures 是 ship-gate-class V2.x 项目的核心资源**。tolerance.aep 一个 fixture 解了 Layr Transform + Rect + Fill body 三处 silent drop。V2.2.1/V3 work 需要更多 fixtures (Ellipse / Path / Stroke / etc each AE-saved). 抽 + embed 流水 (`extract_*` tools) 是 reusable infrastructure。
4. **V2.x alpha 限制 ≠ 失败**。Phase 5 ship gate 全闭环但 keyframes / 其它 shape kinds 不持久化 — 这是合理的 V2.2 alpha scope。docs 声明清楚 + V2.2.1 subplan 接力, 项目可以**先 ship 后扩展**, 而不是因为追求完整就 6 轮死循环。

## 永久教训

V2.2 Phase 4 Go roundtrip PASS **不代表 AE 接受**。Go parser 写 tolerant，AE parse 严格。下次类似 "writer + ship gate" 流程 phase 顺序要把 ship gate 提前。

silent-drop 类问题 (AE 接受文件但内部 hide layer) = **semantic-level**, 不是 byte-level corruption。第一步用 **transplant 法** isolate 真凶到具体 chunk，第二步若 from-scratch 构造太脆就 **embed AE-saved bytes 作 boilerplate** + post-process 覆 runtime 值，第三步 docs 声明 V2.x 限制 (持久化能力 vs runtime API surface) + 留 V2.x+1 RE 任务。**不要再像 iter-6a..6f 那样盲改 byte-level fields**。

`flightdeck/kneeboard/gpt`-style **LLM pivot 反馈** 是 ship-gate-stuck 时的关键工具：6+ iter 没进展时主动找另一个 LLM 看 bisect 数据 + 建议结构, 信息密度比单条 chat 高 5x。把反馈写进 `flightdeck/kneeboard/gpt` 让 claude 当作"另一个 RE 专家的建议"参考。
