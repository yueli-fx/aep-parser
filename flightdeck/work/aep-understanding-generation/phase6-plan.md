# AEP Understanding Phase 6 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first generation-readiness loop: reuse one sentinel-frame plan for an original AEP and a generated/clone AEP, render both, compare all matching frames, and feed render differences into structured gaps before introducing recipe IR.

**Architecture:** Keep render comparison in `internal/aeoracle` and `cmd/aeoracle`; keep slice/gap aggregation in `internal/sliceworkflow` and `cmd/aepslices`. Phase 6 starts with measured clone fidelity, not automatic correction. Recipe IR work starts only after a source-vs-clone render compare can be produced from generic commands.

**Tech Stack:** Go 1.25.1, `internal/aeoracle`, `internal/gapledger`, `internal/sliceworkflow`, `cmd/aeoracle`, `cmd/aepslices`, `scripts/aeoracle_render.jsx`, existing AE runner `scripts/ae_run.ps1`.

---

## File Structure

- Create `internal/aeoracle/frameset.go`: frame-set metadata loading, tag alignment, multi-frame PNG comparison, and summary schema.
- Create `internal/aeoracle/frameset_test.go`: table tests for exact matches, pixel differences, missing frame tags, duplicate tags, and status propagation.
- Modify `internal/aeoracle/frames.go`: add a helper that clones an existing `RenderRequest` for another AEP/output directory while preserving `Frames`.
- Modify `internal/aeoracle/request_test.go`: cover cloned requests and path fields.
- Modify `cmd/aeoracle/main.go`: add `clone-request` and `compare-set` subcommands.
- Modify `cmd/aeoracle/main_test.go`: command-level tests for `clone-request`, `compare-set`, and render status failure semantics.
- Modify `internal/gapledger/render.go`: accept frame-set compare reports or expose a converter that emits one render gap per differing or missing frame.
- Modify `internal/sliceworkflow/report.go`: allow a frame-set compare report in diagnose mode.
- Modify `cmd/aepslices/main.go`: add `-render-set` to diagnose, reading a frame-set report and merging its gaps.
- Modify `flightdeck/work/aep-understanding-generation/index.md`: record Phase 5 hard render completion and Phase 6 status.
- Modify `flightdeck/cockpit.md`: route next work to Phase 6 render-compare loop.

## Public Contracts

### Clone request

`aeoracle clone-request` creates a second render request with the same frame tags/times as an existing source request:

```powershell
go run ./cmd/aeoracle clone-request `
  -from tmp_debug\aeoracle\booyah\request.json `
  -aep flightdeck\showcase\booyah-clone\booyah-clone.aep `
  -out tmp_debug\aeoracle\booyah_clone `
  -json
```

Rules:
- Preserve `frames` exactly from the source request.
- Default `comp_name` to the source request's `comp_name`.
- Allow `-comp` override for clones whose top comp uses a different name.
- Write `request.json`, `aeoracle_render.done`, and `metadata.json` under the requested output directory.
- Do not mutate the source request file.

### Frame-set compare

`aeoracle compare-set` compares every matching frame tag from two metadata files:

```powershell
go run ./cmd/aeoracle compare-set `
  -expected-meta tmp_debug\aeoracle\booyah\metadata.json `
  -actual-meta tmp_debug\aeoracle\booyah_clone\metadata.json `
  -threshold 0 `
  -json `
  -out tmp_debug\aeoracle\booyah_compare_set.json
```

Report shape:

```go
type FrameSetCompareReport struct {
    SchemaVersion int
    ExpectedMetadata string
    ActualMetadata string
    ExpectedAEPPath string
    ActualAEPPath string
    CompName string
    Summary FrameSetSummary
    Frames []FrameCompareRecord
}

