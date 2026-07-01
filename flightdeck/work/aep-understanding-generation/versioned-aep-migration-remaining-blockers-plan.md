# Versioned AEP Migration Remaining Blockers Implementation Plan

> Frozen historical plan. Do not append new execution steps here.
> Current execution state lives in `versioned-aep-migration-current.json`;
> current coverage results live in `versioned-aep-migration-coverage.json`.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Continue the `aep-understanding-generation` mainline by turning the post-text-animator migration boundary into a reviewed blocker ledger, then remove the next smallest blocked family.

**Architecture:** Use the validation summary as the total control surface. First rerun the full no-AE migration matrix to establish the current pass/blocked boundary after text animator support. Then group blockers by domain and implement the smallest next domain slice with TDD. For this phase, the first implementation target is the layer matte family (`minimal-layer-explicit-matte` and `minimal-layer-track-matte`) because it is smaller than mask path reconstruction and gradient-stroke synthesis.

**Tech Stack:** Go, `cmd/aepmigrate matrix`, `internal/aepmigrate`, `internal/profile`, `internal/profilediff`, Flightdeck validation ledger.

---

## Total View

Expected current boundary after text animator migration:

- Full no-AE matrix should improve from `423 total, 336 pass, 87 blocked` to approximately `423 total, 402 pass, 21 blocked`.
- Remaining blocker families are expected to be:
  - layer matte: `minimal-layer-explicit-matte`, `minimal-layer-track-matte`
  - layer mask: `minimal-layer-mask`
  - shape gradient stroke: `minimal-shape-gradient-stroke*`

Do not claim these numbers until the matrix is rerun and the result is written to `versioned-aep-migration-validation-summary.md`.

## Files

- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
  - Record the refreshed full matrix boundary and remaining blocker groups.
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
  - Append a concise progress note after the slice is verified.
- Modify likely: `internal/aepmigrate/convert_layer.go`
  - Add reopen-backed layer matte materialization if writer APIs support the required matte model.
- Modify likely: `internal/aepmigrate/convert_scope.go`
  - Admit the conservative matte fixtures only when the source layer shape is otherwise supported.
- Modify likely: `internal/aepmigrate/convert_test.go`
  - Add focused regression tests for `minimal-layer-explicit-matte` and `minimal-layer-track-matte`.

## Task 1: Refresh Full Matrix Boundary

- [x] **Step 1: Run the current full no-AE matrix**

```powershell
go run ./cmd/aepmigrate matrix -recipes examples\recipes -sources AE2020 -targets AE2020,AE2022,AE2025 -out tmp\migration_matrix_smoke_all
```

Expected command exit: `1` if intentional blocked cases remain.

- [x] **Step 2: Summarize blocker groups from `matrix.json`**

```powershell
@'
import json
from collections import Counter
from pathlib import Path
m=json.loads(Path('tmp/migration_matrix_smoke_all/matrix.json').read_text())
print(m['summary'])
for recipe, count in sorted(Counter(c['recipe_name'] for c in m['cases'] if c['status']!='pass').items()):
    print(count, recipe)
'@ | python -
```

Expected: no failed cases; blocker count grouped by recipe.

- [x] **Step 3: Update validation summary**

Record the matrix summary and grouped blockers in `versioned-aep-migration-validation-summary.md`.

## Task 2: Layer Matte Slice, Red Tests

- [x] **Step 1: Add focused failing convert tests**

Add tests named:

```go
func TestConvertWritesRecipeLayerExplicitMatteProject(t *testing.T)
func TestConvertWritesRecipeLayerTrackMatteProject(t *testing.T)
```

Both should compile the existing recipe, run `Convert` to `VersionAE2025`, and assert:

```go
if report.Summary.Status != StatusPass {
    t.Fatalf("status = %q, entries=%+v, diffs=%+v", report.Summary.Status, report.Entries, report.Verification.ProfileDiffs)
}
if report.Verification.ProfileDiffStatus != "pass" || report.Verification.ProfileDiffCount != 0 {
    t.Fatalf("profile diff verification = %+v, want pass with 0 diffs", report.Verification)
}
```

- [x] **Step 2: Run red tests**

```powershell
go test ./internal/aepmigrate -run "TestConvertWritesRecipeLayer(ExplicitMatte|TrackMatte)Project" -count=1
```

Expected: fail because matte migration is not implemented or still blocked.

## Task 3: Layer Matte Implementation

- [x] **Step 1: Inspect writer/profile support**

Use `rg` to find existing matte APIs:

```powershell
rg -n "TrackMatte|Matte|Set.*Matte|matte" internal/aep internal/scene internal/serializer internal/profile
```

Expected: identify whether the writer supports classic track matte, explicit matte, or both.

- [x] **Step 2: Implement the smallest supported matte path**

If writer APIs support matte assignment, add a reopen-backed pass in `internal/aepmigrate/convert_layer.go` or a new focused file such as `convert_matte.go`.

If only one recipe is supported by writer APIs, implement that one and leave the other blocked with an explicit scope reason. Do not fake support by dropping matte refs.

- [x] **Step 3: Run green tests**

```powershell
go test ./internal/aepmigrate -run "TestConvertWritesRecipeLayer(ExplicitMatte|TrackMatte)Project" -count=1
```

Expected: implemented supported cases pass; unsupported cases must have clear tests/ledger entries explaining the limitation.

## Task 4: Matrix and Validation Ledger

- [x] **Step 1: Run focused matte matrix**

```powershell
go run ./cmd/aepmigrate matrix -recipe examples\recipes\minimal-layer-explicit-matte.json -recipe examples\recipes\minimal-layer-track-matte.json -sources recipe -targets AE2020,AE2022,AE2025 -out tmp\migration_matrix_layer_matte
```

Actual result: 6 total, 4 pass, 2 blocked, 0 failed, 0 skipped. The blocked
cases are `minimal-layer-explicit-matte` from an AE2025 source to AE2020/AE2022
targets, intentionally blocked by the current writer contract.

- [x] **Step 1b: Refresh full no-AE matrix after implementation**

```powershell
go run ./cmd/aepmigrate matrix -recipes examples\recipes -sources AE2020 -targets AE2020,AE2022,AE2025 -out tmp\migration_matrix_smoke_all
```

Actual result: 423 total, 405 pass, 18 blocked, 0 failed, 0 skipped.

- [x] **Step 2: Update validation summary**

Record the focused matrix result by domain:

- domain: `layer`
- capability: matte / track matte
- writer coverage: `PD-1x3` or explicit blocked scope
- AE host coverage: pending unless separately run
- artifact path: `tmp/migration_matrix_layer_matte/matrix.json`

- [x] **Step 3: Update history**

Append a concise note to `history.md` with the implementation boundary and matrix result.

## Task 5: Commit Gate

- [x] **Step 1: Run full verification**

```powershell
go test ./...
go vet ./...
git diff --check
```

- [x] **Step 2: Read commit/verify knowledge**

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

- [x] **Step 3: Commit**

Commit message should be one of:

```text
feat(aepmigrate): preserve layer mattes
```

or, if the slice is only matrix/ledger triage:

```text
docs(aepmigrate): record remaining migration blockers
```

## Self-Review

- This plan uses the validation summary as the total control surface.
- It does not conflate writer targets with AE host versions.
- It keeps mask and gradient stroke out of the matte slice unless the refreshed matrix proves a different smallest next target.
- It requires TDD for behavior changes and a separate matrix/ledger update before commit.

---

## Task 6: Layer Mask Slice, Red Test

- [x] **Step 1: Add a focused failing convert test**

Add:

```go
func TestConvertWritesRecipeLayerMaskProject(t *testing.T)
```

The test should compile `examples/recipes/minimal-layer-mask.json`, run
`Convert` to `VersionAE2025`, and assert status pass with profile diff pass and
zero diffs.

- [x] **Step 2: Run the red test**

