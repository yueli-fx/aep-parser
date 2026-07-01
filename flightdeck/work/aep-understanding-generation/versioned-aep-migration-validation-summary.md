# Versioned AEP Migration Validation Summary

> Frozen historical Markdown view. Do not treat this file as the current source
> of truth. Current coverage lives in `versioned-aep-migration-coverage.json`;
> current execution state lives in `versioned-aep-migration-current.json`.

This file is the reviewed validation ledger for versioned AEP migration. Raw
matrix output under `tmp/` is supporting evidence, not the source of truth.

Legend:

- `W2020/W2021/W2022/W2023/W2024/W2025`: writer targets produced by Go.
- `H2020/H2021/H2022/H2023/H2024/H2025`: installed AE hosts used to open output.
- `PD-3x3`: profile diff passed for source writers `W2020,W2022,W2025` into target writers `W2020,W2022,W2025`.
- `PD-6x6`: profile diff passed for source writers `W2020,W2021,W2022,W2023,W2024,W2025` into target writers `W2020,W2021,W2022,W2023,W2024,W2025`.
- `PD-1x3`: profile diff passed for source writer `W2020` into target writers `W2020,W2022,W2025`.
- `PD-1x6`: profile diff passed for source writer `W2020` into target writers `W2020,W2021,W2022,W2023,W2024,W2025`.
- `OPEN-H2025`: AE2025 open smoke passed.
- `OPEN-ALL-HOSTS`: one representative writer output opened in every installed AE host from 2020 through 2025.
- `OPEN-ENDPOINTS`: one representative writer output opened in H2020 and H2025.
- `INFER-MID-HOSTS`: H2021-H2024 were not run directly; compatibility is inferred from H2020 and H2025 endpoint success and must remain marked as inference.
- `pending`: not yet formally validated at that level.

## Current Answer

No, before this ledger the project could not quickly answer "which domain and
which AE versions are validated." The current reviewed answer is:

- Text animator migration has `PD-6x6` evidence for 22 recipe-owned animator fixtures, plus `OPEN-ALL-HOSTS` evidence for two representatives: static scalar skew and animated fill-color keyframes.
- Text layer baseline migration has `PD-6x6` evidence for default text layer creation and static transform text fixtures.
- Text style migration has `PD-6x6`, `OPEN-H2025`, and `OPEN-ALL-HOSTS` evidence for the current style fixture set.
- Dynamic effect params have `PD-6x6`, `OPEN-H2025`, and one vector-keyframe `OPEN-ALL-HOSTS` representative.
- Static supported effects have `PD-6x6` evidence for the current recipe-owned static effect surface: text-layer Gaussian Blur params and adjustment-layer Gaussian Blur params.
- Dynamic transforms have `PD-6x6`, `OPEN-H2025`, and one transform-ease `OPEN-ALL-HOSTS` representative.
- Auto-orient migration has `PD-6x6` evidence for the current layer auto-orient recipe with position keyframes.
- Layer track matte migration has 31/36 all-writer matrix evidence for the classic track-matte fixture and `OPEN-ALL-HOSTS` evidence for its W2020 output. AE2025 source downgrades to W2020-W2024 are blocked by the explicit matte source contract; AE2025 source to W2025 passes. AE2025 explicit matte has `PD-1x1` evidence for AE2025 source to AE2025 target; lower-source explicit matte matrix cases are marked as source-contract skips, not conversion blockers.
- Layer mask migration has `PD-6x6` evidence for the current recipe-owned mask surface: mode/options, static outline, and path keyframes. The representative W2020 mask output has `OPEN-ALL-HOSTS` evidence.
- Layer migration has a full-family all-writer boundary matrix for all 16 current `minimal-layer-*` recipes: 576 total, 536 pass, 10 matte-contract blocked, 0 failed, 30 source-contract skipped.
- Shape migration has `PD-6x6` evidence for every current recipe-owned shape fixture. Gradient-stroke variants also have `OPEN-ALL-HOSTS` evidence.
- Structural representatives for project, comp, camera, light, and precomp have `PD-6x6` plus `OPEN-ALL-HOSTS` evidence.
- Project migration has `PD-6x6` evidence for every current recipe-owned project setting fixture.
- Comp migration has `PD-6x6` evidence for every current recipe-owned comp setting fixture.
- Camera migration has `PD-6x6` evidence for every current recipe-owned camera layer and camera option fixture.
- Light migration has `PD-6x6` evidence for every current recipe-owned light layer, option, and light-source fixture.
- Precomp migration has `PD-6x6` and `OPEN-ALL-HOSTS` evidence for the current recipe-owned precomp fixture.
- Native writer targets now exist for W2020-W2025. The narrow writer-target matrix for `minimal-comp-object-profile` is 6 total, 6 pass.
- Latest full no-AE matrix boundary is 846 total, 840 pass, 0 blocked, 0 failed, 6 source-contract skipped. It covers W2020 source into W2020-W2025 targets.
- There are no remaining blocked or failed cases in the current W2020-source full no-AE matrix. The only skipped family is AE2025 explicit matte outside the AE2025 source contract. Other migrated domains are covered by the latest full no-AE matrix boundary, but many do not yet have per-capability host-version ledgers.

