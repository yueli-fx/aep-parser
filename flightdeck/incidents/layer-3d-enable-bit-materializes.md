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

## Gate (DOM-readback, not pixel — and why that's correct here)

`TestLayer3DEnable_AEShipGate_AE2020/2025` PASS. **Deliberately DOM-level**: a 3D
layer with a DEFAULT transform (z=0, no rotation) renders **pixel-identical** to
its 2D self, so "enable" has no visible surface to assert on — the capability's
real surface is the 3D-ness AE reports (threeDLayer + materialized channels +
Position→3D). The **visible** 3D transform under a camera (dolly / parallax /
RotateY perspective) is the next capability and gets its own render-pixel gate
(red line 4). Per delivery-contract this split is honest: enable is shipped at its
surface; visible-3D is NOT yet claimed.

## Pre-existing 3D infra (already shipped, fixture-based)

Material / Geometry options + transform setters (SetRotateX/Y, SetOrientation,
3-comp SetPosition) already work **on fixture 3D layers** (`re_material_options.aep`
/ `re_cameralight.aep`, `layer_3d_test.go`) — they require the channel present in
the tree. The gap this incident closes is making OUR from-scratch layer 3D in the
first place; wiring those setters onto a from-scratch-3D layer is the next step.
