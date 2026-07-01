# Versioned AEP Migration Convert Slice Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or equivalent TDD execution. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first conservative `aepmigrate convert` slices: convert no-layer composition skeleton projects and default null-layer projects to a requested AE target version, and refuse projects that would require unsupported layer reconstruction.

**Architecture:** Extend `internal/aepmigrate` with a `Convert` function that opens the source, builds a profile, runs existing assess checks, adds convert-scope blockers, and writes a new target-version project only when the source is inside the first supported surface. After writing, reopen the target, build its profile, and run `profilediff.Compare` before reporting success. Add `cmd/aepmigrate convert` as a CLI wrapper. Do not copy raw chunks across AE versions.

**Tech Stack:** Go standard library, `internal/aep`, `internal/profile`, existing `aepmigrate.Report`, table-driven tests.

---

## Scope

This slice supports:

- source project opens and profiles successfully;
- target is `AE2020`, `AE2022`, or `AE2025`;
- every composition has zero layers, or every layer is inside the first explicit
  layer-bearing slice: default null layers only;
- each composition is recreated with name, width, height, frame rate, and duration;
- stable profile-visible composition settings are recreated through target-version
  writers: background color, resolution factor, pixel aspect, display start
  time, work area, frame blending, draft 3D, hide-shy, preserve-nested flags,
  motion-blur settings, renderer, Motion Graphics template name, label, and
  comment;
- output project skeleton uses the requested target version.
- every successful output is reopened and profile-diffed against the source;
  unexpected profile diffs are recorded in `verification.profile_diffs` and
  block success.

This slice refuses:

- any source project with layers outside the default-null slice;
- existing assess blockers such as AE2025 explicit matte downgrade;
- unknown target version labels.

The refusal is intentional. A project with unsupported layers should not produce
an output file because that would silently drop content.

## Files

- Modify: `internal/aepmigrate/model.go`
  - Add `OutputPath` to options or report target as needed.
- Create: `internal/aepmigrate/convert.go`
  - `Convert(opts ConvertOptions) (Report, error)`.
  - target-label to `aep.AETarget` mapping.
  - no-layer/default-null profile gate.
  - target project skeleton rebuild.
- Modify: `internal/aepmigrate/assess_test.go`
  - Reuse fixture helpers.
- Create: `internal/aepmigrate/convert_test.go`
- Modify: `cmd/aepmigrate/main.go`
  - Add `convert` subcommand.
- Modify: `cmd/aepmigrate/main_test.go`
  - Add command tests for success and blocked source.
- Modify: `internal/profilediff/diff.go`
  - Compare profile-visible comp settings that migration claims to preserve.
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

- [x] Run:

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

- [x] Add `TestConvertPreservesNoLayerCompRendererAndTemplateName`.
  - Build an AE2020 no-layer source with non-default renderer and Motion
    Graphics template name.
  - Convert to `AE2025`.
  - Reopen and profile the output.
  - Assert renderer match name and template name match.

## Task 1.5: Profile Diff Gate

- [x] Add `profilediff.TestCompareReportsCompSettingsDiffs`.
  - Prove profile diff reports background color, resolution factor, pixel
    aspect, display start time, work area, comp flags, motion blur, renderer,
    Motion Graphics template name, label, and comment differences.

- [x] Extend `profilediff.compareComp`.
  - Compare all profile-visible comp fields that the no-layer convert slice
    claims to preserve.

- [x] Add `TestConvertRunsProfileDiffVerification`.
  - Convert a no-layer comp with non-default settings.
  - Assert `report.verification.profile_diff_status == "pass"`.
  - Assert `report.verification.profile_diff_count == 0`.

- [x] Add CLI report coverage.
  - `cmd/aepmigrate convert` report JSON must expose profile diff status/count.

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
  - `examples/recipes/minimal-comp-object-profile.json` compiled to
    `tmp\migration_verify_gate_smoke\source.aep`.
  - `cmd/aepmigrate convert` retargeted it to AE2025 with
    `profile_diff_status: "pass"` and `profile_diff_count: 0`.
  - AE 2025 open gate passed via `test_data/generators/verify_open.jsx` and
    `scripts/ae-worker/ae_run.ps1`; readback saw 1 composition named `Main`
    and 0 layers.
  - `cmd/aepmigrate convert -ae-open -ae <AfterFX.exe>` now runs the same gate
    directly and writes `ae_open_status: "pass"` into the migration report.
  - Default null-layer source generated through `aep.NewNullLayer` converted to
    AE2025 with `profile_diff_status: "pass"`, `profile_diff_count: 0`,
    `ae_open_status: "pass"`, and `ae_open_exit_code: 0`.
  - `examples/recipes/minimal-default-null-layer.json` validates and compiles,
    but its explicit transform properties are outside the default-null convert
    slice; `convert` blocks it via pre-write profile diff and leaves no target
    output.

- [x] Update `flightdeck/work/aep-understanding-generation/index.md` and `flightdeck/cockpit.md` to state:
  - `assess` is available;
  - `convert` first slices are available only for no-layer comp skeleton projects
    and default null-layer projects;
  - stable no-layer comp settings are preserved through target-version writers;
  - renderer and Motion Graphics template name are preserved through their comp
    writers;
  - item-level comp metadata is preserved through the reopen-backed item writer;
  - successful convert runs source-vs-target profile diff and blocks success on
    unexpected differences;
  - unsupported layer-bearing projects are intentionally blocked until their
    reconstruction enters the migration surface.

## Self-Review

- No silent layer drops: unsupported layer-bearing projects block before writing.
- No raw chunk copying across target versions.
- No claim of full-project migration.
- Target version is proven by reopening the output and reading the version string.
- Successful convert is now also proven by source-vs-target profile diff status
  `pass`; AE open/render gates remain a separate, stronger validation layer.
- First AE 2025 open smoke passed for the no-layer comp object profile
  migration output.
