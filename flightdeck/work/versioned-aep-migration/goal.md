# Mainline Goal Execution Entry

Use this file as the direct goal target for AE understanding mainline work.

Suggested user command:

```text
/goal execute flightdeck/work/versioned-aep-migration/goal.md
```

## Objective

Execute the AE Understanding mainline as a continuous sequence of package
cycles under the AE Understanding Mainline Goal Protocol. Do not execute
field-level work as a goal, and do not interpret one package cycle as
completion of the whole AE understanding mainline.

Default package selection:

```text
Use domain-batch-inventory.json: if active_package is set, resume it;
otherwise use next_package; otherwise choose the first candidate package.
```

Current package queue:

```text
text_domain_residuals -> layer_precomp_domain_residuals -> shape_domain_residuals
```

## Required Reading

Before making source changes, read these files:

0. If context was compacted/resumed/summarized: reload the Flightdeck preflight
   skill/protocol and reread the project Flightdeck state before continuing.
1. `flightdeck/work/versioned-aep-migration/mainline-goal-protocol.md`
2. `flightdeck/work/versioned-aep-migration/2026-07-03-mainline-domain-batch-execution-plan.md`
3. `flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json`
4. `flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json`
5. `flightdeck/work/versioned-aep-migration/mainline-spec.json`

If the requested work conflicts with the protocol, stop and rewrite the goal
instead of coding.

## Execution Contract

Run package cycles continuously until the mainline is complete or blocked.
Each package cycle must remain a coherent package batch:

1. Reconcile the workspace.
2. Establish or update `domain-batch-inventory.json`.
3. Select one active package.
4. Inventory all package members.
5. Classify each member.
6. Add failing tests for all `implement_in_this_batch` members.
7. Implement the package members as one coherent batch.
8. Run package tests.
9. Run one package matrix across AE2020-AE2025.
10. Update coverage/current/spec JSON once.
11. Run registry/checkpoint gates.
12. Commit once.
13. Select the next package in JSON.
14. Commit the completed package cycle.
15. Continue with the next selected package.

Do not combine multiple packages into one commit. Do not start field-level work
between packages. After every completed package, the workspace must be
recoverable from committed JSON state before the next package begins.

When each package cycle finishes, report it internally as:

```text
package cycle complete; mainline remains active; next package is <id>
```

When all package records are `complete`, `blocked`, or `out_of_scope` and no
`next_package` can be selected, run a final closure gate and report:

```text
mainline package queue complete; blocked packages: <ids-or-none>
```

Do not report:

```text
AE understanding complete
overall goal complete
```

unless the final closure gate proves there are no candidate/active packages and
the JSON ledger records the closed state.

## Current Workspace Reconciliation

At the time this goal file was created, the workspace included interrupted
RED-only Essential Graphics point changes:

- `examples/recipes/minimal-essential-graphics-point-controller.json`
- `internal/aepmigrate/convert_essential_graphics_test.go`
- `internal/recipe/compiler_test.go`

This note is historical. Current execution must reconcile the actual workspace
before each package cycle. If similar interrupted RED-only changes exist for a
different package, either absorb them into that package or revert only those
interrupted changes before starting another package.

## Package Inventory Requirements

Create or update:

```text
flightdeck/work/versioned-aep-migration/domain-batch-inventory.json
```

The active package must include:

```json
{
  "id": "essential_graphics_controllers",
  "domain": "essential_graphics",
  "member_status": {
    "effect_param_slider": "already_supported",
    "effect_param_checkbox": "already_supported",
    "effect_param_color": "already_supported",
    "effect_param_point2d": "implement_in_this_batch",
    "effect_param_point3d": "boundary_pending_evidence",
    "effect_param_angle": "boundary_pending_evidence",
    "source_text": "boundary_pending_evidence",
    "dropdown": "boundary_pending_evidence"
  },
  "status": "active"
}
```

Only upgrade a `boundary_pending_evidence` member to `implement_in_this_batch`
when direct evidence exists for all required binary and host-open behavior.

## Evidence Rules

Writer support requires direct evidence from at least one of:

- Parser evidence
- AE-native fixture
- Embedded serializer template
- Generated host proof
- Existing passing matrix artifact

