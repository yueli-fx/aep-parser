---
status: active
when_to_read: implementing or extending gradient write (SetGradient / GradientFillNode / GradientStroke); encoding AE gradient prop.map XML; debugging "AE drops the gradient" or stops not surviving resave; deciding whether a gradient fixture is AE-version-portable
applies_to: [gradient, g-fill, grad-colors, prop-map, xml-encode, gcst, gcky, utf8, shape-layer, length-variable, ship-gate, ae2020, ae2025, elision]
last_updated: 2026-05-31
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

Consequence: the extracted G-Fill template (`v2_2_shape_gradfill_body.bin`)
contains ONLY `ADBE Vector Grad Colors` — Grad Type / Start Pt / End Pt were
default in the source and AE elided them, so there is no slot to overwrite.
**V2.2 models the color/alpha stops only; the ramp geometry stays at AE's
default linear** (deferred — would need a fixture with Type/Start/End set
non-default, which requires UI authoring since JSX can't set the stops either).

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
