# ⚠ AddEffect — Effect Parade splice RE + ship findings

SUMMARY: AddEffect — Effect Parade splice RE + ship findings
READ WHEN: implementing or extending AddEffect / the effect-template library; adding a new effect to the embedded set; debugging "AE drops/rejects a Go-added effect" or "cannot find layer ID=N in composition" on open; deciding whether an effect is splice-portable; extending parade auto-create to another group kind (Mask Parade); reasoning about the (tdmn, sspc) effect chunk unit or the tdpi host-layer binding

---

`aep.AddEffect(layer, matchName)` appends a built-in effect to a layer's
`ADBE Effect Parade`. Shipped 2026-06-10 (two commits: core mechanic + 12-effect
library). AE 2020 + AE 2025 ship-gate 24/24 across the full 12-effect library
(2026-06-11; initially shipped 2026-06-10 on a 5-effect payload-size sample).
Promoted Alpha → Stable as a structural op on the full-library gate,
together with RemoveEffect (rides the gated Effect-Parade child removal).

## Effect chunk structure (RE'd from re_property_struct_baseline.aep)

The Effect Parade is a `LIST(tdgp)` whose direct children are:

```
[0] tdsb            (4B group header)
[1] tdsn            (14B group display name)
[2] tdmn  "ADBE Gaussian Blur 2"   (40B fixed match-name field)
[3] LIST(sspc)      effect #1 payload  (params + pard + built-in-params group)
[4] tdmn  "ADBE Tint"
[5] LIST(sspc)      effect #2 payload
[6] tdmn  "ADBE Fill"
[7] LIST(sspc)      effect #3 payload
[8] tdmn  "ADBE Group End"          (lone sentinel, NO payload after it)
```

So **one effect = a `(tdmn[40B], LIST:sspc)` pair**. The parade has no
count/index chunk (same finding as RemovePropertyGroup/MoveTo — see
[[property-indexed-group-structural-re]]). The trailing `ADBE Group End` tdmn is
the sentinel; **AddEffect splices the new pair in just before it**.

## Mechanic (identical to DuplicatePropertyGroup, template-sourced)

`DuplicatePropertyGroup` already proved AE accepts a spliced `(tdmn, payload)`
effect pair (it clones a sibling). AddEffect is the same splice but sources the
pair from an **embedded AE-native template** instead of an existing instance —
so it works even when the layer lacks that effect. The returned `*Effect` is the
spliced pair re-parsed via `collectEffects` (back-refs point at the spliced
chunks, never the cache), so callers can tune params immediately
(`Property.SetStaticValue` on effect params already works).

- **LIST sizes auto-fix.** `rifx.Chunk.Write` calls `c.PayloadSize()` (computed
  from children) at write time, so every ancestor LIST size recomputes bottom-up
  — splicing whole chunks needs no manual size fixup. Same length-variable path
  as gradient write / `Footage.SetPath`. **Do not hand-patch LIST sizes.**
- **Atomic.** snapshot parade chunk children + parade scene `Children` + flat
  `layer.Effects` + `Project.Warnings`; roll back on any new parser warning.

## Key findings / gotchas

1. **Version-portable.** Templates are extracted from an **AE-2020-saved** file
   yet AE 2025 accepts them verbatim — across payload sizes 1.7KB (Gaussian Blur)
   to 20.6KB (Pro Levels2). Same portability principle as the gradient finding
   ([[gradient-fill-write-re]]). One AE-2020 fixture serves both gate versions.
2. **AE tolerates a "duplicate" internal effect-instance id.** Adding e.g. a 2nd
   Gaussian Blur (verbatim template) is accepted — AE does not require a fresh
   per-instance id at the chunk level (Duplicate showed the same).
3. **Parser surfaces only non-default effect params.** A default Gaussian Blur
   instance exposes a single `ADBE Gaussian Blur 2-0000` param (the others are
   default-elided), NOT `-0001` as the docgen example suggests. Don't assume a
   param match-name is present — iterate `Effect.Parameters`.
4. **From-scratch shape layers have NO Effect Parade** — and neither does ANY
   effect-less layer: AE only persists the parade once ≥1 effect exists. Solved
   2026-06-10 by parade auto-create (see § Parade auto-create below).
