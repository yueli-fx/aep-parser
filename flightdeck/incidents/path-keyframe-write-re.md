---
status: active
last_updated: 2026-06-12
when_to_read: implementing shape-path keyframe write (animated ADBE Vector Shape); wiring LowerPathStream into lowerPathNode; emitting multi-frame om-s (tdbs time table + N shap); debugging "AE drops animated path" / wrong vertices on a keyframed shape path; deciding from-scratch vs embed-template for animated path lower
applies_to: [path-keyframe, shape-path, om-s, omks, shap, shph, tdbs, lhd3, ldat, bbox-normalization, LowerPathStream, lowerPathNode, encodeBezier, ae2020, ae2025, version-portable, linear-interp, phase-0-re]
---

# Shape-path keyframe write — Phase 0 RE

逐帧 bezier path 动画的磁盘结构。fixture `test_data/re_path_anim.jsx`（3 个 LINEAR 关键帧，各帧顶点数/bbox 不同）→ AE 2020 + AE 2025 双版本存盘 → `tmp_debug/dump_path_anim`。

## 结论速览

动画 shape path 与**动画 mask path 完全同构**，且 **AE 2020/2025 字节结构可移植**（仅运行时指针差异）。**Phase 2 不是从零造**：单帧 bbox 归一化编码器 `encodeBezier` 已实现并有测试（`lower_property_stream.go` + `..._test.go`），缺的只是多帧组装 + 接线。

## 磁盘结构（== 动画 mask）

```
om-s LIST
  tdbs LIST            ← 关键帧 TIME 表（每帧一条）
    list(kfl) LIST
      lhd3 (52B)       count@0x08 = #kf, @0x0C = 1, bpk@0x10 = 64
      ldat             #kf × 64B block
  omks LIST            ← 每个关键帧一份几何
    shap LIST × N      （N == #kf）
      shph (24B)       本帧独立 bbox
      list(kfl) LIST
        lhd3 (52B)     count@0x08 = nVerts×3, @0x0C = nVerts, bpk@0x10 = 8
        ldat           nVerts × 3 × (f32 X, f32 Y)  ← anchor / in-tan / out-tan
      omtn             （空 name）
```

`parse_mask.go` 的 `decodeMask` + `readMaskPathTimes` + `decodeMaskVertices` 已能读这套结构 → **Phase 1 hydrate 直接借用**（shape path 现走 `parse_shape.go::decodeShapePath`，每个 shap 单独产出一个 `ShapePath`，**不建模多帧**，需改成 mask 那样的 N-shap→keyframes 聚合 + 读 tdbs times）。

## (b) tdbs 时间块 = 标量 spatial-ease 64B 布局

每块 64B（与 `readMaskPathTimes` 注释一致，也 == spatial scalar ease 块）：

| off | 字段 | linear 实测 |
|---|---|---|
| 0x00 | time u32 BE（ticks，**per-comp tickRate**） | 0 / 30720 / 61440（comp 30fps → tickRate=30720=30×1024，1s=30720） |
| 0x04 | inInterp u8 | 1 = LINEAR（注意 ScriptingAPI 枚举 6612，磁盘上 LINEAR=1） |
| 0x05 | outInterp u8 | 1 |
| 0x07 | u8 const | **0x01（所有帧）**。lower：`blk[0x07]=0x01` |
| 0x08 | u32 BE const | **0x00000002（所有帧）**。lower：`PutUint32(blk[0x08:0x0C],2)` |
| **0x10** | f64 | **1.0（非末帧）/ 0.0（末帧）**；linear 下 AE 仍写 1.0。lower：`i != n-1 → PutUint64(blk[0x10:0x18],1.0)` |
| 0x18 / 0x20 / 0x28 / 0x30 | in/out speed·influence f64 | linear 全 0 |
| 0x38 | 8B trailer | **运行时指针/缓存，每次运行+每版本都变 → 置零**（2020 `60ef0ab4…` vs 2025 `2078fc85c2010000`，形似堆地址；末帧本就 0）。lower 一律置零 |

> ⚠️ **本表 2026-05-31 二次纠错（双版本 ship-gate 通过后回填）**：先前一版把 1.0 写成 @0x30、@0x07 写成「首帧 2 余 0」，**两处都错**（误读通道导致；详 [[feedback_bisection_over_stacking]]）。实测 byte-diff `re_path_anim.aep`：1.0 在 **@0x10**、@0x07=**0x01 全帧**、另有 **@0x08 u32=2 全帧**。最初版的 @0x10 反而是对的。

## (c) shph (24B) — 每帧独立 bbox + 顶点归一化（**关键**）

