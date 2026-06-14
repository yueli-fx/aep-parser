---
status: active
when_to_read: implementing/extending source-less layer creation (NewCameraLayer / NewLightLayer); adding a new embed-whole-Layr layer type; debugging "AE drops/mis-types a Go-created camera/light"; deciding embed-whole-Layr vs from-scratch for a new layer kind
applies_to: [new-camera-layer, new-light-layer, source-less-layer, embed-whole-layr, ldta-subtype, camera-options, light-options, parse-the-clone, property-tree, option-setter, structural-write, layer-create, ship-gate, ae2020, ae2025]
last_updated: 2026-06-14
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

## Fresh-layer setters: ldta-based work, property-based deferred (2026-06-14)

The "setters don't work on fresh camera/light" deferral was **too broad**. Split by
where the field lives:

- **`SetLightKind` (ldta @0x88) — WORKS from-scratch, AE-gated dual-version.**
  `newTemplatedLayer` wires the fresh `*Layer`'s `layerBackrefs.ldta` to the cloned
  template ldta, so any ldta-byte setter reaches it. The light template default is
  **Parallel** (not Ambient — the old "默认只能环境光" note was wrong). Patching
  ldta @0x88 to any of the 4 kinds is accepted by AE 2020 **and** 2025: a fresh
  light set to Spot/Point/Ambient/Parallel reports the matching `lightType` and
  resave preserves it — AE tolerates the kind/Light-Options mismatch (it synthesizes
  the missing per-kind props like Cone Angle at runtime). Gated by
  `TestNewCameraLight_AEShipGate_AE20{20,25}` (Light1 built Spot, JSX asserts
  `lightType===SPOT`, Go resave asserts `LightKind==spot`).
- **Property-based option setters — RESOLVED 2026-06-14 via parse-the-clone.**
  The deferral was framed as needing "property synthesis", but the easier half
  sufficed: the embed-whole-Layr template **already carries** the full Camera/
  Light Options group with every property leaf (tdmn→tdbs→cdat) — it just had no
  parsed scene tree. `newTemplatedLayer` now runs the **same** property-tree parse
  the read path uses (`parseProperties` + `buildAEPropertyGroupTree` +
  `wirePropertyTreeLeaves`, warnings to a LOCAL sink so they can't trip the splice
  rollback) on the cloned Layr. This lights up the **entire existing** option
  setter/accessor surface (`SetCameraZoom`/`Focus`/`Aperture`/`Blur`/`DoF`/Iris*;
  `SetLightIntensity`/`ConeAngle`/`ConeFeather`/`Falloff*`/`Shadow*`) on a fresh
  layer — they were always implemented (`setScalarProperty` → property backref),
  just unreachable without the tree.
  - **Purely read-only on the bytes**: an untouched fresh camera/light still
    serializes byte-identically (the create ship-gate stays green); a setter then
    overwrites only its own cdat in place (length-preserving).
  - Gated: `TestCameraLightOptions_FromScratch_Roundtrip` (pure Go: create →
    set → write → re-parse → assert getters) + the dual-version
    `TestNewCameraLight_AEShipGate_AE20{20,25}` extended to set 7 options and read
    them back from **AE's DOM** (`cameraOption.zoom`=850 / `focusDistance`=1200 /
    `aperture`=180 / `depthOfField`=1; `lightOption.intensity`=65 / `coneAngle`=72
    / `coneFeather`=35) + Go re-parse confirms resave preserved them. Both AE
    versions identical.
  - **Scope honesty**: this is an **AE-model-readback + resave-preservation** gate
    (AE ingests the values into its DOM and keeps them), NOT a 3D-render pixel
    gate — pixel-proving a light's intensity/cone or a camera's DoF needs a lit
    **3D** scene, which needs 3D-layer creation we don't have yet. For these
    numeric option scalars (AE reads them into its model, no "enabled-bit" render
    trap) the DOM-readback ceiling is the appropriate gate.
  - **Still deferred — `ADBE Light Color`**: AE **elides** it at its default, so it
    is absent from the template → no slot to overwrite and `LightColor()` returns
    nil on a fresh light. This one genuinely needs the gradient-style fix (author
    a non-default color in AE + re-extract the template, or true synthesis-insert
    the leaf). Low priority.
- ~~NewTextLayer~~ — SHIPPED 2026-06-11 via the same embed-whole-Layr path
  (mutate_layer_text.go; the btdk blob travels verbatim, so its complexity never
  materialized for creation — length-variable text WRITE is the remaining wall,
  see [[text-btdk-length-variable-write-scoping]]). ~~Solid/Null/Adjustment~~ —
  shipped 2026-06-10 ([[new-layer-types-scoping]]). All layer kinds now creatable.
