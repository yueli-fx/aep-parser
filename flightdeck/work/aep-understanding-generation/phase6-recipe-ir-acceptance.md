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

Seventh follow-up completed:

- Text recipes now support a first static text-style slice:
  - `text_style.font_size` sets `Layer.SetRunFontSize`.
  - `text_style.tracking` sets `Layer.SetRunTracking`.
  - `text_style.justification` sets `Layer.SetParagraphJustification`.
- Optional `run_index` and `paragraph_index` default to `0`; validation
  rejects negative indexes, non-positive font sizes, and unsupported
  justifications.
- `expected_profile.text_styles[]` can assert font size, tracking, and
  paragraph justification for a named text layer. Justification comparison is
  enum-like and case-insensitive because profile display values are title case
  (`Center`) while recipe input is lowercase (`center`).
- `examples/recipes/minimal-text-shape.json` now styles the `Title` layer and
  asserts those style values in the embedded profile contract.
- Verification:
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-textstyle.aep -json` returned
    valid and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request: `tmp_debug/aeoracle/minimal_text_shape_textstyle/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_textstyle/aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs: `f000000.png`, `f000060.png`, `f000120.png`

Eighth follow-up completed:

- Text recipes now support additional static text-style controls:
  - `text_style.fill_color` sets `Layer.SetRunFillColor`.
  - `text_style.faux_bold` sets `Layer.SetRunFauxBold`.
  - `text_style.faux_italic` sets `Layer.SetRunFauxItalic`.
  - `text_style.apply_stroke` sets `Layer.SetRunApplyStroke`.
  - `text_style.stroke_color` sets `Layer.SetRunStrokeColor`.
  - `text_style.stroke_width` sets `Layer.SetRunStrokeWidth`.
- `expected_profile.text_styles[]` can assert fill color, faux bold/italic,
  apply stroke, stroke color, and stroke width in addition to the earlier font
  size, tracking, and paragraph justification checks. Text-style colors compare
  against the profile's normalized `0..1` RGBA values.
- `examples/recipes/minimal-text-shape.json` now exercises blue text fill,
  faux bold/italic, and magenta text stroke with embedded profile checks.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-textstyle2.aep -json` returned
    valid and all `profile_checks` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_text_shape_textstyle2/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_textstyle2/aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs: `f000000.png`, `f000060.png`, `f000120.png`

Ninth follow-up completed:

- Recipe embedded profile contracts now support `expected_profile.keyframes[]`
  for property keyframe checks:
  - lookup by `layer_name` + property `match_name`;
  - checks keyframe count;
  - checks each expected keyframe's time and value.
- Recipe validation now rejects unsorted `transform.position_keyframes`, matching
  the original schema rule that keyframes must be sorted by time.
- `examples/recipes/minimal-text-shape.json` now animates the `Title` layer's
  Position across three keyframes and asserts the parsed `ADBE Position`
  keyframes in `expected_profile.keyframes[]`.
- Note: recipe Position keyframes are authored as 2D `[x,y]`, while profile
  `ADBE Position` keyframe values read back as 3D `[x,y,0]`; see
  `knowledge/layer/recipe-position-keyframe-profile-3d.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-keyframes.aep -json` returned
    valid and all `profile_checks` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_text_shape_keyframes/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_keyframes/aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000030.png`, `f000060.png`, `f000090.png`,
      `f000120.png`

Tenth follow-up completed:

- Recipes now support `transform.opacity_keyframes[]` as scalar keyframes in
  AE-style percent units `0..100`.
- Validation rejects opacity keyframes outside the comp duration, unsorted
  opacity keyframes, and opacity keyframe values outside `0..100`.
- `expected_profile.keyframes[]` now exercises both Position and Opacity on the
  updated shape/text example. Note: recipe Opacity keyframes are authored as
  percent values, while profile `ADBE Opacity` keyframe values read back as
  unit opacity `0..1`; see
  `knowledge/layer/recipe-opacity-keyframe-profile-units.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-opacity-keyframes.aep -json`
    returned valid and all `profile_checks` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_text_shape_opacity_keyframes/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_opacity_keyframes/aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000030.png`, `f000060.png`, `f000090.png`,
      `f000120.png`

Eleventh follow-up completed:

- Recipes now support `transform.scale_keyframes[]` as 2D vector keyframes in
  AE-style percent units `[x,y]`.
- Validation rejects scale keyframes outside the comp duration, unsorted scale
  keyframes, and malformed scale keyframe vectors.
- `expected_profile.keyframes[]` now exercises Position, Scale, and Opacity on
  the updated shape/text example. Note: recipe Scale keyframes are authored as
  2D percent values, while profile `ADBE Scale` keyframe values read back as
  3D unit scale `[x/100,y/100,1]`; see
  `knowledge/layer/recipe-scale-keyframe-profile-units.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-scale-keyframes.aep -json`
    returned valid and all `profile_checks` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_text_shape_scale_keyframes/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_scale_keyframes/aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000030.png`, `f000060.png`, `f000090.png`,
      `f000120.png`

Twelfth follow-up completed:

- Recipes now support `transform.rotation_keyframes[]` as scalar keyframes in
  degree units.
- Validation rejects rotation keyframes outside the comp duration and unsorted
  rotation keyframes. Rotation values intentionally have no value range limit,
  matching AE angle behavior.
