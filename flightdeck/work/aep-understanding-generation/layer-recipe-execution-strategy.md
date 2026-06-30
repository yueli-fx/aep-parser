# Layer Recipe/Profile Execution Strategy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move layer-level recipe/profile work from scattered field examples to
an object-level execution strategy with reusable `expected_profile.layers[]`
contracts, fixed field groups, and whole-recipe verification.

**Architecture:** Treat `Layer` / `profile.Layer` as the work unit. Existing
writer fields still land in narrow TDD slices, but each slice should attach to
the layer object contract when the profile already exposes the value. Use
`properties[]`, `effects[]`, `text_styles[]`, and `keyframes[]` only for nested
property-like surfaces; use `layers[]` for layer identity, metadata, timing,
flags, and refs.

**Tech Stack:** Go, `internal/aep`, `internal/scene`, `internal/profile`,
`internal/recipe`, `cmd/aeprecipe`, `scripts/verify_recipe_profiles.ps1`, AE
2025 render/profile gates when needed.

---

## Operating Rules

- Do not ask for per-field direction once this strategy is active.
- Keep implementation commits small, but choose the next field from this
  matrix.
- Prefer one readable object-level recipe over many isolated layer recipes when
  fields share evidence level and failure mode.
- Keep dedicated recipes for layer families with distinct semantics: parent
  refs, matte refs, camera/light settings, shape/text/effect/property content,
  keyframes, and expressions.
- Use `scripts/verify_recipe_profiles.ps1` after each slice to check all
  recipe examples validate, compile, and pass `profile_checks`.
- Add knowledge notes only when the field has non-obvious byte layout, AE DOM
  readback caveats, or false-green risk.

## Evidence Ladder For Layer Fields

- `L3 AE accept`: AE opens/re-saves or renders the compiled project and the
  setting is reflected by AE-visible behavior or DOM readback.
- `L2 roundtrip/profile`: compiled AEP reopens through this parser and
  profile/raw bytes match, but AE DOM/render does not provide reliable
  confirmation.
- `L1 parsed`: parser/profile exposes the value, but recipe writer is not in
  scope yet.
- `Blocked`: existing parser/writer cannot provide a defensible contract.

## Layer Field Matrix