```powershell
go test ./internal/aepmigrate -run TestConvertWritesRecipeLayerMaskProject -count=1
```

Expected: fail because mask migration is not implemented or still blocked by
the current convert scope.

## Task 7: Layer Mask Implementation

- [x] **Step 1: Inspect writer/profile support**

Use narrow searches around `AddMask`, mask writer setters, profile mask fields,
and recipe mask materialization.

- [x] **Step 2: Implement the smallest supported mask path**

Support the existing `minimal-layer-mask` fixture only if all visible profile
fields can be reconstructed through existing writer APIs. Do not silently drop
mask path keyframes or mask options.

- [x] **Step 3: Run the green test**

```powershell
go test ./internal/aepmigrate -run TestConvertWritesRecipeLayerMaskProject -count=1
```

Expected: supported mask fixture passes profile diff verification.

## Task 8: Mask Matrix and Ledger

- [x] **Step 1: Run focused mask matrix**

```powershell
go run ./cmd/aepmigrate matrix -recipe examples\recipes\minimal-layer-mask.json -sources AE2020 -targets AE2020,AE2022,AE2025 -out tmp\migration_matrix_layer_mask
```

Actual result: 3 total, 3 pass, 0 blocked, 0 failed, 0 skipped.

- [x] **Step 2: Refresh full no-AE matrix**

```powershell
go run ./cmd/aepmigrate matrix -recipes examples\recipes -sources AE2020 -targets AE2020,AE2022,AE2025 -out tmp\migration_matrix_smoke_all
```

Actual result: 423 total, 408 pass, 15 blocked, 0 failed, 0 skipped.

- [x] **Step 3: Update validation summary, index, and history**

Record the focused mask matrix, refreshed full boundary, remaining blockers,
and any explicit unsupported scope.

## Task 9: Mask Commit Gate

- [x] **Step 1: Run full verification**

```powershell
go test ./...
go vet ./...
git diff --check
```

- [x] **Step 2: Read commit/verify knowledge**

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

- [x] **Step 3: Commit**

Use:

```text
feat(aepmigrate): preserve layer masks
```

---

## Task 10: Shape Gradient Stroke Slice, Red Tests

- [x] **Step 1: Add focused failing convert tests**

Add tests for the four existing gradient-stroke recipes:

```go
func TestConvertWritesRecipeShapeGradientStrokeProjects(t *testing.T)
```

The table should cover:

- `minimal-shape-gradient-stroke.json`
- `minimal-shape-gradient-stroke-alpha-stops.json`
- `minimal-shape-gradient-stroke-highlight.json`
- `minimal-shape-gradient-stroke-style.json`

Each case should compile the recipe, run `Convert` to `VersionAE2025`, and
assert status pass with profile diff pass and zero diffs.

- [x] **Step 2: Run the red tests**

```powershell
go test ./internal/aepmigrate -run TestConvertWritesRecipeShapeGradientStrokeProjects -count=1
```

Expected: fail because gradient stroke migration is not implemented or still
blocked by the current convert scope.

## Task 11: Shape Gradient Stroke Implementation

- [x] **Step 1: Inspect writer/profile support**

Use narrow searches around `GradientStrokeNode`, `AddGradientStroke`,
`hasGradientFillGraphic`, `materializeShapeGradientFill`, and shape profile
properties.

- [x] **Step 2: Implement the supported gradient stroke path**

Reuse the gradient fill replay pattern for gradient stroke, including gradient
type, start/end points, highlight length/angle, gradient color/alpha stops,
stroke width, line cap, line join, and miter limit.

- [x] **Step 3: Run the green tests**

```powershell
go test ./internal/aepmigrate -run TestConvertWritesRecipeShapeGradientStrokeProjects -count=1
```

Expected: all four recipe-owned gradient-stroke fixtures pass profile diff
verification.

## Task 12: Gradient Stroke Matrix and Ledger

- [x] **Step 1: Run focused gradient-stroke matrix**

```powershell
go run ./cmd/aepmigrate matrix -recipe examples\recipes\minimal-shape-gradient-stroke.json -recipe examples\recipes\minimal-shape-gradient-stroke-alpha-stops.json -recipe examples\recipes\minimal-shape-gradient-stroke-highlight.json -recipe examples\recipes\minimal-shape-gradient-stroke-style.json -sources AE2020 -targets AE2020,AE2022,AE2025 -out tmp\migration_matrix_shape_gradient_stroke
```

Actual result: 12 total, 12 pass, 0 blocked, 0 failed, 0 skipped.

- [x] **Step 2: Refresh full no-AE matrix**

```powershell
go run ./cmd/aepmigrate matrix -recipes examples\recipes -sources AE2020 -targets AE2020,AE2022,AE2025 -out tmp\migration_matrix_smoke_all
```

Actual result: 423 total, 420 pass, 3 blocked, 0 failed, 0 skipped.

- [x] **Step 3: Update validation summary, index, and history**

Record the focused gradient-stroke matrix, refreshed full boundary, and
remaining blockers.

## Task 13: Gradient Stroke Commit Gate

- [x] **Step 1: Run full verification**

```powershell
go test ./...
go vet ./...
git diff --check
```

- [x] **Step 2: Read commit/verify knowledge**

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

- [x] **Step 3: Commit**

Use:

```text
feat(aepmigrate): preserve gradient strokes
```

---

## Task 14: Host-Open Representative Plan

- [x] **Step 1: Treat writer targets and AE host versions as separate axes**

Current writer targets remain `W2020/W2022/W2025` because those are the
implemented writer fingerprints. AE host-open validation should cover installed
hosts `H2020/H2021/H2022/H2023/H2024/H2025` through `-ae-versions all`.

- [x] **Step 2: Select representative W2020 outputs for all-host smoke**

Use W2020 source and W2020 target outputs so older hosts can open the files.
This slice covers recent and previously inferred evidence without expanding the
writer surface:

- `minimal-text-animator-skew.json`
- `minimal-text-animator-color-value-keyframes.json`
- `minimal-layer-track-matte.json`
- `minimal-layer-mask.json`
- `minimal-shape-gradient-stroke.json`

The AE2025 explicit matte fixture is excluded from all-host smoke because its
current writer contract is AE2025-target only.

## Task 15: Host-Open Representative Matrix

- [x] **Step 1: Run all-host representative matrix**

```powershell
go run ./cmd/aepmigrate matrix -recipe examples\recipes\minimal-text-animator-skew.json -recipe examples\recipes\minimal-text-animator-color-value-keyframes.json -recipe examples\recipes\minimal-layer-track-matte.json -recipe examples\recipes\minimal-layer-mask.json -recipe examples\recipes\minimal-shape-gradient-stroke.json -sources AE2020 -targets AE2020 -ae-open -ae-root E:\adobe -ae-versions all -max-ae-open-cases 30 -out tmp\migration_matrix_representative_all_hosts
```

Actual result: 30 total, 30 pass, 0 blocked, 0 failed, 0 skipped. The matrix
converted five W2020 writer outputs and opened each output in H2020-H2025.

- [x] **Step 2: Summarize host-open result by recipe and host**

Record any skipped host as missing environment, any blocked case as conversion
scope, and any failed case as a real host-open regression until proven
otherwise.

## Task 16: Host-Open Ledger and Commit Gate

- [x] **Step 1: Update validation summary, index, and history**

Promote the all-host representative matrix into the validation summary. Remove
`INFER-MID-HOSTS` wording for the two text animator representatives only if the
all-host matrix passes for H2021-H2024.

- [x] **Step 2: Run full verification**

```powershell
go test ./...
go vet ./...
git diff --check
```

- [x] **Step 3: Read commit/verify knowledge and commit**

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Use:

```text
docs(aepmigrate): record all-host migration evidence
```

---

## Task 17: Generated Coverage Ledger Plan

- [x] **Step 1: Define the narrow generated-ledger surface**

