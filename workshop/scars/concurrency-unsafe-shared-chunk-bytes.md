---
when_to_read: writing concurrent code touching aep.Project / Composition / Layer / Property / Keyframe; debugging race conditions in Set* calls
applies_to: [concurrency, thread-safety, rifx-chunk-data, mutate, Set*]
last_updated: 2026-05-27
---

# Project/Composition/Layer/Property/Keyframe 全部 mutate 共享 chunk bytes

所有 Go 类型（Project / Composition / Layer / Property / Keyframe）都是 `rifx.Chunk.Data` 字节的视图 / mutator。**Set\* 直接改共享底层字节**，没有 copy-on-write。

- 多 reader 无锁可行（只读字节）
- 任何 Set\* 调用必须由**调用方自己同步**

## How to apply

- 在 goroutine 里调 Set\* → 加锁
- 同时多个 reader 可以并发跑 getter / 读 field，安全
- 不要假设有内部 mutex —— 没有

## Why

length-preserving splice 是性能优化（不重新 alloc + 不重建 chunk 树）。代价是 mutate 共享。Library 不内置锁是为了让调用方按需选粒度。
