# Codebase Hygiene Survey Spec

## Goal

Create a deliberate cleanup and knowledge-stabilization pass for the now-large
`aep-parser` codebase. The work should leave future sessions with a reliable
map of what the codebase does, which knowledge is still valid, which JSX
generators are active, and which `tmp` JSON files are disposable.

This is an investigation and planning effort first. It must not begin by
deleting files.

## Initial Findings

`internal/` has grown into multiple strata:

- Core parser/writer: `rifx`, `codec`, `scene`, `serializer`, `aep`.
- Acceptance and comparison: `aep_test`, `aeptest`, `aehost`, `host`,
  `aeoracle`, `profile`, `profilediff`, `gapledger`, `sliceworkflow`.
- Version migration and generation: `aepmigrate`, `recipe`, `recipedoc`.
- Governance: `apidoc`, `capindex`, `registry`.
- Technique internalization and services: `technique`, `selfhost`, `server`.

The old README architecture section still describes the early core stack, but
not the later migration, registry, recipe, and technique layers.

Knowledge is useful but uneven. The local knowledge base has 140 files. A quick
header scan found 44 incomplete recipe entries, mainly under `composition`,
`layer`, and `shape`. That makes preflight routing weaker because `READ WHEN`
cannot fire for those entries.

`test_data/generators` is large but not obviously stale. There are 239 JSX
scripts. A text-reference scan found about 26 scripts with no direct stem match
in code/docs/registry main files, but all `verify_v2_2_*` scripts are referenced.
`v2_2` is a project milestone label, not AE 2022, and the shape gates still use
those scripts.

`tmp` has 116 JSON files totaling about 4.0 MB. It is ignored and intended as
disposable output. Registry files contain many commands that write reports to
`tmp`, and some durable matrix JSON files record historical `tmp` paths as
generation metadata. The cleanup pass must distinguish "a tool writes a fresh
report to tmp" from "durable truth depends on an existing tmp file".

Current sanity gates pass:

- `go run ./cmd/aepregistry audit -root .`
- `go run ./cmd/aepregistry layout -root .`
- `go run ./cmd/aepregistry ownership -root .`
- `go run ./cmd/capindex -check`

## Scope

### 1. Internal Codebase Map

Produce a current map of `internal/` by responsibility, not just by package
name. The map should answer:

- What each package does.
- Which packages are public-facing through `internal/aep`.
- Which packages own binary byte layout versus typed scene state.
- Which packages are governance/reporting tools rather than core parser code.
- Where new functionality should land.

Expected outputs:

- Keep `flightdeck/knowledge/architecture/internal-codebase-map.md` current.
- Optionally update `README.md` `## 原理与分层` so humans see the modern layers
  without opening Flightdeck.
- Produce an internal package ledger for large packages with "keep", "split
  later", or "needs owner note" classification.

### 2. Knowledge Freshness And Merge Audit

Audit every local `flightdeck/knowledge/**/*.md` entry. Do not read and rewrite
blindly; classify first.

Classification:

- `valid`: still changes future action.
- `header-repair`: body still useful, but routing header is missing or weak.
- `merge-candidate`: overlaps another note enough that one entry should absorb
  the durable conclusion.
- `stale-codepath`: references moved package paths, old template locations, or
  an implementation that no longer exists.
- `archive/delete-candidate`: does not change future work after its conclusion
  has been absorbed elsewhere.

Rules:

- Knowledge must remain self-contained. Do not replace a durable conclusion with
  "see work package".
- Fix incomplete headers before doing deeper merge decisions.
- Treat old path mentions as stale only after confirming the current code path.
  Historical context can remain if the current rule is still accurate.

Expected outputs:

- A ledger under this work package listing each knowledge file and class.
- Small batches of header repair / merge commits after each group is verified.
- No mass deletion without a grepable replacement note or an explicit "no future
  action" reason.

### 3. JSX Generator Audit

Build a reference matrix for `test_data/generators/*.jsx`.

Classification:

- `active-ship-gate`: directly used by `internal/aep_test` or current showcase.
- `active-re-fixture`: used to regenerate a tracked fixture or embedded
  template.
- `historical-re`: useful as reverse-engineering provenance, but not part of a
  current gate.
- `duplicate-or-superseded`: equivalent to a newer generator or replaced by a
  broader one.
- `delete-candidate`: no reference, no tracked fixture/template provenance, and
  no knowledge note depends on it.

Rules:

- Do not delete `verify_v2_2_*` because of the name alone. It is still active in
  shape gates and is documented as a milestone label.
- Before deleting any JSX, search code/docs/knowledge/registry and run the
  relevant tests or registry ownership gate.
- If a generator only exists to document how a binary template was extracted,
  prefer preserving that provenance in knowledge before deleting the script.

Expected outputs:

- A generator ledger with path, prefix, references, produced fixture/template
  if known, class, and cleanup action.
- A first cleanup batch limited to obvious unreferenced candidates only after
  gate confirmation.

### 4. `tmp` JSON Cleanup Audit

Treat `tmp` as disposable, but prove no durable truth depends on an existing
file before deleting.

Process:

1. List `tmp/**/*.json` by prefix, size, and modified time.
2. Search `registry/`, `docs/`, `flightdeck/knowledge`, and active work packages
   for direct references to each tmp file.
3. Separate command examples that write fresh tmp reports from references that
   require reading an existing tmp file.
4. If any durable registry/current/coverage/evidence state depends on an
   existing tmp artifact, promote it to `registry/evidence/<topic>/` first.
5. Run registry gates before and after cleanup.

Expected outputs:

- A tmp cleanup ledger with `safe-delete`, `regenerate-only`, `promote-first`,
  and `keep-for-active-work` classifications.
- Removal of safe disposable JSON in a separate commit, not mixed with knowledge
  or generator changes.

## Suggested Phases

1. Baseline ledgers: generate internal, knowledge, generator, and tmp ledgers
   without modifying existing source files.
2. Routing repair: fix knowledge headers so future preflights can route
   correctly.
3. Documentation refresh: update README / architecture notes once the internal
   package map is settled.
4. Generator cleanup: prune or archive only ledger-proven JSX candidates.
5. Tmp cleanup: delete safe ignored JSON and document any promoted evidence.
6. Final gate: run registry audit/layout/ownership, capindex check, and targeted
   Go tests for touched areas.

## Acceptance Criteria

- A future session can answer "what does `internal/` do?" from one knowledge
  note and, if updated, README.
- Every knowledge file is classified; incomplete routing headers are either
  repaired or explicitly marked for deletion/merge.
- Every JSX generator has a class and an owner/reason before any deletion.
- `tmp` contains no file that a clean clone needs to explain current state or
  pass registry gates.
- Registry and capindex gates pass after each cleanup batch.

## Non-goals

- No public API behavior changes.
- No broad refactor of `serializer`, `scene`, or `aepmigrate` during the survey.
- No cleanup of `data/` user-approved fixtures.
- No deletion of `v2_2` named assets solely because the name looks old.