Add a generated ledger command that reads one `matrix.json` and emits a
Markdown table grouped by inferred recipe domain. This is not a replacement for
the reviewed validation summary; it is a repeatable input that answers "which
recipes/domains passed, blocked, failed, or skipped in this matrix?" without
manual counting.

Files:

- Create: `internal/aepmigrate/matrix_ledger.go`
- Create: `internal/aepmigrate/matrix_ledger_test.go`
- Modify: `cmd/aepmigrate/main.go`
- Modify: `cmd/aepmigrate/main_test.go`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`

- [x] **Step 2: Keep domain inference conservative**

Infer domain from recipe names only:

- `minimal-text-*` -> `text`
- `minimal-layer-*` -> `layer`
- `minimal-shape-*` -> `shape`
- `minimal-effect-*` or `minimal-adjustment-*` -> `effect`
- `minimal-transform-*` -> `transform`
- `minimal-comp-*` -> `comp`
- `minimal-camera-*` or `minimal-light-*` -> `camera-light`
- `minimal-precomp-*` -> `precomp`
- `minimal-project-*` -> `project`
- anything else -> `other`

The command must preserve raw recipe names and case counts so manual review can
override any rough domain bucket later.

## Task 18: Generated Coverage Ledger TDD

- [x] **Step 1: Add failing internal ledger tests**

Add tests that build a small `MatrixReport` with mixed text/layer/shape cases
and assert:

- ledger rows are grouped by recipe name,
- domain inference follows the conservative recipe prefixes,
- per-row counts include pass/blocked/failed/skipped,
- AE-open host versions are listed only when present,
- Markdown output includes a stable table header and rows.

- [x] **Step 2: Run the red internal test**

```powershell
go test ./internal/aepmigrate -run TestBuildMatrixLedger -count=1
```

Expected: fail because the ledger API does not exist.

- [x] **Step 3: Add failing CLI test**

Add a CLI test that writes a tiny `matrix.json`, runs:

```powershell
aepmigrate ledger -matrix matrix.json -out ledger.md
```

and asserts the output file contains the Markdown table and stdout prints
`migration ledger:`.

- [x] **Step 4: Run the red CLI test**

```powershell
go test ./cmd/aepmigrate -run TestRunLedgerWritesMarkdown -count=1
```

Expected: fail because the `ledger` subcommand does not exist.

## Task 19: Generated Coverage Ledger Implementation

- [x] **Step 1: Implement internal ledger builder and renderer**

Implement:

```go
func ReadMatrixLedger(path string) (MatrixLedger, error)
func BuildMatrixLedger(report MatrixReport) MatrixLedger
func RenderMatrixLedgerMarkdown(ledger MatrixLedger) string
```

The Markdown table columns are:

```text
Domain | Recipe | Status | Pass | Blocked | Failed | Skipped | Sources | Targets | AE Hosts
```

- [x] **Step 2: Implement `aepmigrate ledger` CLI**

Flags:

```text
-matrix path/to/matrix.json
-out path/to/ledger.md
```

If `-out` is omitted, print Markdown to stdout. If `-out` is present, write the
file and print `migration ledger: <path>`.

- [x] **Step 3: Run green focused tests**

```powershell
go test ./internal/aepmigrate -run TestBuildMatrixLedger -count=1
go test ./cmd/aepmigrate -run TestRunLedgerWritesMarkdown -count=1
```

Expected: both pass.

## Task 20: Generated Coverage Ledger Matrix Use and Commit Gate

- [x] **Step 1: Generate ledgers for current matrices**

```powershell
go run ./cmd/aepmigrate ledger -matrix tmp\migration_matrix_smoke_all\matrix.json -out tmp\migration_matrix_smoke_all\ledger.md
go run ./cmd/aepmigrate ledger -matrix tmp\migration_matrix_representative_all_hosts\matrix.json -out tmp\migration_matrix_representative_all_hosts\ledger.md
```

- [x] **Step 2: Update validation summary and history**

Record the generated ledger artifacts as supporting evidence. Do not replace
the reviewed summary table with generated output.

- [x] **Step 3: Run full verification**

```powershell
go test ./...
go vet ./...
git diff --check
```

- [x] **Step 4: Read commit/verify knowledge and commit**

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Use:

```text
feat(aepmigrate): generate matrix coverage ledgers
```

---

## Task 21: Host-Open Policy

- [x] **Step 1: Add the host-open policy to the validation plan**

Define when a capability needs `OPEN-ALL-HOSTS`, when `OPEN-H2025` is enough,
and when `INFER-MID-HOSTS` is allowed. Keep writer targets and AE hosts
separate: policy cannot imply W2021/W2023/W2024 writer support.

- [x] **Step 2: Apply the policy to current gaps**

Current policy application:

- Text animator variants: `OPEN-ALL-HOSTS` representatives plus `PD-3x3` are
  enough for non-representative variants unless a new writer path appears.
- Layer mask: the only current recipe-owned mask fixture already has
  `OPEN-ALL-HOSTS`.
- Classic track matte: current representative already has `OPEN-ALL-HOSTS`.
- AE2025 explicit matte: version-gated; do not run H2020-H2024 against a
  W2025-only contract as if it were normal all-host evidence.
- Shape gradient stroke: alpha stops, highlight, and style use distinct
  gradient/stroke writer paths, so run all current gradient-stroke variants
  across H2020-H2025.

## Task 22: Shape Gradient Stroke All-Host Matrix

- [x] **Step 1: Run all-host matrix for every gradient-stroke variant**

```powershell
go run ./cmd/aepmigrate matrix -recipe examples\recipes\minimal-shape-gradient-stroke.json -recipe examples\recipes\minimal-shape-gradient-stroke-alpha-stops.json -recipe examples\recipes\minimal-shape-gradient-stroke-highlight.json -recipe examples\recipes\minimal-shape-gradient-stroke-style.json -sources AE2020 -targets AE2020 -ae-open -ae-root E:\adobe -ae-versions all -ae-timeout-sec 240 -max-ae-open-cases 24 -out tmp\migration_matrix_shape_gradient_stroke_all_hosts
```

Actual result: 24 total, 24 pass, 0 blocked, 0 failed, 0 skipped. This proves
W2020 gradient-stroke variant outputs open in H2020-H2025.

- [x] **Step 2: Generate the matrix ledger**

```powershell
go run ./cmd/aepmigrate ledger -matrix tmp\migration_matrix_shape_gradient_stroke_all_hosts\matrix.json -out tmp\migration_matrix_shape_gradient_stroke_all_hosts\ledger.md
```

## Task 23: Host-Open Policy Ledger and Commit Gate

- [x] **Step 1: Update validation summary, index, and history**

Record the policy and the shape gradient-stroke all-host artifact. Remove the
missing-evidence item for non-representative layer/shape variants if the matrix
passes and the policy says no remaining current variant requires more evidence.

- [x] **Step 2: Run full verification**

```powershell
go test ./...
go vet ./...
git diff --check
```

- [x] **Step 3: Read commit/verify knowledge and commit**

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Use:

```text
docs(aepmigrate): define host-open validation policy
```

---

## Task 24: Matrix Ledger-Out Flag TDD

- [x] **Step 1: Add failing CLI test for `matrix -ledger-out`**

Add a focused test in `cmd/aepmigrate/main_test.go` that runs:

```powershell
aepmigrate matrix -recipe minimal.json -sources AE2020 -targets AE2020 -out matrix -ledger-out matrix\ledger.md
```

and asserts:

- `matrix.json` exists,
- `ledger.md` exists,
- stdout contains both `migration matrix:` and `migration ledger:`,
- the ledger contains the recipe row.

- [x] **Step 2: Run the red test**

```powershell
go test ./cmd/aepmigrate -run TestRunMatrixWritesLedgerOut -count=1
```

Expected: fail because `-ledger-out` is not defined.

## Task 25: Matrix Ledger-Out Implementation

- [x] **Step 1: Implement `-ledger-out` on `aepmigrate matrix`**

Add an optional `-ledger-out` flag to `runMatrix`. After `RunMatrix` returns,
build a ledger from the returned report, set its source to the matrix JSON path,
write the Markdown file, and print `migration ledger: <path>`.

- [x] **Step 2: Run focused green tests**

```powershell
go test ./cmd/aepmigrate -run "TestRun(MatrixWritesLedgerOut|LedgerWritesMarkdown)" -count=1
go test ./internal/aepmigrate -run "Test(Build|Render)MatrixLedger" -count=1
```

## Task 26: Recurring Matrix Command Ledger and Commit Gate

- [x] **Step 1: Run recurring full no-AE matrix command with ledger output**

```powershell
go run ./cmd/aepmigrate matrix -recipes examples\recipes -sources AE2020 -targets AE2020,AE2022,AE2025 -out tmp\migration_matrix_smoke_all -ledger-out tmp\migration_matrix_smoke_all\ledger.md
```

Actual result: command exited 1 because intentional blockers remain, and wrote
both `matrix.json` and `ledger.md`. Summary: 423 total, 420 pass, 3 blocked, 0
failed, 0 skipped.

- [x] **Step 2: Update validation plan, summary, index, and history**

Record the recurring command form and keep the reviewed summary as the source
of truth.

- [x] **Step 3: Run full verification**

```powershell
go test ./...
go vet ./...
git diff --check
```

- [x] **Step 4: Read commit/verify knowledge and commit**

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Use:

```text
feat(aepmigrate): write matrix ledgers during runs
```

---

## Task 27: Native Writer Targets 2021/2023/2024 TDD

- [x] **Step 1: Add failing `NewProject` target tests**

Extend `TestNewProject_AllTargetsParseAndSvap` to include existing embedded
templates:

- AE2021 svap `0b120e04`, 26 root children
- AE2023 svap `0b3a8634`, 30 root children
- AE2024 svap `0f010e02`, 31 root children

Expected: compile fails until `TargetAE2021`, `TargetAE2023`, and
`TargetAE2024` exist.

- [x] **Step 2: Add failing migration version/matrix tests**

Extend `TestParseVersionLabelAcceptsSupportedTargets` and
`TestMatrixAllPresetsExpandWritersAndAEOpenHosts` so writer `all` expands to
AE2020-AE2025 inclusive.

Expected: compile/test failure until migration labels and matrix writer labels
include the new targets.

## Task 28: Native Writer Target Implementation

- [x] **Step 1: Add target constants, aliases, and embedded templates**

Expose `TargetAE2021`, `TargetAE2023`, and `TargetAE2024` through scene,
serializer aliases, and public aep aliases. Embed the existing project
templates under `internal/serializer/templates/project/`.

- [x] **Step 2: Add migration version labels and target mapping**

Add `VersionAE2021`, `VersionAE2023`, and `VersionAE2024`; update
`ParseVersionLabel`, `NormalizeSourceVersion`, `matrixWriterLabels`, and
`aepTarget`.

- [x] **Step 3: Run focused green tests**

```powershell
go test ./internal/aep_test -run TestNewProject_AllTargetsParseAndSvap -count=1
go test ./internal/aepmigrate -run "TestParseVersionLabel|TestMatrixAllPresetsExpandWritersAndAEOpenHosts" -count=1
```

## Task 29: Native Writer Target Matrix and Ledger

- [x] **Step 1: Run a narrow 6-target writer matrix**

```powershell
go run ./cmd/aepmigrate matrix -recipe examples\recipes\minimal-comp-object-profile.json -sources AE2020 -targets all -out tmp\migration_matrix_writer_targets_all -ledger-out tmp\migration_matrix_writer_targets_all\ledger.md
```

Actual result: 6 total, 6 pass, 0 blocked, 0 failed, 0 skipped. This proves the
matrix can produce all writer targets for a stable comp-only fixture.

Also refreshed the full no-AE boundary with all writer targets:

```powershell
go run ./cmd/aepmigrate matrix -recipes examples\recipes -sources AE2020 -targets all -out tmp\migration_matrix_smoke_all -ledger-out tmp\migration_matrix_smoke_all\ledger.md
```

Actual result: 846 total, 840 pass, 6 blocked, 0 failed, 0 skipped. The blocked
cases are still only `minimal-layer-explicit-matte` outside the AE2025 source
contract.

- [x] **Step 2: Update validation plan, summary, index, and history**

Record that W2021/W2023/W2024 are now native writer targets for the project
skeleton and matrix runner, with narrow fixture evidence before broad migration
claims.

- [x] **Step 3: Run full verification**

```powershell
go test ./...
go vet ./...
git diff --check
```

- [x] **Step 4: Read commit/verify knowledge and commit**

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Use:

```text
feat(aepmigrate): add native 2021 2023 2024 writer targets
```

---

## Task 30: Source-Contract Skips for Explicit Matte Matrix Cases

**Goal:** Keep the full W2020-source no-AE matrix focused on recipes that can
actually be authored by the requested source writer, while preserving explicit
matte evidence in an AE2025-source focused matrix.

**Architecture:** Do not loosen the recipe validator or the explicit matte
writer contract. Classify compile reports whose only refusal is
`explicit_matte_requires_ae2025` as `skipped` with reason
`source_contract_unsupported`; all other invalid compile reports remain
`blocked`.

**Files:**
- Modify: `internal/aepmigrate/matrix.go`
- Modify/Test: `internal/aepmigrate/matrix_test.go`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`
- Possibly modify: `flightdeck/work/aep-understanding-generation/index.md`
- Possibly modify: `flightdeck/work/aep-understanding-generation/history.md`

