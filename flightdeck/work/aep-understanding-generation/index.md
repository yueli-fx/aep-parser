# Index — aep-understanding-generation

## State

This package is now an umbrella/router only. The old monolithic AEP
understanding package was split into focused topic packages so each recovery
point has one owner, one state file, and one next action.

The historical monolith remains in `history.md` for archaeology. Do not resume
execution from this package unless the work is explicitly about cross-package
coordination.

## Next

Resume the focused topic package that matches the task:

- `../versioned-aep-migration/index.md` — active mainline: AE2020-AE2025
  version migration, coverage JSON, matrix/host-open gates, and batch inventory.
- `../aep-profile-understanding/index.md` — parser/profile/diff/render-oracle
  understanding foundation.
- `../recipe-generation-workflow/index.md` — from-scratch generation, recipe IR,
  comp/layer execution strategy, and render-compare workflow.
- `../fx-technique-internalization/index.md` — reference project
  internalization, technique facts, portraits, and reports.
- `../asset-registry-cleanup/index.md` — data/example/test/tmp ownership,
  registry ledgers, cleanup policy, and generated artifact boundaries.

## Read now

- `history.md` — only if reconstructing old completed-slice history.

## Read if

- `../versioned-aep-migration/goal.md` — when the user asks to run the mainline
  AEP-understanding goal.
- `../versioned-aep-migration/current.md` — when recovering the active versioned
  migration state.

## Progress

Split complete:

- versioned migration state and JSON truth sources moved to
  `work/versioned-aep-migration/`.
- profile foundation plans moved to `work/aep-profile-understanding/`.
- recipe/generation plans moved to `work/recipe-generation-workflow/`.
- technique facts/portrait/report files moved into
  `work/fx-technique-internalization/`.
- data and generated-output governance now has a dedicated
  `work/asset-registry-cleanup/` package.

## Open questions

- Whether old `history.md` should later be cold-archived once all useful
  historical facts have been absorbed by the focused packages or knowledge
  files.
