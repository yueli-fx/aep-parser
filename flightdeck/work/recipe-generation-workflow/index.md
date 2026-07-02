# Index — recipe-generation-workflow

## State

This package owns from-scratch project generation and the recipe IR workflow:
comp/layer execution strategy, Phase 5 replication acceptance, Phase 6 recipe
IR, and source-vs-clone render comparison.

It is currently parked. Versioned migration uses recipe fixtures as evidence,
but ordinary migration batches should resume from
`../versioned-aep-migration/index.md`.

## Next

Resume when adding new recipe-owned authoring families, changing recipe/profile
contracts, or tightening from-scratch render comparison.

## Read now

- `comp-recipe-execution-strategy.md`
- `layer-recipe-execution-strategy.md`
- `phase6-recipe-ir-plan.md`
- `phase6-recipe-ir-acceptance.md`

## Read if

- `phase5-plan.md`, `phase5-booyah-acceptance.md`, and
  `phase5-non-booyah-acceptance.md` — if revisiting replication acceptance.
- `phase6-plan.md` and `phase6-booyah-render-compare.md` — if revisiting
  render compare or Booyah-driven recipe validation.

## Progress

Done:

- Recipe IR grew into broad comp, layer, transform, effect, shape, text, camera,
  light, precomp, and Essential Graphics fixture coverage.
- Recipe outputs now provide many of the representative fixtures used by the
  versioned migration matrix.

## Open questions

- Which missing authoring family should become the next recipe domain once
  migration evidence needs it.