- [x] **Step 1: Add red matrix test for source-contract skip**

Add a test to `internal/aepmigrate/matrix_test.go` that runs
`minimal-layer-explicit-matte.json` with source `AE2020` and target `AE2025`.
Expected result before implementation: the case is currently `blocked` with
reason `recipe_compile_blocked`; the new expected result is `skipped` with
reason `source_contract_unsupported`.

```go
func TestRunMatrixSkipsRecipeWhenSourceWriterViolatesRecipeContract(t *testing.T) {
	root := t.TempDir()
	recipePath := filepath.Join("..", "..", "examples", "recipes", "minimal-layer-explicit-matte.json")
	outRoot := filepath.Join(root, "matrix")

	report, err := RunMatrix(MatrixOptions{
		RecipePaths:  []string{recipePath},
		SourceLabels: []string{"AE2020"},
		TargetLabels: []string{"AE2025"},
		OutRoot:      outRoot,
	})
	if err != nil {
		t.Fatalf("RunMatrix: %v", err)
	}
	if report.Summary.Total != 1 || report.Summary.Skipped != 1 || report.Summary.Blocked != 0 {
		t.Fatalf("summary = %+v, want total=1 skipped=1 blocked=0; cases=%+v", report.Summary, report.Cases)
	}
	if got := report.Cases[0].Status; got != MatrixStatusSkipped {
		t.Fatalf("status = %s, want %s; case=%+v", got, MatrixStatusSkipped, report.Cases[0])
	}
	if got := report.Cases[0].Reason; got != "source_contract_unsupported" {
		t.Fatalf("reason = %q, want source_contract_unsupported; case=%+v", got, report.Cases[0])
	}
}
```

Run:

```powershell
go test ./internal/aepmigrate -run TestRunMatrixSkipsRecipeWhenSourceWriterViolatesRecipeContract -count=1
```

Expected before implementation: FAIL because the case is still blocked.

- [x] **Step 2: Implement source-contract skip classification**

In `internal/aepmigrate/matrix.go`, add a small helper near `runMatrixCase`:

```go
func matrixCompileInvalidReason(report recipe.Report) (MatrixStatus, string) {
	if len(report.Refusals) == 1 && report.Refusals[0].Code == "explicit_matte_requires_ae2025" {
		return MatrixStatusSkipped, "source_contract_unsupported"
	}
	return MatrixStatusBlocked, "recipe_compile_blocked"
}
```

