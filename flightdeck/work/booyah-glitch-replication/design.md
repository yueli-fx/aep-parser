# Booyah Glitch 全工程复刻 — 理解金标准检验

## 1. 背景 / 为什么

「理解一个工程」研究 arc 的金标准:能不能用咱们的 Go API **从零重搭出一个真实工程**。round-trip 只证「读得回字节」;**从零重建证「真的懂每个 chunk 怎么来的」**。目标工程 = `data/samples/motionbox/glitch/booyah-glitch/Booyah Glitch.aep`(用户指定;我们 glitch showcase 就是照它拆的,先验理解最足、全 native 无插件)。

用户铁律:**不逃避未知字段**——撞到库不支持的 effect 参数 / chunk / 编码,走 RE 流水线实现,而不是 skip 或近似糊弄;物理不可写的(无 scripting API)给实证负结论。

## 2. 目标工程真实规模(aepdissect 实测)

- **12 个合成**,全 1920×1080 / 30fps / 6.0s,4+ 层嵌套。
- **61 层**:34 precomp 实例(av)、22 调整层、4 shape、1 text(日文)。
- **~100 mask**:glitch 撕裂切片是 mask 定义的(グリッチテキスト 调整层每层 9–21 个 mask 路径)。
- **54 条动画属性**、42 effect 实例、99 非默认参数。
- effect 全原生(11 种):Displacement Map×12 · Transform(Geometry2)×7 · Glow(Glo2)×5 · Fractal Noise×4 · Fill×3 · Gaussian Blur 2×3 · Venetian Blinds×3 · Exposure2×2 · CurvesCustom×1 · Noise2×1 · Ramp×1。RENDER DEPENDENCIES = none(stock AE 可渲)。

## 3. 成功判据(用户拍板:结构保真 + 终帧渲染对照)

逐 comp 三关:
1. **AE 接受** —— 打开不报「项目文件似乎已损坏」。
2. **结构保真** —— `verify.jsx` dump 重建 comp 的 DOM 值(层数/类型/transform/effect 参数/mask 数/关键帧),与原工程**同 comp 解析值逐项对账**。
3. **终帧渲染对照** —— 顶层 `メインコンプ！` render 一帧,与原工程同帧**像素对照**(红线4:值对≠渲染对)。

凡**新 RE 出来的写路径**(库当前没有/没 ship-gate 的),必须过 **AE 2020 + 2025 双版本 ship-gate** 才算 ship(红线6)。

## 4. 方法(用户拍板:方法 A)

**手写 per-comp 生成器,原工程仅当「取值神谕」,chunk 全由咱们 API 重建。**

- 值从解析原工程读出来(什么 mask 路径 / 什么参数 / 什么关键帧时间),但输出字节**全部经 `NewProject`/`New*`/`AddEffect`/`SetEffectParam`/`AddMask`/`SetMaskPath(Keyframes)`/`Animate*` 构造**。
- **关键纪律(这次检验的意义)**:原工程**只读、当值的神谕**,绝不当字节源 copy。copy 原 chunk 字节 = 退化成 round-trip,不算复刻,本 spec 视为失败。
- 按需长出可复用的小提取器(批量 copy mask / copy keyframes from parsed layer),先住 showcase 包;证明通用了再提升 `tools/replicate/`——这些提取器是未来全自动 transcoder 的种子,增量喂「理解自动化」、不担前置基建风险。

## 5. 工件落位

`flightdeck/showcase/booyah-clone/`(对齐 CLAUDE.md「showcase 生成器」工件规则 + showcase 眼验面):
- `gen.go` + 按 comp 拆的 `gen_<comp>.go`(同一 package main)
- `render.jsx`(saveFrameToPng 终帧)+ `verify.jsx`(dump DOM 值对账)
- `INDEX.md`(frontmatter + **覆盖账本**,见 §8)

原工程留 `data/samples/`。`*.aep`/`*.png` gitignored,`gen.go`/`*.jsx`/`INDEX.md` tracked,干净 clone 可一键重生成。

## 6. DAG 拓扑序(叶→根,逐 comp 单独 AE 验过再上层)

