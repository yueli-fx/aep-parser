# Project operating rules

Core aep-parser operating rules: API stability, package boundaries, write/delivery gates, comments/docs, and truth-source priority.

Project rules are organized as focused topics under `flightdeck/knowledge/`.
Use the current task, Work Goal, Current and Next to select only the applicable
topics.

## API Stability

- Core Stable R/W APIs (`Open`, `FromReader`, `WriteAEP`, `WriteJSON`, shipped
  getters/setters) keep signatures, types, and JSON fields stable.
- Stable structural operations (`New*`, `Delete*`, `Insert*`, `Move*`,
  `Duplicate*`, `Add*`, `Remove*`, `SetDimensionsSeparated`) keep semantics
  stable, but call shape may move between scene methods and facade free
  functions when package boundaries require it. Mark such changes as breaking in
  the commit message and sync public docs/capindex.
- Alpha, deferred, or not ship-gated APIs may change or be removed; mark
  breaking changes honestly.
- Symbol status, verification level, gates, and boundaries are queried through
  capindex: `go run ./cmd/capindex -q "<term>"`.

## Package Boundaries

The dependency direction is:

1. `rifx` / `codec`
2. `scene` / `serializer`
3. `aep`
4. acceptance, profile, recipe, migration, registry, technique, and services

Core boundaries:

- `rifx` is format-generic RIFX container I/O and does not know AEP semantics.
- `codec` owns low-level value encoders/layout helpers.
- `scene` owns typed runtime state and writer contracts; it should not build
  concrete RIFX chunks.
- `serializer` parses/lowers/writes chunks and implements the byte-backed writer
  contracts. It may import `scene`, `codec`, and `rifx`; it must not import
  `aep`.
- `aep` is the public facade and public documentation surface.

See `internal/README.md` and `knowledge/architecture/internal-codebase-map.md`
for the current package map.

## Write Semantics

- Default writeback should be length-preserving. Length-variable exceptions
  include names, comments, expressions, font names, and text strings; parent
  LIST sizes must be reflowed correctly.
- Structural operations must maintain atomic invariants: warnings-as-failure,
  rollback to the pre-call state on failure, and AE gate coverage before being
  claimed as shipped.
- Mutations that combine project-owned objects must verify identity ownership
  before touching scene or chunk state. Matching numeric IDs is insufficient:
  a Composition, Layer, Effect, Mask, queue item, or property from another
  Project/owner can carry the same ID while referring to unrelated chunks.
- Validate every fallible back-reference, table header, count, stride, and
  target membership before the commit point. A structural method must not
  discover malformed backing data after it has started appending or splicing.
- Unknown or not-yet-modeled chunks are opaque preservation data. Parser paths
  must round-trip them byte-identically unless a deliberate mutation owns the
  bytes.
- Read-only/debug views of chunk bytes return detached snapshots. Do not expose
  live `Chunk.Data` slices through scene/facade accessors; all byte mutation
  must stay behind serializer writer methods so state and validation remain
  local to one module.
- Embedded serializer resources live under
  `internal/serializer/templates/` because Go `//go:embed` requires package
  local files.

## Delivery And Verification

- Do not claim a capability is shippable unless it is covered by AE acceptance
  at the scale and combination being used. Read
  `knowledge/workflow/delivery-contract.md` before making that claim.
- AE-sensitive structural write paths need AE 2020 + AE 2025 ship-gate coverage
  unless a documented physical version boundary makes dual-version acceptance
  impossible.
- Rendering capabilities must validate their effect surface. Value round-trip
  alone is not enough for visual behavior; use render-pixel gates or showcase
  review where appropriate.
- For test and gate commands, read `knowledge/workflow/verify.md` and
  `knowledge/workflow/re-fixture.md`.

## Documentation And Comments

- Internal implementation comments should stay sparse and explain only
  non-obvious why/constraints.
- Exported API doc comments are the public documentation source. They are
  English source text and feed `cmd/docgen`.
- Changing exported public API comments or symbols requires regenerated docs and
  capindex/docgen alignment. Free facade functions must be registered in
  `docs/docgen.json` where relevant.

## Truth-Source Priority

Use this order when facts disagree:

1. Current code and AE ship-gate tests.
2. Capindex tags via `go run ./cmd/capindex -q`.
3. Generated docs.
4. Work specs, backlog prose, and archived plans.

Do not treat `tmp/` as truth. Durable evidence belongs in tracked locations such
as `registry/evidence/<topic>/`, `data/reference/<topic>/`, or committed
fixtures/templates depending on lifecycle.
