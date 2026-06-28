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

This spec defines the bridge between those goals: convert AEP files into a
stable profile, compare original vs generated projects mechanically, use render
diff only as a focused oracle, and turn repeated findings into reusable
techniques and generation recipes.

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
human-written ledger notes. The next work should consolidate those pieces before
attempting broader AI generation.

## Non-Goals

- Do not train or fine-tune a model in this phase.
- Do not attempt full 1:1 visual equivalence for every imported project.
- Do not ask AI to write raw RIFX chunks.
- Do not move work-in-progress specs or plans into long-term `knowledge/`.
- Do not declare a feature "known" unless it has evidence: parse, roundtrip,
  AE accept, render-pixel, or user verification.

## Recommended Architecture

### 1. Profile Core

Create a reusable package, likely `internal/profile`, that replaces the private
`jProfile` types inside `cmd/aepdissect`.

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
    source: parsed | inferred | raw | ae_dom | render
    confidence
```

`cmd/aepdissect` should become a thin CLI around this package. Its current text
report remains useful, but JSON should become stable and testable.

### 2. Profile Diff

Add a deterministic diff layer, likely `internal/profilediff` plus a CLI:

```text
go run ./cmd/aepdiff original.aep clone.aep
go run ./cmd/aepdiff -json original.aep clone.aep
```

Diff categories:

- **P0 structural mismatch**: missing comp, missing layer, wrong source, wrong
  parent/matte, AE would render a different graph.
- **P1 semantic mismatch**: effect missing, tuned param differs, keyframe count
  differs, expression differs, layer flag differs.
- **P2 fidelity mismatch**: interpolation/ease differs, timing offset differs,
  mask geometry differs, text style differs.
- **P3 known acceptable delta**: different item IDs, expected 30 vs 29.97
  workaround, default-elided vs explicitly written equivalent value.
- **Unknown**: parser cannot compare because field is not surfaced or only raw
  bytes are available.

The diff should support ignore rules by stable path, not by ad hoc text. Example:

```yaml
ignore:
  - path: comps["main"].layers["Curves"].effects["ADBE CurvesCustom"].params.curve_data
    reason: "arbitrary curve data not writable yet"
```

Booyah becomes the first fixture, but the diff model must be generic.

### 3. Render Oracle

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
- metadata JSON containing AE version, comp, frame, PNG path, status,
- deterministic frame naming: `NN_original.png`, `NN_clone.png`,
- optional per-layer solo render when a comp frame fails.

Default frame selection should be sentinel based:
- first frame,
- last visible frame,
- every keyframe time bucket,
- midpoints between keyframes,
- a small fixed sample for long quiet spans.

Full frame-by-frame comparison remains an escalation mode, not the normal path.

### 4. Gap Ledger

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
  suggested_next_action
  linked_capability
  linked_knowledge
```

This becomes the operational bridge from "clone differs" to "what code must be
upgraded next".

### 5. Technique and Recipe Layer

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

## Phased Plan

### Phase 1 — Profile Stabilization

Goal: make "understand this AEP" a stable library function and JSON contract.

Work:
- Extract `cmd/aepdissect` profile structs/build logic into `internal/profile`.
- Merge useful fields from `Project.WriteJSON` into profile output where they
  serve comparison or understanding.
- Keep `cmd/aepdissect` text output intact by consuming the new package.
- Add golden tests on small fixtures and one Booyah slice.
- Document profile schema version, e.g. `schema_version: 1`.

Exit criteria:
- `go test ./internal/profile ./cmd/aepdissect` passes.
- `go run ./cmd/aepdissect -json <known fixture>` emits deterministic JSON.
- Profile includes enough layer/effect/property identity to build stable paths.

### Phase 2 — Structural and Semantic Diff

Goal: stop using manual ledger inspection for original-vs-clone structure.

Work:
- Add `internal/profilediff`.
- Add `cmd/aepdiff`.
- Implement path-based diff records and ignore rules.
- Cover comp/layer/source/effect/property/keyframe/mask/text high-value fields.
- Run against Booyah original vs current clone and record the first gap report.

