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
| Timing | `start_time` / `in_point` / `out_point` | `Layer.SetStartTime` / `SetInPoint` / `SetOutPoint` | `layers[].timing.*` | `expected_profile.layers[].timing.*` | L3 | done |
| Basic flags | `visible` / `solo` / `locked` / `shy` / `motion_blur` | `Layer.Set*` | `layers[].flags.*` | `expected_profile.layers[].flags.*` | L3 | done |
| AV flags | `effects_enabled` / `audio_enabled` / `frame_blend_enabled` / `frame_blend_pixel_motion` / `collapse_transform` / `sampling_bicubic` / `preserve_transparency` | `Layer.Set*` | `layers[].flags.*` | `expected_profile.layers[].flags.*` | L3 | done |
| Type flags | `is_3d` / `is_adjust` / `is_null` / `is_guide` | `Layer.Set*` plus type constructors | `layers[].flags.*` | `expected_profile.layers[].flags.*` | L3 | done |
| Quality/blend | `quality` / `blending_mode` | `Layer.SetQuality` / `SetBlendingMode` | `layers[].quality` / `blending_mode` | `expected_profile.layers[].quality` / `blending_mode` | L3 | done |
| Auto-orient | `auto_orient` | `Layer.SetAutoOrient` | `layers[].auto_orient` | `expected_profile.layers[].auto_orient` | L3 | done |
| Parent refs | `parent` | `Layer.SetParent` | `layers[].parent_ref` | `expected_profile.layers[].parent` | L3 | done |
| Classic matte refs | `track_matte` | `Layer.SetTrackMatte` | `layers[].flags.track_matte_name` / `matte_ref` | `expected_profile.layers[].track_matte` / `matte` | L3 | done |
| Explicit matte refs | not recipe-owned yet | `Layer.SetTrackMatteSource` / `SetTrackMatteLayer` | `layers[].matte_ref` | planned when writer enters recipe scope | L3 AE2025-only | blocked |
| Transform statics | `transform.position` / `scale` / `anchor_point` / `rotation` / `opacity` | `SetLayerTransform` | `properties[]` | `expected_profile.properties[]` | L3 | keep separate |
| Transform keyframes/ease | `transform.*_keyframes` | `SetLayerTransform` | `properties[].keyframes[]` | `expected_profile.keyframes[]` | L3 | keep separate |
| Transform expressions | `transform.expressions.*` | `Property.SetExpression` | `properties[].expression` | `expected_profile.properties[].expression` | L3 | keep separate |
| Text style | `text_style.*` | text run/paragraph setters | `layers[].text.*` | `expected_profile.text_styles[]` | L3 | keep separate |
| Shape contents | `shape.*` | vector group writers | `layers[].shapes[]` / `properties[]` | `expected_profile.properties[]` | L3/L4 | keep separate |
| Effects | `effects[]` | `AddEffect` / `SetEffectParam` | `layers[].effects[]` | `expected_profile.effects[]` | L3/L4 | keep separate |
| Camera/light layer identity | `type: camera` / `type: light` | layer constructors | `layers[].name` / `type` | `expected_profile.layers[].name` / `type` | L3 | done |
| Camera options | `camera.*` | camera setters | `properties[]` | `expected_profile.properties[]` | L3/L4 | done in consolidated camera baseline |
| Light options | `light.kind` / `light.*` | light setters | `layers[].light_kind` / `properties[]` | `expected_profile.layers[].light_kind` / `expected_profile.properties[]` | L3/L4 | done in consolidated light baseline for profile-visible properties |

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
- timing: `start_time`, `in_point`, `out_point`
- flags: `visible`, `solo`, `locked`, `shy`, `motion_blur`,
  `effects_enabled`, `audio_enabled`, `frame_blend_enabled`,
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
color, shadow, falloff, and cone controls in one recipe. `light.source_layer`
remains writer/capability-covered but does not yet have a stable profile field.

Next, leave explicit AE2025 `SetTrackMatteSource` / `SetTrackMatteLayer`
blocked until recipe target-version handling is explicit. Continue with the
next recipe-owned family instead of mixing AE2025-only layout requirements into
the AE2020-targeted recipe compiler.

## Self-Review

- Spec coverage: object-level sequencing, field matrix, evidence levels,
  recipe strategy, verification, and commit discipline are covered.
- Placeholder scan: no `TBD`, `TODO`, or unspecified "add tests" steps remain.
- Type consistency: paths use existing `recipe.Layer`, `profile.Layer`,
  `Layer.Set*`, and `cmd/aeprecipe` names.
