# Mainline Domain Batch Execution Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Continue AE understanding / versioned AEP migration by executing coherent domain batches, not single-field increments.

**Architecture:** Treat JSON ledgers as the source of truth and matrix outputs as rebuildable evidence. Each execution cycle selects one domain package, inventories the whole package, implements all evidence-backed members together, records unsupported members as explicit boundaries, runs AE2020-AE2025 matrix coverage, then updates JSON once and commits once.

**Tech Stack:** Go, `cmd/aepmigrate`, `cmd/aepregistry`, recipe JSON, capability registry JSON, AE2020-AE2025 installed hosts.

---

## Required Protocol

Before executing this plan, read and follow
`flightdeck/work/aep-understanding-generation/mainline-goal-protocol.md`.
If the requested goal is field-level rather than package-level, stop and
rewrite the goal instead of making source changes.

## Non-Negotiable Execution Rules

- No single-field goal targets. A valid target is a domain package such as `essential_graphics_controllers`, `effect_controls_static_values`, `text_domain_residuals`, `shape_domain_residuals`, or `layer_precomp_domain_residuals`.
- No per-run Markdown churn. During execution update only the batch JSON/state files, source/tests/recipes, and final generated registry reports required by gates.
- No tiny commits. One coherent domain batch should normally produce one commit. Split only when the first commit is infrastructure and the second is the domain implementation.
- No guessed binary support. If a control/property has no parser evidence, AE-native fixture, template, or generated host proof, mark it as `boundary_pending_evidence` in JSON rather than implementing it.
- Matrix axis is AE2020-AE2025: `AE2020`, `AE2021`, `AE2022`, `AE2023`, `AE2024`, `AE2025`.
- Endpoint inference is allowed only when direct endpoint hosts pass: if `AE2020` and `AE2025` pass, mark `AE2021-AE2024` as inferred by endpoint in the host-open evidence, not as unverified pass.
- Before execution begins, reconcile the current RED-only Essential Graphics point edits. They must be either absorbed into the Essential Graphics domain batch or reverted before another package starts.

## Source Of Truth Files

- Read: `flightdeck/work/versioned-aep-migration/mainline-spec.json`
- Read/modify: `flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json`
- Read/modify: `flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json`
- Read/modify as needed: `internal/aepmigrate/version_capability_ledger.json`
- Create/modify during execution: `flightdeck/work/versioned-aep-migration/domain-batch-inventory.json`
- Rebuildable reports: `tmp/registry_coverage*.json`, `tmp/migration_matrix_*`, `tmp/host_open_gaps/*`

## Current Baseline

- Branch: `mainline/aep-understanding`
- Existing coverage-axis checkpoint: `150 atom rows / 5400 cells / 5365 pass / 5 blocked / 0 failed / 30 skipped`
- Completed package: `text-animator-remaining-value-keyframes`, matrix `tmp/migration_matrix_text_animators_all_6x6/matrix.json`, totals `1044/1044 pass`
- Current dirty state from interrupted RED-only attempt:
  - `examples/recipes/minimal-essential-graphics-point-controller.json`
  - `internal/aepmigrate/convert_essential_graphics_test.go`
  - `internal/recipe/compiler_test.go`

---

## Task 1: Freeze The Workspace And Reconcile Interrupted RED Work

**Files:**
- Inspect: `examples/recipes/minimal-essential-graphics-point-controller.json`
- Inspect: `internal/aepmigrate/convert_essential_graphics_test.go`
- Inspect: `internal/recipe/compiler_test.go`

- [ ] **Step 1: Inspect dirty state**

Run:

```powershell
git status --short
git diff -- examples/recipes/minimal-essential-graphics-point-controller.json internal/aepmigrate/convert_essential_graphics_test.go internal/recipe/compiler_test.go
```

Expected: only the interrupted Essential Graphics point RED changes are dirty.

- [ ] **Step 2: Decide package ownership**

If the next package is `essential_graphics_controllers`, keep the dirty RED changes and absorb them into Task 5.

If the next package is not `essential_graphics_controllers`, revert only these interrupted RED changes:

```powershell
git restore -- internal/aepmigrate/convert_essential_graphics_test.go internal/recipe/compiler_test.go
Remove-Item -LiteralPath examples/recipes/minimal-essential-graphics-point-controller.json
```

Expected: no unrelated files are touched.

- [ ] **Step 3: Confirm clean or intentional dirty state**

Run:

```powershell
git status --short
```

Expected: either clean, or only intentional files for the selected package are dirty.

---

## Task 2: Create The Domain Batch Inventory JSON

**Files:**
- Create/modify: `flightdeck/work/versioned-aep-migration/domain-batch-inventory.json`
- Read: `flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json`
- Read: `flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json`

- [ ] **Step 1: Create the inventory skeleton**

