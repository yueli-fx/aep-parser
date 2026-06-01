---
status: active
when_to_read: 3D 层 / 分离维度 fixture 的 Transform Group 属性少了一截；ADBE Orientation 解成 0-child group；Position_2/Scale/Rotate Z/Opacity 读不到
applies_to: [parser, property-stream, transform-group, 3d-layer, separated-dimensions, coverage-gap]
last_updated: 2026-06-02
---

# 3D 层 Transform Group 解析截断（Orientation 起 desync）

## 症状

py-aep fixture `samples/models/property/transform_separated.aep` / `transform_unseparated.aep`（都是 3D 的 "Gray Solid 1"）经我们 parser 后，Transform Group 比 py-aep golden 少一大截：

| | golden（py-aep）Transform Group 子项 | 我们 parser 解出 |
|---|---|---|
| separated | AnchorPoint, Position, Position_0/1/2, Scale, Orientation, RotateX/Y/Z, Opacity, Envir (11) | Position, Position_0, Position_1, **Orientation(误判为 0-child group)**, RotateX, RotateY, Envir (7) |
| unseparated | 同上（Position_0/1/2 也在，只是 dimSep=false） | **Orientation(group), RotateX, RotateY, Envir (4)** — Position 之前的全丢 |

两个共性：
1. `ADBE Orientation` 被解析成 **AEPropertyGroup（0 children）**，正确应是 3-component leaf Property。
2. Orientation 之后的 `Rotate Z / Opacity` 丢失；separated 的 `Position_2`、两者的 `Anchor Point / Scale` 也丢。
3. unseparated 更严重：连 `Anchor Point / Position / Position_0..2 / Scale` 都没解出，flat 列表直接从 Orientation 系列 + 3D 材质项开始。

非 3D 普通层不受影响（`re_batch.aep` 的 Position 正常解出）。

## 影响

- `DimensionsSeparated` / `IsSeparationFollower` 等 readers 本身正确（separated Position dimSep=true、Position_0/1 follower 验证通过），但**这两个 fixture 无法覆盖 Position_2 / unseparated leader=false** 的真实字节路径 —— 见 `property_separation_test.go` 的 NOTE，已退化用 bare property 验 Position_2 的 match-name 逻辑。
- 3D 层的 transform 写路径（若将来做）会踩同一 desync，必须先修这里。

## 待查方向（未定位根因）

- `ADBE Orientation` 在 3D 层的 chunk 布局：是不是它带了 om-s / 子结构让 `parse_properties` 误判成 group？Orientation 是 3D 专属，普通层没有 → 解释了为何只在 3D 层炸。
- desync 是从 Orientation 开始还是更早（unseparated 连 Position 都没了，可能更早就偏移）。建议用 `tmp_debug/list_props` 或 `parse_btdk` dump 这两个 fixture 的原始 chunk 序列，对着 `parse_properties.go` 的 group/leaf 判别逻辑数。
- 怀疑点：separated 维度 follower（Position_0/1/2）+ Orientation 的混合让属性流游标错位。

## 现状

记录为 coverage gap，未修。分离 readers 已 ship（R-only，[coverage.md](../plans/coverage.md)）。修这个属于 parser property-stream 的独立 arc。