| Group | Recipe path | Current write path | Profile path target | Expected-profile path | Evidence | Status |
| --- | --- | --- | --- | --- | --- | --- |
| Identity | `layers[].name` / `type` | layer constructors | `layers[].name` / `type` | `expected_profile.layers[].name` / `type` | L3 | done |
| Metadata | `label` / `comment` | `Layer.SetLabel` / `SetComment` | `layers[].label` / `comment` | `expected_profile.layers[].label` / `comment` | L3 | done |
| Timing | `start_time` / `in_point` / `out_point` / parsed `duration` / `stretch` | `Layer.SetStartTime` / `SetInPoint` / `SetOutPoint`; duration/stretch parsed from resulting layer span | `layers[].timing.*` | `expected_profile.layers[].timing.*` | L3/L2 | done |
| Basic flags | `visible` / `solo` / `locked` / `shy` / `motion_blur` / `markers_locked` | `Layer.Set*` | `layers[].flags.*` | `expected_profile.layers[].flags.*` | L3 | done |
| AV flags | `effects_enabled` / `audio_enabled` / `frame_blend_enabled` / `frame_blend_pixel_motion` / `collapse_transform` / `sampling_bicubic` / `preserve_transparency` | `Layer.Set*` | `layers[].flags.*` | `expected_profile.layers[].flags.*` | L3 | done |
| Type flags | `is_3d` / `is_adjust` / `is_null` / `is_guide` | `Layer.Set*` plus type constructors | `layers[].flags.*` | `expected_profile.layers[].flags.*` | L3 | done |
| Quality/blend | `quality` / `blending_mode` | `Layer.SetQuality` / `SetBlendingMode` | `layers[].quality` / `blending_mode` | `expected_profile.layers[].quality` / `blending_mode` | L3 | done |
| Auto-orient | `auto_orient` | `Layer.SetAutoOrient` | `layers[].auto_orient` | `expected_profile.layers[].auto_orient` | L3 | done |
| Source refs | solid/precomp source layer creation | layer constructors / `Layer.SetSource` | `layers[].source_ref` | `expected_profile.layers[].source` / `source_kind` | L3 | done for solid footage source |
| Masks | `layers[].masks[]` static path/name/mode/inverted/options/path keyframes | `AddMask` after Reopen plus `Mask.Set*` option setters and `SetMaskPathKeyframes` | `layers[].masks[]` | `expected_profile.masks[]` | L3 | static + options + path keyframes done |
| Parent refs | `parent` | `Layer.SetParent` | `layers[].parent_ref` | `expected_profile.layers[].parent` | L3 | done |
| Classic matte refs | `track_matte` | `Layer.SetTrackMatte` | `layers[].flags.track_matte_name` / `matte_ref` | `expected_profile.layers[].track_matte` / `matte` | L3 | done |
| Explicit matte refs | `matte` + non-none `track_matte` with `project.target_version: "AE2025"` | `Layer.SetTrackMatteSource` | `layers[].matte_ref` | `expected_profile.layers[].matte` | L3 AE2025-only | done |
| Transform statics | `transform.position` / `scale` / `anchor_point` / `rotation` / `opacity` | `SetLayerTransform` | `properties[]` | `expected_profile.properties[]` | L3 | keep separate |
| Transform keyframes/ease | `transform.*_keyframes` | `SetLayerTransform` | `properties[].keyframes[]` | `expected_profile.keyframes[]` | L3 | keep separate |
| Transform expressions | `transform.expressions.*` | `Property.SetExpression` / `Property.SetExpressionEnabled` | `properties[].expression` / `expression_enabled` | `expected_profile.properties[].expression` / `expression_enabled` | L3 | keep separate |
| Text style | `text_style.*` | text run/paragraph setters | `layers[].text.*` | `expected_profile.text_styles[]` | L3 | keep separate |
| Text animators | `text_animators[].property: opacity/position/scale/rotation/color/tracking/character_offset/fill_opacity/stroke_opacity/stroke_width/skew/rotation_x/rotation_y/stroke_color` static Range Selector + `range_offset_keyframes` | `AddTextOpacityAnimator` / `AddTextPositionAnimator` / `AddTextScaleAnimator` / `AddTextRotationAnimator` / `AddTextColorAnimator` / `AddTextTrackingAnimator` / `AddTextCharacterOffsetAnimator` / `AddTextFillOpacityAnimator` / `AddTextStrokeOpacityAnimator` / `AddTextStrokeWidthAnimator` / `AddTextSkewAnimator` / `AddTextRotationXAnimator` / `AddTextRotationYAnimator` / `AddTextStrokeColorAnimator` / `AnimateTextRangeOffset` | `properties[]` / `properties[].keyframes[]` | `expected_profile.properties[]` / `expected_profile.keyframes[]` | L3/L4 | static animator properties + range offset done |
| Shape contents | `shape.*` | vector group writers | `layers[].shapes[]` / `properties[]` | `expected_profile.properties[]` | L3/L4 | keep separate |
| Effects | `effects[]` | `AddEffect` / `SetEffectParam` | `layers[].effects[]` | `expected_profile.effects[]` | L3/L4 | keep separate |
| Camera/light layer identity | `type: camera` / `type: light` | layer constructors | `layers[].name` / `type` | `expected_profile.layers[].name` / `type` | L3 | done |
| Camera options | `camera.*` | camera setters | `properties[]` | `expected_profile.properties[]` | L3/L4 | done in consolidated camera baseline |
| Light options | `light.kind` / `light.source_layer` / `light.*` | light setters | `layers[].light_kind` / `layers[].light_source_ref` / `properties[]` | `expected_profile.layers[].light_kind` / `light_source` / `expected_profile.properties[]` | L3/L4 | done |

## Preferred Execution Order

1. Lock the layer object profile contract for identity, metadata, timing, and
   flags in one readable baseline recipe.
