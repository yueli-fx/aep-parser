# Versioned AEP Migration Validation Summary

This file is the reviewed validation ledger for versioned AEP migration. Raw
matrix output under `tmp/` is supporting evidence, not the source of truth.

Legend:

- `W2020/W2022/W2025`: writer targets produced by Go.
- `H2020/H2021/H2022/H2023/H2024/H2025`: installed AE hosts used to open output.
- `PD-3x3`: profile diff passed for source writers `W2020,W2022,W2025` into target writers `W2020,W2022,W2025`.
- `PD-1x3`: profile diff passed for source writer `W2020` into target writers `W2020,W2022,W2025`.
- `OPEN-H2025`: AE2025 open smoke passed.
- `OPEN-ALL-HOSTS`: one representative writer output opened in every installed AE host from 2020 through 2025.
- `OPEN-ENDPOINTS`: one representative writer output opened in H2020 and H2025.
- `INFER-MID-HOSTS`: H2021-H2024 were not run directly; compatibility is inferred from H2020 and H2025 endpoint success and must remain marked as inference.
- `pending`: not yet formally validated at that level.

## Current Answer

No, before this ledger the project could not quickly answer "which domain and
which AE versions are validated." The current reviewed answer is:

- Text animator migration has `PD-3x3` evidence for 22 recipe-owned animator fixtures, plus `OPEN-ALL-HOSTS` evidence for two representatives: static scalar skew and animated fill-color keyframes.
- Text style migration has `PD-1x3`, `OPEN-H2025`, and `OPEN-ALL-HOSTS` evidence for the current style fixture set.
- Dynamic effect params have `PD-1x3`, `OPEN-H2025`, and one vector-keyframe `OPEN-ALL-HOSTS` representative.
- Dynamic transforms have `PD-1x3`, `OPEN-H2025`, and one transform-ease `OPEN-ALL-HOSTS` representative.
- Layer track matte migration has `PD-1x3` evidence for the classic track-matte fixture and `OPEN-ALL-HOSTS` evidence for its W2020 output. AE2025 explicit matte has `PD-1x1` evidence for AE2025 source to AE2025 target; AE2020/AE2022 targets remain intentionally blocked by the current explicit-matte writer contract.
- Layer mask migration has `PD-1x3` evidence for the current recipe-owned mask surface: mode/options, static outline, and path keyframes. The representative W2020 mask output has `OPEN-ALL-HOSTS` evidence.
- Shape gradient stroke migration has `PD-1x3` evidence for the current recipe-owned gradient stroke surface: base gradient stroke, alpha stops, radial highlight, and stroke style. The representative W2020 base gradient-stroke output has `OPEN-ALL-HOSTS` evidence.
- Latest full no-AE matrix boundary is 423 total, 420 pass, 3 intentionally blocked, 0 failed, 0 skipped.
- Remaining blocked family is AE2025 explicit matte outside the AE2025 target contract. Other migrated domains are covered by the latest full no-AE matrix boundary, but many do not yet have per-capability host-version ledgers.

## Text Domain

| Capability | Recipes / scope | Profile-diff writer coverage | AE host open coverage | Evidence |
| --- | --- | --- | --- | --- |
| Text layer baseline | `minimal-default-text-layer`, `minimal-default-text-static-transform` | Included in latest full `PD-1x3` boundary | pending per-capability host ledger | `tmp/migration_matrix_smoke_all` previous recorded full matrix: 423 total, 336 pass, 87 blocked |
| Text style | `minimal-text-style`, `minimal-text-shape` | `PD-1x3`: W2020 -> W2020/W2022/W2025 | `OPEN-H2025`; `OPEN-ALL-HOSTS` for `minimal-text-style` as W2020 output | `history.md` text style entry |
| Text animator opacity | `minimal-text-animator-opacity`, `minimal-text-animator-opacity-value-keyframes` | `PD-3x3`: W2020/W2022/W2025 -> W2020/W2022/W2025 | pending migration AE-host fanout | `tmp/migration_matrix_text_animators_all_writers/matrix.json`: 198 total, 198 pass |
| Text animator position | `minimal-text-animator-position`, `minimal-text-animator-position-value-keyframes` | `PD-3x3` | pending migration AE-host fanout | same text animator matrix |
| Text animator scale | `minimal-text-animator-scale`, `minimal-text-animator-scale-value-keyframes` | `PD-3x3` | pending migration AE-host fanout | same text animator matrix |
| Text animator rotation Z | `minimal-text-animator-rotation`, `minimal-text-animator-rotation-value-keyframes` | `PD-3x3` | pending migration AE-host fanout | same text animator matrix |
| Text animator rotation X/Y | `minimal-text-animator-rotation-x`, `minimal-text-animator-rotation-y` | `PD-3x3` | pending migration AE-host fanout | same text animator matrix |
| Text animator fill color | `minimal-text-animator-color`, `minimal-text-animator-color-value-keyframes` | `PD-3x3` | `OPEN-ALL-HOSTS` for `minimal-text-animator-color-value-keyframes` W2020 output on H2020-H2025 | `tmp/migration_matrix_text_animators_all_writers/matrix.json`; `tmp/migration_matrix_representative_all_hosts/matrix.json`: representative all-host matrix, 30 total, 30 pass |
| Text animator stroke color | `minimal-text-animator-stroke-color` | `PD-3x3` | pending migration AE-host fanout | same text animator matrix |
| Text animator tracking | `minimal-text-animator-tracking`, `minimal-text-animator-tracking-value-keyframes` | `PD-3x3` | pending migration AE-host fanout | same text animator matrix |
| Text animator character offset | `minimal-text-animator-character-offset`, `minimal-text-animator-character-offset-value-keyframes` | `PD-3x3` | pending migration AE-host fanout | same text animator matrix |
| Text animator fill/stroke opacity and stroke width | `minimal-text-animator-fill-opacity`, `minimal-text-animator-stroke-opacity`, `minimal-text-animator-stroke-width` | `PD-3x3` | pending migration AE-host fanout | same text animator matrix |
| Text animator skew | `minimal-text-animator-skew` | `PD-3x3` | `OPEN-ALL-HOSTS` for W2020 output on H2020-H2025 | `tmp/migration_matrix_text_animators_all_writers/matrix.json`; `tmp/migration_matrix_representative_all_hosts/matrix.json`: representative all-host matrix, 30 total, 30 pass |
| Text animator range selector offset keyframes | `minimal-text-animator-range-offset` | `PD-3x3` | pending migration AE-host fanout | same text animator matrix |

