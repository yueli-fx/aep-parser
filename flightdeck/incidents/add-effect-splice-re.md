---
status: active
when_to_read: implementing or extending AddEffect / the effect-template library; adding a new effect to the embedded set; debugging "AE drops/rejects a Go-added effect" or "cannot find layer ID=N in composition" on open; deciding whether an effect is splice-portable; extending parade auto-create to another group kind (Mask Parade); reasoning about the (tdmn, sspc) effect chunk unit or the tdpi host-layer binding
applies_to: [add-effect, effect-parade, sspc, tdmn, tdpi, host-layer-binding, effect-template, structural-write, splice, group-end-sentinel, parade-auto-create, reopen, version-portable, ae2020, ae2025, ship-gate, embed-fs, gaussian-blur, levels]
last_updated: 2026-06-16
---

# AddEffect — Effect Parade splice RE + ship findings

`aep.AddEffect(layer, matchName)` appends a built-in effect to a layer's
`ADBE Effect Parade`. Shipped 2026-06-10 (two commits: core mechanic + 12-effect
library). AE 2020 + AE 2025 ship-gate 24/24 across the full 12-effect library
(2026-06-11; initially shipped 2026-06-10 on a 5-effect payload-size sample).
Promoted Alpha → Stable (structural op, CLAUDE.md #2) on the full-library gate,
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

## Effect-template library (193, embed.FS)

`internal/serializer/templates/effect_adbe_*.bin`, each a `LIST(tdgp)` wrapper
around one `(tdmn, sspc)` pair, extracted from AE-2020 fixtures
`test_data/re_effect_library.aep` (wave 1, 12 effects) +
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
- **4 foreign-tdpi layer-ref deferred** (tdpi=[15,0…]): 3D Glasses · Warp
  Stabilizer (`ADBE SubspaceStabilizer`) · Timewarp · CC Particle World — need
  the materialized-template + SetEffectLayerParam path (future layer-ref wave).
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
- Pre-commit template hygiene: `tmp_debug/effect_id_scan <bin> 15` on each new
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
4. **Scene can't hold rifx chunks** (CLAUDE.md #3: scene 禁 import rifx). So the
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
