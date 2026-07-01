# Versioned AEP Migration Remaining Blockers Implementation Plan

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