Create `flightdeck/work/versioned-aep-migration/domain-batch-inventory.json` with this shape:

```json
{
  "schema_version": 1,
  "updated_at": "2026-07-03",
  "goal": "Batch AE understanding work by domain package instead of single-field increments.",
  "version_axis": ["AE2020", "AE2021", "AE2022", "AE2023", "AE2024", "AE2025"],
  "selection_policy": {
    "minimum_package_size": "one coherent domain family",
    "forbidden_target_shape": "single field or single property unless it is the only remaining member of a documented package",
    "evidence_required_for_writer_support": ["parser evidence", "template evidence", "AE-native fixture", "generated host proof"],
    "unsupported_status": "boundary_pending_evidence"
  },
  "packages": [],
  "active_package": null,
  "last_completed_package": null
}
```

- [ ] **Step 2: Populate current package candidates**

Add at least these package records:

```json
{
  "id": "essential_graphics_controllers",
  "domain": "essential_graphics",
  "scope": "Motion Graphics Template controller exposure for effect parameters and supported source properties.",
  "current_supported": ["slider", "checkbox", "color"],
  "known_evidence": [
    "test_data/fixtures/eg_slider_controller.aep",
    "test_data/fixtures/eg_checkbox_controller.aep",
    "test_data/fixtures/eg_color_controller.aep",
    "test_data/fixtures/eg_point_controller.aep",
    "test_data/fixtures/eg_dropdown_controller.aep",
    "test_data/fixtures/eg_text_source_text.aep"
  ],
  "candidate_members": ["effect_param_slider", "effect_param_checkbox", "effect_param_color", "effect_param_point2d", "effect_param_point3d", "effect_param_angle", "source_text", "dropdown"],
  "execution_status": "candidate"
}
```

Also add package candidates for:

```json
[
  {
    "id": "effect_controls_static_values",
    "domain": "effects",
    "scope": "Expression control effects and static values across scalar, angle, boolean, color, 2D point, 3D point, and slider."
  },
  {
    "id": "text_domain_residuals",
    "domain": "text",
    "scope": "Text source/style/animator residuals after completed value-keyframe animator package."
  },
  {
    "id": "layer_precomp_domain_residuals",
    "domain": "layer_precomp",
    "scope": "Layer, precomp, source, matte, and timing residuals."
  },
  {
    "id": "shape_domain_residuals",
    "domain": "shape",
    "scope": "Shape graph, modifiers, gradients, masks, and path residuals."
  }
]
```

- [ ] **Step 3: Validate JSON parses**

Run:

```powershell
Get-Content -Raw flightdeck/work/versioned-aep-migration/domain-batch-inventory.json | ConvertFrom-Json | Out-Null
```

Expected: command exits with code 0.

---

## Task 3: Select One Active Package By Evidence And Blast Radius

**Files:**
- Modify: `flightdeck/work/versioned-aep-migration/domain-batch-inventory.json`

- [ ] **Step 1: Query current coverage by domain**

Run:

```powershell
go run ./cmd/aepregistry coverage -root . -out tmp/registry_coverage_axis.json -axis
go run ./cmd/aepregistry coverage -root . -out tmp/registry_coverage_axis_text_domain.json -axis -domain text
go run ./cmd/aepregistry coverage -root . -out tmp/registry_coverage_axis_blocked.json -axis -case-status blocked
```

Expected: global coverage passes with zero failed cells; blocked cells are known boundaries.

- [ ] **Step 2: Select the package**

Set `active_package` to the package with the best combination of:

- Existing fixtures/templates for the whole family.
- Shared implementation path.
- Matrix recipes can be grouped into one coverage record.
- Minimal risk of speculative binary writes.

Recommended first active package from current evidence: `essential_graphics_controllers`.

- [ ] **Step 3: Record selection rationale**

In `domain-batch-inventory.json`, update the chosen package:

```json
{
  "execution_status": "active",
  "selection_rationale": "Existing EG parser supports controller type readback; writer supports slider/checkbox/color; AE-native point/dropdown/text fixtures exist; point has documented CVal/CDef layout; unsupported members can be explicit boundaries instead of single-field goals."
}
```

---

## Task 4: Inventory The Entire Active Package Before Implementation

**Files:**
- Modify: `flightdeck/work/versioned-aep-migration/domain-batch-inventory.json`
- Read: active package source files and tests

- [ ] **Step 1: List source files for the active package**

For `essential_graphics_controllers`, inspect:

```powershell
rg -n "EssentialGraphics|AddEssentialProperty|EGSlider|EGCheckbox|EGColor|EGPoint|EGMultiDimensional|EGDropdown|EGText|PCTL" internal/aep internal/scene internal/serializer internal/recipe internal/aepmigrate
```