2. Add missing profile-visible layer state for quality, blend, and auto-orient.
3. Add parent-ref object checks using a two-layer baseline.
4. Add classic track-matte mode checks with a two-layer baseline.
5. Keep transform, text, shape, effect, camera, and light checks in their
   existing nested expected-profile families.

## Standard Slice Template

### Task N: Add One Layer Field Group

**Files:**
- Modify: `internal/profile/profile.go`
- Modify: `internal/recipe/schema.go`
- Modify: `internal/recipe/compiler.go`
- Modify: `internal/recipe/schema_test.go` when validation changes
- Modify: `internal/recipe/compiler_test.go`
- Modify or create: `examples/recipes/minimal-layer-*.json`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/layer-recipe-execution-strategy.md`

- [ ] **Step 1: Write the failing layer-object profile test**

Use `internal/recipe/compiler_test.go` and assert stable paths under
`expected_profile.layers[]`, for example:

```go
assertProfileCheck(t, report, "expected_profile.layers[0].flags.visible", true)
```

- [ ] **Step 2: Run the focused failing test**

```powershell
go test ./internal/recipe -run TestCompileToFileChecksLayerObjectProfileExample -count=1
```

- [ ] **Step 3: Implement the minimal profile/schema/compiler contract**

Add only fields already backed by writer/parser evidence. If `profile.Layer`
does not expose the value, add it there first and populate it from parsed
`JSONLayer` / scene data.

- [ ] **Step 4: Add or update a readable object-level recipe**

Prefer `examples/recipes/minimal-layer-object-profile.json` for same-evidence
identity/metadata/timing/flag fields. Keep separate recipes for parent/matte or
type-specific content.

- [ ] **Step 5: Run focused verification**

```powershell
go test ./internal/recipe -run TestCompileToFileChecksLayerObjectProfileExample -count=1
pwsh -NoProfile -File scripts\verify_recipe_profiles.ps1 -Recipe examples\recipes\minimal-layer-object-profile.json
```

- [ ] **Step 6: Run branch verification**

```powershell
go test ./internal/recipe ./internal/profile
pwsh -NoProfile -File scripts\verify_recipe_profiles.ps1
git diff --check
```

Run `go test ./...` and `go vet ./...` when shared parser/writer behavior
changes beyond expected-profile plumbing.

- [ ] **Step 7: Update flightdeck and commit**

Commit each completed field group locally. Do not push.

## Completed Baseline

`examples/recipes/minimal-layer-object-profile.json` now asserts one text layer
as an object-level contract:

- identity: `name`, `type`
- display/mode: `quality`, `blending_mode`, `auto_orient`
- metadata: `label`, `comment`
- timing: `start_time`, `in_point`, `out_point`, parsed span `duration`, and
  `stretch`
- flags: `visible`, `solo`, `locked`, `shy`, `motion_blur`,
  `markers_locked`, `effects_enabled`, `audio_enabled`, `frame_blend_enabled`,
  `frame_blend_pixel_motion`, `collapse_transform`, `sampling_bicubic`,
  `is_3d`, `is_adjustment`, `is_guide`, and `preserve_transparency`

## Next Concrete Slice

Add parent-ref object checks using `expected_profile.layers[].parent`, because
recipe writer support and `profile.Layer.parent_ref` already exist. Completed
in `minimal-layer-parent.json`.

Classic `track_matte` mode is now recipe-owned and covered by
`minimal-layer-track-matte.json`: the fill layer asserts both
`expected_profile.layers[].track_matte` and the inferred classic matte source
through `expected_profile.layers[].matte`.

Layer render/sampling flags are now part of the consolidated object baseline:
`minimal-layer-object-profile.json` asserts both
`expected_profile.layers[].flags.frame_blend_pixel_motion` and
`expected_profile.layers[].flags.sampling_bicubic`.

Camera and light layer creation now has object identity coverage in
`minimal-camera-layer.json` and `minimal-light-layer.json`; parameter values
remain in the property-based camera/light recipe family.

Camera option values now also have a consolidated baseline:
`minimal-camera-object-profile.json` sets and asserts zoom, depth of field,
focus distance, aperture, blur level, and the iris controls in one recipe.

Light option values now also have a consolidated baseline:
`minimal-light-object-profile.json` sets and asserts light kind, intensity,
color, shadow, falloff, and cone controls in one recipe. Light source refs are
covered by `minimal-light-source.json` through
`expected_profile.layers[].light_source`.

Generic layer source refs are now covered by `minimal-layer-source-ref.json`,
which asserts a generated solid layer's footage source through
`expected_profile.layers[].source` and `source_kind`.

Null and adjustment base recipes now assert their object identity and type
flags through `expected_profile.layers[]`: `minimal-null-layer.json` covers the
null controller plus child parent ref, and `minimal-adjustment-layer.json`
covers the adjustment layer flag alongside its effect contract.

The older standalone layer recipes for label, comment, timing, common switches,
motion blur, shy, advanced switches, quality/blending, auto-orient, and null
flag now also assert their `expected_profile.layers[]` object contracts instead
of remaining count-only smoke tests.

`minimal-light-kind.json` now asserts the light layer identity and
`expected_profile.layers[].light_kind`, clearing the last count-only recipe
profile contract.

Camera and light single-option recipes now also assert camera/light layer
identity through `expected_profile.layers[]` alongside their property value
checks, so those examples are anchored to the object contract instead of only
the property list.

Shape, text/effect, transform, and comp-draft examples that author layers now
also assert `expected_profile.layers[]`. The recipe test suite includes a guard
that fails any authored-layer example missing an object-level layer contract or
asserting fewer layer entries than it authored.

Transform keyframe examples now also require `expected_profile.keyframes[]`.
The guard counts authored transform keyframe streams and fails examples that
do not assert at least that many keyframed profile entries; this caught and
fixed `minimal-layer-auto-orient.json`, whose Position keyframes are required
for the `along_path` behavior.

Transform keyframes now have a complete minimal object-family baseline:
`minimal-transform-keyframes.json` authors Position, Anchor Point, Scale,
Rotation, and Opacity keyframe streams on one text layer and asserts every
stream through `expected_profile.keyframes[]`. The existing
`minimal-transform-keyframe-ease.json` remains the dedicated temporal-ease
slice for Position and Opacity.

Expression examples now assert enabled state as part of the property contract.
`minimal-transform-expression.json` checks both default-enabled Position and
disabled Opacity expressions through `expected_profile.properties[]`, while
`minimal-effect-param-expression.json` checks disabled Gaussian Blur parameter
expression state through `expected_profile.effects[].params[]`.

Text style now has a dedicated minimal recipe/profile baseline:
`minimal-text-style.json` authors font size, fill color, tracking, faux bold,
faux italic, stroke enable/color/width, and paragraph justification on one text
layer, and asserts them through `expected_profile.text_styles[]`. The larger
`minimal-text-shape.json` still exercises text style in a mixed content scene,
but it is no longer the only profile contract for text style.

Text animators now have a first minimal recipe/profile baseline:
`minimal-text-animator-opacity.json` authors a static opacity Range Selector
through `layers[].text_animators[]` and asserts the resulting
`ADBE Text Opacity` property through `expected_profile.properties[]`.
`minimal-text-animator-position.json` authors a static 3D position text animator
and asserts `ADBE Text Position 3D` through `expected_profile.properties[]`.
`minimal-text-animator-scale.json` authors a static 3D scale text animator and
asserts `ADBE Text Scale 3D` through `expected_profile.properties[]`.
`minimal-text-animator-rotation.json` authors a static rotation text animator
and asserts `ADBE Text Rotation` through `expected_profile.properties[]`.
`minimal-text-animator-color.json` authors a static fill-color text animator
and asserts `ADBE Text Fill Color` through `expected_profile.properties[]`.
`minimal-text-animator-tracking.json` authors a static tracking text animator
and asserts `ADBE Text Tracking Amount` through `expected_profile.properties[]`.
`minimal-text-animator-character-offset.json` authors a static character offset
text animator and asserts `ADBE Text Character Offset` through
`expected_profile.properties[]`.
`minimal-text-animator-fill-opacity.json` authors a static fill-opacity text
animator and asserts `ADBE Text Fill Opacity` through
`expected_profile.properties[]`.
`minimal-text-animator-stroke-opacity.json` authors a static stroke-opacity text
animator and asserts `ADBE Text Stroke Opacity` through
`expected_profile.properties[]`.
`minimal-text-animator-stroke-width.json` authors a static stroke-width text
animator and asserts `ADBE Text Stroke Width` through
`expected_profile.properties[]`.
`minimal-text-animator-skew.json` authors a static skew text animator and
asserts `ADBE Text Skew` through `expected_profile.properties[]`.
`minimal-text-animator-rotation-x.json` and
`minimal-text-animator-rotation-y.json` author static 3D rotation text
animators and assert `ADBE Text Rotation X` / `ADBE Text Rotation Y` through
`expected_profile.properties[]`.
`minimal-text-animator-stroke-color.json` authors a static stroke-color text
animator and asserts `ADBE Text Stroke Color` through
`expected_profile.properties[]`.
`minimal-text-animator-range-offset.json` adds animated Range Selector Offset
through `range_offset_keyframes` and asserts `ADBE Text Percent Offset` through
`expected_profile.keyframes[]`. Later text animator slices should extend this
family by property kind only when the expected-profile property/keyframe
contract stays explicit.

Nested content examples also require the matching expected-profile family:
shape/camera/light property content must assert `properties[]`, text style
content must assert `text_styles[]`, text animator content must assert
`properties[]`, and effect content must assert `effects[]`. A solid layer's
`shape.fill_color` remains source-generation metadata and is not treated as
shape-layer property content.

Shape filter coverage has been split into dedicated minimal recipes for the
stable filter families that were previously exercised mainly through the larger
`minimal-text-shape.json` baseline: `minimal-shape-trim.json`,
`minimal-shape-round-corners.json`, `minimal-shape-offset-paths.json`,
`minimal-shape-zigzag.json`, `minimal-shape-pucker-bloat.json`, and
`minimal-shape-twist.json`. `TestCompileToFileChecksShapeFilterProfileExamples`
keeps these examples anchored to their `expected_profile.properties[]` checks.

Static vector masks now have a first recipe/profile slice:
`minimal-layer-mask.json` authors `layers[].masks[]`, materializes it through
the existing Reopen + `AddMask` path, applies mode/inverted through stable mask
setters, and asserts the parsed mask object with `expected_profile.masks[]`.
The same baseline now also covers byte-level mask options: locked state,
timeline color, motion-blur override, and feather-falloff mode. Mask
opacity, feather, and expansion are now covered in the same recipe/profile
family through stable mask setters. Mask path keyframes are also covered in the
same baseline through `path_keyframes[]` authoring and
`expected_profile.masks[].path_keyframes[]` count/time/vertex-count checks.
Per-keyframe interpolation/ease and tangent detail remain future work only when
the parser/profile can expose a defensible contract.

Explicit AE2025 source mattes now have a target-version-gated recipe slice:
`minimal-layer-explicit-matte.json` sets `project.target_version: "AE2025"`,
uses `layers[].matte` to name the source layer, and keeps the channel mode in
`layers[].track_matte`. The recipe asserts both
`expected_profile.layers[].track_matte` and `expected_profile.layers[].matte`.
Continue with the next recipe-owned family; do not add more AE2025-only fields
without the same explicit target-version boundary.

## Self-Review

- Spec coverage: object-level sequencing, field matrix, evidence levels,
  recipe strategy, verification, and commit discipline are covered.
- Placeholder scan: no `TBD`, `TODO`, or unspecified "add tests" steps remain.
- Type consistency: paths use existing `recipe.Layer`, `profile.Layer`,
  `Layer.Set*`, and `cmd/aeprecipe` names.
