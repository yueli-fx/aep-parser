# AEP Understanding, Replication Diagnostics, and Generation Spec

## Problem

The project has made strong progress as an `.aep` parser and writer, and the
Booyah Glitch replication proved that a real project can be rebuilt from Go
APIs. But the current replication flow does not scale: a single 1:1 clone can
turn into manual frame-by-frame visual debugging, and the conclusions are easy
to trap inside one showcase directory.

The long-term target has two parts:

1. Learn as much of the AEP structure and parameters as possible, with explicit
   evidence levels and unknowns.
2. Feed many real projects into a system that lets AI generate new AEPs from a
   user request without guessing raw chunks.

This spec defines the bridge between those goals: audit the current export
surfaces, convert AEP files into a stable profile, compare original vs generated
projects mechanically, use render diff only as a focused oracle, and turn
verified findings into reusable techniques and generation recipes.

The spec intentionally separates two tracks:

- **Understanding and replication diagnostics**: parse, profile, diff, render
  selected frames, and produce capability gaps.
- **Intent-to-AEP generation**: compile a higher-level recipe into supported
  writer APIs only after the diagnostic track can prove what is known, unknown,
  and unsupported.

Generation is blocked until the diagnostic track has stable profiles, stable
paths, actionable gap reports, and at least one non-Booyah project processed by
the generic workflow.

## Current Codebase Assessment

The codebase supports more of this than it first appears to:

- `cmd/aepdissect` already parses a project, classifies effect dependencies,
  extracts effect parameters, motion summaries, expression references, graph
  hints, and has a `-json` mode.
- `internal/scene/write_json.go` already exports a broad JSON view including
  comps, layers, effects, properties, keyframes, masks, shape primitives, text,
  guides, essential graphics, and render queue.
- `internal/scene` exposes a rich parsed model: comp settings, layer flags,
  source IDs, parent/matte references, masks, text, shape graph, properties,
  effects, keyframes, expressions, markers, render queue, and raw-byte escape
  hatches for some fields.
- `internal/aep` exposes many stable/alpha write APIs, with `docs/capabilities.md`
  acting as the capability truth source.
- `flightdeck/showcase/booyah-clone` contains practical replication tools:
  original oracle reads, per-comp generators, AE render scripts, timing dumps,
  solo-layer renders, and original-vs-clone frame comparisons.
- `flightdeck/knowledge/techniques/understand-a-project.md` and
  `fx-techniques.md` already define the beginning of a technique ontology.

The missing piece is a reusable pipeline. Today the profile shape is split
between `aepdissect -json`, `Project.WriteJSON`, Booyah-specific helpers, and
human-written ledger notes. The next work should audit and consolidate those
pieces before attempting broader AI generation.

## Non-Goals

- Do not train or fine-tune a model in this phase.
- Do not attempt full 1:1 visual equivalence for every imported project.
- Do not ask AI to write raw RIFX chunks.
- Do not move work-in-progress specs or plans into long-term `knowledge/`.
- Do not declare a feature "known" unless it has explicit evidence under the
  evidence ladder below.

## Core Contracts

These contracts are Phase 0/1 requirements, not cleanup tasks. Diff, ignore
rules, gap ledgers, and future generation all depend on them.

### Evidence Ladder

Every surfaced field, capability assertion, and generated gap should be able to
name the strongest evidence currently available.

```yaml
evidence:
  level: L0_raw | L1_parsed | L2_roundtrip | L3_ae_accept | L4_render | L5_user_verified
  source:
  confidence:
  notes:
```

Definitions:

- `L0_raw`: bytes or raw chunks are present, but the structure is not decoded.
- `L1_parsed`: the parser exposes a structured field with deterministic tests.
- `L2_roundtrip`: the field survives read/write/read without structural loss.
- `L3_ae_accept`: After Effects opens or renders the written project without
  rejecting the construct.
- `L4_render`: selected rendered pixels match within the configured tolerance.
- `L5_user_verified`: a human review confirms the result for the target use.

Evidence is field-scoped. A project can have `L4_render` evidence for a visible
text layer while a hidden effect parameter remains `L1_parsed`.

### Stable Path Contract

Profile paths are machine identifiers, not display strings. They must be
deterministic across repeated parses of the same AEP and resilient to common
display-name collisions.

Path records should carry both a human path and an identity key:

```yaml
path: comps.by_id[17].layers.by_index[3].effects.by_match_name["ADBE Fill"]
display_path: comps["main"].layers[3:"white flash"].effects["Fill"]
identity:
  comp_id:
  layer_id:
  layer_index:
  property_match_name:
  occurrence:
```

