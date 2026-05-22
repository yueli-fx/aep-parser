# 架构 ground truth

动代码前必读。架构边界、文件地图（当前现状）、length-preserving 约束、关键帧字节布局、测试惯例、已知非完美区。

> 公开 API 文档见 `docs/`，AE 字段覆盖矩阵见 `workshop/plans/coverage.md`。

## 一句话

Go 原生 `.aep` 项目文件解析器 + length-preserving 写回库。基于 RIFX (Big-Endian RIFF) 二进制格式逆向工程，**无需 AE 运行实例**。

## 兼容范围

**读取下限**：AE 2020 (CC 17.0)。任何 AE 2020+ 写的 .aep 必须可解。AE 24+ 引入的新字段在旧文件缺失时 graceful degrade 为 nil / 默认值，**不报错**。

**写回支持**：AE 2020 共有字段全部 length-preserving 可写；AE 24+ 新字段要求文件由 AE 24+ 写出（旧文件没对应 chunk slot）。

**版本稳定性实证**：cdta 内部布局 AE 2020 ↔ AE 2025 bit-for-bit 完全一致（见 `workshop/board.md` 归档 / `re_cdta_ae2020.aep` 跟 `re_cdta_probe.aep` diff）。不需要任何版本-条件分支。

**RE 工具链双轨**：
- `test_data/re_*.jsx` — AE 2020 ScriptingAPI 写的 fixture
- `test_data/re_*_ae24.jsx` — AE 24+ ScriptingAPI 写的 fixture（24+ 才解封的字段，如 fontCapsOption / strokeOverFill）

参考资料常驻仓内：
- `after-effects-scripting-guide/` — docsforadobe 镜像
- `Types-for-Adobe/AfterEffects/{8.0..26.0}/` — 按 AE 版本目录的 TS 类型定义。**判断"某字段在哪个版本引入"看这里最准**

## 架构数据流

```
io.ReadSeeker
   ↓
internal/rifx        ── 通用 RIFX chunk 树解析 (磁盘字节 ↔ Chunk 树)
   ↓
internal/aep/parse*  ── AEP 语义层 (Chunk 树 → Project / Composition / Layer / ...)
   ↓
*aep.Project ──┬── *aep.WriteAEP(io.Writer)        二进制写回 (length-preserving)
               └── *aep.WriteJSON(io.Writer)       JSON 单向导出 (无反序列化)
```

**关键边界**：`internal/rifx` 不知道 AEP 语义，只懂 RIFX 容器。`internal/aep` 不直接读字节流，只通过 `rifx.Chunk` 操作。两层职责分清，不交叉。

## 文件地图

### 入口

| 文件 | 职责 |
|---|---|
| `main/main.go` | Demo CLI：解析 → 打印关键帧 → 演示 SetTime/SetValue → 重解析验证 |
| `tmp_debug/*` | RE 工具集（16 个）：`dump_cdta` / `dump_ldta_trackmatte` / `parse_btdk` / `list_items` / `list_props` / `list_item_chunks` / `probe_effects` 等。详 `playbooks/verify.md` |
| `scripts/split_*.py` | 一次性脚本：拆 test (12 文件) / 拆 parse_text (3 文件) / 拆 write (3 文件) + Layer.SetText 并入 write_layer |

### internal/rifx

| 文件 | 职责 |
|---|---|
| `internal/rifx/rifx.go` | RIFX 容器读写。Chunk 树解析/序列化、ChunkID 常量集中表（含 PRin / prin / prda for renderer）、`U8/U16/U32/Text` 越界安全访问器、`FindFirst/FindFirstList/FindAllList` 查找助手 |

### internal/aep — 解析与类型（25 个 .go 文件）

**类型 + 公共访问器**

| 文件 | 行数 | 职责 |
|---|---|---|
| `types_core.go` | 705 | 核心模型：`Project / Composition / Footage / Folder / Layer / Property / Keyframe / TemporalEase / InterpType`。`Composition.{Renderer, ResolutionFactor}` 在此声明。**含 5 个 transform getter** (AnchorPoint/Position/Scale/Rotation/Opacity) |
| `types_features.go` | 307 | 扩展类型：`Effect / Marker / Mask / MaskVertex / MaskPathKeyframe / MaskMode / ShapePath / LayerQuality / BlendingMode / TrackMatteType / AutoOrientType` + String() |
| `layer_accessors.go` | 814 | Camera/Light/Material/Geometry/Transform 类型化访问器（getter + setter）+ `LightKind / MaterialCastsShadowsMode` 枚举 + MatchName 常量集中表。**53 个 typed setter** 都在这里 |
| `text_types.go` | 326 | 文本类型 + 8 enum 三件套（`TextJustification / TextCapsOption / TextAutoKernType / TextLineJoinType / TextDigitSet / TextLeadingType / TextParagraphDirection / TextBaselineOption`）+ `TextStyleRun / TextParagraph / TextSource` |

**解析（parse_*.go，9 个）**

