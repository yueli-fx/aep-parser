---
status: active
when_to_read: SetExpression 写入后 AE 不求值/expressionEnabled 读回 false；AE 打开后表达式文本被丢；touching tdb4 @0x77/@0x78 or SetExpression/SetExpressionEnabled; 评估「Go round-trip 绿但 AE 行为不对」的表达式类症状
applies_to: [expression, expression-enabled, tdb4, 0x77, 0x78, has-expression-marker, settext, render-dead, mg-roadmap, ship-gate, ae2020, ae2025]
last_updated: 2026-06-12
resolved_by:
---

# 表达式激活 = tdb4 @0x77/@0x78 字节对 — 历史反语义解读翻案

## Signature
- symptom: Go SetExpression 写入的表达式 AE 不求值（JSX 读 expressionEnabled=false / valueAtTime 不变）；写「disabled 表达式」时 AE 打开后表达式文本整个被丢（expression 读回空串）；Go 自读回全程正常（假绿）
- error_type: —
- where: internal/serializer/back_property.go SetExpression/SetExpressionEnabled + parse_properties.go @0x78 解析
- trigger: from-scratch 或局部改写工程中给属性挂表达式后在 AE 打开/渲染（MG roadmap S2 首撞；delivery-contract 红线1 的原始教训源）

## 症状/复现

`SetExpression("time*90")` + `SetExpressionEnabled(true)` → Go 读回全对，AE 里 expressionEnabled=false、不求值（render-dead）。第二阶段：修了 enabled 位后写 disabled 变体，AE 打开直接**丢掉表达式文本**。

## 根因

tdb4 @0x77/@0x78 是**两个独立字节**，历史 RE 把它们混为一个「@0x78 反语义 disabled 位」：

| 状态 | @0x77 | @0x78 |
|---|---|---|
| 有表达式，enabled | 00 | 01 |
| 有表达式，disabled | 01 | 01 |
| 无表达式 | 00 | 00 |

- **@0x78 = has-expression 标记**（与 tdbs 里的表达式 Utf8 必须同步——Utf8 在而 @0x78=0 时 AE 当无表达式、open 时把 Utf8 丢掉）。
- **@0x77 = disabled 位**（1 = 保留文本但不求值）——这才是 expressionEnabled 的真身。

历史解读「@0x78 inverted: 0=enabled」恰好对无表达式属性（00 00）成立、又被「Go 自读回」循环验证遮蔽，致 SetExpression 全线 render-dead 多月未察觉。RE 法：AE 2025 原生 enabled/disabled fixture 对（`tmp_debug/expr_re/gen_native_expr*.jsx`）+ `difftdb4` 全字节 diff。

## 修法

- `SetExpression`：写/删 Utf8 时同步 @0x78（1/0）。
- `SetExpressionEnabled`：写 @0x77（enabled→0 / disabled→1），不再碰 @0x78。
- parse：`Expression != ""` 时 `ExpressionEnabled = (@0x77 == 0)`，否则默认 true。
- gate `TestExpression_AEShipGate_*`（红线4 渲染像素）：ON 层 `time*90` 求值后 dot 渲染在锚点下方、OFF 层同表达式不动；JSX 读回 enabled/求值双态 + resave 双态存活。**AE 2020 + AE 2025 双版本 PASS**。

覆盖边界：rotation 1D scalar 表达式实测；`time*90` 单表达式；loopOut/wiggle/属性间引用（thisComp.layer 链）未单独 gate——给 AI 生成 MG 用前建议补常用表达式语汇 gate（roadmap S2 followup）。

## Cases
- 2026-06-12 首次（MG roadmap S2；第二阶段 disabled-丢文本是修复过程中的次生发现，一并修复）
