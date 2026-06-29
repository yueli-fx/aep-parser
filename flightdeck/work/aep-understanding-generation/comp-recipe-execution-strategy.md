# Comp Recipe/Profile Execution Strategy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move comp-level recipe work from field-by-field ad hoc slices to an object-level execution strategy with a fixed field order, stable evidence requirements, and small commits.

**Architecture:** Treat `CompSpec` as the work unit and advance through a predeclared field matrix. Each field still lands through the same narrow TDD loop: profile contract when needed, schema/capability validation, compiler write, `expected_profile` assertion, example recipe, and verification. Fields that share one byte family or one API group can be batched only when they have the same evidence level and failure mode.

**Tech Stack:** Go, `internal/aep`, `internal/scene`, `internal/serializer`, `internal/profile`, `internal/recipe`, `cmd/aeprecipe`, AE 2025 render/profile gates.

---

## Operating Rules

- Do not ask the user for field-by-field direction once this strategy is active.
- Keep implementation commits small, but choose the next field from this document.
- Prefer object-level fixtures and tests that name `CompSpec` behavior over isolated one-off tests.
- Keep special shipgate tests only for evidence-sensitive fields: AE DOM readback mismatch, raw byte flags, renderer chunk replacement, or cross-version behavior.
- Add knowledge notes only when the field has a non-obvious byte layout, false-green risk, or AE acceptance caveat.
- Update this document when a field moves from planned to done.

## Evidence Ladder For Comp Fields

- `L3 AE accept`: AE opens/re-saves or renders the compiled project and the setting is reflected by AE-visible behavior or DOM readback.
- `L2 roundtrip/profile`: compiled AEP reopens through this parser and the profile/raw bytes match, but AE DOM does not provide reliable confirmation.
- `L1 parsed`: parser exposes the value but writer is not in recipe scope yet.
- `Blocked`: existing parser/writer does not provide a defensible contract.

`draft_3d` remains `L2 roundtrip/profile`; do not promote it unless a new AE-visible readback or render behavior proves it.

## Comp Field Matrix

| Group | Recipe path | Current write path | Profile path target | Evidence | Status |
| --- | --- | --- | --- | --- | --- |
| Identity | `comps[].name` | `aep.NewComposition` / `Composition.SetName` | `comps[].name` | L3 | base create + expected_profile done; rename recipe not planned yet |
| Geometry | `comps[].width` / `height` | `aep.NewComposition` / `Composition.SetSize` | `comps[].width` / `height` | L3 | base create + expected_profile done; resize recipe not planned yet |
| Timing | `comps[].frame_rate` / `duration` | `aep.NewComposition` / `SetFrameRate` / `SetDuration` | `comps[].frame_rate` / `duration_seconds` | L3 | base create + expected_profile done; mutation recipe not planned yet |
| Display | `background_color` | `Composition.SetBGColor` | `comps[].background_color` | L3 | done |
| Display | `resolution_factor` | `Composition.SetResolutionFactor` | `comps[].resolution_factor` | L3 | done |
| Display | `pixel_aspect` | `Composition.SetPixelAspect` | `comps[].pixel_aspect` | L3 | done |
| Display | `display_start_time` | `Composition.SetDisplayStartTime` | `comps[].display_start_time` | L3 | done |
| Display | `renderer` | `aep.SetRenderer` | `comps[].renderer` | L3 | done |
| Preview flags | `frame_blending` | `Composition.SetFrameBlending` | `comps[].frame_blending` | L3 | done |
| Preview flags | `draft_3d` | `Composition.SetDraft3D` | `comps[].draft_3d` | L2 | done |
| Preview flags | `hide_shy_layers` | `Composition.SetHideShyLayers` | `comps[].hide_shy_layers` | L3 | done |
| Nesting flags | `preserve_nested_frame_rate` | `Composition.SetPreserveNestedFrameRate` | `comps[].preserve_nested_frame_rate` | L3 | done |
| Nesting flags | `preserve_nested_resolution` | `Composition.SetPreserveNestedResolution` | `comps[].preserve_nested_resolution` | L3 | done |
| Motion blur | `motion_blur.enabled` | `Composition.SetCompMotionBlur` | `comps[].motion_blur.enabled` | L3 | done |
| Motion blur | `motion_blur.shutter_angle` | `Composition.SetShutterAngle` | `comps[].motion_blur.shutter_angle_degrees` | L3 | done |
| Motion blur | `motion_blur.shutter_phase` | `Composition.SetShutterPhase` | `comps[].motion_blur.shutter_phase` | L3 | done |
| Motion blur | `motion_blur.adaptive_sample_limit` | `Composition.SetMotionBlurAdaptiveSampleLimit` | `comps[].motion_blur.adaptive_sample_limit` | L3 | done |
| Motion blur | `motion_blur.samples_per_frame` | `Composition.SetMotionBlurSamplesPerFrame` | `comps[].motion_blur.samples_per_frame` | L3 | done |
| Work area | `work_area.start` / `end` | `Composition.SetWorkArea` | `comps[].work_area.start_seconds` / `end_seconds` | L3 | done |
| Item metadata | `label` | item label writer through `applyCompItemSettings` | `comps[].label` | L3 | done |
| Item metadata | `comment` | item comment writer through `applyCompItemSettings` | `comps[].comment` | L3 | done |

