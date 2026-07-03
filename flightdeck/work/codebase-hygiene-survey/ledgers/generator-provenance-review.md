# Generator Provenance Review

Manual review of the 26 `delete-candidate` rows from `generator-ledger.md`.

The mechanical ledger only counted direct references to each `.jsx` filename. This pass also checked declared outputs, `scripts/fixtures/fixtures_manifest.json`, current Go/knowledge references, and obvious superseding scripts.

## Summary

- keep-active: 6
- keep-provenance-gap: 5
- archive-or-delete-candidate: 15
- direct-delete-now: 0

No generator should be deleted solely from the original `delete-candidate` class. Several candidates are fixture provenance sources whose script path is not directly referenced by tests.

## Keep Active

These are active because `scripts/fixtures/fixtures_manifest.json` already regenerates them:

| Generator | Reason |
|---|---|
| `test_data/generators/fdta_probe_AE2020.jsx` | Manifest entry; produces `test_data/fixtures/fdta_probe/AE2020_*.aep`. |
| `test_data/generators/fdta_probe_AE2025.jsx` | Manifest entry; produces `test_data/fixtures/fdta_probe/AE2025_*.aep`. |
| `test_data/generators/re_camera_filmsize.jsx` | Manifest entry; output also appears in split-roundtrip baseline. |
| `test_data/generators/re_cdta_ae2020.jsx` | Manifest entry; output also appears in split-roundtrip baseline. |
| `test_data/generators/re_cross_project_insert.jsx` | Manifest entry; produces generated cross-project source/dest fixtures. |
| `test_data/generators/re_sepdim_anim3.jsx` | Manifest entry; generated outputs appear in split-roundtrip baseline. |

## Keep, But Repair Provenance

These should not be deleted, but their generation story is incomplete or stale.

| Generator | Current evidence | Needed repair |
|---|---|---|
| `test_data/generators/build_material_classic_2020.jsx` | Produces `test_data/fixtures/re_material_classic_2020.aep`; used by `internal/aep_test/material_classic_shipgate_test.go`. | Add to fixture manifest or document why it stays manual. |
| `test_data/generators/build_re_comp_idta.jsx` | `re_comp_idta.aep` is cited by serializer comments, tests, and knowledge. | Script currently saves to `test_data/re_comp_idta.aep`, while tracked fixture is `test_data/fixtures/re_comp_idta.aep`; fix output path before adding manifest. |
| `test_data/generators/build_re_layer_comment.jsx` | `re_layer_comment.aep` is cited by codec/test comments and layer knowledge. | Script currently saves to `test_data/re_layer_comment.aep`, while tracked fixture is `test_data/fixtures/re_layer_comment.aep`; fix output path before adding manifest. |
| `test_data/generators/gen_text_range_adv_smoothness.jsx` | Produces tracked `test_data/fixtures/re_text_range_adv_sm.aep`; comments say it is fixture B for Range Advanced template synthesis. | Either document the synthesis/extraction dependency or prove the embedded template no longer needs this source fixture. |
| `test_data/generators/verify_ge_duplicate_layer_explicit_matte.jsx` | Generated `.aep` appears in split-roundtrip baseline and `knowledge/layer/ae-duplicatelayer-re.md` as historical AE 2025 ship-gate evidence. | Decide whether historical generated ship-gate AEPs in `test_data/generated/ship-gate` stay as baseline evidence or move to durable registry evidence. |

## Archive Or Delete Candidates

These had no current Go/manifest/knowledge owner beyond their own script, an obsolete generated baseline, or an obvious newer replacement. Delete only after deciding how to handle tracked fixture outputs and `tools/debug/split_roundtrip_baseline/baseline.txt`.

| Generator | Why candidate |
|---|---|
| `test_data/generators/probe_open.jsx` | Superseded by `verify_open.jsx`; no current references. |
| `test_data/generators/probe_text_range_advanced.jsx` | Produces tracked `probe_text_range_advanced.aep`, but no current code/knowledge references were found. |
| `test_data/generators/re_resfactor.jsx` | Produces tracked `re_resfactor_{half,third,full}.aep`, but no current references were found. |
| `test_data/generators/re_transparency.jsx` | Produces tracked `re_transp_{off,on}.aep`, but no current references were found. |
| `test_data/generators/smoke_ae_run.jsx` | Placeholder `$DONE_PLACEHOLDER$` smoke script; no current references. |
| `test_data/generators/verify_dropframe.jsx` | Obsolete side-car for `re_wave2_ae24.aep`; canonical source is `re_wave2_ae24.jsx` in the manifest. |
| `test_data/generators/verify_ge_cross_project_insert.jsx` | No current Go ship-gate references; generated AEPs remain only in split-roundtrip baseline. |
| `test_data/generators/verify_ge_delete_layer.jsx` | No current Go ship-gate references; generated AEPs remain only in split-roundtrip baseline. |
| `test_data/generators/verify_ge_duplicate_composition.jsx` | No current Go ship-gate references; generated AEP remains only in split-roundtrip baseline. |
| `test_data/generators/verify_ge_duplicate_layer.jsx` | No current Go ship-gate references; generated AEPs remain only in split-roundtrip baseline. |
| `test_data/generators/verify_ge_move_layer.jsx` | No current Go ship-gate references; generated AEPs remain only in split-roundtrip baseline. |
| `test_data/generators/verify_light_kinds.jsx` | Opens tracked `light_kinds_verify.aep`, but no current references were found; light-kind coverage appears to live in `re_wave2_ae24.aep`. |
| `test_data/generators/verify_open_batch.jsx` | Superseded by configurable `verify_open.jsx`; no current references. |
| `test_data/generators/verify_open_multicomp.jsx` | Superseded by configurable `verify_open.jsx`; `aepmigrate` uses `verify_open.jsx`. |
| `test_data/generators/verify_resfactor.jsx` | Opens tracked `resfactor_go_verify.aep`, but no current references were found. |

## Next Actions

1. Repair provenance for the 5 `keep-provenance-gap` rows before deleting anything around them.
2. For the 15 archive/delete candidates, first decide whether their tracked `.aep` outputs and `tools/debug/split_roundtrip_baseline/baseline.txt` entries are still useful. If not, delete the script, its unreferenced tracked fixture(s), and any stale baseline hash in one cleanup commit.
3. Re-run `scripts/fixtures/regen_fixtures.ps1 -CheckOnly` after manifest edits, if AE availability allows it; otherwise run non-AE registry/capindex gates and record AE as not run.
