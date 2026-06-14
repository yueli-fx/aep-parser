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

字节机制与表达式内容无关（tdb4 @0x77/@0x78 + tdbs Utf8），gate 证明的是 AE **求值**这三类 idiom + 渲染像素（LEAD/LINK/LOOP 确定性命中，WIG 非确定故仅验「渲染未丢」）。**AE 2020 + AE 2025 双版本 PASS**。

**仍未 gate：**
- 表达式驱动 1D 标量以外维度的更复杂语汇（`linear()`/`ease()` remap、`valueAtTime` 组合）——边际证明值低（机制已证内容无关），按需补。
- **表达式引用 effect 参数（`effect(1)("Slider")` slider-control 绑定）= 被一个 AddEffect 问题挡住，非表达式问题。负向发现（2026-06-14 尝试）：把 Slider Control 加到 SLD 后在 7 层从零 comp 里 `effect(1)` 在 AE 读回 null——AE **静默 drop 了效果**；改挂 SOLID 宿主 CTRL（essential-graphics/set-effect-param gate 已证单层 solid+slider 可接受）仍被 AE drop。Go round-trip 全程保留效果（`l.Effects`=1、params=2），AE 不保留——典型「Go round-trip ≠ AE 接受」。即 AddEffect 在「从零多层 comp」上下文回归（单层 EG gate 绿、本 7 层场景 drop），疑似 item-ID / tdpi host-binding / 层序交互（关联 `add-effect-splice-re.md`、`multi-layer-silent-drop.md`、`nextitemid-must-include-layer-ids.md`）。修法走最小失败 bisection（剥到 solid+slider 单测 → 逐层加回定位 drop 触发点），未做，deferred。表达式侧已就绪，解开 AddEffect 后即可补 SLD idiom。**注**：AddEffect 在 SHAPE 层上本就未 gate（同样 AE drop）。**

> 写 AI 生成 MG 表达式时查语义：`flightdeck/references/after-effects-expression-reference/`（docsforadobe，docs/ 按 objects/layer/general/text 分组）。

## Cases
- 2026-06-12 首次（MG roadmap S2；第二阶段 disabled-丢文本是修复过程中的次生发现，一并修复）
- 2026-06-14 表达式语汇 gate（S2 followup）：loopOut / wiggle / 跨层引用三 idiom 双版本 ship-gate PASS；Go 侧零改动（机制内容无关），新增 `expr_vocab_shipgate_test.go`。JSX 数组日志须逐元素索引（`v2s()`），直接拼数组对象触发 ExtendScript「数字结果无效（除以零？）」throw（同 `effect-param-elision-synthesis-lite.md` 坑）。