## Preferred Execution Order

1. Lock the comp object profile contract for already parsed core fields: `name`, `width`, `height`, `frame_rate`, `duration_seconds`, `renderer`, `work_area`, and motion-blur numeric settings.
2. Add missing profile-visible display fields: `background_color`, `resolution_factor`, `pixel_aspect`, `display_start_time`.
3. Add missing profile-visible comp flag fields: `frame_blending`, `hide_shy_layers`, `preserve_nested_frame_rate`, `preserve_nested_resolution`, `motion_blur.enabled`.
4. Add item metadata to profile in one deliberately named slice: `label` and `comment`.
5. Replace scattered comp recipe examples with one object-level fixture only when it stays readable; keep dedicated examples for fields with render gates or evidence caveats.
6. Add a comp object-level expected profile test that exercises all profile-visible comp settings in one recipe.
7. Keep `draft_3d` as a separate evidence note until AE evidence changes.

## Standard Slice Template

### Task N: Add One Comp Field Or One Same-Evidence Field Group

**Files:**
- Modify: `internal/profile/profile.go`
- Modify: `internal/recipe/schema.go`
- Modify: `internal/recipe/compiler.go`
- Modify: `internal/recipe/schema_test.go`
- Modify: `internal/recipe/compiler_test.go`
- Modify or create: `examples/recipes/minimal-comp-*.json`
- Modify when evidence is non-obvious: `flightdeck/knowledge/composition/*.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/comp-recipe-execution-strategy.md`

- [ ] **Step 1: Write the failing profile/expected-profile test**

Use `internal/recipe/compiler_test.go` when the field is recipe-owned. The failure should name the exact expected path, for example:

```go
assertProfileCheck(t, report, "expected_profile.background_color", true)
```

Expected failure before implementation: missing profile field, missing expected-profile check, or mismatch at the exact path.

- [ ] **Step 2: Run the focused failing test**

Run:

```powershell
go test ./internal/recipe -run TestCompileToFileSetsCompObjectProfile
```

Expected: FAIL for the field path under implementation.

- [ ] **Step 3: Implement the minimal profile contract**

Add the field to `profile.Composition` or a focused nested struct. Populate it from `scene.Composition` or cdta raw bytes only when the scene model does not already expose it.

```go
type Composition struct {
    BackgroundColor [3]uint8 `json:"background_color,omitempty"`
}
```

If `omitempty` would hide a meaningful default such as black `[0,0,0]`, use a representation that preserves the contract intentionally.

- [ ] **Step 4: Implement `expected_profile` checking**

Add the matching field to `ExpectedProfile` or a nested expected struct, include it in `hasExpectedProfile`, and report a stable path such as:

```go
add("expected_profile.background_color", expected.BackgroundColor, actual, equalFloatSlices(expected.BackgroundColor, actual))
```

- [ ] **Step 5: Confirm schema capability reporting**

