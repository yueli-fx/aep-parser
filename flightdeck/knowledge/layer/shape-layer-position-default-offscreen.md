# ⚠ From-scratch shape layer renders blank: layer Position defaults to (0,0), not comp-center

From-scratch shape layer renders blank: layer Position defaults to (0,0), not comp-center

## Signature
- symptom: `from-scratch shape layer renders blank in AE; verify.jsx shows all shapes+values correct; Layer.SetPosition → "Position property not present"`
- error_type: —
- where: NewShapeLayer / lowerShapeLayer (internal/serializer/lower_layer.go) · Layer.SetPosition (internal/scene/scene_layer_accessors.go)
- trigger: build a shape layer from scratch whose shape nodes use AE-relative (often negative) positions, then render

## Symptom / repro
Booyah 复刻 comp ① (4 animated rects + fill). Go round-trip clean, AE accepts the
file, `verify.jsx` dumps the full shape tree (4 `ADBE Vector Shape - Rect` + Fill),
and AE's DOM reads every value correctly (rect Size animates to 230×12 at t=0.5,
group opacity 100, fill cyan opaque). Yet `saveFrameToPng` renders a **blank**
frame at every time (identical byte size across t). Classic red-line-4: value-correct
≠ render-correct.

## Root cause
I assumed the cockpit/ledger note "结构 delta (组嵌套 + transform 默认物化) = render-neutral"
was true. It was not. **AE centers a new shape layer's Transform Position at comp-center
(e.g. 960,540 for 1920×1080); `NewShapeLayer` leaves it at (0,0) (top-left).** The
original's rects carry negative shape-space positions (e.g. [-452,-68]) measured from
the layer origin, so with the layer at comp-center they land on-screen, but with the
layer at (0,0) every rect maps to negative comp coords → off the top-left edge →
nothing visible. The DOM is entirely correct; only the layer's world placement is wrong.

Decisive diagnosis = read AE's computed values, not the stored ones: dump
`layer.Transform.Position.value` for BOTH the original and the clone — orig=[960,540,0],
clone=[0,0,0]. Everything else (group opacity/scale, rect Size.valueAtTime, fill) matched.

## Fix
Set the shape layer's position via its **own transform stream**, not the generic
Layer accessor:

- ✅ `sl.Transform().Position().SetStaticValue([2]float64{px, py})` — `LayerTransform.Position()`
  is the `*PropertyStream[[2]float64]` that `lowerShapeLayer` writes into the embedded
  transform-group template (overwrites Position_0/_1 cdat). px,py copied from the
  original layer's `ADBE Transform Group / ADBE Position` (replication: value from oracle).
- ❌ `Layer.SetPosition(v)` — goes through `l.Position()` = `PropertyByMatchName("ADBE Position")`,
  which is **nil** for a from-scratch shape layer (its transform is an embedded byte
  template, not a materialized scene property) → returns "Position property not present".
  This is the transform-group default-omission / property-synthesis gap ([[transform-group-default-omission]]).

After the fix: comp ① t=0 frame is byte-identical to the original, t=0.5 within 4 bytes.

**Library follow-up (not done):** `NewShapeLayer` arguably should center the layer at
comp-center by default (match AE) instead of (0,0). That changes a Stable-API default and
needs a gate sweep, so it's deferred; the gen sets position explicitly for now.

## Cases
- 2026-06-20 first — Booyah 复刻 comp ① rendered blank; root-caused to layer Position (0,0) vs comp-center; fixed in gen via Transform().Position().
- 2026-06-22 second — Booyah 复刻 comp ④ カクッ (2 stroked-rect+trim shape layers). **New twist: this time I DID read the original's position to copy it — and still landed at top-left.** The original's layer Transform stores Position **SEPARATED** (`ADBE Position_0` / `ADBE Position_1`, no unified `ADBE Position`), and the parser surfaces both channels as **0,0** even though the layer is visually centered (user confirmed on the real project). I trusted that read → didn't center → clone's rect (centered at shape-(0,0), (0,0) anchor) rendered with its center at comp (0,0) = top-left (render showed only the bottom-right quadrant's corner). **Fix:** center explicitly via `SetLayerTransform.Position()` = comp-centre (this comp animates Scale/Opacity so it uses the SetLayerTransform path, not `sl.Transform().Position()` like comp ①; both center the same way). After fix, render shows the trimmed rect centered (corners as two opposite brackets, L0's 180° rotation putting its trim-gap corner diagonally opposite L1's). **Read-side open question (not chased):** is this a parser MISREAD of the separated X/Y channels (should be 960,540) or genuine AE semantics (separated 0,0 + comp-centre base)? Needs an AE-DOM `xPosition`/`yPosition` dump of the original to settle. **Lesson:** for a separated-Position layer, do NOT trust the parser's Position_0/Position_1 as the world position during replication — center explicitly (or settle the read-side first). Related write-side: [[separate-dimensions-write-mechanics]].