| 文件 | 行数 | 职责 |
|---|---|---|
| `parse.go` | 223 | 顶层入口 `Open / FromReader`，`parseProject / parseItem / classifyItem`，公用辅助 `walkTdmnPairs / trimNUL / readFloat64BE` |
| `parse_composition.go` | 279 | Composition 解析（cdta 偏移表 / TickRate 推导 / WorkArea / MotionBlur / Shutter / Renderer from PRin）。**定义 `parseCtx` 警告机制** |
| `parse_layer.go` | 373 | Layer 解析（ldta 160/164 字节偏移表 + 3 字节 LayerAttrBits + 子类型 enum @0x83 + light kind @0x88 + trackMatte @0xA0） |
| `parse_footage.go` | 121 | Footage 解析（sspc/opti/Pin/Als2/Cpth chunks） |
| `parse_properties.go` | 217 | 属性树遍历（tdgp/tdbs/tdmn）+ Effect Parade / Marker / Mask Parade 分流 |
| `parse_keyframe.go` | 199 | 关键帧字节布局解码：spatial vs non-spatial dispatcher (`layoutFor`)、ease/tangent (`decodeEasing`)。**silent-fail 全部接入 ctx.warn** |
| `parse_marker.go` | 162 | "ADBE Marker" mrst 解码（layer + comp marker 共用） |
| `parse_mask.go` | 377 | Mask Parade 解码（mkif 48-byte / om-s/omks/shap 顶点流 / 动画路径 + per-snapshot ease） |
| `parse_shape.go` | 204 | Shape Layer Bezier 路径 (`collectShapePaths`) + 参数化基元 (`collectShapePrimitives`: Rect/Ellipse/Star) |
| `parse_text.go` | 354 | 文本图层 btds 解码（slim, 仅 7 个 decoder）。**类型在 `text_types.go`；PS 解析器在 `postscript.go`** |
| `postscript.go` | 347 | btdk PostScript mini-parser（lexer + parser + path navigator + 字符串解码）。unexported，仅 parse_text / write_text 用 |

**写回（write_*.go，7 个）**

| 文件 | 行数 | 职责 |
|---|---|---|
| `write.go` | 171 | 顶层：`Project.WriteAEP` + `Footage.SetPath` + `Project.SetBitsPerChannel` + 共享 helper |
| `write_keyframe.go` | 252 | `Keyframe.*` (8 public + 5 helper)：SetTime/Value/Interp/TemporalEase/SpatialTangent，全 length-preserving block 写 |
| `write_property.go` | 306 | `Property.*` (6 public + 1 helper)：SetStaticValue / SetExpression / SetExpressionEnabled / InsertKeyframe / DeleteKeyframe |
| `write_layer.go` | 830 | Layer ldta 字节/位 + Utf8/cmta 文本 + 时间字段 + parent/source ID + **Layer.SetText**（文本 length-preserving 字符串替换）。ldta flag bit 集中在 `ldtaFlagBit` 表 |
| `write_composition.go` | 411 | Composition cdta 字节 + name/framerate/duration/pixelAspect/resolutionFactor + 标志位 (`SetDraft3D / SetHideShyLayers / SetCompMotionBlur / SetFrameBlending / SetPreserveNestedFrameRate / SetPreserveNestedResolution`) |
| `write_marker.go` | 141 | Marker 写回（layer + comp 共用）：SetTime / SetDuration / SetLabel / SetComment / SetChapter / SetURL / SetFrameTarget / SetCuePointName |
| `write_mask.go` | 138 | Mask 写回：SetMode/SetInverted/SetColor/SetClosed/SetLocked/SetMaskMotionBlur |
| `write_item.go` | 89 | Item-level：Composition / Footage 的 SetComment / SetLabel |
| `write_text.go` | 664 | 文本 per-run / per-paragraph + 字体表扩展。通用 `splicePSValue` helper + 16 typed setter + `Layer.AddFont`。**核心 invariant**：length-changing splice 必须同步更新内嵌 `LIST btdk` 的 size header（bodyOff-8 处 4 字节 BE） |

**序列化**

| 文件 | 行数 | 职责 |
|---|---|---|
| `json.go` | 536 | JSON 视图（snake_case 镜像类型 + `ToJSON / MarshalJSON / WriteJSON`）。**单向导出，无 ReadJSON** |

### 测试（18 个 _test.go，按域拆分）

| 类别 | 文件 |
|---|---|
| 共享 testutil | `testutil_test.go` (rifxBuilder + buildMinimalAEP) / `testutil_keyframe / mask / layer / text / aep24` |
| Domain test | `aep_test.go` (smoke + JSON) / `keyframe_test / mask_test / shape_test / text_test / layer_test / composition_test / marker_test / property_test / item_test / expression_test / parse_warning_test` |

**写新测试**：`*_test.go` 全部 `package aep_test`（公开测试）。新加 fixture 用 `t.Skipf` 缺文件跳过，不阻塞 CI。

### 测试数据

`test_data/re_*.aep` + `.jsx` 配对，~28 个 fixture。完整清单见 `workshop/plans/coverage-detail.md` 末尾。