Then replace the invalid compile block in `runMatrixCase`:

```go
if !compileReport.Valid {
	c.Status, c.Reason = matrixCompileInvalidReason(compileReport)
	return c
}
```

Run:

```powershell
go test ./internal/aepmigrate -run TestRunMatrixSkipsRecipeWhenSourceWriterViolatesRecipeContract -count=1
```

Expected after implementation: PASS.

- [x] **Step 3: Refresh focused and full matrices**

Run the explicit matte source-contract check:

```powershell
go run ./cmd/aepmigrate matrix -recipe examples\recipes\minimal-layer-explicit-matte.json -sources AE2020 -targets all -out tmp\migration_matrix_explicit_matte_source_contract -ledger-out tmp\migration_matrix_explicit_matte_source_contract\ledger.md
```

Expected: 6 total, 0 pass, 0 blocked, 0 failed, 6 skipped.

Run the valid explicit matte contract check:

```powershell
go run ./cmd/aepmigrate matrix -recipe examples\recipes\minimal-layer-explicit-matte.json -sources AE2025 -targets AE2025 -out tmp\migration_matrix_explicit_matte_ae2025 -ledger-out tmp\migration_matrix_explicit_matte_ae2025\ledger.md
```

Expected: 1 total, 1 pass, 0 blocked, 0 failed, 0 skipped.

Refresh the recurring full no-AE matrix:

```powershell
go run ./cmd/aepmigrate matrix -recipes examples\recipes -sources AE2020 -targets all -out tmp\migration_matrix_smoke_all -ledger-out tmp\migration_matrix_smoke_all\ledger.md
```

Expected: 846 total, 840 pass, 0 blocked, 0 failed, 6 skipped.

- [x] **Step 4: Update validation summary and index/history**

Update `versioned-aep-migration-validation-summary.md` so the current remaining
blocker section says there are no remaining blocked cases in the W2020-source
full no-AE matrix, and the six explicit matte cases are source-contract skips.
Add raw artifacts for:

- `tmp/migration_matrix_explicit_matte_source_contract/matrix.json`
- `tmp/migration_matrix_explicit_matte_ae2025/matrix.json`
- refreshed `tmp/migration_matrix_smoke_all/matrix.json`

Update `index.md` and `history.md` with the same current totals only if their
current-state paragraphs mention the old blocked count.

- [x] **Step 5: Full verification and commit**

Run:

```powershell
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
fix(aepmigrate): skip source-incompatible matrix recipes
```

---

## Task 31: Recurring Matrix Verification Script

**Goal:** Turn the reviewed no-AE migration matrix boundary into a single
tracked command that future workers can run before claiming the matrix is still
clean.

**Architecture:** Add a PowerShell script under `scripts/migration/` because
existing repository automation is PowerShell-based and matrix artifacts belong
under `tmp/`. The script runs the full W2020-source no-AE matrix and the
AE2025 explicit-matte valid contract matrix, parses each `matrix.json`, and
fails if totals drift from the reviewed boundary.

**Files:**
- Create: `scripts/migration/verify_matrix.ps1`
- Modify: `flightdeck/knowledge/workflow/verify.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-plan.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`

- [x] **Step 1: Add the recurring verification script**

Create `scripts/migration/verify_matrix.ps1` with these behaviors:

- default output root: `tmp\migration_matrix_verify`
- run full no-AE matrix:
  `go run ./cmd/aepmigrate matrix -recipes examples\recipes -sources AE2020 -targets all`
- assert summary is `846 total, 840 pass, 0 blocked, 0 failed, 6 skipped`
- run explicit matte valid contract matrix:
  `go run ./cmd/aepmigrate matrix -recipe examples\recipes\minimal-layer-explicit-matte.json -sources AE2025 -targets AE2025`
- assert summary is `1 total, 1 pass, 0 blocked, 0 failed, 0 skipped`
- print a concise PASS line for each matrix

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
```

Expected: both matrix checks print `PASS` and the process exits 0.

- [x] **Step 2: Document the recurring command**

Add the command to `flightdeck/knowledge/workflow/verify.md` under the
versioned migration area. Also point `index.md`,
`versioned-aep-migration-validation-plan.md`, and
`versioned-aep-migration-validation-summary.md` at the script as the recurring
no-AE matrix gate.

- [x] **Step 3: Full verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
chore(aepmigrate): add recurring matrix verification
```

---

## Task 32: Structural Domain All-Host Representative Evidence

**Goal:** Add direct H2020-H2025 open evidence for migrated structural domains
that currently have only generic representative coverage: project, comp,
camera, light, and precomp.

**Architecture:** Use the matrix runner's separated writer-target and AE host
axes. Produce W2020 output from stable representative recipes, then open that
same target writer output in every installed AE host from 2020 through 2025.
Keep the AE-open case count explicit at 30.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`

- [x] **Step 1: Run the structural all-host representative matrix**

Run:

```powershell
go run ./cmd/aepmigrate matrix `
  -recipe examples\recipes\minimal-project-display-settings.json `
  -recipe examples\recipes\minimal-comp-object-profile.json `
  -recipe examples\recipes\minimal-camera-object-profile.json `
  -recipe examples\recipes\minimal-light-object-profile.json `
  -recipe examples\recipes\minimal-precomp-layer.json `
  -sources AE2020 `
  -targets AE2020 `
  -out tmp\migration_matrix_structural_all_hosts `
  -ledger-out tmp\migration_matrix_structural_all_hosts\ledger.md `
  -ae-root E:\adobe `
  -ae-open `
  -ae-versions all `
  -ae-timeout-sec 240 `
  -max-ae-open-cases 30
```

Expected: 30 total, 30 pass, 0 blocked, 0 failed, 0 skipped. If any AE host is
missing or an open fails, do not broaden the summary; record the exact missing
host or failure.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, add `tmp/migration_matrix_structural_all_hosts/matrix.json`
and `ledger.md` to raw artifacts, then update Other Domains host-open coverage:

- `project`: `OPEN-ALL-HOSTS` for `minimal-project-display-settings`
- `comp`: `OPEN-ALL-HOSTS` for `minimal-comp-object-profile`
- `camera-light`: `OPEN-ALL-HOSTS` for `minimal-camera-object-profile` and
  `minimal-light-object-profile`
- `precomp`: `OPEN-ALL-HOSTS` for `minimal-precomp-layer`

Also update `index.md` and append `history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record structural all-host evidence
```

---

## Task 33: Text Animator PD-6x6 Writer Matrix Evidence

**Goal:** Upgrade text animator writer coverage from the historical W2020/W2022/W2025
`PD-3x3` matrix to a full W2020-W2025 source-and-target matrix now that native
writer targets exist for every installed AE year.

**Architecture:** Reuse the existing matrix runner without AE-open. Run every
`minimal-text-animator*.json` recipe across `-sources all -targets all` and
record the reviewed result as `PD-6x6` evidence. Do not infer host-open coverage
from this profile-diff matrix.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`

- [x] **Step 1: Run the text animator all-writer matrix**

Run:

```powershell
$recipes = Get-ChildItem examples\recipes\minimal-text-animator*.json | ForEach-Object { @('-recipe', $_.FullName) }
go run ./cmd/aepmigrate matrix @recipes -sources all -targets all -out tmp\migration_matrix_text_animators_all_6x6 -ledger-out tmp\migration_matrix_text_animators_all_6x6\ledger.md
```

Expected: 792 total, 792 pass, 0 blocked, 0 failed, 0 skipped. This is
22 recipes x 6 source writers x 6 target writers.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, update the text animator rows from `PD-3x3` to `PD-6x6`.
Add `tmp/migration_matrix_text_animators_all_6x6/matrix.json` and its ledger to
raw artifacts. Update `index.md` and append `history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record text animator six-writer evidence
```

