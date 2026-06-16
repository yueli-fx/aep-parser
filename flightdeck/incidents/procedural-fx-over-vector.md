---
status: active
when_to_read: 被要求从零生成火焰/烟/能量/光效类风格化视觉（火焰 demo 番外重启等）;想用矢量形状+模糊糊某种自然现象前
applies_to: [showcase, fire, smoke, energy, procedural-fx, fractal-noise, animate-effect-param, vector-shape, flame-番外]
last_updated: 2026-06-16
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

## How to apply

被要求从零生成此类视觉时:
1. 默认走**程序化效果链**,别用纯矢量+模糊糊形状。
2. 动画优先用 `AnimateEffectParam`/`AnimateEffectParamVec`(已 ship + 双版本 gate)keyframe **效果参数**(如 Fractal Noise Evolution),而非 path/position 关键帧。
3. 若所需效果(Fractal Noise 参数控制 / Turbulent Displace / Glow 渲染、Colorama/Displacement Map 模板)尚未 render-gate 验证或缺嵌入模板,**先如实说这是未验证/缺失的基础设施缺口**(查 `go run ./cmd/capindex -q`),别硬糊一个交付(交付准则,CLAUDE.md #7)。

关联:CLAUDE.md 工作风格「渲染类 bug 先看图」「做完一个方向」· `effect-param-elision-synthesis-lite.md`(AnimateEffectParam 机制)。
