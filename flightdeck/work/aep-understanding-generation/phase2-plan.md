# AEP Understanding Phase 2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add deterministic structural/semantic diff tooling over `internal/profile`, so original-vs-clone inspection can report specific profile paths instead of relying on manual ledger review.

**Architecture:** `internal/profilediff` owns the library contract: compare two `profile.Profile` values and return schema-versioned `Report` records. `cmd/aepdiff` is a thin CLI that opens two AEPs, builds profiles, optionally loads ignore rules, and writes either text or JSON output. Phase 2 does not invoke After Effects, render frames, generate gaps, or write clone projects.

**Tech Stack:** Go 1.25.1, `internal/profile`, `internal/aep`, standard `encoding/json`. Ignore files use JSON first to avoid adding a YAML dependency in this slice.

---

## File Structure

- Create `internal/profilediff/diff.go`: public diff record/report types and `Compare`.
- Create `internal/profilediff/ignore.go`: ignore rule schema and matching.
- Create `internal/profilediff/diff_test.go`: table tests over hand-built profiles plus ignore behavior.
- Create `cmd/aepdiff/main.go`: CLI for `aepdiff [-json] [-ignore rules.json] original.aep clone.aep`.
- Create `cmd/aepdiff/main_test.go` only if command behavior needs focused non-AE coverage beyond package tests.
- Modify `flightdeck/work/aep-understanding-generation/index.md`: mark Phase 1 self-review accepted and Phase 2 active.
- Modify `flightdeck/cockpit.md`: update current next step away from human profile review.

## Tasks

### Task 1: RED Tests for Diff Record Contract

- [x] Add `internal/profilediff/diff_test.go`.
- [x] Build two minimal `profile.Profile` values in test helpers.
- [x] Assert `Compare` reports:
  - missing layer as `kind=missing_object`, `action_type=write`, path from expected profile
  - extra layer as `kind=extra_object`, `action_type=investigate`
  - wrong scalar value as `kind=wrong_value`, with expected/actual values
  - effect parameter value differences at parameter paths
  - keyframe count/value differences at property paths
- [x] Assert report schema version is `1`.
- [x] Run `go test ./internal/profilediff`.
- [x] Expected result: fail because package does not exist.

### Task 2: GREEN Minimal Profile Diff

- [x] Create `internal/profilediff/diff.go`.
- [x] Implement:
  - `Report`, `Diff`, `Evidence`, `Options`
  - constrained `kind`, `severity`, and `action_type` string constants
  - `Compare(expected, actual *profile.Profile, opts Options) (*Report, error)`
- [x] Compare high-value fields first:
  - profile schema version and project fingerprint counts
  - compositions by ID with index fallback
  - layers by stable layer path
  - layer type/source/parent/matte/timing/flags
  - effects by match name occurrence
  - effect params/properties by path
  - keyframes by index under property path
  - masks/shapes/text presence and core scalar summaries
- [x] Keep severity deterministic and conservative:
  - missing comp/layer/effect/property/keyframe = `fidelity`
  - wrong visible/source/timing/effect param/text = `fidelity`
  - noisy metadata such as schema/fingerprint mismatch = `unknown` unless tied to a specific object
- [x] Run `go test ./internal/profilediff`.

### Task 3: RED/GREEN Ignore Rules

- [x] Add tests that load/match ignore rules with:
  - `schema_version: 1`
  - `path`
  - `kind`
  - non-empty `reason`
  - optional exact-value condition string for later extension
- [x] Reject ignore rules with empty reason or unsupported schema version.
- [x] Apply ignore rules after generating diffs and before report counts.
- [x] Mark ignored diffs separately only if needed for JSON traceability; text output may omit ignored diffs by default.
- [x] Run `go test ./internal/profilediff`.

### Task 4: CLI

- [x] Add `cmd/aepdiff/main.go`.
- [x] Flags:
  - `-json`: emit full JSON report
  - `-ignore <path>`: load JSON ignore rules
  - `-dict <path>`: optional effect dictionary path, defaulting to current `aepdissect` dictionary behavior
- [x] Positional args: `expected.aep actual.aep`.
- [x] Exit codes:
  - `0`: no unignored diffs
  - `1`: unignored diffs found
  - `2`: usage/open/profile/ignore errors
- [x] Text output should be compact: count summary plus first N diffs with kind/severity/action/path.
- [x] Verify with two identical small fixtures and one intentionally different pair.

### Task 5: Integration Run and Flightdeck Update

- [x] Run `go run ./cmd/aepdiff -json flightdeck/showcase/text/text.aep flightdeck/showcase/text/text.aep`.
- [x] Run an intentionally different fixture pair and confirm non-zero diff exit.
- [x] If Booyah fixture paths are available and command runtime is reasonable, run original vs current clone and store only summary notes under this work package, not generated bulk output.
- [x] Update `index.md` and `cockpit.md` with Phase 2 result.
- [x] Run:
  - `go test ./internal/profilediff ./cmd/aepdiff`
  - `go test ./...`
  - `go vet ./...`
  - reserved-word scan on touched files
  - `git diff --check`
- [x] Commit with `feat(profilediff): add aep structural diff`.

## Verification Notes

- Identical fixture: `go run ./cmd/aepdiff -json flightdeck/showcase/text/text.aep flightdeck/showcase/text/text.aep` returned `diff_count: 0`.
- Different small fixtures: `text.aep` vs `effects.aep` returned exit code 1 with 16 path-level diffs.
- Booyah smoke run: original `samples/motionbox/glitch/booyah-glitch/Booyah Glitch.aep` vs clone `flightdeck/showcase/booyah-clone/booyah-clone.aep` returned exit code 1 with 123 path-level diffs. The first implementation originally over-reported comp-level ID differences; Phase 2 now separates matching identity from report paths and falls back to comp name plus layer index/name for from-scratch clones.

## Stop Rule

Do not start Phase 3 render oracle, gap ledger generation, or recipe/generation work until Phase 2 has a committed diff library and CLI with tests.
