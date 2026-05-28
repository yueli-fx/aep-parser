---
when_to_read: looking up a chunk ID constant in rifx; debugging "field reads OK on file X but nil on file Y"; adding a new chunk ID
applies_to: [rifx, chunk-id, tdb4, case-sensitivity, parser]
last_updated: 2026-05-27
---

# Chunk ID 大小写敏感 — Tdb4 ≠ tdb4

RIFX chunk ID 是 4 字节 ASCII，**大小写敏感**。`Tdb4` (uppercase, legacy) 跟 `tdb4` (lowercase, modern) 是**两个不同的 chunk**。同义但不同名。

## 现状

`internal/rifx/rifx.go` ChunkID 常量表里**两个都声明**。parser 必须同时认。

## How to apply

- 加新 chunk ID 常量前，**grep rifx.go 看是否已有同名 lower/upper 变体**
- 解析时遇到旧文件读不到字段、新文件 OK（或反过来）→ 先查是不是 case 变体没都认上
- 写 `rifx.Find*` 调用时，调用方传哪个 ID 决定能匹配到哪个变体

## Why

AE 在某些版本切换了 chunk 命名 case，旧文件保留 legacy 名。RIFX 规范没强制 case 中立，二进制 byte 直接比对，所以两个变体共存。
