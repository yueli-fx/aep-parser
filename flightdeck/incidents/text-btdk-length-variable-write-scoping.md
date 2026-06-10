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

## 修法（解封路径，按今日证据）

- **布局缓存大概率不用精确重算——AE 载入时自己重算**。证据：NewTextLayer
  ship-gate 的 T2 probe（fresh 层等长 SetText "A"→"B"，btdk 布局缓存仍是 "A"
  的字形度量），AE 2020 + AE 2025 双版本照常打开、`sourceText.value.text`
  读回 "B"、resave 保留（`TestNewTextLayer_AEShipGate_AE20{20,25}` 2/2 PASS，
  2026-06-11）。等长 ≠ 字形等宽（A/B 度量不同），所以「stale 像素度量可被接受」
  已证；**未证**的是 stale 的缓存内字符计数（/5 键）在长度变化时是否也被容忍。
- **长度可变 splice 机制已有现成范本**：`layerBackrefs.AddFont`（back_layer.go）
  ——在 btds Data 内插字节 + 手动修内嵌 LIST btdk 的 size header
  （`bodyOff-8` 处 u32 += delta）；外层 LIST size 由 WriteAEP bottom-up 重算。
  文本串替换 = 同机制：替换 `/1/1[0]/0/0` 串字节 + 改写 2/3 处的整数计数
  （PsNum 重序列化也是 length-variable，同样走 splice）+ size header 修正。
- **建议实施顺序**：v1 限单段落（多段落要 splice 新段落 dict entry，结构性
  更大）：改串 + 段落计数 + run 计数 + size header，布局缓存原样不动 →
  AE 双版本 ship-gate。若 AE 拒（缓存内 /5 计数不容忍 stale），fallback 把
  `/1/1[0]/1` 整棵换成已知可接受的最小形态（如模板 "A" 的）再 gate。
- 多 run/多段落文本的计数分配策略（新字符归哪个 run）等同 AE 行为 RE，
  v1 可统一归 run[0]/段落[0]（单 run 单段模板下天然成立）。

## 相关

- [[kerning-first-enable]] — btdk 内 structural-add 被拒、splice 可行的先例
- [[camera-light-layer-create-re]] — embed-whole-Layr 创建机制（NewTextLayer 同族）
- `internal/serializer/mutate_layer_text.go` — NewTextLayer（模板 = re_text.aep
  baseline_A，创建时接 btds back-ref，fresh 层即可等长 SetText）

## Cases
- 2026-06-11 首次（NewTextLayer ship 时的 scoping pass）