---

## Task 34: Dynamic Transform PD-6x6 Writer Matrix Evidence

**Goal:** Upgrade dynamic transform writer coverage from `PD-1x3` to a full
W2020-W2025 source-and-target matrix now that native writer targets exist for
every installed AE year.

**Architecture:** Reuse the matrix runner without AE-open. Run the three
dynamic transform recipes across `-sources all -targets all` and record the
reviewed result as `PD-6x6` evidence. Keep host-open coverage separate: the
existing transform-ease representative already has `OPEN-ALL-HOSTS`.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`

- [x] **Step 1: Run the dynamic transform all-writer matrix**

Run:

```powershell
go run ./cmd/aepmigrate matrix `
  -recipe examples\recipes\minimal-transform-keyframes.json `
  -recipe examples\recipes\minimal-transform-keyframe-ease.json `
  -recipe examples\recipes\minimal-transform-expression.json `
  -sources all `
  -targets all `
  -out tmp\migration_matrix_transform_all_6x6 `
  -ledger-out tmp\migration_matrix_transform_all_6x6\ledger.md
```

Expected: 108 total, 108 pass, 0 blocked, 0 failed, 0 skipped. This is
3 recipes x 6 source writers x 6 target writers.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, update the Transform Domain rows from `PD-1x3` to `PD-6x6`.
Add `tmp/migration_matrix_transform_all_6x6/matrix.json` and its ledger to raw
artifacts. Update `index.md` and append `history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record transform six-writer evidence
```

---

## Task 35: Dynamic Effect PD-6x6 Writer Matrix Evidence

**Goal:** Upgrade dynamic effect parameter writer coverage from `PD-1x3` to a
full W2020-W2025 source-and-target matrix for the current recipe-owned dynamic
effect surface.

**Architecture:** Reuse the matrix runner without AE-open. Run the four dynamic
effect recipes across `-sources all -targets all` and record the reviewed result
as `PD-6x6` evidence. Keep host-open coverage separate: the vector-keyframe
representative already has `OPEN-ALL-HOSTS`.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`

- [x] **Step 1: Run the dynamic effect all-writer matrix**

Run:

```powershell
go run ./cmd/aepmigrate matrix `
  -recipe examples\recipes\minimal-effect-layer-param.json `
  -recipe examples\recipes\minimal-effect-param-expression.json `
  -recipe examples\recipes\minimal-effect-param-keyframes.json `
  -recipe examples\recipes\minimal-effect-param-vector-keyframes.json `
  -sources all `
  -targets all `
  -out tmp\migration_matrix_effect_params_all_6x6 `
  -ledger-out tmp\migration_matrix_effect_params_all_6x6\ledger.md
```

Expected: 144 total, 144 pass, 0 blocked, 0 failed, 0 skipped. This is
4 recipes x 6 source writers x 6 target writers.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, update the Effect Domain dynamic rows from `PD-1x3` to
`PD-6x6`. Add `tmp/migration_matrix_effect_params_all_6x6/matrix.json` and its
ledger to raw artifacts. Update `index.md` and append `history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record effect six-writer evidence
```

---

## Task 36: Text Style PD-6x6 Writer Matrix Evidence

**Goal:** Upgrade text style writer coverage from `PD-1x3` to a full
W2020-W2025 source-and-target matrix for the current recipe-owned text style
surface.

**Architecture:** Reuse the matrix runner without AE-open. Run the two current
text style recipes across `-sources all -targets all` and record the reviewed
result as `PD-6x6` evidence. Keep host-open coverage separate: the existing
`minimal-text-style` W2020 representative already has `OPEN-ALL-HOSTS`.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`

- [x] **Step 1: Run the text style all-writer matrix**

Run:

```powershell
go run ./cmd/aepmigrate matrix `
  -recipe examples\recipes\minimal-text-style.json `
  -recipe examples\recipes\minimal-text-shape.json `
  -sources all `
  -targets all `
  -out tmp\migration_matrix_text_style_all_6x6 `
  -ledger-out tmp\migration_matrix_text_style_all_6x6\ledger.md
```

Expected: 72 total, 72 pass, 0 blocked, 0 failed, 0 skipped. This is
2 recipes x 6 source writers x 6 target writers.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, update the Text Domain text style row from `PD-1x3` to
`PD-6x6`. Add `tmp/migration_matrix_text_style_all_6x6/matrix.json` and its
ledger to raw artifacts. Update `index.md` and append `history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record text style six-writer evidence
```

---

## Task 37: Layer Mask PD-6x6 Writer Matrix Evidence

**Goal:** Upgrade layer mask writer coverage from `PD-1x3` to a full
W2020-W2025 source-and-target matrix for the current recipe-owned layer mask
surface.

**Architecture:** Reuse the matrix runner without AE-open. Run the current
`minimal-layer-mask` recipe across `-sources all -targets all` and record the
reviewed result as `PD-6x6` evidence. Keep host-open coverage separate: the
existing W2020 representative already has `OPEN-ALL-HOSTS`.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`
- Modify: `flightdeck/cockpit.md`

- [x] **Step 1: Run the layer mask all-writer matrix**

Run:

```powershell
go run ./cmd/aepmigrate matrix `
  -recipe examples\recipes\minimal-layer-mask.json `
  -sources all `
  -targets all `
  -out tmp\migration_matrix_layer_mask_all_6x6 `
  -ledger-out tmp\migration_matrix_layer_mask_all_6x6\ledger.md
```

Expected: 36 total, 36 pass, 0 blocked, 0 failed, 0 skipped. This is
1 recipe x 6 source writers x 6 target writers.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, update the Layer Domain layer mask row from `PD-1x3` to
`PD-6x6`. Add `tmp/migration_matrix_layer_mask_all_6x6/matrix.json` and its
ledger to raw artifacts. Update `index.md`, `cockpit.md`, and append
`history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record layer mask six-writer evidence
```

---

## Task 38: Classic Track Matte Writer Matrix Boundary Evidence

**Goal:** Review classic track matte writer coverage across W2020-W2025
sources and targets, and record the exact downgrade boundary instead of
inferring full `PD-6x6` support.

**Architecture:** Reuse the matrix runner without AE-open. Run the current
`minimal-layer-track-matte` recipe across `-sources all -targets all` and
record the reviewed result. AE2025-authored classic track matte can surface as
an explicit matte source in the profile; lower targets must remain blocked if
the conversion cannot preserve that source binding.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`
- Modify: `flightdeck/cockpit.md`

- [x] **Step 1: Run the classic track matte all-writer matrix**

Run:

```powershell
go run ./cmd/aepmigrate matrix `
  -recipe examples\recipes\minimal-layer-track-matte.json `
  -sources all `
  -targets all `
  -out tmp\migration_matrix_layer_track_matte_all_6x6 `
  -ledger-out tmp\migration_matrix_layer_track_matte_all_6x6\ledger.md
```

Actual: 36 total, 31 pass, 5 blocked, 0 failed, 0 skipped. The blocked cases
are AE2025 source into W2020-W2024 targets; `convert_report.json` records
`comps["Main"].layers["Fill"].matte_ref` blocked because explicit matte source
requires AE2025 in the current writer contract.

- [x] **Step 2: Update reviewed validation summary**

Update the Layer Domain classic track matte row from `PD-1x3` to the reviewed
31/36 writer boundary. Add
`tmp/migration_matrix_layer_track_matte_all_6x6/matrix.json` and its ledger to
raw artifacts. Update `index.md`, `cockpit.md`, and append `history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record track matte writer boundary
```

---

## Task 39: Shape Gradient Stroke PD-6x6 Writer Matrix Evidence

**Goal:** Upgrade shape gradient stroke writer coverage from `PD-1x3` to a
full W2020-W2025 source-and-target matrix for all current recipe-owned
gradient stroke variants.

