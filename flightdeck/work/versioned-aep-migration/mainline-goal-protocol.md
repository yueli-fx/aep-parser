# AE Understanding Mainline Goal Protocol

This protocol defines how mainline AE-understanding goals must be selected,
executed, verified, checkpointed, and closed. It exists to prevent single-field
work from spreading across the five-layer framework as repeated low-value
validation.

## 1. Goal Unit

A valid goal is a mainline upgrade package, not a single field, recipe, test, or
unsupported error.

Valid goal units:

- Domain package: `text`, `layer`, `shape`, `effect`, `essential_graphics`, `precomp`
- Capability family: `text_animators`, `layer_switches`, `effect_controls`, `shape_modifiers`
- Framework upgrade: `coverage_ledger`, `matrix_runner`, `cleanup_policy`, `capability_inventory`
- Validation closure: complete one domain's AE2020-AE2025 total/detail/total coverage loop

Invalid goal units:

- Support one `layer.xxx` field
- Support one text animator property
- Run one recipe
- Fix one coverage cell
- Add one isolated matrix result

Exception: a single member may be a valid goal only when it is the final
documented blocker inside an already-declared package.

## 2. Goal Entry Requirements

Before any mainline goal starts, answer these questions in the package JSON or
in the written plan:

- Which mainline upgrade package is this?
- Which members are inside the package?
- Which members are already supported?
- Which members are implemented in this batch?
- Which members are blocked because evidence is missing?
- Which JSON file is the source of truth for this batch?
- Where will matrix evidence be written?
- What is the maximum commit count for this batch?

If these answers are missing, do not write code.

## 3. Five-Layer Framework Contract

The five layers are not five separate reasons to validate every field one by
one.

Layer responsibilities:

1. `serializer / scene / aep`: prove AEP bytes and public API can express the capability.
2. `recipe / examples`: prove the capability can be declared as a workflow.
3. `profile / aepmigrate`: prove parsed AEP data can be rebuilt.
4. `cmd/aepmigrate matrix`: prove AE2020-AE2025 source/target behavior.
5. `registry / coverage / current JSON`: prove results are queryable, summarized, and boundary-marked.

Required order:

```text
define the layer-5 upgrade package
then decide which layer-1 through layer-4 gaps must be filled
then write layer-5 evidence and status once
```

Forbidden order:

```text
pick one layer-1 field
add tests upward through every layer
call that a mainline goal
```

## 4. Batch Package Record

Every upgrade package needs a machine-readable record. Use
`flightdeck/work/versioned-aep-migration/domain-batch-inventory.json` when
available.

Minimum package shape:

```json
{
  "id": "essential_graphics_controllers",
  "domain": "essential_graphics",
  "members": {
    "slider": "already_supported",
    "checkbox": "already_supported",
    "color": "already_supported",
    "point2d": "implement_in_this_batch",
    "point3d": "boundary_pending_evidence",
    "angle": "boundary_pending_evidence"
  },
  "matrix": "tmp/migration_matrix_essential_graphics_controllers_all_6x6/matrix.json",
  "status": "active"
}
```

Allowed member statuses:

- `already_supported`
- `implement_in_this_batch`
- `verified_in_this_batch`
- `boundary_pending_evidence`
- `known_version_boundary`
- `out_of_scope`
- `blocked`

Do not use informal states such as "try", "maybe", "looks okay", or "quick check".

## 5. Verification Rules

Verification serves the package, not an isolated field.

Allowed package-level verification examples:

```powershell
go test ./internal/aepmigrate -run EssentialGraphics
go run ./cmd/aepmigrate matrix -recipe "examples/recipes/minimal-essential-graphics-*.json" -sources all -targets all -out tmp/migration_matrix_essential_graphics_controllers_all_6x6 -ledger-out tmp/migration_matrix_essential_graphics_controllers_all_6x6/ledger.md
go run ./cmd/aepregistry coverage -root . -out tmp/registry_coverage_axis.json -axis
```

Single-member tests are allowed only as local RED/GREEN steps. They are not a
mainline deliverable by themselves.

## 6. Commit Rules

The continuous mainline goal defaults to:

- No intermediate commits
- One final commit after the full package queue is validated

Package checkpoints still need to be recoverable from JSON/Markdown state before
the next package starts, but they stay in the working tree until the queue is
complete. Commit early only when the user explicitly asks for a checkpoint
commit.

Forbidden:

- One field per commit
- One matrix per commit
- One package per commit during the continuous queue
- Updating `history.md`, `index.md`, or `cockpit.md` for every validation slice
- Turning temporary validation into a pile of Markdown logs

Default changed files should be limited to:

- Source, test, and recipe files
- `flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json`
- `flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json`
- `flightdeck/work/versioned-aep-migration/mainline-spec.json` when the mainline pointer changes
- `flightdeck/work/versioned-aep-migration/domain-batch-inventory.json` when package state changes

## 7. Package Completion Definition

A package cycle is complete only when:

- Every package member has a recorded status.
- Every implemented member has tests.
- Package recipes have AE2020-AE2025 matrix evidence.
- Matrix results are reflected in coverage JSON.
- Unsupported members have explicit boundary reasons.
- Registry coverage/checkpoint gates pass.
- The working tree contains only intentional package queue changes plus any
  user-approved unrelated changes.
- No commit is required for the package checkpoint during a continuous run.

If any item is missing, the package cycle is still active or blocked, not
complete.

Completing one package cycle does not complete the whole AE understanding
mainline. After a successful package cycle, the persistent mainline status is
`ready_for_next_batch` with a selected next package when more candidates remain.

## 7.1 Continuous Goal Completion Definition

A continuous mainline goal is complete only when:

- `domain-batch-inventory.json` has no package with `execution_status` of
  `candidate` or `active`.
- Every package is `complete`, `blocked`, or `out_of_scope`.
- Every blocked package has a concrete `blocked_reason` or boundary reason.
- Registry coverage/checkpoint gates pass for the final ledger state.
- The final state is committed.

If candidates remain, continue to the next package in the same goal run after
the previous package checkpoint. Do not stop merely because one package cycle
passed, and do not commit merely because one package cycle passed.

## 8. Stop Conditions

Stop immediately when any of these happen:

- The goal is actually a single-field target.
- There is no JSON total-table entry point.
- Implementation requires guessing AEP binary structure.
- Matrix results cannot be summarized back into JSON.
- Markdown churn starts replacing JSON state.
- Commits are trending toward field-by-field or package-by-package fragments.
- The user says the execution granularity has drifted.

After stopping, only do one of these:

1. Update the protocol or execution plan.
2. Redefine the upgrade package.

Do not continue writing field-level implementation code.

## 9. Goal Wording

Good goal wording:

```text
Execute the versioned AEP migration mainline package queue. For each package in
domain-batch-inventory.json, inventory all members, implement or verify the
evidence-backed members as one batch, run AE2020-AE2025 matrix coverage, update
coverage JSON once, checkpoint the package in JSON/worktree state, then continue
to the next package until no candidate packages remain or a package is
explicitly blocked. Commit once after the full queue is validated.
```

Bad goal wording:

```text
Support EG point.
Continue finding unsupported fields.
Run the next matrix.
Fix layer.xxx.
```

## 10. Required Preflight For Future Mainline Goals

Before resuming a mainline goal, read this file and confirm the requested goal
matches this protocol. If it does not match, stop and rewrite the goal before
making source changes.

If context was compacted, resumed, or otherwise summarized, first reload the
Flightdeck preflight skill/protocol, then reread the project Flightdeck state
before continuing this mainline goal. Do not continue from memory after context
compression.
