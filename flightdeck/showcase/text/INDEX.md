---
showcase: text
direction: 文字层 from-scratch — 纯 Go 从零创建文字层并写入多行字符串（NewTextLayer + SetText），AE 实渲
capabilities: [new-text-layer, set-text, multiline-text, length-variable-string]
gates: [TestNewTextLayer_AEShipGate_AE2020, TestNewTextLayer_AEShipGate_AE2025]
status: complete
last_updated: 2026-06-13
regenerate: "go run ./flightdeck/showcase/text  +  scripts/ae-worker/ae_run.ps1 render.jsx"
---

# text — 文字层 from-scratch showcase

## 这个方向测什么

从零 `NewTextLayer` 建一个文字层，`SetText` 写入用 `\r` 分段的多行字符串（`Reopen` 后走 chunk-backed 变长路径）。AE 2020 渲染 frame 0 → `text.png`，文字以 AE 默认样式（88pt、默认字体、浅色）渲出三行。

## ⚠ 真实边界（交付准则 · 诚实标注）

**from-scratch 文字只覆盖「建层 + 写字符串」**，**没有**位置 / 颜色 / 字号 / 字体的公共 setter：
- 文字层的 `Position` 属性未物化 → `SetPosition` 报「Position property not present」，**多个文字层会全部堆在同一默认位置重叠**。故本 showcase 用**单层 + 多行字符串**，而非多个独立摆放的文字层。
- 颜色/字号/字体 = AE 种子模板默认值（浅蓝、88pt、默认字体），不可改。
- 字符串长度可变（变长 btds 路径），`\r` 分段多行有效。

这是当前文字能力的**完整作用面**——演示里不暗示「可排版文字」。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `text.aep`（2 层 = 1 文字 + BG） |
| `render.jsx` | 渲染脚本(tracked) | 渲 frame 0 + dump 每层 text/fontSize 到 .done |
| `text.aep` | 产出工程 (gitignored) | 1920×1080，AE2020 target |
| `text.png` | 渲染帧 (gitignored) | frame 0 |

## 布局

| 图层 | 内容 | 样式 |
|---|---|---|
| T | `AEP-PARSER` / `from-scratch text` / `no AE needed`（`\r` 三段） | 默认 88pt 浅色，默认位置（偏右） |

> 审核要点：三行文字清晰可读 = SetText 多行变长写入生效；位置/颜色为默认 = 符合上述边界。

## 溯源

文字层创建 + 变长 SetText：`incidents/text-btdk-length-variable-write-scoping.md`、`kerning-first-enable.md`；gate：`new_text_layer_shipgate_test.go`。