## Text Domain

| Capability | Recipes / scope | Profile-diff writer coverage | AE host open coverage | Evidence |
| --- | --- | --- | --- | --- |
| Text layer baseline | `minimal-default-text-layer`, `minimal-default-text-static-transform` | `PD-6x6`: W2020-W2025 -> W2020-W2025 | pending per-capability host ledger | `tmp/migration_matrix_text_baseline_all_6x6/matrix.json`: 72 total, 72 pass |
| Text style | `minimal-text-style`, `minimal-text-shape` | `PD-6x6`: W2020-W2025 -> W2020-W2025 | `OPEN-H2025`; `OPEN-ALL-HOSTS` for `minimal-text-style` as W2020 output | `tmp/migration_matrix_text_style_all_6x6/matrix.json`: 72 total, 72 pass; `history.md` text style entry |
| Text animator opacity | `minimal-text-animator-opacity`, `minimal-text-animator-opacity-value-keyframes` | `PD-6x6`: W2020-W2025 -> W2020-W2025 | pending migration AE-host fanout | `tmp/migration_matrix_text_animators_all_6x6/matrix.json`: 792 total, 792 pass |
| Text animator position | `minimal-text-animator-position`, `minimal-text-animator-position-value-keyframes` | `PD-6x6` | pending migration AE-host fanout | same text animator 6x6 matrix |
| Text animator scale | `minimal-text-animator-scale`, `minimal-text-animator-scale-value-keyframes` | `PD-6x6` | pending migration AE-host fanout | same text animator 6x6 matrix |
| Text animator rotation Z | `minimal-text-animator-rotation`, `minimal-text-animator-rotation-value-keyframes` | `PD-6x6` | pending migration AE-host fanout | same text animator 6x6 matrix |
| Text animator rotation X/Y | `minimal-text-animator-rotation-x`, `minimal-text-animator-rotation-y` | `PD-6x6` | pending migration AE-host fanout | same text animator 6x6 matrix |
| Text animator fill color | `minimal-text-animator-color`, `minimal-text-animator-color-value-keyframes` | `PD-6x6` | `OPEN-ALL-HOSTS` for `minimal-text-animator-color-value-keyframes` W2020 output on H2020-H2025 | `tmp/migration_matrix_text_animators_all_6x6/matrix.json`; `tmp/migration_matrix_representative_all_hosts/matrix.json`: representative all-host matrix, 30 total, 30 pass |
| Text animator stroke color | `minimal-text-animator-stroke-color` | `PD-6x6` | pending migration AE-host fanout | same text animator 6x6 matrix |
| Text animator tracking | `minimal-text-animator-tracking`, `minimal-text-animator-tracking-value-keyframes` | `PD-6x6` | pending migration AE-host fanout | same text animator 6x6 matrix |
| Text animator character offset | `minimal-text-animator-character-offset`, `minimal-text-animator-character-offset-value-keyframes` | `PD-6x6` | pending migration AE-host fanout | same text animator 6x6 matrix |
| Text animator fill/stroke opacity and stroke width | `minimal-text-animator-fill-opacity`, `minimal-text-animator-stroke-opacity`, `minimal-text-animator-stroke-width` | `PD-6x6` | pending migration AE-host fanout | same text animator 6x6 matrix |
| Text animator skew | `minimal-text-animator-skew` | `PD-6x6` | `OPEN-ALL-HOSTS` for W2020 output on H2020-H2025 | `tmp/migration_matrix_text_animators_all_6x6/matrix.json`; `tmp/migration_matrix_representative_all_hosts/matrix.json`: representative all-host matrix, 30 total, 30 pass |
| Text animator range selector offset keyframes | `minimal-text-animator-range-offset` | `PD-6x6` | pending migration AE-host fanout | same text animator 6x6 matrix |