| # | comp(id) | 层 | 依赖 |
|---|---|---|---|
| ① | シェイイイイプ！！！(173) | 1 | — |
| ② | テキスト変えるならココ！(27) | 1 text | — |
| ③ | マップ用フラクタルノイズ(69) | 2 | — |
| ④ | カクッ(337) | 2 shape | — |
| ⑤ | プリコンポジション 1(205) | 3 | ① |
| ⑥ | シェイプの塊(243) | 7 | ⑤ |
| ⑦ | ここは開けない方が身のため(287) | 3 | ⑥ |
| ⑧ | RGBズレ(132) | 3 | ② |
| ⑨ | なんか周りのやつ(355) | 3 | ④ |
| ⑩ | **グリッチテキスト(85)** | **27 / ~100 mask** | ⑦②⑧③ ← 怪物 |
| ⑪ | 背景変えるならココ！(43) | 6 | — |
| ⑫ | **メインコンプ！(14)** | 3 | ⑩⑨⑪ ← 顶 |

## 7. 未知字段前沿(撞上 RE,不绕;每个给 ship-gate 或实证 blocked)

- **wiggle 表达式**(头号):`Exposure` 层挂 `wiggle(34,0.29)`。`SetExpression` 是已知 false-green 高危区(`incidents/expression-enable-byte-pair.md`:Go round-trip 绿 ≠ AE 真求值)。**先做最小验证**:Go 写一个 wiggle 表达式 → AE 是否真求值(DOM expressionEnabled + 帧间值变化 / render-pixel)。
- **go/no-go(用户已拍板)**:若实证 AE 不认 Go 写的表达式 → 该属性**降级为「值 + keyframe 复刻、表达式标 blocked + 实证理由」**可接受,不必死磕到 AE 求值为止。
- **Curves(CurvesCustom)曲线数据**:arbitrary-data param(dissect 都没 surface),RE 曲线点编码 + 能否 SetEffectParam 物化。
- **单层 ~21 mask**:压测 mask parade 多 mask 写路径(`AddMask` gate 只验过少数 mask);确认多 mask + 各自 path 不被 silent-drop。
- **shape 几何(4 层)/ text(日文 + 字体)/ Gradient Ramp 颜色+点**。

## 8. 覆盖账本(诚实标 + blocked 给实证)

`showcase/booyah-clone/INDEX.md` 维护逐特征状态表:**已复刻 / 待 RE / 物理 blocked**。blocked 必须给实证理由(对齐交付准则 + 红线7:未验证禁混入「能用」)。终帧渲染对照前,任何「值 round-trip 绿但未 AE 验」的项标黄,不计入复刻完成度。

## 9. 知识沉淀

新坑 → `incidents/`(双版本 gate 失败信号、RE 发现);新能力 → 源码 `aep:cap` tag(capindex);本 arc 进度 → 本 spec + 对应 plan。预期会催生若干新 incident(多 mask 写、wiggle 表达式实证、curves 编码等)——这正是「深化学习」的产出。

## 10. 风险 / 已知坑

- **grad テキスト怪物**(⑩,27 层 + ~100 mask + 11 个 displacement)是主要工作量与风险集中点;叶子优先正是为了在抵达它之前把 mask/displacement/keyframe 写路径都验顺。
- **wiggle 表达式**可能整段 blocked(见 §7),提前承认、不阻塞主线。
- 调整层 `src=footage(N)`(footage 100/110/169/319/376)预计是 solid footage(adjustment 层底层) —— **首个 comp 即确认**这些 footage 是 solid 还是需额外 RE。
- 多 mask / 多 keyframe 触 lhd3 capacity-pages(`incidents/lhd3-keyframe-capacity-pages.md`)——已知机制,留意分页正确。

## 11. 用户已决策(brainstorming 定档)

- 判据 = **结构保真 + 终帧渲染对照**(非纯渲染近似、非字节等同)。
- 范围 = **全工程 12 comp,DAG 叶子优先逐 comp 验**。
- 方法 = **A(手写 per-comp 生成器,原工程=取值神谕)**。
- 工件落 `showcase/booyah-clone/`;wiggle 实证 blocked 时**降级可接受**。
