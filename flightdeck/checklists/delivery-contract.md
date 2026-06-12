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

## 四条认知红线（每条都有血的教训）

1. **Go round-trip 0 警告 ≠ AE 接受。** Go parser 宽松：它只验证「字节能读回成模型对象」，**不验证 AE 的几何 / 表达式 / 渲染引擎能否消化**。结构自洽 ≠ AE 认账。
   - 教训：`SetExpression` 只做过 Go round-trip（合成 fixture），从没 AE-gated → AE 根本不激活表达式（`expressionEnabled` 读回 false）。

2. **`Stable` / Verified 有「覆盖边界」，不是无条件。** ship-gate 只验过**特定规模 / 组合**（典型：2 个关键帧、单层、单特性）。**超出边界即退回未验证**——更多关键帧、更大数据、更多层、多特性叠加，都要重新 gate。
   - 教训：transform 关键帧 gate 只验过 **2 个** Position 关键帧；demo 用了 **13 个空间 Position 关键帧**（运动路径）→ AE 打开急切解码畸形运动路径 → 卡死无响应（hang，非 crash）。

3. **「组合 / 端到端」是独立交付项。** 把多个各自 gated 的能力拼成一个真实产物（完整工程、端到端流程），**组合本身**必须有它自己的 ship-gate；单点 gate **不为组合背书**。
   - 教训：从没有任何 gate 测过「从零生成一个多层动画工程」这种端到端用例——所有 gate 都是单特性 / 小规模。

4. **「值 round-trip 绿」≠「渲染正确」。** 一个 gate「绿」只代表它**验到的那部分**对，不代表能力真能用——如果 gate 没验到能力的**作用面**（见下章）。`AE 接受 + 值读回相等` 只证明「AE 存住了正确的数」，**不证明 AE 渲染时用了这个数**。
   - 教训：shape fill/stroke/gradient 颜色的 gate 只验「值 round-trip + 不丢层」，从没验渲染。orbit demo 里颜色值全部存对（probe 证实 ARGB 正确）、gate 全绿，但 AE **渲染成统一默认红**——「假绿」整整覆盖了「颜色能用」这个最关键的点。

## 验证必须匹配能力的「作用面」（gate 验什么，不只是 gate 跑没跑）

铁律规定「AE 双版本 gate PASS 才算数」，但**没规定 gate 必须验证什么**。补上：**gate 的验收内容必须覆盖能力的「作用面」**——即用户最终**看到 / 用到**的那一面。

| 能力的作用面 | gate 必须验到 | 只验值 round-trip = |
|---|---|---|
| **视觉渲染**（fill/stroke/gradient 颜色、opacity、blur、blend mode、可见几何的位置/形状/缩放/旋转的实际呈现） | **渲染像素**：让 AE render 一帧 → 采样像素颜色 / 位置，或与参考帧比对 | ❌ 假绿，不算 Verified |
| **数据 / 结构**（chunk 被 AE 接受、不丢层、ID、数值字段被保留） | AE 接受 + 值 round-trip | ✅ 足够 |
| **纯元数据**（name / comment 等不影响渲染） | 值 round-trip | ✅ 足够 |

判定法：问「这个能力，用户最终**看到 / 用到**的是什么？」——若是**渲染出来的画面**，gate 必须验到**像素级**；只验「文件里存了正确的值」远远不够。

> 渲染验证的实现：AE 脚本 render 一帧（`comp.saveFrameToPng` / RQ 输出 PNG）→ Go 读 PNG 采样目标像素颜色 / 位置 → 断言。这是 ship-gate harness 的标准扩展，凡渲染类能力都该走这条。

## 这个库擅长 / 不擅长（避免用错方向）

- **擅长（Verified 密集）**：读真实 `.aep` → 局部改某字段（length-preserving）→ 写回；单点结构性写（加一个 mask / 删一个 effect / 移一层）。
- **不擅长 / 未验证**：从零拼复杂工程。from-scratch 创建的层天然缺东西（transform 默认通道被 AE 省略、表达式不认、关键帧规模 / 组合未测）。**拿它当「AE 工程生成器」= 用在最弱、最没验证的方向**——必须为这条路径单独补 gate 才谈得上交付。

## 交付前自检（5 问，任一为否 → 不是交付，是原型）

1. 产物用到的**每个能力**都是 Verified 吗？
2. 都在各自 gate 的**覆盖边界内**（规模 / 组合）吗？
3. 这个「组合 / 端到端」本身**有没有自己的 gate**？没有 → 它是原型，不是交付。
4. 演示 / README / 示例 / 给用户的能力描述里，有没有混进 Go-only 或超边界的东西？有 → 标红「未验证」或移除。
5. 用到的**渲染类能力**（颜色 / opacity / 可见效果），它的 gate **验了渲染像素**吗？只验了值 round-trip → 渲染未验证，不算 Verified。

## 与既有规范的关系

- CLAUDE.md #6（AE 接受 gate）= 本准则的「单点结构性写」部分；本准则把它**扩展到组合 / 端到端 / 演示 / 对外描述**。
- CLAUDE.md #2（API 分级 Stable/Alpha）= 验证级别的 API 契约视角；本准则补上「Stable 有覆盖边界」「组合需独立 gate」「演示禁用未验证能力」三条。
