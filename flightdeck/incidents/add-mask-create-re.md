---
status: active
when_to_read: implementing or extending AddMask / mask-atom creation; debugging AE crash (0 :: 42) or 参数值无效 on a Go-written mask; reasoning about mask shape coordinate units (fraction vs pixel) or the shph open/closed flag; touching encodeBezier lhd3 fields for non-4-vertex paths; needing the (tdmn, mkif, tdgp) atom triple layout
applies_to: [add-mask, mask-parade, mask-atom, mkif, om-s, shph, lhd3, tdb4, coordinate-space, open-path, closed-flag, structural-write, splice, parade-auto-create, ae2020, ae2025, ship-gate, crash-0-42]
last_updated: 2026-06-11
resolved_by:
---

# AddMask — Mask Parade from-scratch atom RE + ship findings

## Signature
- symptom: `After Effects 无法继续: 抱歉，After Effects 已崩溃。(0 :: 42)` on opening a Go-written mask file; or JSX `EXC Error: 参数值无效` reading a mask shape; or mask vertices read back ×layer-size / ÷layer-size off
- error_type: —
- where: internal/serializer/mutate_mask_add.go (makeMaskShapeOmS / maskLayerDims) · parse_mask.go Closed decode
- trigger: writing a from-scratch mask atom with shape-path conventions (canonical scalar tdb4 / closed flag at shph[3]-as-shape / lhd3 vertex-count at @0x14) or with pixel coords on a source-backed layer

`aep.AddMask(layer, name, path)` — fully from-scratch mask atom (NO embed
template; first structural write where every chunk is Go-built), spliced into
`ADBE Mask Parade`. Shipped 2026-06-11, AE 2020 + AE 2025 ship-gate 4/4
(AE-native fixture + 100% Go-built project, closed + open paths, exact vertex
readback, resave preservation). Ground-truth fixture: `test_data/re_mask_open.aep`
(+ `.jsx`), AE-2020-saved masks across solid/shape layers × open/closed.

## Atom 结构（与 effect 的差异）

一个 mask = parade tdgp 里的 **(tdmn "ADBE Mask Atom", mkif[48B], LIST:tdgp) 三件套**
——比 effect 的 `(tdmn, sspc)` pair 多一个 mkif。后果：

- `childTdmnPayload`（mutate_property_structural）的 pair 假设对 mask atom 不成立
  （tdgp 前是 mkif 不是 tdmn）→ **Remove/Move/Duplicate 对 mask 子项会拒绝**（安全失败，
  不损坏）。RemoveMask 需要 triple-aware 实现 → deferred。
- atom tdgp 内容：`tdsb 0x01 + tdsn(掩码显示名) + tdmn "ADBE Mask Shape" + LIST(om-s)
  + tdmn "ADBE Group End"`。Feather/Opacity/Expansion 默认省略（同 effect param elision）。
- **mask 显示名存在 atom tdgp 的 tdsn**（AE 面板名），omtn 恒空——parse 侧已加 tdsn
  fallback（`Mask.Name`），合成 fixture 才用 omtn。
- mkif 48B：`@0x04` mode、`@0x08` per-layer index（1 起，AE 容忍 gap）、`@0x0C` "om"
  tag、`@0x10`=0x0E、`@0x18`=0x0F、`@0x20` 存盘时间戳（任意值可）、`@0x24`=65C00000、
  `@0x2C` FF+RGB label 色（首 mask 黄 E4D84C）。
- Parade 位置 anchor：**Effect Parade 之前**（有则），否则 Transform Group 之前；
  auto-create 复用 `spliceEmptyParade`（由 ensureEffectParade 泛化而来）。

## 三个 gate 踩出来的字节语义（全部 ground-truth 校正）

1. **Mask Shape 的 tdb4 ≠ canonical 标量 tdb4**。`@0x04..0x0B = 00 07 00 01 00 02 00 07`
   （path-stream 标记）+ `@0x10` 双精度 0.0001 + 四个 1.0。发 canonical 标量头
   （`00 01` + headerByte）→ **AE 2020 开文件即崩 (0 :: 42)**。同样的 canonical 头在
   shape layer 的 "ADBE Vector Shape" om-s 里却被接受（已 gate）——严格性是 mask 特有
   （AE 打开工程即急切解码 mask 轮廓）。修法：整块 124B 常量 `maskShapeTdb4`
   （`@0x0C` 字段 78 00 / 5D A8 两种存盘变体都接受）。
2. **坐标空间分层**：mask 顶点存盘单位 = **源 item 像素空间的分数**（solid/footage/
   precomp：`px / sourceDim`）；**source-less 层（shape/text）= 裸像素**（除数 1）。
   写错的症状不是 reject 而是顶点读回 ×100 / ÷1920（gate 数值读回逮住的）。
   `maskLayerDims` 负责换算，公共 API 始终收像素。
