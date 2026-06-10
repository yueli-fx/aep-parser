---
status: active
when_to_read: implementing arbitrary-length SetText / NewTextLayer text param; editing btds bytes and wondering which counters must stay in sync; debugging text layout after a btds edit; assuming the btdk layout cache must be kept accurate
applies_to: [text, btdk, btds, settext, length-variable, layout-cache, char-count, new-text-layer, postscript, cooltype, re-finding]
last_updated: 2026-06-11
resolved_by:
---

# Text btdk length-variable write — scoping findings

## Signature
- symptom: `SetText length mismatch (new=N bytes, old=M bytes — length-preserving only; pad input to match)`
- error_type: —
- where: Layer.SetText (scene_layer_writers.go) / layerBackrefs.SetText (back_layer.go)
- trigger: writing a text string whose UTF-16BE encoding differs in byte length from the stored one

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
