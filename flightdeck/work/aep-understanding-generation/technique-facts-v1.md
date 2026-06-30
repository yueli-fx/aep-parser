# Technique Facts v1

## Goal

Add a first AEP understanding layer that turns a normalized
`profile.Profile` into structured technique facts. This layer explains what a
project appears to contain without parsing raw AEP bytes, writing AEPs, or
guessing beyond current profile evidence.

## Scope

Technique Facts v1 creates a reusable package and CLI:

- `internal/technique`
  - `Build(*profile.Profile) (*FactSet, error)`
  - stable JSON model for comp, layer, effect, text animator, shape operator,
    dependency, and unknown facts
- `cmd/aeptechnique`
  - opens an AEP
  - builds `profile.Profile`
  - emits technique facts as JSON

The first slice is intentionally descriptive. It does not generate recipes,
score projects, cluster corpus data, or produce natural language summaries.

## Inputs

The authoritative input is `internal/profile.Profile`.

The builder must not depend on raw `aep.Project`, `scene.Project`, or
serializer chunks. If a fact cannot be derived from profile data, it is out of
scope until profile exposes a defensible contract.

## Output Model

`FactSet` should include:

- `schema_version`
- `source_path`
- `summary`
- `comps[]`
- `layers[]`
- `effects[]`
- `text_animators[]`
- `shape_operators[]`
- `dependencies[]`
- `unknowns[]`

Every fact that represents a profile entity should carry:

- stable profile path
- display path when available
- evidence copied or derived from profile evidence
- confidence string, initially `high`, `medium`, or `low`

## v1 Fact Rules

### Comp Summary

For each comp:

- record ID, name, dimensions, frame rate, duration, layer count
- mark a main-comp candidate when it has the largest layer count, tie-broken by
  source order

### Layer Roles

For each layer:

- `text` for profile layer type `text`
- `shape` for profile layer type `shape`
- `camera`, `light`, `null`, `adjustment` from type/flags
- `precomp` for source refs with kind `composition`
- `solid` for source refs with kind `footage`
- `controller` when the layer is null or adjustment and has no visible source
  content but participates through parent, effect, or expression facts
- otherwise `unknown`

Role assignment is a heuristic. Each layer fact should include the selected
role and confidence. Do not hide the original profile layer type.

### Dependencies

Emit dependency facts for:

- layer source refs
- parent refs
- matte refs
- light source refs
- effect parameter layer refs

Dependencies should include source layer path/name when available, target
name/ID when available, and relation type.

### Effects

For each effect:

- record match name, display name, dependency class, layer path, occurrence
- count changed/tuned params
- flag whether any param has keyframes
- flag whether any param has expression
- flag whether any param has a layer ref
- record unknown param count

Third-party or pseudo effects are not failures; they are facts with their
dependency class and unknown counts.

### Text Animators

Profile text animator support currently appears as stable property match names
under text layers. v1 should detect known `ADBE Text ...` property match names
from `profile.Layer.Properties` and emit:

- property match name
- inferred property kind such as `opacity`, `position`, `scale`, `rotation`,
  `fill_color`, `stroke_color`, `tracking`, `character_offset`, `skew`
- keyframed/expression/static flags

### Shape Operators

Detect shape operator facts from `profile.Layer.Shapes[].Kind` and
`profile.Layer.Properties` match names. v1 should cover stable broad families:

- fill
- stroke
- gradient_fill
- gradient_stroke
- trim
- repeater
- merge_paths
- offset_paths
- round_corners
- zigzag
- pucker_bloat
- twist
- wiggle_paths
- wiggle_transform
- star_or_polygon

### Unknowns

Carry through:

- `profile.Unknowns`
- effect `UnknownParams`

Unknown facts should preserve path, reason, evidence, and owning context when
it can be derived.

## Non-Goals

- No recipe generation.
- No automatic correction loop.
- No AE render validation.
- No ML or clustering.
- No direct raw AEP traversal.
- No new parser or writer fields.
- No unsupported-field inference presented as fact.

## Testing

Use TDD.

Minimum tests:

- synthetic profile with comp/layer/effect/dependency facts
- text layer profile properties produce text animator facts
- shape profile properties produce shape operator facts
- unknowns and effect unknown params are preserved
- CLI emits valid JSON for a small fixture

## Acceptance

The first slice is accepted when:

- `internal/technique` has deterministic tests
- `cmd/aeptechnique -in <fixture>` emits valid JSON; `-json` is accepted as
  an explicit compatibility flag
- `go test ./internal/technique ./cmd/aeptechnique -count=1` passes
- `go test ./...` passes
- `go vet ./...` passes
- `scripts/verify_recipe_profiles.ps1` still passes
- `scripts/regen_fixtures.ps1 -CheckOnly` has no new unmanaged fixture changes

## Progress

Current:

- Spec written.
- `internal/technique` implemented with deterministic tests for summary,
  layer roles, dependencies, effect facts, text animator facts, shape operator
  facts, and unknown preservation.
- `cmd/aeptechnique` implemented as a thin JSON CLI over `aep.Open`,
  `profile.Build`, and `technique.Build`.
