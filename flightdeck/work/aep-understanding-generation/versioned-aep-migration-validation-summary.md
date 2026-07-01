# Versioned AEP Migration Validation Summary

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
- Text style migration has `PD-1x3`, `OPEN-H2025`, and `OPEN-ALL-HOSTS` evidence for the current style fixture set.
- Dynamic effect params have `PD-6x6`, `OPEN-H2025`, and one vector-keyframe `OPEN-ALL-HOSTS` representative.
- Dynamic transforms have `PD-6x6`, `OPEN-H2025`, and one transform-ease `OPEN-ALL-HOSTS` representative.
- Layer track matte migration has `PD-1x3` evidence for the classic track-matte fixture and `OPEN-ALL-HOSTS` evidence for its W2020 output. AE2025 explicit matte has `PD-1x1` evidence for AE2025 source to AE2025 target; lower-source explicit matte matrix cases are now marked as source-contract skips, not conversion blockers.
- Layer mask migration has `PD-1x3` evidence for the current recipe-owned mask surface: mode/options, static outline, and path keyframes. The representative W2020 mask output has `OPEN-ALL-HOSTS` evidence.
- Shape gradient stroke migration has `PD-1x3` and `OPEN-ALL-HOSTS` evidence for the current recipe-owned gradient stroke surface: base gradient stroke, alpha stops, radial highlight, and stroke style.
- Native writer targets now exist for W2020-W2025. The narrow writer-target matrix for `minimal-comp-object-profile` is 6 total, 6 pass.
- Latest full no-AE matrix boundary is 846 total, 840 pass, 0 blocked, 0 failed, 6 source-contract skipped. It covers W2020 source into W2020-W2025 targets.
- There are no remaining blocked or failed cases in the current W2020-source full no-AE matrix. The only skipped family is AE2025 explicit matte outside the AE2025 source contract. Other migrated domains are covered by the latest full no-AE matrix boundary, but many do not yet have per-capability host-version ledgers.

## Text Domain

| Capability | Recipes / scope | Profile-diff writer coverage | AE host open coverage | Evidence |
| --- | --- | --- | --- | --- |
| Text layer baseline | `minimal-default-text-layer`, `minimal-default-text-static-transform` | Included in latest full `PD-1x6` boundary | pending per-capability host ledger | `tmp/migration_matrix_smoke_all`: 846 total, 840 pass, 0 blocked, 6 skipped |
| Text style | `minimal-text-style`, `minimal-text-shape` | `PD-1x3`: W2020 -> W2020/W2022/W2025 | `OPEN-H2025`; `OPEN-ALL-HOSTS` for `minimal-text-style` as W2020 output | `history.md` text style entry |
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
| Static supported effects | `minimal-text-effect`, `minimal-adjustment-layer`, supported static built-in effect params | Included in latest full `PD-1x6` boundary | pending per-capability host ledger | `history.md`; latest full no-AE matrix 846 total, 840 pass, 0 blocked, 6 skipped |
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
| Auto-orient unlock | `minimal-layer-auto-orient` | Included in latest full `PD-1x6` boundary after transform keyframes | pending per-capability host ledger | `history.md` dynamic transform entry; latest full no-AE matrix 846 total, 840 pass, 0 blocked, 6 skipped |

## Layer Domain

| Capability | Recipes / scope | Profile-diff writer coverage | AE host open coverage | Evidence |
| --- | --- | --- | --- | --- |
| Classic track matte | `minimal-layer-track-matte` | `PD-1x3`: recipe source into W2020/W2022/W2025 targets | `OPEN-ALL-HOSTS` for W2020 output on H2020-H2025 | `tmp/migration_matrix_layer_matte/matrix.json`: recipe-source focused matrix, 4 pass / 2 intentional blocked overall; `tmp/migration_matrix_representative_all_hosts/matrix.json`: representative all-host matrix, 30 total, 30 pass |
| AE2025 explicit matte | `minimal-layer-explicit-matte` | `PD-1x1`: AE2025 source into W2025 target; AE2020 source into W2020-W2025 targets skipped by source contract | pending per-capability host ledger | `tmp/migration_matrix_explicit_matte_ae2025/matrix.json`: 1 total, 1 pass; `tmp/migration_matrix_explicit_matte_source_contract/matrix.json`: 6 total, 6 skipped |
| Layer mask | `minimal-layer-mask` | `PD-1x3`: W2020 source into W2020/W2022/W2025 targets | `OPEN-ALL-HOSTS` for W2020 output on H2020-H2025 | `tmp/migration_matrix_layer_mask/matrix.json`: 3 total, 3 pass; `tmp/migration_matrix_representative_all_hosts/matrix.json`: representative all-host matrix, 30 total, 30 pass |

