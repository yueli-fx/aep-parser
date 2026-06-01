---
status: active
when_to_read: implementing or debugging Layer.SetAlternateSource; touching Essential Properties slot / blsi chunk; comparing JSX-saved vs Go-built media-replacement layers
applies_to: [layer, alternate-source, blsi, jsx-quirk, media-replacement, ae24]
last_updated: 2026-05-21
---

# setAlternateSource 自动包 wrapper precomp（AE 脚本 quirk）

## 症状

JSX 里 `layer.setAlternateSource(otherCompItem)` 后，dump .aep 看到 layer 的 blsi 指向的不是 `otherCompItem` 本身，而是一个新生成的 wrapper precomp：名字 `<slotName>_<origName> 2`，进了项目面板 "媒体替换合成" 文件夹。

## 根因

AE 24+ Media Replacement (Essential Properties slot) 的设计：脚本 set 时自动包一层 precomp（让你能在 wrapper 里独立调整 alternate source 的 transform/effect），原 item 不变。这是 *AE 的产品行为*，不是 bug。

## 教训

我们的 Go setter `Layer.SetAlternateSource(item)` **不** 包装 wrapper，直接写 `item.ItemID()` 到 blsi。两个不同语义：

- **JSX 脚本**: "我要 alternate source = X" → AE 包 wrapper（保护原 item，方便后续独立编辑）
- **Go setter**: "我要 blsi 字节 = X 的 id" → 直接写，调用方自己负责管理 wrapper

两个都对，但要在 docs/layer.md `SetAlternateSource` 段说明区别，否则调用方 RE 时会以为我们漏写了 wrapper 步骤。

## 关联约束

`Layer.HasAlternateSourceSlot() == false` 时 setter refused —— 该 layer 没被 `addToMotionGraphicsTemplateAs` promoted 过，缺 slot。"首次创建 slot" 是结构性变更，不实现。