```
0x00 magic u32 = b3de0201（常量）
0x04 minX f32 ; 0x08 minY f32 ; 0x0C maxX f32 ; 0x10 maxY f32
0x14 byte = 01（本 fixture 全闭合路径；mask 读法把 0x14 当 closed 标志，
            而 lower_property_stream 写法把 0x14..0x17 当常量 0x01000000——
            open path 能否区分 closed=0 未测，本 fixture 无开放路径）
```

实测 bbox 逐帧独立：square 0..40 → `[0,0,40,40]`；square 10..150 → `[10,10,150,150]`；triangle bbox 0..200 → `[0,0,200,200]`。

⚠️ **顶点存的是 bbox 归一化 0..1，不是绝对层坐标**。输入 40px 方块 → ldat 存单位方块 `(0,0)(1,0)(1,1)(0,1)` + shph bbox `[0,40]`。两个不同尺寸的方块归一化后 ldat **完全相同**，差异只在 shph bbox。Phase 2 lower 必须算 bbox 再归一化（`encodeBezier` 已这么做：bbox = min/max over verts ∪ verts+inTan ∪ verts+outTan）；Phase 1 hydrate 要 de-normalize 还原绝对坐标。**现有 `decodeMaskVertices` 读原始 f32 = 归一化坐标，不乘 bbox** —— Phase 1 必须确认静态 shape-path 读路径是否也受此影响（疑似 latent gap）。

## (d) lhd3 (52B) 随帧/顶点变的字段

`magic@0x00 = 0x00d00bee`（常量）、`@0x04 = 0`。变的：

| | @0x08 count | @0x0C | @0x10 bpk |
|---|---|---|---|
| tdbs 时间表 | #kf (=3) | 1 | 64 |
| shap 几何 | nVerts×3 (=12 / 9) | nVerts (=4 / 3) | 8 |

## 版本可移植性（== gradient 结论）

AE 2020 vs AE 2025 输出**逐字段同构**，唯一差异：64B 时间块 `@0x38` 运行时指针 trailer + `@0x08` flag(2 vs 0)。一份 codec/模板服务双版本，**dual-version ship-gate 廉价**。fixture：`re_path_anim_ae2020.aep`（canonical = `re_path_anim.aep`）+ `re_path_anim_ae2025.aep`。

## Phase 2 决策：扩 + 接线，非从零

`LowerPathStream`（`lower_property_stream.go`）**已有但无调用方**；`StreamModeAnimated` 当前只发 `keyframes[0]` 单 shap、**无 tdbs**（行 102-116）。`encodeBezier`（单帧 shph/lhd3/ldat，bbox 归一化）已 RE-verified-correct（dump 与之逐字段吻合）。所以：

1. 扩 `LowerPathStream` `StreamModeAnimated`：循环 keyframes 发 N 个 shap（每帧 `encodeBezier`）+ 造 tdbs 时间表（复用标量 64B time-block 编码器，ease 置 0，trailer 置 0）。
2. 接线 `LowerPathStream` → `lowerPathNode`（`lower_shape_node.go:454-456` 现走 embed 模板单帧 fallback）；`numKeys≤1` 保留 embed 静态路径。
3. 既往"from-scratch shape path 崩 AE 2020 (0::42)"风险已缓解：AE 双版本**自己**就产出这套字节，byte-match 即 AE 接受。

deferred：temporal ease（首版 linear only）、mask path write（只做 shape path）、open path 的 shph 0x14 语义、time-block `@0x30` 1.0 的作用。

## Phase 2 实现结果（2026-05-31）

**未走 from-scratch `LowerPathStream`，改为扩 embed 模板**（`lower_shape_node.go::lowerPathNode` + 新 `spliceAnimatedPath` / `spliceShapGeometry` / `encodePathTimeTable`）。理由：static path from-scratch 曾崩 AE2020(0::42)，故 static 走 embed 模板 splice；动画路径与 static 共用 om-s/omks/per-shap 脚手架，唯一新增 = ①value tdbs(tdsb+tdsn+**tdb4+cdat**) 转 time-table tdbs(tdsb+tdsn+**tdb4(patch flag)**+**LIST(kfl){lhd3,ldat}**) + ②omks 单 shap 克隆成 N shap 各 splice 几何。全部复用 AE 原生脚手架，比从零造低风险。判据：`numKeys≥2` 动画，`≤1` 走原 static splice。

