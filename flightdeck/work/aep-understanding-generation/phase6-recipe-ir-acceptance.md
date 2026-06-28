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

Next generation work should broaden recipe coverage in small proven slices,
starting with a recipe-to-profile expected-output contract or richer shape/text
controls, each gated by profile and render-oracle evidence.