## Effect Domain

| Capability | Recipes / scope | Profile-diff writer coverage | AE host open coverage | Evidence |
| --- | --- | --- | --- | --- |
| Static supported effects | `minimal-text-effect`, `minimal-adjustment-layer`, supported static built-in effect params | Included in latest full `PD-1x3` boundary | pending per-capability host ledger | `history.md`; latest full no-AE matrix 423 total, 336 pass, 87 blocked |
| Effect layer-ref params | `minimal-effect-layer-param` | Included in latest full `PD-1x3` boundary | pending per-capability host ledger | `history.md` |
| Effect param expression | `minimal-effect-param-expression` | `PD-1x3`: W2020 -> W2020/W2022/W2025 | `OPEN-H2025` | `history.md` dynamic effect-param entry |
| Effect scalar keyframes | `minimal-effect-param-keyframes` | `PD-1x3` | `OPEN-H2025` | `history.md` dynamic effect-param entry |
| Effect vector keyframes | `minimal-effect-param-vector-keyframes` | `PD-1x3` | `OPEN-H2025`; `OPEN-ALL-HOSTS` for W2020 representative | `history.md` dynamic effect-param entry |

## Transform Domain

| Capability | Recipes / scope | Profile-diff writer coverage | AE host open coverage | Evidence |
| --- | --- | --- | --- | --- |
| Transform keyframes | `minimal-transform-keyframes` | `PD-1x3`: W2020 -> W2020/W2022/W2025 | `OPEN-H2025` | `history.md` dynamic transform entry |
| Transform keyframe ease | `minimal-transform-keyframe-ease` | `PD-1x3` | `OPEN-H2025`; `OPEN-ALL-HOSTS` for W2020 representative | `history.md` dynamic transform entry |
| Transform expressions | `minimal-transform-expression` | `PD-1x3` | `OPEN-H2025` | `history.md` dynamic transform entry |
| Auto-orient unlock | `minimal-layer-auto-orient` | Included in latest full `PD-1x3` boundary after transform keyframes | pending per-capability host ledger | `history.md` dynamic transform entry |

## Layer Domain

| Capability | Recipes / scope | Profile-diff writer coverage | AE host open coverage | Evidence |
| --- | --- | --- | --- | --- |
| Classic track matte | `minimal-layer-track-matte` | `PD-1x3`: recipe source into W2020/W2022/W2025 targets | `OPEN-ALL-HOSTS` for W2020 output on H2020-H2025 | `tmp/migration_matrix_layer_matte/matrix.json`: recipe-source focused matrix, 4 pass / 2 intentional blocked overall; `tmp/migration_matrix_representative_all_hosts/matrix.json`: representative all-host matrix, 30 total, 30 pass |
| AE2025 explicit matte | `minimal-layer-explicit-matte` | `PD-1x1`: AE2025 source into W2025 target; W2020/W2022 targets intentionally blocked | pending per-capability host ledger | `tmp/migration_matrix_layer_matte/matrix.json`; low targets blocked with `explicit matte source requires AE2025` |
| Layer mask | `minimal-layer-mask` | `PD-1x3`: W2020 source into W2020/W2022/W2025 targets | `OPEN-ALL-HOSTS` for W2020 output on H2020-H2025 | `tmp/migration_matrix_layer_mask/matrix.json`: 3 total, 3 pass; `tmp/migration_matrix_representative_all_hosts/matrix.json`: representative all-host matrix, 30 total, 30 pass |

