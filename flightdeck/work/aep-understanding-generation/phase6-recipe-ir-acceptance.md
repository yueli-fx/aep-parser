# Phase 6 Recipe IR Acceptance

## Inputs

- Recipe: `examples/recipes/minimal-text-shape.json`
- Generated AEP: `tmp_debug/recipes/minimal-text-shape.aep`
- Render request: `tmp_debug/aeoracle/minimal_recipe/request.json`
- AE executable: `E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe`

## Commands

```powershell
go run ./cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json -out tmp_debug\recipes\minimal-text-shape.aep -json
go run ./cmd/aeoracle plan -aep tmp_debug\recipes\minimal-text-shape.aep -out tmp_debug\aeoracle\minimal_recipe -json
go run ./cmd/aeoracle render -request tmp_debug\aeoracle\minimal_recipe\request.json -ae 'E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe' -timeout-sec 600
```

## Result

- `cmd/aeprecipe compile` returned a valid report and wrote
  `tmp_debug/recipes/minimal-text-shape.aep`.
- `cmd/aeoracle plan` selected 3 sentinel frames:
  - `f000000`, frame 0, comp start
  - `f000060`, frame 60, quiet midpoint
  - `f000120`, frame 120, comp end
- `cmd/aeoracle render` passed on AE `25.1x68`.
- Done marker: `tmp_debug/aeoracle/minimal_recipe/aeoracle_render.done` = `ok`
- Metadata: `tmp_debug/aeoracle/minimal_recipe/metadata.json`, status `ok`
- PNG outputs:
  - `tmp_debug/aeoracle/minimal_recipe/f000000.png`
  - `tmp_debug/aeoracle/minimal_recipe/f000060.png`
  - `tmp_debug/aeoracle/minimal_recipe/f000120.png`

## Interpretation

The minimal recipe IR can now compile a one-comp text/shape recipe into an AEP
using exported writer APIs, and the generated AEP opens and renders through the
generic AE oracle. This proves the first deterministic generation slice, not an
AI correction loop.

Follow-up completed after this gate:

- `cmd/aeprecipe` now loads `docs/capabilities.json` through reusable
  `internal/capindex` instead of using a temporary static map.
- Recipe reports include used capabilities and downgrades.
- At this follow-up point, effect requests were explicitly refused instead of
  silently dropped until the next slice materialized them.
- The stale `SetLayerTransform` capability boundary was corrected and
  `docs/capabilities.{json,md}` were regenerated.

Second follow-up completed:

- `examples/recipes/minimal-text-effect.json` adds a supported built-in effect
  (`ADBE Gaussian Blur 2`) to a text layer.
- `internal/recipe` now compiles effects with a two-stage path: build the base
  project, `Reopen` it to materialize parsed layer backing, then call
  `AddEffect` on the indexed parsed layer.
- `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-effect.json -out
  tmp_debug\recipes\minimal-text-effect.aep -json` returned valid.
- `cmd/aepdissect -json tmp_debug\recipes\minimal-text-effect.aep` showed
  `ADBE Gaussian Blur 2` on `Blurred Title`.
- AE 2025 render oracle passed:
  - request: `tmp_debug/aeoracle/minimal_text_effect/request.json`
  - done marker: `tmp_debug/aeoracle/minimal_text_effect/aeoracle_render.done`
    = `ok`
  - metadata: status `ok`, AE `25.1x68`
  - PNG outputs: `f000000.png`, `f000030.png`, `f000060.png`

Third follow-up completed:

- `Effect.params[]` supports static parameter values:
  - numeric scalar
  - boolean, normalized to `1.0` / `0.0`
  - numeric arrays, normalized from JSON arrays to `[]float64`
- `internal/recipe` validates param match names and value types, records
  `SetEffectParam` in capability reports, and rejects unsupported value types
  such as strings.
- `examples/recipes/minimal-text-effect.json` now sets Gaussian Blur params:
  - `ADBE Gaussian Blur 2-0001` = `25`
  - `ADBE Gaussian Blur 2-0002` = `2`
  - `ADBE Gaussian Blur 2-0003` = `true` → `1`
- `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-effect.json -out
  tmp_debug\recipes\minimal-text-effect-params.aep -json` returned valid.