**关键结构发现（dump `re_path_anim.aep` 实测）—— 2026-05-31 二次纠错**：动画 path 的 time-table tdbs **保留 tdb4**（124B；= tdsb + tdsn + tdb4 + LIST(kfl)），只是把 cdat 换成 time-table，并对 tdb4 打三个 static→animated flag patch（@0x05 `&^=0x01`、@0x44`=0x01`、@0x4f `&^=0x01`）—— **与标量流 `injectAnimatedStream` 完全同构**。先前一版「time-table tdbs 没有 tdb4、转换=删 tdb4」是误读，已修正（byte-diff 静态 embed tdb4 vs 真 fixture tdb4 = 仅这三个 offset 差）。lhd3 时间表头常量与 `encodeKeyframes` 完全一致（magic 00d00bee / @0x08=#kf / @0x0C=1 / @0x10=64=bpk / @0x14=4 / @0x18=1 / @0x1C=4）。

**验证状态：AE 2020 + AE 2025 双版本 ship-gate 均 PASS ✅（2026-05-31）**。`shape_pathkf_shipgate_test.go::TestV2_2_PathKf_AEShipGate_AE2020/_AE2025`——AE 打开动画 path 文件、JSX 确认 path numKeys≥2 未被 drop、re-save，Go 端 re-decode resaved om-s 确认 ≥3 shap + time-table kfl 存活，全链路过。Go 多帧 round-trip ✅ + hydrate ✅ + `go vet`/`go test ./...` 全绿。

**踩坑链（gate 走通前的三关，皆非数据问题之外的连环误读）**：
1. **首跑 exit 2 被 cockpit「默认当 flake」误导**——实为 AE 弹「项目文件似乎已损坏（跳过部分：1）(26::0)」真损坏框（[[ae2020-shape-ldta-164-corrupt.md]] 同签名），切屏只是叠加了 OCR 遮挡。根因 = 上面的 time-block + tdb4 字节错。修字节后损坏框消失。
2. **JSX 导航太浅**：`contents.property(1).property("ADBE Vector Shape")` 找不到——path 流嵌在 `Root Vectors Group → Vector Group → Vectors Group → Vector Shape - Group → Vector Shape` 共 3 层下，改成递归 `findByMatch`。
3. **测试 helper 用错匹配**：`findShipChunk(root, IDOmS)` 按 `ch.ID` 找，但 om-s 是 `LIST`（FormType="om-s"）→ 永远 nil。加 `findShipListByForm` 按 FormType 找。

dump/build 工具（tmp_debug，gitignored）：`dump_path_anim`（按块解码）、`build_pathkf`（复刻 gate build 落盘）、`dump_tdbs`（tdbs leaf）、`dump_tdmn`（match-name 树）。

## encodeBezier lhd3/shph 偏离 AE-native — AE 容忍但非 byte-faithful（2026-06-12，二次纠错）

[[add-mask-create-re]] 掀出 encodeBezier 的 geometry lhd3 写 `@0x14=n / @0x18=closedFlag / @0x1C=16`，
留 follow-up 问 shape 侧是否 n≠4 broken。**首轮结论（「shape 不受影响 = mask 独有严格性」）经 dump
AE-native fixture `v2_2_shape_path_re.aep`（含开放 path）后部分推翻**：AE-native shape path 的这些字段值
与 **mask 完全一致**（shph[3] open=**0x09**、@0x14=**4 恒**、@0x18=**1 恒**、@0x08=3n、@0x0C/@0x1C 疑
cap=nextPow2(n)/4·cap），而 encodeBezier 写 `shph[3] open=0x00 / @0x14=n / @0x18=closedFlag / @0x1C=16`，
**全部偏离、只 n=4 闭合巧合**——这是与 mask 同源的通用偏差，**不是 mask 独有**。

**对的部分留下**：shape path 的 **AE 接受性** 确实不受影响——n=3 closed gate
（`shape_path_shipgate_test.go` + 精确读回）+ 本 gate frame 2（n=3）双版本 PASS，AE 对
"ADBE Vector Shape" om-s **容忍**偏差值（不像 mask 急切解码 outline → 硬崩 0::42）。

**已修的确证 bug**：`decodeShapePath`（parse_shape.go）用恒 0x01 的 shph[0x14] 读 closed → 公共 API 对
开放 shape path 误报 Closed=true，已改 shph[3]（与本文件 `bezierFromShap` 一致）。**开放 shape path 写
功能性 = 双版本 ship-gated**（`TestV2_2_PathOpen_AEShipGate_AE2020/_AE2025` PASS——AE 接受偏差字节、顶点
+ 开放标志保真）。**writer byte-faithfulness 降为纯优化 follow-up**（AE 容忍现状，需求驱动）。
详 [[add-mask-create-re]] §三 finding 4。
