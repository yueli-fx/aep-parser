---
status: active
summary: 让本库纯 Go 离线把 AE Pseudo Effect(.ffx 自定义伪效果)splice 进层 Effect Parade,免开 AE。.ffx=RIFX FaFX form,内含 sspc+ppar(pard)+tdgp = 一个 Effect-Parade 成员单元(同 AddEffect splice)。头号 crux=AE 认不认未 applyPreset 注册的 Pseudo/<uID> matchName,spike-first 双版本验。v1=apply-only 写侧,读侧解 .ffx 进 scene 列后续。
last_updated: 2026-06-20
---

> **进度**(2026-06-20):Phase 0 spike ✅ **GO**(AE 2020+2025 双版本 CRUX-PASS)。Phase 1 ✅ `rifx` 读 .ffx + 抽 effect-unit。Phase 2 🔶 `ApplyPseudoEffect` 已建(facade+serializer,复用 AddEffect 核心)、**离线端到端通**,但 AE 报 "file is damaged" → 实测发现 **.ffx sspc 需转换成 in-parade 形态**(fnam→Utf8 / parT 追加 built-in / tdgp 重构,详 §「Phase 2 关键修正」)。**下一步 = 实现该转换 + AE 复验**(先验 elision-非必需 hypothesis)。

# 支持 AE Pseudo Effect(自定义伪效果)纯 Go 离线应用

> 配套 RE 详档:`references/pseudoeffect-support-re.md`(.ffx 字节 dissect + rendertom 机制 + 5 条开放问题)。本 spec = **要做什么 + 怎么分阶段 + 判据**;那份 reference = **底层字节真相 + 参考实现**。两者勿重复,有冲突以本 spec 决策为准。

## 1. 动机 / 目标

AE 的 **Pseudo Effect**(自定义伪效果,Pseudo Effect Maker 产物)= 一组自定义控件(Slider/Color/Checkbox/Point/Angle/Layer/Dropdown…)包成的、长得像原生效果的属性树,matchName 形如 `Pseudo/<uID>/<name>`,导出为 Animation Preset(`.ffx`)。脚本作者大量用它给脚本造控制面板。

**目标**:让本库**纯 Go 离线**把一个 `.ffx` 伪效果 splice 进某层的 `ADBE Effect Parade` —— 拿一个 `.ffx` + 一个真实 `.aep`,产出「该层挂上了这个伪效果」的新 `.aep`,**全程不启动 AE**。

**与参考实现(rendertom/PseudoEffect)的本质区别**——这正是本库的价值所在:

| | rendertom/PseudoEffect | 本 spec |
|---|---|---|
| 运行环境 | **必须 AE 在跑** | **离线纯 Go** |
| 机制 | 写临时 .ffx → 临时 comp 临时层 `applyPreset` 让 AE 点亮 matchName → addProperty | 直接读 .ffx 字节 → splice 进 Effect Parade |
| 依赖 | ExtendScript + AE prefs「允许脚本写文件」 | 无 |

非目标(本 spec 不做):不抄 rendertom 的 AE 运行时路径;不做 `.ffx`↔binaryString 互转工具(我们直接读字节);Pseudo Effect Maker 那种「造伪效果」的编辑器不做。

## 2. 关键洞察(为什么可行)

`.ffx` **本身就是 RIFX 文件**,但 form type = **`FaFX`**(不是工程的 `Egg!`)。其内部 `sspc + ppar(pard 参数定义) + tdgp(参数值)` 这一坨,**就是一个效果在真实 `.aep` 的 `ADBE Effect Parade` LIST 里的成员单元** —— 与我们 `AddEffect` splice 的 `(tdmn, sspc, …)` 单元**同构**(见 `incidents/add-effect-splice-re.md`)。

差别只在来源:
- `AddEffect` 的模板 = 216 个**内置原生**效果(matchName 已被 AE 认识),嵌在 `internal/serializer/templates/`。
- Pseudo 效果 = **用户自定义**、matchName 不在 AE 原生注册表;**但它的 pard 控件定义随字节自带**——效果「长什么样」全写在 `.ffx` 里。

→ 支持路径 = **「AddEffect 但模板来自调用方传入的 `.ffx`,而非嵌入集」**。复用现有 splice 机制(原子不变量 + warnings-as-failure + rollback)。

## 3. 头号风险(make-or-break,spike 先验)

**★ AE 会接受一个「从未 `applyPreset` 注册过」的 `Pseudo/<uID>` matchName 吗?**

rendertom 整套「点亮(make live)」hack 的存在,暗示 AE 运行时**可能**要求 matchName 先注册才认。但伪效果的 pard 控件定义随 `.aep` 自带 → AE **也可能**能 standalone 消化渲染。**这条决定整个 feature 成立与否**,必须最先用 spike 实测,不靠推理。

