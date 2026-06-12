---
status: active
last_updated: 2026-06-12
when_to_read: 把任何能力/产物对外宣称「能用」之前；做演示/示例/端到端工程之前；判断某次实现到底算「交付」还是「原型」；review 别人或自己声称 ship 的东西；纠结「Go 测试过了算不算数」
applies_to: [delivery, ship-gate, ae-acceptance, verification-level, go-roundtrip-not-ae, coverage-boundary, composition-gate, end-to-end, deliverable-vs-prototype, honest-scope]
---

# 交付准则 — 什么算「能用 / 合法交付」

> 2026-06-12 立。触发：orbit 动画 demo 把一堆「各自验证过、组合起来没验证过」的能力拼成一个工程，还混进了「从没经 AE 验证」的表达式，AE 打开直接卡死。用户判定「做了很多但很多不一定能用 = 等于白做」。此准则把「可交付」从含糊变成红线。

## 铁律

**只有经过 AE 2020 + AE 2025 双版本 ship-gate 实测 PASS 的能力，且仅在其 gate 覆盖的「规模 + 组合」边界内，才可对外宣称「能用 / 可交付」。** 其余一律视为未验证，不得在交付物、演示、示例、对用户的能力描述中当现成能力使用——除非当场显式标「未验证」。

## 三级验证状态

| 级别 | 定义 | 能否交付 / 演示 |
|---|---|---|
| **Verified** | AE 双版本 ship-gate 实测 PASS（有 AE 读回输出为证） | ✅ 可，但**仅限 gate 覆盖的规模 / 组合** |
| **Go-roundtrip-only** | 仅 Go 解析器往返自洽（`WriteAEP`→`Open` 无警告），AE 接受性未知 | ❌ 默认**不可用**；必须显式标「未验证」；禁止用于交付 / 演示 / 示例 |
| **未测** | 既无 AE gate 也无明确 round-trip 证据 | ❌ 无交付资格 |

库里的 `Stable` ≈ Verified，`Alpha` ≈ Go-roundtrip-only（详 CLAUDE.md #2 + `plans/coverage.md`）。**但 `Stable` 不等于无条件可用**——见红线 2。

## 三条认知红线（每条都有血的教训）

1. **Go round-trip 0 警告 ≠ AE 接受。** Go parser 宽松：它只验证「字节能读回成模型对象」，**不验证 AE 的几何 / 表达式 / 渲染引擎能否消化**。结构自洽 ≠ AE 认账。
   - 教训：`SetExpression` 只做过 Go round-trip（合成 fixture），从没 AE-gated → AE 根本不激活表达式（`expressionEnabled` 读回 false）。

2. **`Stable` / Verified 有「覆盖边界」，不是无条件。** ship-gate 只验过**特定规模 / 组合**（典型：2 个关键帧、单层、单特性）。**超出边界即退回未验证**——更多关键帧、更大数据、更多层、多特性叠加，都要重新 gate。
   - 教训：transform 关键帧 gate 只验过 **2 个** Position 关键帧；demo 用了 **13 个空间 Position 关键帧**（运动路径）→ AE 打开急切解码畸形运动路径 → 卡死无响应（hang，非 crash）。

3. **「组合 / 端到端」是独立交付项。** 把多个各自 gated 的能力拼成一个真实产物（完整工程、端到端流程），**组合本身**必须有它自己的 ship-gate；单点 gate **不为组合背书**。
   - 教训：从没有任何 gate 测过「从零生成一个多层动画工程」这种端到端用例——所有 gate 都是单特性 / 小规模。

## 这个库擅长 / 不擅长（避免用错方向）

- **擅长（Verified 密集）**：读真实 `.aep` → 局部改某字段（length-preserving）→ 写回；单点结构性写（加一个 mask / 删一个 effect / 移一层）。
- **不擅长 / 未验证**：从零拼复杂工程。from-scratch 创建的层天然缺东西（transform 默认通道被 AE 省略、表达式不认、关键帧规模 / 组合未测）。**拿它当「AE 工程生成器」= 用在最弱、最没验证的方向**——必须为这条路径单独补 gate 才谈得上交付。

## 交付前自检（4 问，任一为否 → 不是交付，是原型）

1. 产物用到的**每个能力**都是 Verified 吗？
2. 都在各自 gate 的**覆盖边界内**（规模 / 组合）吗？
3. 这个「组合 / 端到端」本身**有没有自己的 gate**？没有 → 它是原型，不是交付。
4. 演示 / README / 示例 / 给用户的能力描述里，有没有混进 Go-only 或超边界的东西？有 → 标红「未验证」或移除。

## 与既有规范的关系

- CLAUDE.md #6（AE 接受 gate）= 本准则的「单点结构性写」部分；本准则把它**扩展到组合 / 端到端 / 演示 / 对外描述**。
- CLAUDE.md #2（API 分级 Stable/Alpha）= 验证级别的 API 契约视角；本准则补上「Stable 有覆盖边界」「组合需独立 gate」「演示禁用未验证能力」三条。
