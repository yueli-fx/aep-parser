# AEP Understanding Phase 5 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a reusable slice-replication workflow report so a non-Booyah AEP can be planned, classified, diffed, rendered, and converted into gaps without bespoke showcase tooling.

**Architecture:** `internal/sliceworkflow` owns deterministic slice classification, representative comp selection, and a combined workflow report. `cmd/aepslices` is a thin CLI with `plan` and `diagnose` subcommands; it reuses `internal/profile`, `internal/profilediff`, `internal/aeoracle`, and `internal/gapledger`. Phase 5 does not introduce recipe IR or automatic full-project generation.

**Tech Stack:** Go 1.25.1, `internal/profile`, `internal/profilediff`, `internal/aeoracle`, `internal/gapledger`, standard `encoding/json`.

---

## File Structure

- Create `internal/sliceworkflow/report.go`: report schema, comp classification, representative slice selection, and gap report merging.
- Create `internal/sliceworkflow/report_test.go`: table tests over hand-built profiles and render/diff gap inputs.
- Create `cmd/aepslices/main.go`: CLI for `plan` and `diagnose`.
- Create `cmd/aepslices/main_test.go`: command-level render/diff-free tests for exit semantics and JSON output.
- Modify `flightdeck/work/aep-understanding-generation/index.md`: mark Phase 5 active/result.
- Modify `flightdeck/cockpit.md`: route next step to Phase 5 implementation.

## Public Report Contract

`internal/sliceworkflow.Report` should encode:

```go
type Report struct {
    SchemaVersion int
    SourceProject string
    ObservedIn string
    Mode string
    Summary Summary
    Slices []Slice
    GapReports []gapledger.Report
    Commands []Command
}
```

`Slice` should include stable fields only:

```go
type Slice struct {
    CompID uint32
    Name string
    ProfilePath string
    Classification string
    Score int
    Reasons []string
    LayerCount int
    TextLayerCount int
    ShapeLayerCount int
    FootageLayerCount int
    EffectCount int
    PluginDependencyCount int
}
```

Classification values:

- `procedural`: no footage refs and no third-party plugin deps.
- `footage-assembly`: at least one footage layer ref and no third-party plugin deps.
- `plugin-dependent`: at least one third-party effect dependency.
- `mixed`: footage refs plus unsupported or ambiguous signals that do not fit the first three categories.

## Tasks

### Task 1: RED/GREEN Slice Classification

- [x] Add `internal/sliceworkflow/report_test.go`.
- [x] Write `TestBuildReportClassifiesProceduralComp` with a hand-built profile containing one text layer and one shape layer, no footage, no third-party effects.
- [x] Run `go test ./internal/sliceworkflow` and confirm it fails because the package does not exist.
- [x] Create `internal/sliceworkflow/report.go`.
- [x] Implement `BuildReport(prof *profile.Profile, opts Options) (Report, error)`.
- [x] Implement classification helpers:
  - count text layers with `Layer.Text != nil`
  - count shape layers with `len(Layer.Shapes) > 0`
  - count footage layers where `Layer.SourceRef.Kind == "footage"`
  - count effects from `Layer.Effects`
  - count plugin deps where `Effect.DependencyClass == "third_party"`
- [x] A procedural comp should include reasons `text-layer` and `shape-layer` when those signals exist.
- [x] Run `go test ./internal/sliceworkflow` and confirm it passes.

### Task 2: RED/GREEN Representative Selection

- [x] Add `TestBuildReportSelectsDeterministicRepresentativeSlices`.
- [x] Build a profile with four comps: procedural, footage-assembly, plugin-dependent, and another procedural.
- [x] Set `Options{MaxSlices: 3}`.
- [x] Assert selected slices are sorted by descending score, then comp ID, then name.
- [x] Assert plugin-dependent comps score higher than footage-assembly, and footage-assembly scores higher than simple procedural comps.
- [x] Implement `scoreSlice`.
- [x] Implement max-slice capping; `MaxSlices <= 0` should default to 3.
- [x] Run `go test ./internal/sliceworkflow`.

### Task 3: RED/GREEN Gap Report Coupling