若 spike 证 AE **拒收**未注册 matchName → feature 退化为「只能在已装该伪效果的 AE 里用」,价值大减,届时给实证负结论、转记 incident,本 spec 收为「已探明不可达」。

## 4. 范围与分阶段

### Phase 0 — Spike(de-risk,必须最先)✅ **PASS / GO**(2026-06-20)
**结论:头号 crux 在 AE 2020 + 2025 双版本决定性通过 —— feature 物理成立。**

实测路径(`tmp_debug/pseudo-spike/`):
1. **离线**:Node 从 test.js binaryString 重建 `Scribe.ffx`(4572B,RIFX `FaFX` form);我们的 `rifx.ReadChunk` **完整解出** FaFX 树,effect-unit = `LIST sspc`(fnam + parT(8 pard 控件定义) + tdgp(值区))定位清晰。→ **Phase 1 离线 de-risk 完成**。
2. **Launch A(注册 session)**:JSX 新建 shape 层 → `applyPreset(Scribe.ffx)` → 效果活、8 参数全在 → 存 `scribe_baked.aep`。AE 2020 + 2025 各一份。
3. **离线核对**:Go parser 读 baked.aep → **AE 把完整自包含定义烤进字节**(sspc + 8 pard 全在,非裸 matchName 引用)。
4. **★ Launch B(决定性:全新进程 = 未注册)**:打开 baked.aep → 效果**完全活**(`name=Scribe` 非「Missing:」、`enabled=true`、8 参数带名)。**AE 2020 + 2025 均 CRUX-PASS**。

→ **AE 接受未经 `applyPreset` 注册的自包含伪效果;pard 控件定义随 `.aep` 字节自带、跨进程存活。** rendertom 的运行时点亮 hack 对「离线写 .aep」**不必要**。

**意外收获 = 字节 oracle**:`scribe_baked_2020.aep` / `scribe_baked.aep`(2025)是 AE 亲手写的「正确的 in-parade 伪效果」样本 → Phase 2 splice 直接 byte-diff 对照,无需猜目标布局。

**仍未验(Phase 2 才验)**:① 我们 **Go-splice 出的字节**(非 AE-baked)被 AE 接受 —— API 未建;② **render-pixel**(本 spike 用空路径 shape 层,Scribe 无路径可描 → 帧空白 234B,符合预期,非渲染验证)。

判据(原始,已满足 accept 面、render 面留 Phase 2):
- **判据 A(accept)**:✅ AE 2020 + 2025 DOM 读回 `matchName == "Pseudo/9db0uID/Scribe"`、8 参数、`name` 不含 Missing、`enabled=true`。
- **判据 B(render)**:⏸ 移到 Phase 2 ship-gate(需带路径的层让 Scribe 可见)。

### Phase 1 — `.ffx` reader(FaFX form)〔Phase 0 已离线 de-risk:`rifx.ReadChunk` 已能读 FaFX 树〕
- `internal/rifx`:`Parse` 硬拒非 `Egg!`,但 `ReadChunk` 不校验 form → 读 `.ffx` 走 `ReadChunk`(已验)或给 `Parse` 加 `FaFX` 白名单。
- `internal/serializer` 增 `.ffx` 解析:抽出 effect-unit 子树(`sspc + ppar + tdgp`),丢弃 preset 专属的 `besc/beso/tdsp`(路径描述)+ `pgui`。
- 输出一个内部「pseudo-effect 模板」表示(matchName + 控件定义 + 默认值字节)。

### ⚠ Phase 2 关键修正(2026-06-20 实测发现)— **.ffx sspc ≠ in-parade sspc,需转换**
**「verbatim splice」假设证伪**。已建 `ApplyPseudoEffect`(facade + serializer,复用 AddEffect 的 `addEffectFromChunks` 核心),**离线端到端通**(splice→WriteAEP→重读,matchName+8 参数对)。但 AE 2025 打开 Go-spliced .aep 报 **"file is damaged"**(经典红线7a:Go round-trip 绿 ≠ AE 接受)。

byte-diff(go-spliced vs AE-baked oracle `scribe_baked.aep`)定位:effect 单元同为 `sspc(fnam,parT,tdgp,pgui)`,但 AE **applyPreset 烤进工程时重构 .ffx**,三处 delta:
1. **fnam**:.ffx 48B 定长 NUL-pad → in-parade `Utf8` 包装(14B:`Utf8`+len+name)。
2. **parT**:.ffx 17 children(parn + 8×(tdmn+pard))→ in-parade 19(**追加 `ADBE Effect Built In Params` tdmn+pard** = Compositing Options,通用常量)。
3. **tdgp 值区**:.ffx 19(`tdsb`+`tdsn`+8×(tdmn+tdbs)+GroupEnd)→ in-parade 11(**去 tdsb/tdsn 前缀** + **按 elision 丢默认值参数**,只留非默认 -0000/-0003/-0007 + built-in 值组 + GroupEnd)。

