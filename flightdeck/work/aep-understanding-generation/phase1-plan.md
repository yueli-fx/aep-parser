# AEP Understanding Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extract the `aepdissect -json` profile logic into a reusable `internal/profile` package with schema version, stable paths, evidence records, and admitted `WriteJSON` detail fields.

**Architecture:** `internal/profile` owns the stable normalized model and builder. `cmd/aepdissect` keeps text-report behavior local but delegates `-json` output to `internal/profile`. Phase 1 stops after profile extraction and CLI compatibility; it does not add `profilediff`, `aepdiff`, or render oracle code.

**Tech Stack:** Go 1.25.1, `internal/aep`, `internal/scene` JSON view methods, existing `cmd/aepdissect`, standard `encoding/json`.

---

## File Structure

- Create `internal/profile/profile.go`: public schema types, evidence constants, stable path types, `Build`.
- Create `internal/profile/dict.go`: effect dictionary loading and default-comparison helpers moved from `cmd/aepdissect`.
- Create `internal/profile/profile_test.go`: focused builder tests using existing showcase fixtures.
- Modify `cmd/aepdissect/main.go`: remove private JSON profile structs/build function from CLI and use `profile.Build`.
- Modify `flightdeck/work/aep-understanding-generation/index.md`: mark Phase 1 result and next review point.
- Modify `flightdeck/cockpit.md`: route next step to post-Phase-1 human review before diff tooling.

## Tasks

### Task 1: RED Test for Stable Profile Core

- [x] Add `internal/profile/profile_test.go` with `TestBuildTextFixtureIncludesSchemaPathsEvidenceAndText`.
- [x] Test opens `flightdeck/showcase/text/text.aep`, calls `profile.Build`, and asserts:
  - `SchemaVersion == 1`
  - `Meta.Path` is the fixture path
  - at least one comp and layer are present
  - comp and layer have non-empty machine `path`, `display_path`, and `L1_parsed` evidence
  - text layer exposes `Text.Text`
- [x] Run `go test ./internal/profile`.
- [x] Expected result: fail because `internal/profile` does not exist.

### Task 2: GREEN Implement Minimal Profile Builder

- [x] Create `internal/profile/profile.go`.
- [x] Implement schema types:
  - `Profile`, `Meta`, `Fingerprint`, `Items`, `Item`, `Composition`, `Layer`, `Effect`, `Property`, `Keyframe`, `Mask`, `ShapePath`, `ShapePrimitive`, `TextSource`, `PathRef`, `Evidence`, `Unknown`.
- [x] Implement `Build(project *aep.Project, opts Options) (*Profile, error)`.
- [x] Build from `project.ToJSON()` and direct `project` fields where needed.
- [x] Emit schema version 1, `L1_parsed` evidence, stable comp/layer/effect/property paths, project items, comp settings, layer identity/timing/flags, effects/params, properties/keyframes, masks, shape paths/primitives, text source summaries.
- [x] Run `gofmt` and `go test ./internal/profile`.
- [x] Expected result: pass.

### Task 3: RED/GREEN Effect Dictionary and Fingerprint

- [x] Add `TestBuildEffectsFixtureIncludesEffectUsageAndTunedParams`.
- [x] Test opens `flightdeck/showcase/effects/effects.aep`, loads `data/effects-dict/effects_en_US_25.1x68.json`, builds the profile, and asserts:
  - `Fingerprint.EffectUsage["ADBE Gaussian Blur 2"] == 1`
  - Gaussian Blur effect has dependency class `native`
  - effect parameter `ADBE Gaussian Blur 2-0001` is in `tuned_params`
- [x] Run the targeted test and confirm it fails before dictionary support.
- [x] Create `internal/profile/dict.go` with dictionary loading and default comparison helpers.
- [x] Add `Options.Dict *EffectDictionary` and `LoadEffectDictionary(path string)`.
- [x] Re-run targeted test and `go test ./internal/profile`.

### Task 4: RED/GREEN CLI JSON Delegation

- [x] Add or modify `cmd/aepdissect` tests only if existing structure makes CLI testing practical; otherwise use command-level verification in this task.
- [x] Modify `cmd/aepdissect/main.go` so `-json` calls `profile.Build` with the loaded effect dictionary.
- [x] Keep text output behavior unchanged.
- [x] Run `go run ./cmd/aepdissect -json flightdeck/showcase/effects/effects.aep`.
- [x] Expected JSON includes `schema_version`, `meta`, `fingerprint`, `items`, and `comps`.

### Task 5: Full Verification and Review Stop

- [x] Run `go test ./internal/profile ./cmd/aepdissect`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run reserved-word scan on this work package.
- [x] Run `git diff --check`.
- [x] Update Flightdeck status to say Phase 1 profile extraction is ready for human review before `profilediff/aepdiff`.
- [x] Commit with `feat(profile): add stable project profile`.

## Human Review Gate

Stop after Task 5. Do not start `internal/profilediff`, `cmd/aepdiff`, ignore rules, or render oracle until the Phase 1 profile JSON shape has been reviewed.
