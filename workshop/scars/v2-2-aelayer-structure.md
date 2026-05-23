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

`test_data/verify_v2_2.jsx` (Phase 5 Task 5.1) 只验 PASS/FAIL 不验 silent drop。`test_data/verify_open.jsx` (iter 4 新写) 显式 dump `items[i].typeName / layers.length / layers[k].name` 才暴露 silent drop。

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

## 永久教训 → CLAUDE.md / scars

V2.2 Phase 4 Go roundtrip PASS **不代表 AE 接受**。Go parser 写 tolerant，AE parse 严格。下次类似 "writer + ship gate" 流程 phase 顺序要把 ship gate 提前。