## 关键约束（修改时容易踩坑）

1. **Length-preserving 写回是硬约束**。任何 `Set*` API 改的 byte 数必须等于原 chunk 长度。length-variable 例外（Utf8 文本 / Expression / PostScript splice）必须 1) 重算父 LIST 的 size header；2) 文本类还要同步内嵌 `LIST btdk` 的 size header（`bodyOff - 8` 处 4 字节 BE）。

2. **TickRate 是 per-composition 的**。不要假设统一 8000。`Composition.TickRate` 由 cdta `@0x08` / `@0xA8` 推导，关键帧 / Marker 时间都用它换算。`aeLegacyTimeBase = 8000` 只是 cdta 缺失时的 fallback。

3. **关键帧字节布局两种**：
   - **Spatial-style**：4D color / Position / Anchor (3D)。block 头部 byte `0x07 = 0x07` 或 `0x01 + dims≥2`；scalar ease 在 0x18/0x20/0x28/0x30；values 在 0x38；bpk = 0x38 + 3·N·8
   - **Non-spatial**：Opacity (1D) / Scale (3D) / Mask Feather (2D)。block 头部 byte `0x07 = 0x00`；values 在 0x08；per-component ease 在 `0x08 + (N+i)·8`；bpk = 0x08 + 5·N·8
   - 由 `layoutFor(header07, dims)` 集中分发。**不要在 caller 端手算 offset**

4. **chunk ID 大小写敏感**。`Tdb4` (uppercase, legacy) ≠ `tdb4` (lowercase, modern)。rifx 同时声明两个常量。

5. **不安全的并发**。Project / Composition / Layer / Property / Keyframe **全部 mutate 共享 `rifx.Chunk.Data` 字节**。多 reader 无锁可行；任何 Set* 调用必须由调用方自己同步。

6. **mkif 已结构化但有剩余字节未解**。不假设 48 字节全已知；0x10/0x18/0x20-0x27 字段未 RE。round-trip 用 `Mask.MkifRaw` 保留原字节。

## 测试惯例

- **合成 RIFX fixture 优于真实文件**。`rifxBuilder` + `build*` helpers 让 corruption 类测试 reproducible。新增解析路径时优先合成 fixture。
- **真实文件走 `TestManualFile`**：`go test -aep "C:/path/to/your.aep"`，不入 CI baseline。
- **每个解析改动都应该有对应 fixture 测试 + roundtrip 测试**。看 `TestKeyframeRoundtrip` 是好模板。
- **警告路径用 `buildCorruptKeyframedLeaf` 注入异常 lhd3**。不要改 `leafKeyframed` 的健康 helper，分开 corruption 与正常 case。
- **AE 24 字段 fixture 走 `re_*_ae24.jsx`**。AE 2020 写不出来的字段（`fontCapsOption / strokeOverFill / autoHyphenate` 等）必须 AE 24+ 写。测试代码 `t.Skipf(...)` 缺文件不阻塞 CI。

## 已知非完美区

- `ReadJSON` 不存在；JSON 仅作只读快照。修改必须走二进制路径。
- 文本 `Layer.TextSource` 已解出 Text / Fonts / per-paragraph (9 个字段) / per-run (22 个字段) / ManualKerning + Kerning / FontAxes (R only) / IsBoxText + BoxBounds。**未 RE 的 btdk 字段**：`ligature`（默认 false 无 diff，需 OT `liga` 字体 fixture）/ `lineOrientation` 竖排切换（结构性）。
- Shape primitives (Rect/Star/Ellipse) 已结构化为 `Layer.ShapePrimitives []*ShapePrimitive`。**不支持**：增删 primitive / Vector Group 嵌套层级（GroupName 只取最内层）。
- 4D 颜色 32bpc 项目的范围 (0..1 vs 0..255) 未验证。
- ldta 长度只验证过 160 (AE ≤22) / 164 (AE 23+) 两种；AE 25 实测仍 164（见 `scars/ldta-length-third-variant-not-found.md`）。
- 旧 `Layer.LightTypeID` (4101..4104) 已 deprecated；正确 light kind 在 ldta `@0x88` (0..3) → `Layer.LightKind`。
- ldta `@0x98..0x9F` 疑似 camera FilmSize 但 ScriptingAPI 不可写（见 `scars/camera-filmsize-ldta-write-blocked.md`），未 ship setter。
- 增删 Layer / Mask vertex / Marker / Effect / ShapePrimitive 仍不支持（结构性）。Keyframe 增删已 ✅。
- comp.renderer 已 ship **R only**（PRin LIST → prin chunk @0x04）；setter 是 P3（prin 双段 NUL-sep + prda 长度随 renderer 变）。

## 与上游差异

相对 boltframe/aftereffects-aep-parser 主要增量：per-comp TickRate、4D 颜色、Mask 完整解码、Shape Layer Pen 路径 + 参数化基元、length-preserving 写回完整 API（53+ typed setter）、Project.Warnings、Camera/Light/Material/Geometry 类型化访问器、Composition.Renderer / ResolutionFactor 读。
