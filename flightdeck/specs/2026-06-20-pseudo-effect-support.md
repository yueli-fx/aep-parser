---
status: active
summary: 让本库纯 Go 离线把 AE Pseudo Effect(.ffx 自定义伪效果)splice 进层 Effect Parade,免开 AE。.ffx=RIFX FaFX form,内含 sspc+ppar(pard)+tdgp = 一个 Effect-Parade 成员单元(同 AddEffect splice)。头号 crux=AE 认不认未 applyPreset 注册的 Pseudo/<uID> matchName,spike-first 双版本验。v1=apply-only 写侧,读侧解 .ffx 进 scene 列后续。
last_updated: 2026-06-20
---

> **🏗️ 从零造(2026-06-20,方向转向)**:用户拍板——**应用 .ffx 只对伪效果无价值,转做纯 Go 生成伪效果;且不许 clone 模板偷懒、要解析理解后合成**。`BuildPseudoEffect(layer, uid, name, displayName, controls)` —— **零 .ffx、零 AE、零模板字节**,每个 `pard` 按 RE 出的 148B 布局逐字段合成。控件类型(全)Slider/Color/Checkbox/Angle/Point/Point3D/Dropdown/Group/Label/Layer,AE 2020+2025 ship-gate 绿(`TestBuildPseudoEffect*_AEShipGate_*`,3 组 gate)。RE 真相源 = `test_data/pseudo_rich_demo.aep`(用户 Pseudo Effect Maker 产物,13 控件全类型)。pard 格式:`@0x0F`类型·`@0x10`名(ANSI)·`@0x30`=2结构/0有值·`@0x40+`值区;checkbox/dropdown 需尾随 `pdnm`;`@0x50+` 是 AE 栈垃圾(清零无妨)。**支线已收口**(见 §9 收口段 + `docs/pseudo-effect-continuation-handoff.md`):Dropdown/Group/Label = pard-only(组=扁平标记非嵌套);Point 坐标/Layer-picker = 值条目合成(cdat=坐标分数·tdpi=层 ID)。**唯一未解 = CJK 控件标签**(AE 架构限,byte-equiv-only)。
>
> **额外能力(2026-06-20)**:`ApplyPseudoEffectNamed(layer, ffx, displayName)` —— 可设效果实例**显示名**,**支持任意 UTF-8 含中文**(如「伪效果」)。RE 发现显示名落在**值组 tdsn**(非 fnam);Utf8 子记录按**字节长**编码,故多字节中文精确 round-trip。AE 2020+2025 ship-gate 绿(`TestApplyPseudoEffectNamed_CJK_AEShipGate_*`,读回 U+4F2A/6548/679C)。matchName 仍 ASCII。**(社区工具普遍栽在中文名,本库过)**。
>
> **进度**(2026-06-20):Phase 0 ✅ GO · Phase 1 ✅ · **Phase 2 ✅ DONE** —— `ApplyPseudoEffect`(facade+serializer)**AE 2020+2025 双版本 ship-gate 绿**(committed:`TestApplyPseudoEffect_AEShipGate_AE2020/2025`)。转换 = fnam→Utf8 + parT 追加 `ADBE Effect Built In Params`(parn 重算)+ tdgp 丢 .ffx 值仅留骨架。**v1 范围 = 以 pard 默认值应用**(.ffx 作者值暂不保留——AE 拒裸值 'missing data';Set* 调参未接)。下一步(可选增强,非 v1):保留作者值(per-param 值 materialize 进合法 in-parade tdbs)+ Set* 调参 + 多控件类型/读侧。

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

---

## 9. 方向转向 — 从零生成(`BuildPseudoEffect`,2026-06-20)

用户拍板:**apply-only 无价值,转「纯 Go 从零合成伪效果、禁用模板」**(「不能什么都用模版偷懒」)。`ApplyPseudoEffect*` 保留(双版本 gate 绿,含 CJK 实例名),但主交付转 `BuildPseudoEffect`。

### 已交付(双版本 AE ship-gate 绿)
- `BuildPseudoEffect(layer, uid, name, displayName, controls)` — 每个 pard 按 RE 出的 148B 布局**逐字段合成**,无 .ffx / 无 AE / 不 clone 模板字节。控件类型:Slider/Color/Checkbox/Angle/Point/Point3D。
- **自定义值(pard 级,AE 实测回读)**:`PseudoControl{Min,Max,Default,Checked,Color}`。gate 实测 AE 读回 slider value 50 / range -100..100、checkbox checked、angle 45°。

