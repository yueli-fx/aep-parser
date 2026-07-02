# AEP Understanding Phase 0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce the Phase 0 export-surface audit and contract lock so Phase 1 can extract `internal/profile` without guessing the schema.

**Architecture:** This phase is documentation and audit-first. It reads current code and existing fixtures, compares `cmd/aepdissect -json` with `internal/scene.Project.WriteJSON`, then records the profile coverage matrix, field admission rules, stable path contract, and Phase 1 entry gate under this work package.

**Tech Stack:** Go 1.25.1, existing `cmd/aepdissect`, existing `internal/aep.Open`, existing `internal/scene.WriteJSON`, Flightdeck work documents.

---

### Task 1: Confirm Export Surfaces and Fixture Set

**Files:**
- Read: `cmd/aepdissect/main.go`
- Read: `internal/scene/write_json.go`
- Read: `internal/aep/facade.go`
- Modify: `flightdeck/work/aep-understanding-generation/phase0-audit.md`

- [x] **Step 1: Record the two export surfaces**

Read these code locations:

```text
cmd/aepdissect/main.go:
- `jProfile`, `jComp`, `jLayer`, `jEffect`, `jParam`, `jAnim`
- `buildProfile`
- `-json` flag handling

internal/scene/write_json.go:
- `JSONProject`
- `JSONComposition`
- `JSONLayer`
- `JSONEffect`
- `JSONProperty`
- `JSONTextSource`
- `JSONMask`
- `JSONShapePath`
- `JSONShapePrimitive`
- `JSONRenderQueue`
- `Project.ToJSON`
- `Project.WriteJSON`
```

Add a `## Export surfaces` section to `phase0-audit.md` with this conclusion:

```markdown
## Export surfaces

- `cmd/aepdissect -json` is a diagnostic/understanding profile embedded inside the CLI. It exposes a compact project summary: project path, effect usage, third-party effect list, comps, layers, effect params, tuned params, animated transform/effect summaries, expression references, and precomp/footage source labels.
- `Project.WriteJSON` is a broad parsed-scene export. It exposes project items, render queue, comp settings, layer identity/flags/timing/source IDs, properties/keyframes/ease, effects, masks, shape paths/primitives, text source/style runs, markers, guides, folders, and footage.
- Phase 1 must not copy either surface wholesale. `Profile` should be a normalized comparison/learning layer with evidence and stable paths; `WriteJSON` remains the detailed export.
```

- [x] **Step 2: Record the fixture set**

Use these fixtures for Phase 0:

```text
Booyah stress fixture:
- flightdeck/showcase/booyah-clone/booyah-clone.aep

Small deterministic fixtures:
- flightdeck/showcase/text/text.aep
- flightdeck/showcase/masks/masks.aep
- flightdeck/showcase/keyframes-ease/keyframes_ease.aep
- flightdeck/showcase/effects/effects.aep
- internal/serializer/templates/project/2025.aep
```

Add a `## Fixture set` section to `phase0-audit.md` explaining why each fixture is used:

```markdown
## Fixture set

- `booyah-clone.aep`: current real-project stress fixture; useful for nested comps, effects, motion, and authored visual complexity.
- `text.aep`: focused text-source/style coverage.
- `masks.aep`: focused mask/path coverage.
- `keyframes_ease.aep`: focused keyframe/interpolation/ease coverage.
- `effects.aep`: focused effect/parameter coverage.
- `2025.aep`: tiny skeleton project for baseline item/project shape.
```

### Task 2: Run Existing CLI Audit Commands

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/phase0-audit.md`

- [x] **Step 1: Run `aepdissect -json` against representative fixtures**

Run:

```powershell
go run ./cmd/aepdissect -json flightdeck/showcase/text/text.aep
go run ./cmd/aepdissect -json flightdeck/showcase/masks/masks.aep
go run ./cmd/aepdissect -json flightdeck/showcase/keyframes-ease/keyframes_ease.aep
go run ./cmd/aepdissect -json flightdeck/showcase/effects/effects.aep
go run ./cmd/aepdissect -json flightdeck/showcase/booyah-clone/booyah-clone.aep
```

Expected:

```text
Each command exits 0 and emits JSON with `project`, `effectUsage`, and `comps`.
```

- [x] **Step 2: Summarize the observed `aepdissect` shape**

Add this section and adjust counts only if command output contradicts it:

```markdown
## Observed aepdissect-json shape

