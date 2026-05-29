---
when_to_read: adding any NewX layer/structural write path; debugging AE 2020 "项目文件似乎已损坏（跳过部分）" / silent layer skip; touching buildLdtaBytes or the capability matrix
applies_to: [ldta, shapelayer, capability-matrix, ae2020, ship-gate, cross-version]
last_updated: 2026-05-29
---

# AE 2020 把 164B ShapeLayer ldta 判为损坏并跳过该层

**ShapeLayer 的 ldta 大小必须按 target 分支：AE 2020/2022 = 160B，AE 2025 = 164B。** 写错会让 AE 2020 把整个图层判为损坏并静默跳过。

## 症状

AE 2020 打开含 ShapeLayer 的 .aep 时弹警告：
> 您的项目文件似乎已损坏（跳过部分：1）。请使用不同名称保存。(26 :: 0)

AE 2025 打开同一文件**无任何提示**（宽容接受 164B）。纯 comp（无图层）AE 2020 也正常——只有 shape 图层触发。

## 根因

`buildLdtaBytes` 曾硬编码 `make([]byte, ldtaSize2025)`（164B），无视 target。AE 2020 的 ldta reader 对 ShapeLayer ldta 期望 160B；读到 164B 即判该层数据损坏、跳过（"跳过部分：1" = 跳过 1 个图层）。AE-2020-native 存的 ShapeLayer ldta 实测 160B，AE-2025-native 实测 164B（diff 仅尾部 4 个 zero-pad；所有有效字段都在前 0x88 内，所以尺寸只影响尾部补零）。

## 为何长期未发现

V2.2 的 AE-2020 shape ship-gate（`TestV2_2_AEShipGate_AE2020`）**一直是 `t.Skip()` 状态**，从未真正跑过 from-scratch shape 图层过 AE 2020。capability matrix 的 RE-S9 评估曾把 `LdtaSize` 列为"不满足 §1.4 admission rule"——基于"AE 2020 也接受 164B"的**未验证假设**。该假设错误。cross-Project InsertLayer 虽过 AE 2020，但它插进的是 **AE-native** comp（ldta 来自 AE，非我们 buildLdtaBytes），所以没暴露。

## How to apply

- 任何 `NewX` 结构性写路径必须 **AE 2020 + AE 2025 双版本 ship-gate** 真跑（CLAUDE.md 硬约束 #6）——skip 的 gate 等于没验证。
- ldta 尺寸走 `ctx.capabilities.LdtaSize`（capability matrix），勿硬编码。新版本 target 加进 `Capabilities()`。
- 调 AE 2020 "损坏/跳过" 类问题：先取 **AE-2020-native 参考**（让 AE 2020 存同款），dump 对比 chunk 尺寸；optical OCR 在共享桌面会被遮挡，用 **PrintWindow**（PW_RENDERFULLCONTENT，遮挡免疫）抓对话框，或 Win32 枚举子控件。
- 此修复同时解锁了 V2.2 Rect+Fill 在 AE 2020（同 ldta 路径，之前同样 broken）。

## 修复

- `capability_matrix.go`: `AECapabilities.LdtaSize`；`Capabilities(target)` 按版本填 160/164。
- `lower_layer.go buildLdtaBytes`: `size := ctx.capabilities.LdtaSize`（0 回退 164）。
- 测试：`TestLowerShapeLayer_LdtaSizeByTarget`（单测，三 target）+ `TestV2_2_Ellipse_AEShipGate_AE2020/2025`（集成，双版本 PASS）。
