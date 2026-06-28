# Index — aep-understanding-generation

## State

This work package defines the next long-running direction: turn `aep-parser`
from a parser/writer plus one-off showcase generator into an AEP understanding,
replication-diagnostics, and future generation system.

The current codebase already has the key raw materials:
- `cmd/aepdissect -json` can emit a shallow project profile.
- `internal/scene.WriteJSON` can export a broad parsed scene view.
- `flightdeck/showcase/booyah-clone` contains real replication and render-diff
  tooling, but much of it is bespoke to one project.
- `docs/capabilities.md` is the write-capability truth source.

The gap is not "the project cannot do this"; the gap is that the pieces are not
yet a stable, reusable pipeline.

## Next

Review `design.md`. If accepted, write an implementation plan that starts with
Phase 0 export-surface audit and contract lock before extracting
`internal/profile` or expanding generation.

## Read now

- `design.md` — full spec and phased direction.
- `phase0-audit.md` — current export-surface audit.
- `profile-coverage.md` — coverage matrix and Phase 1 field admission list.
- `profile-contract.md` — evidence ladder, stable path object, and Phase 1
  entry gate.
- `plan.md` — Phase 0 execution plan.
- `phase1-plan.md` — Phase 1 profile extraction plan and verification record.
- `phase2-plan.md` — Phase 2 structural/semantic diff plan.
- `flightdeck/knowledge/techniques/understand-a-project.md` — existing
  reference-project internalization workflow.
- `flightdeck/knowledge/techniques/fx-techniques.md` — current technique
  library.
- `docs/capabilities.md` — current write capability matrix.

## Progress

Done:
- Direction scoped from current codebase.
- Existing support and gaps identified.
- External review feedback triaged into the spec: Phase 0 audit, evidence
  ladder, stable path contract, profile/export boundary, render oracle limits,
  and generation readiness gate.
- Phase 0 audit artifacts written: export-surface audit, profile coverage
  matrix, and profile contract.
- Phase 1 implemented `internal/profile` and moved `aepdissect -json` onto the
  stable profile builder.
- Phase 1 self-review fixed profile gaps for track matte refs and parsed
  in/out points.

Current:
- Phase 2 structural/semantic diff is active.
- Implement `internal/profilediff`, `cmd/aepdiff`, and JSON ignore rules before
  any render oracle or generation work.

## Open questions

- How broad the first Booyah diff report should be before Phase 3 render oracle
  exists.