## Shape Domain

| Capability | Recipes / scope | Profile-diff writer coverage | AE host open coverage | Evidence |
| --- | --- | --- | --- | --- |
| Gradient stroke | `minimal-shape-gradient-stroke`, `minimal-shape-gradient-stroke-alpha-stops`, `minimal-shape-gradient-stroke-highlight`, `minimal-shape-gradient-stroke-style` | `PD-1x3`: W2020 source into W2020/W2022/W2025 targets | `OPEN-ALL-HOSTS` for all current gradient-stroke W2020 outputs on H2020-H2025 | `tmp/migration_matrix_shape_gradient_stroke/matrix.json`: 12 total, 12 pass; `tmp/migration_matrix_shape_gradient_stroke_all_hosts/matrix.json`: 24 total, 24 pass |

## Other Domains

| Domain | Current reviewed state | Writer coverage | AE host open coverage |
| --- | --- | --- | --- |
| project | Implemented in convert surface and included in full no-AE matrix boundary | latest full `PD-1x6` boundary | `OPEN-ALL-HOSTS` for `minimal-project-display-settings` |
| comp | Implemented for stable comp settings, work area, renderer, metadata | latest full `PD-1x6` boundary | `OPEN-ALL-HOSTS` for `minimal-comp-object-profile` |
| layer | Implemented for default layer creation, switches, refs, timing, parent/source refs, supported matte slices, and supported mask slices | latest full `PD-1x6` boundary plus focused matte/mask matrices | representative checks only |
| shape | Implemented for supported parametric graphic/filter shape slices, gradient fill slices, and gradient stroke slices | latest full `PD-1x6` boundary plus focused gradient-stroke matrix | representative checks only |
| camera-light | Implemented for supported camera/light options and light source refs | latest full `PD-1x6` boundary | `OPEN-ALL-HOSTS` for `minimal-camera-object-profile` and `minimal-light-object-profile` |
| precomp | Implemented for precomp refs/layers | latest full `PD-1x6` boundary | `OPEN-ALL-HOSTS` for `minimal-precomp-layer` |

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
- `tmp/migration_matrix_transform_all_6x6/matrix.json`: 108 total, 108 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for `minimal-transform-keyframes`, `minimal-transform-keyframe-ease`, and `minimal-transform-expression`.
- `tmp/migration_matrix_effect_params_all_6x6/matrix.json`: 144 total, 144 pass, 0 blocked, 0 failed, 0 skipped. This is W2020-W2025 source writers into W2020-W2025 target writers for layer-ref, expression, scalar-keyframe, and vector-keyframe effect params.
- `tmp/migration_matrix_text_animators_endpoint_hosts/matrix.json`: 4 total, 4 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 writer output for `minimal-text-animator-skew` and `minimal-text-animator-color-value-keyframes` opened in H2020 and H2025. H2021-H2024 were not run and are only inferred low-risk.
- `tmp/migration_matrix_representative_all_hosts/matrix.json`: 30 total, 30 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020 target with AE-open on H2020-H2025 for `minimal-text-animator-skew`, `minimal-text-animator-color-value-keyframes`, `minimal-layer-track-matte`, `minimal-layer-mask`, and `minimal-shape-gradient-stroke`.
- `tmp/migration_matrix_structural_all_hosts/matrix.json`: 30 total, 30 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020 target with AE-open on H2020-H2025 for `minimal-project-display-settings`, `minimal-comp-object-profile`, `minimal-camera-object-profile`, `minimal-light-object-profile`, and `minimal-precomp-layer`.
- `tmp/migration_matrix_layer_matte/matrix.json`: 6 total, 4 pass, 2 intentionally blocked, 0 failed, 0 skipped. This is recipe-source focused matte coverage: classic track matte passes W2020/W2022/W2025 targets; AE2025 explicit matte passes W2025 and is intentionally blocked for W2020/W2022 targets.
- `tmp/migration_matrix_layer_mask/matrix.json`: 3 total, 3 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020/W2022/W2025 targets for the current recipe-owned layer mask surface.
- `tmp/migration_matrix_shape_gradient_stroke/matrix.json`: 12 total, 12 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020/W2022/W2025 targets for the current recipe-owned gradient stroke surface.
- `tmp/migration_matrix_shape_gradient_stroke_all_hosts/matrix.json`: 24 total, 24 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020 target with AE-open on H2020-H2025 for every current recipe-owned gradient-stroke variant.
- `tmp/migration_matrix_writer_targets_all/matrix.json`: 6 total, 6 pass, 0 blocked, 0 failed, 0 skipped. This proves the matrix can produce W2020-W2025 writer targets for the stable `minimal-comp-object-profile` fixture.
- `tmp/migration_matrix_explicit_matte_source_contract/matrix.json`: 6 total, 0 pass, 0 blocked, 0 failed, 6 skipped. This proves AE2020 source does not author the AE2025-only explicit matte recipe and records those cases as source-contract skips.
- `tmp/migration_matrix_explicit_matte_ae2025/matrix.json`: 1 total, 1 pass, 0 blocked, 0 failed, 0 skipped. This proves the valid AE2025 explicit matte source-to-target contract still passes.
- `tmp/migration_matrix_smoke_all/matrix.json`: 846 total, 840 pass, 0 blocked, 0 failed, 6 skipped. This is the current post-source-contract-classification full no-AE boundary for W2020 source into W2020-W2025 targets.
- `tmp/migration_matrix_verify/`: generated by `scripts/migration/verify_matrix.ps1`; contains the recurring full W2020-source matrix and AE2025 explicit-matte contract matrix with asserted totals.
- Generated coverage ledger artifacts:
  - `tmp/migration_matrix_writer_targets_all/ledger.md`: generated from the narrow writer-target matrix; 1 recipe row with W2020-W2025 target evidence.
  - `tmp/migration_matrix_text_animators_all_6x6/ledger.md`: generated from the text animator 6x6 writer matrix; 22 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_transform_all_6x6/ledger.md`: generated from the dynamic transform 6x6 writer matrix; 3 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_effect_params_all_6x6/ledger.md`: generated from the dynamic effect-param 6x6 writer matrix; 4 recipe rows, each with W2020-W2025 source and target evidence.
  - `tmp/migration_matrix_explicit_matte_source_contract/ledger.md`: generated from the AE2020-source explicit-matte source-contract matrix; 1 recipe row, 6 skipped cases.
  - `tmp/migration_matrix_explicit_matte_ae2025/ledger.md`: generated from the valid AE2025 explicit-matte matrix; 1 recipe row, 1 passed case.
  - `tmp/migration_matrix_smoke_all/ledger.md`: generated by the recurring full no-AE matrix command via `-ledger-out`; 141 recipe rows grouped by inferred domain.
  - `tmp/migration_matrix_representative_all_hosts/ledger.md`: generated from the all-host representative matrix; 5 recipe rows, each with H2020-H2025 evidence.
  - `tmp/migration_matrix_structural_all_hosts/ledger.md`: generated from the structural all-host representative matrix; 5 recipe rows, each with H2020-H2025 evidence.
  - `tmp/migration_matrix_shape_gradient_stroke_all_hosts/ledger.md`: generated from the all-host gradient-stroke matrix; 4 recipe rows, each with H2020-H2025 evidence.

## Immediate Missing Evidence

- Broader host-open fanout for non-representative text animator, effect,
  transform, project/comp, camera-light, and precomp variants if the host-open
  policy later raises them above representative coverage.
