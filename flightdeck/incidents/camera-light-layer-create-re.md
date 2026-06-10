---
status: active
when_to_read: implementing/extending source-less layer creation (NewCameraLayer / NewLightLayer); adding a new embed-whole-Layr layer type; debugging "AE drops/mis-types a Go-created camera/light"; deciding embed-whole-Layr vs from-scratch for a new layer kind
applies_to: [new-camera-layer, new-light-layer, source-less-layer, embed-whole-layr, ldta-subtype, camera-options, light-options, structural-write, layer-create, ship-gate, ae2020, ae2025]
last_updated: 2026-06-11
---

# NewCameraLayer / NewLightLayer — embed-whole-Layr create

Shipped 2026-06-10. AE 2020 + AE 2025 ship-gate PASS (`TestNewCameraLight_AEShipGate_AE20{20,25}`):
AE accepts a Go-created Camera + Light in a fresh comp, types them correctly
(`Cam1(cam)`, `Light1(light)`), resave preserves, Go re-parse confirms
`LayerTypeCamera`/`LayerTypeLight` from AE's output. Promoted Alpha → Stable
(structural op, CLAUDE.md #2) in the 2026-06-11 New\* family audit on this
gate evidence.

## Why these were easy (vs shape layers / solids)

- **Source-less**: a Camera/Light is defined entirely by its ldta + a
  layer-specific group (Camera Options / Light Options) — no backing footage Item
  (unlike Solid/Null/Adjustment, see [[new-layer-types-scoping]]).
- **Same 4-child Layr shape as a shape layer**: `ldta, Utf8(name),
  LIST(tdgp){Transform Group + Camera/Light Options + Group End}, LIST(Gide)`.
- **Embed-WHOLE-Layr, not from-scratch**: we clone an AE-native Camera/Light Layr
  extracted verbatim from `re_cameralight.aep` (`templates/layer_{camera,light}_body.bin`),
  so every AE-internal flag byte (@0x25, @0x27, attr/quality, Camera/Light fields)
  is faithful. This sidesteps the from-scratch silent-drop pitfalls shape layers
  needed extensive RE for ([[multi-layer-silent-drop]], [[ae2020-shape-ldta-164-corrupt]]).
  Same principle as cross-Project `InsertLayer` (clone a Layr, remap per-instance
  fields).

## Per-instance patches (only these)

`newTemplatedLayer` (mutate_layer_camera.go) clones the template and patches:
- **ldta @0x00** layer ID (`maxLayerIDInItemList+1`, per-comp namespace).
- **ldta time span** start/in/out → 0 / 0 / comp-duration ticks, divisor=TickRate
  (the template carries its source comp's ticks; an out-point past the new comp's
  duration risks an AE clamp/reject).
- **ldta @0x84** ParentID → 0 (template may carry a stale parent).
- **ldta size** pad/trim to capability `LdtaSize` (160 AE2020/22, 164 AE2025) —
  camera/light fields all fit in ≤0xA0, so size only varies the trailing zero-pad.
- **Utf8 name** child → caller's name (length-variable).
Then splices `[Layr, Ewst, lowerLayerSiblings()…]` at `insertLayrPosition` with the
same atomic snapshot + warnings-rollback as NewShapeLayer, + nextItemID bump + cdta
@0x18 "comp has user content" bump.

## Version portability

Template extracted from an **AE-2020/22-form** fixture (160B ldta) is accepted by
**both** AE 2020 and AE 2025 (2025 pads to 164B). Mirrors the gradient/effect
version-portability finding.

## Gotchas

- **Test cache hid a JSX edit.** Go's test cache reuses a cached PASS+output when
  Go inputs are unchanged; a `.jsx` edit is invisible to it, so a stale
  instanceof-label was displayed. Use `-count=1` when iterating on a ship-gate's
  JSX. (Did not affect correctness — the Go re-parse preservation check is the
  authoritative type signal.)

## Deferred

- **Setters on fresh camera/light** (Zoom / Intensity / LightKind / color …): the
  existing Camera*/Light* setters operate on a parsed layer; on a freshly-built
  layer there's no scene tree, so values inherit the template's AE defaults until
  write+reopen. Same scene-vs-chunk split as [[add-effect-splice-re]] Phase 2.
- ~~NewTextLayer~~ — SHIPPED 2026-06-11 via the same embed-whole-Layr path
  (mutate_layer_text.go; the btdk blob travels verbatim, so its complexity never
  materialized for creation — length-variable text WRITE is the remaining wall,
  see [[text-btdk-length-variable-write-scoping]]). ~~Solid/Null/Adjustment~~ —
  shipped 2026-06-10 ([[new-layer-types-scoping]]). All layer kinds now creatable.
