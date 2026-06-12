---
status: active
when_to_read: debugging wrong RENDERED colours/styles on a Go-built layer while every stored value and JSX DOM readback looks correct; touching appendLayerStylesPlaceholder / makeTdsb flag words; adding any layer-level property-group placeholder; building a render-pixel gate
applies_to: [layer-styles, tdsb, enabled-bit, render-pixel, red-line-4, new-shape-layer, solidFill, bevelEmboss, ship-gate, ae2020, ae2025, saveFrameToPng, 32bpc]
last_updated: 2026-06-12
resolved_by:
---

# Layer Styles tdsb bit0 = render-time enabled bit

## Signature
- symptom: Go-built shape layers render collapsed into dark red (red solid-fill overlay + bevel/emboss relief on every shape) while every stored value、Go round-trip、AE resave、JSX DOM readback (fill/stroke color、opacity、blend mode) 全部正确；32bpc 下额外整体偏暗
- error_type: —
- where: internal/serializer/lower_layer.go appendLayerStylesPlaceholder / makeTdsb (lower_property_stream.go)
- trigger: 渲染任何 NewShapeLayer 产出的层（AE 渲染像素才暴露；打开工程检查值看不出）

## 症状/复现

orbit demo（红线4 首案）：从零 Go 拼 5 层 shape 工程，所有颜色 cdat 值对、双版本 ship-gate（值级）全绿，但 AE 渲染帧整体塌成暗红——BG indigo 渐变渲成 flat 50% 红、所有 dot/ring 同色调浮雕红。链路上每一级 probe 都"正确"：

1. `dump_named_cdat`：fill/stroke cdat = ARGB×255 正确；
2. JSX DOM readback：`ADBE Vector Fill Color` = 设定值精确、opacity 100、blendingMode normal；
3. AE **resave 后渲染逐字节不变**（排除 shape 数据，锁定项目/层级状态）；
4. pcms 色彩管理 JSON（`{}` vs ACES 1.2）字节补丁后渲染不变（证伪色彩管理假说）；
5. **看渲染 PNG 图像**才破案：dot 是带高光的红浮雕球 → 这是 Layer Styles（solidFill 默认红 + bevelEmboss + shadows/glows），JSX probe 确认 10 种 style 全 `enabled=true`。

## 根因

`tdsb` flag word 的 **bit0 = 该组渲染期 enabled 位**。AE 原生 shape layer 的 Layer Styles 家族 tdsb（AE 2020/2025 字节一致）：

| 组 | 原生 tdsb |
|---|---|
| Layer Styles body / Blend Options Group | `0x00000003` |
| ADBE Adv Blend Group | `0x00000001` |
| 10 × fx/enabled（dropShadow…frameFX） | `0x00000002`（present、**OFF**） |
| Extrsn/Material/Audio/Layer Sets placeholder | `0x00000003` |

我们的 `appendLayerStylesPlaceholder` 全部套用泛用叶值 `makeTdsb()=0x00000001`——bit0=1 把 10 种 style 全部置为启用。AE 的 DOM 值读回看不出问题（style 参数组是空 placeholder，读回的是各 style 的默认参数——solidFill 默认红色），只有渲染像素暴露。**这正是交付准则红线4 的实证：值 round-trip 绿 ≠ 渲染正确。**

次要发现（同次 RE）：
- 种子模板项目是 **32bpc**（AE 默认 8）；32bpc 下 `saveFrameToPng` 输出 **linear 编码**（整体偏暗，(v/255)^2.2 关系），渲帧前先 `app.project.bitsPerChannel = 8`。
- `saveFrameToPng` 异步写盘，会被 `app.quit()` 抢跑致空文件——save 后 `$.sleep(2000)` 再 quit。
- `pcms`（AE 2024+ 色彩管理 JSON，AE 2025 原生 `{"ocioConfigurationFile":"ACES 1.2"}`）：旧工程缺失时 AE 2025 resave 补 `{}`，对渲染无影响（本案证伪）。

## 修法

`makeTdsbFlags(flags uint32)`（lower_property_stream.go）+ `emptyPropGroupFlags(flags)`（lower_layer.go），Layer Styles 家族按上表写原生值；其余 emptyPropGroup 调用点（Vector Transform/Materials placeholder 等）保持 0x01 不动。修复 commit 同步把 orbit ship-gate 升级为**渲染像素 gate**：JSX 渲帧 0 → Go `image/png` 采样 5 个点（BG/Ring/3 dots）断言 RGB ±8——双版本 PASS。

渲染验证基建（可复用）：JSX `comp.saveFrameToPng(0, png)`（先 8bpc、后 $.sleep）+ Go 采样断言（`orbit_demo_shipgate_test.go` checkOrbitRenderedPixels / `tmp_debug/pixel_read`）。诊断工具：`tmp_debug/dump_styles_tdsb`（打 Layer Styles 家族 tdsb）、`tmp_debug/patch_pcms`、`tmp_debug/anim_demo/probe_colors.jsx`（DOM 色彩读回）、`probe_styles.jsx`（style enabled 读回）。

## Cases
- 2026-06-12 首次（orbit demo 红线4 首案；前一日 ship 的 orbit gate 只验了值级+不卡+resave，未验像素）