type FrameCompareRecord struct {
    Tag string
    Frame int
    Seconds float64
    Reason string
    ExpectedPath string
    ActualPath string
    Status string // ok, different, missing_expected, missing_actual, error
    Compare *CompareReport
    Error string
}
```

Exit semantics:
- Exit 0 only when every frame status is `ok`.
- Exit 1 when frames compare but any frame differs or is missing.
- Exit 2 on usage, malformed metadata, unreadable PNG, duplicate tag, or unsupported schema.

## Tasks

### Task 1: RED/GREEN Clone Render Requests

**Files:**
- Modify: `internal/aeoracle/frames.go`
- Modify: `internal/aeoracle/request_test.go`
- Modify: `cmd/aeoracle/main.go`
- Modify: `cmd/aeoracle/main_test.go`

- [x] Add `TestCloneRenderRequestPreservesFrameTargets` in `internal/aeoracle/request_test.go`.

Expected assertion: cloned request changes `AEPPath`, `OutputDir`, `DonePath`, and `MetadataPath`, preserves all `Frames`, and keeps `CompName` unless overridden.

- [x] Implement:

```go
func CloneRenderRequest(source RenderRequest, aepPath, compName, outputDir string) RenderRequest
```

Use `NewRenderRequest(aepPath, selectedCompName, outputDir, source.Frames)`, where `selectedCompName` is `compName` if non-empty, otherwise `source.CompName`.

- [x] Add `aeoracle clone-request` CLI flags:
  - `-from`
  - `-aep`
  - `-out`
  - `-comp`
  - `-json`

- [x] Verify:

```powershell
go test ./internal/aeoracle ./cmd/aeoracle
go run ./cmd/aeoracle plan -aep 'data\samples\motionbox\glitch\booyah-glitch\Booyah Glitch.aep' -comp 'グリッチテキスト' -out tmp_debug\aeoracle\booyah_compare\source -json
go run ./cmd/aeoracle clone-request -from tmp_debug\aeoracle\booyah_compare\source\request.json -aep flightdeck\showcase\booyah-clone\booyah-clone.aep -out tmp_debug\aeoracle\booyah_compare\clone -json
```

Expected: `tmp_debug/aeoracle/booyah_clone/request.json` exists and has the same 8 frame tags as the source request.

### Task 2: RED/GREEN Frame-Set Compare Core

**Files:**
- Create: `internal/aeoracle/frameset.go`
- Create: `internal/aeoracle/frameset_test.go`

- [x] Add `TestCompareFrameSetsReportsAllOKForIdenticalPNGs`.
- [x] Add `TestCompareFrameSetsReportsDifferentPixels`.
- [x] Add `TestCompareFrameSetsReportsMissingActualFrame`.
- [x] Add `TestCompareFrameSetsRejectsDuplicateTags`.

Use generated 2x2 PNGs in temp dirs, following `internal/aeoracle/pngcompare_test.go`.

- [x] Implement:

```go
func ReadRenderMetadata(path string) (RenderMetadata, error)
func CompareFrameSets(expectedMetaPath, actualMetaPath string, opts CompareOptions) (FrameSetCompareReport, error)
```

Rules:
- Require metadata `status == "ok"` for both sides.
- Align frames by `tag`.
- Preserve expected frame order.
- Missing expected or actual tags are report records, not fatal errors.
- Duplicate tags in either metadata are fatal errors.
- PNG read/size errors are fatal errors because the compare result is not reliable.

- [x] Verify:

```powershell
go test ./internal/aeoracle
```

### Task 3: CLI Compare-Set

**Files:**
- Modify: `cmd/aeoracle/main.go`
- Modify: `cmd/aeoracle/main_test.go`

- [x] Add `compare-set` usage and flags:
  - `-expected-meta`
  - `-actual-meta`
  - `-threshold`
  - `-json`
  - `-out`

- [x] Implement command output:
  - text: one summary line plus one line per non-ok frame.
  - JSON: full `FrameSetCompareReport`.
  - `-out`: write JSON report regardless of `-json`.

- [x] Implement exit codes:
  - 0 when `Summary.DifferentFrames == 0` and missing counts are zero.
  - 1 when report is valid but any frame differs or is missing.
  - 2 on usage or compare error.

- [x] Verify:

```powershell
go test ./cmd/aeoracle
```

### Task 4: Gap Ledger Coupling

**Files:**
- Modify: `internal/gapledger/render.go`
- Modify: `internal/gapledger/render_test.go`
- Modify: `internal/sliceworkflow/report.go`
- Modify: `internal/sliceworkflow/report_test.go`
- Modify: `cmd/aepslices/main.go`
- Modify: `cmd/aepslices/main_test.go`

- [x] Add `gapledger.FromFrameSetCompare(report aeoracle.FrameSetCompareReport) Report`.

Mapping:
- `different` -> `semantic-gap`, evidence kind `render_frame_delta`.
- `missing_actual` -> `write-gap`, evidence kind `render_frame_missing_actual`.
- `missing_expected` -> `investigate-gap`, evidence kind `render_frame_missing_expected`.
- Include frame tag, frame number, seconds, expected path, actual path, different pixel count, and max channel delta in context.

- [x] Add `aepslices diagnose -render-set <report.json>`.

Rules:
- Existing `-expected-png`/`-actual-png` remains supported for one-off PNG compare.
- `-render-set` may be combined with profile diff.
- Exit code remains 1 when any gap report has gaps.

- [x] Verify:

```powershell
go test ./internal/gapledger ./internal/sliceworkflow ./cmd/aepslices
```

### Task 5: Real Booyah Source-vs-Clone Render Compare

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/phase5-booyah-acceptance.md`
- Create: `flightdeck/work/aep-understanding-generation/phase6-booyah-render-compare.md`

