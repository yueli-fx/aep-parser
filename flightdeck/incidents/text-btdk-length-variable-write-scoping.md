---
status: active
when_to_read: implementing arbitrary-length SetText / NewTextLayer text param; editing btds bytes and wondering which counters must stay in sync; debugging text layout after a btds edit; assuming the btdk layout cache must be kept accurate
applies_to: [text, btdk, btds, settext, length-variable, layout-cache, char-count, new-text-layer, postscript, cooltype, re-finding]
last_updated: 2026-06-12
resolved_by:
---

# Text btdk length-variable write — scoping findings

## Signature
- symptom: `SetText length mismatch (new=N bytes, old=M bytes — length-preserving only; pad input to match)`
- error_type: —
- where: Layer.SetText (scene_layer_writers.go) / layerBackrefs.SetText (back_layer.go)
- trigger: writing a text string whose UTF-16BE encoding differs in byte length from the stored one

> **STATUS (2026-06-12, v3)**: empty + multi-paragraph + **multi-run** SetText 全
> **SHIPPED**（ship-gate PASS）。变长 refuse 集现仅剩 **手动 kerning 表**一项。单/多
> 段落 × 单/多 run × 空串 = 全支持。findings 见 § 修法 v2 / v3。v1 守卫描述存历程。

## 症状/复现

`Layer.SetText` 只接受等长替换——任意长度文本写（"A"→"Hello"）今天做不了。
本 incident 记录 2026-06-11 NewTextLayer ship 过程中对「为什么难、还差什么」的
字节级 scoping，供解封时直接开工。

## 根因 — btdk 里跟文本长度耦合的三处（RE: re_text.aep baseline_A，6657B btdk）

一段式 point text "A"（存储为 "A\r"，2 个 UTF-16 字符）：

1. **文本串本体** `/1/1[0]/0/0` — UTF-16BE + BOM 的 PostScript 串（含尾 `\r`，
   每段一个）。串字节范围由 `ParsePSDict` 的 SrcStart/SrcEnd 给出。
2. **段落字符计数** `/1/1[0]/0/5/0[i]/1` — 每段落 style entry 的兄弟 `/1` =
   该段字符数（baseline_A: `2`）。
3. **style-run 字符计数** `/1/1[0]/0/6/0[i]/1` — 每 run entry 的兄弟 `/1` =
   该 run 覆盖字符数（baseline_A: `2`）。
4. **布局缓存** `/1/1[0]/1` — `/PC → /F → /R → /L → /S → /G` 嵌套树：逐行/逐
   段字符计数（`/5` 键，baseline_A 多处 `2`）+ 字形像素度量（`/14 -75.625`
   `/15 19.9375`、bbox 数组 `[0,-75.625,44,19.9375]`、宽度表 `[36,3]`…）——
   全部跟「渲染出来的具体字符串」耦合。这棵树才是 length-variable 写的大头。

## 修法 — v1 已 SHIP（2026-06-11，同日解封）

`Layer.SetText` 升级 length-variable（签名不变，语义扩展）：三次 `splicePSValue`
（串 `/1/1/0/0/0` + 段落计数 `/1/1/0/0/5/0/0/1` + run 计数 `/1/1/0/0/6/0/0/1`，
每次 re-parse 刷新偏移 + 修内嵌 btdk size header `bodyOff-8`，AddFont 同机制），
布局缓存原样不动；快照回滚保原子。**AE 2020 + AE 2025 双版本 ship-gate 2/2
PASS**（`TestSetTextVariable_AEShipGate_AE20{20,25}`：Go-built 工程
"A"→"Hello AEP parser"〔17 units〕+ "A"→"你好世界"〔5 units〕，AE 读回精确、
resave 保留）——**stale 布局缓存在长度变化尺度下也被 AE 容忍**（缓存内 /5
计数同样 stale，AE 载入全部重算；此前 NewTextLayer T2 probe 只证了等长尺度）。

