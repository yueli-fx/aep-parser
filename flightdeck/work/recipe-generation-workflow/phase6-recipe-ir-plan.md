# Phase 6 Recipe IR Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the smallest useful recipe IR that can compile an intent-like structured recipe into an AEP through supported writer APIs and produce a report of used capabilities, downgrades, and refusals.

**Architecture:** Add `internal/recipe` as a pure IR/parser/compiler layer that depends on `internal/aep` writer APIs, not raw chunk construction. Add `cmd/aeprecipe` as a thin CLI for `validate`, `compile`, and `explain`. Keep render validation outside the compiler: compiled AEPs are verified by existing `cmd/aeoracle` and `cmd/aepslices` commands.

**Tech Stack:** Go 1.25.1, `encoding/json`, `internal/aep`, `internal/profile`, `cmd/capindex` capability truth source, `cmd/aeoracle` render gate, `cmd/aepslices` diagnose.

---

## Scope

First supported recipe surface:

- project metadata: name/comment only when writer APIs support it,
- one composition,
- solid layers,
- text layers,
- shape rectangle/ellipse layers,
- transform static values: position, scale, rotation, opacity, anchor point,
- transform keyframes only for APIs already verified by current tests,
- effect application only when capability lookup reports supported writer status.

Explicitly unsupported in first slice:

- raw chunk injection,
- arbitrary third-party plugins,
- footage import synthesis,
- nested precomp generation beyond referencing an explicitly created comp,
- automatic visual correction loops,
- LLM free-form JSON accepted without schema validation.

## File Structure

- Create `internal/recipe/schema.go`: recipe structs, enums, validation errors, downgrade/refusal records.
- Create `internal/recipe/schema_test.go`: validation tests for required fields, unsupported layer types, invalid times, and duplicate names.
- Create `internal/recipe/compiler.go`: compiler from recipe to `*aep.Project` using only exported writer APIs.
- Create `internal/recipe/compiler_test.go`: compile tests using profile round-trip checks.
- Create `internal/recipe/capability.go`: capability lookup interface and static test implementation.
- Create `internal/recipe/report.go`: compile report schema with capabilities used, downgrades, refusals, output AEP path, and follow-up commands.
- Create `cmd/aeprecipe/main.go`: CLI with `validate`, `compile`, and `explain`.
- Create `cmd/aeprecipe/main_test.go`: command-level JSON tests and failure exit semantics.
- Create `examples/recipes/minimal-text-shape.json`: one-comp recipe fixture.
- Modify `flightdeck/work/aep-understanding-generation/index.md`: mark recipe IR active when implementation starts.
- Modify `flightdeck/cockpit.md`: route recipe work only after render compare remains green.

## Recipe Schema

```json
{
  "schema_version": 1,
  "project": {
    "name": "Minimal title card"
  },
  "comps": [
    {
      "name": "Main",
      "width": 1920,
      "height": 1080,
      "frame_rate": 30,
      "duration": 4,
      "background_color": [0, 0, 0],
      "layers": [
        {
          "type": "text",
          "name": "Title",
          "text": "BOOYAH",
          "transform": {
            "position": [960, 540],
            "scale": [100, 100],
            "opacity": 100
          }
        },
        {
          "type": "shape",
          "name": "Underline",
          "shape": {
            "kind": "rect",
            "size": [640, 12],
            "fill_color": [255, 255, 255]
          },
          "transform": {
            "position": [960, 650]
          }
        }
      ]
    }
  ]
}
```

Validation rules:

- `schema_version` must equal 1.
- At least one comp is required.
- First slice supports exactly one comp; more comps are rejected with a refusal record.
- Comp width/height/frame_rate/duration must be positive.
- Layer type must be `solid`, `text`, or `shape`.
- Transform arrays must have the exact component count.
- Keyframe times must be within `[0, comp.duration]` and sorted ascending.
- Unsupported effect requests produce refusal records; they do not silently disappear.

## Tasks

### Task 1: RED/GREEN Schema Validation

- [x] Create `internal/recipe/schema.go`.
- [x] Create `internal/recipe/schema_test.go`.
- [x] Add `TestValidateAcceptsMinimalTextShapeRecipe`.
- [x] Add `TestValidateRejectsUnsupportedLayerType`.
- [x] Add `TestValidateRejectsOutOfRangeKeyframes`.
- [x] Implement `Validate(recipe Recipe) Report`.
- [x] Run `go test ./internal/recipe`.

### Task 2: RED/GREEN Capability Gate

- [x] Create `internal/recipe/capability.go`.
- [x] Add `CapabilityIndex` interface:

```go
type CapabilityIndex interface {
    Lookup(query string) CapabilityStatus
}
```

- [x] Add statuses: `supported`, `unsupported`, `unknown`.
- [x] Add `TestValidateRefusesUnsupportedEffects`.
- [x] Make validation produce refusal records for unsupported effect requests.
- [x] Run `go test ./internal/recipe`.

### Task 3: RED/GREEN Compiler

- [x] Create `internal/recipe/compiler.go`.
- [x] Create `internal/recipe/compiler_test.go`.
- [x] Add `TestCompileMinimalTextShapeRecipeBuildsProfile`.
- [x] Compile one comp with a text layer and one shape layer through exported `internal/aep` APIs.
- [x] Write the output AEP to a temp dir.
- [x] Re-open the AEP and build `internal/profile`.
- [x] Assert the profile has one comp, two layers, one text layer, and one shape layer.
- [x] Run `go test ./internal/recipe`.

### Task 4: CLI Validate and Compile

- [x] Create `cmd/aeprecipe/main.go`.
- [x] Implement `validate -recipe file.json [-json]`.
- [x] Implement `compile -recipe file.json -out out.aep [-json]`.
- [x] Implement `explain -recipe file.json [-json]`, which prints used capabilities and refusals without writing an AEP.
- [x] Create `cmd/aeprecipe/main_test.go`.
- [x] Add command tests:
  - valid recipe returns exit 0,
  - unsupported effect returns exit 1 with refusal records,
  - malformed JSON returns exit 2.
- [x] Run `go test ./cmd/aeprecipe`.

### Task 5: Example Recipe and Render Gate

- [x] Create `examples/recipes/minimal-text-shape.json`.
- [x] Compile it:

```powershell
go run ./cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json -out tmp_debug\recipes\minimal-text-shape.aep -json
```

- [x] Plan render frames:

```powershell
go run ./cmd/aeoracle plan -aep tmp_debug\recipes\minimal-text-shape.aep -out tmp_debug\aeoracle\minimal_recipe -json
```

- [x] Render with AE 2025:

```powershell
go run ./cmd/aeoracle render -request tmp_debug\aeoracle\minimal_recipe\request.json -ae 'E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe' -timeout-sec 600
```

- [x] Record generated AEP path, metadata path, PNG outputs, and any warnings in `flightdeck/work/aep-understanding-generation/phase6-recipe-ir-acceptance.md`.

### Task 6: Verification and Commit

- [x] Run `go test ./internal/recipe ./cmd/aeprecipe`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `git diff --check`.
- [x] Run the example compile and render commands from Task 5.
- [x] Update Flightdeck status.
- [x] Commit with `feat(recipe): add minimal AEP recipe compiler`.

## Stop Rule

Stop before automated correction loops. This recipe IR only proves deterministic
generation through supported APIs. A correction loop requires a separate plan
with a bounded parameter set, objective metric, max iterations, and rollback
behavior.