Expected: implementation paths are concentrated in `internal/serializer/mutate_essential_graphics.go`, `internal/scene/scene_essential_graphics.go`, recipe compiler materialization, and migrate tests.

- [ ] **Step 2: Classify every candidate member**

For each package member, assign one status:

- `already_supported`
- `implement_in_this_batch`
- `boundary_pending_evidence`
- `out_of_scope_for_this_package`

For `essential_graphics_controllers`, use this starting classification:

```json
{
  "member_status": {
    "effect_param_slider": "already_supported",
    "effect_param_checkbox": "already_supported",
    "effect_param_color": "already_supported",
    "effect_param_point2d": "implement_in_this_batch",
    "effect_param_point3d": "boundary_pending_evidence",
    "effect_param_angle": "boundary_pending_evidence",
    "source_text": "boundary_pending_evidence",
    "dropdown": "boundary_pending_evidence"
  }
}
```

Only upgrade a member from `boundary_pending_evidence` to `implement_in_this_batch` when the implementation has direct evidence for its `CTyp`, `CVal`, `CDef`, property path JSON, and host-open behavior.

- [ ] **Step 3: Record package acceptance criteria**

For `essential_graphics_controllers`, record:

```json
{
  "acceptance": {
    "unit_tests": [
      "go test ./internal/recipe -run EssentialGraphics -count=1",
      "go test ./internal/aepmigrate -run EssentialGraphics -count=1",
      "go test ./internal/aep_test -run EssentialGraphics -count=1"
    ],
    "matrix": "AE2020-AE2025 6x6 for all EG controller recipes in the package",
    "coverage_axis": "new or updated coverage rows have zero failed cells; unsupported members are boundary rows, not missing ad-hoc notes"
  }
}
```

---

## Task 5: Implement The Active Package As One Batch

**Files for `essential_graphics_controllers`:**
- Create/modify recipes under `examples/recipes/minimal-essential-graphics-*.json`
- Modify tests: `internal/recipe/compiler_test.go`
- Modify tests: `internal/aepmigrate/convert_essential_graphics_test.go`
- Modify writer: `internal/serializer/mutate_essential_graphics.go`
- Modify aliases/docs only if new public symbols are required

- [ ] **Step 1: Write all failing package tests before production code**

For every `implement_in_this_batch` member, add recipe + compiler + migrate tests first.

For `effect_param_point2d`, the expected failure is:

```text
AddEssentialProperty: control type 6 is not supported yet
```

Run:

```powershell
go test ./internal/recipe -run EssentialGraphics -count=1
go test ./internal/aepmigrate -run EssentialGraphics -count=1
```

Expected: tests fail only for missing package support, not schema errors or malformed recipes.

- [ ] **Step 2: Implement shared writer support**

For `essential_graphics_controllers`, implement only evidence-backed CCtl value chunk builders in `internal/serializer/mutate_essential_graphics.go`.

For `PCTLTwoD`, write `scene.EGPoint` with:

```go
case PCTLTwoD:
    vals, ok := def.lastValue.([]float64)
    if !ok || len(vals) < 2 {
        return 0, nil, fmt.Errorf("AddEssentialProperty: 2D point parameter has no 2D default value")
    }
    val := make([]byte, 16)
    binary.BigEndian.PutUint64(val[0:8], math.Float64bits(vals[0]))
    binary.BigEndian.PutUint64(val[8:16], math.Float64bits(vals[1]))
    return scene.EGPoint, []*rifx.Chunk{
        leafChunk(rifx.IDCVal, append([]byte(nil), val...)),
        leafChunk(rifx.IDCDef, append([]byte(nil), val...)),
    }, nil
```

Do not implement `angle`, `point3d`, `dropdown`, or `source_text` unless Task 4 found equivalent evidence.

- [ ] **Step 3: Run package unit tests**

Run:

```powershell
go test ./internal/recipe -run EssentialGraphics -count=1
go test ./internal/aepmigrate -run EssentialGraphics -count=1
go test ./internal/aep_test -run EssentialGraphics -count=1
```

Expected: all pass.

- [ ] **Step 4: Run focused broader tests**

Run:

```powershell
go test ./internal/serializer ./internal/recipe ./internal/aepmigrate -count=1
```

Expected: all pass.

---

## Task 6: Run The Package Matrix And Update JSON Once

**Files:**
- Modify: `flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json`
- Modify: `flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json`
- Modify: `flightdeck/work/versioned-aep-migration/mainline-spec.json`
- Modify: `flightdeck/work/versioned-aep-migration/domain-batch-inventory.json`

- [ ] **Step 1: Run one package matrix**

For `essential_graphics_controllers`, run a single grouped matrix command for all EG recipes:

```powershell
go run ./cmd/aepmigrate matrix -recipe "examples/recipes/minimal-essential-graphics-*.json" -sources all -targets all -out tmp/migration_matrix_essential_graphics_controllers_all_6x6 -ledger-out tmp/migration_matrix_essential_graphics_controllers_all_6x6/ledger.md
```

Expected: matrix completes. Any non-pass must map to an explicit boundary in the package inventory and coverage JSON.

- [ ] **Step 2: Refresh registry reports**

Run:

```powershell
go run ./cmd/aepregistry coverage -root . -out tmp/registry_coverage.json
go run ./cmd/aepregistry coverage -root . -out tmp/registry_coverage_summary.json -summary
go run ./cmd/aepregistry coverage -root . -out tmp/registry_coverage_axis.json -axis
go run ./cmd/aepregistry boundaries -root . -out tmp/registry_version_boundaries.json
```

Expected: no failed registry gates.

- [ ] **Step 3: Update JSON ledgers once**

Update only one coverage record for the package, for example:

```json
{
  "id": "essential-graphics-controllers",
  "domain": "essential_graphics",
  "scope": "Essential Graphics controller exposure for supported effect parameter control types",
  "recipes": [
    "minimal-essential-graphics-controller",
    "minimal-essential-graphics-checkbox-controller",
    "minimal-essential-graphics-color-controller",
    "minimal-essential-graphics-point-controller"
  ],
  "artifact": "tmp/migration_matrix_essential_graphics_controllers_all_6x6/matrix.json",
  "ledger": "tmp/migration_matrix_essential_graphics_controllers_all_6x6/ledger.md",
  "writer_coverage": "AE2020-AE2025 sources into AE2020-AE2025 targets",
  "writer_status": "PD-6x6",
  "totals": {
    "total": 144,
    "pass": 144,
    "blocked": 0,
    "failed": 0,
    "skipped": 0
  }
}
```

Use the real matrix totals; do not copy the example numbers if they differ.

- [ ] **Step 4: Update the active package status**

In `domain-batch-inventory.json`, set:

```json
{
  "active_package": null,
  "last_completed_package": "essential_graphics_controllers"
}
```

Set the package `execution_status` to `complete` only if unit tests, matrix, and registry reports pass. Otherwise set `blocked` with a concrete boundary reason.

---

## Task 7: Run Final Gates And Commit The Batch

**Files:**
- All source/test/recipe/JSON files changed in the batch

- [ ] **Step 1: Run recurring checkpoint gates**

Run:

```powershell
go run ./cmd/aepregistry checkpoint -root . -current flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json -coverage flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json -summary tmp/migration_coverage_summary.json -include-coverage-batch
go test ./...
```

Expected: checkpoint passes; Go tests pass.

- [ ] **Step 2: Review diff as a single batch**

Run:

```powershell
git status --short
git diff --stat
git diff -- flightdeck/work/versioned-aep-migration/domain-batch-inventory.json flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json
```

Expected: diff shows one coherent package, not scattered unrelated churn.

- [ ] **Step 3: Commit once**

For the Essential Graphics package:

```powershell
git add examples/recipes internal flightdeck/work/versioned-aep-migration/domain-batch-inventory.json flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json flightdeck/work/versioned-aep-migration/mainline-spec.json
git commit -m "feat(eg): batch essential graphics controller coverage"
```

Expected: one commit contains the domain implementation, tests, recipes, and JSON ledger update.

---

## Task 8: Select The Next Batch, Do Not Start It Automatically

**Files:**
- Modify: `flightdeck/work/versioned-aep-migration/domain-batch-inventory.json`

- [ ] **Step 1: Record next candidate**

After completing a package, choose the next candidate by evidence and blast radius. Recommended order:

1. `effect_controls_static_values`
2. `text_domain_residuals`
3. `layer_precomp_domain_residuals`
4. `shape_domain_residuals`

- [ ] **Step 2: Stop before implementation**

Do not start another package in the same execution cycle unless the user explicitly resumes with a new goal against this plan.

Expected: worktree is clean, the completed package is committed, and the next package is only selected in JSON.

---

## Self-Review Checklist

- [ ] Plan uses domain packages, not single-field tasks.
- [ ] JSON is the source of truth.
- [ ] Current interrupted RED-only changes are accounted for.
- [ ] AE2020-AE2025 axis is explicit.
- [ ] Unsupported members become explicit JSON boundaries, not hidden notes.
- [ ] Commit cadence is one coherent batch, not per atom.
- [ ] No implementation is allowed without prior failing tests.
- [ ] No binary format is guessed without evidence.

## Suggested Goal Text

Use this as the next `/goal` objective:

```text
Execute flightdeck/work/aep-understanding-generation/2026-07-03-mainline-domain-batch-execution-plan.md through one complete domain batch, starting with workspace reconciliation and stopping after the batch commit plus next-batch selection.
```
