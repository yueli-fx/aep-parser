---
status: obsolete
when_to_read: editing parse_keyframe.go or write_keyframe.go; adding new keyframe property type; debugging "ease/value at wrong offset"
applies_to: [keyframe, layout, spatial, non-spatial, layoutFor, parse_keyframe, write_keyframe]
last_updated: 2026-05-27
---

# 关键帧字节布局两种 — 必须走 layoutFor 分发

Keyframe block 字节布局**不是统一的**。两种 layout 互斥，按 property 类型决定。**不要在 caller 端手算 offset**。

## 两种 layout

**Spatial-style**：4D color / Position / Anchor (3D)
- block 头部 byte `0x07 = 0x07` 或 `0x01 + dims≥2`
- scalar ease 在 0x18 / 0x20 / 0x28 / 0x30
- values 在 0x38
- bytes-per-keyframe = 0x38 + 3·N·8

**Non-spatial**：Opacity (1D) / Scale (3D) / Mask Feather (2D)
- block 头部 byte `0x07 = 0x00`
- values 在 0x08
- per-component ease 在 `0x08 + (N+i)·8`
- bytes-per-keyframe = 0x08 + 5·N·8

## How to apply

- 加新 keyframe property 类型 → 走 `layoutFor(header07, dims)` 集中分发，**不要新加 if/else 链**
- 改 keyframe 写回 → 用同一个 `layoutFor` 算 offset，保证 read/write 对称
- 调试 ease/value 错位 → 第一步 dump block 头 byte `0x07` 确认是哪种 layout

## Why

早期实现 caller 端到处手算 offset → 改一个 layout 漏改另一个 → silent 错位。`layoutFor` 是唯一权威分发点。
