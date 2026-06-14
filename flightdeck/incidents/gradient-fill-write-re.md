---
status: active
when_to_read: implementing or extending gradient write (SetGradient / GradientFillNode / GradientStroke / gradient ramp direction Start·End Pt / radial type / radial HiLite Length·Angle highlight); encoding AE gradient prop.map XML; debugging "AE drops the gradient" or stops/direction/highlight not surviving resave; deciding whether a gradient fixture is AE-version-portable; wanting to control the linear ramp angle/direction or shift a radial gradient's bright centre
applies_to: [gradient, g-fill, g-stroke, gradient-stroke, grad-colors, grad-start-pt, grad-end-pt, grad-type, hilite-length, hilite-angle, radial-highlight, ramp-direction, prop-map, xml-encode, gcst, gcky, utf8, shape-layer, length-variable, ship-gate, ae2020, ae2025, elision, mg-roadmap, s5, render-pixel]
last_updated: 2026-06-14
---

# Gradient fill write (SetGradient) — RE + ship findings

V2.2.1 子项⑭. Gradient **read** already existed (`ParseGradientXML` + the
GCst→GCky→Utf8 parse path). This is the **write** path: a from-scratch gradient
fill via the V2.2 shape builder.

## On-disk structure

A gradient fill node (`ADBE Vector Graphic - G-Fill`) carries the stops in:

```
tdmn("ADBE Vector Grad Colors")
LIST(GCst)
  LIST(tdbs)          base metadata (tdsb / tdsn / tdb4-124B / cdat-4B placeholder)
  LIST(GCky)          gradient-keyframe container
    Utf8              prop.map XML (one per keyframe; static = the single leaf)
```

The XML (`<prop.map version='4'>`) holds, under `Gradient Color Data`, **Alpha
Stops first then Color Stops** — each a `Stops List` of `Stop-N` entries plus a
`Stops Size` int. A color stop's `Stops Color` array is **6 floats**
`[offset, midpoint, r, g, b, 1]` (the read parser uses the first 5; emit the
trailing `1` to match AE). Alpha stop `Stops Alpha` = 3 floats
`[offset, midpoint, alpha]`. A trailing sibling pair `Gradient Colors` =
`<string>1.0</string>` closes the map. Lines are newline-separated, no indent.

## Length-variable write is nearly free

Overwriting the Utf8 XML changes its byte length per stop count. No manual size
fixup is needed: `rifx.Chunk.Write` recomputes every enclosing LIST size
bottom-up from current child data (same machinery `Footage.SetPath` relies on).
The GCst's `tdb4` (124B) is NOT a redundant length header — AE 2020 + 2025 both
accept a re-emitted Utf8 of arbitrary length without it being touched. So the
serializer just swaps `Utf8.Data = EncodeGradientXML(g)` and lets Write reflow.

## Two elision traps (why only the colors are modeled)

1. **The default gradient is elided.** A gradient fill with default 2-stop
   black→white emits *no* GCst at all (re_gradient.aep, AE 2020 JSX: set Grad
   Type/Start/End but the colors chunk was absent). Only a **non-default** stop
   set forces AE to emit GCst.
