# Index — apidoc-tag-schema

## State

The `@tag` documentation schema is implemented: facade conversion, dual-read, docgen integration, validation, and repo-wide jargon cleanup are complete. The active remainder is optional cleanup flip(c).

Dual-read still works and `aep:cap` count is already zero, so this is non-blocking.

## Next

Optional flip(c): remove the legacy parser path and old enum maps, add guard tests and strict validation in CI, update the documentation-source note, regenerate docs, then commit the flip.

## Read now

- `plan.md` — implementation history, flip checklist, and gate commands

## Read if

- `design.md` — if changing the schema contract, validator semantics, or docgen/capindex ownership boundary
- `flightdeck/knowledge/docgen/docgen-alias-blindspot-false-green.md` — if docgen output goes empty after alias/facade changes
- `flightdeck/knowledge/docgen/docgen-nolint-jargon-leaks-into-docs.md` — before suppressing jargon in exported doc comments
- `flightdeck/knowledge/workflow/verify.md` — before committing the flip

## Progress

Done:
- `internal/apidoc` schema/parser/validator and jargon lint are in place.
- `cmd/capindex` and `cmd/docgen` consume `@tag` data.
- Facade and bulk API tags are converted; generated docs and capabilities were regenerated during the arc.
- Jargon cleanup is complete except explicitly allowed provenance/test contexts.

Current:
- Optional flip(c) remains parked.

## Open questions

- Whether to spend a separate cleanup slice on flip(c), since the dual-read state is currently working and non-blocking.
