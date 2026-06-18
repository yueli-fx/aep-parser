# specs/ — INDEX

<!-- AUTO:specs -->
### Backlog (idea)
- [per-version-template-versioning.md](per-version-template-versioning.md) — idea — 为 effect/shape body 跨 AE 版本分裂准备的升级响应机制(YAGNI,gate 红了再建):局部字段分裂走版本条件 CODE(先例 ldta 160/164),整份 body 分裂把 effectTemplateFiles 升 (name,target)→path + per-target 覆盖文件 fallback 默认。检测靠现有双版本 ship-gate(红线6)。

### Active · Done
- [2026-06-19-booyah-glitch-full-replication.md](2026-06-19-booyah-glitch-full-replication.md) — active — 把整个 Booyah Glitch 真实工程(12 comp/61 层/~100 mask/wiggle 表达式/Curves)用咱们的 Go API 从零完整复刻,作为理解金标准检验。判据=结构保真(AE 接受 + DOM 读回值对账)+ 终帧渲染像素对照;范围=全工程 DAG 叶子优先逐 comp 验;方法=手写 per-comp 生成器、原工程仅当取值神谕、chunk 全由 API 重建(绝不 copy 字节)。
- [2026-06-18-technique-ontology.md](2026-06-18-technique-ontology.md) — active — 投大规模工程前冻结的数据骨架:角色/技法/机制三轴本体 + 技法条目 schema v2(等价效果集·可复刻性两轴·迁移分proven/hypothesized)+ 工程画像 schema + 受控词表 v0。schema 闭 instance 开,约束所有工程理解的产出形式,使跨工程可聚合出通用技巧。经火焰(生成型)/闪电(素材装配型)/控制器(lib-blocked)三类压测。
- [2026-06-18-roundtrip-ae-accept-residual.md](2026-06-18-roundtrip-ae-accept-residual.md) — active — 补验 arc(批1-28,2026-06-17/18)完成记录 + 剩 53 个 verify=roundtrip 残值的 someday-backlog:逐项分类(N/A read/helper · 实勘负结论 · false-green · 硬尾 · 可做但低 ROI),有需要时按本表挑。ae-accept 35→261,roundtrip 279→53。
- [2026-06-18-procedural-fx-generator.md](2026-06-18-procedural-fx-generator.md) — active — 用户 NL 描述 → 网站一键产出可在 AE 打开的 .aep。架构=离线配方提取 + 运行时(NL→参数→库出字节)。AI 永不碰字节,Go 库是保证合法的执行引擎。v1=火焰,Phase 0(确定性造一个好火焰)为 make-or-break 门槛。
- [2026-06-18-fx-technique-internalization.md](2026-06-18-fx-technique-internalization.md) — active — 通用机制:任何参考 .aep → 解析 → 拆角色 → 抽跨域技法原子 → 存 flightdeck 两层结构(技法库 references/fx-techniques + 现象配方 checklists/build-X)→ AE gate 验证。火焰=实例#1已验证。适配风/雨/雷电/转场=复用技法+增量。是 procedural-fx-generator 的「配方提取」引擎的通用化。
<!-- /AUTO -->