**Architecture:** Reuse the matrix runner without AE-open. Run the four
current gradient stroke recipes across `-sources all -targets all` and record
the reviewed result as `PD-6x6` evidence. Keep host-open coverage separate:
the existing all-host matrix already opens every current W2020 gradient-stroke
variant in H2020-H2025.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`
- Modify: `flightdeck/cockpit.md`

- [x] **Step 1: Run the gradient stroke all-writer matrix**

Run:

```powershell
go run ./cmd/aepmigrate matrix `
  -recipe examples\recipes\minimal-shape-gradient-stroke.json `
  -recipe examples\recipes\minimal-shape-gradient-stroke-alpha-stops.json `
  -recipe examples\recipes\minimal-shape-gradient-stroke-highlight.json `
  -recipe examples\recipes\minimal-shape-gradient-stroke-style.json `
  -sources all `
  -targets all `
  -out tmp\migration_matrix_shape_gradient_stroke_all_6x6 `
  -ledger-out tmp\migration_matrix_shape_gradient_stroke_all_6x6\ledger.md
```

Expected: 144 total, 144 pass, 0 blocked, 0 failed, 0 skipped. This is
4 recipes x 6 source writers x 6 target writers.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, update the Shape Domain gradient stroke row from `PD-1x3` to
`PD-6x6`. Add
`tmp/migration_matrix_shape_gradient_stroke_all_6x6/matrix.json` and its
ledger to raw artifacts. Update `index.md`, `cockpit.md`, and append
`history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record gradient stroke six-writer evidence
```

---

## Task 40: Static Effect PD-6x6 Writer Matrix Evidence

**Goal:** Upgrade static supported effect writer coverage from the latest full
`PD-1x6` boundary to a focused full W2020-W2025 source-and-target matrix for
the current recipe-owned static effect surface.

**Architecture:** Reuse the matrix runner without AE-open. Run the two static
effect recipes that actually materialize effects and static params:
`minimal-text-effect` and `minimal-adjustment-layer`. Do not include
`minimal-default-adjustment-layer` in this slice because it validates
adjustment-layer creation without effect params.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`
- Modify: `flightdeck/cockpit.md`

- [x] **Step 1: Run the static effect all-writer matrix**

Run:

```powershell
go run ./cmd/aepmigrate matrix `
  -recipe examples\recipes\minimal-text-effect.json `
  -recipe examples\recipes\minimal-adjustment-layer.json `
  -sources all `
  -targets all `
  -out tmp\migration_matrix_static_effects_all_6x6 `
  -ledger-out tmp\migration_matrix_static_effects_all_6x6\ledger.md
```

Expected: 72 total, 72 pass, 0 blocked, 0 failed, 0 skipped. This is
2 recipes x 6 source writers x 6 target writers.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, update the Effect Domain static supported effects row from
latest full `PD-1x6` boundary evidence to focused `PD-6x6` evidence. Add
`tmp/migration_matrix_static_effects_all_6x6/matrix.json` and its ledger to raw
artifacts. Update `index.md`, `cockpit.md`, and append `history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record static effects six-writer evidence
```

---

## Task 41: Auto-Orient PD-6x6 Writer Matrix Evidence

**Goal:** Upgrade the auto-orient migration evidence from the latest full
`PD-1x6` boundary to a focused full W2020-W2025 source-and-target matrix.

**Architecture:** Reuse the matrix runner without AE-open. Run
`minimal-layer-auto-orient` across `-sources all -targets all` and record the
reviewed result as `PD-6x6` evidence. This remains separate from the dynamic
transform 6x6 matrix because the recipe also validates layer auto-orient state.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`
- Modify: `flightdeck/cockpit.md`

- [x] **Step 1: Run the auto-orient all-writer matrix**

Run:

```powershell
go run ./cmd/aepmigrate matrix `
  -recipe examples\recipes\minimal-layer-auto-orient.json `
  -sources all `
  -targets all `
  -out tmp\migration_matrix_auto_orient_all_6x6 `
  -ledger-out tmp\migration_matrix_auto_orient_all_6x6\ledger.md
```

Expected: 36 total, 36 pass, 0 blocked, 0 failed, 0 skipped. This is
1 recipe x 6 source writers x 6 target writers.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, update the Transform Domain auto-orient row from latest full
`PD-1x6` boundary evidence to focused `PD-6x6` evidence. Add
`tmp/migration_matrix_auto_orient_all_6x6/matrix.json` and its ledger to raw
artifacts. Update `index.md`, `cockpit.md`, and append `history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record auto-orient six-writer evidence
```

---

## Task 42: Text Baseline PD-6x6 Writer Matrix Evidence

**Goal:** Upgrade text layer baseline migration evidence from the latest full
`PD-1x6` boundary to a focused full W2020-W2025 source-and-target matrix.

**Architecture:** Reuse the matrix runner without AE-open. Run
`minimal-default-text-layer` and `minimal-default-text-static-transform` across
`-sources all -targets all` and record the reviewed result as `PD-6x6`
evidence. Keep this separate from text style and text animator matrices, which
already have focused `PD-6x6` evidence.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`
- Modify: `flightdeck/cockpit.md`

- [x] **Step 1: Run the text baseline all-writer matrix**

Run:

```powershell
go run ./cmd/aepmigrate matrix `
  -recipe examples\recipes\minimal-default-text-layer.json `
  -recipe examples\recipes\minimal-default-text-static-transform.json `
  -sources all `
  -targets all `
  -out tmp\migration_matrix_text_baseline_all_6x6 `
  -ledger-out tmp\migration_matrix_text_baseline_all_6x6\ledger.md
```

Expected: 72 total, 72 pass, 0 blocked, 0 failed, 0 skipped. This is
2 recipes x 6 source writers x 6 target writers.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, update the Text Domain text layer baseline row from latest
full `PD-1x6` boundary evidence to focused `PD-6x6` evidence. Add
`tmp/migration_matrix_text_baseline_all_6x6/matrix.json` and its ledger to raw
artifacts. Update `index.md`, `cockpit.md`, and append `history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record text baseline six-writer evidence
```

---

## Task 43: Structural Representative PD-6x6 Writer Matrix Evidence

**Goal:** Upgrade the structural representative writer coverage from latest
full `PD-1x6` boundary evidence to focused W2020-W2025 source-and-target
evidence for the same representatives that already have all-host open evidence.

**Architecture:** Reuse the matrix runner without AE-open. Run the five
structural representatives from the all-host matrix across `-sources all
-targets all`: project display settings, comp object profile, camera object
profile, light object profile, and precomp layer. Record the result as
representative `PD-6x6` evidence, not as exhaustive project/comp/camera/light
coverage.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`
- Modify: `flightdeck/cockpit.md`

- [x] **Step 1: Run the structural representative all-writer matrix**

Run:

```powershell
go run ./cmd/aepmigrate matrix `
  -recipe examples\recipes\minimal-project-display-settings.json `
  -recipe examples\recipes\minimal-comp-object-profile.json `
  -recipe examples\recipes\minimal-camera-object-profile.json `
  -recipe examples\recipes\minimal-light-object-profile.json `
  -recipe examples\recipes\minimal-precomp-layer.json `
  -sources all `
  -targets all `
  -out tmp\migration_matrix_structural_representatives_all_6x6 `
  -ledger-out tmp\migration_matrix_structural_representatives_all_6x6\ledger.md
```

Expected: 180 total, 180 pass, 0 blocked, 0 failed, 0 skipped. This is
5 recipes x 6 source writers x 6 target writers.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, update the Other Domains rows for project, comp,
camera-light, and precomp to cite focused representative `PD-6x6` evidence
alongside their existing `OPEN-ALL-HOSTS` evidence. Add
`tmp/migration_matrix_structural_representatives_all_6x6/matrix.json` and its
ledger to raw artifacts. Update `index.md`, `cockpit.md`, and append
`history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record structural representative writer evidence
```

---

## Task 44: Camera PD-6x6 Writer Matrix Evidence

**Goal:** Upgrade current recipe-owned camera layer and camera option writer
coverage from latest full `PD-1x6` boundary evidence to focused W2020-W2025
source-and-target evidence.

**Architecture:** Reuse the matrix runner without AE-open. Run every current
`minimal-camera-*.json` recipe across `-sources all -targets all`. Record the
result as camera `PD-6x6` evidence inside the camera-light domain; do not
broaden AE host-open claims beyond the existing `minimal-camera-object-profile`
representative.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`
- Modify: `flightdeck/cockpit.md`

