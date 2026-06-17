---
status: active
graduate: true
summary: 用户 NL 描述 → 网站一键产出可在 AE 打开的 .aep。架构=离线配方提取 + 运行时(NL→参数→库出字节)。AI 永不碰字节,Go 库是保证合法的执行引擎。v1=火焰,Phase 0(确定性造一个好火焰)为 make-or-break 门槛。
last_updated: 2026-06-18
---

# AI 参数化程序化 FX 生成器（产品方案 + 路线）

## 一句话

用户用自然语言描述一个程序化视觉效果(「高瘦、偏蓝、闪得快的火焰」)→ 网站一键产出**可在 AE 打开的 `.aep`**。底层 = 从真实样本学到的**参数化配方** + Go 库**确定性出字节**。

## 背景 / 为什么是这个形态(决策依据)

经过 roundtrip→ae-accept 补验 arc(批1-28),这个库被**实证**确认的形状:**极强于「读真实 .aep + 局部改 + 写回」**(全部双版本 ae-accept gated),**极弱于「从零拼复杂工程」**(default-omission、无 property synthesis、表达式 AE 端不求值、规模/组合未测、红线4「值对≠渲染对」)。火焰 demo 当年丑,正是踩在弱项 + 用错技术(矢量+模糊,见 incident `procedural-fx-over-vector.md`)。

排除掉的两条路(brainstorm 结论):
- **❌ fine-tune AI 直接吐 .aep 字节**:.aep 是长度敏感二进制,LLM 无法可靠产出;一字节错 AE 就打不开。「拿样本训练模型吐字节」方法错误,不是数据问题。
- **❌ 纯选模板填参数**:那 ExtendScript 直接能做、且要开 AE,Go 库无差异化价值(用户原话「区别不大」)。

选中的中间路(可行且正中强项):**AI 在「从真实样本学到的参数空间」里插值合成**。火焰的「设计」不在几何里,在于**效果栈 + 参数**(`固态层 + Fractal/Turbulent Noise + Turbulent Displace + 调色 + evolution 动画`)——而库的 effects 域「全收口」(AddEffect + 设参数 ae-accept gated),组装效果栈 + 填参**正是强项**。这既不是选固定模板,也不是从零造任意设计。

## 架构

```
[离线] 配方提取:N 个真实样本.aep ──(库解析,强项)──> 效果栈 + 参数范围 + 哪些在动 ──> recipe schema(JSON)
                                                                                          │
[运行] ① 意图→参数:用户 NL ──(LLM,约束在 schema 参数范围内)──> 一组具体参数
       ② 参数→.aep:recipe + 参数 ──(Go 库:固态层+效果栈+设参+动画,全 ae-accept)──> 合法 .aep
       ③ 交付:网站一键 → 下载 .aep(headless、可规模化)
```

**组件边界(各单一职责、接口清晰):**
- **recipe schema(数据契约)**:一个 FX 族 = 一个 schema = `{效果栈[效果 match-name + 各参数: {范围, 默认, 是否动画, 动画方式}]}`。离线产出、运行时消费。是 AI 层与库层之间**唯一的接口**。
- **recipe→.aep 编译器(Go,库之上的薄层)**:吃 recipe + 具体参数,调库的 NewComposition/NewSolidLayer/AddEffect/设参/打关键帧,产 .aep。**确定性、无 AI、每条路径 ae-accept gated**。
- **意图层(LLM,库之外)**:NL → 受 schema 约束的参数。**永不碰字节**,只产 JSON 参数。
- **配方提取器(离线,Go 解析 + 可选 LLM 辅助)**:解析样本 → 抽共同效果栈 + 参数分布 → 草拟/校准 schema。

## 铁律(刻进契约)

1. **AI 永不碰字节**:AI/LLM 只产「用哪个配方 + 填什么参数」的结构化 JSON;Go 库是保证「一定能在 AE 打开」的安全执行引擎。
2. **每条出字节路径 ae-accept gated**:编译器用的库操作必须是双版本(或单版本破例,见 `checklists/delivery-contract.md`)验过的;新需要的从零能力须补 gate 才算可交付。
3. **红线4 先看图**:FX 是渲染类,Phase 0/1 产物必须 AE 实渲 + Read png 自验 + 用户真机验,不靠值 round-trip 假绿。
4. **诚实边界**(对用户也要说死):质量天花板 = 样本库质量(只在样本张成的参数空间插值,不无中生有);只对「有界 FX 族(设计=效果栈+参数)」成立,**不承诺「任意视频/整场景构图」**;配方里效果须 AE 原生,否则只能从真实文件整层搬运(灵活度低)。