`aepdissect -json` is compact and comparison-friendly but shallow:

- Has project path, effect usage histogram, third-party effect list, comp ID/name/size/fps/duration, layer index/name/type/blend/visible/timing/source label, parent/matte index, effects, effect params, changed/tuned hints, animated summaries, and expression refs.
- Does not expose project folders, full footage metadata, render queue, comp renderer/work area/motion blur settings, layer IDs, many layer flags, raw source IDs, full property trees, full keyframe values/ease arrays, masks, shapes, text source/style runs, markers, guides, essential graphics, or unknown/raw escape hatches.
- Uses human display labels in several places and parent/matte by layer index, so it is not yet a stable diff path contract.
```

### Task 3: Inspect `WriteJSON` Coverage

**Files:**
- Read: `internal/scene/write_json.go`
- Modify: `flightdeck/work/aep-understanding-generation/phase0-audit.md`

- [x] **Step 1: Record the broad WriteJSON shape**

Add this section:

```markdown
## Observed WriteJSON shape

`Project.WriteJSON` is broad and deterministic, but it is a detailed parsed export rather than a curated profile:

- Project-level: compositions, footage, folders, render queue.
- Composition-level: ID, name, dimensions, frame rate, duration, tick rate, background color, resolution factor, renderer, work area, motion blur/shutter settings, layers, markers, guides, motion graphics template name, essential graphics controllers.
- Layer-level: index, name, type, ID, parent ID/name, source ID, timing, stretch, quality, label, blending mode, track matte, comments, 3D/solo/shy/locked/visible/adjustment/null/guide/motion blur/effects/audio/frame blend/collapse/shape flags.
- Content-level: properties, effects and their parameters, markers, masks, shape paths, shape primitives, text source/style runs/paragraphs.
- Property-level: name, match name, static value, expression, expression enabled state, keyframes, interpolation, temporal ease, and spatial tangents where surfaced.
- Render-level: queue items, output modules, render settings, output module settings, and format options.
```

- [x] **Step 2: Record WriteJSON limits**

Add this section:

```markdown
## WriteJSON limits for profile use

- It is one-way export; there is no corresponding `ReadJSON`.
- It is intentionally broad, so direct diffs would be noisy.
- It lacks evidence levels, schema version, stable path objects, field admission metadata, and explicit unknown classifications.
- Some values are display-friendly rather than profile contracts, such as names beside IDs.
- Large projects can produce bulky output; Phase 1 profile should summarize repeated/noisy fields unless detailed values are needed for diff or generation.
```

### Task 4: Produce Coverage Matrix and Field Admission

**Files:**
- Create: `flightdeck/work/aep-understanding-generation/profile-coverage.md`

- [x] **Step 1: Create the coverage matrix**

Create `profile-coverage.md` with this table:

```markdown
# Profile Coverage Matrix

