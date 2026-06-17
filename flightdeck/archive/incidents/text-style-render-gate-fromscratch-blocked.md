---
status: obsolete
when_to_read: 想给文字样式 setter（SetRunFontSize/SetRunLeading/SetRunTracking/SetParagraphJustification/SetRunCapsOption）补 render-pixel gate；从零 NewTextLayer 的文字渲染不出来 / 渲染在画面外；SetRunFontSize 写完 AE 读回字号 ÷65536；纠结文字样式族为何只到 verify=roundtrip；**saveFrameToPng 在 headless 多 comp gate 里每张 PNG 内容都一样 / 跨进程返回旧帧**；判断要不要投入「materialize 从零文字 transform/matrix」
applies_to: [text, text-style, SetRunFontSize, SetRunLeading, SetRunTracking, SetParagraphJustification, SetRunCapsOption, render-gate, red-line-4, from-scratch-text, NewTextLayer, text-matrix, transform-omission, position-not-materialized, fontsize-65536, btdk, run-style, roundtrip-not-render, negative-finding, saveframetopng, disk-cache, comp-id-collision, stale-frame, false-green, formatpsreal]
last_updated: 2026-06-16
resolved_by: FormatPSReal(point-key REAL 编码) + TestMGTextStyle 双版本 render-pixel gate(clearAEDiskCache harness)
---

# 文字样式 render-gate：从零文字的 transform/matrix 缺陷挡路（红线4 揪出真问题）

## TL;DR

文字样式族（`SetRunFontSize` / `SetRunLeading` / `SetRunTracking` / `SetParagraphJustification` / `SetRunCapsOption`）当前 **verify=roundtrip**（btdk PostScript splice，Go round-trip 绿，从未 AE 实渲验证）。2026-06-16 尝试把它们升到 render-pixel gate，**render-gating 当场揪出从零文字不可渲染**：纯 Go `NewTextLayer` 造的文字层在 AE 里**完全渲染不出来**（连默认 size=88 都看不见），且 `SetRunFontSize(150)` 后 AE 读回 `doc.fontSize = 0.00228 = 150/65536`。**两个症状同源**：从零文字层缺正确的 **text matrix / 坐标 scale**，AE 用 65536 当基把字号/位置全压扁。**writer 没错**（见下 byte 对账），是从零构造缺陷。**这正是 verify=roundtrip 标签掩盖的红线4 缺口**（值 round-trip 绿 ≠ AE 渲对）。

## Bisection（已定位，勿重走）

1. **不是 SetText / from-scratch 文本本身**：同样走 `NewTextLayer`+`SetText` 但**不调** SetRunFontSize 的对照层（TS_DEFAULT），`doc.fontSize` 读回 **88（对）**——唯一变量是有没有调 SetRunFontSize。
2. **不是 writer 写错位置/格式**：dump 两边 btdk（`tools/debug/parse_btdk`）——
   - AE 原生 `re_text.aep` 的 `size_200` 层：run-style 路径 `/1/1[0]/0/6/0[0]/0/0/6/1` 存 **`/1 200` 纯整数**（非定点）。
   - 我方从零 `SetRunFontSize(150)`：**完全相同路径**存 **`/1 150` 纯整数**，字节格式一致。
   - → writer 写对了。decoder 读 `/1` 为 `v.Num`（plain，与 AE 原生值一致）。Go 编/解自洽。
3. **不是文字颜色**：白 BG + 暗字测量也全空白（TS_DEFAULT 渲出纯白，无字形）。
4. **不是测量阈值**：目视 png 确认画面真的没有任何字形。

→ 结论：AE 读 size_200 的 `/1=200` 得 200，读我方 `/1=150` 得 150/65536。**同字节格式不同解读 = 从零层缺某个 AE 据以缩放字号/定位的字段（text matrix / 文档级 size / 坐标 base，疑 16.16 fixed 上下文）**。fixture 有、从零没有。这也解释从零文字渲染在画面外/不可见。

## 关联已知限制

- **从零文字 Position 不可设**：`Layer.Position()` 对文字层恒 nil（fixture 文字层也 nil），transform Position 默认省略（[[transform-group-default-omission]]）。`showcase/text` 注释明确「无 position/colour/size setter」，只放**单个**默认文字层。本次 render-gate 需把多个文字摆进可测位置，撞上同一堵墙。
- **字体轴用 16.16 定点**：`scene_text_decode.go:58-71` 已知 font variation axes（wght 等）走 `v.Num/65536`；字号 `/1` 当前**不**走定点（decoder `r.FontSize = v.Num`），与 AE 原生 plain 值一致——故定点不是字号存储格式，而是 AE 渲染期对缺 matrix 的从零层施加的 scale。

## 可渲染验证的子集（若将来要做）

