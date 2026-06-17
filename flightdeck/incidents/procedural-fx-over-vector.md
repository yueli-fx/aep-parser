---
status: active
when_to_read: 被要求从零生成火焰/烟/能量/光效类风格化视觉（火焰 demo 番外重启等）;想用矢量形状+模糊糊某种自然现象前;手搓程序化效果链配方想直接宣称「能用/好看」前
applies_to: [showcase, fire, smoke, energy, procedural-fx, fractal-noise, animate-effect-param, vector-shape, flame-番外]
recurrences: 1
last_updated: 2026-06-18
---

# 生成火焰/烟/能量类视觉：程序化效果链 > 矢量形状+模糊

## Signature
- symptom: 生成的火焰/烟/能量视觉被用户判定「很丑 / 像玩具 / 只是形态糊了上去 / 动画是简单位移这也叫火焰动画?」
- error_type: —
- where: showcase / 从零生成自然现象类风格化视觉
- trigger: 用「矢量形状 + 渐变 + 高斯模糊 + 混合模式」拼火焰/烟/能量,动画用 path/position 关键帧

## 经过

2026-06-16 用户让从零生成"类似参考图"的动画火焰。先后两版都用「矢量 teardrop 形状 + 渐变 + 高斯模糊 + Add 混合」拼——**两版都被否**(丑/玩具/形态糊上去/动画只是位移),用户指出参考火焰明显是**各种效果堆叠**出来的。火焰遂定为**番外**(后续再做)。

## Why（要害）

真实/风格化的火焰·烟·能量本质是**程序化(generative)效果**,不是矢量插画:
- **形态靠噪声纹理生成,不是画形状再糊模糊**——糊模糊只得到"软边玩具"。
- **动画来自效果参数随时间演化**(Evolution / Offset),不是移动/变形形状。

经典 AE 从零做火焰套路:`Fractal Noise(Evolution 关键帧)→ 颜色映射(Colorama / Gradient Ramp)→ Turbulent Displace(舔动)→ Glow`。用户对"用对技术"的敏感度高于"形似"——糊个大概形状不算数。

## [Case 2] 2026-06-18 — 用对了程序化效果链,手搓配方仍被否（命门质量未过）

火焰 demo 番外重启,这次**严格按本 incident 的 How to apply 走**:固态层 + Fractal Noise(竖向拉伸+高对比)→ Tint(黑→暗红/白→橙黄)→ Turbulent Displace(有机扰动)+ 羽化水滴 mask + `AnimateEffectParam` 给 Evolution 打关键帧(翻腾)。技术全对、双版本 AE ship-gate 绿(`TestFlameDemo_AEShipGate_AE2020/AE2025`,两版本渲染帧逐像素一致)、`AnimateEffectParam` 给 elided Evolution 打关键帧没撞上已知 elision 缺口。

**但用户真机验收仍一票否决**:「只有火焰的形态,但和火焰差很多」。产物=橙色噪声水滴,缺真火关键——**白热的芯 / 由内到外色温渐变(白→黄→橙→红→暗尖)/ Glow 泛光 / 向上舔的细节**。

**升级要害**:不是"技术选错"(那是 Case 1 的教训),而是**手搓配方只到「可辨认」、到不了「好」**。用对程序化效果链 ≠ 能产出专业质量;**好视觉的效果栈/参数/分层必须从真实人做的样本里学**,凭空调参只能到玩具档。决策:roadmap 改顺序——**先学真实样本(解析专业火焰 .aep 抽效果栈+参数)再回头重做,过用户关后才参数化**。详 `archive/plans/2026-06-18-phase0-flame-deterministic.md` § 用户验收结论 + `specs/2026-06-18-procedural-fx-generator.md` § 更新(2026-06-18)。

## How to apply

被要求从零生成此类视觉时:
1. 默认走**程序化效果链**,别用纯矢量+模糊糊形状。
2. 动画优先用 `AnimateEffectParam`/`AnimateEffectParamVec`(已 ship + 双版本 gate)keyframe **效果参数**(如 Fractal Noise Evolution),而非 path/position 关键帧。
3. 若所需效果(Fractal Noise 参数控制 / Turbulent Displace / Glow 渲染、Colorama/Displacement Map 模板)尚未 render-gate 验证或缺嵌入模板,**先如实说这是未验证/缺失的基础设施缺口**(查 `go run ./cmd/capindex -q`),别硬糊一个交付(交付准则,CLAUDE.md #7)。
4. **(Case 2 教训)用对效果链 ≠ 能产出「好」——别拿手搓配方直接宣称交付**。手搓凭空调参的天花板是「可辨认」(一眼是火苗轮廓),离专业质量(白热芯 / 色温渐变 / 泛光 / 舔动细节)有质的差距。**好火焰/烟/能量的效果栈+参数+分层要从真实人做的样本里学**(解析专业 .aep 抽配方),不能凭空猜。顺序:**先有样本参照造出「好」的,再谈参数化/泛化**。

5. **(配方真相源)造火焰/烟/能量前先读 `checklists/build-good-fire.md`** —— plugin-free 原生配方 + 参数→效果对照表 + 插件分级，从真实样本解析得出，省得重新摸索。

关联:CLAUDE.md 工作风格「渲染类 bug 先看图」「做完一个方向」· `effect-param-elision-synthesis-lite.md`(AnimateEffectParam 机制) · `checklists/build-good-fire.md`(火焰配方)。