3. **open/closed 标志在 shph[3] bit3**：closed=`02 01`，open=`02 09`；`shph[0x14]`
   恒 0x01（**不是** closed 位——旧 parse/SetClosed 读写 0x14 只是在 closed mask 上
   碰巧对，已改 bit3，合成 testutil 同步）。shape path 的约定（[3] 0x01/0x00）写给
   mask → AE 读 mask shape 抛 `参数值无效`。
4. **lhd3 字段对 n≠4 顶点是错的（encodeBezier 遗留）**：mask 实测 `@0x14`=**4 恒定**
   （encodeBezier 写顶点数 n，仅 n=4 巧合成立——n=3 的 mask 让 AE 2020 硬崩）、
   `@0x18`=**1 恒定**（不是 closed 位）、`@0x1C`=**4·n**（不是常量 16）。AddMask 在
   encodeBezier 后补丁这三个字段。✅ **follow-up 已核（2026-06-12，纯代码）——shape path
   侧不受影响，是 mask 特有严格性，非 encodeBezier 通用 bug**：前提（「shape gated fixtures
   恰全是 n=4」）不成立。静态 path ship-gate（`shape_path_shipgate_test.go`）用的就是
   **n=3 三角形** `{{10,20},{70,30},{40,90}}`（closed），走 `splicePathGeometry →
   spliceShapGeometry → encodeBezier`（lower_shape_node.go:506 整块替换 lhd3），产出
   `@0x14=3 / @0x18=1 / @0x1C=16`，**AE 2020 + AE 2025 双版本接受 + `assertResavedPathAnchors`
   精确读回 3 顶点（容差 0.5px）**；动画 path gate（`shape_pathkf_shipgate_test.go`）frame 2
   同为 n=3 三角形，也双版本 PASS。即 AE 对 shape layer 的 "ADBE Vector Shape" om-s 容忍
   encodeBezier 的 n-依赖 lhd3 字段（不像 mask 打开工程即急切解码 outline → n≠4 硬崩 0::42）。
   **唯一未 gate 的次要轴**：所有 gated shape path 皆 `Closed=true`（→ `@0x18=closedFlag=1`，
   恰合 mask 的恒定 1），**开放 shape path** 下 encodeBezier 写 `@0x18=0` 无 ship-gate 覆盖——
   但这是 closed/open 轴而非 n≠4 轴，且 shape 开闭独立编码在 shph[3]，需求驱动再核。
   （见 [[path-keyframe-write-re]]）

ldat 三元组布局与 shape path 完全同构（`[anchor, 本点出控制点, 下点入控制点]`，绝对值、
bbox 内归一化）——encodeBezier 直接复用；parse 侧 `MaskVertex.InTangent/OutTangent`
字段名是误称（实为 out-ctrl / next-in-ctrl），读 API 不动，文档已注。

## 自动化体系收获（同次会话落地）

- **Crash action（exit 8）**：AE 崩溃对话框规则 `ae-crashed`（OCR 含 ASCII 锚点
  `learn_ae`——CJK 常被 OCR 打碎）；按确定收尸 + 立即 exit 8；gate harness 对
  1/2/**8** 各 warm retry 一次。teardown 强杀后 `WaitForExit` 防重试撞 exit 6。
- **PrintWindow 采集**：`Capture-WindowBitmap` 升级 `PrintWindow(PW_RENDERFULLCONTENT)`
  （遮挡免疫，交互会话里编辑器压住 AE 对话框不再毒化 OCR），失败回退屏幕矩形。
  详 [[ae-automation-occlusion-crashstate]]。
- **取证保留**：gate 失败时 `.fail/` 从 t.TempDir 搬到 `tmp_debug/gate_fails/`
  （TempDir teardown 会把取证包一起清掉，2026-06-11 排障时踩过）。

## Coverage / gate

- `TestAddMask_AEShipGate_AE20{20,25}` — AE-native baseline 层（3 effects）+ AddMask：
  接受、name/mode/closed/顶点像素级读回、effects 完好、resave 保留。
- `TestAddMaskAutoGo_AEShipGate_AE20{20,25}` — 100% Go-built（NewProject→NewShapeLayer→
  Reopen→AddMask×2，closed 矩形 + open 折线）：双 mask、open closed=false、顶点精确。
- Go-only：`add_mask_test.go` 6 用例（round-trip / 二 mask 索引 / open / fresh 拒绝 +
  Reopen / camera-light 拒绝 / 空 path 拒绝）。

## Deferred

- **RemoveMask**（triple-aware 删除）/ mask path 改写（既有 mask 的 SetMaskPath）/
  animated mask path（om-s 多 shap + tdbs 时间表，机制同 [[path-keyframe-write-re]]）。
- mask mode/color/feather 创建参数化（今天 AE 默认 + 返回 *Mask 后 Set* 可改 mode/
  inverted/color）。
- precomp 层 mask 未单独 gate（按 footage 分数处理，理论一致）。

## Cases
- 2026-06-11 首次（AddMask v1 实现 + 双版本 gate；崩溃三连环：canonical tdb4 → 像素坐标 → shape 约定 open 标志）
