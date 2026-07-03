# Knowledge Body Validity Review

This pass is evidence-first. It does not merge files because their prose looks
similar. A knowledge entry is changed only when a current code path, fixture
owner, registry entry, example, or test proves the old wording can mislead a
future session.

## Summary

- files scanned: 142
- routing header gaps after scan: 0
- stale current-path references fixed: 6
- generated-evidence status notes added: 4 topics
- simple similarity merges applied: 0

## Method

1. Extract path-like references from every `flightdeck/knowledge/**/*.md`.
2. Check exact paths against the current checkout.
3. For missing paths, search current code/tests/registry/examples by stem.
4. Classify each hit as:
   - `stale-path-fix`: current implementation/test path is known.
   - `manifest-owned-generated`: generated evidence is owned by a generator or
     manifest but is not present on disk.
   - `historical-source`: old probe/source path is useful context but not a
     current file.
   - `false-positive-fragment`: scanner matched a prose fragment, glob, or
     command example rather than a required file.
   - `keep-separate`: similar notes cover distinct APIs, fields, or gates.

## Fixed Stale Paths

| Knowledge | Old reference | Current evidence | Action |
|---|---|---|---|
| `knowledge/comp/recipe-draft-3d-profile.md` | `internal/aep/comp_settings_shipgate_test.go` | `internal/aep_test/comp_settings_shipgate_test.go` exists | Updated path. |
| `knowledge/effects/pseudo-effect-continuation-handoff.md` | `internal/aep/pseudo_effect_shipgate_test.go` | `internal/aep_test/pseudo_effect_shipgate_test.go` exists | Updated path. |
| `knowledge/workflow/ae-drops-unknown-chunks-on-resave.md` | `internal/aep/custom_chunk_probe_shipgate_test.go` | `internal/aep_test/custom_chunk_probe_shipgate_test.go` exists | Updated path. |
| `knowledge/workflow/re-fixture.md` | `internal/aep/testutil_shipgate_test.go` | `internal/aep_test/testutil_shipgate_test.go` exists | Updated path. |
| `knowledge/workflow/re-fixture.md` | `internal/aep/mg_text_style_shipgate_test.go` | `internal/aep_test/mg_text_style_shipgate_test.go` exists | Updated path. |
| `knowledge/workflow/ae25-acceptance-gate.md` | `internal/serializer/templates/2020_dummy_comp.aep` | `internal/serializer/templates/project/2020_dummy_comp.aep` exists | Updated path. |

## Generated Evidence Notes

| Knowledge | Missing evidence path | Current evidence | Action |
|---|---|---|---|
| `knowledge/effects/add-effect-splice-re.md` | `test_data/generated/fixtures/re_effect_library*.aep` | `scripts/fixtures/fixtures_manifest.json` owns the generated AEPs; committed truth is `internal/serializer/templates/effects/effect_*.bin` | Added regenerate-before-use note. |
| `knowledge/effects/effect-param-elision-synthesis-lite.md` | `test_data/generated/fixtures/re_effect_param_types*.aep` | Manifest owns generated AEPs; committed truth is `internal/serializer/templates/effects/effectparam_*.bin` | Added regenerate-before-use note. |
| `knowledge/layer/camera-light-layer-create-re.md` | `test_data/generated/fixtures/re_camera_iris.aep`, `re_light_color.aep` | Current truth is `internal/serializer/templates/options/camera_iris_leaves.bin` and `light_color_leaf.bin`; ship gates remain under `internal/aep_test` | Marked generated AEPs historical and fixed template paths. |
| `knowledge/shape/trim-paths-vector-filter-re.md` | `data/reference/shape/v2_2_wiggle_modulation.aep` | Current truth is `internal/serializer/templates/shapes/{roughen_points,correlation,temporal_phase,spatial_phase}_leaf.bin`; current gates are `mg_wiggle_mod*` | Marked source AEP historical and named current embedded templates. |

## Merge Findings

No body-level merge was applied in this pass.

High-similarity pairs such as `recipe-label` comp/layer, `recipe-comment`
comp/layer, `recipe-motion-blur` comp/layer, and many shape recipe enum notes
share template wording, but the current evidence points at distinct recipe
fields, examples, tests, or binary properties. They are not safe merge targets
without first designing a recipe-family index note that preserves all per-field
facts.

The right future merge pattern is:

- create a family note only when it states a reusable rule with current evidence;
- keep per-field notes when they route to distinct examples, tests, match names,
  binary offsets, or validation domains;
- delete or redirect an old note only after its full actionable content appears
  in the family note and all path references are current.