Rules:

- Prefer stable project IDs where the parsed model exposes them.
- Include index and occurrence where AE permits duplicate names or duplicate
  effect match names.
- Use `matchName` for effects/properties; keep display names as annotations.
- Treat unstable paths as first-class diff results instead of silently falling
  back to text matching.
- Add deterministic path tests before ignore rules or gap ledgers depend on
  those paths.

### Profile vs Detailed JSON Export

`Project.WriteJSON` remains the detailed parsed-scene export. `Profile` is the
normalized comparison and learning layer. It should not copy every field from
`WriteJSON` just because it exists.

Field admission criteria for `Profile`:

- The field affects understanding, diffing, render fidelity, capability mapping,
  or future recipe generation.
- The field has a stable identity path or can explicitly report that it does
  not.
- The field can carry evidence and unknown state.
- The field is compact enough for repeated project runs, or it is summarized
  with a pointer to the detailed export.

Phase 0 must produce a coverage matrix:

```text
field | aepdissect-json | WriteJSON | profile-core | evidence | notes
```

### Parameter Semantics

Use separate buckets for raw visibility and generation relevance:

- `params`: all surfaced effect/property parameters that the parser can expose
  in stable form.
- `tuned_params`: the subset likely to materially affect the authored result,
  selected by rules or evidence, not by display-name guesswork alone.
- `unknown_params`: present but not decoded or not safely comparable.

### Diff Record Model

Severity and kind are separate. Severity is context-sensitive; diff kind is the
stable classification.

```yaml
diff:
  path:
  kind: missing_object | extra_object | wrong_value | unsupported_construct | unstable_path | unknown_field | render_mismatch
  severity: blocker | fidelity | polish | acceptable | unknown
  evidence:
  expected:
  actual:
  action_type: parse | write | semantics | render | asset | plugin | ignore | investigate
  human_notes:
```

Ignore rules must be structured and versioned:

```yaml
ignore:
  schema_version: 1
  rules:
    - path:
      kind:
      condition:
      reason:
      expires_when:
```

### Gap Ledger Boundary

Diffs are observations. Gaps are triaged work items derived from diffs, render
failures, or manual investigation.

Gap entries should use a constrained `action_type` plus freeform notes, not a
single unconstrained suggestion string:

```yaml
gap:
  id:
  type:
  profile_path:
  source_project:
  observed_in:
  evidence:
  severity:
  action_type: parse | write | semantics | render | asset | plugin | docs | knowledge | investigate
  human_notes:
  linked_capability:
  linked_knowledge:
```

Only curated, self-contained conclusions move into `flightdeck/knowledge/**`.
One verified project can be enough if the conclusion is reusable and evidence is
clear; one-off diagnostics stay in `flightdeck/work/**` or `tmp_debug/**`.

## Recommended Architecture

### 1. Export Surface Audit

Before creating `internal/profile`, audit the two existing JSON-facing surfaces:

- `cmd/aepdissect -json`: current diagnostic summary and shallow project
  understanding view.
- `internal/scene.Project.WriteJSON`: broad parsed-scene export.

Run both against Booyah and at least one small deterministic fixture, then
record:

- overlapping fields,
- fields unique to each surface,
- conflicting names or meanings,
- fields that belong in profile core,
- fields that should stay in detailed JSON only,
- fields that need new parser or writer work before they can be trusted.

This is Phase 0 because the profile package should be based on observed overlap
and gaps, not on copying either current shape wholesale.

### 2. Profile Core

Create a reusable package, likely `internal/profile`, that replaces the private
`jProfile` types inside `cmd/aepdissect` after the audit.

The profile is not just a pretty JSON export. It is the canonical normalized
view used for:
- project understanding,
- original-vs-clone structural diff,
- capability gap detection,
- technique extraction,
- future generation recipes.

Initial profile sections:

```yaml
profile:
  schema_version
  meta:
    path
    parse_warnings
    bits_per_channel
    expression_engine
  fingerprint:
    comp_count
    layer_count
    footage_count
    effect_usage
    plugin_dependencies
    max_nest_depth
  items:
    comps
    footage
    folders
  comps:
    id
    name
    size
    fps
    duration
    tick_rate
    renderer
    work_area
    motion_blur_settings
    layers
  layers:
    id
    index
    name
    type
    source_ref
    timing
    flags
    blend
    parent_ref
    matte_ref
    transform
    effects
    masks
    shapes
    text
    expressions
    unknowns
  effects:
    match_name
    display_name
    dependency_class
    params
    tuned_params
    unknown_params
    layer_refs
    expressions
  properties:
    match_name
    path
    value_kind
    static_value
    keyframes
    interpolation
    temporal_ease
    spatial_tangents
    expression
  evidence:
    level
    source
    confidence
  unknowns:
    path
    reason
```

