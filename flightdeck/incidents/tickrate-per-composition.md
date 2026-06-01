---
status: active
when_to_read: writing keyframe / marker time arithmetic; debugging "time off by factor"; adding new time-based setter
applies_to: [tickrate, time-arithmetic, keyframe, marker, composition, cdta]
last_updated: 2026-05-27
---

# TickRate 是 per-composition 的，不是全局常量

**不要假设统一 8000**。`Composition.TickRate` 由 cdta `@0x08` / `@0xA8` 推导，**每个 comp 自己一份**。Keyframe / Marker 时间换算都必须用 owning comp 的 TickRate。

`aeLegacyTimeBase = 8000` 只是 cdta 缺失时的 fallback，不是 "正确值"。

## How to apply

- 写新 setter 涉及时间字段 → 通过 layer/property → owning comp 拿 TickRate，**不要硬编码 8000**
- 测试 fixture 时间值对不上 → 第一步 check fixture comp 的 cdta @0x08/@0xA8 实际 tickrate
- Marker / Keyframe 时间字段都用 ticks，不是秒；除以 TickRate 才是秒

## Why

不同 AE 版本 / 不同 comp duration / 不同 fps 可能算出不同 tickrate。早期 boltframe 假设 8000 导致跨 fixture 时间漂移。
