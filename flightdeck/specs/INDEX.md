# specs/ — INDEX

<!-- AUTO:specs -->
### Backlog (idea)
- [per-version-template-versioning.md](per-version-template-versioning.md) — idea — 为 effect/shape body 跨 AE 版本分裂准备的升级响应机制(YAGNI,gate 红了再建):局部字段分裂走版本条件 CODE(先例 ldta 160/164),整份 body 分裂把 effectTemplateFiles 升 (name,target)→path + per-target 覆盖文件 fallback 默认。检测靠现有双版本 ship-gate(红线6)。
- [property-synthesis-and-item-ops.md](property-synthesis-and-item-ops.md) — idea — 暂搁的大 feature 池:泛型 DuplicateItem · ImportComposition · Property synthesis(从零合成属性树,详 incidents/transform-group-default-omission.md)。撞真实需求再 flip active。

### Active · Done
- [2026-06-21-api-doc-tag-schema.md](2026-06-21-api-doc-tag-schema.md) — active — Replace free-prose doc comments + the separate aep:cap directive with ONE swaggo-style @tag block per exported symbol (docgen reads @summary/@description/@param/@returns; capindex reads @domain/@stability/@verify/@gate/@since/@boundary/@incident/@alias). Plus repo-wide cleanup of internal jargon (version codenames V2.2/V3/M8/wave-N, project-file/process refs CLAUDE.md/spec/probe/py-aep) from ALL comments, enforced by lint. Migrate behind a dual-read window (CI green throughout); 486 tags/26 files for schema + ~114 files for cleanup.
- [2026-06-19-booyah-glitch-full-replication.md](2026-06-19-booyah-glitch-full-replication.md) — active — 把整个 Booyah Glitch 真实工程(12 comp/61 层/~100 mask/wiggle 表达式/Curves)用咱们的 Go API 从零完整复刻,作为理解金标准检验。判据=结构保真(AE 接受 + DOM 读回值对账)+ 终帧渲染像素对照;范围=全工程 DAG 叶子优先逐 comp 验;方法=手写 per-comp 生成器、原工程仅当取值神谕、chunk 全由 API 重建(绝不 copy 字节)。
- [2026-06-18-technique-ontology.md](2026-06-18-technique-ontology.md) — active — 投大规模工程前冻结的数据骨架:角色/技法/机制三轴本体 + 技法条目 schema v2(等价效果集·可复刻性两轴·迁移分proven/hypothesized)+ 工程画像 schema + 受控词表 v0。schema 闭 instance 开,约束所有工程理解的产出形式,使跨工程可聚合出通用技巧。经火焰(生成型)/闪电(素材装配型)/控制器(lib-blocked)三类压测。
- [2026-06-18-fx-technique-internalization.md](2026-06-18-fx-technique-internalization.md) — active — 通用机制:任何参考 .aep → 解析 → 拆角色 → 抽跨域技法原子 → 存 flightdeck 两层结构(技法库 references/fx-techniques + 现象配方 checklists/build-X)→ AE gate 验证。火焰=实例#1已验证。适配风/雨/雷电/转场=复用技法+增量。是 procedural-fx-generator 的「配方提取」引擎的通用化。
<!-- /AUTO -->
