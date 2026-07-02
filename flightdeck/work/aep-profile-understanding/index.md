# Index — aep-profile-understanding

## State

This package owns the foundation for understanding existing `.aep` projects:
stable project profiles, structural diffs, render oracle framing, gap ledger
concepts, and the early slice workflow.

It is not the active mainline execution package right now; it is a parked
foundation package to resume when parser/profile semantics or diagnostic
architecture need work.

## Next

Resume only when changing profile contracts, diff semantics, render-oracle
evidence, or the early understanding pipeline.

## Read now

- `design.md` — high-level understanding pipeline.
- `plan.md` — original implementation plan.
- `profile-contract.md` — profile shape and contract notes.
- `profile-coverage.md` — profile coverage notes.

## Read if

- `phase0-audit.md` through `phase4-plan.md` — when reconstructing the early
  phased rollout or changing a specific foundation phase.

## Progress

Done:

- Phase 0-4 foundation was implemented and then used by later recipe,
  technique, and migration work.
- The current active migration package depends on this profile layer, but does
  not require these old plans for ordinary batch execution.

## Open questions

- Which profile gaps should become explicit capability atoms instead of staying
  as broad diagnostic notes.
