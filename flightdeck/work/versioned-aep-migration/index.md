# Index — versioned-aep-migration

## State

This is the active mainline package for AE-version-aware migration. It owns the
AE2020-AE2025 writer/source/target matrix, capability coverage ledger, current
batch state, and recurring registry gates.

Machine-readable truth sources:

- `versioned-aep-migration-current.json`
- `versioned-aep-migration-coverage.json`
- `domain-batch-inventory.json`
- `mainline-spec.json`

There is no next selected batch; the last completed batch was
`shape_domain_residuals`.

## Next

When the user says to execute the mainline goal, run the continuous package
queue from `goal.md`. Each package is still a separate batch, but intermediate
package checkpoints stay in JSON/worktree state. The goal should continue to
the next package until no candidates remain or a package is explicitly blocked,
then commit once after final validation unless the user asks otherwise.

## Read now

- `goal.md` — execution contract for one mainline package cycle.
- `mainline-goal-protocol.md` — package-cycle rules and recovery semantics.
- `current.md` — current migration commands and registry checks.
- `domain-batch-inventory.json` — package list, status, active/next package.
- `versioned-aep-migration-current.json` — current state, commands, gates.
- `versioned-aep-migration-coverage.json` — coverage ledger and evidence rows.

## Read if

- `2026-07-03-mainline-domain-batch-execution-plan.md` — if changing the batch
  execution algorithm itself.
- `versioned-aep-migration-spec.md` — if changing product scope or migration
  architecture.
- `versioned-aep-migration-convert-plan.md` — if changing conservative
  conversion behavior.
- `versioned-aep-migration-assess-plan.md` — if changing `aepmigrate assess`.
- `versioned-aep-migration-matrix-plan.md`,
  `versioned-aep-migration-validation-plan.md`,
  `versioned-aep-migration-validation-summary.md`, and
  `versioned-aep-migration-remaining-blockers-plan.md` — frozen historical
  context superseded by the JSON ledgers.

## Progress

Done:

- `cmd/aepmigrate assess`, conservative `convert`, and matrix gates exist for
  AE2020-AE2025 writer targets.
- JSON coverage/current ledgers replaced the old long Markdown execution logs
  as the active truth source.
- Current recurring no-AE matrix gate, registry checkpoint, coverage summary,
  and host-open planning commands are tracked in `current.md` and the JSON
  ledgers.
- `essential_graphics_controllers` batch is complete and committed.

Current:

- Package queue status is `complete` in `domain-batch-inventory.json`.
- Last completed package is `shape_domain_residuals`: shape family and
  ellipse-fill recipe records were re-run through AE2020-AE2025 matrix coverage
  and refreshed in the coverage JSON.
- Next step is the final closure gate and one final commit unless the user asks
  to leave the validated state uncommitted.
- Do not start a full AE-open run unless explicitly expanding the host-open
  axis or closing a specific host-open gap.

## Open questions

- Whether the next package should broaden static effect controls or first
  promote another already-clean no-AE domain into host-open evidence.