## Shape Domain

| Capability | Recipes / scope | Profile-diff writer coverage | AE host open coverage | Evidence |
| --- | --- | --- | --- | --- |
| Gradient stroke | `minimal-shape-gradient-stroke`, `minimal-shape-gradient-stroke-alpha-stops`, `minimal-shape-gradient-stroke-highlight`, `minimal-shape-gradient-stroke-style` | `PD-1x3`: W2020 source into W2020/W2022/W2025 targets | `OPEN-ALL-HOSTS` for base W2020 output on H2020-H2025 | `tmp/migration_matrix_shape_gradient_stroke/matrix.json`: 12 total, 12 pass; `tmp/migration_matrix_representative_all_hosts/matrix.json`: representative all-host matrix, 30 total, 30 pass |

## Other Domains

| Domain | Current reviewed state | Writer coverage | AE host open coverage |
| --- | --- | --- | --- |
| project | Implemented in convert surface and included in full no-AE matrix boundary | latest full `PD-1x3` boundary | representative checks only |
| comp | Implemented for stable comp settings, work area, renderer, metadata | latest full `PD-1x3` boundary | representative checks only |
| layer | Implemented for default layer creation, switches, refs, timing, parent/source refs, supported matte slices, and supported mask slices | latest full `PD-1x3` boundary plus focused matte/mask matrices | representative checks only |
| shape | Implemented for supported parametric graphic/filter shape slices, gradient fill slices, and gradient stroke slices | latest full `PD-1x3` boundary plus focused gradient-stroke matrix | representative checks only |
| camera-light | Implemented for supported camera/light options and light source refs | latest full `PD-1x3` boundary | representative checks only |
| precomp | Implemented for precomp refs/layers | latest full `PD-1x3` boundary | representative checks only |

## Current Remaining Blockers

Latest full no-AE matrix:

- Artifact: `tmp/migration_matrix_smoke_all/matrix.json`
- Summary: 423 total, 420 pass, 3 blocked, 0 failed, 0 skipped
- Writer coverage: W2020 source into W2020/W2022/W2025 targets

Blocked recipe groups:

| Domain | Recipe | Cases | Status |
| --- | --- | ---: | --- |
| layer | `minimal-layer-explicit-matte` | 3 | blocked in W2020-source full matrix because explicit matte requires AE2025 source/target contract |

## Raw Matrix Artifacts

Reviewed raw artifacts currently known:

- `tmp/migration_matrix_text_animators/matrix.json`: 66 total, 66 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020/W2022/W2025 targets for all 22 text animator recipes.
- `tmp/migration_matrix_text_animators_all_writers/matrix.json`: 198 total, 198 pass, 0 blocked, 0 failed, 0 skipped. This is W2020/W2022/W2025 source writers into W2020/W2022/W2025 target writers for all 22 text animator recipes.
- `tmp/migration_matrix_text_animators_endpoint_hosts/matrix.json`: 4 total, 4 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 writer output for `minimal-text-animator-skew` and `minimal-text-animator-color-value-keyframes` opened in H2020 and H2025. H2021-H2024 were not run and are only inferred low-risk.
- `tmp/migration_matrix_representative_all_hosts/matrix.json`: 30 total, 30 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020 target with AE-open on H2020-H2025 for `minimal-text-animator-skew`, `minimal-text-animator-color-value-keyframes`, `minimal-layer-track-matte`, `minimal-layer-mask`, and `minimal-shape-gradient-stroke`.
- `tmp/migration_matrix_layer_matte/matrix.json`: 6 total, 4 pass, 2 intentionally blocked, 0 failed, 0 skipped. This is recipe-source focused matte coverage: classic track matte passes W2020/W2022/W2025 targets; AE2025 explicit matte passes W2025 and is intentionally blocked for W2020/W2022 targets.
- `tmp/migration_matrix_layer_mask/matrix.json`: 3 total, 3 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020/W2022/W2025 targets for the current recipe-owned layer mask surface.
- `tmp/migration_matrix_shape_gradient_stroke/matrix.json`: 12 total, 12 pass, 0 blocked, 0 failed, 0 skipped. This is W2020 source into W2020/W2022/W2025 targets for the current recipe-owned gradient stroke surface.
- `tmp/migration_matrix_smoke_all/matrix.json`: 423 total, 420 pass, 3 intentionally blocked, 0 failed, 0 skipped. This is the current post-gradient-stroke full no-AE boundary.
- Generated coverage ledger artifacts:
  - `tmp/migration_matrix_smoke_all/ledger.md`: generated from the full no-AE matrix; 141 recipe rows grouped by inferred domain.
  - `tmp/migration_matrix_representative_all_hosts/ledger.md`: generated from the all-host representative matrix; 5 recipe rows, each with H2020-H2025 evidence.

## Immediate Missing Evidence

- Per-domain host-open policy: which capabilities need all-host fanout versus AE2025 smoke only.
- AE-host open evidence for non-representative layer and shape variants where policy requires more than a representative smoke.
