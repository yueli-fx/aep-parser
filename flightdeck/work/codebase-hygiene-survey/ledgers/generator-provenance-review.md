# Generator Provenance Review

Manual review and cleanup execution for the 26 original `delete-candidate` rows from `generator-ledger.md`.

The mechanical ledger only counted direct references to each `.jsx` filename. This pass also checked declared outputs, `scripts/fixtures/fixtures_manifest.json`, current Go/knowledge references, and obvious superseding scripts.

## Summary

- keep-active: 10
- deleted-script: 16
- deleted-tracked-fixture: 8
- stale-baseline-hashes-removed: 16

No generator was deleted solely from the original `delete-candidate` class. Deletions were applied only after checking fixture manifest ownership, produced fixture names, current references, tracked outputs, and split-roundtrip baseline rows.

## Keep Active

These are active because `scripts/fixtures/fixtures_manifest.json` regenerates them:

| Generator | Reason |
|---|---|
| `test_data/generators/build_material_classic_2020.jsx` | Manifest entry added; produces `test_data/fixtures/re_material_classic_2020.aep`, used by material classic tests. |
| `test_data/generators/build_re_comp_idta.jsx` | Output path repaired to `test_data/fixtures/re_comp_idta.aep`; manifest entry added. |
| `test_data/generators/build_re_layer_comment.jsx` | Output path repaired to `test_data/fixtures/re_layer_comment.aep`; manifest entry added. |
| `test_data/generators/fdta_probe_AE2020.jsx` | Manifest entry; produces `test_data/fixtures/fdta_probe/AE2020_*.aep`. |
| `test_data/generators/fdta_probe_AE2025.jsx` | Manifest entry; produces `test_data/fixtures/fdta_probe/AE2025_*.aep`. |
| `test_data/generators/gen_text_range_adv_smoothness.jsx` | Manifest entry added; source fixture B for text range advanced template synthesis. |
| `test_data/generators/re_camera_filmsize.jsx` | Manifest entry; output also appears in split-roundtrip baseline. |
| `test_data/generators/re_cdta_ae2020.jsx` | Manifest entry; output also appears in split-roundtrip baseline. |
| `test_data/generators/re_cross_project_insert.jsx` | Manifest entry; produces generated cross-project source/dest fixtures. |
| `test_data/generators/re_sepdim_anim3.jsx` | Manifest entry; generated outputs appear in split-roundtrip baseline. |

## Deleted

These rows had no current Go/manifest/knowledge owner beyond their own script, an obsolete generated baseline, or a newer replacement. They were removed from `test_data/generators/`.

| Generator | Cleanup |
|---|---|
| `test_data/generators/probe_open.jsx` | Superseded by `verify_open.jsx`; no tracked output. |
| `test_data/generators/probe_text_range_advanced.jsx` | Removed with unreferenced `test_data/fixtures/probe_text_range_advanced.aep`. |
| `test_data/generators/re_resfactor.jsx` | Removed with unreferenced `test_data/fixtures/re_resfactor_{half,third,full}.aep`. |
| `test_data/generators/re_transparency.jsx` | Removed with unreferenced `test_data/fixtures/re_transp_{off,on}.aep`. |
| `test_data/generators/smoke_ae_run.jsx` | Placeholder `$DONE_PLACEHOLDER$` smoke script; no current references. |
| `test_data/generators/verify_dropframe.jsx` | Obsolete side-car for active `re_wave2_ae24.jsx`; `re_wave2_ae24.aep` was retained. |
| `test_data/generators/verify_ge_cross_project_insert.jsx` | Removed with stale `test_data/generated/ship-gate/ge_cross_project_insert_*.aep` baseline hashes. |
| `test_data/generators/verify_ge_delete_layer.jsx` | Removed with stale `test_data/generated/ship-gate/ge_delete_layer_*.aep` baseline hashes. |
| `test_data/generators/verify_ge_duplicate_composition.jsx` | Removed with stale `test_data/generated/ship-gate/ge_duplicate_composition.aep` baseline hash. |
| `test_data/generators/verify_ge_duplicate_layer_explicit_matte.jsx` | Removed with stale generated baseline hash; `knowledge/layer/ae-duplicatelayer-re.md` now preserves the historical conclusion. |
| `test_data/generators/verify_ge_duplicate_layer.jsx` | Removed with stale `test_data/generated/ship-gate/ge_duplicate_layer_*.aep` baseline hashes. |
| `test_data/generators/verify_ge_move_layer.jsx` | Removed with stale `test_data/generated/ship-gate/ge_move_layer_*.aep` baseline hashes. |
| `test_data/generators/verify_light_kinds.jsx` | Removed with unreferenced `test_data/fixtures/light_kinds_verify.aep`; light-kind coverage remains in active `re_wave2_ae24.aep`. |
| `test_data/generators/verify_open_batch.jsx` | Superseded by configurable `verify_open.jsx`; no tracked output. |
| `test_data/generators/verify_open_multicomp.jsx` | Superseded by configurable `verify_open.jsx`; `aepmigrate` uses `verify_open.jsx`. |
| `test_data/generators/verify_resfactor.jsx` | Removed with unreferenced `test_data/fixtures/resfactor_go_verify.aep`. |

## Deleted Tracked Fixtures

- `test_data/fixtures/probe_text_range_advanced.aep`
- `test_data/fixtures/re_resfactor_full.aep`
- `test_data/fixtures/re_resfactor_half.aep`
- `test_data/fixtures/re_resfactor_third.aep`
- `test_data/fixtures/re_transp_off.aep`
- `test_data/fixtures/re_transp_on.aep`
- `test_data/fixtures/light_kinds_verify.aep`
- `test_data/fixtures/resfactor_go_verify.aep`

## Removed Baseline Hashes

The `test_data/generated/ship-gate/` directory is absent from the current tree, and no `ge_*` AEPs are tracked. Removed stale split-roundtrip baseline hashes for:

- `ge_cross_project_insert_{dedup,footage,precomp}.aep`
- `ge_delete_layer_{baseline,matte_ae20,matte_ae25,middle,parent}.aep`
- `ge_duplicate_composition.aep`
- `ge_duplicate_layer_{dup_child,dup_parent,explicit_matte,solo}.aep`
- `ge_move_layer_{first_to_last,last_to_first,mid_swap}.aep`

The full `tools/debug/split_roundtrip_baseline` output is broader than this cleanup batch and currently differs from the checked-in baseline for existing generated-fixture drift outside the removed `ge_*` rows. Do not refresh the whole baseline in a generator cleanup commit without owning that larger fixture-baseline reset.

## Verification

- `pwsh -NoProfile -File scripts/fixtures/regen_fixtures.ps1 -CheckOnly` -> `79 manifest jobs, 47 with missing outputs (46 driveable, 1 manual-only), 0 ungoverned on-disk aep`.
- `go run ./cmd/aepregistry audit -root .` -> pass, 0 errors, 0 warnings.
- `go run ./cmd/aepregistry layout -root .` -> 18 locations, 0 cleanup candidates, 0 blocked unowned.
- `go run ./cmd/aepregistry ownership -root .` -> 18 locations, 1550 files, 1250 owned, 300 unowned.
- `go run ./cmd/capindex -check` -> pass.
