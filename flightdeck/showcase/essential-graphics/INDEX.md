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

| 文件                        | 类型                       | 说明                                                |
| --------------------------- | -------------------------- | --------------------------------------------------- |
| `gen.go`                  | 生成器(Go, tracked)        | host solid + 3 控制效果 + 提升为 EG 控件 + 命名模板 |
| `verify.jsx`              | 读值脚本(tracked)          | dump EG DOM →`.done`（无 png）                   |
| `essential_graphics.aep`  | 产出工程 (gitignored)      | 1920×1080，AE2020                                  |
| `essential_graphics.done` | readback 日志 (gitignored) | 用户读这个核值                                      |

## 设值 → 期望 readback（AE2020 readback 一致）

| 操作                                                  | DOM 字段                              | 期望值                 |
| ----------------------------------------------------- | ------------------------------------- | ---------------------- |
| SetMotionGraphicsTemplateName("Showcase EG Template") | motionGraphicsTemplateName            | "Showcase EG Template" |
| AddEssentialProperty ×1（Slider）                     | motionGraphicsTemplateControllerCount | 1                      |
| Slider Control → expose                               | controller#1                          | "Blur Amount"          |

> 审核要点：模板名对 + 控件数=1 + 控件名对。**但 readback 通过 ≠ 面板无恙** —— 见下「崩溃发现」,请务必**真机展开「基本图形」面板**确认不崩。

## 🛑 崩溃发现（红线4b：超出 gate 覆盖边界）

**最初版用 3 个混合控件(slider + 物化 color + checkbox),DOM readback 全过(模板名+3 控件名读得到),但用户真机一展开「基本图形」面板就崩溃 AE。** 根因方向：EG ship-gate **只验过 1 个 slider 控件,且只验 load + DOM 读 + resave,从不打开 EG 面板** → 多控件 / color / checkbox 的 EG 结构(CIF3/CCtl/OvG2/CprC)在面板渲染时不被 AE 接受,属 gate 未覆盖的组合/规模(红线4b)。

**本档已退回到 gate 唯一证过的「单 slider 控件」。** ⚠ **仍待用户真机验**:展开基本图形面板,单 slider 是否也崩?
- 若**不崩** → 单 slider 从零 EG 可交付(本档成立);多控件/color/checkbox 标为 RE 候选。
- 若**仍崩** → 即便 gate-proven 的单 slider 从零 EG 面板也不安全(gate 从不开面板=假绿根源),整个 from-scratch EG 退回未验证,需独立 RE。

> 「Go round-trip + DOM readback 双绿 ≠ AE 真能用」的又一活样本——验证必须验到能力的真实作用面(EG 的作用面是**面板**,非 DOM 计数)。

## 注记（elision 行为，备查）

**Color 控件须先 materialize**（已从本档移除）：color 参数默认被 elide(无值流),直接 expose 报错;须先 `SetEffectParam(fx, "ADBE Color Control-0001", []float64{...})` 物化值再提升。SetEffectParam 值类型是 `[]float64`(切片)非 `[4]float64`。

## 溯源

EG 写路径（CIF*/CCtl/OvG2/CPrp）：`incidents/essential-graphics-write-re.md`；参数 elision/materialize：`incidents/effect-param-elision-synthesis-lite.md`；ship-gate `essential_graphics_shipgate_test.go`。