- **Tracking / Justification** 在 fixture（已正确定位 + 有 matrix 的 AE 原生文字层）上 round-trip 正常（trk=1000✓、just=center✓ 读回），理论可在 fixture 上 render-gate；从零层因 matrix 缺陷一律渲不出。
- **FontSize / Leading** 即使在 fixture 上也需先确认 AE 是否对其也 ÷65536（未测——本次只在从零层观察到，fixture 期望正常因 matrix 在）。

## 建议（交给需求方决策）

要把文字样式族升到 render-pixel，先解 **从零文字 transform/text-matrix materialize**（synthesis-insert 文字层完整 transform + 文本 matrix，使 Position 可设 + 字号按 plain points 渲染）——这是独立的中-大型 RE arc，**非「中等收益」小活**。在此之前文字样式族**诚实保持 verify=roundtrip**，边界标注「从零文字渲染未验证（matrix 缺陷）；setter 在已 parse 的真实文字层上 round-trip 正常」。**不要**因 Go round-trip 绿就标 render-pixel（假绿）。

## UPDATE 2026-06-16（render-gate 实跑，部分推翻 TL;DR）

前述 TL;DR「从零文字完全渲染不出来 / 默认 size=88 也看不见 / `SetRunFontSize` 读回 ÷65536」**已不再成立**。本次先扫清挡路的 KBar evalScript-timeout 模态（ship-gate 卡 exit-2，已加 `ae_dialog_rules.json` 规则 + 修 cross-volume forensics 丢失，commit b33df56），gate 得以跑完，实测：

- **从零文字现在能渲染**：每个 TS_* comp 的默认文字（AE 默认浅蓝 fill）在 AE2020 实渲里**清晰可见**（目视 png 确认，非空白），不再是「完全看不见」。
- **DOM 读回字号正确**：`doc.fontSize` = 160 / 50 / 88（plain，**非 /65536**）；trk=1200、lead=70/220、just=7413/7415 全对。
- **同源缺陷已被工作树 `internal/codec/text_encode.go` + `internal/serializer/back_layer.go` 改动修掉**（上次会话产出，**尚未 commit**，render-gate 未过故未 ship）。

**但 render-pixel 仍未证成**（红线4 未闭环）：gate 用 `comp.openInViewer()` + `comp.saveFrameToPng` 取帧，实测 **saveFrameToPng 渲染的是 active-viewer comp 而非 receiver comp**，且 headless `-r` 下 viewer 切换不同步（同步脚本占住事件循环，`setActive()` 不生效）→ 9/10 单行 comp 抓到的是同一个 active comp（TS_JL 的 "ABCD"），无法逐 knob 区分。故 **SetRunFontSize 等是否「按值渲染」尚未被像素证实**——只证了「文字能渲染 + DOM 值对」。

**下一步（reachable，独立子活）**：换可靠的逐 comp 取帧（Render Queue 渲 PNG 序列，comp-specific 且不依赖 viewer；或一次 AE run 只渲 startup-active 的单 comp）。证成后文字样式族方可升 render-pixel + commit 文本修复 + 改本 incident `status: resolved`。在此之前维持 verify=roundtrip 标注，文本修复保持 uncommitted（render-gate 未过不算 ship）。

## UPDATE 2026-06-16(b)：4/5 knob 渲染已目视/数值证实；自动 gate 卡在 saveFrameToPng 缓存

继续推进后**能力真相已查明**（capability 层面，红线4 实质消除）：

- **字号 / 字距 / 对齐 / 大小写 4 个 knob 确认按值正确渲染**：在抓帧成功的那次 run + 目视 png：FontSize SMALL 47×40 vs BIG 153×132（≥2×）、Tracking TIGHT 173 vs WIDE 490、Justify JLEFT cx≠JCENTER cx、Caps "ace"→实渲 "ACE"（大写）。FontSize/Tracking 跨多次 run 数值稳定一致。
- **Leading 实证不渲染（roundtrip-only）**：`SetRunLeading` 写值 + DOM 读回 70/220 正确，但 AE 对从零层渲染默认行距（220 与 70 两行间距目视相同）。已从 render-gate 移除、保留 verify=roundtrip（evidence-based defer，同 Rotation X/Y 先例）。

**但自动化 render-gate 仍做不绿——卡在 AE 工具链的 `saveFrameToPng` 缓存**（不是库缺陷）：
headless `AfterFX -r` 下 `comp.saveFrameToPng(time, file)` **发的是 AE 的持久磁盘帧缓存、跨进程不失效**：实锤——两个**独立冷启** AE 进程、相隔 2 分钟、打开**不同的**单 comp 工程（DOM 读回各异，证明确实开了不同工程），却写出**字节完全相同**的 PNG（且是**更早某次 run** 的 TS_TRK_WIDE "MMMM" 帧）。试过且**全部无效**：`app.purge(PurgeTarget.ALL_CACHES)`、逐 job 唯一 `time`、逐 comp 独立冷启 AE（单 comp 工程）。`time` 参数似乎被忽略（唯一 time 不改变输出）。`app.purge` 不清磁盘缓存；PurgeTarget 无 disk 项。