**关键 hypothesis(待验,决定 Phase 2 难度)**:**elision 非接受必需**——.ffx 自身含全量非默认值、AE applyPreset 照吃,AE 只在**存盘**时 elide。若成立,转换 = fnam→Utf8 + parT 追加 built-in pard + tdgp(去 tdsb/tdsn 前缀 + **保留全量值** + 追加 built-in 值组),**无需实现 per-param 默认值比对**(那才是难点)。两个 built-in 常量块(parT 的 pard + tdgp 的值组)可从任一原生 effect 模板 runtime 抽。
→ 下一步 = 实现此转换 → AE 2020+2025 复验(先验 hypothesis:非 elide 是否被接受)。

### Phase 2 — `ApplyPseudoEffect`(写侧,v1 主交付)
- facade 自由函数(对齐 CLAUDE.md #2/#3 结构性 op 住 serializer):
  ```go
  func ApplyPseudoEffect(layer *Layer, ffxBytes []byte) (*Effect, error)
  ```
  读 `.ffx` → 抽 effect-unit → splice 进 `layer` 的 `ADBE Effect Parade` → 返回解析后的 `*Effect`,调用方可继续 `Set*` 调参。
- 复用 `AddEffect` 的 Effect-Parade 插入路径(原子不变量 + rollback)。
- 核对 `tdpi`(add-effect incident 记其为 host-layer 绑定)是否需重指向目标层。
- **byte oracle**:byte-diff 我们的输出 vs Phase 0 的 `tmp_debug/pseudo-spike/scribe_baked{,_2020}.aep`(AE 亲手写的正确 in-parade 伪效果)。
- **render-pixel**:用**带路径的层**(让 Scribe 实际描边)出可见帧,采样像素验(红线4)。
- **Gate**:AE 2020 + 2025 双版本 ship-gate(accept + render-pixel),过了才升 Stable;未过保持 Alpha。

### Phase 3 — 多控件类型覆盖
Scribe 只示范 group / Color / Slider。补验 Point / Angle / Checkbox / Dropdown / Layer-ref 各自 pard 布局 + 值 round-trip。逐类型加 fixture。

### 后续(本 spec 之外,需求驱动再立)
- **读侧**:把 `.ffx` 伪效果解进 scene 模型 + JSON 导出(当前 v1 只 apply,不读)。
- **showcase**:effect 域已有眼验面,Pseudo apply 过 gate 后并入 effect showcase 批次。

## 5. 公共 API 契约(初版,Alpha)

- `ApplyPseudoEffect(layer *Layer, ffxBytes []byte) (*Effect, error)` — Alpha,过双版本 gate(accept+render)后升 Stable。
- 可能的读侧 helper(Phase 1 副产物,内部优先):解析 `.ffx` → 模板;是否导出公共 API 待定(见决策点)。
- capindex:新增 `aep:cap` tag,domain=effect,初始 `stable=alpha · verify=none`,gate 过后升 `verify=render-pixel`。

## 6. 待 RE / 验证清单(撞墙前别宣称能用,对齐红线7)

1. **(crux)** 未注册 matchName 的 AE 接受性 —— Phase 0 spike。
2. **FaFX form + besc/beso/tdsp 语义** —— 抽 effect-unit 的精确边界(哪些 chunk 丢、哪些必带)。
3. **`tdpi` 绑定** —— 值区里的 `tdpi 0x26` 是否需随目标层重写。
4. **版本可移植性** —— `.ffx` 由某版 AE 导出,splice 进另一版工程是否 portable(参 `docs/embed-template-architecture.md`)。
5. **多控件 pard 布局** —— Point/Angle/Checkbox/Dropdown/Layer-ref(Phase 3)。

## 7. 决策点(需用户/实测拍板)

- **D1 v1 范围**:默认 **apply-only 写侧**(读侧解 .ffx 进 scene 列后续)。← 已默认,除非反对。
- **D2 API 入参**:`ffxBytes []byte` vs 也提供 `ApplyPseudoEffectFile(layer, path)` 便捷重载。倾向先只 `[]byte`,path 重载是糖。
- **D3 matchName 冲突策略**:同一 `Pseudo/<uID>/<name>` 在同工程被 apply 两次 —— AE 行为待 RE(uID 是否要求全局唯一 / 重复是否合并定义)。

## 8. 依据

`references/pseudoeffect-support-re.md`(字节 dissect + 参考实现)· `references/PseudoEffect/`(rendertom 源 + Scribe 样本)· `incidents/add-effect-splice-re.md`(effect splice 机制,本 feature 复用)· `docs/embed-template-architecture.md`(模板/跨版本)· capindex `AddEffect`(216 内置库基线)· CLAUDE.md 红线 4/7 + #2/#3(API 分级 + 物理分层)。
