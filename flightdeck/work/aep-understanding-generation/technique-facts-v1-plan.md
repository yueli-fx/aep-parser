# Technique Facts v1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first reusable technique-fact extraction layer over `profile.Profile`.

**Architecture:** Add `internal/technique` as a pure derivation package that consumes `internal/profile` data and returns stable JSON-ready facts. Add `cmd/aeptechnique` as a thin CLI around `aep.Open`, `profile.Build`, and `technique.Build`.

**Tech Stack:** Go, existing `internal/aep`, `internal/profile`, JSON encoding, TDD with package tests and CLI tests.

---

## Files

- Create: `internal/technique/model.go`
  - JSON model types and schema version.
- Create: `internal/technique/build.go`
  - `Build(*profile.Profile) (*FactSet, error)` and helper classifiers.
- Create: `internal/technique/build_test.go`
  - synthetic profile tests for comp/layer/effect/text/shape/dependency/unknown facts.
- Create: `cmd/aeptechnique/main.go`
  - CLI entry point and `run` helper.
- Create: `cmd/aeptechnique/main_test.go`
  - JSON CLI smoke test against a small fixture.
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
  - add read-now link and progress bullet.
- Modify: `flightdeck/work/aep-understanding-generation/technique-facts-v1.md`
  - update progress.

## Task 1: Technique Model And Core Build

**Files:**
- Create: `internal/technique/model.go`
- Create: `internal/technique/build.go`
- Create: `internal/technique/build_test.go`

- [x] **Step 1: Write failing core build test**

Add `TestBuildSummarizesCompsLayersEffectsDependenciesAndUnknowns` to
`internal/technique/build_test.go`. It should construct a small
`profile.Profile` with:

- two comps, one with more layers
- text, shape, null, precomp, and solid-like layers
- source, parent, matte, light-source, and effect-param dependencies
- one effect with tuned params, keyframes, expression, layer ref, and unknown
  param
- one top-level profile unknown

- [x] **Step 2: Verify red**

Run:

```powershell
go test ./internal/technique -run TestBuildSummarizesCompsLayersEffectsDependenciesAndUnknowns -count=1
```

Expected: fail because `internal/technique` does not exist.

- [x] **Step 3: Implement model and minimal builder**

Create the model and implement enough build logic to pass the test:

- validate nil profile with an error
- copy source path from profile meta
- summarize counts
- choose main comp candidate by largest layer count
- classify layer roles
- collect dependencies
- summarize effects
- preserve unknowns

- [x] **Step 4: Verify green**

Run:

```powershell
go test ./internal/technique -run TestBuildSummarizesCompsLayersEffectsDependenciesAndUnknowns -count=1
```

Expected: pass.

## Task 2: Text And Shape Fact Classification

**Files:**
- Modify: `internal/technique/build.go`
- Modify: `internal/technique/build_test.go`

- [x] **Step 1: Write failing text/shape classification tests**

Add:

- `TestBuildDetectsTextAnimatorFactsFromProfileProperties`
- `TestBuildDetectsShapeOperatorFactsFromProfileShapesAndProperties`

The tests should use profile property match names that already appear in
recipe/profile tests:

- `ADBE Text Opacity`
- `ADBE Text Position 3D`
- `ADBE Text Fill Color`
- `ADBE Vector Graphic - Stroke`
- `ADBE Vector Filter - Trim`
- `ADBE Vector Filter - Repeater`

- [x] **Step 2: Verify red**

Run:

```powershell
go test ./internal/technique -run "TestBuildDetects(TextAnimator|ShapeOperator)" -count=1
```

Expected: fail because the classifiers are not implemented yet.

- [x] **Step 3: Implement classifiers**

Add deterministic match-name classifiers for v1 text animator and shape
operator families. Unknown match names should be ignored, not emitted as
unsupported facts.

- [x] **Step 4: Verify green**

Run the same focused command. Expected: pass.

## Task 3: CLI

**Files:**
- Create: `cmd/aeptechnique/main.go`
- Create: `cmd/aeptechnique/main_test.go`

- [x] **Step 1: Write failing CLI JSON test**

Test a helper such as:

```go
var buf bytes.Buffer
err := run([]string{"-in", fixture, "-json"}, &buf, io.Discard)
```

Use `flightdeck/showcase/text/text.aep` as the fixture and assert the output is
valid JSON with `schema_version` and at least one layer fact.

- [x] **Step 2: Verify red**

Run:

```powershell
go test ./cmd/aeptechnique -run TestRunEmitsTechniqueJSON -count=1
```

Expected: fail because the command does not exist.

- [x] **Step 3: Implement CLI**

Implement:

- flags: `-in`, `-json`
- require `-in`
- open AEP with `aep.Open`
- build profile with `profile.Build`
- build facts with `technique.Build`
- encode indented JSON to stdout

For v1, `-json` is accepted and defaults to JSON behavior.

- [x] **Step 4: Verify green**

Run the focused CLI test. Expected: pass.

## Task 4: Documentation And Verification

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/technique-facts-v1.md`

- [x] **Step 1: Update work docs**

Add `technique-facts-v1.md` to `Read now`, and add progress that v1 package and
CLI landed.

- [x] **Step 2: Run full verification**

Run:

```powershell
go test ./internal/technique ./cmd/aeptechnique -count=1
go test ./...
go vet ./...
go run ./cmd/aepverify recipe-profiles
pwsh -NoProfile -File scripts\fixtures\regen_fixtures.ps1 -CheckOnly
git diff --check
```

Expected:

- all tests pass
- vet exits 0
- recipe verifier reports 132 passed / 0 failed unless recipe count changes for a deliberate reason
- fixture check has no new unmanaged outputs; the known manual-only `re_template.jsx` entry remains acceptable

- [x] **Step 3: Commit**

Commit with:

```powershell
git add internal\technique cmd\aeptechnique flightdeck\work\aep-understanding-generation\index.md flightdeck\work\aep-understanding-generation\technique-facts-v1.md flightdeck\work\aep-understanding-generation\technique-facts-v1-plan.md
git commit -m "feat: add technique facts v1"
```

## Self-Review

- Spec coverage: tasks cover model, builder, text/shape classifiers, CLI, docs, and verification.
- Placeholder scan: no TODO/TBD placeholders.
- Type consistency: all planned types live under `internal/technique`; CLI consumes `aep.Open`, `profile.Build`, and `technique.Build`.