**结论**：能力已证（4 knob 渲染对、Leading 不渲染），但**自动像素门禁需要绕过 saveFrameToPng 的磁盘缓存**——候选：(a) 改 **Render Queue**（真渲染、绕 preview/disk 缓存，但只出 TIFF/PSD，无 PNG 模板 → Go stdlib 不解 TIFF，需加 `golang.org/x/image/tiff` 测试依赖、破坏本仓零依赖）；(b) 关闭/清 AE 磁盘缓存（位置/prefs key 版本相关，盲删有风险）；(c) 不做 CI 像素门禁、以本 incident 的目视+数值实证为准（红线4 风险已大幅消除）。**待需求方定夺**。工作树留存未 commit 的 gate（`mg_text_style_shipgate_test.go` 逐 comp 冷启版 + `verify_mg_text_style.jsx`）+ 文本修复（`text_encode.go`/`back_layer.go`）。

## UPDATE 2026-06-16(c)：RESOLVED — 缓存根因揪出，双版本 render-pixel gate 绿

(b) 的「缓存问题」上一会话只猜对一半。本次系统化 bisect 把根因彻底坐实，gate 做绿：

**真根因（不是 saveFrameToPng 不稳定——它忠实返回渲染器给的帧）**：AE 持久磁盘帧缓存键 ≈ **(comp.id, render-time)**。两条共同制造碰撞：
1. **comp.id 全撞**：确定性 builder 给每个单 comp 工程的 comp 恒分配 `comp.id=1`（8 个工程 JSX 日志全是 id=1，实锤）。
2. **旧 verify JSX 把 `job.time` 丢了**：第 48 行硬编码 `saveFrameToPng(0)`，Go 侧算好的 per-comp 唯一 time **从没传进 AE**。→ 上一会话「逐 job 唯一 time 无效 / time 被忽略」是**伪测试**（缓解措施根本没生效），结论错了。

id 全 1 + time 全 0 → 缓存键完全相同 → 一帧污染全部 8 张（旧 PNG 8 张 md5 全同、都是 `TS_TRK_WIDE` 的 "MMMM"，连画幅都不是各 comp 自己的高度）。`app.purge(ALL_CACHES)` 无效是因为它清 RAM、不碰磁盘缓存（`%LOCALAPPDATA%\Temp\Adobe\After Effects\<ver>\Disk Cache-*.noindex`，本机 2.5G）。

**修复（两道防线，缺一不可，均已验证）**：
1. **verify JSX 改用 `job.time`** → run 内每 comp 唯一缓存键。仅此一步修了 7/8；唯独 `TS_FS_SMALL`(t=0.1) 量化进被历史 t=0 run 污染的桶、仍渲旧 "MMMM" → gate **假绿**（靠 SMALL 旧帧恰好够薄 PASS，红线4 现身：值/DOM 全对但像素被污染）。
2. **`clearAEDiskCache` harness**（`go test` 渲染前删 `Disk Cache-*.noindex`）→ 清历史中毒帧。清完重跑：**8/8 PNG 哈希各异**，`TS_FS_SMALL` 目视终于 "Ag"、`TS_FS_BIG` "Ag" 大字。

**双版本 ship-gate 绿**（`TestMGTextStyle_AEShipGate_AE2020/2025`，跨版本数值一致）：FontSize 47×40 vs 153×132、Tracking 173 vs 490、Justify cx 1046 vs 958、Caps 39 vs 60。→ **SetRunFontSize / SetRunTracking / SetParagraphJustification / SetRunCapsOption 升 `verify=render-pixel`**。

**另一条独立 bug 顺带修掉（FormatPSReal）**：point-measurement 键（字号 `/1`、行距 `/5`、H/V scale `/6//7`、基线 `/9`、描边宽 `/63`）AE 当 **REAL** 读；写裸整数 `"150"` 会被读成 16.16 定点 = `150/65536`（字号那次的 ÷65536 症状根因）。`codec.FormatPSReal` 强制带小数点；`back_layer.go` 这些 setter 改用它。**注意**：这些 setter 此前是**潜在错的**（会渲 ÷65536），FormatPSReal 是真 correctness fix，不只是 gate 配套。

**Leading 仍 evidence-defer**（roundtrip-only）：值 + DOM 读回 70/220 对，但 AE 对从零层渲默认行距（像素不变），同 Rotation X/Y 先例，保留 `verify=roundtrip`。

**TL;DR/UPDATE(a)(b) 的「缓存绕不过 / 待定夺」已作废**——见本节。