## Effect Domain

| Capability | Recipes / scope | Profile-diff writer coverage | AE host open coverage | Evidence |
| --- | --- | --- | --- | --- |
| Static supported effects | `minimal-text-effect`, `minimal-adjustment-layer`, supported static built-in effect params | `PD-6x6`: W2020-W2025 -> W2020-W2025 | pending per-capability host ledger | `tmp/migration_matrix_static_effects_all_6x6/matrix.json`: 72 total, 72 pass |
| Effect layer-ref params | `minimal-effect-layer-param` | `PD-6x6`: W2020-W2025 -> W2020-W2025 | pending per-capability host ledger | `tmp/migration_matrix_effect_params_all_6x6/matrix.json`: 144 total, 144 pass |
| Effect param expression | `minimal-effect-param-expression` | `PD-6x6` | `OPEN-H2025` | `tmp/migration_matrix_effect_params_all_6x6/matrix.json`; `history.md` dynamic effect-param entry |
| Effect scalar keyframes | `minimal-effect-param-keyframes` | `PD-6x6` | `OPEN-H2025` | `tmp/migration_matrix_effect_params_all_6x6/matrix.json`; `history.md` dynamic effect-param entry |
| Effect vector keyframes | `minimal-effect-param-vector-keyframes` | `PD-6x6` | `OPEN-H2025`; `OPEN-ALL-HOSTS` for W2020 representative | `tmp/migration_matrix_effect_params_all_6x6/matrix.json`; `history.md` dynamic effect-param entry |

## Transform Domain

| Capability | Recipes / scope | Profile-diff writer coverage | AE host open coverage | Evidence |
| --- | --- | --- | --- | --- |
| Transform keyframes | `minimal-transform-keyframes` | `PD-6x6`: W2020-W2025 -> W2020-W2025 | `OPEN-H2025` | `tmp/migration_matrix_transform_all_6x6/matrix.json`: 108 total, 108 pass |
| Transform keyframe ease | `minimal-transform-keyframe-ease` | `PD-6x6` | `OPEN-H2025`; `OPEN-ALL-HOSTS` for W2020 representative | `tmp/migration_matrix_transform_all_6x6/matrix.json`; `history.md` dynamic transform entry |
| Transform expressions | `minimal-transform-expression` | `PD-6x6` | `OPEN-H2025` | `tmp/migration_matrix_transform_all_6x6/matrix.json`; `history.md` dynamic transform entry |
| Auto-orient unlock | `minimal-layer-auto-orient` | `PD-6x6`: W2020-W2025 -> W2020-W2025 | pending per-capability host ledger | `tmp/migration_matrix_auto_orient_all_6x6/matrix.json`: 36 total, 36 pass |

## Layer Domain

| Capability | Recipes / scope | Profile-diff writer coverage | AE host open coverage | Evidence |
| --- | --- | --- | --- | --- |
| Classic track matte | `minimal-layer-track-matte` | 31/36 all-writer boundary: W2020-W2024 sources -> W2020-W2025 pass, AE2025 source -> W2025 pass, AE2025 source -> W2020-W2024 blocked by explicit matte source contract | `OPEN-ALL-HOSTS` for W2020 output on H2020-H2025 | `tmp/migration_matrix_layer_track_matte_all_6x6/matrix.json`: 36 total, 31 pass, 5 blocked; `tmp/migration_matrix_representative_all_hosts/matrix.json`: representative all-host matrix, 30 total, 30 pass |
| AE2025 explicit matte | `minimal-layer-explicit-matte` | `PD-1x1`: AE2025 source into W2025 target; AE2020 source into W2020-W2025 targets skipped by source contract | pending per-capability host ledger | `tmp/migration_matrix_explicit_matte_ae2025/matrix.json`: 1 total, 1 pass; `tmp/migration_matrix_explicit_matte_source_contract/matrix.json`: 6 total, 6 skipped |
| Layer mask | `minimal-layer-mask` | `PD-6x6`: W2020-W2025 -> W2020-W2025 | `OPEN-ALL-HOSTS` for W2020 output on H2020-H2025 | `tmp/migration_matrix_layer_mask_all_6x6/matrix.json`: 36 total, 36 pass; `tmp/migration_matrix_representative_all_hosts/matrix.json`: representative all-host matrix, 30 total, 30 pass |

## Shape Domain

