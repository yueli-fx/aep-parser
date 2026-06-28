# AEP Understanding Phase 4 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn structural/render observations into machine-readable gap ledger entries that can become code, documentation, or knowledge work without reinterpreting freeform notes.

**Architecture:** `internal/gapledger` owns the stable gap schema and mapping from `profilediff.Diff` / `aeoracle.CompareReport` into gaps. `cmd/aepgaps` is a thin CLI with `diff` and `render` subcommands; it reuses `internal/profile`, `internal/profilediff`, and `internal/aeoracle` instead of duplicating comparison logic. Phase 4 does not promote entries into `flightdeck/knowledge/**`; it emits reports to stdout or a caller-provided output path.

**Tech Stack:** Go 1.25.1, `internal/profilediff`, `internal/aeoracle`, standard `encoding/json`.

---

## File Structure

- Create `internal/gapledger/gap.go`: schema types and constructors.
- Create `internal/gapledger/gap_test.go`: mapping tests.
- Create `cmd/aepgaps/main.go`: CLI for `diff` and `render`.
- Modify `flightdeck/work/aep-understanding-generation/index.md`: mark Phase 4 active/result.
- Modify `flightdeck/cockpit.md`: route next step to Phase 4 closeout or Phase 5.

## Tasks

### Task 1: RED/GREEN Gap Mapping From Profile Diffs

- [ ] Add `internal/gapledger/gap_test.go`.
- [ ] Test diff `ActionWrite` maps to `write-gap`.
- [ ] Test diff `ActionSemantics` maps to `semantic-gap`.
- [ ] Test diff `ActionPlugin` maps to `plugin-gap`.
- [ ] Test diff `ActionParse` maps to `parse-gap`.
- [ ] Test diff `ActionInvestigate` maps to `investigate-gap`.
- [ ] Assert each gap has deterministic ID, source project, observed target, evidence, severity, action type, profile path, and human notes.
- [ ] Implement `internal/gapledger/gap.go`.
- [ ] Run `go test ./internal/gapledger`.

### Task 2: RED/GREEN Render Gap Mapping

- [ ] Add tests for `FromRenderCompare`.
- [ ] When `DifferentPixels == 0`, expect no gaps.
- [ ] When `DifferentPixels > 0`, emit one `render-gap` with metric fields in `details`.
- [ ] Include `expected_path`, `actual_path`, total pixels, different pixels, percent, max channel delta, and threshold.
- [ ] Run `go test ./internal/gapledger`.

### Task 3: CLI Diff Gaps

- [ ] Create `cmd/aepgaps/main.go`.
- [ ] Implement `diff`:
  - flags: `-json`, `-out`, `-ignore`, `-dict`, `-source`, `-observed`
  - positional args: `expected.aep actual.aep`
  - builds profiles, runs `profilediff.Compare`, maps unignored diffs to gaps
  - exit 0 when no gaps, exit 1 when gaps exist, exit 2 on usage/errors
- [ ] Text output should show count plus first 50 `id type severity action profile_path`.
- [ ] JSON output should include schema version, source project, observed target, and gaps.
- [ ] Verify with identical fixture and different fixture pair.

### Task 4: CLI Render Gaps

- [ ] Implement `render`:
  - flags: `-json`, `-out`, `-expected`, `-actual`, `-threshold`, `-source`, `-observed`
  - runs `aeoracle.ComparePNG`
  - maps report to gaps
  - exit semantics same as `diff`
- [ ] Verify with generated 2x2 PNGs.

### Task 5: Verification and Commit

- [ ] Run `go test ./internal/gapledger ./cmd/aepgaps`.
- [ ] Run `go test ./...`.
- [ ] Run `go vet ./...`.
- [ ] Run `git diff --check`.
- [ ] Run reserved-word scan on touched files.
- [ ] Run `go run ./cmd/aepgaps diff -json flightdeck/showcase/text/text.aep flightdeck/showcase/text/text.aep`.
- [ ] Run `go run ./cmd/aepgaps diff flightdeck/showcase/text/text.aep flightdeck/showcase/effects/effects.aep` and confirm exit code 1.
- [ ] Run `go run ./cmd/aepgaps render` against generated test PNGs and confirm exit code 1.
- [ ] Update Flightdeck status.
- [ ] Commit with `feat(gapledger): add structured gap reports`.

## Stop Rule

After Phase 4, continue to Phase 5 slice replication workflow planning and implementation unless a real blocker appears. Do not start recipe IR or intent-to-AEP generation before Phase 5 proves a second project can move through profile, diff, render, and gaps.
