---
status: active
when_to_read: 想给文字样式 setter（SetRunFontSize/SetRunLeading/SetRunTracking/SetParagraphJustification/SetRunCapsOption）补 render-pixel gate；从零 NewTextLayer 的文字渲染不出来 / 渲染在画面外；SetRunFontSize 写完 AE 读回字号 ÷65536；纠结文字样式族为何只到 verify=roundtrip；判断要不要投入「materialize 从零文字 transform/matrix」
applies_to: [text, text-style, SetRunFontSize, SetRunLeading, SetRunTracking, SetParagraphJustification, SetRunCapsOption, render-gate, red-line-4, from-scratch-text, NewTextLayer, text-matrix, transform-omission, position-not-materialized, fontsize-65536, btdk, run-style, roundtrip-not-render, negative-finding]
last_updated: 2026-06-16
resolved_by:
---

# 文字样式 render-gate：从零文字的 transform/matrix 缺陷挡路（红线4 揪出真问题）

## TL;DR

文字样式族（`SetRunFontSize` / `SetRunLeading` / `SetRunTracking` / `SetParagraphJustification` / `SetRunCapsOption`）当前 **verify=roundtrip**（btdk PostScript splice，Go round-trip 绿，从未 AE 实渲验证）。2026-06-16 尝试把它们升到 render-pixel gate，**render-gating 当场揪出从零文字不可渲染**：纯 Go `NewTextLayer` 造的文字层在 AE 里**完全渲染不出来**（连默认 size=88 都看不见），且 `SetRunFontSize(150)` 后 AE 读回 `doc.fontSize = 0.00228 = 150/65536`。**两个症状同源**：从零文字层缺正确的 **text matrix / 坐标 scale**，AE 用 65536 当基把字号/位置全压扁。**writer 没错**（见下 byte 对账），是从零构造缺陷。**这正是 verify=roundtrip 标签掩盖的红线4 缺口**（值 round-trip 绿 ≠ AE 渲对）。

## Bisection（已定位，勿重走）

1. **不是 SetText / from-scratch 文本本身**：同样走 `NewTextLayer`+`SetText` 但**不调** SetRunFontSize 的对照层（TS_DEFAULT），`doc.fontSize` 读回 **88（对）**——唯一变量是有没有调 SetRunFontSize。
2. **不是 writer 写错位置/格式**：dump 两边 btdk（`tmp_debug/parse_btdk`）——
   - AE 原生 `re_text.aep` 的 `size_200` 层：run-style 路径 `/1/1[0]/0/6/0[0]/0/0/6/1` 存 **`/1 200` 纯整数**（非定点）。
   - 我方从零 `SetRunFontSize(150)`：**完全相同路径**存 **`/1 150` 纯整数**，字节格式一致。
   - → writer 写对了。decoder 读 `/1` 为 `v.Num`（plain，与 AE 原生值一致）。Go 编/解自洽。
3. **不是文字颜色**：白 BG + 暗字测量也全空白（TS_DEFAULT 渲出纯白，无字形）。
4. **不是测量阈值**：目视 png 确认画面真的没有任何字形。

→ 结论：AE 读 size_200 的 `/1=200` 得 200，读我方 `/1=150` 得 150/65536。**同字节格式不同解读 = 从零层缺某个 AE 据以缩放字号/定位的字段（text matrix / 文档级 size / 坐标 base，疑 16.16 fixed 上下文）**。fixture 有、从零没有。这也解释从零文字渲染在画面外/不可见。

## 关联已知限制

- **从零文字 Position 不可设**：`Layer.Position()` 对文字层恒 nil（fixture 文字层也 nil），transform Position 默认省略（[[transform-group-default-omission]]）。`showcase/text` 注释明确「无 position/colour/size setter」，只放**单个**默认文字层。本次 render-gate 需把多个文字摆进可测位置，撞上同一堵墙。
- **字体轴用 16.16 定点**：`scene_text_decode.go:58-71` 已知 font variation axes（wght 等）走 `v.Num/65536`；字号 `/1` 当前**不**走定点（decoder `r.FontSize = v.Num`），与 AE 原生 plain 值一致——故定点不是字号存储格式，而是 AE 渲染期对缺 matrix 的从零层施加的 scale。

## 可渲染验证的子集（若将来要做）

- **Tracking / Justification** 在 fixture（已正确定位 + 有 matrix 的 AE 原生文字层）上 round-trip 正常（trk=1000✓、just=center✓ 读回），理论可在 fixture 上 render-gate；从零层因 matrix 缺陷一律渲不出。
- **FontSize / Leading** 即使在 fixture 上也需先确认 AE 是否对其也 ÷65536（未测——本次只在从零层观察到，fixture 期望正常因 matrix 在）。

## 建议（交给需求方决策）

要把文字样式族升到 render-pixel，先解 **从零文字 transform/text-matrix materialize**（synthesis-insert 文字层完整 transform + 文本 matrix，使 Position 可设 + 字号按 plain points 渲染）——这是独立的中-大型 RE arc，**非「中等收益」小活**。在此之前文字样式族**诚实保持 verify=roundtrip**，边界标注「从零文字渲染未验证（matrix 缺陷）；setter 在已 parse 的真实文字层上 round-trip 正常」。**不要**因 Go round-trip 绿就标 render-pixel（假绿）。

## UPDATE 2026-06-16（render-gate 实跑，部分推翻 TL;DR）

前述 TL;DR「从零文字完全渲染不出来 / 默认 size=88 也看不见 / `SetRunFontSize` 读回 ÷65536」**已不再成立**。本次先扫清挡路的 KBar evalScript-timeout 模态（ship-gate 卡 exit-2，已加 `ae_dialog_rules.json` 规则 + 修 cross-volume forensics 丢失，commit b33df56），gate 得以跑完，实测：

- **从零文字现在能渲染**：每个 TS_* comp 的默认文字（AE 默认浅蓝 fill）在 AE2020 实渲里**清晰可见**（目视 png 确认，非空白），不再是「完全看不见」。
- **DOM 读回字号正确**：`doc.fontSize` = 160 / 50 / 88（plain，**非 /65536**）；trk=1200、lead=70/220、just=7413/7415 全对。
- **同源缺陷已被工作树 `internal/codec/text_encode.go` + `internal/serializer/back_layer.go` 改动修掉**（上次会话产出，**尚未 commit**，render-gate 未过故未 ship）。

**但 render-pixel 仍未证成**（红线4 未闭环）：gate 用 `comp.openInViewer()` + `comp.saveFrameToPng` 取帧，实测 **saveFrameToPng 渲染的是 active-viewer comp 而非 receiver comp**，且 headless `-r` 下 viewer 切换不同步（同步脚本占住事件循环，`setActive()` 不生效）→ 9/10 单行 comp 抓到的是同一个 active comp（TS_JL 的 "ABCD"），无法逐 knob 区分。故 **SetRunFontSize 等是否「按值渲染」尚未被像素证实**——只证了「文字能渲染 + DOM 值对」。

**下一步（reachable，独立子活）**：换可靠的逐 comp 取帧（Render Queue 渲 PNG 序列，comp-specific 且不依赖 viewer；或一次 AE run 只渲 startup-active 的单 comp）。证成后文字样式族方可升 render-pixel + commit 文本修复 + 改本 incident `status: resolved`。在此之前维持 verify=roundtrip 标注，文本修复保持 uncommitted（render-gate 未过不算 ship）。