| Capability | Recipes / scope | Profile-diff writer coverage | AE host open coverage | Evidence |
| --- | --- | --- | --- | --- |
| Gradient stroke | `minimal-shape-gradient-stroke`, `minimal-shape-gradient-stroke-alpha-stops`, `minimal-shape-gradient-stroke-highlight`, `minimal-shape-gradient-stroke-style` | `PD-6x6`: W2020-W2025 -> W2020-W2025 | `OPEN-ALL-HOSTS` for all current gradient-stroke W2020 outputs on H2020-H2025 | `tmp/migration_matrix_shape_gradient_stroke_all_6x6/matrix.json`: 144 total, 144 pass; `tmp/migration_matrix_shape_gradient_stroke_all_hosts/matrix.json`: 24 total, 24 pass |

## Other Domains

| Domain | Current reviewed state | Writer coverage | AE host open coverage |
| --- | --- | --- | --- |
| project | Implemented for supported project display, color, bit-depth, and preference settings | `PD-6x6` for all 4 current `minimal-project-*` recipes | `OPEN-ALL-HOSTS` for `minimal-project-display-settings` |
| comp | Implemented for stable comp settings, work area, renderer, metadata, and nested-comp options | `PD-6x6` for all 18 current `minimal-comp-*` recipes | `OPEN-ALL-HOSTS` for `minimal-comp-object-profile` |
| layer | Implemented for default layer creation, switches, refs, timing, parent/source refs, supported matte slices, and supported mask slices | all-family 6x6 boundary: 576 total, 536 pass, 10 matte-contract blocked, 30 source-contract skipped; 14 non-matte recipes pass `PD-6x6` | representative checks only |
| shape | Implemented for supported parametric graphic/filter shape slices, gradient fill/stroke slices, stroke details, and shape operators | `PD-6x6` for all 27 current `minimal-shape-*` recipes | `OPEN-ALL-HOSTS` for current gradient-stroke variants |
| camera-light | Implemented for supported camera/light options and light source refs | camera `PD-6x6` for all 15 current `minimal-camera-*` recipes; light `PD-6x6` for all 14 current `minimal-light-*` recipes | `OPEN-ALL-HOSTS` for `minimal-camera-object-profile` and `minimal-light-object-profile` |
| precomp | Implemented for precomp refs/layers | `PD-6x6` for all 1 current `minimal-precomp-*` recipe | `OPEN-ALL-HOSTS` for `minimal-precomp-layer` |

## Current Remaining Source-Contract Skips

Latest full no-AE matrix:

- Artifact: `tmp/migration_matrix_smoke_all/matrix.json`
- Recurring gate: `pwsh -File scripts/migration/verify_matrix.ps1`
- Summary: 846 total, 840 pass, 0 blocked, 0 failed, 6 skipped
- Writer coverage: W2020 source into W2020/W2021/W2022/W2023/W2024/W2025 targets

Skipped recipe groups:

| Domain | Recipe | Cases | Status |
| --- | --- | ---: | --- |
| layer | `minimal-layer-explicit-matte` | 6 | skipped in W2020-source full matrix because explicit matte requires AE2025 source contract |

## Raw Matrix Artifacts

Reviewed raw artifacts currently known:

- `tmp/migration_matrix_text_animators/matrix.json`: 66 total, 66 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020/W2022/W2025 targets for all 22 text animator recipes.
- `tmp/migration_matrix_text_animators_all_writers/matrix.json`: 198 total, 198 pass, 0 blocked, 0 failed, 0 skipped. This is W2020/W2022/W2025 source writers into W2020/W2022/W2025 target writers for all 22 text animator recipes.
- `tmp/migration_matrix_text_animators_all_6x6/matrix.json`: 792 total, 792 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for all 22 text animator recipes.
- `tmp/migration_matrix_text_baseline_all_6x6/matrix.json`: 72 total, 72 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for `minimal-default-text-layer` and `minimal-default-text-static-transform`.
- `tmp/migration_matrix_text_style_all_6x6/matrix.json`: 72 total, 72 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for `minimal-text-style` and `minimal-text-shape`.
- `tmp/migration_matrix_transform_all_6x6/matrix.json`: 108 total, 108 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for `minimal-transform-keyframes`, `minimal-transform-keyframe-ease`, and `minimal-transform-expression`.
- `tmp/migration_matrix_auto_orient_all_6x6/matrix.json`: 36 total, 36 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for `minimal-layer-auto-orient`.
- `tmp/migration_matrix_effect_params_all_6x6/matrix.json`: 144 total, 144 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for layer-ref, expression, scalar-keyframe, and vector-keyframe effect params.
- `tmp/migration_matrix_static_effects_all_6x6/matrix.json`: 72 total, 72 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for `minimal-text-effect` and `minimal-adjustment-layer`.
- `tmp/migration_matrix_text_animators_endpoint_hosts/matrix.json`: 4 total, 4 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 writer output for `minimal-text-animator-skew` and `minimal-text-animator-color-value-keyframes` opened in H2020 and H2025. H2021-H2024 were not run and are only inferred low-risk.
- `tmp/migration_matrix_representative_all_hosts/matrix.json`: 30 total, 30 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020 target with AE-open on H2020-H2025 for `minimal-text-animator-skew`, `minimal-text-animator-color-value-keyframes`, `minimal-layer-track-matte`, `minimal-layer-mask`, and `minimal-shape-gradient-stroke`.
- `tmp/migration_matrix_structural_all_hosts/matrix.json`: 30 total, 30 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020 target with AE-open on H2020-H2025 for `minimal-project-display-settings`, `minimal-comp-object-profile`, `minimal-camera-object-profile`, `minimal-light-object-profile`, and `minimal-precomp-layer`.
- `tmp/migration_matrix_structural_representatives_all_6x6/matrix.json`: 180 total, 180 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for structural representatives `minimal-project-display-settings`, `minimal-comp-object-profile`, `minimal-camera-object-profile`, `minimal-light-object-profile`, and `minimal-precomp-layer`.
- `tmp/migration_matrix_project_all_6x6/matrix.json`: 144 total, 144 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for all 4 current `minimal-project-*` recipes.
- `tmp/migration_matrix_comp_all_6x6/matrix.json`: 648 total, 648 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for all 18 current `minimal-comp-*` recipes.
- `tmp/migration_matrix_precomp_all_6x6/matrix.json`: 36 total, 36 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for all 1 current `minimal-precomp-*` recipe.
- `tmp/migration_matrix_layer_all_6x6/matrix.json`: 576 total, 536 pass, 10 blocked, 0 failed, 30 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for all 16 current `minimal-layer-*` recipes; non-pass cases are limited to explicit/track matte source-contract boundaries.
- `tmp/migration_matrix_shape_all_6x6/matrix.json`: 972 total, 972 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for all 27 current `minimal-shape-*` recipes.
- `tmp/migration_matrix_camera_all_6x6/matrix.json`: 540 total, 540 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for all 15 current `minimal-camera-*` recipes.
- `tmp/migration_matrix_light_all_6x6/matrix.json`: 504 total, 504 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for all 14 current `minimal-light-*` recipes.
- `tmp/migration_matrix_layer_matte/matrix.json`: 6 total, 4 pass, 2 intentionally blocked, 0 failed, 0 skipped. This is recipe-source focused matte coverage: classic track matte passes W2020/W2022/W2025 targets; AE2025 explicit matte passes W2025 and is intentionally blocked for W2020/W2022 targets.
- `tmp/migration_matrix_layer_track_matte_all_6x6/matrix.json`: 36 total, 31 pass, 5 blocked, 0 failed, 0 skipped. This is the current all-writer boundary for `minimal-layer-track-matte`; only AE2025 source into W2020-W2024 targets is blocked by the explicit matte source contract.
- `tmp/migration_matrix_layer_mask/matrix.json`: 3 total, 3 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020/W2022/W2025 targets for the current recipe-owned layer mask surface.
- `tmp/migration_matrix_layer_mask_all_6x6/matrix.json`: 36 total, 36 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for the current recipe-owned layer mask surface.
- `tmp/migration_matrix_shape_gradient_stroke/matrix.json`: 12 total, 12 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020/W2022/W2025 targets for the current recipe-owned gradient stroke surface.
- `tmp/migration_matrix_shape_gradient_stroke_all_6x6/matrix.json`: 144 total, 144 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for every current recipe-owned gradient stroke variant.
- `tmp/migration_matrix_shape_gradient_stroke_all_hosts/matrix.json`: 24 total, 24 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020 target with AE-open on H2020-H2025 for every current recipe-owned gradient-stroke variant.
- `tmp/migration_matrix_writer_targets_all/matrix.json`: 6 total, 6 pass, 0 blocked, 0 failed, 0 skipped. This proves the matrix can produce W2020-W2025 writer targets for the stable `minimal-comp-object-profile` fixture.
- `tmp/migration_matrix_explicit_matte_source_contract/matrix.json`: 6 total, 0 pass, 0 blocked, 0 failed, 6 skipped. This proves AE2020 source does not author the AE2025-only explicit matte recipe and records those cases as source-contract skips.
- `tmp/migration_matrix_explicit_matte_ae2025/matrix.json`: 1 total, 1 pass, 0 blocked, 0 failed, 0 skipped. This proves the valid AE2025 explicit matte source-to-target contract still passes.
- `tmp/migration_matrix_smoke_all/matrix.json`: 846 total, 840 pass, 0 blocked, 0 failed, 6 skipped. This is the current post-source-contract-classification full no-AE boundary for W2020 source into W2020-W2025 targets.
- `tmp/migration_matrix_verify/`: generated by `scripts/migration/verify_matrix.ps1`; contains the recurring full W2020-source matrix and AE2025 explicit-matte contract matrix with asserted totals.
- Generated coverage ledger artifacts:
  - `tmp/migration_matrix_writer_targets_all/ledger.md`: generated from the narrow writer-target matrix; 1 recipe row with W2020-W2025 target evidence.
  - `tmp/migration_matrix_text_animators_all_6x6/ledger.md`: generated from the text animator 6x6 writer matrix; 22 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_text_baseline_all_6x6/ledger.md`: generated from the text baseline 6x6 writer matrix; 2 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_text_style_all_6x6/ledger.md`: generated from the text style 6x6 writer matrix; 2 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_transform_all_6x6/ledger.md`: generated from the dynamic transform 6x6 writer matrix; 3 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_auto_orient_all_6x6/ledger.md`: generated from the auto-orient 6x6 writer matrix; 1 recipe row with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_effect_params_all_6x6/ledger.md`: generated from the dynamic effect-param 6x6 writer matrix; 4 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_static_effects_all_6x6/ledger.md`: generated from the static effect 6x6 writer matrix; 2 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_layer_track_matte_all_6x6/ledger.md`: generated from the classic track matte all-writer matrix; 1 recipe row with 31 passed cases and 5 blocked AE2025-source downgrade cases.
  - `tmp/migration_matrix_layer_mask_all_6x6/ledger.md`: generated from the layer mask 6x6 writer matrix; 1 recipe row with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_shape_gradient_stroke_all_6x6/ledger.md`: generated from the gradient stroke 6x6 writer matrix; 4 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_explicit_matte_source_contract/ledger.md`: generated from the AE2020-source explicit-matte source-contract matrix; 1 recipe row, 6 skipped cases.
  - `tmp/migration_matrix_explicit_matte_ae2025/ledger.md`: generated from the valid AE2025 explicit-matte matrix; 1 recipe row, 1 passed case.
  - `tmp/migration_matrix_smoke_all/ledger.md`: generated by the recurring full no-AE matrix command via `-ledger-out`; 141 recipe rows grouped by inferred domain.
  - `tmp/migration_matrix_representative_all_hosts/ledger.md`: generated from the all-host representative matrix; 5 recipe rows, each with H2020-H2025 evidence.
  - `tmp/migration_matrix_structural_all_hosts/ledger.md`: generated from the structural all-host representative matrix; 5 recipe rows, each with H2020-H2025 evidence.
  - `tmp/migration_matrix_structural_representatives_all_6x6/ledger.md`: generated from the structural representative 6x6 writer matrix; 5 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_project_all_6x6/ledger.md`: generated from the project 6x6 writer matrix; 4 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_comp_all_6x6/ledger.md`: generated from the comp 6x6 writer matrix; 18 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_precomp_all_6x6/ledger.md`: generated from the precomp 6x6 writer matrix; 1 recipe row with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_layer_all_6x6/ledger.md`: generated from the layer all-family 6x6 boundary matrix; 16 recipe rows with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_shape_all_6x6/ledger.md`: generated from the shape 6x6 writer matrix; 27 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_camera_all_6x6/ledger.md`: generated from the camera 6x6 writer matrix; 15 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_light_all_6x6/ledger.md`: generated from the light 6x6 writer matrix; 14 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_shape_gradient_stroke_all_hosts/ledger.md`: generated from the all-host gradient-stroke matrix; 4 recipe rows, each with H2020-H2025 evidence.

## Immediate Missing Evidence

- Broader host-open fanout for non-representative text animator, effect,
  transform, project/comp, and camera-light variants if the host-open policy
  later raises them above representative coverage.
