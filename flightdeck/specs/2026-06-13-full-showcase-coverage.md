---
status: active
summary: 把所有已 ship 能力补齐 showcase 供用户逐个真机验收（代码过≠效果对）；已回填 8 个 from-scratch 可视方向，列出剩余缺口 + 非可视能力的验证方式，交接下个对话逐方向补
last_updated: 2026-06-13
---

# 全能力 showcase 覆盖 — 每个已 ship 能力都要有用户可真机验证的 showcase

## 目标（一句话）

**库里每个已 ship 的能力，都要有一个用户能在真机打开 .aep 逐个验证的 showcase**。理由：**「代码过 ≠ 效果对」**（交付准则红线4 + 用户原话）——ship-gate 像素验证是 agent 跑的，用户要亲自开 .aep 核一遍才算数。规约见 `rules.md` § Showcase + `checklists/showcase.md`（含 `待review`↔`complete` review-gate）。

## 现状：已回填 8 方向（全部 🔍 待review，等用户真机验）

`flightdeck/showcase/` 下：shape-filters · shape-primitives · keyframes-ease · expressions · precomp-nesting · gradient · text · layers。每个有 `INDEX.md`（布局表）+ `gen.go` + `render.jsx`，产物 `.aep/.png` 已本地生成（gitignore）。**这 8 个等用户逐个真机验收后翻 complete。**

## 缺口：还没 showcase 的已 ship 能力

> 权威清单以 `plans/coverage.md` 为准（看板可能漂移，建时用 grep/Explore 核实代码 + ship-gate test 真在）。下面是分组待办，**逐方向补、每个落 `待review`**。

### A. 可像素验证（同既有 8 个的 from-scratch + AE 渲染套路）

| 方向 | 覆盖能力 | 渲染思路 |
|---|---|---|
| `masks` | AddMask（mask 形状裁切/显隐图层） | 一个填充层 + mask → 渲染只露 mask 内区域 |
| `effects` | AddEffect（Gaussian Blur / Levels / Tint 等）+ 参数 | 同一图形 before/after 并排，看效果差异 |
| `structural-ops` | DeleteLayer / DuplicateLayer / MoveLayer / 属性 Remove·Duplicate·MoveTo / SetDimensionsSeparated | 建基线 → 应用 op → 渲染出结果（如 duplicate→两份、separate→XY 分离动画） |
| `stroke-detail` | 虚线 dashes / Line Cap / Line Join / Miter | 几条不同端点·连接·虚线的描边并排（可并入 shape-primitives 或单列） |
| `animated-path` | animated shape path / mask path 关键帧 | 渲中间帧看路径插值形态 |

### B. 非可视 / 结构性（不出像素，改用 readback 验证产物）

这些能力**没有可渲染的视觉**（EG 面板控件 / 标记 / 渲染队列 / 相机灯光参数 / 媒体替换 / 合成设置等）。其 showcase = `gen.go` 建工程 + `verify.jsx` 把相关 DOM 值 dump 到 `.done` 日志，**用户读日志核值**（而非看图）。INDEX 的「产物」表注明是 readback 验证、不是像素。

| 方向 | 覆盖能力 | readback 验证 |
|---|---|---|
| `essential-graphics` | AddEssentialProperty / SetMotionGraphicsTemplateName | dump EG 控件名/类型/值 |
| `markers` | layer / comp markers | dump marker time/comment |
| `render-queue` | RQ add / insert | dump RQ item 设置 |
| `camera-light` | NewCameraLayer / NewLightLayer + 选项 | dump 层类型 + camera/light 参数（需 3D 才有像素，暂 readback） |
| `comp-settings` | shutter angle/phase 等合成设置 | dump cdta 字段 |
| `media-replace` | SetAlternateSource | 需真实 footage；dump source 引用（或暂搁） |

> B 类是否值得逐个出，按需——用户点名要验哪个就补哪个；纯 round-trip 的可以攒批。

## 做法（交接下个对话）

1. 一次一个方向（A 类优先，视觉最直观）：照 `checklists/showcase.md` 6+1 步——`gen.go`→`go run`→`render.jsx`→AE 实渲眼验→写 INDEX(`待review`)→commit→**通知用户真机复核**。
2. **agent 不自标 complete**；用户真机验过才翻。
3. 每补一个，顶层 `showcase/INDEX.md` 加一行。
4. 全部补齐（A 类 + 用户点名的 B 类）+ 用户逐个验收 complete = 本 spec done。

## 验收口径

每个方向 done 的标准 = **用户在真机打开该方向 .aep、对照 INDEX 布局表确认效果对** → 翻 `complete`。不是 agent 渲染眼验，不是 ship-gate 绿。