5. **tdpi = host-layer binding, AE validates it on open (the Phase-1 latent
   bug).** Every effect param's tdbs carries a 4-byte BE `tdpi` chunk holding
   the OWNING layer's ID (re_effect_library host id 15, re_shape_effect host id
   13 — tracks the host in both). The embedded templates carried the extraction
   fixture's id 15 verbatim; splicing into a layer whose ID ≠ 15 makes AE
   reject the project on open with 无法在合成"X"中找到图层 ID=15 (an EXC from
   app.open, NOT a corrupt-file dialog). **The Phase-1 ship-gate passed only by
   coincidence** — the baseline fixture's host layer ID is also 15 (both
   fixtures: one comp + one layer built by similar JSX → same ID allocation).
   Fix: `retargetEffectHostLayer` rewrites every tdpi in the cloned pair to the
   destination layer's ID before splicing (white-box regression
   `TestAddEffect_RetargetsTdpiHostLayer`; re-gated AE 2020 Phase-1 sample +
   auto-parade both versions). Corollary: effects with layer/path REFERENCE
   params carry tdpi pointing at OTHER layers — a blind retarget-all would
   corrupt those; the parameter-only curation rule keeps retarget-all safe.
   **2026-06-20 relocation:** the retarget now lives in `AddEffect`, NOT the
   shared `addEffectFromChunks` splice core. `BuildPseudoEffect` synthesizes its
   own tdpi (header→host, layer-picker→a deliberately chosen OTHER layer) and the
   blanket retarget-all in the core silently clobbered the picker → host. Only
   the template path (AddEffect) has foreign tdpi to fix, so retarget moved there;
   the pseudo paths splice through the core untouched. See
   [[pseudo-layer-picker-tdpi-retarget-clobber]] (incl. the 0-based-Index /
   1-based-AE false-green gate that masked it).

## Effect-template library (193, embed.FS)

`internal/serializer/templates/effects/effect_adbe_*.bin`, each a `LIST(tdgp)` wrapper
around one `(tdmn, sspc)` pair, extracted from AE-2020 fixtures
`test_data/generated/fixtures/re_effect_library.aep` (wave 1, 12 effects) +
`re_effect_library2.aep` (wave 2, 17 effects, 2026-06-11) — generated by the
matching `re_effect_library*.jsx` applying a curated **parameter-only**
built-in set. Extractor: `tmp_debug/extract_effect_lib <fixture.aep>` (parses
the fixture, walks to the parade, wraps each `(tdmn, sspc)` pair, writes
`effect_<sanitized-matchname>.bin`; reusable replacement for wave 1's throwaway
in-package test). The `.bin`s are the committed source of truth — `test_data/`
is gitignored so the fixtures/JSX live locally.

Wave 1: Gaussian Blur 2 · Fill · Tint · Brightness & Contrast 2 · Tritone ·
Easy Levels2 · Pro Levels2 · HUE SATURATION · Box Blur · Glo2 · Invert ·
Exposure2.