- [x] **Step 1: Run the camera all-writer matrix**

Run:

```powershell
$recipes = Get-ChildItem examples\recipes\minimal-camera-*.json | ForEach-Object { @('-recipe', $_.FullName) }
go run ./cmd/aepmigrate matrix @recipes `
  -sources all `
  -targets all `
  -out tmp\migration_matrix_camera_all_6x6 `
  -ledger-out tmp\migration_matrix_camera_all_6x6\ledger.md
```

Expected: 540 total, 540 pass, 0 blocked, 0 failed, 0 skipped. This is
15 recipes x 6 source writers x 6 target writers.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, update the Other Domains camera-light row to cite focused
camera `PD-6x6` evidence alongside the existing structural representative
evidence. Add `tmp/migration_matrix_camera_all_6x6/matrix.json` and its ledger
to raw artifacts. Update `index.md`, `cockpit.md`, and append `history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record camera six-writer evidence
```

---

## Task 45: Light PD-6x6 Writer Matrix Evidence

**Goal:** Upgrade current recipe-owned light layer, light option, and light
source writer coverage from latest full `PD-1x6` boundary evidence to focused
W2020-W2025 source-and-target evidence.

**Architecture:** Reuse the matrix runner without AE-open. Run every current
`minimal-light-*.json` recipe across `-sources all -targets all`. Record the
result as light `PD-6x6` evidence inside the camera-light domain; do not
broaden AE host-open claims beyond the existing `minimal-light-object-profile`
representative.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`
- Modify: `flightdeck/cockpit.md`

- [x] **Step 1: Run the light all-writer matrix**

Run:

```powershell
$recipes = Get-ChildItem examples\recipes\minimal-light-*.json | ForEach-Object { @('-recipe', $_.FullName) }
go run ./cmd/aepmigrate matrix @recipes `
  -sources all `
  -targets all `
  -out tmp\migration_matrix_light_all_6x6 `
  -ledger-out tmp\migration_matrix_light_all_6x6\ledger.md
```

Expected: 504 total, 504 pass, 0 blocked, 0 failed, 0 skipped. This is
14 recipes x 6 source writers x 6 target writers.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, update the Other Domains camera-light row to cite focused
light `PD-6x6` evidence alongside focused camera evidence and the existing
host-open representatives. Add `tmp/migration_matrix_light_all_6x6/matrix.json`
and its ledger to raw artifacts. Update `index.md`, `cockpit.md`, and append
`history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record light six-writer evidence
```

---

## Task 46: Project PD-6x6 Writer Matrix Evidence

**Goal:** Upgrade current recipe-owned project setting writer coverage from
structural representative `PD-6x6` evidence to focused W2020-W2025
source-and-target evidence for every current `minimal-project-*` recipe.

**Architecture:** Reuse the matrix runner without AE-open. Run all four current
project recipes across `-sources all -targets all`. Record the result as
project `PD-6x6` evidence while keeping AE host-open scope at the existing
`minimal-project-display-settings` representative.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`
- Modify: `flightdeck/cockpit.md`

- [x] **Step 1: Run the project all-writer matrix**

Run:

```powershell
go run ./cmd/aepmigrate matrix `
  -recipe examples\recipes\minimal-project-display-settings.json `
  -recipe examples\recipes\minimal-project-bits-per-channel.json `
  -recipe examples\recipes\minimal-project-linear-color.json `
  -recipe examples\recipes\minimal-project-preferences.json `
  -sources all `
  -targets all `
  -out tmp\migration_matrix_project_all_6x6 `
  -ledger-out tmp\migration_matrix_project_all_6x6\ledger.md
```

Expected: 144 total, 144 pass, 0 blocked, 0 failed, 0 skipped. This is
4 recipes x 6 source writers x 6 target writers.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, update the Other Domains project row to cite focused project
`PD-6x6` evidence for all four current project recipes. Add
`tmp/migration_matrix_project_all_6x6/matrix.json` and its ledger to raw
artifacts. Update `index.md`, `cockpit.md`, and append `history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record project six-writer evidence
```

---

## Task 47: Comp PD-6x6 Writer Matrix Evidence

**Goal:** Upgrade current recipe-owned comp setting writer coverage from
structural representative `PD-6x6` evidence to focused W2020-W2025
source-and-target evidence for every current `minimal-comp-*` recipe.

**Architecture:** Reuse the matrix runner without AE-open. Run all 18 current
comp recipes across `-sources all -targets all`. Record the result as comp
`PD-6x6` evidence while keeping AE host-open scope at the existing
`minimal-comp-object-profile` representative.

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/work/aep-understanding-generation/history.md`
- Modify: `flightdeck/work/aep-understanding-generation/versioned-aep-migration-remaining-blockers-plan.md`
- Modify: `flightdeck/cockpit.md`

- [x] **Step 1: Run the comp all-writer matrix**

Run:

```powershell
$recipes = Get-ChildItem examples\recipes\minimal-comp-*.json | ForEach-Object { @('-recipe', $_.FullName) }
go run ./cmd/aepmigrate matrix @recipes `
  -sources all `
  -targets all `
  -out tmp\migration_matrix_comp_all_6x6 `
  -ledger-out tmp\migration_matrix_comp_all_6x6\ledger.md
```

Expected: 648 total, 648 pass, 0 blocked, 0 failed, 0 skipped. This is
18 recipes x 6 source writers x 6 target writers.

- [x] **Step 2: Update reviewed validation summary**

If Step 1 passes, update the Other Domains comp row to cite focused comp
`PD-6x6` evidence for all 18 current comp recipes. Add
`tmp/migration_matrix_comp_all_6x6/matrix.json` and its ledger to raw artifacts.
Update `index.md`, `cockpit.md`, and append `history.md`.

- [x] **Step 3: Verification and commit**

Run:

```powershell
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
git diff --check
```

Read:

```powershell
Get-Content C:\Users\yl\.flightdeck\knowledge\git\commits.md -Raw
Get-Content flightdeck\knowledge\workflow\verify.md -Raw
```

Commit:

```text
docs(aepmigrate): record comp six-writer evidence
```

---

## Task 48: Layer / Shape / Precomp Batch Matrix Evidence

**Goal:** Batch the remaining structural/layer/shape writer evidence updates
into one validation pass and one commit.

| Domain | Scope | Artifact | Result |
| --- | --- | --- | --- |
| precomp | all 1 current `minimal-precomp-*` recipe | `tmp/migration_matrix_precomp_all_6x6/matrix.json` | 36 total, 36 pass, 0 blocked, 0 failed, 0 skipped |
| layer | all 16 current `minimal-layer-*` recipes | `tmp/migration_matrix_layer_all_6x6/matrix.json` | 576 total, 536 pass, 10 blocked, 0 failed, 30 skipped |
| shape | all 27 current `minimal-shape-*` recipes | `tmp/migration_matrix_shape_all_6x6/matrix.json` | 972 total, 972 pass, 0 blocked, 0 failed, 0 skipped |

Layer non-pass cases are the known matte contract boundary: `minimal-layer-explicit-matte`
has 30 source-contract skips and 5 AE2025-source downgrade blocks; `minimal-layer-track-matte`
has 5 AE2025-source downgrade blocks. No layer matrix case failed.

- [x] Run batch matrices.
- [x] Update validation summary, index, history, and cockpit once for the batch.
- [x] Run recurring verification and commit once.
