# V2.2 iter-5 byte-level RE — tdb4 / tdbs / cdat schema differences

**Date**: 2026-05-25
**State**: code reverted to iter-5b (PASS but layers.length=0 + AE 警告 "Anchor dim 2 → 3")
**Purpose**: offline catalogue of every byte-level diff between our V2.2 builder output (iter5_check.aep) and AE-saved tolerance.aep, with hypothesis on which trigger silent-drop vs cosmetic-warning.

Tools used (all offline, no AE):
- `tmp_debug/dump_root/` (decodes tdmn / tdsn names)
- `tmp_debug/diff_ldta/` (byte-level ldta diff — confirmed identical 164B ✓)
- `tmp_debug/diff_stream/` (per-stream tdbs body diff)
- `tmp_debug/dump_gide/` (Gide constant verification)

## 已对齐的字段（iter 4 + 5/5b 修过）

- ✅ ldta 164B byte-identical (LayerID/Quality/StretchDivisors/time-ticks/AttrBytes/LayerSubtype/ParentID 全对)
- ✅ Layr 第 4 child = LIST Gide (gdta 8B + LIST list / lhd3 52B), Gide content byte-identical
- ✅ Item-level Ewst (0 child) sibling immediately after Layr
- ✅ Item LIST 结构性 child 类型计数完全一致 (Layr/DLay/SLay/CLay/SecL/Ewst/PRin/dats)
- ✅ Root Vectors Group 5-层嵌套 + tdsb 0x00000401 / 0x00000001 标志位
- ✅ outer Layr LIST(tdgp) 7 个 property group placeholders (Root Vectors / Transform / Layer Styles / Extrsn / Material / Audio / Layer Sets)

## 未对齐字段 — 候选 silent-drop 触发器

按嫌疑度排序（基于"AE 不会在 cosmetic 上 silent drop layer"假设）：

### P0 — Transform stream tdsb 标志位 (0x01 vs 0x03)

所有 Transform Group 内的 6 stream (Position_0/_1, Orientation, Rotate X/Y, Envir Appear) tolerance 用 **tdsb = 0x00000003**，我们用 0x00000001。

| Stream | Ours tdsb | Tolerance tdsb |
|---|---|---|
| ADBE Position_0 | 0x00000001 | **0x00000003** |
| ADBE Position_1 | 0x00000001 | **0x00000003** |
| ADBE Orientation (tdbs inside otst) | 0x00000001 | **0x00000003** |
| ADBE Rotate X | 0x00000001 | **0x00000003** |
| ADBE Rotate Y | 0x00000001 | **0x00000003** |
| ADBE Envir Appear in Reflect | 0x00000001 | **0x00000003** |

Shape sub-prop streams (Rect Size, Fill Color etc) tolerance **保持 0x00000001** = 跟我们一致。

**Hypothesis**: 0x03 = "this stream belongs to the Layer Transform Group" marker. AE 校 Transform Group 完整性时按 tdsb=0x03 计数；若 tdsb=0x01 AE 视作"非 Transform stream"→ 整个 Transform 视作 incomplete → silent drop layer。

**Fix candidate**: 给 lowerLayerTransform 里所有 stream 用一个新 `makeTdsbTransform()` 出 0x00000003。

### P1 — Position_0/_1 spatial 流缺 tdum/tduM 后缀

| | Ours | Tolerance |
|---|---|---|
| Position_0 tdbs children | 4 (tdsb + tdsn + tdb4 + cdat) | 6 (上 4 + **tdum 8B + tduM 8B**) |
| Position_1 tdbs children | 4 | 6 |

tdum/tduM = spatial bounds (f64 min/max)，tolerance 全 0 (因为 Position default = 0,0 = 静态无范围)。

**Hypothesis**: AE 看 spatial stream 必须有 tdum/tduM bound markers，缺 → "数据不完整" → silent drop。

**Fix candidate**: 给 LowerFloat64Stream 加一个 `withSpatialBounds bool`，spatial=true 时尾巴 append tdum(8B 0) + tduM(8B 0)。或更精确的 spatial bounds 计算 (静态 value 时 min=max=value)。

### P2 — Transform stream tdb4 head bytes @0x08-0F

| Field | Ours | Tolerance | 含义猜测 |
|---|---|---|---|
| Position_0 tdb4 @0x08-0B | 00000000 | **00000001** | "spatial flag" / "stream version" |
| Position_0 tdb4 @0x0C-0D | 0000 | **ffff** | uint16 = 65535 = "no limit" / "default cap" |
| Rotate X tdb4 @0x08-0B | 00000000 | **00000001** | 同上 |
| Rotate X tdb4 @0x0C-0D | 0000 | **ffff** | 同上 |
| Rect Size tdb4 @0x08-0B | 00000000 | **ffffffff** | uint32 = -1 / max / sentinel |
| Fill Color tdb4 @0x08-0B | 00000000 | **0002ffff** | type-specific |
| Anchor Point tdb4 @0x04-05 | 0001 | **000f** | flags 0b1111 |
| Anchor Point tdb4 @0x06 | 07 | **03** | headerByte 不同 |

**Hypothesis**: AE 用 tdb4 @0x08-0F 编码 stream metadata (version / type / range)。我们 makeTdb4 hardcoded 0 → AE 解码失败但容错继续 → layer 解析完但 marked malformed → silent drop。

