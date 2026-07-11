# ShapeLayer transform 必须写入 runtime stream

SUMMARY: ShapeLayer 写出时会从 runtime shape tree 重新 lower；recipe 和 migration 必须填充 `ShapeLayer.Transform()`，不能只替换 chunk-side Transform Group。
READ WHEN: 修改 ShapeLayer transform、recipe shape 编译、shape migration/rebuild，或出现 profile 中 Position 正确但迁移输出变成 `[0,0,0]` 时
RECHECK WHEN: ShapeLayer 不再通过 `syncShapeLayerChunks` 从 runtime tree 统一重新 lower 时

---

`NewShapeLayer`/`WrapShapeLayer` 会建立 runtime `LayerTransform`，并把 shape layer 标为 dirty。`WriteAEP` 前的 `syncShapeLayerChunks` 会调用 `lowerShapeLayer`，以 runtime shape tree 和 runtime transform 重建整个 Layr 内容。

因此，对普通 text/solid/null/precomp layer 有效的 `aep.SetLayerTransform(layer, transform)` 不足以修改 dirty ShapeLayer：它替换的是已有 chunk 中的 Transform Group，随后可能被 runtime 默认值重新 lower 覆盖。

正确做法：

- recipe 编译拿到 `*aep.ShapeLayer` 后，直接填充 `shapeLayer.Transform()` 的 Position/Scale/Rotation/Opacity/AnchorPoint streams。
- migration 重建 shape layer 时，从 source profile 填充同一个 runtime transform；静态值和关键帧走同一 helper。
- 测试必须先让源 recipe/profile 拥有非默认 transform，再迁移并比较输出。只用默认 `[0,0]` 会形成假绿。

2026-07-11 的回归即由此暴露：recipe 修复后源 profile 首次正确包含中心 Position，但 migration 的 shape rebuild 只恢复 shape primitive/filter，漏掉 runtime transform，导致迁移输出回到 `[0,0,0]`。修复是在 `aepmigrate` 中统一使用 `populateLayerTransformFromProfile` 填充普通 layer 的新 transform 或 ShapeLayer 的现有 runtime transform。
