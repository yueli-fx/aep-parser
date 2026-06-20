---
showcase: pseudo-effect
direction: 纯 Go 从零合成一个伪效果，单个效果里铺满全部 11 种控件类型，出英文 + 中文两份供眼验 Effect Controls 面板
capabilities: [build-pseudo-effect, pseudo-slider, pseudo-color, pseudo-checkbox, pseudo-angle, pseudo-point, pseudo-point3d, pseudo-dropdown, pseudo-group, pseudo-label, pseudo-layer, with-label-codepage]
gates: [TestBuildPseudoEffect_AEShipGate_AE2020, TestBuildPseudoEffect_AEShipGate_AE2025, TestBuildPseudoEffectRich_AEShipGate_AE2020, TestBuildPseudoEffectRich_AEShipGate_AE2025, TestBuildPseudoEffectValueEntry_AEShipGate_AE2020, TestBuildPseudoEffectValueEntry_AEShipGate_AE2025]
status: 待review
last_updated: 2026-06-20
regenerate: "go run ./flightdeck/showcase/pseudo-effect  (B 类读值：AE 打开看 Effect Controls 面板，或跑 test_data/verify_pseudo_effect.jsx 结构 dump)"
---

# pseudo-effect — showcase（B 类 读值档）

## 这个方向测什么

`BuildPseudoEffect` 纯 Go 从零合成伪效果（无 .ffx、无 AE、不 clone 模板字节），单个 "Demo" 效果里
**铺满全部 11 种控件类型**，证明 Pseudo Effect Maker 能造的控件都能离线合成。伪效果控件是**参数容器、
本身不渲染像素**，故这是 **B 类读值**示例：判据 = AE 打开后 Effect Controls 面板里控件树活、类型/值/标签对，
不是渲染帧。

含的全部控件类型（出现顺序）：Label · Slider · Angle · Color · Checkbox · Dropdown · Label · Point ·
Point3D · Layer-picker · Group(起 + 子 Slider + 子 Checkbox + 止)。

两份产物的区别只在标签语言 + 码页：
- **en**：纯 ASCII 标签 → 码页无关，任何系统显示对。
- **zh**：中文标签（GBK 编码 pard 名）→ **中文(GBK)Windows 显示对**；效果显示名「演示」走值组 tdsn(UTF-8)、
  码页无关处处对。日语版同理改 `WithLabelCodepage(PseudoLabelShiftJIS)`（本 showcase 未出，机制见
  `incidents/pseudo-control-label-ansi-codepage.md`）。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| gen.go | 生成器(Go) | 纯 Go 构建两份 .aep（comp 400×400 / 2 shape 层：Host 挂效果 + Source 供 Layer-picker 绑定） |
| pseudo_demo_en.aep | 产出工程 (gitignored) | 英文标签（纯 ASCII） |
| pseudo_demo_zh.aep | 产出工程 (gitignored) | 中文标签（GBK） |

## 已自验（AE 2020，结构 dump verify_pseudo_effect.jsx）

- 两份都 **AE 接受、效果活**（matchName `Pseudo/{en,zh}/Demo`、enabled、numProperties=17）。
- 全 11 种控件读回齐全；**组是扁平标记控件**（控件平铺，非属性树嵌套——AE 原生行为，
  详 `incidents/pseudo-control-label-ansi-codepage` 同批 RE 与 facade doc）。
- en 标签全部正常显示；zh 效果显示名「演示」正常（charCodes 28436,31034），zh **控件标签**在西文 gate 机
  显示乱码 = 预期（GBK 字节被 cp1252 误读），**待用户中文真机复核**翻 complete。

> ⚠ unverified（待办）: zh 控件标签在中文 Windows 的显示 — 需用户真机打开 `pseudo_demo_zh.aep` 复核
> （西文 gate 机不可验，byte-equivalence 单测 `TestPardNameBytes_GBKMatchesAENative` 兜底）。
