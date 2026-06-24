# ⚠ 手动 kerning 首次启用 = 结构性添加

SUMMARY: 手动 kerning 首次启用 = 结构性添加
READ WHEN: implementing or debugging Layer.SetManualKerning; deciding between length-preserving splice and structural-add for text setters

---

## 症状

`Layer.SetManualKerning(values)` 在某些 layer 上 refuse with "no manual-kerning slot present"，即使我们给的 values 长度正确。

## 根因

AE 只在 *已经有过* 非零 NoAuto kerning 值的时候才 emit btdk `/1/1[0]/0/8` 子树。Layer 默认 AutoKernType=Metric/Optical 的话整个 `/8` 不存在 —— "添加 slot" 是结构性变更（要扩 btdk 树容器 + N+1 entry array + sentinel），破坏 length-preserving 约束。

## 教训

- "first-time enable" 跟 "modify existing value" 是两种事：第一种是 structural，第二种是 length-variable splice。
- Setter 只支持第二种。需要 enable 时让用户先在 AE UI 设一次非零 kerning，AE 自己 emit slot 后我们再改值。
- 检测：`len(TextSource.ManualKerning) == 0` 即 slot 不存在。

> **更新（2026-06-12）**：「无 scripting API enable kerning」是本 incident 写就时（2026-05-19）
> 的认知。**AE 24+ 有 `TextDocument.kerning` 统一属性**——`td.kerning = -50` 能 materialize
> `/1/1/0/0/8` per-char slot（`re_text_ae24_more.jsx` / `re_text_kern_resize.jsx` 实证）。
> 故造 kerning fixture 不再必须手动 GUI。`Layer.SetManualKerning`（本 incident 主题，等长改值）
> 仍只支持「modify existing」。反向：整文本替换时 AE **丢弃**整个 kerning slot，详
> [[text-btdk-length-variable-write-scoping]] § 修法 v4。
