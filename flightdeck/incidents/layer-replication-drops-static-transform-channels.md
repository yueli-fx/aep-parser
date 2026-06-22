---
status: active
when_to_read: replicating a real layer's transform from-scratch (precomp / AV / shape / text) and the clone renders too big / un-rotated / off-position vs the original while the ANIMATED channels (Position/Opacity keyframes) look right; a nested-comp composite spreads wider or sits differently than the source though keyframe times match; deciding which transform channels to copy when mirroring an original layer; SetLayerTransform leaves Scale 100% / Rotation 0 because the gen never set them
applies_to: [layer-replication, setlayertransform, transform-channels, scale, rotation, anchor, static-non-default, from-scratch, precomp, booyah-clone, fidelity-gap, composite-mismatch, render-pixel, flightdeck/showcase/booyah-clone/gen_precomp1.go]
last_updated: 2026-06-22
resolved_by: comp ⑤ gen now copies static Scale+Rotate Z (booyah-clone, 2026-06-22); general lesson — replicate ALL transform channels, not just animated ones
---

# From-scratch layer replication silently drops static non-default transform channels

## Signature
- symptom: a replicated precomp/AV layer renders FULL-SIZE (Scale 100%) and un-rotated while the original is scaled-down / rotated; the nested composite's glitch spreads far wider than the original even though every keyframe TIME and the leaf source match pixel-for-pixel
- error_type: —
- where: `flightdeck/showcase/booyah-clone/gen_precomp1.go` (any gen building a `LayerTransform` for `SetLayerTransform`)
- trigger: the gen copies only the channels it "knows are animated" (Position keyframes, Opacity keyframes) and leaves Anchor/Scale/Rotation at `NewLayerTransform` defaults — but the original carries a **static, non-default** Scale/Rotation

## Symptom / repro

booyah comp ⑤ プリコンポジション 1 (3 nested copies of comp ①): the user's frame-18
side-by-side showed the clone's glitch bars spread far wider than the original. Bisected
by solo-rendering each layer: the **static** layer (L2, Scale 100%) matched pixel-for-pixel,
but the two **transformed** copies diverged. The leaf comp ① matched frame-for-frame, and
the AE DOM `keyTime`s matched to 4 dp — so it looked like a baffling "nested time" bug.
It was not time at all: the user opened the original's transform panel and saw
**Scale 44 % (L0) / 45 % + Rotation 180° (L1)** vs the clone's 100 % / 0°.

This masqueraded as a deep nested-time-mapping mystery for a long bisection (cdta / @0xA8 /
lhd3 / fingerprint all ruled out) **because the divergence correlated with "layer has
keyframes"** — but that was a coincidence: the two animated layers also happened to be the
two with non-default Scale/Rotation. **Lesson: don't pattern-match on "animated vs static";
read the actual transform values.**

## Root cause

`NewLayerTransform()` seeds Anchor 0,0 · Position 0,0 · **Scale 100 %** · **Rotation 0°** ·
Opacity 100. The gen set only AnchorPoint + Position(+kf) + Opacity(kf) and never read the
original's Scale or Rotate Z, so they stayed at the 100 %/0° defaults. A precomp layer at
44 % scale shows its source shrunk toward the anchor (compact); at 100 % it shows full-size
(spread) — hence the composite mismatch. The static-but-non-default channels are invisible
in a keyframe/`keyTime` dump (no keyframes to dump), so value-oracle checks that only walk
keyframed props miss them entirely (red-line 4: value-round-trip green ≠ render correct).

## Fix

Read EVERY transform channel from the original and apply it, animated or not. comp ⑤ gen now:
```go
if sc := findProp(otg, "ADBE Scale"); sc != nil {        // parser returns a FRACTION
    s := toFloats(sc.StaticValue)
    tr.Scale().SetStaticValue([2]float64{s[0]*100, s[1]*100})  // SetLayerTransform.Scale wants PERCENT → ×100
}
if r := findProp(otg, "ADBE Rotate Z"); r != nil {
    tr.Rotation().SetStaticValue(toScalar(r.StaticValue))      // degrees as-is
}
```
Units: `SetLayerTransform.Scale` is percent (lowerTransformScale divides by 100 → AE fraction);
the parser reads Scale as a fraction (0.44), so multiply by 100. Rotate Z is degrees both ways.
Verified: comp ⑤ composite frames 6 + 18 now match the original pixel-for-pixel (AE 2025),
user-accepted on real machine.

**General rule**: when mirroring a real layer, iterate the original's full Transform Group
(Anchor/Position/Scale/Rotation/Opacity + any 3D channels) and copy each — a "static default"
assumption per channel is a silent fidelity hole. Other booyah comps (②③④) may carry the same
omission; sweep their transform panels.

Related: [[shape-layer-position-default-offscreen]] (the separated-Position centering
workaround, a sibling transform-read trap) · [[setlayertransform-av-anchor-fraction]]
(anchor units) · [[ntsc-tickrate-derive-3x-off]] (the tdb4 keyframe-TIME fix that was the
*other* half of comp ⑤'s mismatch — orthogonal: WHEN keyframes fire vs WHAT the static
transform is).

## Cases
- 2026-06-22 first seen — booyah comp ⑤; user caught Scale/Rotation diff in the AE transform
  panel after a long (mis-aimed) nested-time bisection. Fixed by copying static Scale+Rotate Z.
  Verification workflow (per-frame orig-vs-clone PNG + solo-layer isolation) → `checklists/showcase.md`.