Wave 2: Drop Shadow · Sharpen · Mosaic · Noise · Geometry2 (Transform) ·
Ramp (Gradient Ramp) · Fractal Noise · Tile (Motion Tile) · Motion Blur
(Directional Blur) · Linear Wipe · Wave Warp · CurvesCustom (Curves) ·
Slider/Point/Color/Angle/Checkbox Control (expression controls; "ADBE Layer
Control" stays excluded — layer reference).

Wave 3 (2026-06-12): Point3D Control — extracted from the untouched instance
in `re_effect_param_types.aep` (the control-type param-template fixture,
[[effect-param-elision-synthesis-lite]]); dual-version gated via the
SetEffectParam ship-gate (AddEffect + materialize + readback both versions).

Wave 4 (2026-06-12, `re_effect_lib3.jsx`): Turbulent Displace · Roughen Edges ·
Echo · Radial Blur · 4-Color Gradient · Checkerboard · Grid · Stroke ·
Corner Pin · Venetian Blinds (10). `TestAddEffectWave4_AEShipGate_AE2020/2025`.

Wave 5 (2026-06-16, `re_effect_lib5.jsx` → 41 candidates / 38 OK / 3 wrong
match-name [BULGE, Bezier Warp, Channel Mixer — corrected names TBD next wave]):
38 MG parameter-only built-ins — distort (Twirl/Polar Coordinates/Spherize/
Magnify/Ripple/Optics Compensation), stylize (Posterize/Threshold2/Find Edges/
Color Emboss/Emboss/Strobe/Brush Strokes), perspective (Bevel Alpha/Edges),
color (Photo Filter/Vibrance/Color Balance 2/Color Balance HLS/Black&White/
Gamma·Pedestal·Gain2), blur (Channel Blur/Bilateral/Smart Blur/Unsharp Mask2),
channel (Shift Channels/Solid Composite/Minimax/Arithmetic), generate (Circle/
Lens Flare/Cell Pattern/Lightning 2/Laser/Paint Bucket), time (Posterize Time),
matte (Simple/Matte Choker). All tdpi-host=15 uniform (no dangling layer-ref).
`TestAddEffectWave5_AEShipGate_AE2020/2025` — 38-in-one-run on a 100% Go-built
file, both versions readback-in-order + resave-preserved.

Wave 6 (2026-06-16, `re_effect_lib6.jsx` → 45 candidates / 23 OK; corrected
wave-5 Bulge = `ADBE Bulge` not `ADBE BULGE`): 23 more parameter-only built-ins
— distort (Bulge/Offset/Mirror/Fractal), generate (Write-on/Scribble Fill/
Eyedropper Fill/Audio Spectrum/Audio Waveform), color (Auto Levels/Auto Color/
Auto Contrast/Equalize/Leave Color/Change To Color/Change Color), perspective
(Radial Shadow), channel (Remove Color Matting), noise-grain (Dust & Scratches/
Noise Alpha2/Noise HLS2), transition (Radial Wipe/Block Dissolve). **Audio
Spectrum/Waveform kept**: their Audio Layer param defaults to None so the
template's tdpi pair is [host,host] (verified all-tdpi audit, no foreign id) —
splice-safe, though the audio source must be wired separately (no setter yet).
`TestAddEffectWave6_AEShipGate_AE2020/2025` — 23-in-one-run, both versions.
Still-wrong match-names parked: Bezier Warp (≠ `ADBE BezMesh`), Channel Mixer
(≠ `ADBE ChannelMixer`), + many tried-and-FAILed in lib6 (Warp/Mesh Warp/
Liquify/Reshape/Vegas/Radio Waves/Shadow-Highlight/Pixel Motion Blur/…) — their
real match-names need a probe pass (enumerate via canAddProperty sweep).

Wave 7 (2026-06-16, `re_effect_lib7.jsx` probe sweep → 46/46 canAdd, 45 kept):
**Lumetri Color** (`ADBE Lumetri`) + **Lightning** (`ADBE Lightning`, the old one;
`ADBE Lightning 2` = Advanced Lightning was wave 4) + **43 Cycore (CC) effects**
(CC Radial Fast/Radial Blur · Bend It/Bender/Blobbylize/Flo Motion/Griddler/Lens/
Page Turn/Power Pin/Ripple Pulse/Slant/Smear/Split/Split 2/Tiler/WarpoMatic ·
Light Burst 2.5/Light Rays/Light Sweep/Threads · Cylinder/Sphere/Spotlight ·
Glass/HexTile/Kaleida/Mr. Smoothie/Plastic/RepeTile/Threshold/Threshold RGB ·
Pixel Polly/Scatterize/Star Burst · Force Motion Blur/Wide Time · Color Offset/
Toner · Burn Film/Vignette/Simple Wire Removal).
- **KEY: 4 CC effects store a `CS …` internal match-name ≠ the `CC …` addProperty
  alias** — CC Cross Blur→`CS CrossBlur`, CC Threads→`CS Threads`, CC HexTile→
  `CS HexTile`, CC Vignette→`CS Vignette`. `addProperty("CC X")` accepts the alias
  but the stored tdmn (what `collectEffects` reads back + what `cloneEffectTemplate`
  must be keyed on) is the `CS` form. So the const VALUE = the stored `CS …` name.
  Lesson: when the parade readback name ≠ the addProperty name, key the template
  map + const on the **readback** name (gate's paradeChildNames compares readback).
- **CC Vector Blur EXCLUDED**: all-tdpi audit (throwaway scanner) showed tdpi=[15,0]
  — its "Vector Map" layer pickwhip defaults to None=0 (foreign vs host 15);
  retarget-all would point it at self. Same curation rule as Set Matte siblings.
  Other CC layer-pickwhip effects (Glass bump / Page Turn back-page / Blobbylize
  blob / Mr. Smoothie) audited CLEAN (no foreign tdpi when default None) → kept.
`TestAddEffectWave7_AEShipGate_AE2020/2025` — 45-in-one-run, both versions.

Wave 9 (2026-06-16, `re_effect_lib9.jsx` BIG probe sweep → 43 NEW kept): keying
(Color Key/Color Range/Extract/Luma Key/KeyCleaner) · channel (Channel Combiner)
· color/utility (Broadcast Colors/ProfileToProfile/Cineon Converter/GROW BOUNDS/
Median/Noise HLS Auto) · perspective (Basic 3D/Geometry-legacy) · time (Time
Displacement/Timecode) · transition (Gradient Wipe) · controls (Layer Control) ·
+ 25 more Cycore CC (Ball Action/Bubbles/Composite/Drizzle/Environment/Glass Wipe/
Glue Gun/Grid Wipe/Hair/Image Wipe/Jaws/Light Wipe/Mr. Mercury/Particle Systems II/
Radial ScaleWipe/Rain/Scale Wipe/Snow/Twister/BlockLoad/Color Neutralizer/Kernel/
LineSweep/Rainfall/Snowfall). `TestAddEffectWave9_AEShipGate_AE2020/2025`.
- **Probe-sweep traps (unattended automation)**: 3 effects open a MODAL on add →
  hang ae_run (exit 2): **Apply Color LUT** (LUT file picker), **PS Arbitrary Map**
  (map file picker), **Numbers** (font dialog). Excluded — can't add headlessly.
- **6 more CS-prefix stored names** (≠ CC alias): BlockLoad/Color Neutralizer/
  Kernel/LineSweep, and Rainfall=`CSRainfall`/Snowfall=`CSSnowfall` (no space).
- **Dropdown Control excluded**: its match-name is a per-instance PSEUDO
  (`Pseudo/@@…` random) — not a stable reusable key.
- **4 foreign-tdpi layer-ref** (tdpi=[15,0…]): 3D Glasses · Warp Stabilizer
  (`ADBE SubspaceStabilizer`) · Timewarp · CC Particle World — **SHIPPED wave 11
  (2026-06-17)** via the materialized-template + SetEffectLayerParam path (see
  § Wave 11 below).
- Genuinely unavailable in AE 2020 (canAdd=false, removed/renamed): Add/Remove
  Grain, Turbulent Noise, Cartoon, Ellipse, Radio Waves, Iris Wipe, Pixel Motion
  Blur, Shadow/Highlight, Selective Color, Liquify, HDR Compander, Foam/Wave
  World/Caustics/Shatter/Particle Playground, Card Wipe, Spill Suppressor.

**Curation rule:** only **parameter-only** effects. Effects with layer/path
**reference** params (e.g. Set Matte, Displacement Map, Calculations, Compound
Blur's "layer" pickwhip) would carry a dangling layer-id in their sspc that a
standalone splice can't satisfy — excluded until a Phase-2 remap handles refs.

## Coverage / gate

- `TestAddEffect_AllTemplates_RoundTrip` — all 30 splice + WriteAEP + re-parse
  (Go, no AE; iterates `SupportedEffects()`, so new templates are auto-covered).
- `TestAddEffect_AEShipGate_AE20{20,25}` — table-driven over the FULL
  29-template library: wave 1 24/24 PASS (2026-06-11; originally a 5-effect
  payload-size sample, the remaining 7 promoted to gated), wave 2 34/34 PASS
  (2026-06-11, both versions): AE opens the Go-added file without corruption,
  reads back 4 effects in order, and AE's own resave preserves the addition.
- Pre-commit template hygiene: `tools/debug/effect_id_scan <bin> 15` on each new
  template — non-tdpi hits in `pard`/`pdnm`/`fnam` are coincidental definition
  bytes (gated wave-1 templates show the same pattern); only tdpi carries the
  host binding (retargeted at splice).
- `TestAddEffectAutoParade_AEShipGate_AE20{20,25}` — parade auto-create end to
  end on a 100% Go-built file (fresh project → shape layer → Reopen →
  AddEffect), 2/2 PASS incl. resave preservation.
- Go-only: `TestAddEffect_AutoCreateParade_ReopenedFreshLayer` (self-contained,
  parade-before-Transform order assert) / `_ParsedFixtureLayer` (AE-native
  parade-less layer, re_text.aep) / `TestAddEffect_RefuseCameraLight` /
  `TestReopen_WriteStable` / `TestAddEffect_RetargetsTdpiHostLayer` (white-box
  tdpi remap).

## Parade auto-create (Phase 2 — SHIPPED 2026-06-10)

Landed as `ensureEffectParade` + `aep.Reopen`, double-version ship-gated
(`TestAddEffectAutoParade_AEShipGate_AE20{20,25}` — 100% Go-built file:
NewProject → NewComposition → NewShapeLayer → Reopen → AddEffect, AE opens
clean, reads back the effect, keeps it across resave):

- **Auto-create (parsed layers):** when `EffectsParade() == nil`, AddEffect
  splices `tdmn("ADBE Effect Parade") + LIST(tdgp){tdsb 0x00000001,
  tdsn "-_0_/-", tdmn("ADBE Group End")}` into the layer's outer tdgp
  **immediately before the `ADBE Transform Group` tdmn** — AE's emitted order on
  every observed layer kind (shape: after Root Vectors Group; AV/solid: after
  Time Remapping). The tdsn is AE's never-renamed placeholder name `-_0_/-`,
  NOT an empty string. Scene tree gets the mirroring AEPropertyGroup node at the
  same anchor; rollback restores both splices.
- **Fresh (built) layers still refuse** — no property tree to splice into, and
  `syncShapeLayerChunks` re-lowers dirty shape layers at write (chunk edits
  discarded). The refuse error points at `aep.Reopen(p)`: one write→parse round
  trip upgrades every built layer to a parsed one, after which auto-create +
  full param fidelity work. Reopen is byte-stable (rewrite == original write).
- **Camera / light layers refuse** (AE does not allow effects on them; we never
  create a parade there).
- AddMask can reuse the same auto-create pattern for Mask Parade (anchor scan +
  empty-group bytes TBD for masks).

## Phase 2 original findings (2026-06-10 investigation, pre-ship)

Concrete findings kept for reference:

1. **Parade position in a shape layer** (RE `re_shape_effect.aep`, an AE-native
   shape layer + Gaussian Blur): the outer `LIST(tdgp)` group order is
   `tdsb, tdsn, [Root Vectors Group], [Effect Parade], [Transform Group],
   [Layer Styles], [Extrsn Options], [Material Options], [Audio Group],
   [Layer Sets], Group End`. **Effect Parade sits immediately after Root Vectors
   Group, before Transform Group.**
2. **AE emits the Effect Parade ONLY when ≥1 effect exists.** The from-scratch
   shape templates (`v2_2_*`) were extracted from effect-less AE shape layers and
   carry **no** parade — so emitting an *empty* parade is non-canonical (AE never
   does it for an effect-less layer). A populated parade is the faithful form.
3. **Write-time re-lowering wall.** `syncShapeLayerChunks` (back_project.go
   WriteAEP) re-lowers **every dirty shape layer from scratch**
   (`lb.layrList.Children = fresh.Children`). So **chunk-level edits to a fresh
   shape layer's parade are discarded at write** — effects must be emitted by
   `lowerShapeLayer` from **scene** state, not patched into the chunk (unlike
   Phase 1, which patches a *parsed* layer's chunk that is never re-lowered).
4. **Scene can't hold rifx chunks** (scene 禁 import rifx). So the
   effect's sspc payload (and any tuned param values) can't live on the scene
   `*Effect`. Emitting effect *parameter values* set on a fresh layer therefore
   needs either (a) generic effect-param lowering (re-encode `Effect.Parameters`
   → sspc — large), or (b) a serializer-side "pending effect chunks" map keyed by
   `*Layer` that `lowerShapeLayer` consumes. Default-param effects (no tuning)
   are tractable via (b) but the "tuned-value silently lost" footgun must be
   handled (refuse/warn on SetStaticValue for fresh-layer effects).

The "write → reopen" workaround in these findings became the shipped path:
`aep.Reopen` is that round trip as a one-liner, and auto-create removes the
"once it has effects" precondition.

## Other deferred

- ~~**Reference-param effects** (Set Matte …)~~ — Set Matte SHIPPED 2026-06-15
  (`aep.SetEffectLayerParam` + `EffectSetMatte` lib template; dual-version render
  gate `TestSetMatte_AEShipGate_*`). RE (`re_set_matte.aep` + `tmp_debug/dump_set_matte`):
  a layer-reference param stores the **target layer's ID in the param's own tdpi**
  — the same 4B binding the always-present `-0000` host stream uses, just aimed at
  another layer (host tdpi=17, matte-source `-0001` tdpi=15). So the setter is a
  length-preserving 4B tdpi rewrite; no selective-remap machinery needed after all.
  Note AddEffect's `retargetEffectHostLayer` rewrites ALL tdpi to the host on add
  (incl. `-0001` → self-matte, a harmless no-op default for opaque layers); the
  setter then aims `-0001` at the real source. Displacement Map / Compound Blur /
  path-reference params: same mechanism, add on demand. Finding 5's "retarget-all
  corrupts ref params" caveat now means "call SetEffectLayerParam after AddEffect".
- ~~**Displacement Map / Compound Blur / CC Vector Blur**~~ — Wave 8 SHIPPED
  2026-06-16 (the layer-ref family, fixture `re_effect_layerref.aep`). RE: each
  effect's layer param is materialized by pointing it at a layer in the fixture
  (`pr.propertyValueType === PropertyValueType.LAYER_INDEX` → setValue(idx)), so
  the extracted template carries the param WITH a tdpi (Displacement Map `-0001`
  "用户图层" tdpi=15, Compound Blur `-0001` "模糊图层", CC Vector Blur `-0005`
  "Vector Map"; host `-0000`=17). AddEffect retargets both → host; SetEffectLayerParam
  aims the ref at the real source — same flow as Set Matte. Library consts +
  layer-ref param consts (`EffectDisplacementMapLayer` etc.) added.
  - **Render gate gotchas** (`mg_effect_layerref_shipgate_test.go`):
    (1) **A video-OFF map layer reads as EMPTY** → zero displacement (first gate
    try set MAP invisible, got delta=0). Fix: keep MAP video-ON, park it at the
    bottom behind a full-frame black BG solid so it does not composite over HOST
    (HOST = white-left shape, transparent right → black). (2) **Displacement Map
    param order**: `-0002` = "Use For Horizontal Displacement" (a dropdown),
    `-0003` = "Max Horizontal Displacement" (the AMOUNT). Setting `-0002` did
    nothing visible; `-0003`=180 shifts. (3) Displacement direction is NEGATIVE
    for white(255) → seam moves LEFT; the two-point ±90px flip test (x1050→white
    OR x870→black) is direction-agnostic. Displacement Map + Compound Blur are
    render-pixel double-version gated; **CC Vector Blur = accept+round-trip+resave
    only** (gradient-driven blur has no clean spatial pixel proof here —
    render-pixel deferred).
- ~~**AddMask**~~ — SHIPPED 2026-06-11 (dual-version gated, from-scratch atom,
  path parameterizable at creation; the auto-create pattern transferred via
  `spliceEmptyParade`). See [[add-mask-create-re]]. NOTE the earlier "same
  INDEXED_GROUP splice" framing was half-right: a mask atom is a (tdmn, mkif,
  tdgp) TRIPLE, so RemovePropertyGroup refuses mask children (pair assumption)
  — RemoveMask stays deferred.
- **Library expansion** beyond the 31 (more fixture RE; ref params now tractable
  via SetEffectLayerParam).
- **Per-effect typed param helpers** (today: raw `Property.SetStaticValue` by match-name).

Typed effect match-name constants (`aep.EffectGaussianBlur` … `aep.EffectExposure`,
single-sourced in the registry, divergence-guarded by `TestEffectConstants_MatchRegistry`)
shipped 2026-06-10 — call sites no longer hardcode AE's internal strings.

## 2026-06-17 — Wave 10: audio-processing effects (need an audio layer)

加齐 10 个音频效果（库 193→203）：Backwards `ADBE Aud Reverse` · Bass&Treble `ADBE Aud BT` ·
Delay `ADBE Aud Delay` · Flange&Chorus `ADBE Aud_Flange`（注下划线，非空格）· High-Low Pass
`ADBE Aud HiLo` · Modulator `ADBE Aud Modulator` · Parametric EQ `ADBE Param EQ`（off-pattern，
非 `ADBE Aud …`）· Reverb `ADBE Aud Reverb` · Stereo Mixer `ADBE Aud Stereo Mixer` · Tone
`ADBE Aud Tone`。

**关键差异：音频效果只能挂到有音频的层**——`parade.canAddProperty("ADBE Aud BT")` 在 solid 上
返回 false。所以 RE fixture 不能用 solid，必须 import 一个 mp3 当音频层：`importFile` → `comp.layers.add(foot)`
（`foot.hasAudio=true`）→ 在该层 parade 上 addProperty。fixture = `re_effect_audio.jsx`
（探针 + 全加 + 存 `re_effect_audio.aep`），探针扫候选 match-name 用 canAddProperty 过滤、读回 stored 名。
extract 走既有 `tmp_debug/extract_effect_lib`（音频效果 sspc 结构与其它一致：tdpi-host 绑定，retarget 照常）。
⚠ **同一 effect 加两次**：第二个实例字节更小（686B vs 1502B，AE 复用/elide），extract 的 last-write-wins
会抓到坏的——fixture 里每个效果只加一次。

**ship-gate（非视觉 → ae-accept + DOM readback，无渲染像素）**：`TestAddEffectAudio_AEShipGate_AE2020/2025`
从 clean 音频 base fixture（`re_audio_base.aep`，mp3 层零效果）`aep.Open` → AddEffect 全 10 个 →
WriteAEP → AE 开 + 读回 parade 全 10 名按序 + resave 存活。双版本 PASS。复用 `verify_property_struct.jsx`
（找首个非空 parade 比 match-name）。base fixture 内嵌 mp3 故 gitignored，缺则 gate skip。

## 2026-06-17 — Wave 11: the 4 foreign-tdpi layer-ref effects (library 203→207)

收口 wave 9 deferred 的 4 个 layer-ref 效果：**3D Glasses (`ADBE 3D Glasses`) · Warp
Stabilizer (`ADBE SubspaceStabilizer`) · Timewarp (`ADBE Timewarp`) · CC Particle
World (`CC Particle World`)**。同 wave-8 物化流程（`re_effect_layerref2.jsx`）：HOST
solid 上 addProperty 每个效果，**递归**找 LAYER_INDEX param 指向 MAP 层物化（这些复杂
效果的 layer pickwhip 嵌在子 group 里，不像 Displacement Map 在顶层——wave-1 顶层-only
遍历会漏），存 fixture，`extract_effect_lib` 抽出带 tdpi 的模板。

**关键 RE 发现**：
- **3D Glasses & Timewarp 各有 TWO layer-ref param**（wave 9 的 all-tdpi 审计只数到 1，
  漏数）：3D Glasses `-0001`/`-0002`（左/右视图）· Timewarp `-0029`/`-0031`（matte/source，
  语义按 param 顺序推测——中文 AE 的 UI 名是 CJK）· Warp Stabilizer `-0046`（单个，
  reference）· CC Particle World `-0045`（Texture Layer）。tdpi 计数实测 = host + N：
  3d_glasses 3 · subspacestabilizer 2 · timewarp 3 · particle_world 2。
- **Warp Stabilizer 可加到 solid**（canAdd=true，无分析 modal）——推翻了"分析类效果加不上/
  弹 modal hang"的预判。AddEffect + SetEffectLayerParam → MAP 后 AE 2020+2025 都接受。
- 模板较大（CC Particle World 32KB · Warp Stabilizer 13KB · Timewarp 12KB）但 splice 照常
  accept；retargetEffectHostLayer 递归重写全部 tdpi → host，SetEffectLayerParam 再逐 param
  指真实 source（多 layer-ref 按 param match-name 区分，机制无需改）。

**ship-gate（非渲染 → accept + DOM readback + resave）**：`TestAddEffectWave11_AEShipGate_
AE2020/2025`——100% Go-built 文件（NewProject→Comp→HOST/MAP solid→Reopen→AddEffect ×4 +
SetEffectLayerParam 每个 layer-ref param→MAP→WriteAEP），双版本 AE 开 + 读回 4 名按序 +
resave 存活。**render-pixel 诚实 deferred**（同 CC Vector Blur 先例）：这 4 个的视觉作用面
无干净单帧像素证明（Warp Stabilizer=分析驱动 · Timewarp=时间重映射 · CC Particle World=
程序化 · 3D Glasses=立体合成）。Go round-trip `TestLayerRefWave11_GoRoundTrip` 验 6 个
layer-ref param 的 tdpi 物化全部回读 MAP id。新 const：4 effect match-name + 6 layer-ref
param（`EffectWarpStabilizerRefLayer`/`Effect3DGlassesLeftView`/`…RightView`/
`EffectTimewarpMatteLayer`/`…SourceLayer`/`EffectCCParticleWorldTexture`）。**Curation rule
扩展**：layer-ref 效果只要 RE 时把 pickwhip 物化（带 tdpi 出模板）即可入库，不再受
parameter-only 限制——`SetEffectLayerParam` 处理 ref 重指。

## 2026-06-17 — Wave 12: wave-6 parked 经典效果（库 207→216，全大写 match-name 是关键）

收回 wave 6 parked 的经典效果——当时 camelCase 猜名全 FAIL，根因是**老 AE effect match-name
是 ALL-CAPS-带空格且大小写敏感**（同 chunk-id 大小写敏感 [[chunk-id-case-tdb4]]）。probe11
（`re_effect_probe11.jsx` 试大写/格式变体）一次定位 9 个正确名：

- **5 parameter-only**：`ADBE BEZMESH`(Bezier Warp) · `ADBE MESH WARP` · `ADBE CHANNEL MIXER` ·
  `ADBE RESHAPE` · `ADBE Vector Paint`。
- **4 layer-ref**（probe 顺带发现，物化同 wave 11）：`ADBE Texturize`(-0001) · `ADBE Color Link`
  (-0001) · `ADBE Compound Arithmetic`(-0001) · `ADBE Set Channels`(**-0001/-0003/-0005/-0007 四个
  source layer**)。

**关键教训**：camelCase 变体（`ADBE BezierWarp`/`ADBE ChannelMixer`/`ADBE Mesh Warp`）全
`canAdd=false`，唯有全大写（`ADBE BEZMESH`/`ADBE CHANNEL MIXER`/`ADBE MESH WARP`）+ `ADBE RESHAPE`
（大写）能加——match-name 大小写敏感。`ADBE Reshape`(首字母大写) FAIL，`ADBE RESHAPE` OK。

**确认不可达**（AE 2020 canAdd=false，非名字错）：Vegas（所有变体）· Warp · PS Express · Smear ·
Wave World2 —— 这些 AE 2020 真没有（renamed/removed）。

extract 走 `re_effect_lib12.jsx`（递归物化 layer-ref + 读 param matchName，一文件抽 9 个），
tdpi 审计：param-only 全 1(host)、layer-ref 1+N（Set Channels host+4）。**ship-gate**
`TestAddEffectWave12_AEShipGate_AE2020/2025`——100% Go-built（HOST/MAP solid + AddEffect ×9 +
SetEffectLayerParam 全 layer-ref param→MAP），双版本 accept + 读回 9 名按序 + resave 存活。
非渲染（distortion/channel 类无干净单帧 layer-ref 像素证明）→ accept+readback+resave。
新 const：9 effect match-name + 7 layer-ref param。
