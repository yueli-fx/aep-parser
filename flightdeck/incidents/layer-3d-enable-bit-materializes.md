---
status: active
when_to_read: implementing 3D-layer support (roadmap priority 2); making a 2D AV/shape/solid layer 3D; wondering whether flipping Is3D needs companion Z-position / Orientation / X·Y-rotation / Material-Options channel synthesis; debugging "AE accepts my 3D-flagged layer but the 3D channels are missing on resave"; planning the visible-3D (camera/Z) render gate
applies_to: [layer, is3d, three-d-layer, ldta, 0x26, attr-byte1, SetIs3D, 3d-enable, transform-3d-channels, position-z, orientation, rotate-x, rotate-y, material-options, channel-synthesis, default-omission, parse-the-clone, ship-gate, ae2020, ae2025, dom-readback, roadmap-priority-2]
last_updated: 2026-06-15
resolved_by:
---

# 3D-enable: flipping the Is3D bit alone makes AE materialize a full 3D layer

## TL;DR

`SetIs3D(true)` (flip ldta @0x26 **bit2**, length-preserving — already shipped in
`scene_layer_writers.go` + `back_layer.go::setFlagBit(flagIs3D)`) is **sufficient
to turn a from-scratch 2D shape layer into a full 3D layer**. AE re-synthesizes
the entire 3D transform tree from the single bit on open — **no companion channel
synthesis is needed for ENABLE**. This collapses the roadmap's "Transform 3D 通道
合成" sub-item *for the enable case*: AE does it for us.

## Ground truth (probe, `verify_3d_enable.jsx` + `TestLayer3DEnable_AEShipGate`)

Built a from-scratch shape layer (rect+fill, 2-comp Position [960,540]), `Reopen`
(parse-the-clone to get the ldta backref), `SetIs3D(true)`, `WriteAEP`. AE opens it
and reports — **AE 2020 + AE 2025 byte-identical behavior**:

- `layer.threeDLayer == true`
- **Position dims = 3, value = [960,540,0]** — AE expanded our 2-comp Position to
  3-comp, z=0.
- `ADBE Orientation` / `ADBE Rotate X` / `ADBE Rotate Y` / `ADBE Rotate Z` all
  present (DOM).
- `ADBE Material Options Group` present.

So the layer is a genuine, complete 3D layer in AE's DOM, built from a 2D shape +
one flipped bit.

## Two consequences for the WRITE side

1. **Resave elides the materialized channels.** Re-parsing the AE-resaved file,
   our parser sees `Is3D=true` but `Orientation()/RotateX()/RotateY()` return nil —
   AE materialized them in the DOM but **default-omits them on disk** (same
   `transform-group-default-omission` rule). The durable on-disk signal of 3D-ness
   is the **Is3D bit**, not the channel chunks.
2. **Setting a NON-default 3D transform still needs the slot.** Our from-scratch
   Position is 2-comp (16-byte cdat); a 3-comp Position is 24-byte → writing Z is
   **NOT length-preserving**. RotateX/Y/Orientation aren't in our emit at all. So
   the **visible** 3D transform (Z-depth / rotation) is a separate capability that
   needs either (a) a 3-comp Position emit when Is3D, or (b) synthesis-insert of
   the rotation channels (camera/light-option vein), or (c) parse-the-clone off an
   AE-materialized 3D layer. TBD — next priority-2 step.

## Two gates: DOM enable + visible camera-dolly (both PASS, AE2020+2025)

1. **`TestLayer3DEnable_AEShipGate`** (DOM-readback). **Deliberately DOM-level**: a
   3D layer with a DEFAULT transform (z=0, no rotation) renders **pixel-identical**
   to its 2D self, so "enable" has no visible surface — the surface is the 3D-ness
   AE reports (threeDLayer + materialized channels + Position→3D).
2. **`TestLayer3DCamDolly_AEShipGate`** (render-pixel, red line 4). The **visible**
   proof, and it needed **zero new channel-write code**: build a from-scratch 3D
   BOX (z=0) + a `NewCameraLayer`, then the JSX dollies the **camera's** Position Z
   (cameras are inherently 3D, 3-comp Position already settable) — at z=-700 the
   300px box renders **426px wide**, at z=-2400 it renders **124px** (3.41× ratio,
   byte-identical both versions). A 2D layer ignores the camera → constant width,
   so the ratio is the 3D proof. This is the roadmap's 推拉镜 in miniature.

So "图层 3D flag" is closed with both an acceptance and a render proof.

## Per-layer Z = PARALLAX — shipped with ZERO new write code (2026-06-15)