**v1 守卫（refuse 集，解封需 entry-splicing RE）**：变长改字仅限单段落 + 单
style-run + 无手动 kerning 表（`/1/1/0/0/8/0`）+ 非空。多段落变长需 splice 新
段落 dict entry；多 run 需计数分配策略（新字符归哪个 run = AE 行为，未 RE）。

## 修法 — v2 已 SHIP（2026-06-12，空串 + 多段落解封）

用户点名解封空串 + 多段落（跳过多 run）。两件事 RE 自 `re_text_multipara.jsx`
（AE 2020：`empty_addText` / `empty_setValue` / `three_para` / `one_para`）+ 已有
`re_text.aep` 的 `text_two_lines`：

**RE finding 1 — 空串不是「distinct form」**（v1 注释的担忧证伪）。AE 空文本层
btdk = 单段落路径：串 `/1/1/0/0/0` = `"\r"`、段落计数 `/0/5/0/0/1` = 1、run 计数
`/0/6/0/0/1` = 1，其余全是 layout 缓存（AE 载入重算）。`addText("")` 与
`setValue("")` 仅 leading 度量一处差（runtime，AE 自给）。**=> 删 `normalized=="\r"`
守卫即解封，现有 splice 路径原样产出。**

**RE finding 2 — 多段落 = N 个逐字节相同的段落 entry 克隆**。`/1/1/0/0/5/0` 段落
数组每段一个 entry，**仅 `/1` 单位计数不同**（该段文本 + 尾 `\r`，UTF-16 units）；
样式 dict（`/0/0/0/5` 的 40-key 块）逐字节相同。**单 style-run 不变**——run 数组
仍 1 个 entry，`/1` = 总 units。例：`L1\rL2\rL3\r` → 段落计数 `[3,3,3]`、run 计数 `9`。

**修法**：变长路径改为 rebuild 段落数组——取现存 entry `[0]` 原始字节当模板，按
目标段落数克隆，每份 patch `/1`（`buildParagraphArray` / `paragraphEntryCountRange`
in `back_layer.go`）。三次 splice：串 `/1/1/0/0/0` + 段落数组 `/1/1/0/0/5/0`（整组替换）
+ run 计数 `/1/1/0/0/6/0/0/1`。layout 缓存（含段落数不符的 stale 结构）原样留给 AE
重算。**AE 2020 + AE 2025 双版本 ship-gate PASS**（`runSetTextGate` T1-T4：ASCII 1→17
units、CJK 1→5、3 段落块、空串）——**stale 缓存在段落数变化下也被 AE 完全重算容忍**
（v1 只证了单段落长度变化尺度）。

**实现细节**：`jsExpect`（AE `value.text` 用 `\r` 段落分隔，无尾 terminator）≠ Go
`TextSource.Text`（`\r`→`\n`、剥尾 `\r`）——ship-gate verify JSX 比 `\r` 形、Go resave
preservation 比 `\n` 形。`jsStringEscape` 须转义控制字符（`\r`/`\n` → `\uXXXX`）否则
`eval()` 见裸 line terminator 炸。

## 修法 — v3 已 SHIP（2026-06-12，多 run 解封）

用户点名继续解封多 run。**RE finding（`re_text_multirun.jsx` @ AE 2025，characterRange
造 2-run "Hello" 红/蓝）= 整文本替换时 AE 把 run 数组 collapse 成单个，保留 run[0] 样式**：
- `multirun_src`：run 数组 `/1/1/0/0/6/0` n=2，run[0] fill `[1,1,0,0]`(红)、run[1] `[1,0,0,1]`(蓝)。
- `multirun_setvalue`（`td.text=...` + `setValue`）→ **run 数组 n=1，fill `[1,1,0,0]`=run[0] 红**。
- `multirun_charrange`（`characterRange(0,5).text=...`）→ 同样 collapse 到 n=1。

=> 「多 run 计数分配」根本不需要 RE——AE 自己就是 **collapse-to-first-run**。`multirun_setvalue`
就是 AE 亲手产出的 ground truth。

