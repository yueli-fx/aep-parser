# ⚠ NewCameraLayer / NewLightLayer — embed-whole-Layr create

SUMMARY: NewCameraLayer / NewLightLayer — embed-whole-Layr create
READ WHEN: implementing/extending source-less layer creation (NewCameraLayer / NewLightLayer); adding a new embed-whole-Layr layer type; debugging "AE drops/mis-types a Go-created camera/light"; deciding embed-whole-Layr vs from-scratch for a new layer kind

---

Shipped 2026-06-10. AE 2020 + AE 2025 ship-gate PASS (`TestNewCameraLight_AEShipGate_AE20{20,25}`):
AE accepts a Go-created Camera + Light in a fresh comp, types them correctly
(`Cam1(cam)`, `Light1(light)`), resave preserves, Go re-parse confirms
`LayerTypeCamera`/`LayerTypeLight` from AE's output. Promoted Alpha → Stable
as structural ops in the 2026-06-11 New\* family audit on this
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
- **Property-based option setters — RESOLVED 2026-06-14 via parse-the-clone (for
  the slots the template carries).** The deferral was framed as needing "property
  synthesis", but the easier half sufficed for the **non-elided** options: the
  embed-whole-Layr template carries those property leaves (tdmn→tdbs→cdat) — they
  just had no parsed scene tree. `newTemplatedLayer` now runs the **same**
  property-tree parse the read path uses (`parseProperties` +
  `buildAEPropertyGroupTree` + `wirePropertyTreeLeaves`, warnings to a LOCAL sink
  so they can't trip the splice rollback) on the cloned Layr, lighting up the
  setters/accessors for the slots present — they were always implemented
  (`setScalarProperty` → property backref), just unreachable without the tree.
  - **CORRECTION (2026-06-14, code-verified `tmp_debug/probe_fromscratch_opts`):
    parse-the-clone only reaches slots AE did NOT elide in the template.** The
    earlier "lights up the entire existing surface incl. Iris\*" was an
    overstatement. Coverage by slot presence:
    - **Light: 10/10 OK + all dual-version AE-DOM-gated** (2026-06-14 finish) —
      Intensity / Color (spliced) / Cone Angle / Cone Feather / Falloff Type /
      Falloff Start / Falloff Distance / Casts Shadows / Shadow Darkness / Shadow
      Diffusion all set on the from-scratch Spot light + read back from AE's DOM
      in `TestNewCameraLight_AEShipGate`. **Gotcha: Falloff Distance default is
      500** — gating it at 500 made AE elide it on resave (Go re-parse → nil); the
      gate uses 750. (Same default-elision trap as any "distinctive value happens
      to equal the AE default" — pick a value provably off-default.)
    - **Camera: was 5/13, now 13/13** — the 5 (Zoom / DoF / Focus / Aperture /
      Blur Level) have template slots; the 8 elided Iris\*/Highlight\* were
      resolved by synthesis-insert below.
  - **Camera Iris\*/Highlight\* — RESOLVED 2026-06-14 via synthesis-insert
    (AE2020+2025 PASS).** AE elides these 8 DoF-bokeh controls at default (even
    AE's own parsed cameras lack them), so the embed template had no slots and
    `SetIrisShape`/`…`/`SetIrisHighlightSaturation` failed from-scratch with
    `property not present`. Fixed with the **same leaf-splice as Light Color**,
    batched: author a camera with **DoF on** + all 8 set non-default
    (`tmp_debug/gen_camera_iris.jsx` → `test_data/re_camera_iris.aep`), extract
    all 8 `(tdmn, LIST:tdbs)` pairs into one `templates/camera_iris_leaves.bin`
    (a LIST(tdgp) of 8 pairs in canonical order; `extract_camera_iris_leaves`),
    and `newTemplatedLayer` (camera only) splices them **after Blur Level** +
    resets each cdat to AE's default (Shape 1 / Rotation 0 / Roundness 0 /
    AspectRatio 1 / DiffractionFringe 0 / HighlightGain 0 / HighlightThreshold 1
    / HighlightSaturation 0). Canonical Camera Options order (probed): Zoom[1],
    DoF[2], Focus[3], Aperture[4], BlurLevel[5], then the 8 iris[6..13].
    - **BUG also found + fixed**: the `MatchNameCameraIrisHighlightSaturation`
      constant was the **correct** spelling `"ADBE Iris Highlight Saturation"`,
      but Adobe's real on-disk match-name is the **misspelled** `"ADBE Iris
      Hightlight Saturation"` ("Hightlight"). The correct spelling never matched,
      so `IrisHighlightSaturation()` / `SetIrisHighlightSaturation` were silently
      always-nil **even on parsed cameras** — a latent bug surfaced only because
      this work probed AE's actual match-names. Constant corrected.
    - **Untouched plain camera (DoF off, iris leaves at default) is AE-accepted**
      (verified `tmp_debug/verify_plain_camera.jsx`, AE2025): file opens, camera
      intact, irisShape reads back default 1 — no corruption from the always-on
      splice.
    - Gated: `TestCameraLightOptions_FromScratch_Roundtrip` (Go set/read 6 iris
      values) + dual-version `TestNewCameraLight_AEShipGate_AE20{20,25}` extended
      to set all 8 + read them back from **AE's DOM** (irisShape=4 / rotation=25 /
      roundness=60 / aspectRatio=1.8 / diffractionFringe=30 / highlightGain=40 /
      highlightThreshold=0.7 / highlightSaturation=50) + Go re-parse confirms
      resave kept them. Both AE versions identical.
    - **Scope honesty**: AE-model-readback + resave-preservation (pixel-proving
      DoF bokeh needs a lit 3D scene we don't build yet).
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
  - **`ADBE Light Color` — RESOLVED 2026-06-14 via synthesis-insert (AE2020+2025
    PASS).** AE **elides** it at its default white, so it is absent from the
    embed-whole-Layr template → no slot, `LightColor()` nil on a fresh light.
    Fixed by the **leaf-splice** half (not whole-template re-extraction): author a
    light with a non-default colour in AE 2020 (`tmp_debug/gen_light_color.jsx` →
    `test_data/re_light_color.aep`), extract just the `(tdmn "ADBE Light Color",
    LIST:tdbs)` pair (`tmp_debug/extract_light_color_leaf` →
    `templates/light_color_leaf.bin`, a LIST(tdgp) wrapper), and have
    `newTemplatedLayer` (light only) splice it into the cloned Light Options group
    + reset the cdat to white **before** the property-tree parse — so the existing
    `SetLightColor`/`LightColor()` light up from scratch. Same synthesis-insert
    blueprint as `SetEffectParam` (no child-count header to bump; LIST sizes
    reflow on Write). The whole current light template stays byte-unchanged → zero
    regression on the create gate.
    - **Two traps cost two failed gate runs (both classic red-line-4 false-greens):**
      1. **Colour cdat is `[A,R,G,B] × 255` f64, NOT raw `[r,g,b,a] × 1.0`** —
         already RE'd for shape fill/stroke (`encodeShapeColorBE`). A fresh light's
         cdat dump: DOM `[0.9,0.1,0.2]` → disk `[255, 229.5, 25.5, 51]` (alpha
         first, all ×255). `SetLightColor`'s established contract takes these **raw
         0..255 alpha-first** values (`[]float64{200,100,50,255}` in the parsed-
         fixture test); white = `[255,255,255,255]`. A Go round-trip with `[0.2,
         0.4,0.8,1]` was self-consistently GREEN yet AE rendered white — the writer
         and parser agreed on the wrong scale/order.
      2. **Property order within a group is significant.** AE's canonical Light
         Options order is **Intensity[1], Color[2]**, Cone Angle[3], … (probed:
         `tmp_debug/gen_light_order.jsx`). Splicing Color at the group **front**
         (ahead of Intensity) made AE **silently drop it on open** — DOM read white,
         resave re-elided it (`LightColor()` nil on re-parse). Inserting the pair
         **right after Intensity's payload** fixed it. Insert leaves in canonical
         order, never blindly at the front.
    - Gated: `TestCameraLightOptions_FromScratch_Roundtrip` (Go: fresh light →
      `SetLightColor([255,51,102,204])` → write → re-parse → assert raw value) +
      the dual-version `TestNewCameraLight_AEShipGate_AE20{20,25}` extended to set
      the colour and read it back from **AE's DOM** (`lightOption.color` =
      `[0.2,0.4,0.8]`, i.e. 51/102/204 ÷255) + Go re-parse confirms resave kept the
      raw `[255,51,102,204]`. Both AE versions identical.
    - **Scope honesty**: AE-model-readback + resave-preservation (same ceiling as
      the other light options — pixel-proving a light's colour needs a lit 3D
      scene we don't build yet).
- ~~NewTextLayer~~ — SHIPPED 2026-06-11 via the same embed-whole-Layr path
  (mutate_layer_text.go; the btdk blob travels verbatim, so its complexity never
  materialized for creation — length-variable text WRITE is the remaining wall,
  see [[text-btdk-length-variable-write-scoping]]). ~~Solid/Null/Adjustment~~ —
  shipped 2026-06-10 ([[new-layer-types-scoping]]). All layer kinds now creatable.