The write-side worry in consequence #2 above was half-wrong for **Position**: the
from-scratch shape Position is **already a 3-component spatial slot on disk**
(`lowerTransformVec2Spatial` → `encode3D([x,y,0])`, bpk-128, value@0x38 — the
template is the "6-axis 3D-compatible" transform schema, Z just pinned 0). So
after Reopen the Position cdat is 24 bytes, and `Layer.SetPosition([x,y,z])` on
the reopened 3D layer is **length-preserving** — AE honours the Z. No builder
change, no synthesis.
- `TestLayer3DParallax_AEShipGate_AE2020/2025` PASS: two same-size (300px) 3D
  boxes, NEAR z=-800 / FAR z=+1200, Go-created camera at [960,540,-1800] (its
  Position also set in Go — cameras are natively 3-comp). Rendered: NEAR=300px,
  FAR=99px (**3.03× ratio, byte-identical both versions**) — depth from per-layer
  Z. Visually eyeballed (big box left, small box right on a dark backdrop).
- The whole scene is Go-built (no JSX mutation) — the gate proves OUR output
  renders parallax.

## RotateX/Y/Orientation also ZERO new code — the template carries every 3D slot (2026-06-15)

The earlier worry that rotation/orientation "needs synthesis-insert (channels
absent from the tree)" was **WRONG**. Dumping `templates/v2_2_transform_group_body.bin`
(`tmp_debug/xform_dump`) shows the from-scratch transform body already carries
**every** channel: Anchor · Position (+Position_0/_1 separated) · Scale ·
**Orientation · Rotate X · Rotate Y · Rotate Z** · Opacity · Envir Appear. So
after Reopen the parser surfaces `RotateY()` / `RotateX()` / `Orientation()` as
real (non-nil) properties, and the existing `SetRotateX/Y` / `SetOrientation`
setters overwrite an existing cdat — length-preserving, no synthesis.
- `TestLayer3DRotateY_AEShipGate_AE2020/2025` PASS: a 400px 3D box, RotateY=50°,
  Go camera at [960,540,-800] (close → strong perspective). Renders a
  **trapezoid** — near vertical edge magnified (leftH=422 / rightH=612, tall/short
  **1.45×**, byte-identical both versions); a flat rect / dropped RotateY / 2D
  layer would be ≈1.0. AE reads back RotateY=50. Visually eyeballed (clean
  perspective-skewed quad). NB the foreshorten direction depends on the rotation
  sign, so the gate asserts the asymmetry direction-agnostically.
- RotateX / Orientation / RotateZ ride the same proven path (same template slots
  + same setter family) — gate on demand.

**Net:** the entire 3D transform group (enable + Z parallax + rotation/orientation)
is writable from scratch with NO new serializer code — `SetIs3D` + the existing
transform setters on the reopened layer, because the embedded transform template
was already the full 6-axis 3D schema. Camera dolly + Z parallax + RotateY tumble
all render-gated; the headline 推拉镜/视差/透视 are shipped.

## Camera Depth of Field — render-gated (2026-06-15)

The camera DoF setters (`SetCameraDepthOfField` / `SetCameraFocusDistance` /
`SetCameraAperture` / `SetCameraBlurLevel`) existed but were only DOM-readback-
checked (camera-light showcase) — never render-verified to actually defocus.
`TestLayer3DDoF_AEShipGate_AE2020/2025` PASS: a SHARP 3D box at the camera's
focus distance (z=-500, focus 1300) + a far BLUR box (z=+1400), DoF on, aperture
300, blur level 200. Rendered frame: SHARP left-edge transition band = **1px**
(crisp), BLUR band = **21-22px** (defocused) — a >20× ratio, both versions. The
setters must run on the **reopened** camera (from-scratch camera options elide).
This is the 景深 of the "景深视差推拉镜" milestone.

## Lighting (from scratch) — ZERO new code, render-gated (2026-06-15)

A from-scratch POINT light illuminates a from-scratch 3D layer with **no new
serializer code**: `SetLightKind(LightKindPoint)` is ldta @0x88 length-preserving
(works on the fresh light directly — no reopen), and AE's default material has
**Accepts Lights = ON**, so a 3D layer is lit without touching its (empty)
Material Options group. `TestLayer3DLight_AEShipGate_AE2020/2025` PASS: a 3D gray
panel under a POINT light placed up-left/in-front renders a clear brightness
falloff — near-light interior **185.6** luma vs far **88.0** (2.11×, byte-identical
both versions, unclipped). A flat-shaded panel would be uniform; the gradient is
the proof.

**`SetLightKind` from scratch is genuine.** The earlier worry that JSX read back
`light type=4414` (not a clean Point) was unfounded: **4414 IS ExtendScript
`LightType.POINT`** (the enum runs 4412 PARALLEL / 4413 SPOT / 4414 POINT / 4415
AMBIENT — showcase camera-light INDEX records 4415=AMBIENT). This retires the old
"NewLightLayer 默认环境光, no from-scratch SetLightType" caveat.

## Shadows (from scratch) — Material-Options synthesis-insert, render-gated (2026-06-15)

A from-scratch 3D layer emits an **EMPTY** Material Options group
(`lower_layer.go` `emptyPropGroupFlags`) — AE materializes the full 15-leaf
material tree (all defaults) in its DOM on open but **elides every leaf on disk**,
so `MaterialCastsShadows()` is nil and `SetMaterialCastsShadows` reports "property
not present". Casting a shadow needs **Casts Shadows = On** (defaults Off), a
non-default value AE persists — so the leaf must be synthesized back.