`cmd/aepdissect` should become a thin CLI around this package. Its current text
report remains useful, but JSON should become stable and testable.

### 3. Profile Diff

Add a deterministic diff layer, likely `internal/profilediff` plus a CLI:

```text
go run ./cmd/aepdiff original.aep clone.aep
go run ./cmd/aepdiff -json original.aep clone.aep
```

Diff categories should use the `kind` plus `severity` model in the core
contracts. Static P0/P1/P2/P3 labels are not enough because the same field can be
critical or cosmetic depending on the comp. For example, mask geometry can be a
blocker in a matte-driven comp and polish in a hidden helper layer.

The diff should support ignore rules by stable path, not by ad hoc text. Example:

```yaml
ignore:
  - path: comps["main"].layers["Curves"].effects["ADBE CurvesCustom"].params.curve_data
    reason: "arbitrary curve data not writable yet"
```

Booyah becomes the first fixture, but the diff model must be generic.

### 4. Render Oracle

Render comparison should become targeted, not frame-by-frame by default.

Add a generic render harness around existing AE scripts:

```text
go run ./cmd/aeoracle render --aep file.aep --comp "main" --frames 0,15,30
go run ./cmd/aeoracle compare --original original.aep --clone clone.aep --comp "main"
```

Initial implementation can reuse PowerShell/JSX scripts instead of replacing
them. The important change is a stable contract:

- input sidecar format,
- output directory convention,
- metadata JSON containing AE version, OS, project path, comp, frame, PNG path,
  status, detected plugin dependencies, font/dependency warnings where
  available,
- deterministic frame naming: `NN_original.png`, `NN_clone.png`,
- optional per-layer solo render when a comp frame fails.

Default frame selection should be sentinel based:
- first frame,
- last visible frame,
- bounded keyframe buckets,
- midpoints between selected keyframe buckets,
- a small fixed sample for long quiet spans.

Frame selection must have a project-configurable cap. Full frame-by-frame
comparison remains an escalation mode, not the normal path.

Comparison output should report the metric and tolerance used, for example exact
pixel count, per-channel threshold, PSNR, and/or SSIM. Exact match is allowed
only when the environment is controlled enough to make that meaningful.

### 5. Gap Ledger

Add a machine-readable gap ledger generated from profile/diff/render findings.

Gap types:

- `parse-gap`: field exists in AEP but profile cannot surface it.
- `write-gap`: parser can read it, but no public write API exists.
- `semantic-gap`: write API exists but does not preserve AE semantics.
- `render-gap`: structure looks equal but rendered pixels differ.
- `plugin-gap`: effect depends on third-party plugin.
- `asset-gap`: visual content is footage, not procedurally generated.
- `lib-blocked`: AE can do it, but this library currently cannot express it
  safely, such as known expression evaluation limitations.

Each gap should include:

```yaml
gap:
  id
  type
  profile_path
  source_project
  observed_in
  evidence
  severity
  action_type
  human_notes
  linked_capability
  linked_knowledge
```

This becomes the operational bridge from "clone differs" to "what code must be
upgraded next".

### 6. Technique and Recipe Layer

Do not make AI learn chunks. AI should learn:

- project profiles,
- reusable techniques,
- phenomenon recipes,
- capability boundaries,
- successful generation examples.

Generation should eventually look like:

```text
user request
→ technique selection
→ recipe IR
→ AEP compiler using internal/aep APIs
→ profile diff against intended structure
→ AE render oracle
→ correction loop
```

The recipe IR can be introduced later. It should not be the first milestone
because the project first needs profile and diff stability.

Minimum model relationships:

- `Capability`: what the parser/writer/render harness can prove today.
- `Technique`: a reusable authored mechanism, with evidence and constraints.
- `Recipe`: an intended construction plan that references techniques and must
  compile only through supported capabilities.

Recipes must degrade or refuse unsupported constructs explicitly. They should
not invent raw chunks or silently substitute effects without reporting the
change.

## Phased Plan

### Phase 0 — Export Audit and Contract Lock

Goal: decide what the stable profile is allowed to be before extracting code.

Work:
- Run `cmd/aepdissect -json` and `Project.WriteJSON` on Booyah plus small
  deterministic fixtures.
