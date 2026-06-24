# ⚠ Transform Group 属性比 py-aep 少 = AE 默认值省略，不是 parser bug

SUMMARY: Transform Group 属性比 py-aep 少 = AE 默认值省略，不是 parser bug
READ WHEN: 我们解出的 Transform Group 属性比 py-aep golden 少；以为 parser 截断/desync；ADBE Orientation 在 tree 里是空 group

---

## 结论（先纠一个误判）

最初 `transform_separated.aep` / `transform_unseparated.aep`（3D "Gray Solid 1"）解出的
Transform Group 比 py-aep golden 短一大截，我一度记成"parser 截断 / desync /
Orientation 误解析"。**raw chunk dump 推翻了这个判断 —— 不是 bug。** 两件事叠加：

### 1. AE 不持久化未修改的默认属性（核心）

raw Transform Group 里**物理上只有非默认（被改过）的属性**：

| fixture | 文件里实际存在的 Transform 子项 |
|---|---|
| separated | Position, Position_0, Position_1, Orientation(otst), Rotate X, Rotate Y, Envir |
| unseparated | Orientation(otst), Rotate X, Rotate Y, Envir |

`Anchor Point / Scale / Position_2 / Rotate Z / Opacity` 等**根本没写进 .aep**（值在默认，AE 省略）。

py-aep 会**合成（synthesize）**完整的 canonical transform schema（把省略项按默认值补出来，
标 elided）。我们**故意不合成**（见 `scene_property_flags.go` 的 `Elided()` 注释 "we don't
synthesize anything yet"）。所以我们的 flat/tree 只列文件里真实存在的属性 —— 这是**预期行为
+ 已知 parity gap（property synthesis 暂搁）**，不是解析错误。

> 验证手法：`go run ./tmp_debug/dump_tdgp <fixture>` 打印 tdgp 原始 tdmn→payload 序列。
> 对照 py-aep golden 的 properties 列表，差集就是 AE 省略的默认项。

### 2. `ADBE Orientation` 用 otst 包装（次要，by design）

3D 层的 Orientation payload 是 `LIST:otst`（含 `tdbs` 值 + `otky` 关键帧），不是裸 `tdbs`：

- **flat parser**：`parse_properties.go::collectFromGroup` 现有专门的 `case rifx.IDOtst`
  → `parseOrientationProperty`（2026-06-02 落地）。静态值修对：Components=3、cdat 按
  **小端**解（`decodeCdatValueLE`），fixture `orientation_5_0_0`/`orientation_0_279_0` 验证
  `[5,0,0]`/`[0,279,0]`。动画值修对：keyframe 的 X/Y/Z 从 **otky/otda**（大端，每 otda 一个 kf）
  取，fixture `orientation_with_keyframes` 验证 `[5,0,0]`→`[0,0,0]`。
  ⚠ **仍缺**：animated orientation 的 easing/tangents 是用 Components=1 的旧 layout 解的，未校验，
  需要时再修（值已对，影响的是缓动）。
- **tree builder**：`scene_property_group.go::addNamedChildren()` 的 `default:` 分支**有意**
  把 otst/parT/mrst 这类未知 wrapper 包成 opaque 空 group（注释写明 "don't descend further"）。
  所以 tree 里 Orientation 显示成 0-child group —— **by design**，非 bug。flat 列表才是值的来源。

## 对分离 readers 的影响

`DimensionsSeparated` / `IsSeparation*` readers 本身正确。但这两个 fixture **无法覆盖
Position_2（文件里没有）和 unseparated leader（Position 整个被省略）**的真实字节路径 ——
见 `property_separation_test.go` 的 NOTE，已退化用 bare property 验 match-name 逻辑。

## 真正可做的后续（都不是"修 bug"）

1. **Property synthesis**（大 feature，暂搁）：补出 AE 省略的默认 transform 属性，对齐 py-aep
   的完整 schema + `Elided()`。这是唯一能让我们的 transform group "看起来和 py-aep 一样长"的路。
2. **otst Orientation fidelity**：静态值 + keyframe 值已修（见上 § 2）。剩 animated orientation
   的 easing/tangents（旧 1D layout）+ tree 里把 otst 当叶子而非空 group。需要时另立。
