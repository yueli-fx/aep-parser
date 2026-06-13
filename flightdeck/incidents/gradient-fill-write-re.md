---
status: active
when_to_read: implementing or extending gradient write (SetGradient / GradientFillNode / GradientStroke / gradient ramp direction Start·End Pt); encoding AE gradient prop.map XML; debugging "AE drops the gradient" or stops/direction not surviving resave; deciding whether a gradient fixture is AE-version-portable; wanting to control the linear ramp angle/direction
applies_to: [gradient, g-fill, grad-colors, grad-start-pt, grad-end-pt, ramp-direction, prop-map, xml-encode, gcst, gcky, utf8, shape-layer, length-variable, ship-gate, ae2020, ae2025, elision, mg-roadmap, s5, render-pixel]
last_updated: 2026-06-13
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