| Field group | aepdissect-json | WriteJSON | Profile core decision | Evidence floor | Notes |
| --- | --- | --- | --- | --- | --- |
| Project identity | `project` path | absent as explicit path | Admit `meta.path`, `meta.schema_version`, `meta.parse_warnings` | L1_parsed | Add schema version in Phase 1. |
| Project item graph | precomp/footage source labels only | `compositions`, `footage`, `folders` with IDs | Admit normalized item table with IDs, names, type, parent/folder where available | L1_parsed | Needed for source refs and generation inventory. |
| Effect usage fingerprint | `effectUsage`, `thirdParty` | effects nested per layer | Admit summarized effect usage and plugin dependency list | L1_parsed | Keep detailed params per path under effects. |
| Composition settings | ID/name/size/fps/duration | plus tick rate, renderer, work area, bg, shutter/motion blur, markers/guides/EG | Admit ID/name/size/fps/duration/tick/render/work area/motion blur; detail-only markers/guides/EG until diff needs them | L1_parsed | Renderer/work area affect render and slice selection. |
| Layer identity | index/name/type | index/name/type/ID/source ID/parent ID | Admit ID, index, name, type, source ref, parent ref, matte ref | L1_parsed | Stable path uses ID where present plus index/occurrence fallback. |
| Layer timing | in/out points | start/duration/stretch | Admit normalized in/out/start/duration/stretch | L1_parsed | Required for active intervals and render sentinel selection. |
| Layer flags | visible/blend only | broad flags and switches | Admit visible, blend, 3D, solo, shy, locked, adjustment/null/guide, motion blur, effects enabled, audio, frame blend, collapse | L1_parsed | Flags frequently explain render deltas. |
| Source refs | precomp/footage display label | raw source ID plus footage table | Admit source ID and resolved display label | L1_parsed | Display label is annotation, not identity. |
| Effects | match name/class/name/tuned/params | match name/name/parameters | Admit match name, display name, occurrence, dependency class, params, tuned params, unknown params | L1_parsed | `tuned_params` remains hint, not proof of semantic importance. |
| Effect params | static value/default/changed/expression | full property shape through parameters | Admit static value, expression, keyframes, interpolation/ease when present | L1_parsed | Defaults from dict are advisory metadata. |
| Transform/property tree | animated summary only | full `properties` list | Admit stable property paths, static values, keyframes, interpolation/ease, expression | L1_parsed | Summaries can remain fingerprint fields. |
| Keyframes | count/motion/start/end | values/interp/ease/tangents | Admit full keyframe timeline for admitted properties | L1_parsed | Required before diff and sentinel frames. |
| Expressions | expression refs from effect params | expression + enabled state per property | Admit expression text, enabled state, and parsed refs where available | L1_parsed | Expression reference extraction is heuristic unless parser-backed. |
| Masks | absent | masks, vertices, path keyframes, interp/ease | Admit mask metadata and path timeline; mark geometry evidence carefully | L1_parsed | Existing knowledge says some mask geometry may be partial. |
| Shapes | absent | shape paths and primitives | Admit shape paths/primitives for shape-layer diff | L1_parsed | Avoid full graph overreach until path identity is stable. |
| Text | absent | text source, fonts, runs, paragraphs, justification | Admit text source summary and style runs | L1_parsed | Font availability becomes render oracle metadata later. |
| Markers/guides/EG | absent | present | Detail-only for Phase 1 unless fixture proves clone/diff need | L1_parsed | Can be admitted later without changing core identity model. |
| Render queue | absent | present | Detail-only for Phase 1; profile fingerprint may include presence/count | L1_parsed | Generation/replication first targets comps, not render queue. |
| Unknown/raw escapes | absent | limited by scene model | Admit explicit `unknowns` records where parser exposes uncertainty | L0_raw/L1_parsed | Do not silently omit unsupported structures. |
```

- [x] **Step 2: Add field admission rules**

Append:

```markdown
## Phase 1 field admission list

Admit immediately:

- `schema_version`
- `meta.path`
- `meta.parse_warnings`
- project item table: comps, footage, folders
- comp ID/name/size/fps/duration/tick rate/renderer/work area/motion blur settings
- layer ID/index/name/type/source ref/timing/stretch/core flags/blend/parent/matte
- effect match name/display name/dependency class/occurrence/params/tuned params/unknown params
- property match name/path/static value/keyframes/interpolation/ease/spatial tangents/expression
- masks with metadata/path timeline
- shapes with path/primitive summaries
- text source and style run summaries
- evidence records and unknown records

Keep detail-only for Phase 1:

- render queue full settings
- guides
- essential graphics controllers
- complete marker payloads
- detailed output module format options
```

### Task 5: Lock Stable Path and Evidence Contracts

**Files:**
- Create: `flightdeck/work/aep-understanding-generation/profile-contract.md`

- [x] **Step 1: Create stable path contract**

Create `profile-contract.md` with:

```markdown
# Profile Contract

