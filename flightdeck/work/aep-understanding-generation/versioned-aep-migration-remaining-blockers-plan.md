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