- `cmd/aepdissect -json tmp_debug\recipes\minimal-text-effect-params.aep`
  showed the three static values as `25`, `2`, and `1`.
- AE 2025 render oracle passed:
  - request: `tmp_debug/aeoracle/minimal_text_effect_params/request.json`
  - done marker:
    `tmp_debug/aeoracle/minimal_text_effect_params/aeoracle_render.done` = `ok`
  - metadata: status `ok`, AE `25.1x68`
  - PNG outputs: `f000000.png`, `f000030.png`, `f000060.png`

Fourth follow-up completed:

- Recipes now support an embedded `expected_profile` contract. The compiler
  builds a profile from the exact AEP bytes it is about to write, records
  `profile_checks` in the report, and refuses to write the final output when a
  profile contract check fails.
- Covered checks:
  - `comp_count`
  - `layer_count`
  - `text_layer_count`
  - `shape_layer_count`
  - expected effect on a named layer
  - expected static effect parameter values
- `examples/recipes/minimal-text-shape.json` now asserts comp/layer/text/shape
  counts.
- `examples/recipes/minimal-text-effect.json` now asserts comp/layer/text
  counts plus `ADBE Gaussian Blur 2` and its three static params.
- Example compiles:
  - `tmp_debug/recipes/minimal-text-shape-contract.aep`
  - `tmp_debug/recipes/minimal-text-effect-contract.aep`
- AE 2025 render oracle passed for the effect contract output:
  - request: `tmp_debug/aeoracle/minimal_text_effect_contract/request.json`
  - done marker:
    `tmp_debug/aeoracle/minimal_text_effect_contract/aeoracle_render.done` =
    `ok`
  - metadata: status `ok`, AE `25.1x68`
  - PNG outputs: `f000000.png`, `f000030.png`, `f000060.png`

Fifth follow-up completed:

- Shape recipes now support static stroke controls:
  - `shape.stroke.color` accepts 3/4-channel recipe RGBA color values.
  - `shape.stroke.width` sets `StrokeNode.SetWidth`.
  - `shape.stroke.opacity` sets `StrokeNode.SetOpacity`.
- `internal/recipe` validates stroke color arity/range, non-negative width,
  and opacity in `0..100`, records these capabilities:
  - `VectorGroup.AddStroke`
  - `StrokeNode.SetColor`
  - `StrokeNode.SetWidth`
  - `StrokeNode.SetOpacity`
- `expected_profile.properties[]` can now assert static profile properties on
  a named layer. The lookup covers both layer-level properties and primitive
  shape properties.
- `examples/recipes/minimal-text-shape.json` now includes a red stroke on the
  underline and asserts stroke color/width/opacity via embedded profile checks.
  Note: profile color static values are read back in AE profile order; the
  recipe RGBA red `[255, 0, 0, 255]` appears in the profile as
  `[255, 255, 0, 0]`.
- Verification:
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-stroke.aep -json` returned valid
    and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request: `tmp_debug/aeoracle/minimal_text_shape_stroke/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_stroke/aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs: `f000000.png`, `f000060.png`, `f000120.png`

Sixth follow-up completed:

- Shape recipes now support additional static detail controls:
  - `shape.position` sets local primitive position for rect/ellipse shapes.
  - `shape.roundness` sets rectangle corner roundness.
  - `shape.fill_opacity` sets `FillNode.SetOpacity`.
- `internal/recipe` records these capabilities:
  - `RectNode.SetPosition` / `EllipseNode.SetPosition`
  - `RectNode.SetRoundness`
  - `FillNode.SetOpacity`
- Validation rejects malformed shape position vectors, negative rectangle
  roundness, ellipse roundness, and fill opacity outside `0..100`.
- `examples/recipes/minimal-text-shape.json` now asserts rect position,
  roundness, and fill opacity via `expected_profile.properties[]`.
- Verification:
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-detail.aep -json` returned
    valid and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request: `tmp_debug/aeoracle/minimal_text_shape_detail/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_detail/aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs: `f000000.png`, `f000060.png`, `f000120.png`

Next generation work should broaden recipe coverage in small proven slices,
starting with text style fields, expression / keyframe coverage, or additional
shape filters, each gated by embedded profile checks and render-oracle evidence.
