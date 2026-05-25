---
when_to_read: implementing or debugging Layer.SetManualKerning; deciding between length-preserving splice and structural-add for text setters
applies_to: [text, kerning, btdk, structural-vs-splice, length-preserving, write-refused]
last_updated: 2026-05-19
---

# 手动 kerning 首次启用 = 结构性添加

## 症状

`Layer.SetManualKerning(values)` 在某些 layer 上 refuse with "no manual-kerning slot present"，即使我们给的 values 长度正确。

## 根因

AE 只在 *已经有过* 非零 NoAuto kerning 值的时候才 emit btdk `/1/1[0]/0/8` 子树。Layer 默认 AutoKernType=Metric/Optical 的话整个 `/8` 不存在 —— "添加 slot" 是结构性变更（要扩 btdk 树容器 + N+1 entry array + sentinel），破坏 length-preserving 约束。

## 教训

- "first-time enable" 跟 "modify existing value" 是两种事：第一种是 structural，第二种是 length-variable splice。
- Setter 只支持第二种。需要 enable 时让用户先在 AE UI 设一次非零 kerning，AE 自己 emit slot 后我们再改值。
- 检测：`len(TextSource.ManualKerning) == 0` 即 slot 不存在。