- [x] Generate clone request:

```powershell
go run ./cmd/aeoracle clone-request -from tmp_debug\aeoracle\booyah\request.json -aep flightdeck\showcase\booyah-clone\booyah-clone.aep -out tmp_debug\aeoracle\booyah_clone -json
```

- [x] Render source and clone with AE 2025:

```powershell
go run ./cmd/aeoracle render -request tmp_debug\aeoracle\booyah_compare\source\request.json -ae 'E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe' -timeout-sec 900
go run ./cmd/aeoracle render -request tmp_debug\aeoracle\booyah_compare\clone\request.json -ae 'E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe' -timeout-sec 900
```

- [x] Compare frame sets:

```powershell
go run ./cmd/aeoracle compare-set -expected-meta tmp_debug\aeoracle\booyah_compare\source\metadata.json -actual-meta tmp_debug\aeoracle\booyah_compare\clone\metadata.json -threshold 0 -json -out tmp_debug\aeoracle\booyah_compare\compare_set.json
```

- [x] Merge profile and render gaps:

```powershell
go run ./cmd/aepslices diagnose -expected 'data\samples\motionbox\glitch\booyah-glitch\Booyah Glitch.aep' -actual 'flightdeck\showcase\booyah-clone\booyah-clone.aep' -render-set tmp_debug\aeoracle\booyah_compare\compare_set.json -json -out tmp_debug\aepslices\booyah_vs_clone_with_render.json
```

- [x] Record exact AE version, frame count, differing frame count, and top render-gap evidence in `phase6-booyah-render-compare.md`.

### Task 6: Readiness Decision for Recipe IR

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/cockpit.md`
- Create or modify: `flightdeck/work/aep-understanding-generation/phase6-recipe-ir-plan.md`

- [x] If Booyah compare produces valid frame-set render gaps, write `phase6-recipe-ir-plan.md` for a minimal recipe compiler.

Minimum recipe IR scope:
- project settings,
- one comp,
- solid/text/shape layers,
- transform static values,
- transform keyframes when supported,
- effects only when `cmd/capindex -q <effect>` reports supported writer capability,
- explicit downgrade/refusal records for unsupported requested constructs.

- [x] Do not start automated correction loops in this phase. Correction loops require a bounded parameter set and a render-delta objective already proven by `compare-set`.

### Task 7: Verification and Commit

- [x] Run `go test ./internal/aeoracle ./cmd/aeoracle ./internal/gapledger ./internal/sliceworkflow ./cmd/aepslices`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `git diff --check`.
- [x] Run the Booyah render compare commands from Task 5.
- [x] Update Flightdeck docs.
- [x] Commit with `feat(aeoracle): compare render frame sets`.

## Stop Rule

Stop before implementing recipe IR if any of these are true:
- `aeoracle render` cannot produce `ok` metadata for both source and clone.
- `compare-set` cannot produce a valid report due missing metadata or unreadable PNGs.
- The clone request cannot target the same semantic top comp without a manual comp-name mapping.
- Render gaps are dominated by automation failure rather than AEP/profile/writer differences.

If none of those blockers occur, proceed to `phase6-recipe-ir-plan.md` and keep the first recipe compiler deliberately small.