- Produce the coverage matrix for high-value fields: comp graph, layers,
  sources, effects, properties, keyframes, expressions, masks, text, shape
  graph, render queue, and unknown/raw escapes.
- Decide which fields enter profile core, which remain detailed JSON, and which
  need parser upgrades.
- Define schema version, evidence ladder encoding, stable path format, and
  field admission rules.
- Add deterministic path contract tests using hand-built or small fixture
  profiles.

Exit criteria:
- The profile design references observed current exports, not assumptions.
- Stable path examples cover duplicate names, duplicate effects, and missing
  IDs.
- The first implementation slice can be scoped without copying all of
  `WriteJSON`.

### Phase 1 — Profile Stabilization

Goal: make "understand this AEP" a stable library function and JSON contract.

Work:
- Extract `cmd/aepdissect` profile structs/build logic into `internal/profile`.
- Add only the `Project.WriteJSON` fields admitted by the Phase 0 coverage
  matrix.
- Keep `cmd/aepdissect` text output intact by consuming the new package.
- Add a path/debug mode, such as `aepdissect -paths`, or equivalent test helper
  output for stable identity paths.
- Add golden tests on small fixtures and one Booyah slice.
- Document profile schema version, e.g. `schema_version: 1`.

Exit criteria:
- `go test ./internal/profile ./cmd/aepdissect` passes.
- `go run ./cmd/aepdissect -json <known fixture>` emits deterministic JSON.
- Profile includes enough layer/effect/property identity to build stable paths.
- Keyframe timelines are complete enough for Phase 2 structural/semantic diff
  and Phase 3 sentinel selection to reason about active intervals.

### Phase 2 — Structural and Semantic Diff

Goal: stop using manual ledger inspection for original-vs-clone structure.

Work:
- Add `internal/profilediff`.
- Add `cmd/aepdiff`.
- Implement path-based diff records and ignore rules.
- Split diff `kind`, context-derived `severity`, and constrained
  `action_type`.
- Cover comp/layer/source/effect/property/keyframe/mask/text high-value fields.
- Run against Booyah original vs current clone and record the first gap report.

Exit criteria:
- Diff output points to specific profile paths.
- Known acceptable deltas can be ignored with explicit reasons.
- Unknown fields are reported as unknown, not silently skipped.
- Ignore rules include schema version, conditions, and expiry or replacement
  criteria where practical.

### Phase 3 — Render Oracle Harness

Goal: make render comparison targeted and repeatable.

Work:
- Wrap existing `scripts/ae_run.ps1` and Booyah JSX patterns into generic
  `cmd/aeoracle` or tracked reusable scripts.
- Implement capped sentinel frame selection from profile keyframes and active
  intervals.
- Produce render metadata JSON and deterministic image names.
- Add solo-layer escalation for failing comp frames.
- Report comparison metric, tolerance, and environment metadata.

Exit criteria:
- A single command can render original/clone sentinel frames for a comp.
- When a frame differs, a solo-layer run can identify candidate layers.
- Output is suitable for user review and for automated gap notes.
- Render results are reproducible enough to explain whether a delta is content,
  environment, plugin, font, or unsupported-authoring related.

### Phase 4 — Gap Ledger and Capability Coupling

Goal: turn failures into prioritized code upgrade tasks.

Work:
- Add `cmd/aepgaps` or integrate gap output into `aepdiff`.
- Link profile paths to `docs/capabilities.json` where possible.
- Emit gap YAML/JSON into `tmp_debug` for runs; only curated conclusions move
  into `flightdeck/knowledge`.
- Define severity and action type separately.
- Define knowledge-promotion rules: reusable, self-contained, evidence-tagged,
  and not dependent on a flowing spec or plan.

Exit criteria:
- Booyah remaining deltas are represented as structured gaps.
- New projects can produce a gap report before any hand-written clone code.
- Gap reports can be converted into either code work, documentation updates,
  or knowledge entries without reinterpreting freeform notes.

### Phase 5 — Slice Replication Workflow

Goal: replace "whole project 1:1 immediately" with a scalable loop.

Workflow:

```text
input AEP
→ profile
→ classify procedural vs footage+assembly vs plugin-dependent
→ choose 1-3 representative comps/slices
→ generate clone slice
→ profile diff
→ render oracle
→ gap ledger
→ code upgrade or knowledge update
```

Exit criteria:
- A second real project can be processed without creating bespoke one-off tools.
- The system can say which parts are cloneable now and which require code work.

### Phase 6 — Recipe IR and Generation Readiness

Goal: generate new AEPs from intent without raw chunk guessing.