**修法**：变长路径加一步——run 数组 `/1/1/0/0/6/0` rebuild 成单个 entry（clone run[0]、
`/1`=总计数），跟段落数组同一个 `buildEntryArray` helper（段落/run entry 同构，top-level
`/1` = 计数）。删多 run refuse 守卫。最终 splice 三处：串 + 段落数组（整组）+ run 数组（整组）。

**ship-gate**：`runSetTextMultiRunGate` 加载 `re_text_multirun.aep`（2-run）→ SetText collapse
→ AE 读回 + resave 单 run。**AE 2024 + AE 2025 双版本 PASS**（fixture 用 AE 2024 生成 =
AE-24 stamped：AE 2024 native 开、AE 2025 向后兼容开）。**AE 2020 仍 N/A**——fixture 需
characterRange(AE 24+) 故最低 AE-24 stamped，AE 2020 前向拒开（「使用版本 X 保存，无法用此
版本打开」，非我方字节问题）；AE 2020 也无 multi-run scripting API 造不出 2020-openable
fixture。`buildEntryArray` 机制另由段落 case（T3 双版本含 2020 PASS）跨版本证过。
（**AE 2024 2026-06-12 解锁** = 给 24+ 引入字段补「24+ 但非 25」第二门禁版本，详
`checklists/re-fixture.md` § Adobe 软件路径。）

**v3 守卫（剩余 refuse 集）**：仅 **手动 kerning 表**（`/1/1/0/0/8/0`，per-char 数组会 desync）。
单/多段落 × 单/多 run × 空串 = 全支持。多 run 输入按 AE 行为 collapse 到单 run（保 run[0] 样式）。

> **AE 2025 ship-gate 注意**：首跑 + warm-retry 都 exit 2（unknown modal），OCR 抓到
> 的是 **About/credits 启动闪屏**（"1992...2024 / 保留所有权利 / 署名"），非数据 reject
> 对话框。clear crash state + kill AfterFX 后干净重跑即 PASS（12s）。冷启动闪屏 flake，
> 非 reject——但仍照 [[feedback_ship_gate_exit2_capture_dialog]] 先验 OCR 再判。

**实现陷阱（写时踩到）**：

1. **等长快路径必须比「段落剖面」而非总量**：`"A\rB"→"XYZ"` 字节数、总字符数
   全相等，但段落边界移动——原地写会让段落计数 [2,2] 失真（应为 [4]）且不报错。
   快路径条件 = 字节等长 **且逐段 UTF-16 计数相等**（`sameParagraphProfile`）。
   旧 SetText（只比字节）有此潜伏 bug，本次一并堵死。
2. **`DecodePSString` 逐 unit `rune(cu)` 打碎代理对**：astral 字符（𝄞/emoji）
   decode 成 U+FFFD——改 `utf16.Decode` 合成。计数侧 `UTF16CodeUnitLen` 按
   code units（代理对=2），与 AE 的计数单位一致（你好=3 排除字节假设，fixture
   text_unicode/text_hello 验证）。
3. **`re_text.aep` 有历史累积的重复 RE_TEXT comp**（老 JSX 无 fresh_project
   守卫，[[jsx-state-leak]] 活例）：按名字找层会命中多个，测试须按 parse 序
   固定取首个，否则「改了最后一个、验了第一个」假阴。

## 相关

- [[kerning-first-enable]] — btdk 内 structural-add 被拒、splice 可行的先例
- [[camera-light-layer-create-re]] — embed-whole-Layr 创建机制（NewTextLayer 同族）
- `internal/serializer/mutate_layer_text.go` — NewTextLayer（模板 = re_text.aep
  baseline_A，创建时接 btds back-ref，fresh 层即可等长 SetText）

## Cases
- 2026-06-11 首次（NewTextLayer ship 时的 scoping pass）
- 2026-06-12 v2 ship：空串 + 多段落解封（双版本 ship-gate PASS）；refuse 集缩到多 run + kerning
- 2026-06-12 v3 ship：多 run 解封（AE = collapse-to-first-run，AE 2024+2025 双版本 gate PASS，2020 N/A 前向不兼容）；refuse 集仅剩手动 kerning
