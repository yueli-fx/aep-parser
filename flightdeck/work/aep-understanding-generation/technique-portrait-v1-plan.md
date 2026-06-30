# Technique Portrait v1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a conservative project portrait layer above Technique Facts.

**Architecture:** Keep `internal/technique` as the single package for facts and portraits. `BuildPortrait` consumes only `*technique.FactSet`, so the portrait layer cannot reach around the stable facts contract. Extend `cmd/aeptechnique` with output modes while keeping the existing default facts behavior.

**Tech Stack:** Go, existing `internal/technique`, existing `cmd/aeptechnique`, JSON encoding, TDD.

---

## Files

- Modify: `internal/technique/model.go`
  - Add portrait JSON model types.
- Create: `internal/technique/portrait.go`
  - Implement `BuildPortrait(*FactSet) (*Portrait, error)`.
- Create: `internal/technique/portrait_test.go`
  - Synthetic facts tests for portrait summaries and empty facts.
- Modify: `cmd/aeptechnique/main.go`
  - Add `-mode facts|portrait` and `-portrait`.
- Modify: `cmd/aeptechnique/main_test.go`
  - Add CLI portrait JSON tests.
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
  - Add read-now and progress bullets.
- Modify: `flightdeck/work/aep-understanding-generation/technique-portrait-v1.md`
  - Update implementation progress.

## Task 1: Portrait Model And Builder

**Files:**
- Modify: `internal/technique/model.go`
- Create: `internal/technique/portrait.go`
- Create: `internal/technique/portrait_test.go`

- [x] **Step 1: Write failing portrait summary test**

Add `TestBuildPortraitSummarizesMechanismsGraphSignalsAndHints`. It should
construct a `technique.FactSet` with:

- text, shape, precomp, controller, and solid layers
- one native changed/keyframed/expression effect
- one third-party effect
- text animator facts
- shape operator facts
- source, matte, parent, and effect-param-layer dependencies

Assert fingerprint counts, mechanism counts, graph relation counts, signal
layer labels, and hints:

- `kinetic_text`
- `shape_operator_stack`
- `precomp_assembly`
- `effect_driven_layer`
- `matte_composite`
- `controller_rig`
- `plugin_dependent`

- [x] **Step 2: Verify red**

Run:

```powershell
go test ./internal/technique -run TestBuildPortraitSummarizesMechanismsGraphSignalsAndHints -count=1
```

Expected: fail because `BuildPortrait` is undefined.

- [x] **Step 3: Implement minimal portrait model and builder**

Add model types and implement deterministic aggregation from `FactSet`.

- [x] **Step 4: Verify green**

Run:

```powershell
go test ./internal/technique -run TestBuildPortraitSummarizesMechanismsGraphSignalsAndHints -count=1
```

Expected: pass.

## Task 2: Empty Portrait Behavior

**Files:**
- Modify: `internal/technique/portrait_test.go`
- Modify: `internal/technique/portrait.go`

- [x] **Step 1: Write failing empty-facts test**

Add `TestBuildPortraitAllowsEmptyFactSet` and assert a non-nil portrait with
schema version, source path, zero counts, empty maps, and no hints.

- [x] **Step 2: Verify red**

Run:

```powershell
go test ./internal/technique -run TestBuildPortraitAllowsEmptyFactSet -count=1
```

Expected: fail until empty maps are initialized.

- [x] **Step 3: Implement empty-map initialization**

Ensure all map fields are non-nil even when facts are empty.

- [x] **Step 4: Verify green**

Run:

```powershell
go test ./internal/technique -count=1
```

Expected: pass.

## Task 3: CLI Portrait Mode

**Files:**
- Modify: `cmd/aeptechnique/main.go`
- Modify: `cmd/aeptechnique/main_test.go`

- [x] **Step 1: Write failing CLI portrait tests**

Add:

- `TestRunEmitsPortraitJSONWithMode`
- `TestRunAcceptsPortraitFlag`

Use `flightdeck/showcase/text/text.aep` and assert the output JSON contains
`fingerprint` and `mechanisms`.

- [x] **Step 2: Verify red**

Run:

```powershell
go test ./cmd/aeptechnique -run "TestRunEmitsPortraitJSONWithMode|TestRunAcceptsPortraitFlag" -count=1
```

Expected: fail because `-mode portrait` and `-portrait` are not implemented.

- [x] **Step 3: Implement CLI mode selection**

Add `-mode` with values `facts` and `portrait`; keep existing facts output as
the default. Add `-portrait` as a shorthand that overrides mode to `portrait`.

- [x] **Step 4: Verify green**

Run:

```powershell
go test ./cmd/aeptechnique -count=1
```

Expected: pass.

## Task 4: Documentation, Verification, Commit

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/technique-portrait-v1.md`
- Modify: `flightdeck/cockpit.md`

- [x] **Step 1: Update work docs**

Add portrait v1 to read-now, progress, and cockpit.

- [x] **Step 2: Run verification**

Run:

```powershell
go test ./internal/technique ./cmd/aeptechnique -count=1
go test ./...
go vet ./...
pwsh -NoProfile -File scripts\verify_recipe_profiles.ps1
pwsh -NoProfile -File scripts\regen_fixtures.ps1 -CheckOnly
git diff --check
```

Expected: all pass; fixture check only reports the known manual-only
`re_template.jsx` entry.

- [x] **Step 3: Commit**

Commit with:

```powershell
git add internal\technique cmd\aeptechnique flightdeck\cockpit.md flightdeck\work\aep-understanding-generation
git commit -m "feat: add technique portrait v1"
```

## Self-Review

- Spec coverage: model, builder, CLI, docs, verification, and commit are covered.
- Placeholder scan: no TODO/TBD placeholders.
- Type consistency: all new APIs live under `internal/technique`; CLI keeps the existing facts default.
