# AEP Understanding Phase 3 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a generic render-oracle harness that can select sentinel frames, render them through existing AE automation, and compare PNG outputs with deterministic metadata.

**Architecture:** `internal/aeoracle` owns non-AE logic: sentinel frame selection from `profile.Profile`, sidecar request/metadata structs, and PNG comparison metrics. `cmd/aeoracle` is a thin CLI with `plan`, `render`, and `compare` subcommands. `scripts/aeoracle_render.jsx` is the generic AE-side renderer invoked through existing `scripts/ae_run.ps1`; it reads a JSON request and writes metadata/done files. Phase 3 does not generate gap ledgers or recipes.

**Tech Stack:** Go 1.25.1, standard `image/png`, `internal/aep`, `internal/profile`, existing `scripts/ae_run.ps1`, ExtendScript JSX.

---

## File Structure

- Create `internal/aeoracle/frames.go`: sentinel frame selection and request model.
- Create `internal/aeoracle/frames_test.go`: table tests for capped frame selection.
- Create `internal/aeoracle/pngcompare.go`: PNG exact/threshold comparison.
- Create `internal/aeoracle/pngcompare_test.go`: generated image comparison tests.
- Create `cmd/aeoracle/main.go`: CLI subcommands.
- Create `scripts/aeoracle_render.jsx`: AE renderer reading JSON sidecar.
- Modify `flightdeck/work/aep-understanding-generation/index.md`: mark Phase 3 active/result.
- Modify `flightdeck/cockpit.md`: route next step to Phase 3 closeout or Phase 4.

## Tasks

### Task 1: RED/GREEN Sentinel Frame Selection

- [x] Add `internal/aeoracle/frames_test.go`.
- [x] Test `SelectFrames` includes first and last comp frames for a simple comp.
- [x] Test keyframe times from admitted layer properties/effect params are included.
- [x] Test long quiet spans add midpoints between selected keyframes.
- [x] Test `MaxFrames` caps output deterministically and always keeps endpoints.
- [x] Implement `internal/aeoracle/frames.go`.
- [x] Run `go test ./internal/aeoracle`.

### Task 2: RED/GREEN Request and Metadata Contract

- [x] Add tests for JSON round-trip of `RenderRequest`:
  - schema version
  - AEP path
  - comp name
  - output directory
  - frame list with frame index, seconds, tag, and reason
- [x] Implement `WriteRequest(path string, req RenderRequest) error` and `ReadRequest(path string) (RenderRequest, error)`.
- [x] Validate request rejects empty AEP path, empty output dir, and empty frame list.
- [x] Run `go test ./internal/aeoracle`.

### Task 3: RED/GREEN PNG Comparison

- [x] Add tests that generate tiny PNGs in a temp dir.
- [x] Assert identical images produce `DifferentPixels=0`.
- [x] Assert one changed pixel is counted with exact threshold.
- [x] Assert a per-channel tolerance can ignore small deltas.
- [x] Implement `ComparePNG(expectedPath, actualPath string, opts CompareOptions) (CompareReport, error)`.
- [x] Include width/height mismatch as a structured error.
- [x] Run `go test ./internal/aeoracle`.

### Task 4: CLI Dry-Run and Compare

- [x] Create `cmd/aeoracle/main.go`.
- [x] Implement `plan`:
  - flags: `-aep`, `-comp`, `-out`, `-max-frames`, `-json`
  - opens AEP, builds profile, selects frames, prints or writes JSON request
- [x] Implement `compare`:
  - flags: `-expected`, `-actual`, `-threshold`, `-json`
  - prints compact metric or JSON
- [x] Implement `render`:
  - flags: `-request`, `-ae`, `-jsx`, `-timeout-sec`, `-dry-run`
  - dry-run validates request and prints the `ae_run.ps1` command without launching AE
  - non-dry-run invokes PowerShell `scripts/ae_run.ps1`
- [x] Run command-level smoke tests without launching AE.

### Task 5: Generic JSX Renderer

- [x] Create `scripts/aeoracle_render.jsx`.
- [x] Read JSON request path from `AEORACLE_REQUEST` environment variable first; fall back to `aeoracle_request.json` beside the JSX.
- [x] Open requested AEP, find requested comp by exact name or first comp when comp name is empty.
- [x] Set 8bpc, purge caches, render each frame to deterministic PNG names under output dir.
- [x] Write metadata JSON with AE version, project path, comp name, frame records, output paths, and status.
- [x] Write a done marker path from request when provided; otherwise `aeoracle_render.done` beside the request.

### Task 6: Verification and Commit

- [x] Run `go test ./internal/aeoracle ./cmd/aeoracle`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `git diff --check`.
- [x] Run reserved-word scan on touched files.
- [x] Run `go run ./cmd/aeoracle plan -aep flightdeck/showcase/text/text.aep -out tmp_debug/aeoracle/text -json`.
- [x] Run `go run ./cmd/aeoracle render -request tmp_debug/aeoracle/text/request.json -dry-run`.
- [x] Run `go run ./cmd/aeoracle compare` against generated test PNGs or document that package tests cover it.
- [x] Update Flightdeck status.
- [x] Commit with `feat(aeoracle): add sentinel render harness`.

## Verification Notes

- `plan` generated `tmp_debug/aeoracle/text/request.json` with frames 0, 45, and 90 for `flightdeck/showcase/text/text.aep`.
- `render -dry-run` validated the request and printed the `scripts/ae_run.ps1` invocation without launching AE.
- `compare` against generated 2x2 PNGs returned exit code 1 and reported 1 different pixel.

## Stop Rule

After Phase 3, stop before generating gap ledgers. Phase 4 must turn profile/render observations into structured gaps and capability-linked work items.