- [x] Add `TestBuildDiagnoseReportIncludesDiffAndRenderGaps`.
- [x] Build one `profilediff.Report` with a write diff and one `aeoracle.CompareReport` with `DifferentPixels > 0`.
- [x] Implement `BuildDiagnoseReport(source *profile.Profile, observed *profile.Profile, diffReport *profilediff.Report, renderReports []aeoracle.CompareReport, opts Options) (Report, error)`.
- [x] Convert the diff report with `gapledger.FromDiffReport`.
- [x] Convert each render compare report with `gapledger.FromRenderCompare`.
- [x] Omit empty gap reports.
- [x] Set `Mode` to `diagnose`; `BuildReport` should set `Mode` to `plan`.
- [x] Run `go test ./internal/sliceworkflow`.

### Task 4: CLI Plan

- [x] Create `cmd/aepslices/main.go`.
- [x] Implement `plan` flags:
  - `-aep`
  - `-max`
  - `-dict`
  - `-json`
  - `-out`
- [x] Open the AEP through `internal/aep`, build a profile, call `sliceworkflow.BuildReport`.
- [x] Exit 0 on success, exit 2 on usage/errors.
- [x] Text output should show project path, selected slice count, and one line per slice: `classification score comp_id name`.
- [x] JSON output should include schema version, mode, summary, slices, commands, and no gaps.
- [x] Add `cmd/aepslices/main_test.go` for JSON emit and usage failure using command-level fixture tests.
- [x] Verify with `go run ./cmd/aepslices plan -aep flightdeck/showcase/text/text.aep -json`.

### Task 5: CLI Diagnose

- [x] Implement `diagnose` flags:
  - `-expected`
  - `-actual`
  - `-max`
  - `-dict`
  - `-ignore`
  - `-expected-png`
  - `-actual-png`
  - `-threshold`
  - `-json`
  - `-out`
- [x] Build expected and actual profiles.
- [x] Run `profilediff.Compare`.
- [x] If both PNG flags are provided, run `aeoracle.ComparePNG`.
- [x] Call `sliceworkflow.BuildDiagnoseReport`.
- [x] Exit 0 when total gap count is zero, exit 1 when any gap exists, exit 2 on usage/errors.
- [x] Verify identical AEP pair returns 0 and no gaps.
- [x] Verify different fixture pair returns 1 and gap reports.
- [x] Verify optional PNG pair returns render gap.

### Task 6: Verification and Commit

- [x] Run `go test ./internal/sliceworkflow ./cmd/aepslices`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `git diff --check`.
- [x] Run reserved-word scan on touched files.
- [x] Run `go run ./cmd/aepslices plan -aep flightdeck/showcase/text/text.aep -json`.
- [x] Run `go run ./cmd/aepslices diagnose -expected flightdeck/showcase/text/text.aep -actual flightdeck/showcase/text/text.aep -json`.
- [x] Run `go run ./cmd/aepslices diagnose -expected flightdeck/showcase/text/text.aep -actual flightdeck/showcase/effects/effects.aep`.
- [x] Run `go run ./cmd/aepslices diagnose` with generated or existing 2x2 PNGs and confirm exit code 1.
- [x] Update Flightdeck status.
- [x] Commit with `feat(sliceworkflow): add replication slice reports`.

## Verification Notes

- `plan` on `flightdeck/showcase/text/text.aep` selected `TextShowcase` as one procedural slice with text and shape signals.
- `diagnose` on the same AEP returned exit 0 and no gaps.
- `diagnose` from text fixture to effects fixture returned exit 1 with 16 structural gaps.
- `diagnose` on identical AEPs plus existing 2x2 PNGs returned exit 1 with one render gap.
- This validates the generic workflow mechanics on tracked fixtures. A true external, non-Booyah project still needs to run through the same command before Phase 6 recipe IR should be treated as ready.

## Stop Rule

After Phase 5, stop before recipe IR unless the workflow has produced a second-project report with profile, slice selection, diff gaps, and optional render gaps. If the report shows profile or writer gaps that block meaningful slice generation, fix those first instead of starting Phase 6.
