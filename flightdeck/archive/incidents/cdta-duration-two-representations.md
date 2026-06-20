---
status: obsolete
when_to_read: resolving render-queue time spans; comparing comp Duration against py-aep golden; debugging "comp duration off by ~1.5×" on synthetic fixtures
applies_to: [composition, cdta, duration, work-area, render-queue, py-aep-parity]
last_updated: 2026-06-01
---

# cdta 有两套 duration 表示，可能互相不一致

cdta 同时携带**两种** comp duration：

1. **frame-count**：`cdtaDuration @0xB0`（uint32 帧数）+ mirror `@0xB8`。我们的 parser 用这个：`Duration = frames / fps`。
2. **time-ratio**：`duration_dividend / duration_divisor`（@0x14 区，秒）。py-aep 用这个。

真实 AE 存盘文件里两者一致，所以平时无感。但 **py-aep 的 synthetic fixture（脚本批量改过的）里两者会偏离**——例如 `samples/models/renderqueue/numItems_1.aep`：time-ratio = 10s（py-aep golden），frame-count @0xB0 = 360f/24fps = **15s**（我们）。work area 同理（@0x1C/@0x24 dividend/divisor）。

## How to apply

- **不要拿 py-aep RQ golden 的 `timeSpanDuration` 秒数硬断言我们的值**当 time_span_source = LENGTH_OF_COMP / WORK_AREA_ONLY 时——它依赖 comp duration/work-area，而 synthetic fixture 这俩字段不自洽。改为断言**解析逻辑**（`TimeSpanDuration == comp.WorkAreaEnd - comp.WorkAreaStart`）或用 **CUSTOM source** fixture（值直接来自 settings ldat dividends，与 comp 解析无关）做精确断言。详 `parse_render_queue_test.go`。
- 若将来要让 `Composition.Duration` 与 py-aep 严格对齐，需改读 time-ratio dividend/divisor（@0x14 区），但这会动 Stable comp 字段语义 + 全 fixture 回归——非 RQ slice scope，单独评估。

## Why

frame-count 是"显示用"派生量，time-ratio 是"权威"时间。AE 自己写盘时同步两者；外部脚本（py-aep save）只更新 time-ratio，留下 stale frame-count。两条 RE 路径选了不同字段，于是 synthetic fixture 暴露分歧。相关：[[tickrate-per-composition]]。