### RE 真相(金样本 `test_data/pseudo_rich_demo.aep`,13 控件 AE-authored)
pard 148B,值字段全在 pard 内(被 elide 的控件 AE 仍从 pard 读 min/max/default):
- **Slider(0x0a)**:@0x04=0x200 flag · @0x38 f8 default · @0x68/@0x6C f4 **valid range = AE minValue/maxValue** · @0x70/@0x74 f4 visible(track)range · @0x78 f4 default · @0x7C=0x00050003 精度/显示 flag。
- **Color(0x05)**:@0x38 ARGB last · @0x3C ARGB default(**非** @0x40——早期草稿写错偏移,白色凑巧过 gate)。
- **Angle(0x03)**:@0x38/@0x3C s4 度数 16.16 定点(last/default)。
- **Checkbox(0x04)**:@0x38 u32 last · @0x3C u8 default(1=勾)· parT 内尾随 `pdnm`(Utf8 标签,缺则 AE「must have nameptr set」)。
- **Point(0x06)**:@0x3C=0x00050000 · @0x48=0x00640000 结构常量(缺则 AE「range has no values」);**坐标在值条目,不在 pard**。
- **Dropdown(0x07)**:@0x38 last · @0x3C hi16=nb_options/lo16 · parT 内尾随 `pdnm`=「opt1|opt2」。
- **Layer-picker**:control_type **0x00**(同 effect header)+ @0x30=2;值条目带 `tdpi`+`tdps`。
- **Group/Label(0x0d)**:label 带 @0x04=0x20 flag,group 不带;配对 **GroupEnd(0x0e)** @0x04=0x08。

### CJK 控件标签 — 决定性负结论(2026-06-20 gate 实测,详 `incidents/pseudo-control-label-ansi-codepage.md`)
原 hypothesis「value-entry tdsn 驱动控件标签」**被 gate 证伪**:
- 合成了 checkbox/color/point 非 elide value entry(`tdsb + tdsn(Utf8) + tdb4(RE'd 每类型) + cdat`)→ **AE 接受**(value-entry 机制可行),但
- 控件标签仍来自 **pard @0x10 名**(本机 cp1252 误读成乱码,8364=€ 暴露系统码页为西欧非 GBK),**tdsn 不覆盖标签**。

**真相 = AE 架构限制**:控件标签 = pard @0x10 名,按**查看机系统 ANSI 码页**解码(非 UTF-8、非 tdsn)。AE 自家 Pseudo Effect Maker 写 GBK(金样本 color pard `@0x10=d1d5c9ab`=「颜色」GBK),仅在 GBK Windows 显示对。**无可移植解**。中文系统唯一解 = pard 名 GBK 编码(字节等同 AE 原生输出,需 x/text 依赖 + locale 假设,本西欧码页机**无法 ship-gate**)——已搁,待用户拍依赖/locale 取舍。
**效果显示名无此限**(值组顶层 tdsn / Utf8),`displayName`「中文效果」gate 绿。
value-entry 合成本身已验可行(AE 接受非 elide 值条目),Point/3DPoint 默认坐标 + Layer-picker 仍可走它,**但不为 CJK 标签**。已 revert(不留无收益的 opaque tdb4 字节)。

### ship-gate verify-JSX gotcha(可复用)
ExtendScript 里对**刚 fetch 的伪 slider 属性**直接读 `p.minValue` 返回 stale(=maxValue);必须**先碰 `p.hasMin`** 再读 `minValue`(`maxValue` 同理需先碰 `hasMax`)。`fx.property(matchName)` 读 minValue 也踩此坑,改 `fx.property(index)` + 先碰 hasMin 才稳。RE 决定性证据:同一文件 probe(先碰 hasMin)报 -100、gate(冷读)报 100。

### 仍未做
- **CJK 控件标签** = AE 限制,仅 GBK-pard-name 可行(locale + 依赖 + 不可 gate,待用户拍)。

### ✅ 支线收口(2026-06-20,commit 4fa4a5c + d661bb9)
全部剩余控件类型已实现 + 双版本 AE ship-gate 绿。Pseudo Effect Maker 能造的控件类型现在都能纯 Go 离线合成。
- **Dropdown / Group / Label**(`TestBuildPseudoEffectRich_AEShipGate_*`):pard-only(值条目可省略)。**决定性发现**:伪效果「组」=**扁平标记控件**(视觉分组,非属性树嵌套);AE 原生金样本读回同样扁平,仅内置 Compositing Options 真嵌套。
- **Point/3DPoint 自定义坐标 + Layer-picker**(`TestBuildPseudoEffectValueEntry_AEShipGate_*`):值条目合成。**Point cdat = 坐标空间分数**(读金样本回 AE 反推:[0.000025,2500]@500px → cdat[5e-8,5.0],比值=维度);实测 [0.25,0.125]→[100,50]。**Layer-picker** 绑定=值条目 `tdpi`=层内部 ID(`aep.Layer.ID`),gate 绑指定层、AE 读回该层索引。per-type tdb4(124B)逐字节抄金样本(@0x10=per-dim 常量非值相关);新增 rifx `IDTdps`。
- 收口记录详 `docs/pseudo-effect-continuation-handoff.md` §0。