2. **ExtendScript can't author custom stops.** `ADBE Vector Grad Colors` is XML
   with no typed `setValue`. So no JSX fixture can carry custom stops — the only
   stops-bearing fixture is `v2_2_gradient_src.aep` (← py-aep's gradient.aep),
   which is **AE 25.6-saved**.

Consequence (original): the extracted G-Fill template contained ONLY `ADBE
Vector Grad Colors` — Grad Type / Start Pt / End Pt were default → elided → no
slot. **V2.2 originally modeled the color/alpha stops only; the ramp geometry
stayed at AE's default linear.**

### UPDATE 2026-06-13 — ramp DIRECTION (Start/End Pt) now resolved ✅

The "needs UI authoring" worry was wrong. **JSX CAN author the typed point props
`ADBE Vector Grad Start Pt` / `End Pt`** (only the `Grad Colors` stops XML lacks
a typed setValue). So: open the existing stops-bearing `v2_2_gradient_src.aep` in
AE 2025, `gfill.property("ADBE Vector Grad Start Pt").setValue(...)` +
`End Pt` non-default (a diagonal), resave → AE emits the slots alongside the
already-present stops. Re-extracted `v2_2_shape_gradfill_body.bin` (now **9
children** vs 5; `tmp_debug/gen_gradient_dir.jsx` → `v2_2_gradient_dir.aep`).
- G-Fill defaults (probed): Grad Type=1 (Linear) · **Start Pt=[0,0]** · **End
  Pt=[100,0]** (horizontal ramp) · HiLite Length/Angle=0. 10 DOM children, only
  non-default ones emit.
- `lowerGradientFillNode` overwrites Start/End Pt (Vec2 2×f64 BE @cdat[0:16], same
  layout as Repeater Transform points) with the node's StartPoint/EndPoint;
  default [0,0]→[100,0] reproduces the old behavior (no regression — existing
  `TestV2_2_GradientFill_AEShipGate` still PASS with the 9-child template).
- Gate `TestMGGradientDir_AEShipGate_AE2020/2025` PASS: red→blue diagonal ramp,
  corners TL=red BR=blue TR=BL=mid-purple (horizontal ramp would make TR blue /
  BL red). `GradientFillNode.SetStartPoint/SetEndPoint`.
- **Still deferred**: Grad Type (radial) · HiLite · gradient STROKE direction
  (G-Stroke template unchanged) · read-back of direction (hydrate unchanged —
  write-only from-scratch, consistent with the filter nodes; untouched parsed
  gradients stay opaque-preserved, mutate-sync has known partial fidelity).

### UPDATE 2026-06-14 — Grad TYPE (radial) now resolved ✅

Same template-re-extraction blueprint as the direction work. **`ADBE Vector
Grad Type` is a typed enum** JSX can `setValue` — **range is [1,2]: 1=Linear,
2=Radial** (NOT 1/3; `setValue(3)` throws "值 3 在 1 到 2 的范围外"). Authored
on top of `v2_2_gradient_dir.aep` (Grad Type=2 + Start/End Pt kept non-default
so all three geometry slots emit) → `v2_2_gradient_type.aep`
(`tmp_debug/gen_gradient_type.jsx`). Re-extracted `v2_2_shape_gradfill_body.bin`
is now **15 children** (vs 9): + Grad Type, + HiLite Length/Angle (AE emits the
HiLite pair too once the gradient is radial).
- Grad Type is **1D f64 BE enum @cdat[0:8]** (same family as Merge/Twist enums).
  `lowerGradientFillNode` overwrites it with `float64(n.GradientType())`; default
  `GradientLinear`=1 overwrites the baked Radial(2) → **no regression** (linear
  `TestMGGradientDir` / `TestV2_2_GradientFill` still PASS on the 15-child
  template, AE2020).
- API: `GradientFillNode.SetGradientType(GradientLinear|GradientRadial)` +
  `GradientType()` getter. For radial, StartPoint = centre, EndPoint sets the
  outer radius.
- Gate `TestMGGradientRadial_AEShipGate_AE2020/2025` PASS (red line 4): red
  centre → blue edge, the 4 cardinal points at radius 140 render **identical
  (76,0,178)** = rotational symmetry (a linear ramp would split L=red/R=blue);
  Grad Type=2 read back; resave survives. Both AE versions byte-identical pixels.
- **Still deferred**: HiLite tuning (slots present but untouched, length stays
  embed default) · gradient STROKE type/direction (G-Stroke template unchanged)
  · read-back of type/direction (hydrate write-only, as before).

### UPDATE 2026-06-14 — radial HIGHLIGHT (HiLite Length/Angle) now resolved ✅

No template re-extraction needed — the 15-child radial body already carries the
`ADBE Vector Grad HiLite Length` / `ADBE Vector Grad HiLite Angle` slots (AE
emits the HiLite pair the moment the gradient is radial). Both are **1D f64 BE
at cdat[0:8]** (len-40 cdat, same scalar family as Grad Type / the filter enums),
baked at 0 in the template. `lowerGradientFillNode` now overwrites both with the
node values; default 0/0 overwrites the baked 0 with no visible change →
**no regression** (radial/dir/fill gates still PASS).
- **HiLite Length** = highlight offset magnitude as **percent of the radius**
  (range [-100,100]; 0 = centred). **HiLite Angle** = offset direction in
  **degrees**. They only affect a *radial* gradient (linear ignores them).
- **Angle convention (measured, AE2020+2025 identical):** Angle 0° shifts the
  bright centre (start color) toward **+X (right)**. At Length 70 / Angle 0 the
  red hotspot moves 0.7·radius right: cardinals at radius 130 render
  R=(244,0,10) red · L=(52,0,202) blue · U==D=(79,0,175) — a clean one-axis
  split (X) with the perpendicular (Y) pair byte-identical.
- API: `GradientFillNode.SetHighlightLength` / `SetHighlightAngle` +
  `HighlightLength()` / `HighlightAngle()` getters.
- Gate `TestMGGradientHilite_AEShipGate_AE2020/2025` PASS (red line 4): asserts
  the highlight breaks symmetry along **exactly one** axis (the radial gate's
  inverse) — `max(dx,dy)>60 && min(dx,dy)<45` + the matched pair stays symmetric;
  HiLite Length=70 read back; resave survives. Both AE versions byte-identical.
- **Still deferred**: gradient STROKE type/direction/highlight (G-Stroke template
  unchanged — needs its own re-extraction) · read-back of type/direction/highlight
  (hydrate write-only, as before).

### UPDATE 2026-06-14 — gradient STROKE ramp geometry (direction + type) now resolved ✅

Same template-re-extraction blueprint as G-Fill, applied to `ADBE Vector Graphic
- G-Stroke`. The original G-Stroke template (`v2_2_shape_gradstroke_body.bin` from
`v2_2_gradient_src.aep`) had DEFAULT ramp geometry → Grad Type/Start/End/HiLite
all elided (only Grad Colors + the stroke geometry Width/Cap/Join/… present).
- **Authored the slots:** `tmp_debug/gen_gradstroke_geom.jsx` opens the stops-
  bearing `v2_2_gradient_src.aep`, sets the **G-Stroke's** `ADBE Vector Grad Type`
  = 2 (Radial) + Start Pt=[-120,-120] / End Pt=[120,120] (non-default), resaves
  → `v2_2_gradstroke_geom.aep`. Setting Type=2 makes AE emit the **HiLite pair
  too** (same as G-Fill). Re-extracted via `extract_shape_bodies` (now sourced
  from the geom fixture): **29 children** (vs ~25), carrying all 5 gradient-
  geometry slots. Only `v2_2_shape_gradstroke_body.bin` changed.
- **cdat layout identical to G-Fill:** Grad Type / HiLite Length / HiLite Angle =
  1D f64 BE @cdat[0:8]; Start/End Pt = Vec2 @cdat[0:16]. `lowerGradientStrokeNode`
  overwrites all five (default Linear=1 / [0,0]→[100,0] / HiLite 0/0 reproduces
  AE's pre-geometry stroke → **no regression**: existing stops-only
  `TestV2_2_GradientStroke_AEShipGate_AE2020/2025` still PASS on the 29-child
  template, both versions).
- API: `GradientStrokeNode.SetStartPoint/SetEndPoint` + `SetGradientType` +
  `SetHighlightLength/SetHighlightAngle` (+ getters) — full parity with
  GradientFillNode.
- Gate `TestMGGradStrokeGeom_AEShipGate_AE2020/2025` PASS (red line 4), three
  18px-stroked rects in one comp: **GSDIR** linear left→right ramp renders
  left=pure red (253,0,1) / right=pure blue (1,0,253) → direction works on a
  stroke; **GSRAD** radial (radius 200) ring-mids at radius 100 render the four
  cardinals **identical (127,0,127)** = rotational symmetry → type=radial works
  on a stroke; **GSHL** radial + HiLite Length 70 / Angle 0 (+X) renders the ring
  L=blue (74,b=180) / R=red (224,b=30) / U==D (r=107) = one-axis split → the
  highlight shifts the hotspot on the stroke too. Grad Type 1/2 + HiLite Length
  read back, resave survives, both AE versions byte-identical pixels.
- **Still deferred**: stroke geometry Dashes / Taper / Wave (the nested groups —
  kept at the template's defaults). Width / Cap / Join / Miter resolved below.
  (Type/direction/highlight read-back also resolved below.)

### UPDATE 2026-06-14 — ramp-geometry READ-BACK (hydrate) now resolved ✅

The hydrate path was stops-only ("write-only from-scratch"): opening a parsed
gradient surfaced the color stops but `GradientType()` / `StartPoint()` /
`EndPoint()` / `HighlightLength()` / `HighlightAngle()` returned the constructor
defaults regardless of the on-disk values (raw bytes were still opaque-preserved,
so resave stayed byte-identical — it was a *typed-accessor* gap, not data loss).
- `hydrateGradientGeometry` (shared by G-Fill + G-Stroke) reads the five geometry
  cdats from the body via `nodeStreamValues` + `scalarOf`/`vec2Of` and applies the
  node setters; absent (elided-default) slots leave the constructor default.
  Static only; setter range errors ignored (best-effort).
- **Purely additive to the read side** — write still opaque-preserves untouched
  nodes, so the full round-trip/golden suite stays green (no byte change).
- Gate `TestV2_2_GradientGeometry_Roundtrip` (pure Go: build radial G-Fill +
  G-Stroke with distinct non-default Start/End/Type/HiLite → write → re-parse →
  assert every geometry getter). PASS.
- Remaining gradient debt is now only **stroke geometry** (resolved below).

### UPDATE 2026-06-14 — gradient-stroke geometry (Width / Cap / Join / Miter) ✅

The G-Stroke template already carried these slots (baked Width=18, Cap=Round=2,
Join=Round=2, Miter=4 — no re-extraction needed). `GradientStrokeNode` gains
`SetStrokeWidth` / `SetLineCap` / `SetLineJoin` / `SetMiterLimit` (+ getters,
reusing the solid stroke's `StrokeLineCap`/`StrokeLineJoin` enums);
`lowerGradientStrokeNode` overwrites the four 1D f64 BE @cdat[0:8] slots. The
node **defaults to the baked template values**, so a stroke that overrides none
re-emits byte-identically → **no regression** (existing `TestMGGradStrokeGeom` /
`TestV2_2_GradientStroke` stay green).
- Gate `TestMGGradStrokeStyle_AEShipGate_AE2020/2025` PASS: a 60px stroke renders
  a point 25px outside the rect edge as pure red (inside the 60px band 810±30),
  while a point 42px out is background — an 18px default band (801..819) would
  fail this. AE DOM reads back Width=60 / Cap=3(Projecting) / Join=3(Bevel) /
  Miter=10; resave survives. Both versions identical. (Width is pixel-gated; Cap/
  Join/Miter are value-gated — their render effect is subtle corner/cap geometry.)
- **Still deferred**: the nested **Dashes / Taper / Wave** groups (kept at the
  template defaults) — same nested-group write as the solid stroke's, on demand.

## Cross-version: AE25-shaped gradient is accepted by AE 2020

The template source is AE 25.6 and **AE 2020 refuses to open that whole project
file** (version stamp). The open question was whether a *from-scratch* shape
layer (AE-2020-targeted project seed) carrying the AE25-extracted gradient body
+ a freshly-encoded XML would be accepted by AE 2020. **It is** — gradients are
an ancient/stable format; the GCst/GCky/Utf8 + prop.map v4 structure is
version-portable. Both `TestV2_2_GradientFill_AEShipGate_AE2020` and `_AE2025`
PASS, decoding the resaved stops back to the written red/green/blue. So one
AE25-sourced template serves both gate versions.

## Scope (子项⑭)

`(g *VectorGroup) AddGradientFill()` → `GradientFillNode`; `SetColorStops` /
`SetAlphaStops` (≥2 stops, ranges validated) + `Gradient()` getter. Encoder
`EncodeGradientXML` is the inverse of `ParseGradientXML` and round-trips through
it. Stops are static (animated gradients deferred). Gradient **stroke**
(`G-Stroke`) is a straightforward follow-on (same GCst path; the source fixture
has a G-Stroke body too).