Exit criteria:
- Diff output points to specific profile paths.
- Known acceptable deltas can be ignored with explicit reasons.
- Unknown fields are reported as unknown, not silently skipped.

### Phase 3 — Render Oracle Harness

Goal: make render comparison targeted and repeatable.

Work:
- Wrap existing `scripts/ae_run.ps1` and Booyah JSX patterns into generic
  `cmd/aeoracle` or tracked reusable scripts.
- Implement sentinel frame selection from profile keyframes and active intervals.
- Produce render metadata JSON and deterministic image names.
- Add solo-layer escalation for failing comp frames.

Exit criteria:
- A single command can render original/clone sentinel frames for a comp.
- When a frame differs, a solo-layer run can identify candidate layers.
- Output is suitable for user review and for automated gap notes.

### Phase 4 — Gap Ledger and Capability Coupling

Goal: turn failures into prioritized code upgrade tasks.

Work:
- Add `cmd/aepgaps` or integrate gap output into `aepdiff`.
- Link profile paths to `docs/capabilities.json` where possible.
- Emit gap YAML/JSON into `tmp_debug` for runs; only curated conclusions move
  into `flightdeck/knowledge`.
- Define severity: blocker, fidelity, polish, unsupported.

Exit criteria:
- Booyah remaining deltas are represented as structured gaps.
- New projects can produce a gap report before any hand-written clone code.

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

### Phase 6 — Recipe IR and Generation

Goal: generate new AEPs from intent without raw chunk guessing.

Work:
- Define recipe schema around comps/layers/techniques, not binary chunks.
- Add compiler from recipe IR to `internal/aep` calls.
- Add capability-aware fallbacks for unsupported features.
- Use render oracle for correction.

Exit criteria:
- A simple prompt can produce a recipe, an AEP, and a preview image.
- The output reports which techniques and verified capabilities were used.

## Data Ownership Rules

- Work-in-progress specs and plans stay under `flightdeck/work/**`.
- Long-term, self-contained conclusions go under `flightdeck/knowledge/**`.
- Render output and bulk generated profiles stay under `tmp_debug/**` unless
  intentionally promoted.
- Capability truth remains `docs/capabilities.md` / `docs/capabilities.json`.
- Technique truth remains `flightdeck/knowledge/techniques/**`.

## Testing Strategy

Profile:
- Golden JSON tests on small deterministic fixtures.
- Fixture tests for key fields: source graph, parent/matte, effect params,
  expressions, keyframes, masks, text.

Diff:
- Table tests comparing hand-built minimal profiles.
- Integration tests on tiny fixture pairs with known differences.

Render oracle:
- Unit-test sidecar generation and output path conventions.
- Keep AE-dependent runs gated behind existing ship-gate environment rules.

Gap ledger:
- Tests that known diffs map to the right gap type.
- Tests that ignored deltas require a non-empty reason.

Full verification:
- `go test ./...`
- `go vet ./...`
- AE ship-gates only for milestones that touch render harness or writer behavior.

## Risks

- Profile bloat: copying all of `WriteJSON` into profile would create noisy diffs.
  Mitigation: profile has stable identity paths and evidence/confidence fields;
  detailed raw JSON remains a secondary export.
- False visual failures: antialiasing, font availability, fps differences, and
  plugin availability can create expected render deltas. Mitigation: sentinel
  metadata and explicit ignore/gap reasons.
- AI overreach: LLMs may invent unsupported AEP features. Mitigation: recipe
  compiler must be capability-aware and refuse unsupported operations.
- Tool fragmentation: Booyah scripts could remain one-off. Mitigation: Phase 3
  promotes their contracts into reusable render oracle tooling.

## Recommended First Implementation Slice

Start with Phase 1 only:

1. Create `internal/profile`.
2. Move `cmd/aepdissect` JSON model into it.
3. Add stable IDs/paths for project, comp, layer, effect, property.
4. Add enough fields for profile diff to be possible later.
5. Keep the CLI behavior compatible.

Do not start with generation. The generation system depends on profile and diff
being reliable; otherwise AI will only produce plausible-looking but unverified
projects.
