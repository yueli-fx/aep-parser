# Internal codebase map

`internal/` is layered around AEP binary parsing/writing, typed scene state, public facade exports, verification gates, registry governance, version migration, recipes, and technique internalization.

`internal/` is no longer only the original RIFX parser stack. Treat it as five
working layers:

Repository-facing version: `internal/README.md` mirrors this map for developers
who are browsing the code without Flightdeck context.

1. Binary and value substrate:
   - `rifx` reads/writes Big-Endian RIFF/RIFX chunk trees without AEP semantics.
   - `codec` holds byte/value encoders and layout helpers.

2. Parsed scene model and byte-backed serializer:
   - `scene` owns typed runtime state (`Project`, `Composition`, `Layer`,
     properties, masks, shapes, render queue) and writer interfaces. It must not
     own concrete RIFX chunks.
   - `serializer` bridges chunks to scene and back: parse, lower, write,
     backrefs, structural mutation, and embedded template cloning.
   - `aep` is the public facade over those internals. Public API docs and
     capability tags live on exported facade comments, not inside serializer
     internals.

3. Verification and host integration:
   - `aep_test`, `aeptest`, `aehost`, `host`, and `aeoracle` are the acceptance
     and AE automation layer. Ship-gate tests under `internal/aep_test` often
     call `test_data/generators/*.jsx`.
   - `profile`, `projectindex`, `profilediff`, `gapledger`, and
     `sliceworkflow` turn parsed projects and rendered comparisons into stable
     machine-readable reports.

4. Version migration, recipe, and governance:
   - `aepmigrate` rebuilds projects for target AE versions from normalized
     profiles and verifies the result.
   - `recipe` compiles declarative recipe JSON into AEP projects using the public
     `internal/aep` facade; `recipedoc` documents that schema and capability
     coverage.
   - `apidoc`, `capindex`, and `registry` are governance surfaces: public API
     annotation vocabulary, capability lookup/indexing, ownership, evidence, and
     migration ledgers.

5. Technique internalization and local services:
   - `technique` derives higher-level facts, portraits, and explanation objects
     from normalized profiles. It does not parse or write AEP bytes directly.
   - `selfhost` builds the long-running technique/report acceptance surface.
   - `server` exposes parse/profile operations over HTTP for external consumers.

Practical routing rules:

- New public user-facing capability normally enters through `internal/aep`
  facade comments, a `serializer` implementation, `scene` state/writer contracts,
  `internal/aep_test` tests, generated docs, and capindex.
- New byte-layout knowledge belongs in `serializer`/`codec` plus a routed
  `flightdeck/knowledge/<domain>/` note when it changes future work.
- New evidence or governance truth belongs under `registry/` or
  `registry/evidence/<topic>/`, not `tmp/`.
- New temporary probes should stay outside durable truth paths and must not make
  `go test ./...` depend on scratch output.