Readiness gate before starting:
- Phase 0-4 contracts exist and are tested.
- At least one non-Booyah project has produced a profile, diff, render report,
  and gap ledger through the generic workflow.
- `docs/capabilities` can be queried from generation code or a generated
  machine-readable artifact.
- Unsupported constructs have downgrade/refusal behavior, not silent guessing.

Subphases:
- **6a recipe compiler**: define a minimal recipe schema around
  comps/layers/techniques and compile only through supported `internal/aep`
  APIs.
- **6b editable IR**: allow recipe inspection and deterministic edits before
  writing an AEP.
- **6c render/report loop**: render sentinel previews and report capabilities,
  techniques, gaps, and substitutions.
- **6d limited correction**: allow automated correction only inside bounded,
  typed parameters with evidence, never arbitrary raw chunks.

Exit criteria:
- A simple prompt can produce a recipe, an AEP, and a preview image.
- The output reports which techniques and verified capabilities were used.
- Unsupported requests are degraded or refused with a capability-grounded
  explanation.

## Data Ownership Rules

- Work-in-progress specs and plans stay under `flightdeck/work/**`.
- Long-term, self-contained conclusions go under `flightdeck/knowledge/**`.
- Render output and bulk generated profiles stay under `tmp_debug/**` unless
  intentionally promoted.
- `tmp_debug/**` outputs are disposable run artifacts. Any command that writes
  there should be safe to clean and should not be required to understand a
  knowledge entry.
- Capability truth remains `docs/capabilities.md` / `docs/capabilities.json`.
- Technique truth remains `flightdeck/knowledge/techniques/**`.

## Testing Strategy

Profile:
- Phase 0 coverage-matrix checks comparing admitted profile fields against
  existing export surfaces.
- Golden JSON tests on small deterministic fixtures.
- Fixture tests for key fields: source graph, parent/matte, effect params,
  expressions, keyframes, masks, text.
- Deterministic stable-path tests covering duplicate display names and duplicate
  effect/property occurrences.

Diff:
- Table tests comparing hand-built minimal profiles.
- Integration tests on tiny fixture pairs with known differences.
- Tests that severity can change by context while diff kind remains stable.

Render oracle:
- Unit-test sidecar generation and output path conventions.
- Unit-test sentinel frame selection with caps and long quiet spans.
- Keep AE-dependent runs gated behind existing ship-gate environment rules.

Gap ledger:
- Tests that known diffs map to the right gap type.
- Tests that ignored deltas require a non-empty reason.
- Tests that `action_type` is constrained and `human_notes` is optional
  narrative, not the machine contract.

Full verification:
- `go test ./...`
- `go vet ./...`
- AE ship-gates only for milestones that touch render harness or writer behavior.

Performance and compatibility:
- Profile and diff commands should default to deterministic, non-AE runs.
- Large AEPs should produce bounded summaries for noisy repeated fields.
- AE-dependent commands must record AE version and OS, and should surface
  plugin/font/dependency warnings when detectable.
- Parse failures should return partial profiles with explicit errors where
  possible, not only a process-level failure.

## Risks

- Profile bloat: copying all of `WriteJSON` into profile would create noisy diffs.
  Mitigation: Phase 0 coverage matrix, field admission criteria, stable identity
  paths, evidence/confidence fields, and detailed raw JSON as a secondary export.
- False visual failures: antialiasing, font availability, fps differences, and
  plugin availability can create expected render deltas. Mitigation: sentinel
  metadata and explicit ignore/gap reasons.
- AI overreach: LLMs may invent unsupported AEP features. Mitigation: recipe
  compiler must be capability-aware and refuse or downgrade unsupported
  operations with explicit reporting.
- Tool fragmentation: Booyah scripts could remain one-off. Mitigation: Phase 3
  promotes their contracts into reusable render oracle tooling.
- Premature auto-correction: a correction loop can hide real parser or writer
  bugs. Mitigation: Phase 6d is gated and limited to typed parameters with
  evidence.

## Recommended First Implementation Slice

Start with Phase 0, then Phase 1 only:

1. Audit `cmd/aepdissect -json` versus `Project.WriteJSON`.
2. Write the coverage matrix and profile field admission list.
3. Lock the evidence ladder and stable path contract.
4. Create `internal/profile`.
5. Move the admitted `cmd/aepdissect` JSON model into it.
6. Add stable IDs/paths for project, comp, layer, effect, property.
7. Add enough fields for profile diff to be possible later.
8. Keep the CLI behavior compatible.

Do not start with generation. The generation system depends on profile and diff
being reliable; otherwise AI will only produce plausible-looking but unverified
projects.
