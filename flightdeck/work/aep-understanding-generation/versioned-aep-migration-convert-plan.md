# Versioned AEP Migration Convert Slice Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or equivalent TDD execution. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first conservative `aepmigrate convert` slice: convert only no-layer composition skeleton projects to a requested AE target version, and refuse projects that would require layer reconstruction.

**Architecture:** Extend `internal/aepmigrate` with a `Convert` function that opens the source, builds a profile, runs existing assess checks, adds convert-scope blockers, and writes a new target-version project only when the source is inside the first supported surface. Add `cmd/aepmigrate convert` as a CLI wrapper. Do not copy raw chunks across AE versions.

**Tech Stack:** Go standard library, `internal/aep`, `internal/profile`, existing `aepmigrate.Report`, table-driven tests.

---

## Scope

This slice supports:

- source project opens and profiles successfully;
- target is `AE2020`, `AE2022`, or `AE2025`;
- every composition has zero layers;
- each composition is recreated with name, width, height, frame rate, and duration;
- stable profile-visible composition settings are recreated through target-version
  writers: background color, resolution factor, pixel aspect, display start
  time, work area, frame blending, draft 3D, hide-shy, preserve-nested flags,
  motion-blur settings, label, and comment;
- output project skeleton uses the requested target version.

This slice refuses:

- any source project with layers;
- existing assess blockers such as AE2025 explicit matte downgrade;
- unknown target version labels.

The refusal is intentional. A project with layers should not produce an output
file in this first slice because that would silently drop content.

## Files

- Modify: `internal/aepmigrate/model.go`
  - Add `OutputPath` to options or report target as needed.
- Create: `internal/aepmigrate/convert.go`
  - `Convert(opts ConvertOptions) (Report, error)`.
  - target-label to `aep.AETarget` mapping.
  - no-layer profile gate.
  - target project skeleton rebuild.
- Modify: `internal/aepmigrate/assess_test.go`
  - Reuse fixture helpers.
- Create: `internal/aepmigrate/convert_test.go`
- Modify: `cmd/aepmigrate/main.go`
  - Add `convert` subcommand.
- Modify: `cmd/aepmigrate/main_test.go`
  - Add command tests for success and blocked source.
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/cockpit.md`

## Task 1: Library Convert

- [x] Write a failing test `TestConvertWritesTargetVersionNoLayerProject`.
  - Build an AE2020 source with one no-layer comp.
  - Convert to `AE2025`.
  - Open the output.
  - Assert output AE version normalizes to `AE2025`.
  - Assert output profile has one comp with same name, width, height, frame rate, and duration.

- [x] Write a failing test `TestConvertRefusesLayerProjects`.
  - Build a source with one solid layer.
  - Convert to `AE2025`.
  - Assert error is non-nil or report status is `blocked`.
  - Assert no output `.aep` exists.

- [x] Implement `ConvertOptions`.

```go
type ConvertOptions struct {
	InputPath  string
	OutputPath string
	Target     VersionLabel
}
```

- [x] Implement `Convert`.
  - Open source and build profile.
  - Run source version normalization and `classifyProfile`.
  - Add `ClassBlocked` entries for every layer path in this first convert slice.
  - If summary is blocked, return the report without writing output.
  - Create `aep.NewProject(targetAEPVersion(opts.Target))`.
  - Recreate each no-layer comp via `aep.NewComposition`.
  - Write `opts.OutputPath`.
  - Set `report.Target.Path`.

- [ ] Run:

```powershell
go test ./internal/aepmigrate -run TestConvert -count=1
```

- [x] Add `TestConvertPreservesNoLayerCompSettings`.
  - Build an AE2020 no-layer source with non-default stable comp settings.
  - Convert to `AE2025`.
  - Reopen and profile the output.
  - Assert background color, resolution factor, pixel aspect, display start
    time, work area, comp flags, and motion-blur settings match.

- [x] Add `TestConvertPreservesNoLayerCompMetadata`.
  - Build an AE2020 no-layer source with comp label/comment after reopen.
  - Convert to `AE2025`.
  - Reopen and profile the output.
  - Assert item-level comp label and comment match.

## Task 2: CLI Convert

- [x] Write a failing CLI success test:
  - `run(["convert", "-in", input, "-target", "AE2025", "-out", output, "-report", report])`.
  - Assert exit code `0`, output file exists, report status is `pass`.

- [x] Write a failing CLI blocked test:
  - source has a layer;
  - command returns `1`;
  - report file exists;
  - output file does not exist.

- [x] Add `convert` subcommand:

```powershell
aepmigrate convert -in source.aep -target AE2025 -out migrated.aep -report report.json
```

- [x] Run:

```powershell
go test ./cmd/aepmigrate -count=1
```

## Task 3: Verification And Docs

- [x] Run final verification:

```powershell
go test ./internal/aepmigrate ./cmd/aepmigrate -count=1
go test ./... -count=1
go vet ./...
git diff --check
```

- [x] Run a real command on a generated fixture through tests or temp output.

- [x] Update `flightdeck/work/aep-understanding-generation/index.md` and `flightdeck/cockpit.md` to state:
  - `assess` is available;
  - `convert` first slice is available only for no-layer comp skeleton projects;
  - stable no-layer comp settings are preserved through target-version writers;
  - item-level comp metadata is preserved through the reopen-backed item writer;
  - layer-bearing projects are intentionally blocked until layer reconstruction enters the migration surface.

## Self-Review

- No silent layer drops: layer-bearing projects block before writing.
- No raw chunk copying across target versions.
- No claim of full-project migration.
- Target version is proven by reopening the output and reading the version string.
