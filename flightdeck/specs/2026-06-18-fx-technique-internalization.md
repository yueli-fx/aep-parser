---
status: active
graduate: true
summary: 通用机制:任何参考 .aep → 解析 → 拆角色 → 抽跨域技法原子 → 存 flightdeck 两层结构(技法库 references/fx-techniques + 现象配方 checklists/build-X)→ AE gate 验证。火焰=实例#1已验证。适配风/雨/雷电/转场=复用技法+增量。是 procedural-fx-generator 的「配方提取」引擎的通用化。
last_updated: 2026-06-18
implements: specs/2026-06-18-procedural-fx-generator.md
---

# 技法内化流水线 + 跨域技法库（参考模版 → 可复用能力）

## 问题（眼光放长远）

不要只会做火焰。真正的资产是:**用户给一个参考 `.aep` 模版,我能系统地把它「内化」成可复用、可迁移到其它现象(风/雨/雷电/场景切换/转场/动画)的能力。** 本 spec 定义这套内化机制 = 流水线 + 知识结构 + 跨会话持久化。火焰 arc(2026-06-18)是这套机制的首个实证。

## 核心洞察:资产是「技法」不是「现象」

**每种视觉现象 = 一组跨域「技法(technique)」的组合;技法可复用,现象只是配方。** 火焰教会的技法大半不属于火焰:噪声造质料、位移扭曲、多层 Add 叠深度、亮度→调色板、粒子发射、时间驱动参数、辉光……这些直接复用于烟/风/水/闪电/能量/转场。**所以新现象 = 复用已有技法 + 增量补新技法;技法库越长越厚,新现象越来越便宜。**

（承接上一轮用户定调:「层次感来自多层合成,不是参数」——即"多层 Add 叠深度"是一个技法,不是一个旋钮。技法是结构/手法级,不是参数级。）

## 内化流水线（5 步;每给一个参考模版跑一遍）

1. **PARSE（机械,库已能做）**:参考 .aep → 全结构 dump(comp 树 / 每层 类型·混合模式·源·mask / 每效果 matchName+set/animated 参数 / 关键帧·表达式 / 原生vs插件分类)。
   - 当前:探针 `tmp_debug/dump_fxchain`+`dump_tdmn`(local)。**TODO:升级成正经工具 `cmd/aepdissect`**,任何模版一键出结构化报告。
2. **DECOMPOSE 到角色（AI/人判断,不可机械化）**:原始结构 → 功能角色映射(「什么效果干什么用」)。产物 = 角色词典(见 `checklists/build-good-fire.md` § 效果用途词典)。
3. **EXTRACT 技法（提炼跨域原子）**:每个技法抽成命名条目。**新技法 → 进技法库;已有技法 → 标"现象 Y 也用到"**(交叉链,越用越强)。
4. **STORE（flightdeck,自动跨会话）**:两层结构,见下。
5. **VALIDATE（库已有 gate）**:按配方 build(`NewSolidLayer`×N + `AddEffect` + `SetEffectParam`/`AnimateEffectParam` + `Layer.SetBlendingMode` + `SetEffectLayerParam` 等,全 gated)→ AE 渲 → 眼验/用户验收。同一套火焰 render harness 通用。只有验证过的技法/配方标「validated」。

## 知识结构（关键设计:两层）

```
references/fx-techniques/        ← 技法原子(跨域可复用,按角色索引)
   noise-substance.md  displacement-distortion.md  additive-depth.md
   luminance-color.md  particle-emit.md  time-evolution.md  emissive-glow.md ...
   每条:做什么 · 哪些效果实现(matchName) · 参数按观感 · 哪些现象用它 · 原生/插件 · 验证状态

checklists/build-<现象>.md       ← 现象配方(组合技法,一句话引用一堆技法)
   build-good-fire.md (实例#1,已建)  → 后续 build-wind / build-rain / build-lightning / build-transition-*
```

- 两层都在 flightdeck → **preflight 每次完整载入 checklists/INDEX(+ docs/references INDEX)→ 下个会话开场自动浮现**。这就是"内化成我的"的确切机制:非记忆/非重训,是**结构化、自动浮现、随模版单调增长**的知识库。
- 注:`references/` 当前 gitignore(本地)。技法库要跨 clone 持久 → **技法原子建议落 `docs/`(tracked)或 checklists/,不落 references/**(实施时定;不改本 spec 结论)。

## 技法 → 跨现象复用（证明通用性,非火焰专属）

| 技法原子 | 火焰 | 风 | 雨 | 雷电 | 转场 |
|---|---|---|---|---|---|
| 噪声造质料 (Fractal Noise) | 火料 | 气流 | — | — | 噪声擦除 |
| 位移扭曲 (Displacement/Turbulent) | 火舌 | 吹弯/热浪 | 雨幕扰动 | — | 扭曲转场 |
| 多层 Add 叠深度 | 热芯/层次 | — | — | 辉光叠加 | — |
| 亮度→调色板 (Tritone/Ramp) | 色温 | — | — | 电色 | 调色 |
| 粒子发射 (CC Particle World) | 火星 | 飘尘 | 雨滴/雪 | 火花 | 碎屑转场 |
| 时间驱动参数 (关键帧) | 翻腾/上升 | 风向流动 | 下落 | 闪烁 | 时间扫过 |
| 辉光 (Glow) | 泛光 | — | — | 电弧光 | 光效转场 |
| 分形分支 (Advanced Lightning) | — | — | — | 主体 | — |
| 擦除/形状过渡 (Linear Wipe/CC) | — | — | — | — | 切换主体 |

→ 雨 ≈ 粒子发射(条状)+时间下落+模糊;风 ≈ 噪声造质料+方向位移(作用于目标层)+运动模糊;雷电 ≈ 分形分支+辉光+闪烁+Add;转场 ≈ 位移/擦除+时间。**大量复用,少量增量。**

## 现状与下一步

- **已验证(火焰 arc)**:流水线 1–4 跑通过一遍——解析真实多插件工程 → 角色词典 → 原生/插件分级 → 存 flightdeck。证机制可行。
- **待做(把火焰专属升级成通用)**:
  1. 把探针升级为 `cmd/aepdissect`(PARSE 工具化)。
  2. 把 `build-good-fire.md` 里的技法抽成独立技法原子文件(建立 `fx-techniques/` 首批),建立两层结构。
  3. 火焰 v3(多层合成)作为"additive-depth + 多层"技法的验证实例。
  4. 喂第二个现象(风/雨/雷电其一)验证技法复用率。

## 与父 spec 的关系

本 spec = `procedural-fx-generator` 的「离线配方提取」半边的**通用化引擎**。父 spec 管产品形态(NL→.aep 网站);本 spec 管"知识怎么从模版进到库、且跨域复用"。运行时(NL→参数→出字节)仍属父 spec。

## graduate

标 `graduate: true`:本 spec 定义了**知识结构契约**(技法库两层结构 + 内化流水线)——后续每个现象、每次喂模版都按它走,会被反复参考。用户可否决。
