# Versioned AEP Migration Matrix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or equivalent TDD execution. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a one-command migration matrix gate that batch-validates recipe fixtures across supported writer targets and available After Effects hosts.

**Architecture:** Keep conversion truth in `internal/aepmigrate.Convert`. Add a matrix layer that discovers recipe inputs, optionally discovers AE hosts under an install root such as `E:\adobe`, compiles each recipe once to a source `.aep`, runs `Convert` for each supported target, and writes aggregate JSON. The matrix distinguishes writer support from AE host availability: AE2021/AE2023/AE2024 hosts can be detected for open gates, while current writer targets remain AE2020/AE2022/AE2025 until native writer templates exist.

**Tech Stack:** Go standard library, `internal/aepmigrate`, `internal/recipe`, `internal/capindex`, existing `aehost.Host`.

---

## Scope

- Add `aepmigrate matrix`.
- Input can be a recipe directory or one explicit recipe path.
- Target labels are parsed from CLI values; unsupported writer targets are refused during argument parsing rather than silently remapped.
- `-ae-root E:\adobe` discovers `Adobe After Effects 2020` through `Adobe After Effects 2025` executable paths.
- If `-ae-open` is set without `-ae-versions`, each converted target uses the
  matching discovered or explicitly mapped AE executable.
- If `-ae-open -ae-versions AE2020,...` is set, the matrix opens the same
  converted output through each requested AE host version. This verifies host
  compatibility separately from writer target support.
- Output root contains:
  - `matrix.json` aggregate summary.
  - One directory per case with source `.aep`, target `.aep`, and convert report JSON.
- Status meanings:
  - `pass`: convert report status is pass, profile diff passes, and AE open passes when requested.
  - `blocked`: convert intentionally refuses the fixture.
  - `failed`: IO, compile, profile diff failure, or AE open failure.
  - `skipped`: target requested but writer or AE host is unavailable for the requested gate.

## Files

- Create: `internal/aepmigrate/matrix.go`
  - `MatrixOptions`, `MatrixReport`, `MatrixCase`, `BuildAEHostMap`, and `RunMatrix`.
- Create: `internal/aepmigrate/matrix_test.go`
  - Unit tests for host discovery, target parsing, unsupported target handling, and one recipe x two target batch output without AE.
- Modify: `cmd/aepmigrate/main.go`
  - Add `matrix` subcommand and flags.
- Modify: `cmd/aepmigrate/main_test.go`
  - CLI smoke test for `matrix -recipe ... -targets AE2020,AE2025 -out ...`.
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
  - Record the matrix gate.
- Modify: `flightdeck/cockpit.md`
  - Add the one-command validation entry.

## Tasks

- [x] Write failing tests for AE host discovery from an install root.
- [x] Implement host discovery.
- [x] Write failing tests for matrix running one recipe against multiple supported targets.
- [x] Implement matrix recipe compile, convert loop, per-case report writing, and aggregate JSON.
- [x] Write failing CLI test.
- [x] Implement `cmd/aepmigrate matrix`.
- [x] Add AE-open case-count guard so broad recipe matrices cannot accidentally launch hundreds of AE open gates in one run.
- [x] Add explicit AE-open host-version fanout with `-ae-versions`.
- [x] Run focused tests.
- [x] Run full Go verification.
- [x] Run one real matrix smoke without AE open.
- [x] Run one real matrix smoke with AE2025 open gate if available.
- [x] Run one real matrix smoke across AE2020-AE2025 open hosts.

## Notes

- Do not add PowerShell as the matrix driver. Go owns orchestration.
- Do not claim AE2021/AE2023/AE2024 writer support until scene/serializer exposes those targets.
- Default matrix jobs should be serial. AE automation is process-heavy and currently not designed for parallel open gates.
- `-ae-open` has a default safety cap (`-max-ae-open-cases 25`, set `0` to disable) because a broad `-recipes examples/recipes` matrix can expand to hundreds of AE launches. Full non-AE matrix runs are cheap; broad AE-open runs must be intentionally narrowed or explicitly uncapped.

## Verification

- `go test ./internal/aepmigrate -run Matrix -count=1 -v`
- `go test ./cmd/aepmigrate -run Matrix -count=1 -v`
- `go run ./cmd/aepmigrate matrix -recipe examples\recipes\minimal-comp-object-profile.json -sources AE2020 -targets AE2020,AE2022,AE2025 -out tmp\migration_matrix_smoke_one`
- `go run ./cmd/aepmigrate matrix -recipes examples\recipes -sources AE2020 -targets AE2020,AE2022,AE2025 -out tmp\migration_matrix_smoke_all`
  - Result after layer metadata, text-layer switch, text-layer advanced
    switch, text-parent, solid-backed null-controller, and camera/light static
    option/light-source/layer-object convert slices: 423 total, 291 pass, 132 blocked, 0 failed,
    0 skipped; command exits 1 because intentional blocked cases remain.
- `go run ./cmd/aepmigrate matrix -recipes examples\recipes -sources AE2020 -targets AE2020,AE2022,AE2025 -out tmp\migration_matrix_guard_all_ae -ae-root E:\adobe -ae-open`
  - Result: refused before launching AE: `AE open matrix would run 423 cases, above limit 25`.
- `go run ./cmd/aepmigrate matrix -recipe examples\recipes\minimal-comp-object-profile.json -sources AE2020 -targets AE2025 -out tmp\migration_matrix_smoke_ae2025 -ae-root E:\adobe -ae-open -max-ae-open-cases 1`
  - Result: 1 total, 1 pass, AE2025 open gate passed after clearing AE crash-state.
- `go run ./cmd/aepmigrate matrix -recipe examples\recipes\minimal-comp-object-profile.json -sources AE2020 -targets AE2020 -out tmp\migration_matrix_aehost_2020_2025 -ae-root E:\adobe -ae-open -ae-versions AE2020,AE2021,AE2022,AE2023,AE2024,AE2025 -ae-timeout-sec 240 -max-ae-open-cases 6`
  - Result: 6 total, 6 pass, 0 blocked, 0 failed, 0 skipped after clearing
    AE2020-AE2025 crash-state. This validates all installed AE2020-AE2025
    hosts against one AE2020 writer output.