Do not guess AEP binary layouts. If evidence is incomplete, record a boundary.

## Verification Commands

Use package-level verification. For the default package, expected command shape:

```powershell
go test ./internal/recipe -run EssentialGraphics -count=1
go test ./internal/aepmigrate -run EssentialGraphics -count=1
go test ./internal/aep_test -run EssentialGraphics -count=1
go test ./internal/serializer ./internal/recipe ./internal/aepmigrate -count=1
go run ./cmd/aepmigrate matrix -recipe "examples/recipes/minimal-essential-graphics-*.json" -sources all -targets all -out tmp/migration_matrix_essential_graphics_controllers_all_6x6 -ledger-out tmp/migration_matrix_essential_graphics_controllers_all_6x6/ledger.md
go run ./cmd/aepregistry coverage -root . -out tmp/registry_coverage.json
go run ./cmd/aepregistry coverage -root . -out tmp/registry_coverage_summary.json -summary
go run ./cmd/aepregistry coverage -root . -out tmp/registry_coverage_axis.json -axis
go run ./cmd/aepregistry boundaries -root . -out tmp/registry_version_boundaries.json
go run ./cmd/aepregistry checkpoint -root . -current flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json -coverage flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json -summary tmp/migration_coverage_summary.json -include-coverage-batch
go test ./...
```

Single-member tests are allowed only as local RED/GREEN proof, not as the final
goal result.

## JSON Update Rules

Update JSON once after package matrix evidence exists.

Allowed JSON targets:

- `flightdeck/work/versioned-aep-migration/domain-batch-inventory.json`
- `flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json`
- `flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json`
- `flightdeck/work/versioned-aep-migration/mainline-spec.json`

Do not update `history.md`, `index.md`, or `cockpit.md` during the package run
unless the user explicitly asks.

## Commit Rule

Default commit budget per package:

```text
one implementation commit
```

Optional:

```text
one separate infrastructure commit, only if required before package work
```

Do not create per-field, per-recipe, or per-matrix commits.

Default commit message for the default package:

```powershell
git commit -m "feat(eg): batch essential graphics controller coverage"
```

## Package Cycle Completion Criteria

One package cycle is complete only when:

- One package is completed or explicitly blocked.
- Every package member has a status in JSON.
- Implemented members have tests.
- Package matrix evidence exists for AE2020-AE2025.
- Coverage JSON references the package matrix.
- Unsupported members have explicit boundary reasons.
- Registry/checkpoint gates pass or the blocking gate is recorded.
- The package commit is created if implementation was completed.
- The next package is selected in JSON.

Do not mark the overall AE understanding mainline complete just because one
package cycle passes. The persistent mainline state after a successful cycle is
`ready_for_next_batch`, not `complete`, unless a separate user-approved final
closure plan proves every package is done.

## Continuous Goal Completion Criteria

This goal is complete only when:

- No package has `execution_status` of `candidate` or `active`.
- Every package is `complete`, `blocked`, or `out_of_scope`.
- Every blocked package has a concrete `blocked_reason` or boundary reason.
- The coverage/current/spec JSON files pass registry/checkpoint gates.
- `domain-batch-inventory.json` records `mainline_status` as `complete` or
  `blocked_with_boundaries`.
- The final state is committed.

## Stop Conditions

Stop and report instead of coding when:

- The work starts from a field name instead of a package.
- The package cannot be represented in JSON.
- Evidence is insufficient and would require guessing binary structure.
- Matrix evidence cannot be written back to coverage JSON.
- The diff starts spreading into repeated Markdown logging.
- Commit count would exceed the per-package budget.
- The user says the execution is drifting.

When stopped, only update the protocol/plan or redefine the package.

## Package Loop Semantics

After a successful package cycle, this same `/goal execute .../goal.md` run must:

1. Read `domain-batch-inventory.json`.
2. Use `next_package` as the new active package unless the user overrides it.
3. Treat previous completed packages as baseline evidence, not as the end of the mainline.
4. Execute the next package cycle.
5. Repeat until the continuous goal completion criteria or a stop condition is reached.
