---
status: active
---

# Composition.Renderer W — RE findings (P3, pre-implementation)

**Created**: 2026-06-01（RQ reader/writer arc 收尾时顺手探的）

下一个 P3 结构性候选 `Composition.Renderer` 写。R 已有（`parse_composition.go`：`PRin` LIST → `prin` chunk @offset 4 的 ASCII NUL-padded match-name）。写之前先把 RE 探明记下，省得下次重探。

## chunk 布局

comp 的渲染器存在 Item 下的 `LIST:PRin` 里：
- `prin` chunk：**固定 104B**，renderer match-name 是 ASCII NUL-padded，**@offset 4**。改名是 length-preserving。
- `prda` chunk：renderer-specific 选项，**长度随 renderer 变**：

| renderer (内部名) | UI | prda len |
|---|---|---|
| `ADBE Escher` | Classic 3D | 12 |
| `ADBE Calder` | Advanced 3D | 52 |
| `ADBE Ernst` | Cinema 4D | 20 |
| `ADBE Picasso` | Ray-traced 3D | 16 |

（实测自 py-aep `samples/models/composition/renderer_{classic_3d,advanced_3d,cinema_4d,ray_traced}.aep`，各自单 comp。）

## 结论：Renderer W 是结构性写

换 renderer = 改 prin 名（定长）+ **替换变长 prda chunk** → 父 `PRin` LIST size 变 → **结构性**。按 CLAUDE.md 硬约束 #6 必须跑 AE 2020 + 2025 双版本 ship-gate 才算 ship。

## 实现路线（待执行）

1. 抽 4 个 renderer 的 prda byte 模板（直接从上面 4 个 fixture dump）→ Go table（类似 templates 内嵌）。
2. `(c *Composition) SetRenderer(name string)`：
   - prin @4 写新 match-name（NUL-pad 到 104B 内）
   - 用目标 renderer 的 prda 模板替换现 prda chunk（rifx 自动重算父 LIST size）
   - 走 V2.1 atomic invariants：snapshot + warnings-as-failure + rollback
3. JSX RE fixture：`comp.renderer = "ADBE Escher"` 等，AE 2020+2025 双开校验接受 + readback。

## 开放问题

- **内部名随 AE 版本/上下文变**：现有 R 测试注释说「AE 2025 里 Classic 和 Advanced 都塌成 `ADBE Escher`」，但 py-aep fixture（更早 AE）是 classic=Escher / advanced=Calder。SetRenderer 的入参该用内部名还是 UI 名 + 版本映射？需先定 API 语义（建议入参内部 match-name，映射表另列）。
- prda 模板是否跨 AE 版本稳定？ship-gate 双版本时一并验。
- prin 的其余 100B（@8..103）是否也 renderer-specific？dump 时一并比对，若有差异要连 prin 体一起换。

## 相关

- 现有 R：`parse_composition.go` § Renderer（PRin→prin@4）+ `composition_display_test.go::TestCompositionRendererReal`（注意：本地 `re_renderer.aep` 是 9KB 极简 fixture，无 prin/prda，该测试在本机 vacuous pass）
- 写机制参考：V3 结构性 ops（DeleteLayer/MoveLayer 等）的 snapshot+rollback 模式
- spec §2.2 Renderer 行
