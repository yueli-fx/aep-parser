---
showcase: essential-graphics
direction: Essential Graphics — AddEssentialProperty + SetMotionGraphicsTemplateName，纯 Go 从零把三个效果参数提升为 EG 控件并命名模板，AE EG DOM readback 核对（📋 读值档，不看图）
capabilities: [add-essential-property, set-motion-graphics-template-name, eg-controller]
gates: [TestEssentialGraphics_AEShipGate_AE2020, TestEssentialGraphics_AEShipGate_AE2025, essential-graphics-write-re]
status: 待review
last_updated: 2026-06-14
regenerate: "go run ./flightdeck/showcase/essential-graphics  +  scripts/ae_run.ps1 verify.jsx"
---

# essential-graphics — Essential Graphics 控件 showcase（📋 读值档）

## 这个方向测什么

`AddEssentialProperty` 把图层效果参数提升为 **Essential Graphics 面板控件** + `SetMotionGraphicsTemplateName` 命名模板。EG 控件在 EG 面板里(无渲染视觉) → 读值档：AE 打开后 DOM 确认模板名 + 控件数 + 各控件名。

## 验证方式（📋 读值，非看图）

`verify.jsx` 打开 .aep,读 `comp.motionGraphicsTemplateName` / `motionGraphicsTemplateControllerCount` / `getMotionGraphicsTemplateControllerName(k)` 到 `.done`。无 png。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | host solid + 3 控制效果 + 提升为 EG 控件 + 命名模板 |
| `verify.jsx` | 读值脚本(tracked) | dump EG DOM → `.done`（无 png） |
| `essential_graphics.aep` | 产出工程 (gitignored) | 1920×1080，AE2020 |
| `essential_graphics.done` | readback 日志 (gitignored) | 用户读这个核值 |

## 设值 → 期望 readback（AE2020 实测一致）

| 操作 | DOM 字段 | 期望值 |
|---|---|---|
| SetMotionGraphicsTemplateName("Showcase EG Template") | motionGraphicsTemplateName | "Showcase EG Template" |
| AddEssentialProperty ×3 | motionGraphicsTemplateControllerCount | 3 |
| Slider Control → expose | controller#1 | "Blur Amount" |
| Color Control → materialize+expose | controller#2 | "Accent Color" |
| Checkbox Control → expose | controller#3 | "Enable Glow" |

> 审核要点：模板名对 + 控件数=3 + 三控件名顺序对即通过。

## 注记（elision 行为，非缺陷）

**Color 控件须先 materialize**：color 参数默认被 elide(无值流),直接 expose 报错;须先 `SetEffectParam(fx, "ADBE Color Control-0001", []float64{...})` 物化值再提升（slider/checkbox 可直接 expose）。这是 AE 的参数持久化规则（值≠默认才存),非缺陷。注意 SetEffectParam 的值类型是 `[]float64`(切片)非 `[4]float64`。

## 溯源

EG 写路径（CIF*/CCtl/OvG2/CPrp）：`incidents/essential-graphics-write-re.md`；参数 elision/materialize：`incidents/effect-param-elision-synthesis-lite.md`；ship-gate `essential_graphics_shipgate_test.go`。
