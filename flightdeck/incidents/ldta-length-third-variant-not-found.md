---
status: active
when_to_read: hypothesizing AE 24+/25 added bytes to ldta tail; planning a "new long ldta" RE pass
applies_to: [ldta, wave-2, negative-finding, re-strategy, coverage]
last_updated: 2026-05-21
---

# AE 24/25 ldta 第三种长度 — 找不到 (negative finding)

## 症状

Wave 2 调研期间假设 AE 22+ / 24+ 可能加长 ldta（在 164 字节后追加字段）。穷举测试若干 fixture 想找新字段位置。

## 结论

AE 25 实测 ldta 仍写 164 字节，跟 AE 23+ 一致。未观察到第三种长度。所有 Wave 2 候选字段（trackMatteLayer / LightKind / timeRemapEnabled / displayStartTime）实际都在 *既有* 164 字节范围内（`@0x88` / `@0xA0`），或者在 cdta tail（DisplayStartTime `@0xA4/@0xA8`），不在 ldta 加长部分。

## 教训

- "新版本 AE 加长 X chunk" 是一种常见 RE 假设，但实际经常不发生 —— Adobe 倾向于复用现有 padding 字节而不是扩 chunk。
- 找新字段时先 *穷尽 RE 现有长度内的字节*（很多 0 字节其实是未 RE 的 flag），再考虑 chunk 加长。
- 找不到时也要写下来：`coverage-detail.md` 标 `❌ 未发现` 而不是 ⚠ deferred，否则下次有人又会去找一遍。