## 分阶段路线(按「先证最大未知」排)

- **Phase 0 — 一个手写火焰能不能成(make-or-break,无 AI、无网站)**
  纯 Go 硬写一条火焰配方(固态 + Fractal/Turbulent Noise + Turbulent Displace + 调色 + evolution 动画)→ AE 实渲 → Read png 自看 + 用户真机验。**整个项目的命门**:库连「确定性造一个像样火焰」都做不到则全盘空中楼阁。**这阶段会逼出并补掉库唯一真实缺口 = 从零给效果参数打关键帧**(incident `effect-param-elision-synthesis-lite`:静态→动画的 InsertKeyframe 有限制)。
  **门槛判定**:Phase 0 渲不出像样火焰 → 当场叫停,换更稳的族(数据条动画 / 文字动效——correctness 客观、动画规整),不浪费后面阶段。
- **Phase 1 — 把火焰参数化**:配方旋钮(颜色/高瘦/闪速/规模)抽成 typed 参数,手调 5-10 个差异明显的好火焰,AE 实渲确认参数空间有意义且都不丑。
- **Phase 2 — 真从样本学配方**:拿 ~10 个真实火焰 .aep,库解析,把手写配方对齐/校准到真实参数分布(验证泛化、修范围)。
- **Phase 3 — 接 AI 意图层**:LLM 把 NL 映射成受约束参数(「一键 + 描述」体验)。
- **Phase 4 — 产品化 + 扩族**:网站/API;加第二个族(数据条 / 文字动效 / smoke / 辉光),复用同一条管线。

## v1 取舍

- **v1 族 = 火焰**(原型场景、wow 最强、技术路径已在 incident 标好),Phase 0 为硬门槛。
- Phase 0 不过 → fallback 到「数据条动画」(correctness 客观、效果栈简单、动画规整),把链路先跑通,火焰退到 Phase 4。

## 成功判据 / 测试

- **Phase 0 通过** = AE 双版本实渲一帧火焰,像素目视像火焰(非 flat/默认色/糊矢量),用户真机确认 → 证「库能确定性出好 FX」。
- **Phase 1 通过** = ≥5 个参数组合各渲出明显不同且都不丑的火焰。
- **Phase 2 通过** = 配方参数范围覆盖 ~10 真实样本(解析对齐误差在容忍内)。
- **Phase 3 通过** = NL 描述 → 参数 → 渲染,人评「描述与结果一致」。
- 每阶段产物进 `flightdeck/showcase/`(渲染类必出 showcase 眼验,rules.md)。

## 更新(2026-06-18,Phase 0 用户验收后 — roadmap 顺序调整)

Phase 0 手搓火焰用户真机否决:「只有形态,和火焰差很多」(缺白热芯 / 色温渐变 / Glow 泛光 / 向上舔细节)。**狭义命门达成**(库能确定性造可辨认火焰,effect-stack+param+AnimateEffectParam+mask 全 gated 可行,双版本渲染一致);**但产品质量门槛未过**。

**坐实的核心判断**:**手搓配方只到「可辨认」、到不了「好」——好视觉必须借真实人造样本。** 这正是 brainstorm 时的论点的实证。

**roadmap 顺序改**:原 Phase 1(参数化)→ Phase 2(学样本)。**改成先 Phase 2(学真实样本)再回头重做火焰,过用户关后才参数化**——没有「好」火焰前参数化无意义。**下一步具体动作 = 用户提供真实火焰 .aep(纯 AE 原生效果、非 Particular 插件、非素材视频)放 `samples/flame/` → 库解析抽「好火焰」真实效果栈+参数 → 重做。** 详 `archive/plans/2026-06-18-phase0-flame-deterministic.md` § 用户验收结论。

## 风险 / 未决

- **最大风险 = Phase 0**(库能否确定性造好火焰 + 补掉从零效果参数动画缺口)。故排在最前。
- 配方提取的「学习」程度:Phase 2 是结构化抽取 + LLM 辅助草拟 schema,**不是** fine-tune;具体抽取算法待 Phase 2 细化。
- 火焰若依赖第三方插件(Particular 等)→ 那批样本只能整层搬运;需在 Phase 2 确认样本用的是 AE 原生效果。
- 网站/API/计费 = Phase 4,本 spec 不展开(独立子项目)。
