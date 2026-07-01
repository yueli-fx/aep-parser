# Versioned AEP Migration Convert Slice Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or equivalent TDD execution. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first conservative `aepmigrate convert` slices: convert no-layer composition skeleton projects, default null-layer projects, default solid-layer projects, default adjustment-layer projects, default camera-layer projects, default light-layer projects, default text-layer projects, default empty shape-layer projects, single rect+fill shape-layer projects, and default precomp-layer projects to a requested AE target version, and refuse projects that would require unsupported layer reconstruction.

**Architecture:** Extend `internal/aepmigrate` with a `Convert` function that opens the source, builds a profile, runs existing assess checks, adds convert-scope blockers, and writes a new target-version project only when the source is inside the first supported surface. After writing, reopen the target, build its profile, and run `profilediff.Compare` before reporting success. Add `cmd/aepmigrate convert` as a CLI wrapper. Do not copy raw chunks across AE versions.

**Tech Stack:** Go standard library, `internal/aep`, `internal/profile`, existing `aepmigrate.Report`, table-driven tests.

---

## Scope

This slice supports:

- source project opens and profiles successfully;
- target is `AE2020`, `AE2022`, or `AE2025`;
- every composition has zero layers, or every layer is inside the first explicit
  layer-bearing slices: default null layers, default solid layers, default
  adjustment layers, default camera layers, default light layers, default text
  layers, default empty shape layers, single rect+fill shape layers, and
  default precomp layers;
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