**Fix candidate**: 不同 stream 类型用不同 tdb4 head。需要按 matchName 查 RE'd value 表。

### P3 — Rect Size / Anchor Point dim + headerByte

- **Rect Size** (Vec2 non-positional): tolerance @0x06 = 0x00 (非 spatial), 我们 0x07
- **Anchor Point** (3D vector!): tolerance dim=3 + @0x06 = 0x03, 我们 dim=2 + @0x06 = 0x07

我们的 `LowerVec2Stream` 一律 `headerByte=0x07, spatial=true`，与实际不符。Vec2 不一定是 spatial — Size、Scale 是 Vec2 但非 spatial；Anchor Point 是 3D 非 Vec2。

**Hypothesis**: dim 错 + headerByte 错可能导致 AE 解码 cdat 时偏移错 → 整 stream malformed → drop layer。AE 警告"Anchor dim 2 但需要 3"是直接体现。

**Fix candidate**: 
- LowerVec2Stream 拆分: `LowerVec2NonSpatial` (Size/Scale, headerByte=0) + `LowerVec2Spatial` (deprecated; spatial Position 用 Position_0/_1 split)
- LowerVec3Stream 给 Anchor Point: dim=3, headerByte=0x03 (per tolerance)

### P4 — Rect Size cdat padding (48B vs 80B)

我们 emit cdat 48B (2 × 8B value + 32B 零 padding)，tolerance 80B + tdum(8B) + tduM(8B)。

Tolerance Rect Size tdum = `c0df400000000000` (-32256 as f64)，tduM = `40df400000000000` (+32256)。"32256" 像 AE 内部 max coord 常量。

**Hypothesis**: cdat 48B vs 80B mismatch → AE 期望读到 80B 才 advance 到下一 chunk → 读到 tdum 时 mismatch → "数据缺失"。

**Fix candidate**: spatial-bounded Vec2 stream emit cdat 80B + tdum/tduM。需要 RE per-prop bound 常量。

### P5 — V2.2 always-emit vs AE elide-default

Tolerance Rect body 只 5 children = tdsb + tdsn + Size + LIST tdbs + Group End。只 emit Size（user 显式设过的）。我们 emit Rect 全 4 sub-prop (Direction + Size + Position + Roundness)。

**P5 是 cosmetic** — 多 emit 不会让 AE drop。但 placeholder 形式 (Direction empty) 可能触发 dim 警告 (iter-5b 警告 "Shape Direction dim 2 但要 1"，可能确实是 empty placeholder 让 AE 推断 dim 错)。

**Fix candidate (低优先)**: 实现 PropertyStream "Unset" mode + 选择性 emit (跟 iter 3 scar 提的策略 fix B 真实形态)。

### P6 — tdsn 占位串

Tolerance tdsn payload = `Utf8` magic + size=6 + `"-_0_/-"` 字面量 (AE 内部 sentinel "未命名" placeholder)。我们用 `"Anchor Point"` / `"Size"` 等英文 display name。

**P6 是 cosmetic** — display name 只是 UI 显示用，AE 不会因此 drop。

---

## 影响 priorities 评估

按"silent drop layer 触发"嫌疑：

| 候选 | 嫌疑度 | 修复复杂度 |
|---|---|---|
| P0 Transform tdsb 0x01 → 0x03 | 高 | 低 (改 6 行) |
| P1 Position spatial tdum/tduM | 高 | 中 (tdbs builder + spatial flag plumb) |
| P3 Vec2 spatial 滥用 + Anchor 3D dim | 中-高 (有 AE 警告佐证) | 中 (API + LowerVec2 拆分) |
| P2 tdb4 @0x08-0F bytes | 中 (无直接证据) | 高 (per-matchName 表) |
| P4 cdat padding 48 vs 80 | 中 | 高 (per-prop bound 表) |
| P5 over-emit shape sub-props | 低 (cosmetic 多余) | 高 (Unset mode 重构) |
| P6 tdsn display name | 极低 (纯 UI) | 低但无价值 |

## 建议下一步 AE 验证策略

不再 6 变体 bisect 全跑。**每次只跑 variant #2 (empty ShapeLayer)**，单一变量验证：

1. **iter-6a**: 仅修 P0 (Transform tdsb 0x03)。若 layers.length 变 1 → P0 是 silent-drop 真凶
2. **iter-6b**: 仅修 P1 (Position tdum/tduM)，P0 reverted。若变 1 → P1 凶手
3. **iter-6c**: P0 + P1 都修。若仍 0 → P3/P2/P4 接力

每次 AE 跑只 ~15s (单变体)，不会 9min × 6。 

## 未 RE 候选 / 盲区

- 没看 Orientation otst 里面 tdbs body 跟 tolerance 差多少
- 没看 Layer Styles / Extrsn Options 等 placeholder 实际跟 tolerance 是否字节级一致
- Anchor Point AE 警告 dim 2 → 3，但 tolerance "Nested" ShapeLayer 根本没 Anchor。AE 警告可能源自 ldta 内部对 Anchor 的 expected reference (不是 Transform Group)。需 RE。
- head counter B (我们 14, tolerance 36) — iter 4 doc 说 "≥ nextItemID 即可"，但可能要求更严
- cdta secondary fields 我们 iter 4 改了 @0x18，可能还有别处