`aep.SetMaterialOption(layer, matchName, value)` is the synthesis-lite sibling of
`SetEffectParam` (`internal/serializer/mutate_layer_material.go`): clones the
requested `(tdmn, LIST:tdbs)` leaf from embedded `material_options_leaves.bin`
(the 15-leaf tree extracted from `re_material_options.aep`, group with most
children = all props authored), splices it into the group in AE's **canonical
order** (group order is significant — out-of-order leaves get dropped on open),
re-parses it into a `*Property`, appends to `layer.Properties` + the tree, writes
the caller's value. Atomic: snapshots group-chunk children + flat Properties +
tree + warnings, rolls back on any parser warning. Already-present leaf (parsed/
fixture layer or prior splice) → plain SetStaticValue.

`TestLayer3DShadow_AEShipGate_AE2020/2025` PASS: a Go-built 3D WALL catcher + a 3D
CASTER whose `ADBE Casts Shadows`=On is **synthesized** + a POINT light (Casts
Shadows on) above/in-front → caster casts a hard shadow DOWNWARD onto the wall.
Centre-below-caster luma **0.0** (pure black — single light, no fill) vs lit wall
**114** (symmetric L/R), byte-identical both versions; AE reads back
`materialCastsShadows=1`. A no-shadow render leaves that central region lit, so a
dark centre flanked by a lit wall is the coincidence-proof.

**Two things needed NO synthesis:** (a) the **catcher** — an empty material group
accepts shadows + lights by default; (b) the **light's own** Casts Shadows — it is
already present in a from-scratch light's Light Options (index 7), so
`SetLightCastsShadows(true)` works directly.

**This closes priority-2 (3D).** enable + Z parallax + rotate/orient + camera
dolly + DoF + lighting + shadows are all render-gated double-version. Material
Options other coefficients (Diffuse/Specular/…) ride the same `SetMaterialOption`
path (Go-round-trip-tested; render-gate on demand).

## Pre-existing 3D infra (already shipped, fixture-based)

Material / Geometry options + transform setters (SetRotateX/Y, SetOrientation,
3-comp SetPosition) already work **on fixture 3D layers** (`re_material_options.aep`
/ `re_cameralight.aep`, `layer_3d_test.go`) — they require the channel present in
the tree. The gap this incident closes is making OUR from-scratch layer 3D in the
first place; wiring those setters onto a from-scratch-3D layer is the next step.

## 2026-06-17 — priority-2 真正收口：RotateX/Orientation/RotateZ render-gate + Orientation 静态写 bug

补齐 RotateY 之外的三轴双版本 render-pixel gate（`layer_3d_rotaxes_shipgate_test.go`，
6 个 `TestLayer3D{RotateX,Orientation,RotateZ}_AEShipGate_AE2020/2025`）：
- **RotateX=50°** → 透视下顶/底边 width 不等（侧躺梯形，topW/botW 1.45）。
- **RotateZ=45°**（= `SetRotation`，Z 轴）→ 正方形转成菱形（center col >> edge col，54×）。
- **Orientation [0,50,0]** → 同 RotateY 的 Y 轴梯形（leftH/rightH 1.45）。

RotateX/RotateZ 确如先前判断「同路径零新代码」（标量 BE cdat 覆写）直接过。**Orientation 不是**——
踩出一个真 correctness bug（值 round-trip 绿但 AE 渲染 0，红线红线 #4 活样本）：

**根因（byte-diff AE-authored golden 揪出）**：3D 层 Orientation 的 `otst` wrapper 把静态值存**两份**：
1. `tdbs/cdat`（24B）= **小端**（py-aep 早 RE 的读路径已知）。
2. `otky/otda`（24B，"orientation default value chunk"）= **大端** —— **AE 读静态当前值用的是这份，不是 cdat**。

旧 `SetStaticValue` 两头都错：(a) 用通用 `writeFloat64`（大端）写 cdat，跟 AE 的小端**字节翻转**
（50.0 → 9.26e-320 ≈ 0）；(b) **完全没碰 otda**，留默认 [0,0,0]。两个 bug 叠加 → AE 渲染 0。
仅修 cdat 端序仍 FAIL（AE 不读 cdat 读 otda），必须**双写**。

**修法**（`back_property.go` + `parse_properties.go`）：propertyBackrefs 加 `cdatLE bool` + `otda *rifx.Chunk`，
`parseOrientationProperty` 静态分支同时绑 cdat（标 LE）+ otky 下的 otda；`SetStaticValue` 据 `cdatLE`
选 `writeFloat64LE` 写 cdat，并把同值**大端镜像进 otda**。其它属性 otda=nil、cdatLE=false，零影响。

**波及面**：这 bug 对**任何** SetOrientation（fixture 层也算，非仅 from-scratch）都存在，只是从未 AE-gated
（capindex 标 roundtrip）才一直没暴。现 SetRotateX/Y/Rotation/Orientation 全升 verify=render-pixel。
**遗留**：animated orientation（otda 多帧 + ease）写未验，需要时另立。
