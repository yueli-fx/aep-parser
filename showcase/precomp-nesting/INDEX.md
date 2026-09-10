---
showcase: precomp-nesting
direction: 预合成嵌套 — 纯 Go 从零生成一个 child 合成（含三个 badge 的组合场景），由 parent 合成嵌套一次并渲染出整套 child 内容
capabilities: [new-composition, new-precomp-layer, nested-composition, source-id-resolution, transparent-precomp]
gates: [TestMGPrecomp_AEShipGate_AE2020, TestMGPrecomp_AEShipGate_AE2025]
status: complete
last_updated: 2026-06-13
regenerate: "go run ./showcase/precomp-nesting  +  scripts/ae-worker/ae_run.ps1 render.jsx"
---

# precomp-nesting — 预合成嵌套 showcase

## 这个方向测什么

建两个合成：child「Badge」装一个自包含的组合场景（三个「品红圆盘 + 白星」徽章横排，**无全屏底 → 周围透明**），parent「PrecompMain」在暗底上**嵌套 child 一次**（`NewPrecompLayer`，full-frame）。parent 渲染出整套 child 内容 = 证明预合成层正确解析到 child 的 SourceID 并把整个嵌套合成渲染出来。parent 只有 2 层（BG + 1 个嵌套层），画面里的 3 个徽章全部来自 child。`NewPrecompLayer` 在 `Reopen` 后调用（需 item-list 回引），与 S4 ship-gate 同路。

> 说明：精合成层的**重定位/缩放**（多实例摆位）不在公共 API 覆盖内（`SetPosition` 对精合成层 Position 未物化）——本 showcase 用「child 内部摆 3 个徽章 + 嵌套一次」展示嵌套+组合，而非「同一 child 多实例变换」。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `precomp_nesting.aep`（parent 2 层 + child 6 层） |
| `render.jsx` | 渲染脚本(tracked) | 渲 parent `PrecompMain` frame 0 |
| `precomp_nesting.aep` | 产出工程 (gitignored) | 两个合成，AE2020 target |
| `precomp_nesting.png` | 渲染帧 (gitignored) | parent frame 0 |

## 布局

| 合成 | 层 | 内容 |
|---|---|---|
| Badge (child) | Disc1-3 / Star1-3 | 三个徽章（品红圆盘 + 白五角星）@ x=480/960/1440，透明底 |
| PrecompMain (parent) | BG + NESTED_Badge | 暗底 + 嵌套 child 一次（full-frame） |

> 审核要点：parent 渲染出三个徽章 = 嵌套生效；徽章间是暗底（透明传递）= child 无不透明底，未互相遮挡。

## 溯源

S4 precomp source-ID RE：`incidents/precomp-layer-source-id-re.md`；gate：`mg_precomp_shipgate_test.go`。
