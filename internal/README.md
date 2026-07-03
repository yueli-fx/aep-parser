# internal architecture

`internal/` contains the implementation stack behind the public `aep` facade.
It is not one layer of parser code; it is a set of cooperating layers for
binary I/O, typed scene state, serialization, acceptance gates, governance,
recipes, migration, and higher-level technique analysis.

## Layer Map

### Binary substrate

- `rifx` reads and writes Big-Endian RIFF/RIFX chunk trees. It should stay
  format-generic and avoid AEP semantics.
- `codec` owns low-level value encoders, decoders, chunk layout helpers, and
  shared binary constants.

Use these packages when the change is about bytes, chunks, primitive value
encoding, or stable binary layout mechanics.

### Scene model and serializer

- `scene` owns typed runtime state: `Project`, `Composition`, `Layer`,
  properties, masks, shapes, text, render queue data, and writer-facing
  contracts. It should not own concrete RIFX chunk construction.
- `serializer` bridges between RIFX chunks and scene state. It parses AEP
  bytes, lowers scene objects back to chunks, writes projects, manages backrefs,
  performs structural mutation, and owns embedded template cloning/splicing.
- `aep` is the public facade over the internal implementation. Exported API
  comments and capability tags belong here because docgen and capindex treat
  facade comments as the user-facing contract.

Most user-visible write support touches all three: facade API in `aep`, typed
state or writer contracts in `scene`, and byte-level implementation in
`serializer`.

### Acceptance and host integration

- `aep_test` contains the main black-box, round-trip, and AE ship-gate tests.
- `aeptest`, `aehost`, `host`, and `aeoracle` provide AE automation, render
  requests, host execution, and test helpers.
- `profile`, `projectindex`, `profilediff`, `gapledger`, and `sliceworkflow`
  turn parsed projects, comparisons, and rendered evidence into stable reports.

Use this layer when a change needs AE acceptance, render/value verification,
project indexing, or regression evidence. Ship-gate tests often call
`test_data/generators/*.jsx`.

### Recipes, migration, and governance

- `recipe` compiles declarative recipe JSON into AEP projects through the public
  `internal/aep` facade.
- `recipedoc` documents recipe schema and recipe capability coverage.
- `aepmigrate` rebuilds projects for target AE versions from normalized profile
  data and verifies the result.
- `apidoc`, `capindex`, and `registry` are governance surfaces for public API
  annotations, capability lookup, ownership, evidence, layout, and migration
  ledgers.

Use these packages when the work is about declarative generation, versioned
migration, public capability accounting, or repository ownership rules.

### Technique and services

- `technique` derives higher-level facts, portraits, and explanations from
  normalized project/profile data. It should not parse or write AEP bytes
  directly.
- `selfhost` builds long-running technique and report acceptance surfaces.
- `server` exposes parse/profile operations over HTTP for external consumers.

Use this layer for product-facing analysis and local services built on top of
the parser, not for byte-layout changes.

## Dependency Direction

Keep dependencies flowing upward from low-level bytes to higher-level tools:

1. `rifx` / `codec`
2. `scene` / `serializer`
3. `aep`
4. acceptance, profile, recipe, migration, registry, and technique tools

Avoid making lower layers depend on recipe, registry, technique, or host
automation packages. If a low-level package needs a fact learned from a gate,
move the durable rule into code, tests, registry evidence, or Flightdeck
knowledge rather than importing the higher-level tool.

## Where To Put New Work

- Public API or user-facing capability: start at `internal/aep`, add or update
  doc comments/capability tags, then implement through `scene` and `serializer`.
- New binary field, stream, or template behavior: implement in `serializer` or
  `codec`, add focused tests under `internal/aep_test` or serializer tests, and
  add a knowledge note if the finding changes future reverse-engineering work.
- New declarative generation feature: add schema/compiler support in `recipe`,
  exercise public `aep` APIs, and update `recipedoc`/capindex as needed.
- Version rebuild behavior: work in `aepmigrate` and keep profile/migration
  evidence in registry-owned locations, not in `tmp/`.
- New governance or ownership rule: update `registry`, `capindex`, or `apidoc`;
  do not encode durable truth in scratch reports.
- New external or local service behavior: prefer `server` or `selfhost`, and
  keep byte parsing/writing behind the existing `aep`/profile surfaces.

## Verification Expectations

- Parser/serializer behavior should have Go tests and, for AE-sensitive write
  behavior, AE ship-gate coverage under `internal/aep_test`.
- Public API changes require regenerated docs and capindex alignment.
- Generated evidence that must survive cleanup belongs in `registry/evidence/`
  or another tracked truth-source location. `tmp/` is disposable.
- Fixture generator changes should be checked against
  `scripts/fixtures/fixtures_manifest.json`, generator provenance, and registry
  ownership before deleting or renaming files.
