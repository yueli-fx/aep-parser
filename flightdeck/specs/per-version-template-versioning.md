---
status: idea
summary: 为 effect/shape body 跨 AE 版本分裂准备的升级响应机制(YAGNI,gate 红了再建):局部字段分裂走版本条件 CODE(先例 ldta 160/164),整份 body 分裂把 effectTemplateFiles 升 (name,target)→path + per-target 覆盖文件 fallback 默认。检测靠现有双版本 ship-gate(红线6)。
---

# Per-version embedded-template versioning scheme

> **下个会话起。这是「准备好但不预建」的响应方案**——只有当某个具体 effect/shape 模板真的在某版本 gate 红了才动手。本 spec 是那一刻的现成蓝图。背景调查见 `docs/embed-template-architecture.md` § 版本可移植性。

## 背景 / 动机

库的写策略 = 「parse-the-clone + 覆写值」:effect/shape 的复杂 opaque body 不是从零生成,而是嵌入一份 AE 亲手存的 canonical 字节(`.bin`)再覆写看懂的值槽(详 `docs/embed-template-architecture.md`)。

**当前所有 effect/shape body 都是从 AE 2020 抽的单一模板**(`mutate_effect_add.go` 注释:"extracted verbatim from an AE-2020-saved layer … version-portable, AE 2025 accepts the AE-2020 bytes")。可移植性不是假设,是**双版本 ship-gate(红线6)持续实测**:`TestAddEffect_AEShipGate_AE2020` + `_AE2025` 把整库灌进两个真 AE,断言都接受+读回正确。

**风险(用户 2026-06-17 提出)**:AE 升级可能改某效果/形状的初始字段,使 AE-2020 body 在新版本里被丢/读错。目前**没有 per-version 模板机制**——`effectTemplateFiles` 是 `map[string]string`(name→path 单值);shape body 是单 `//go:embed` var。**唯一按版本选的是工程 seed**(`mutate_project_new.go` `case TargetAE2025` → `templates/project/2025.aep`)。

## 目标 / 非目标

**目标**:把「某模板跨版本分裂时怎么办」从口口相传变成代码就绪的响应路径,使红线6 gate 一旦报警就能最小改动修复。

**非目标(关键)**:
- **不预先建任何 per-version 模板**(YAGNI)。目前零个模板在 2025 gate 失败,无实证收益,2–3× 维护成本不值。
- 不改双版本 gate 机制本身(它已是探测器)。
- 不追求 Adobe 兼容性"保证"——可移植性永远是实测(bookend gate)非承诺。

## 触发条件(什么时候执行本 spec)

某个具体 effect/shape 的双版本 ship-gate **在某 AE 版本上红了**,且根因 = AE-2020 body 的字节结构跟该版本 canonical 不一致(被丢 / 读错值 / 判损坏)。**先做最小失败 bisection 确认是版本分裂**(对照 known-good fixture),再按下面分级响应。

## 设计:两级响应

### Tier A — 局部字段分裂 → 分叉点写版本条件 CODE
若只是 body 里某几个字节/某字段长度按版本不同(不是整份结构变),在 lower/back 的那个写点加 `target` 条件分支。**已有先例,不是新机制**(完整 postmortem: [`../incidents/ae2020-shape-ldta-164-corrupt.md`](../incidents/ae2020-shape-ldta-164-corrupt.md)):
- `lower_layer.go` / `mutate_layer_camera.go`:ldta 按 target 补到 160B(AE2020/22)或 164B(AE2025),尺寸走 `ctx.capabilities.LdtaSize`(capability matrix)勿硬编码。
- 这是首选——成本最低,模板仍单一。

### Tier B — 整份 body 分裂 → 版本键化模板加载
若整份 body 结构在新版本不可调和(Tier A 补不动),才上 per-version 模板:
1. **模板键升维**:`effectTemplateFiles` 从 `map[string]string` 改 `map[string]map[AETarget]string`(或等价的 `(name,target)→path` 查找),`cloneEffectTemplate` 按 target 选;**fallback 到默认(2020)** 当该 target 无覆盖。shape body 同理(per-shape 版本变体)。
2. **覆盖文件命名**:默认 `effects/<name>.bin`,覆盖 `effects/<name>.<ver>.bin`(如 `effects/glow.2025.bin`)。**刚做的子目录重构让这点很干净**——覆盖文件就在同目录,一眼可见哪些效果版本化了。
3. **镜像 seed 模式**:这跟 `mutate_project_new.go` 按 target 选 seed 是同一套思路,实现风格对齐。
4. **抽取**:per-version 覆盖的 `.bin` 由对应 AE 版本跑 RE fixture 抽出(`extract_effect_lib` / `extract_shape_bodies` 已参数化输出路径)。

## 开放问题

- **中间版本 gate**:现在只 gate **bookend(2020 + 2025)**,假设 AE 向后兼容则 2021/2023/2024 稳(seeds 虽全有)。是否需要给关键能力补中间版本 render-gate?——按需,等出现中间版本独有的分裂证据再说。
- **粒度**:版本键化是 per-template(只给分裂的那个加变体)还是 per-domain?倾向 per-template(最小面)。

## 验收标准(执行时)

- [ ] 触发的 gate 失败已 bisection 确认为版本分裂(非 flake / 非其它 bug)。
- [ ] 按 Tier A 或 B 修复;若 Tier B,`cloneEffectTemplate`/shape loader 的 fallback 路径有测试覆盖(默认+覆盖两条)。
- [ ] 修复后**双版本 gate 全绿**(含新加的覆盖路径)。
- [ ] 本 spec 的机制描述对齐实际实现(graduate 候选:它定义了 per-version 加载契约)。

## 评审纪要

（空）