- `expected_profile.keyframes[]` now exercises Position, Scale, Rotation, and
  Opacity on the updated shape/text example. `ADBE Rotate Z` profile keyframe
  values read back in degrees, matching recipe input.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-rotation-keyframes.aep -json`
    returned valid and all `profile_checks` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_text_shape_rotation_keyframes/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_rotation_keyframes/aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000030.png`, `f000060.png`, `f000090.png`,
      `f000120.png`

Thirteenth follow-up completed:

- Recipes now support `transform.anchor_point_keyframes[]` as 2D vector
  keyframes.
- Validation rejects anchor point keyframes outside the comp duration, unsorted
  anchor point keyframes, and malformed anchor point vectors.
- `expected_profile.keyframes[]` now exercises Position, Anchor Point, Scale,
  Rotation, and Opacity on the updated shape/text example. Note: recipe Anchor
  Point keyframes are authored as 2D `[x,y]`, while profile
  `ADBE Anchor Point` keyframe values read back as 3D `[x,y,0]`; see
  `knowledge/layer/recipe-anchor-point-keyframe-profile-3d.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-anchor-keyframes.aep -json`
    returned valid and all `profile_checks` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_text_shape_anchor_keyframes/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_anchor_keyframes/aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000030.png`, `f000060.png`, `f000090.png`,
      `f000120.png`

Fourteenth follow-up completed:

- Shape recipes now support a first shape-filter slice:
  - `shape.trim.start` sets `ADBE Vector Trim Start`.
  - `shape.trim.end` sets `ADBE Vector Trim End`.
  - `shape.trim.offset` sets `ADBE Vector Trim Offset`.
- Validation rejects trim start/end outside `0..100`; trim offset intentionally
  has no value range limit.
- `examples/recipes/minimal-text-shape.json` now includes Trim Paths on the
  `Underline` layer and asserts the three Trim properties through
  `expected_profile.properties[]`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-trim.aep -json` returned valid
    and all `profile_checks` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - AE 2025 render oracle passed:
    - request: `tmp_debug/aeoracle/minimal_text_shape_trim/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_trim/aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000030.png`, `f000060.png`, `f000090.png`,
      `f000120.png`

Fifteenth follow-up completed:

- Shape recipes now support `shape.round_corners.radius`, a static Round
  Corners filter control.
- Validation rejects negative round-corners radius values.
- Capability reporting records both:
  - `VectorGroup.AddRoundCorners`
  - `RoundCornersNode.SetRadius`
- `examples/recipes/minimal-text-shape.json` now includes Round Corners on the
  `Underline` layer and asserts `ADBE Vector RoundCorner Radius` through
  `expected_profile.properties[]`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-round-corners.aep -json`
    returned valid and all `profile_checks` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_text_shape_round_corners/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_round_corners/aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000030.png`, `f000060.png`, `f000090.png`,
      `f000120.png`

Sixteenth follow-up completed:

- Shape recipes now support `shape.offset_paths`, including:
  - `amount` -> `ADBE Vector Offset Amount`
  - `line_join` -> `ADBE Vector Offset Line Join`
  - `miter_limit` -> `ADBE Vector Offset Miter Limit`
  - `copies` -> `ADBE Vector Offset Copies`
  - `copy_offset` -> `ADBE Vector Offset Copy Offset`
- Validation rejects unsupported `line_join` values, `miter_limit < 1`, and
  `copies < 1`; offset amount and copy offset intentionally have no recipe
  range limit beyond the underlying writer.
- Capability reporting records:
  - `VectorGroup.AddOffsetPaths`
  - `OffsetPathsNode.SetAmount`
  - `OffsetPathsNode.SetLineJoin`
  - `OffsetPathsNode.SetMiterLimit`
  - `OffsetPathsNode.SetCopies`
  - `OffsetPathsNode.SetCopyOffset`
- `examples/recipes/minimal-text-shape.json` now includes Offset Paths on the
  `Underline` layer and asserts all five Offset Paths profile properties.
  Note: recipe `line_join` is authored as a string, while profile
  `ADBE Vector Offset Line Join` reads back as numeric enum `1/2/3`; see
  `knowledge/shape/recipe-offset-paths-profile-enums.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-offset-paths.aep -json`
    returned valid and all `profile_checks` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_text_shape_offset_paths/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_offset_paths/aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000030.png`, `f000060.png`, `f000090.png`,
      `f000120.png`

Seventeenth follow-up completed:

- Shape recipes now support `shape.zigzag`, including:
  - `size` -> `ADBE Vector Zigzag Size`
  - `detail` -> `ADBE Vector Zigzag Detail`
  - `points` -> `ADBE Vector Zigzag Points`
- Validation rejects negative size/detail values and unsupported `points`
  values; valid points are `corner` and `smooth`.
- Capability reporting records:
  - `VectorGroup.AddZigZag`
  - `ZigZagNode.SetSize`
  - `ZigZagNode.SetDetail`
  - `ZigZagNode.SetPoints`
- `examples/recipes/minimal-text-shape.json` now includes ZigZag on the
  `Underline` layer and asserts all three ZigZag profile properties. Note:
  recipe `points` is authored as a string, while profile
  `ADBE Vector Zigzag Points` reads back as numeric enum `1/2`; see
  `knowledge/shape/recipe-zigzag-profile-enums.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-zigzag.aep -json` returned
    valid and all `profile_checks` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_text_shape_zigzag/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_zigzag/aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000030.png`, `f000060.png`, `f000090.png`,
      `f000120.png`

Next generation work should broaden recipe coverage in small proven slices:
additional transform keyframe channels/ease where writer support exists,
expression support, or shape filters. Do not start automated correction loops.
