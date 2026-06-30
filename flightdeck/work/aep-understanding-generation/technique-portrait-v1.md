# Technique Portrait v1

## Goal

Add the first project-level understanding layer above Technique Facts. A
portrait compresses raw facts into a stable, corpus-friendly project image that
can be compared across many `.aep` projects.

## Scope

Technique Portrait v1 adds:

- `technique.BuildPortrait(*FactSet) (*Portrait, error)`
- JSON-ready portrait model types in `internal/technique`
- `cmd/aeptechnique -mode portrait`
- `cmd/aeptechnique -portrait` as a shorthand compatibility flag
- `cmd/aeptechnique -mode explain` project explanations with deterministic
  `recreation_steps`

The portrait remains descriptive and conservative. It should expose mechanism
usage, graph shape, signal layers, and low-risk technique hints. It must not
claim style, artistic intent, or recipe-generation readiness.

The explain layer may describe recreation order, but only from data already
present in the portrait. It must not infer AE defaults or claim exact recreation
without profile/render evidence.

## Inputs

The authoritative input is `technique.FactSet`.

Portrait generation must not read raw AEP bytes, `aep.Project`,
`scene.Project`, or `profile.Profile` directly. If data is not present in
facts, the portrait should either omit it or expose it as unknown.

## Output Model

`Portrait` should include:

- `schema_version`
- `source_path`
- `fingerprint`
- `signal_layers[]`
- `mechanisms`
- `graph`
- `technique_hints[]`
- `unknowns`

## v1 Rules

### Fingerprint

Copy project-scale counts from facts:

- comp count
- layer count
- effect count
- text layer count
- shape layer count
- text animator count
- shape operator count
- dependency count
- unknown count
- layer role counts

### Signal Layers

Emit one signal layer for each layer that participates in at least one signal:

- has an effect
- has a keyframed effect param
- has an expression effect param
- has a layer-reference effect param
- has text animator facts
- has shape operator facts
- participates in source, parent, matte, light-source, or effect-param-layer
  dependency edges

Each signal layer records comp, layer, role, score, and signal labels.

### Mechanisms

Summarize:

- effect dependency class counts
- effect match-name counts
- text animator kind counts
- shape operator family counts
- reproducibility counts derived from effect dependency class

The initial reproducibility buckets are:

- `native`
- `cycore`
- `third_party`
- `unknown`

### Graph

Summarize dependency relation counts and record the total edge count.

### Technique Hints

Technique hints are conservative corpus tags, not full interpretation. v1 can
emit:

- `kinetic_text` when text animator facts include keyframes or expressions
- `shape_operator_stack` when a shape layer has two or more shape operator facts
- `precomp_assembly` when a layer sources another composition
- `effect_driven_layer` when an effect has changed/tuned params, keyframes,
  expressions, or layer refs
- `matte_composite` when matte dependencies exist
- `controller_rig` when effect params reference layers or controller-role
  layers participate in graph edges
- `plugin_dependent` when third-party effects are present

Every hint records confidence and the signal labels that caused it.

## Non-Goals

- No natural-language report.
- No recipe generation.
- No style classification.
- No ML/clustering.
- No AE render validation.
- No inference from raw AEP bytes.

## Explanation Recreation Steps

`BuildExplanation` emits `recreation_steps[]` as a deterministic execution
outline for humans and downstream tools. The steps are sorted by `priority` and
use stable IDs:

- `structure`: create compositions and restore graph edges.
- `layers`: restore layer stacks, roles, timing, sources, parents, and mattes.
- `shapes`: restore shape operators and family-specific properties.
- `text`: restore text animator properties.
- `effects`: apply effect stacks after layer identity exists.
- `controllers`: reconnect controller and layer-reference dependencies.
- `unknowns`: resolve unknown parsed fields before claiming exact recreation.

Each step contains `inputs`, `risks`, and `evidence` strings derived from the
portrait summary. These are planning aids, not generated recipes.

## Testing

Use TDD.

Minimum tests:

- synthetic facts produce fingerprint, mechanism, graph, signal-layer, and hint
  summaries
- empty facts produce an empty but valid portrait
- CLI emits valid JSON for `-mode portrait`
- CLI accepts `-portrait`

## Acceptance

- `go test ./internal/technique ./cmd/aeptechnique -count=1` passes
- `go test ./...` passes
- `go vet ./...` passes
- `go run ./cmd/aepverify recipe-profiles` passes
- `scripts/fixtures/regen_fixtures.ps1 -CheckOnly` reports no new unmanaged AEPs
- `git diff --check` passes

## Progress

Current:

- Spec written.
- `internal/technique` implements `BuildPortrait(*FactSet)` with deterministic
  aggregation for fingerprints, signal layers, mechanisms, graph summaries,
  conservative technique hints, and unknown summaries.
- `cmd/aeptechnique` supports `-mode facts`, `-mode portrait`, and `-portrait`.
- `BuildExplanation` emits `recreation_steps[]` so reports can show the
  project-level recreation order and blockers without guessing missing fields.
