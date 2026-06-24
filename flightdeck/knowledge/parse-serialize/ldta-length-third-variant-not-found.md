# ⚠ AE 24/25 ldta 第三种长度 — 找不到 (negative finding)

SUMMARY: AE 24/25 ldta 第三种长度 — 找不到 (negative finding)
READ WHEN: hypothesizing AE 24+/25 added bytes to ldta tail; planning a "new long ldta" RE pass

---

## 症状

Wave 2 调研期间假设 AE 22+ / 24+ 可能加长 ldta（在 164 字节后追加字段）。穷举测试若干 fixture 想找新字段位置。

## 结论

AE 25 实测 ldta 仍写 164 字节，跟 AE 23+ 一致。未观察到第三种长度。所有 Wave 2 候选字段（trackMatteLayer / LightKind / timeRemapEnabled / displayStartTime）实际都在 *既有* 164 字节范围内（`@0x88` / `@0xA0`），或者在 cdta tail（DisplayStartTime `@0xA4/@0xA8`），不在 ldta 加长部分。

## 教训

- "新版本 AE 加长 X chunk" 是一种常见 RE 假设，但实际经常不发生 —— Adobe 倾向于复用现有 padding 字节而不是扩 chunk。
- 找新字段时先 *穷尽 RE 现有长度内的字节*（很多 0 字节其实是未 RE 的 flag），再考虑 chunk 加长。
- 找不到时也要写下来：`coverage-detail.md` 标 `❌ 未发现` 而不是 ⚠ deferred，否则下次有人又会去找一遍。

## 补充 (2026-06-09)：长度确实变过一次，但在 2023（不在 24/25）

本条原结论"AE25=164=AE23+、无第三种长度"**仍成立**（24/25 没在 164 之上再加长）。但当年只搜 24+，**漏了 2023 这一档的 160→164 加长**——164 不是全版本基线：

- **原生 AE2020 ldta = 160 字节**（实测 `2020_dummy_comp` 160×11、`tmp_debug/ellipse_ae2020_native` / `test_data/renderer_ae2020_r0/r1` 均 160×12）。
- **AE2025 ldta = 164 字节**（`selection_both_layers` 164×13）。
- 即 AE2023+ 在 ldta 尾加了 **4 字节**，落在 `@0xA0`(160..163) —— 正是本文上面列的 Wave-2 字段位（它在 164 的**最后 4 字节**里，≤2022 的 160-byte ldta 根本没这段）。

旁证来自逆向外部 AE 降级器（详 [2026-06-09-ae-version-downgrader-re.md](2026-06-09-ae-version-downgrader-re.md)）：它对 source≥2023→target≤2022 **正是砍掉 ldta 尾 4 字节**（164→160），且 ship-gate 实测 AE2020 无损接受降级输出 → 反证 160 = ≤2022 原生长度。精确 pivot=2023 是该工具断言（本仓 2021–2024 模板都空、0 ldta，未独立量到原生分层 2021/2022/2023；已量到 2020=160、2025=164）。

**对 parser 的含义**：ldta `@0xA0` 尾 4 字节是 **2023+-only** 字段；读 ≤2022 文件时该 offset 不存在，按 ldta 实际 size 读、勿对低版本 over-read。
