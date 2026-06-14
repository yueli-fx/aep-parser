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

覆盖边界：rotation 1D scalar 表达式实测；`time*90` 单表达式。**表达式语汇扩展已 gate**（2026-06-14，S2 followup，`expr_vocab_shipgate_test.go` + `verify_expr_vocab.jsx`）：

| 语汇 | 表达式 | 新覆盖点 | t=2.5s 求值 |
|---|---|---|---|
| 跨层引用 | `thisComp.layer("LEAD").transform.position + [0,250]` | 按名解析其它层属性 + 矢量算术 | `[480,500]` ✓ |
| loopOut | `loopOut("cycle")` | **表达式叠加在带关键帧属性上**（新组合：之前只验静态属性） | `[899.997,750]`（循环相位 0.5 中点；无循环则保持末帧 1500）✓ |
| wiggle | `wiggle(2,250)` | 过程式/时变（偏离锚点且 t=1≠t=2.5） | `[857,916]` ✓ |
| effect-param 引用 | `[effect(1)(1), 350]`（SLD 上挂 Slider Control = 880，表达式引用之） | **表达式读 effect 参数**（AddEffect + SetEffectParam + expression 三者组合，MG slider 绑定核心） | `[880,350]`（slider 值经表达式驱动 x）✓ |

字节机制与表达式内容无关（tdb4 @0x77/@0x78 + tdbs Utf8），gate 证明的是 AE **求值**这四类 idiom + 渲染像素（LEAD/LINK/LOOP/SLD 确定性命中，WIG 非确定故仅验「渲染未丢」）。**AE 2020 + AE 2025 双版本 PASS**。

**effect-param 引用的 RE 教训（按索引引用，不是名字）：**`effect(1)("Slider")` 与 DOM `layer.effect("Slider Control")` / `.property("Slider")` **全部解析失败**——AddEffect 生成的效果实例**名 = match-name「ADBE Slider Control」**（非显示名 "Slider Control"），且我们物化的 slider 参数**显示名不是 "Slider"**（set_effect_param 一向按 match-name "ADBE Slider Control-0001" 访问）。故：(a) 表达式引用效果用**效果索引 + 参数索引** `effect(1)(1)`（求值出 880）；(b) JSX 读回走 `layer.property("ADBE Effect Parade").property(1)`（`.effect()` DOM 访问器在此也 flaky）。表达式若按错误名字引用，AE **不报错、静默回退静态值**（实测 `effect(1)("Slider")` → x=960 静态，假绿陷阱）。

**误诊更正（2026-06-14）：**最初以为「AddEffect 在 shape / 多层 comp 被 AE 静默 drop」——**错**。bisection（L1 单 shape+slider → L6 全复杂度 SLD 最后建）逐级全 PASS，效果从未被 drop。真因是上面的「按错误名字访问」：① JSX `effect("Slider Control")` 名字错 → null → throw（被误读成 drop）；② 表达式 `effect(1)("Slider")` 参数名错 → 静默回退静态。AddEffect 在 shape 层、多层、SLD 末位建——全部正常。**教训：负向发现下结论前先把验证脚本的访问路径排除掉（同「先看产物再玩数字」）。**

**仍未 gate：**`linear()`/`ease()` remap、`valueAtTime` 组合等——边际证明值低（机制已证内容无关），按需补。

> 写 AI 生成 MG 表达式时查语义：`flightdeck/references/after-effects-expression-reference/`（docsforadobe，docs/ 按 objects/layer/general/text 分组）。

## Cases
- 2026-06-12 首次（MG roadmap S2；第二阶段 disabled-丢文本是修复过程中的次生发现，一并修复）
- 2026-06-14 表达式语汇 gate（S2 followup）：loopOut / wiggle / 跨层引用 / effect-param 引用**四** idiom 双版本 ship-gate PASS；Go 侧零改动（机制内容无关），新增 `expr_vocab_shipgate_test.go`。JSX 数组日志须逐元素索引（`v2s()`），直接拼数组对象触发 ExtendScript「数字结果无效（除以零？）」throw（同 `effect-param-elision-synthesis-lite.md` 坑）。effect-param idiom 先误诊为「AddEffect 被 AE drop」，bisection 证伪 → 真因是引用按名失败须改索引（见上节）。