If the recipe field already exists, keep the capability name unchanged. If adding a new recipe field, record the existing capability from `docs/capabilities.json`, for example `SetBGColor`, `SetResolutionFactor`, or `SetPixelAspect`.

Run:

```powershell
go test ./internal/recipe -run TestValidateReportsComp
```

Expected: PASS, with the capability asserted by name.

- [ ] **Step 6: Add or update the example recipe**

Prefer one readable object-level recipe:

```json
{
  "schema_version": 1,
  "project": {"name": "comp-object-profile"},
  "comps": [
    {
      "name": "Comp Object",
      "width": 640,
      "height": 360,
      "frame_rate": 30,
      "duration": 2
    }
  ],
  "expected_profile": {
    "comp_count": 1,
    "layer_count": 0
  }
}
```

Keep a dedicated recipe when a field needs its own render acceptance or evidence note.

- [ ] **Step 7: Run focused verification**

Run:

```powershell
go test ./internal/recipe ./internal/profile
go run ./cmd/aeprecipe validate -recipe examples\recipes\minimal-comp-object-profile.json -json
go run ./cmd/aeprecipe compile -recipe examples\recipes\minimal-comp-object-profile.json -out $env:TEMP\aep-parser-comp-object-profile.aep -json
Remove-Item -LiteralPath $env:TEMP\aep-parser-comp-object-profile.aep -Force
```

Expected: tests pass, validate reports `valid: true`, compile reports passing `profile_checks`.

- [ ] **Step 8: Run branch verification**

Run:

```powershell
go test ./...
go vet ./...
pwsh -NoProfile -File scripts\regen_fixtures.ps1 -CheckOnly
git diff --check
```

Expected: all pass. Existing fixture check may still report the known manual-only `re_template.jsx` entry; do not treat that as a new failure.

- [ ] **Step 9: Update flightdeck and commit**

Update the work index with one concise bullet, then commit:

```powershell
git add internal\profile\profile.go internal\recipe\schema.go internal\recipe\compiler.go internal\recipe\schema_test.go internal\recipe\compiler_test.go examples\recipes\minimal-comp-object-profile.json flightdeck\work\aep-understanding-generation\index.md flightdeck\work\aep-understanding-generation\comp-recipe-execution-strategy.md
git commit -m "feat: add recipe comp object profile field"
```

Use a more specific commit message when the slice is a named group, for example `feat: add recipe comp display profile checks`.

## Next Concrete Slice

The completed display slice added profile/expected-profile coverage for
`background_color`, `resolution_factor`, `pixel_aspect`, and
`display_start_time`.

The completed flag slice added profile/expected-profile coverage for
`frame_blending`, `hide_shy_layers`, `preserve_nested_frame_rate`,
`preserve_nested_resolution`, and `motion_blur.enabled`.

The next slice should be **comp item metadata profile checks**:

The completed metadata slice added profile/expected-profile coverage for comp
`label` and `comment`. Decision: expose these directly on
`profile.Composition`, not only on `profile.Items`, because comp recipe
contracts use the comp object as the assertion target while `profile.Items`
is an inventory/index surface.

The completed object consolidation slice expanded
`minimal-comp-object-profile.json` into a no-layer comp baseline that asserts
display, flag, motion-blur, work-area, renderer, and metadata fields together.
It now also asserts the base constructor-backed comp object fields: `name`,
`width`, `height`, `frame_rate`, and `duration`.

The standalone comp recipes for background color, label, comment, resolution
factor, pixel aspect, display start time, frame blending, hide shy layers,
preserve nested frame rate, preserve nested resolution, and motion blur enabled
now also assert their own field-level `expected_profile` contracts instead of
remaining count-only smoke tests.

The next object-level planning slice should create the same kind of execution
strategy for layer-level recipe/profile work, so layer fields stop advancing as
unconnected single-field slices.

## Self-Review

- Spec coverage: this plan covers object-level sequencing, per-field evidence, test shape, example strategy, verification, and commit discipline.
- Placeholder scan: no `TBD`, `TODO`, or unspecified "add tests" steps remain.
- Type consistency: paths use existing `CompSpec`, `ExpectedProfile`, `profile.Composition`, `Composition.Set*`, and `cmd/aeprecipe` names.