## Evidence levels

| Level | Meaning | Use in Phase 1 |
| --- | --- | --- |
| `L0_raw` | Bytes/chunks exist but are not structurally decoded. | Unknown/raw records only. |
| `L1_parsed` | Parser exposes deterministic structured fields. | Default for parsed profile fields. |
| `L2_roundtrip` | Read/write/read preserves the field. | Capability mapping, not required for all read fields. |
| `L3_ae_accept` | AE opens or renders the written construct. | Writer capability confidence. |
| `L4_render` | Selected rendered pixels match tolerance. | Render oracle output. |
| `L5_user_verified` | User/human accepts target result. | Showcase acceptance and final promotion. |

## Stable path object

Profile records that can be diffed must carry:

- `path`: machine path using IDs, indexes, match names, and occurrence indexes.
- `display_path`: human-readable path with names.
- `identity`: structured identity fields.
- `evidence`: strongest evidence for this record.

Example:

```yaml
path: comps.by_id[17].layers.by_id[42].effects.by_match_name["ADBE Fill"]#0.params.by_match_name["ADBE Fill-0002"]
display_path: comps["main"].layers[3:"white flash"].effects[0:"Fill"].params["Color"]
identity:
  comp_id: 17
  layer_id: 42
  layer_index: 3
  effect_match_name: ADBE Fill
  effect_occurrence: 0
  property_match_name: ADBE Fill-0002
evidence:
  level: L1_parsed
  source: internal/scene
  confidence: high
```

## Path rules

- Use comp IDs when present; use `comps.by_index[n]` only for ID-less records.
- Use layer IDs when present; keep layer index as annotation and fallback.
- Use effect `matchName` plus zero-based occurrence, because duplicate effects are legal.
- Use property `matchName` plus occurrence or property-group lineage for duplicate properties.
- Keep display names out of machine identity unless there is no other identity source.
- Emit `unstable_path` rather than creating a text-only path when identity is insufficient.
- Ignore rules must target `path`, `kind`, and optional `condition`, never only display text.
```

- [x] **Step 2: Add Phase 1 entry gate**

Append:

```markdown
## Phase 1 entry gate

Human review is required before extracting `internal/profile`.

Phase 1 may start only when reviewers accept:

- `phase0-audit.md`
- `profile-coverage.md`
- this profile contract
- the decision that `Profile` is a normalized layer, while `WriteJSON` remains the detailed export
- the decision that Phase 1 includes stable paths and evidence records before diff tooling
```

### Task 6: Sync Flightdeck Index and Stop for Review

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/cockpit.md`

- [x] **Step 1: Update work index**

Update the work package index so `Read now` includes:

```markdown
- `phase0-audit.md` — current export-surface audit.
- `profile-coverage.md` — coverage matrix and Phase 1 field admission list.
- `profile-contract.md` — evidence ladder, stable path object, and Phase 1 entry gate.
- `plan.md` — Phase 0 execution plan.
```

Set `Current` to:

```markdown
Current:
- Phase 0 audit artifacts are ready for human review.
- Stop here before extracting `internal/profile`.
```

- [x] **Step 2: Update cockpit next step**

Change the `aep-understanding-generation` next item to say:

```markdown
Phase 0 人工审核点：确认 `phase0-audit.md` / `profile-coverage.md` / `profile-contract.md` 后，再进入 Phase 1 `internal/profile` 抽取。
```

- [x] **Step 3: Verify and commit**

Run:

```powershell
$p = ('TB'+'D','TO'+'DO','place'+'holder','待'+'定','以后'+'再说','不确'+'定','\?\?\?') -join '|'
Select-String -Path flightdeck/work/aep-understanding-generation/*.md -Pattern $p
git diff --check
git status --short
```

Expected:

```text
The reserved-word scan prints no matches.
`git diff --check` exits 0.
`git status --short` shows only the intended Flightdeck work files before commit.
```

Commit:

```powershell
git add flightdeck/work/aep-understanding-generation flightdeck/cockpit.md
git commit -m "docs(flightdeck): add phase0 profile audit"
```
