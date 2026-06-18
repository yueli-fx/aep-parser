---
purpose: 跨分类「技法学习」配方区 — 火/烟/能量 · 文字 · 转场 + 跨分类通用技巧（照着做就能产出某现象的可执行配方）
last_updated: 2026-06-18
---

# checklists/techniques/ — INDEX

> **这里装什么**：可执行的「现象配方」与通用技巧流程 —— 照着做就能造出火/雨/雷电/文字动效/转场，以及各分类都能复用的通用做法（多层合成、缓动套路、displacement 等）。学习方向已从「逆向（incidents/）」转向「内化技法」，这是新方向的配方家。
> **不装什么**：跨现象复用的**技法原子**（解释性、读懂为主）住 `docs/`（现 `docs/fx-techniques.md`，长到多篇再镜像出 `docs/techniques/`）；逆向发现 / 踩坑住 `incidents/`；项目流程（commit/verify/showcase/re-fixture）留 `checklists/` 根。
> 两层结构契约见 `specs/2026-06-18-fx-technique-internalization.md`（流水线）+ `specs/2026-06-18-technique-ontology.md`（三轴本体）。
> 首个实例 `build-good-fire.md`（火焰配方）已入区；后续 build-rain / build-lightning / build-text-* / build-transition-* 同此落位。

<!-- AUTO:techniques -->
- [build-good-fire.md](build-good-fire.md) — active — when_to_read: 被要求从零造火焰/烟/能量类程序化 FX；要重做火焰 v2；调火焰参数想知道某个旋钮的视觉影响；纠结哪些效果是核心、哪些靠插件 — applies_to: [fire, flame, procedural-fx, fractal-noise, displacement-map, ramp, glow, recipe, parameters, plugin-free, sample-analysis, AnimateEffectParam, blend-modes]
- [harvest-motionbox-samples.md](harvest-motionbox-samples.md) — active — when_to_read: 想批量采集 motion-box.net（或同类 Nuxt+Firebase 前端）的 AE 参考工程到 samples/；下载按钮无静态链接、点击不触发下载、要找文件直链；firestore REST 取数据被 403 挡；要给 understand-a-project 流水线补充原料工程 — applies_to: [sample-harvest, motionbox, motion-box.net, nuxt, firebase-storage, firestore, vue-state, dlURL, download-reverse-engineer, aep-samples, reference-projects, playwright, rest-403, no-reprint, license]
- [understand-a-project.md](understand-a-project.md) — active — when_to_read: 拿到一个参考 .aep 要「理解/内化」它(抽可复用技法,而非只读懂这一个);要把某现象工程拆成三轴本体(角色/技法/机制);纠结理解的产出该长什么样、存哪;给学习引擎/生成引擎喂新样本前 — applies_to: [internalization, understand-project, pipeline, aepdissect, three-axis-ontology, role-decompose, technique-extract, project-profile, reproducibility, cross-domain, recipe-signal, elision]
<!-- /AUTO -->
