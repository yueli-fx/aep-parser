---
status: active
when_to_read: replicating / copying an existing layer's masks from-scratch and every mask reads as the SAME unit square (anchors all 0/1) while the original clearly has distinct, positioned masks; Mask.Vertices looks normalized not pixel; reconstructing real mask geometry from a parsed mask; deciding what coordinate the parser surfaces for a mask outline; copying many masks (tearing slices) faithfully via AddMask; wondering why mask feather/opacity/expansion don't survive a from-scratch mask copy
applies_to: [mask, mask-vertices, shph, bbox, bounding-box, read-side-gap, normalized-ldat, fraction, coordinate-space, mask-replication, copyMasksFromOriginal, AddMask, mask-feather, mask-opacity, mask-expansion, booyah, tearing-slice, render-fidelity, flightdeck/showcase/booyah-clone/gen_glitch_text.go]
last_updated: 2026-06-22
resolved_by:
---

# Mask 真几何在 shph bbox，`Mask.Vertices` 只 surface 归一化 ldat（读侧 gap）

## Signature
- symptom: 读一个真实层的多个 mask，`Mask.Vertices` 全部返回**同一个 unit square**（anchor ∈ {0,1}，如 (1,0)(0,0)(0,1)(1,1)），无论原版那些 mask 在画面上明显是不同位置/大小的切片；以为 parser 漏解或 mask 都一样
- error_type: —
- where: internal/serializer/parse_mask.go decodeMaskVertices（只读 ldat）；消费侧 = 任何从 `Mask.Vertices` 重建几何的代码
- trigger: 复刻/拷贝一个层的 mask（booyah ⑩ 131 个撕裂切片 mask 首撞）

## 根因
AE 存一条 mask 轮廓 = **两段**：(1) `shph` 头 @0x04 起的 **4 个 BE float32 bbox** = `[L, T, R, B]`，是**源像素空间的分数**（0..1）；(2) `ldat` 里的 bezier 顶点，是**相对该 bbox 归一化**的坐标（一个矩形 mask → bbox 内的 unit square）。`decodeMaskVertices` 只解 ldat 的归一化顶点、**不应用 shph bbox**，所以每个矩形 mask 都读成同一个 unit square；真正区分各 mask 的位置/尺寸的 bbox 被丢在 `Mask.ShphRaw` 里没解。

## 复刻解法（消费侧，库未改）
从 `Mask.ShphRaw` 自己解 bbox 重建像素几何：
```
d := m.ShphRaw                       // 24 字节
L := f32(d[0x04]); T := f32(d[0x08]) // BE float32, 源空间分数
R := f32(d[0x0C]); B := f32(d[0x10])
// 像素矩形（源 = 该层 source item，booyah 全是 1920×1080 solid）
rect := BezierPath{Vertices: {{L*w,T*h},{R*w,T*h},{R*w,B*h},{L*w,B*h}}, Closed:true}
AddMask(layer, name, rect)           // AddMask 再 ÷源尺寸回分数 → 与原版 bbox byte-exact
```
booyah ⑩ 实测：全 131 个 mask 都是 4 顶点 Add-mode 轴对齐矩形，bbox round-trip `maxErr=0`，双版本 AE 接受不 drop。代码见 `flightdeck/showcase/booyah-clone/gen_glitch_text.go` `maskBBoxRect`/`copyMasksFromOriginal`。

## 边界 / 未覆盖
- **只对轴对齐矩形 mask 忠实**：bbox 是 AABB，丢掉旋转/非矩形 bezier 的形状（ldat 归一化顶点理论上能恢复非矩形，但需同时套 bbox 变换 + 厘清 ldat 三元组真实布局 `[anchor, out-ctrl, next-in-ctrl]`，本次未做——booyah 全是矩形所以没碰）。
- **mask feather / opacity / expansion 未复刻**：`copyMasksFromOriginal` 只搬 geometry + mode + inverted。booyah ⑫ 终帧对照发现 ⑪ flare mask 在 clone 渲成**锐利满不透明亮矩形** vs 原版柔和 → 怀疑原 mask 带 feather/低 opacity。补法:加 `m.SetFeather`/`m.SetOpacity`/`m.SetExpansion`（getter 侧 `Mask.Feather/Opacity/Expansion` 已可读）。属 render-fidelity polish。
- **可选库改进**:给 `Mask` 加一个「de-normalized 像素顶点」accessor（套 shph bbox），免每个消费方各自解 ShphRaw；本次走消费侧绕过，库读语义未动。

## Cases
- 2026-06-22 booyah ⑩：131 撕裂切片 mask 复刻首撞；shph bbox 解法 byte-exact round-trip + 双版本 AE 接受。⑫ 终帧暴露 mask feather/opacity 未复刻的 render delta。