- any source project with layers outside the default-null/default-solid/default-adjustment/default-camera/default-light/default-text/default-empty-shape/single-rect-fill/default-precomp slice;
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
  - `examples/recipes/minimal-default-null-layer.json` now validates, compiles,
    converts to AE2025 with `profile_diff_status: "pass"` and
    `profile_diff_count: 0`, and passes AE2025 open gate with
    `ae_open_status: "pass"` / `ae_open_exit_code: 0`.
  - Default null-layer conversion preserves the source profile's visible
    default transform surface when the source contains transform properties, so
    recipe-generated default null layers and Go-writer default null layers both
    pass source-vs-target profile diff.
  - Profile and profilediff now include profile-visible footage item details
    (`asset_type`, `width`, `height`, and solid RGB color), so solid source
    dimensions/color are part of the migration verification contract.
  - Default solid-layer sources generated through `aep.NewSolidLayer` and
    `examples/recipes/minimal-layer-source-ref.json` convert to AE2025 with
    `profile_diff_status: "pass"` and `profile_diff_count: 0`; the recipe
    fixture also passes AE2025 open gate with `ae_open_status: "pass"` /
    `ae_open_exit_code: 0`.
  - Default adjustment-layer sources generated through `aep.NewAdjustmentLayer`
    and `examples/recipes/minimal-default-adjustment-layer.json` convert to
    AE2025 with `profile_diff_status: "pass"` and `profile_diff_count: 0`;
    the recipe fixture also passes AE2025 open gate with
    `ae_open_status: "pass"` / `ae_open_exit_code: 0`.
  - Default camera/light-layer sources generated through `aep.NewCameraLayer`
    / `aep.NewLightLayer` and
    `examples/recipes/minimal-default-camera-layer.json` /
    `examples/recipes/minimal-default-light-layer.json` convert to AE2025 with
    `profile_diff_status: "pass"` and `profile_diff_count: 0`; both recipe
    fixtures pass AE2025 open gate with `ae_open_status: "pass"` /
    `ae_open_exit_code: 0`.
  - Default precomp-layer sources generated through `aep.NewPrecompLayer` and
    `examples/recipes/minimal-default-precomp-layer.json` convert to AE2025
    with `profile_diff_status: "pass"` and `profile_diff_count: 0`; the recipe
    fixture passes AE2025 open gate with `ae_open_status: "pass"` /
    `ae_open_exit_code: 0`. This slice creates all target comps before
    materializing layers, then resolves composition source refs into the target
    project instead of copying raw source chunks.
  - Default text-layer sources generated through `aep.NewTextLayer` /
    `Layer.SetText` and `examples/recipes/minimal-default-text-layer.json`
    convert to AE2025 with `profile_diff_status: "pass"` and
    `profile_diff_count: 0`; the recipe fixture passes AE2025 open gate with
    `ae_open_status: "pass"` / `ae_open_exit_code: 0`. This slice only opens
    point-text layers without refs/effects/masks/shapes/markers/comments; text
    document/style fidelity is enforced by the source-vs-target profile diff.
  - Default empty shape-layer sources generated through `aep.NewShapeLayer`
    and `examples/recipes/minimal-default-shape-layer.json` convert to AE2025
    with `profile_diff_status: "pass"` and `profile_diff_count: 0`; the recipe
    fixture passes AE2025 open gate with `ae_open_status: "pass"` /
    `ae_open_exit_code: 0`. This slice intentionally excludes shape content
    nodes; rect/fill/stroke/path/filter migration remains a later surface.
  - Single rect/ellipse graphic/filter shape-layer sources generated through
    `aep.NewShapeLayer` with one parametric primitive plus supported graphic
    nodes and/or supported vector filters convert to AE2025 with
    `profile_diff_status: "pass"` and `profile_diff_count: 0`;
    `minimal-shape-rect-fill-default-transform.json`,
    `minimal-shape-stroke-style.json`, and
    `minimal-shape-stroke-dashes.json`, plus the ellipse
    `minimal-shape-stroke-taper.json`, `minimal-shape-stroke-wave.json`,
    `minimal-shape-round-corners.json`, `minimal-shape-offset-paths.json`,
    `minimal-shape-trim.json`, and `minimal-shape-zigzag.json` pass AE2025
    open gate with
    `ae_open_status: "pass"` / `ae_open_exit_code: 0`. This slice reconstructs
    rect size/position/roundness, ellipse size/position, fill
    color/opacity/blend/composite/fill-rule, and stroke
    color/opacity/width/cap/join/miter/composite/blend plus a single Dash 1/Gap
    1 pair, percent-mode stroke taper controls, wavelength-mode stroke wave
    controls, Trim Paths start/end/offset, Round Corners radius, Offset Paths
    amount/line-join/miter/copies/copy-offset, and ZigZag size/detail/points
    from stable profile properties while still excluding stroke offset,
    additional dash/gap pairs, wave units/cycles, gradient, and other filter
    shape content.
  - Static layer transform reconstruction now maps profile-visible
    `ADBE Anchor Point`, `ADBE Position`, `ADBE Scale`, `ADBE Rotate Z`, and
    `ADBE Opacity` back into `LayerTransform`, converting profile scale/opacity
    unit values into writer percent units. `writeTempProjectWithMovedNullLayer`
    and `examples/recipes/minimal-default-text-static-transform.json` convert
    to AE2025 with `profile_diff_status: "pass"` and
    `profile_diff_count: 0`; the recipe fixture passes AE2025 open gate with
    `ae_open_status: "pass"` / `ae_open_exit_code: 0`. Transform keyframes
    remain outside this static slice and are still blocked by profile diff.
  - Layer timing reconstruction now applies profile-visible `start_time`,
    `in_point`, `out_point`, and `stretch` through `Layer.SetStartTime`,
    `Layer.SetInPoint`, `Layer.SetOutPoint`, and `Layer.SetStretch` for all
    supported layer reconstruction paths. `examples/recipes/minimal-layer-timing.json`
    converts to AE2025 with `profile_diff_status: "pass"` and
    `profile_diff_count: 0`, and passes AE2025 open gate with
    `ae_open_status: "pass"` / `ae_open_exit_code: 0`.

- [x] Update `flightdeck/work/aep-understanding-generation/index.md` and `flightdeck/cockpit.md` to state:
  - `assess` is available;
  - `convert` first slices are available only for no-layer comp skeleton projects,
    default null-layer projects, default solid-layer projects, default
    adjustment-layer projects, default camera-layer projects, default
    light-layer projects, default text-layer projects, default empty
    shape-layer projects, single rect+fill shape-layer projects, and default
    precomp-layer projects;
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
