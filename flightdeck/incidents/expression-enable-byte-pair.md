---
status: active
when_to_read: SetExpression 写入后 AE 不求值/expressionEnabled 读回 false；AE 打开后表达式文本被丢；给 effect param（或任何带 tdum/tduM 的属性）挂表达式后 AE 判损坏跳过该层；touching tdb4 @0x77/@0x78 or SetExpression/SetExpressionEnabled or 表达式 Utf8 在 tdbs 里的插入位置；评估「Go round-trip 绿但 AE 行为不对」的表达式类症状
applies_to: [expression, expression-enabled, tdb4, 0x77, 0x78, has-expression-marker, utf8-position, tdum-tduM, effect-param-expression, settext, render-dead, mg-roadmap, ship-gate, ae2020, ae2025]
last_updated: 2026-06-17
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

### 二次纠错（2026-06-15）：Utf8 必须插在 cdat 后、tdum/tduM 前（不是 append）

`SetExpression` 原本 **append** 表达式 Utf8 到 tdbs.Children 末尾。裸 Transform 标量的 tdbs =
`tdsb/tdsn/tdb4/cdat`——append 恰好落在 cdat 后，AE 接受（旧 gate 全绿系巧合）。但**物化的
effect param** tdbs 额外带 `tdum/tduM`（min/max 范围），append 把 Utf8 放到 tduM **之后** →
AE 2020 开工程时急切解码表达式、读到乱序流判**损坏 → 静默跳过该层**（「项目文件似乎已损坏：跳过
部分」）。又一个**红线1 假绿**：Go `SetExpression` 返回 nil，AE 拒。Ground truth（AE 自存
fixture，`tmp_debug/dump_expr_effect` 逐 chunk dump）：AE canonical 序 = `…cdat, Utf8, tdum,
tduM`。修：Utf8 **插在最后一个 cdat（无 cdat 则 tdb4）之后**，tdum/tduM 之前；无 tdum/tduM 时插入位
== 旧 append 位，故 Transform 标量 gate 不受影响。`back_property.go::SetExpression`。
**这把表达式从「只能挂 Transform 等无 tdum/tduM 的属性」扩到「任意属性含 effect param」**。
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

**按函数补 gate = 关闭（closed decision，2026-06-17）：**`linear()`/`ease()` remap、`valueAtTime` 组合等**不再单独 gate**。理由：这些全是 AE 官方内置表达式函数（AE 必认，风险不在"AE 认不认函数"），且字节写入路径与表达式内容无关——上面已 gate 的 5 类 idiom（静态值 / 跨层引用 / 带关键帧叠加 / 时变随机 / 读 effect 参数）已**穷尽写入端的字节情况**：普通标量 tdbs（无 tdum/tduM，Utf8 落 cdat 后）与 effect param tdbs（带 tdum/tduM，Utf8 须插 cdat 后/tduM 前）两种插入位都覆盖了，新函数不会引入新字节路径。"未验"的潜台词从来不是"AE 不认函数"，而是"怕换写法时我们这边又有隐藏耦合"（如当年 tdum/tduM 那次）——该担心已被这 5 类堵死。

> **若哪天某 remap 真出问题：**几乎一定在 AE 求值端（AE 的事），不在我们的写入端。下结论前先 `difftdb4` 拿 AE 自存的同表达式 fixture 逐字节对账（同红线 1 原始教训），确认 @0x77/@0x78 + Utf8 插入位无误后再去怀疑求值——别又凭 Go round-trip 绿假绿。

> 写 AI 生成 MG 表达式时查语义：`flightdeck/references/after-effects-expression-reference/`（docsforadobe，docs/ 按 objects/layer/general/text 分组）。

## Cases
- 2026-06-12 首次（MG roadmap S2；第二阶段 disabled-丢文本是修复过程中的次生发现，一并修复）
- 2026-06-14 表达式语汇 gate（S2 followup）：loopOut / wiggle / 跨层引用 / effect-param 引用**四** idiom 双版本 ship-gate PASS；Go 侧零改动（机制内容无关），新增 `expr_vocab_shipgate_test.go`。JSX 数组日志须逐元素索引（`v2s()`），直接拼数组对象触发 ExtendScript「数字结果无效（除以零？）」throw（同 `effect-param-elision-synthesis-lite.md` 坑）。effect-param idiom 先误诊为「AddEffect 被 AE drop」，bisection 证伪 → 真因是引用按名失败须改索引（见上节）。
