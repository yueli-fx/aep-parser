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

Eighteenth follow-up completed:

- Shape recipes now support `shape.pucker_bloat.amount`, mapped to
  `ADBE Vector PuckerBloat Amount`.
- The Pucker & Bloat amount is intentionally unrestricted at recipe validation
  level, matching the underlying writer; negative values pucker and positive
  values bloat.
- Capability reporting records:
  - `VectorGroup.AddPuckerBloat`
  - `PuckerBloatNode.SetAmount`
- `examples/recipes/minimal-text-shape.json` now includes Pucker & Bloat on
  the `Underline` layer and asserts `ADBE Vector PuckerBloat Amount` through
  `expected_profile.properties[]`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-pucker-bloat.aep -json`
    returned valid and all `profile_checks` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_text_shape_pucker_bloat/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_pucker_bloat/aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000030.png`, `f000060.png`, `f000090.png`,
      `f000120.png`

Nineteenth follow-up completed:

- Shape recipes now support `shape.twist`, including:
  - `angle` -> `ADBE Vector Twist Angle`
  - `center` -> `ADBE Vector Twist Center`
- Validation rejects malformed `center` vectors; angle values intentionally
  have no recipe range limit, matching AE angle behavior.
- Capability reporting records:
  - `VectorGroup.AddTwist`
  - `TwistNode.SetAngle`
  - `TwistNode.SetCenter`
- `examples/recipes/minimal-text-shape.json` now includes Twist on the
  `Underline` layer and asserts both Twist profile properties.
- During Twist AE validation, the generic render oracle exposed a harness bug:
  metadata could say five frames were rendered while only two PNG files existed.
  Root cause: `scripts/aeoracle_render.jsx` did not wait for
  `saveFrameToPng` output files, and `cmd/aeoracle render` did not validate
  requested frame artifacts after `done_path = ok`. The harness now waits for
  each PNG and the CLI verifies all requested PNG files exist and are non-empty;
  see `knowledge/workflow/aeoracle-saveframe-output-validation.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe ./cmd/aeoracle` passed.
  - `cmd/aeprecipe compile -recipe examples\recipes\minimal-text-shape.json
    -out tmp_debug\recipes\minimal-text-shape-twist.aep -json` returned valid
    and all `profile_checks` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - AE 2025 render oracle passed after harness hardening:
    - request:
      `tmp_debug/aeoracle/minimal_text_shape_twist_retry/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_text_shape_twist_retry/aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000030.png`, `f000060.png`, `f000090.png`,
      `f000120.png`

Twentieth follow-up completed:

- Shape recipes now support `shape.wiggle_paths`, including:
  - `size` -> `ADBE Vector Roughen Size`
  - `detail` -> `ADBE Vector Roughen Detail`
  - `wiggles_per_second` -> `ADBE Vector Temporal Freq`
  - `random_seed` -> `ADBE Vector Random Seed`
  - `points` -> `ADBE Vector Roughen Points`
  - `correlation` -> `ADBE Vector Correlation`
  - `temporal_phase` -> `ADBE Vector Temporal Phase`
  - `spatial_phase` -> `ADBE Vector Spatial Phase`
- Validation rejects invalid `points` values and out-of-range `correlation`
  values. Other scalar fields intentionally follow the underlying writer's
  looser bounds.
- Capability reporting records:
  - `VectorGroup.AddWigglePaths`
  - `WigglePathsNode.SetSize`
  - `WigglePathsNode.SetDetail`
  - `WigglePathsNode.SetWigglesPerSecond`
  - `WigglePathsNode.SetRandomSeed`
  - `WigglePathsNode.SetPoints`
  - `WigglePathsNode.SetCorrelation`
  - `WigglePathsNode.SetTemporalPhase`
  - `WigglePathsNode.SetSpatialPhase`
- `examples/recipes/minimal-shape-wiggle-paths.json` is a dedicated Wiggle
  Paths recipe example and asserts all eight profile properties. Do not stack
  every path filter into `minimal-text-shape.json`; the combined filter set can
  make AE 2025 fail to produce `saveFrameToPng` output for the first frame.
- During Wiggle Paths AE validation, the generic render oracle exposed two
  harness gaps:
  - `scripts/aeoracle_render.jsx` used a hard-coded 30-second per-frame PNG
    wait despite the CLI having a much larger `-timeout-sec`; render requests
    now carry `frame_timeout_ms` with a 120-second default.
  - AE's long-running `Executing Script ...` progress dialog was treated as an
    unknown modal; `scripts/ae_dialog_rules.json` now classifies it as a known
    `Ignore` dialog. The same rules file also classifies the
    `Scripting plugin is not installed` dialog as `Abort` instead of blindly
    pressing OK.
  See `knowledge/workflow/aeoracle-slow-frame-script-progress.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe ./internal/aeoracle
    ./cmd/aeoracle` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - `Invoke-Pester scripts\ae_run.Tests.ps1` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-shape-wiggle-paths.json -out
    tmp_debug\recipes\minimal-shape-wiggle-paths.aep -json` returned valid and
    all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_shape_wiggle_paths/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_shape_wiggle_paths/aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Twenty-first follow-up completed:

- Shape recipes now support `shape.wiggle_transform`, including:
  - `anchor` -> `ADBE Vector Wiggler Anchor`
  - `position` -> `ADBE Vector Wiggler Position`
  - `scale` -> `ADBE Vector Wiggler Scale`
  - `rotation` -> `ADBE Vector Wiggler Rotation`
  - `wiggles_per_second` -> `ADBE Vector Xform Temporal Freq`
  - `random_seed` -> `ADBE Vector Random Seed`
  - `correlation` -> `ADBE Vector Correlation`
  - `temporal_phase` -> `ADBE Vector Temporal Phase`
  - `spatial_phase` -> `ADBE Vector Spatial Phase`
- Validation rejects malformed 2D vector fields and out-of-range
  `correlation`; amplitude scalars otherwise follow the underlying writer.
- Capability reporting records:
  - `VectorGroup.AddWiggleTransform`
  - `WigglerTransform.SetAnchor`
  - `WigglerTransform.SetPosition`
  - `WigglerTransform.SetScale`
  - `WigglerTransform.SetRotation`
  - `WiggleTransformNode.SetWigglesPerSecond`
  - `WiggleTransformNode.SetRandomSeed`
  - `WiggleTransformNode.SetCorrelation`
  - `WiggleTransformNode.SetTemporalPhase`
  - `WiggleTransformNode.SetSpatialPhase`
- `examples/recipes/minimal-shape-wiggle-transform.json` is a dedicated Wiggle
  Transform recipe example and asserts all nine profile properties. The profile
  names intentionally mix `ADBE Vector Wiggler ...`, `ADBE Vector Xform
  Temporal Freq`, and shared modulation names; see
  `knowledge/shape/recipe-wiggle-transform-profile-names.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-shape-wiggle-transform.json -out
    tmp_debug\recipes\minimal-shape-wiggle-transform.aep -json` returned valid
    and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_shape_wiggle_transform/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_shape_wiggle_transform/aeoracle_render.done`
      = `ok`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Twenty-second follow-up completed:

- Shape recipes now support `shape.repeater`, including:
  - `copies` -> `ADBE Vector Repeater Copies`
  - `offset` -> `ADBE Vector Repeater Offset`
  - `order` -> `ADBE Vector Repeater Order`
  - `anchor` -> `ADBE Vector Repeater Anchor`
  - `position` -> `ADBE Vector Repeater Position`
  - `scale` -> `ADBE Vector Repeater Scale`
  - `rotation` -> `ADBE Vector Repeater Rotation`
  - `start_opacity` -> `ADBE Vector Repeater Opacity 1`
  - `end_opacity` -> `ADBE Vector Repeater Opacity 2`
- Validation rejects `copies < 1`, unsupported `order` values, malformed 2D
  transform vectors, and start/end opacity outside `0..100`.
- Capability reporting records:
  - `VectorGroup.AddRepeater`
  - `RepeaterNode.SetCopies`
  - `RepeaterNode.SetOffset`
  - `RepeaterNode.SetOrder`
  - `RepeaterTransform.SetAnchor`
  - `RepeaterTransform.SetPosition`
  - `RepeaterTransform.SetScale`
  - `RepeaterTransform.SetRotation`
  - `RepeaterTransform.SetStartOpacity`
  - `RepeaterTransform.SetEndOpacity`
- `examples/recipes/minimal-shape-repeater.json` is a dedicated Repeater
  recipe example and asserts all nine profile properties. Recipe `order` is
  authored as `below` / `above`, while profile readback is numeric `1` / `2`;
  see `knowledge/shape/recipe-repeater-profile-enums.md`.
- The AE 2025 run did not reproduce the earlier `Scripting plugin is not
  installed` modal. Current `scripts/ae_dialog_rules.json` already classifies
  that dialog as `Abort`, so a recurrence exits fast with forensics instead of
  dismissing the environment failure.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-shape-repeater.json -out
    tmp_debug\recipes\minimal-shape-repeater.aep -json` returned valid and
    all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_shape_repeater/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_shape_repeater/aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Twenty-third follow-up completed:

- Shape recipes now support `shape.merge_paths`, including:
  - `type` -> `ADBE Vector Merge Type`
- Valid `type` values are `merge`, `add`, `subtract`, `intersect`, and
  `exclude`; validation rejects unsupported values.
- Capability reporting records:
  - `VectorGroup.AddMergePaths`
  - `MergePathsNode.SetType`
- `examples/recipes/minimal-shape-merge-paths.json` is a dedicated Merge Paths
  recipe example and asserts `ADBE Vector Merge Type`. Recipe `type` is
  authored as a string enum, while profile readback is numeric `1..5`; see
  `knowledge/shape/recipe-merge-paths-profile-enums.md`.
- Boundary: current recipe IR has one `shape` primitive per shape layer, so
  this slice proves node/type generation and AE acceptance; visible multi-path
  boolean composition needs a later multi-shape or vector-group recipe shape.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-shape-merge-paths.json -out
    tmp_debug\recipes\minimal-shape-merge-paths.aep -json` returned valid and
    all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_shape_merge_paths/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_shape_merge_paths/aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Twenty-fourth follow-up completed:

- Shape recipes now support polystar geometry through `shape.kind`:
  - `star` -> default `ADBE Vector Shape - Star`
  - `polygon` -> `ADBE Vector Star Type` = `2`
- Supported polystar fields:
  - `points` -> `ADBE Vector Star Points`
  - `position` -> `ADBE Vector Star Position`
  - `rotation` -> `ADBE Vector Star Rotation`
  - `inner_radius` -> `ADBE Vector Star Inner Radius`
  - `outer_radius` -> `ADBE Vector Star Outer Radius`
  - `inner_roundness` -> `ADBE Vector Star Inner Roundess`
  - `outer_roundness` -> `ADBE Vector Star Outer Roundess`
- Validation rejects `points < 3`, malformed `position`, negative radii, and
  legacy rect-only `roundness` on polystar shapes.
- Capability reporting records:
  - `VectorGroup.AddStar`
  - `StarNode.SetStarType`
  - `StarNode.SetPoints`
  - `StarNode.SetPosition`
  - `StarNode.SetRotation`
  - `StarNode.SetInnerRadius`
  - `StarNode.SetOuterRadius`
  - `StarNode.SetInnerRoundness`
  - `StarNode.SetOuterRoundness`
- `examples/recipes/minimal-shape-polystar.json` is a dedicated polystar
  recipe example and asserts all eight profile-visible `ADBE Vector Star ...`
  properties. See `knowledge/shape/recipe-polystar-profile-fields.md` for enum
  and profile spelling notes.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-shape-polystar.json -out
    tmp_debug\recipes\minimal-shape-polystar.aep -json` returned valid and
    all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_shape_polystar/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_shape_polystar/aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Twenty-fifth follow-up completed:

- Shape stroke recipes now support `shape.stroke.dashes`, including:
  - `dash` -> `ADBE Vector Stroke Dash 1`
  - `gap` -> `ADBE Vector Stroke Gap 1`
- Setting either value enables the writer's Dashes group. The current writer
  models exactly one Dash/Gap pair; additional pairs and dash offset are not
  modeled yet.
- Validation rejects negative `dash` and `gap` values.
- Capability reporting records:
  - `StrokeDashes.SetDash`
  - `StrokeDashes.SetGap`
- `examples/recipes/minimal-shape-stroke-dashes.json` is a dedicated dashed
  stroke recipe example and asserts the profile-visible Dash 1 / Gap 1
  properties. See `knowledge/shape/recipe-stroke-dashes-profile-fields.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-shape-stroke-dashes.json -out
    tmp_debug\recipes\minimal-shape-stroke-dashes.aep -json` returned valid
    and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_shape_stroke_dashes/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_shape_stroke_dashes/aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Twenty-sixth follow-up completed:

- Shape stroke recipes now support line style fields:
  - `line_cap` -> `ADBE Vector Stroke Line Cap`
  - `line_join` -> `ADBE Vector Stroke Line Join`
  - `miter_limit` -> `ADBE Vector Stroke Miter Limit`
- Valid `line_cap` values are `butt`, `round`, and `projecting`.
- Valid `line_join` values are `miter`, `round`, and `bevel`.
- Validation rejects unsupported line cap / line join values and
  `miter_limit < 1`.
- Capability reporting records:
  - `StrokeNode.SetLineCap`
  - `StrokeNode.SetLineJoin`
  - `StrokeNode.SetMiterLimit`
- `examples/recipes/minimal-shape-stroke-style.json` is a dedicated stroke
  style recipe example and asserts all three profile-visible stroke style
  properties. See `knowledge/shape/recipe-stroke-style-profile-enums.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-shape-stroke-style.json -out
    tmp_debug\recipes\minimal-shape-stroke-style.aep -json` returned valid
    and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_shape_stroke_style/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_shape_stroke_style/aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Twenty-seventh follow-up completed:

- Shape stroke recipes now support `shape.stroke.taper`, including:
  - `start_length` -> `ADBE Vector Taper Start Length`
  - `end_length` -> `ADBE Vector Taper End Length`
  - `start_width` -> `ADBE Vector Taper Start Width`
  - `end_width` -> `ADBE Vector Taper End Width`
  - `start_ease` -> `ADBE Vector Taper Start Ease`
  - `end_ease` -> `ADBE Vector Taper End Ease`
- Capability reporting records:
  - `StrokeTaper.SetStartLength`
  - `StrokeTaper.SetEndLength`
  - `StrokeTaper.SetStartWidth`
  - `StrokeTaper.SetEndWidth`
  - `StrokeTaper.SetStartEase`
  - `StrokeTaper.SetEndEase`
- Boundary: the writer models the always-active percent-mode taper controls;
  pixel-mode mirror streams and the length unit enum are not in recipe IR yet.
- `examples/recipes/minimal-shape-stroke-taper.json` is a dedicated stroke
  taper recipe example and asserts all six profile-visible taper properties.
  See `knowledge/shape/recipe-stroke-taper-profile-fields.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-shape-stroke-taper.json -out
    tmp_debug\recipes\minimal-shape-stroke-taper.aep -json` returned valid
    and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug/aeoracle/minimal_shape_stroke_taper/request.json`
    - done marker:
      `tmp_debug/aeoracle/minimal_shape_stroke_taper/aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Twenty-eighth follow-up completed:

- Shape stroke recipes now support `shape.stroke.wave`, including:
  - `amount` -> `ADBE Vector Taper Wave Amount`
  - `wavelength` -> `ADBE Vector Taper Wavelength`
  - `phase` -> `ADBE Vector Taper Wave Phase`
- Capability reporting records:
  - `StrokeWave.SetAmount`
  - `StrokeWave.SetWavelength`
  - `StrokeWave.SetPhase`
- Boundary: the writer models the wavelength-mode stroke wave controls;
  pixel-mode mirror streams and the wave unit enum are not in recipe IR yet.
- `examples/recipes/minimal-shape-stroke-wave.json` is a dedicated stroke wave
  recipe example and asserts all three profile-visible wave properties. See
  `knowledge/shape/recipe-stroke-wave-profile-fields.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-shape-stroke-wave.json -out
    tmp_debug\recipes\minimal-shape-stroke-wave.aep -json` returned valid and
    all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_shape_stroke_wave\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_shape_stroke_wave\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Twenty-ninth follow-up completed:

- Composition recipes now honor `background_color`.
- Validation requires exactly three RGB channels in 0..255 units.
- Capability reporting records:
  - `SetBGColor`
- Boundary: composition background color is not exposed in the current stable
  profile schema, so the contract uses compiled AEP readback plus AE render
  acceptance rather than `expected_profile.properties[]`.
- `examples/recipes/minimal-comp-background-color.json` is a dedicated
  no-layer comp recipe with a non-black background. See
  `knowledge/composition/recipe-background-color.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-comp-background-color.json -out
    tmp_debug\recipes\minimal-comp-background-color.aep -json` returned valid
    and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_comp_background_color\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_comp_background_color\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Thirtieth follow-up completed:

- Composition recipes now support `motion_blur` shutter/sample settings:
  - `shutter_angle` -> profile `motion_blur.shutter_angle_degrees`
  - `shutter_phase` -> profile `motion_blur.shutter_phase`
  - `adaptive_sample_limit` -> profile `motion_blur.adaptive_sample_limit`
  - `samples_per_frame` -> profile `motion_blur.samples_per_frame`
- Capability reporting records:
  - `SetShutterAngle`
  - `SetShutterPhase`
  - `SetMotionBlurAdaptiveSampleLimit`
  - `SetMotionBlurSamplesPerFrame`
- Validation rejects non-integer values, `shutter_angle` outside `0..720`, and
  negative sample settings.
- Boundary: this slice models profile-visible shutter/sample settings only; the
  comp motion blur enable flag is not exposed in the current stable profile
  schema and is not modeled here.
- `examples/recipes/minimal-comp-motion-blur.json` is a dedicated no-layer
  comp recipe that asserts all four `expected_profile.motion_blur` fields. See
  `knowledge/composition/recipe-motion-blur-settings.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-comp-motion-blur.json -out
    tmp_debug\recipes\minimal-comp-motion-blur.aep -json` returned valid and
    all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_comp_motion_blur\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_comp_motion_blur\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Thirty-first follow-up completed:

- Composition recipes now support `work_area` second-based start/end settings:
  - `start` -> profile `work_area.start_seconds`
  - `end` -> profile `work_area.end_seconds`
- Capability reporting records:
  - `SetWorkArea`
- Validation requires both values and rejects ranges outside
  `0 <= start <= end <= comp.duration`.
- Boundary: this slice models second-based work area authoring only. Frame-based
  work area setters remain lower-level `aep` APIs for now.
- `examples/recipes/minimal-comp-work-area.json` is a dedicated no-layer comp
  recipe that asserts both `expected_profile.work_area` fields. See
  `knowledge/composition/recipe-work-area.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-comp-work-area.json -out
    tmp_debug\recipes\minimal-comp-work-area.aep -json` returned valid and all
    `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_comp_work_area\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_comp_work_area\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Thirty-second follow-up completed:

- Composition recipes now support `renderer`:
  - `renderer` -> profile `renderer`
- Capability reporting records:
  - `SetRenderer`
- Boundary: recipe examples and contracts use binary renderer match names (for
  example `ADBE Escher`) because `SetRenderer` may normalize known
  ExtendScript module aliases.
- `examples/recipes/minimal-comp-renderer.json` is a dedicated no-layer comp
  recipe that asserts `expected_profile.renderer`. See
  `knowledge/composition/recipe-renderer.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-comp-renderer.json -out
    tmp_debug\recipes\minimal-comp-renderer.aep -json` returned valid and all
    `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_comp_renderer\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_comp_renderer\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Thirty-third follow-up completed:

- Composition recipes now support `resolution_factor`:
  - `[x, y]` -> `SetResolutionFactor`
- Capability reporting records:
  - `SetResolutionFactor`
- Validation rejects non-two-value arrays, zero values, non-integers, and values
  outside `uint16` range.
- Boundary: the current stable profile schema does not expose resolution
  factor, so this slice uses compiled AEP readback in tests plus AE render
  acceptance for the dedicated example.
- `examples/recipes/minimal-comp-resolution-factor.json` is a dedicated
  no-layer comp recipe with `resolution_factor: [2, 2]`. See
  `knowledge/composition/recipe-resolution-factor.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-comp-resolution-factor.json -out
    tmp_debug\recipes\minimal-comp-resolution-factor.aep -json` returned valid
    and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_comp_resolution_factor\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_comp_resolution_factor\aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Thirty-fourth follow-up completed:

- Composition recipes now support `pixel_aspect`:
  - value -> `SetPixelAspect`
- Capability reporting records:
  - `SetPixelAspect`
- Validation rejects non-positive values.
- Boundary: the current stable profile schema does not expose pixel aspect, so
  this slice uses compiled AEP readback in tests plus AE render acceptance for
  the dedicated example.
- `examples/recipes/minimal-comp-pixel-aspect.json` is a dedicated no-layer
  comp recipe with `pixel_aspect: 2`. See
  `knowledge/composition/recipe-pixel-aspect.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-comp-pixel-aspect.json -out
    tmp_debug\recipes\minimal-comp-pixel-aspect.aep -json` returned valid and
    all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_comp_pixel_aspect\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_comp_pixel_aspect\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Thirty-fifth follow-up completed:

- Composition recipes now support `display_start_time`:
  - seconds -> `SetDisplayStartTime`
- Capability reporting records:
  - `SetDisplayStartTime`
- Validation rejects negative values.
- Boundary: this slice models second-based display start only; frame-based
  display start remains a lower-level `aep` API for now. The current stable
  profile schema does not expose display start time, so this slice uses
  compiled AEP readback in tests plus AE render acceptance for the dedicated
  example.
- `examples/recipes/minimal-comp-display-start-time.json` is a dedicated
  no-layer comp recipe with `display_start_time: 0.5`. See
  `knowledge/composition/recipe-display-start-time.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-comp-display-start-time.json -out
    tmp_debug\recipes\minimal-comp-display-start-time.aep -json` returned valid
    and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_comp_display_start_time\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_comp_display_start_time\aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Thirty-sixth follow-up completed:

- Composition recipes now support `frame_blending`:
  - boolean -> `SetFrameBlending`
- Capability reporting records:
  - `SetFrameBlending`
- Boundary: this is the composition master frame-blending switch only. Visible
  frame interpolation still requires suitable layers with layer frame-blend
  flags/modes enabled. The current stable profile schema does not expose this
  comp flag, so this slice uses compiled AEP `cdta @0x8B bit 0x10` readback in
  tests plus AE render acceptance for the dedicated example.
- `examples/recipes/minimal-comp-frame-blending.json` is a dedicated no-layer
  comp recipe with `frame_blending: true`. See
  `knowledge/composition/recipe-frame-blending.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-comp-frame-blending.json -out
    tmp_debug\recipes\minimal-comp-frame-blending.aep -json` returned valid and
    all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_comp_frame_blending\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_comp_frame_blending\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Thirty-seventh follow-up completed:

- Composition recipes now support `hide_shy_layers`:
  - boolean -> `SetHideShyLayers`
- Capability reporting records:
  - `SetHideShyLayers`
- Boundary: this is the composition master Hide Shy Layers timeline toggle
  only. It only affects layers that are themselves marked shy; recipe
  layer-level shy flags are not modeled in this slice. The current stable
  profile schema does not expose this comp flag, so this slice uses compiled
  AEP `cdta @0x8B bit 0x01` readback in tests plus AE render acceptance for the
  dedicated example.
- `examples/recipes/minimal-comp-hide-shy-layers.json` is a dedicated no-layer
  comp recipe with `hide_shy_layers: true`. See
  `knowledge/composition/recipe-hide-shy-layers.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-comp-hide-shy-layers.json -out
    tmp_debug\recipes\minimal-comp-hide-shy-layers.aep -json` returned valid
    and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_comp_hide_shy_layers\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_comp_hide_shy_layers\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Thirty-eighth follow-up completed:

- Composition recipes now support `preserve_nested_frame_rate`:
  - boolean -> `SetPreserveNestedFrameRate`
- Capability reporting records:
  - `SetPreserveNestedFrameRate`
- Boundary: this is the composition "preserve frame rate when nested" toggle.
  It affects behavior when this comp is used as a nested/precomp source. The
  current stable profile schema does not expose this comp flag, so this slice
  uses compiled AEP `cdta @0x8B bit 0x20` readback in tests plus AE render
  acceptance for the dedicated example.
- `examples/recipes/minimal-comp-preserve-nested-frame-rate.json` is a
  dedicated no-layer comp recipe with `preserve_nested_frame_rate: true`. See
  `knowledge/composition/recipe-preserve-nested-frame-rate.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-comp-preserve-nested-frame-rate.json -out
    tmp_debug\recipes\minimal-comp-preserve-nested-frame-rate.aep -json`
    returned valid and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_comp_preserve_nested_frame_rate\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_comp_preserve_nested_frame_rate\aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Thirty-ninth follow-up completed:

- Composition recipes now support `preserve_nested_resolution`:
  - boolean -> `SetPreserveNestedResolution`
- Capability reporting records:
  - `SetPreserveNestedResolution`
- Boundary: this is the composition "preserve resolution when nested" toggle.
  It affects behavior when this comp is used as a nested/precomp source. The
  current stable profile schema does not expose this comp flag, so this slice
  uses compiled AEP `cdta @0x8B bit 0x80` readback in tests plus AE render
  acceptance for the dedicated example.
- `examples/recipes/minimal-comp-preserve-nested-resolution.json` is a
  dedicated no-layer comp recipe with `preserve_nested_resolution: true`. See
  `knowledge/composition/recipe-preserve-nested-resolution.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-comp-preserve-nested-resolution.json -out
    tmp_debug\recipes\minimal-comp-preserve-nested-resolution.aep -json`
    returned valid and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_comp_preserve_nested_resolution\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_comp_preserve_nested_resolution\aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Fortieth follow-up completed:

- Composition recipes now support `motion_blur.enabled`:
  - boolean -> `SetCompMotionBlur`
- Capability reporting records:
  - `SetCompMotionBlur`
- Boundary: this is the composition motion-blur master switch only. Layers
  still need their own per-layer motion-blur flag to render motion blur. The
  current stable profile schema does not expose this comp flag, so this slice
  uses compiled AEP `cdta @0x8B bit 0x08` readback in tests plus AE render
  acceptance for the dedicated example.
- `examples/recipes/minimal-comp-motion-blur-enabled.json` is a dedicated
  no-layer comp recipe with `motion_blur.enabled: true`. See
  `knowledge/composition/recipe-motion-blur-enabled.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-comp-motion-blur-enabled.json -out
    tmp_debug\recipes\minimal-comp-motion-blur-enabled.aep -json` returned
    valid and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_comp_motion_blur_enabled\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_comp_motion_blur_enabled\aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Forty-first follow-up completed:

- Composition recipes now support `label`:
  - integer `0..16` -> `SetLabel`
- Capability reporting records:
  - `SetLabel`
- Boundary: composition labels are project-panel item-level fields, not cdta
  comp settings. Recipe compilation applies them after building the base
  project by reopening the project and writing the reopened composition's item
  carrier. The current stable profile schema does not expose labels, so this
  slice uses compiled AEP readback in tests plus AE render/open acceptance for
  the dedicated example.
- `examples/recipes/minimal-comp-label.json` is a dedicated no-layer comp
  recipe with `label: 12`. See `knowledge/composition/recipe-label.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-comp-label.json -out
    tmp_debug\recipes\minimal-comp-label.aep -json` returned valid and all
    `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_comp_label\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_comp_label\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Forty-second follow-up completed:

- Composition recipes now support `comment`:
  - non-empty string -> `SetComment`
- Capability reporting records:
  - `SetComment`
- Boundary: composition comments are project-panel item-level fields, not cdta
  comp settings. The comment setter is length-variable and depends on the item
  `cmta` payload plus the idta has-comment flag, so recipe compilation applies
  it after reopening the base project. The current stable profile schema does
  not expose comments, so this slice uses compiled AEP readback in tests plus
  AE render/open acceptance for the dedicated example.
- `examples/recipes/minimal-comp-comment.json` is a dedicated no-layer comp
  recipe with `comment: "reviewed recipe comp"`. See
  `knowledge/composition/recipe-comment.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-comp-comment.json -out
    tmp_debug\recipes\minimal-comp-comment.aep -json` returned valid and all
    `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_comp_comment\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_comp_comment\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Forty-third follow-up completed:

- Layer recipes now support `label`:
  - integer `0..16` -> `Layer.SetLabel`
- Capability reporting records:
  - `Layer.SetLabel`
- Boundary: layer labels are `ldta` item metadata at offset `0x3D`. The
  current stable profile schema does not expose layer labels, so this slice
  uses schema capability reporting, compiled AEP readback in tests, and AE
  render/open acceptance for the dedicated example.
- `examples/recipes/minimal-layer-label.json` is a dedicated one-text-layer
  recipe with `label: 10`. See `knowledge/layer/recipe-label.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-layer-label.json -out
    tmp_debug\recipes\minimal-layer-label.aep -json` returned valid and all
    `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_layer_label\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_layer_label\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Forty-fourth follow-up completed:

- Layer recipes now support `comment`:
  - non-empty string -> `Layer.SetComment`
- Capability reporting records:
  - `Layer.SetComment`
- Boundary: layer comments are length-variable `cmta` metadata. The writer
  must keep the `cmta` double-NUL terminator and `ldta` offset `0x3C`
  has-comment flag in sync. The current stable profile schema does not expose
  layer comments, so this slice uses schema capability reporting, compiled AEP
  readback in tests, and AE render/open acceptance for the dedicated example.
- `examples/recipes/minimal-layer-comment.json` is a dedicated one-text-layer
  recipe with `comment: "reviewed recipe layer"`. See
  `knowledge/layer/recipe-comment.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-layer-comment.json -out
    tmp_debug\recipes\minimal-layer-comment.aep -json` returned valid and all
    `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_layer_comment\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_layer_comment\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Forty-fifth follow-up completed:

- Layer recipes now support `motion_blur`:
  - boolean -> `Layer.SetMotionBlur`
- Capability reporting records:
  - `Layer.SetMotionBlur`
- Boundary: layer motion blur is a length-preserving bit in `ldta` offset
  `0x27`. Visible blur still depends on the composition motion blur switch and
  settings; the dedicated example enables the comp switch too.
- `examples/recipes/minimal-layer-motion-blur.json` is a dedicated
  one-text-layer recipe with `motion_blur: true`. See
  `knowledge/layer/recipe-motion-blur.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-layer-motion-blur.json -out
    tmp_debug\recipes\minimal-layer-motion-blur.aep -json` returned valid and
    all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_layer_motion_blur\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_layer_motion_blur\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Forty-sixth follow-up completed:

- Layer recipes now support `shy`:
  - boolean -> `Layer.SetShy`
- Capability reporting records:
  - `Layer.SetShy`
- Boundary: layer shy is a length-preserving bit in `ldta` offset `0x27`.
  Hiding shy layers still depends on the composition `hide_shy_layers` switch;
  this slice covers the layer flag itself.
- `examples/recipes/minimal-layer-shy.json` is a dedicated one-text-layer
  recipe with `shy: true`. See `knowledge/layer/recipe-shy.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-layer-shy.json -out
    tmp_debug\recipes\minimal-layer-shy.aep -json` returned valid and all
    `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_layer_shy\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_layer_shy\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Forty-seventh follow-up completed:

- Layer recipes now support common switches:
  - `visible` -> `Layer.SetVisible`
  - `solo` -> `Layer.SetSolo`
  - `locked` -> `Layer.SetLocked`
  - `effects_enabled` -> `Layer.SetEffectsEnabled`
  - `audio_enabled` -> `Layer.SetAudioEnabled`
  - `frame_blend_enabled` -> `Layer.SetFrameBlendEnabled`
- Boundary: these are length-preserving layer switches in `ldta`; solo uses
  offset `0x26`, the others are in the common flag region around offset
  `0x27`. Some switches only affect rendering with matching layer context
  (`effects_enabled`, `audio_enabled`, `frame_blend_enabled`), so this slice
  asserts byte/readback behavior and AE open/render acceptance.
- `examples/recipes/minimal-layer-common-switches.json` is a dedicated
  one-text-layer recipe covering all six switches. See
  `knowledge/layer/recipe-common-switches.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-layer-common-switches.json -out
    tmp_debug\recipes\minimal-layer-common-switches.aep -json` returned valid
    and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_layer_common_switches\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_layer_common_switches\aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Forty-eighth follow-up completed:

- Layer recipes now support advanced switches:
  - `collapse_transform` -> `Layer.SetCollapseTransform`
  - `is_3d` -> `Layer.SetIs3D`
  - `is_adjust` -> `Layer.SetIsAdjust`
  - `is_guide` -> `Layer.SetIsGuide`
  - `sampling_bicubic` -> `Layer.SetSamplingBicubic`
  - `frame_blend_pixel_motion` -> `Layer.SetFrameBlendPixelMotion`
  - `preserve_transparency` -> `Layer.SetPreserveTransparency`
- Boundary: these are length-preserving layer switches in `ldta`.
  `frame_blend_pixel_motion` requires frame blending to have a visible effect,
  so the dedicated recipe enables `frame_blend_enabled` too. `is_guide`
  excludes the layer from rendered output, so the oracle PNGs are expected to
  be black while still proving AE open/render acceptance. `markers_locked` is
  intentionally excluded because its current capability verification is
  `roundtrip`, not `ae-accept`.
- `examples/recipes/minimal-layer-advanced-switches.json` is a dedicated
  one-text-layer recipe covering all seven switches. See
  `knowledge/layer/recipe-advanced-switches.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-layer-advanced-switches.json -out
    tmp_debug\recipes\minimal-layer-advanced-switches.aep -json` returned
    valid and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_layer_advanced_switches\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_layer_advanced_switches\aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Forty-ninth follow-up completed:

- Layer recipes now support quality and blending mode:
  - `quality` (`wireframe`, `draft`, `best`) -> `Layer.SetQuality`
  - `blending_mode` (lowercase snake AE blend name) ->
    `Layer.SetBlendingMode`
- Boundary: layer quality is a length-preserving uint16 at `ldta` offset
  `0x04`; blending mode is a length-preserving byte at `ldta` offset `0x63`.
  Both are stable and AE accept-gated. Invalid strings are refused during
  recipe validation.
- `examples/recipes/minimal-layer-quality-blending.json` is a dedicated
  one-text-layer recipe with `quality: "draft"` and
  `blending_mode: "multiply"`. See
  `knowledge/layer/recipe-quality-blending.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-layer-quality-blending.json -out
    tmp_debug\recipes\minimal-layer-quality-blending.aep -json` returned
    valid and all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_layer_quality_blending\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_layer_quality_blending\aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Fiftieth follow-up completed:

- Layer recipes now support timing controls:
  - `start_time` -> `Layer.SetStartTime`
  - `in_point` -> `Layer.SetInPoint`
  - `out_point` -> `Layer.SetOutPoint`
- Boundary: `start_time` is allowed to be negative for pre-roll, while
  `in_point` and `out_point` are validated within comp duration and
  `out_point >= in_point`. `stretch` is intentionally excluded because its
  current writer capability is alpha/roundtrip-only and AE recomputes the span
  from coordinated timing fields.
- `examples/recipes/minimal-layer-timing.json` is a dedicated one-text-layer
  recipe with `start_time: 0.25`, `in_point: 0.1`, and `out_point: 0.9`. See
  `knowledge/layer/recipe-timing.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-layer-timing.json -out
    tmp_debug\recipes\minimal-layer-timing.aep -json` returned valid and all
    `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_layer_timing\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_layer_timing\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000003.png`, `f000027.png`

Fifty-first follow-up completed:

- Layer recipes now support same-comp parenting:
  - `parent` (layer name) -> `Layer.SetParent`
- Boundary: recipe input uses a layer name instead of a raw AE layer ID because
  layer IDs are generated during compilation. Parent links are applied after
  all layers in the comp are created, so the referenced parent may appear
  before or after the child in the recipe. Unknown parent names are refused
  during validation.
- `examples/recipes/minimal-layer-parent.json` is a dedicated two-text-layer
  recipe with `Child.parent: "Parent"`. See
  `knowledge/layer/recipe-parent.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-layer-parent.json -out
    tmp_debug\recipes\minimal-layer-parent.aep -json` returned valid and all
    `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_layer_parent\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_layer_parent\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Fifty-second follow-up completed:

- Layer recipes now support auto-orient mode:
  - `auto_orient` -> `Layer.SetAutoOrient`
- Supported recipe values are `none`, `along_path`,
  `camera_or_point_of_interest`, and `characters_toward_camera`. Invalid
  strings are refused during validation.
- Boundary: auto-orient is a mutually-exclusive bit group spread across `ldta`
  offsets `0x25` and `0x26`. `along_path` only has visible meaning when the
  layer has a motion path, so the dedicated recipe includes Position
  keyframes.
- `examples/recipes/minimal-layer-auto-orient.json` is a dedicated text-layer
  recipe with `auto_orient: "along_path"`. See
  `knowledge/layer/recipe-auto-orient.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-layer-auto-orient.json -out
    tmp_debug\recipes\minimal-layer-auto-orient.aep -json` returned valid and
    all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_layer_auto_orient\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_layer_auto_orient\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Fifty-third follow-up completed:

- Layer recipes now support the null-object marker flag:
  - `is_null` -> `Layer.SetIsNull`
- Boundary: this exposes AE's low-level Null-Object marker bit on an existing
  recipe layer. It is not the same as AE's UI command or the library's
  higher-level `NewNullLayer`, which creates a standard 100x100 solid-backed
  null helper.
- `examples/recipes/minimal-layer-null-flag.json` is a dedicated two-layer
  recipe: a hidden solid `Controller` with `is_null: true`, plus a text child
  parented to it. See `knowledge/layer/recipe-null-flag.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-layer-null-flag.json -out
    tmp_debug\recipes\minimal-layer-null-flag.aep -json` returned valid and
    all `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_layer_null_flag\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_layer_null_flag\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Fifty-fourth follow-up completed:

- Layer recipes now support a high-level null layer type:
  - `type: "null"` -> `NewNullLayer`
- Boundary: `type: "null"` creates the standard solid-backed null helper via
  the library creation API. This differs from `is_null`, which only flips the
  low-level Null-Object marker bit on a layer created by another recipe type.
- `examples/recipes/minimal-null-layer.json` is a dedicated two-layer recipe:
  a `Controller` null layer plus a text child parented to it. See
  `knowledge/layer/recipe-null-layer.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-null-layer.json -out
    tmp_debug\recipes\minimal-null-layer.aep -json` returned valid and all
    `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_null_layer\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_null_layer\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Fifty-fifth follow-up completed:

- Layer recipes now support a high-level adjustment layer type:
  - `type: "adjustment"` -> `NewAdjustmentLayer`
- Boundary: `type: "adjustment"` creates a comp-sized adjustment layer via the
  library creation API. This differs from `is_adjust`, which only flips the
  low-level adjustment marker bit on a layer created by another recipe type.
- `examples/recipes/minimal-adjustment-layer.json` is a dedicated two-layer
  recipe: a `Grade` adjustment layer with Gaussian Blur plus a text layer. See
  `knowledge/layer/recipe-adjustment-layer.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-adjustment-layer.json -out
    tmp_debug\recipes\minimal-adjustment-layer.aep -json` returned valid and
    all `profile_checks` passed, including effect and effect-param checks.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_adjustment_layer\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_adjustment_layer\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Fifty-sixth follow-up completed:

- Layer recipes now support a high-level camera layer type:
  - `type: "camera"` -> `NewCameraLayer`
- Boundary: this slice covers camera creation only. Recipe transform validation
  still requires 2-value vectors, so 3D camera placement and camera option
  setters remain deferred to separate recipe slices.
- `examples/recipes/minimal-camera-layer.json` is a dedicated two-layer recipe:
  a `Camera` layer plus a visible text layer. See
  `knowledge/layer/recipe-camera-layer.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-camera-layer.json -out
    tmp_debug\recipes\minimal-camera-layer.aep -json` returned valid and all
    `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_camera_layer\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_camera_layer\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Fifty-seventh follow-up completed:

- Layer recipes now support a high-level light layer type:
  - `type: "light"` -> `NewLightLayer`
- Boundary: this slice covers light creation only. Recipe transform validation
  still requires 2-value vectors, so 3D light placement and light option
  setters remain deferred to separate recipe slices.
- `examples/recipes/minimal-light-layer.json` is a dedicated two-layer recipe:
  a `Light` layer plus a visible text layer. See
  `knowledge/layer/recipe-light-layer.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-light-layer.json -out
    tmp_debug\recipes\minimal-light-layer.aep -json` returned valid and all
    `profile_checks` passed.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_light_layer\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_light_layer\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Fifty-eighth follow-up completed:

- Camera layer recipes now support the first camera option:
  - `camera.zoom` -> `Layer.SetCameraZoom`
- Boundary: `camera` options are valid only on `type: "camera"` layers. This
  slice covers zoom only; depth of field, focus distance, aperture, iris, and
  other camera options remain separate recipe slices.
- `examples/recipes/minimal-camera-zoom.json` is a dedicated two-layer recipe:
  a `Camera` layer with zoom `850` plus a visible text layer. See
  `knowledge/layer/recipe-camera-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-camera-zoom.json -out
    tmp_debug\recipes\minimal-camera-zoom.aep -json` returned valid and all
    `profile_checks` passed, including `ADBE Camera Zoom = 850`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_camera_zoom\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_camera_zoom\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Fifty-ninth follow-up completed:

- Camera layer recipes now support a second camera option:
  - `camera.depth_of_field` -> `Layer.SetCameraDepthOfField`
- Boundary: `camera` options remain valid only on `type: "camera"` layers.
  This slice covers the Depth of Field toggle only; focus distance, aperture,
  iris, and other camera options remain separate recipe slices.
- `examples/recipes/minimal-camera-depth-of-field.json` is a dedicated
  two-layer recipe: a `Camera` layer with depth of field disabled plus a visible
  text layer. See `knowledge/layer/recipe-camera-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-camera-depth-of-field.json -out
    tmp_debug\recipes\minimal-camera-depth-of-field.aep -json` returned valid
    and all `profile_checks` passed, including
    `ADBE Camera Depth of Field = 0`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_camera_depth_of_field\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_camera_depth_of_field\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Sixtieth follow-up completed:

- Camera layer recipes now support a third camera option:
  - `camera.focus_distance` -> `Layer.SetCameraFocusDistance`
- Boundary: `camera` options remain valid only on `type: "camera"` layers.
  This slice covers Focus Distance only; aperture, iris, and other camera
  options remain separate recipe slices.
- `examples/recipes/minimal-camera-focus-distance.json` is a dedicated
  two-layer recipe: a `Camera` layer with focus distance `1200` plus a visible
  text layer. See `knowledge/layer/recipe-camera-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-camera-focus-distance.json -out
    tmp_debug\recipes\minimal-camera-focus-distance.aep -json` returned valid
    and all `profile_checks` passed, including
    `ADBE Camera Focus Distance = 1200`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_camera_focus_distance\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_camera_focus_distance\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Sixty-first follow-up completed:

- Camera layer recipes now support a fourth camera option:
  - `camera.aperture` -> `Layer.SetCameraAperture`
- Boundary: `camera` options remain valid only on `type: "camera"` layers.
  This slice covers Aperture only; iris and other camera options remain separate
  recipe slices.
- `examples/recipes/minimal-camera-aperture.json` is a dedicated two-layer
  recipe: a `Camera` layer with aperture `180` plus a visible text layer. See
  `knowledge/layer/recipe-camera-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-camera-aperture.json -out
    tmp_debug\recipes\minimal-camera-aperture.aep -json` returned valid and all
    `profile_checks` passed, including `ADBE Camera Aperture = 180`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_camera_aperture\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_camera_aperture\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Sixty-second follow-up completed:

- Camera layer recipes now support a fifth camera option:
  - `camera.blur_level` -> `Layer.SetCameraBlurLevel`
- Boundary: `camera` options remain valid only on `type: "camera"` layers.
  This slice covers Blur Level only; iris and other camera options remain
  separate recipe slices.
- `examples/recipes/minimal-camera-blur-level.json` is a dedicated two-layer
  recipe: a `Camera` layer with blur level `120` plus a visible text layer. See
  `knowledge/layer/recipe-camera-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-camera-blur-level.json -out
    tmp_debug\recipes\minimal-camera-blur-level.aep -json` returned valid and
    all `profile_checks` passed, including
    `ADBE Camera Blur Level = 120`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_camera_blur_level\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_camera_blur_level\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Sixty-third follow-up completed:

- Camera layer recipes now support a sixth camera option:
  - `camera.iris_shape` -> `Layer.SetIrisShape`
- Boundary: `camera` options remain valid only on `type: "camera"` layers.
  This slice covers Iris Shape only; other iris and camera options remain
  separate recipe slices.
- `examples/recipes/minimal-camera-iris-shape.json` is a dedicated two-layer
  recipe: a `Camera` layer with iris shape `4` plus a visible text layer. See
  `knowledge/layer/recipe-camera-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-camera-iris-shape.json -out
    tmp_debug\recipes\minimal-camera-iris-shape.aep -json` returned valid and
    all `profile_checks` passed, including `ADBE Iris Shape = 4`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_camera_iris_shape\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_camera_iris_shape\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Sixty-fourth follow-up completed:

- Camera layer recipes now support a seventh camera option:
  - `camera.iris_rotation` -> `Layer.SetIrisRotation`
- Boundary: `camera` options remain valid only on `type: "camera"` layers.
  This slice covers Iris Rotation only; other iris and camera options remain
  separate recipe slices.
- `examples/recipes/minimal-camera-iris-rotation.json` is a dedicated
  two-layer recipe: a `Camera` layer with iris rotation `25` plus a visible
  text layer. See `knowledge/layer/recipe-camera-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-camera-iris-rotation.json -out
    tmp_debug\recipes\minimal-camera-iris-rotation.aep -json` returned valid
    and all `profile_checks` passed, including `ADBE Iris Rotation = 25`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_camera_iris_rotation\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_camera_iris_rotation\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Sixty-fifth follow-up completed:

- Camera layer recipes now support an eighth camera option:
  - `camera.iris_roundness` -> `Layer.SetIrisRoundness`
- Boundary: `camera` options remain valid only on `type: "camera"` layers.
  This slice covers Iris Roundness only; other iris and camera options remain
  separate recipe slices.
- `examples/recipes/minimal-camera-iris-roundness.json` is a dedicated
  two-layer recipe: a `Camera` layer with iris roundness `60` plus a visible
  text layer. See `knowledge/layer/recipe-camera-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-camera-iris-roundness.json -out
    tmp_debug\recipes\minimal-camera-iris-roundness.aep -json` returned valid
    and all `profile_checks` passed, including `ADBE Iris Roundness = 60`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_camera_iris_roundness\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_camera_iris_roundness\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Sixty-sixth follow-up completed:

- Camera layer recipes now support a ninth camera option:
  - `camera.iris_aspect_ratio` -> `Layer.SetIrisAspectRatio`
- Boundary: `camera` options remain valid only on `type: "camera"` layers.
  This slice covers Iris Aspect Ratio only; other iris and camera options
  remain separate recipe slices.
- `examples/recipes/minimal-camera-iris-aspect-ratio.json` is a dedicated
  two-layer recipe: a `Camera` layer with iris aspect ratio `1.8` plus a
  visible text layer. See `knowledge/layer/recipe-camera-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-camera-iris-aspect-ratio.json -out
    tmp_debug\recipes\minimal-camera-iris-aspect-ratio.aep -json` returned
    valid and all `profile_checks` passed, including
    `ADBE Iris Aspect Ratio = 1.8`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_camera_iris_aspect_ratio\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_camera_iris_aspect_ratio\aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Sixty-seventh follow-up completed:

- Camera layer recipes now support a tenth camera option:
  - `camera.iris_diffraction_fringe` -> `Layer.SetIrisDiffractionFringe`
- Boundary: `camera` options remain valid only on `type: "camera"` layers.
  This slice covers Iris Diffraction Fringe only; other iris and camera
  options remain separate recipe slices.
- `examples/recipes/minimal-camera-iris-diffraction-fringe.json` is a
  dedicated two-layer recipe: a `Camera` layer with iris diffraction fringe
  `30` plus a visible text layer. See
  `knowledge/layer/recipe-camera-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-camera-iris-diffraction-fringe.json -out
    tmp_debug\recipes\minimal-camera-iris-diffraction-fringe.aep -json`
    returned valid and all `profile_checks` passed, including
    `ADBE Iris Diffraction Fringe = 30`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_camera_iris_diffraction_fringe\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_camera_iris_diffraction_fringe\aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Sixty-eighth follow-up completed:

- Camera layer recipes now support an eleventh camera option:
  - `camera.iris_highlight_gain` -> `Layer.SetIrisHighlightGain`
- Boundary: `camera` options remain valid only on `type: "camera"` layers.
  This slice covers Iris Highlight Gain only; other iris highlight and camera
  options remain separate recipe slices.
- `examples/recipes/minimal-camera-iris-highlight-gain.json` is a dedicated
  two-layer recipe: a `Camera` layer with iris highlight gain `40` plus a
  visible text layer. See `knowledge/layer/recipe-camera-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-camera-iris-highlight-gain.json -out
    tmp_debug\recipes\minimal-camera-iris-highlight-gain.aep -json` returned
    valid and all `profile_checks` passed, including
    `ADBE Iris Highlight Gain = 40`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_camera_iris_highlight_gain\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_camera_iris_highlight_gain\aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Sixty-ninth follow-up completed:

- Camera layer recipes now support a twelfth camera option:
  - `camera.iris_highlight_threshold` -> `Layer.SetIrisHighlightThreshold`
- Boundary: `camera` options remain valid only on `type: "camera"` layers.
  This slice covers Iris Highlight Threshold only; highlight saturation and
  other camera options remain separate recipe slices.
- `examples/recipes/minimal-camera-iris-highlight-threshold.json` is a
  dedicated two-layer recipe: a `Camera` layer with iris highlight threshold
  `0.7` plus a visible text layer. See
  `knowledge/layer/recipe-camera-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-camera-iris-highlight-threshold.json -out
    tmp_debug\recipes\minimal-camera-iris-highlight-threshold.aep -json`
    returned valid and all `profile_checks` passed, including
    `ADBE Iris Highlight Threshold = 0.7`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_camera_iris_highlight_threshold\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_camera_iris_highlight_threshold\aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Seventieth follow-up completed:

- Camera layer recipes now support a thirteenth camera option:
  - `camera.iris_highlight_saturation` -> `Layer.SetIrisHighlightSaturation`
- Boundary: `camera` options remain valid only on `type: "camera"` layers.
  This slice covers Iris Highlight Saturation. The profile match name keeps
  Adobe's on-disk misspelling: `ADBE Iris Hightlight Saturation`.
- `examples/recipes/minimal-camera-iris-highlight-saturation.json` is a
  dedicated two-layer recipe: a `Camera` layer with iris highlight saturation
  `50` plus a visible text layer. See
  `knowledge/layer/recipe-camera-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-camera-iris-highlight-saturation.json -out
    tmp_debug\recipes\minimal-camera-iris-highlight-saturation.aep -json`
    returned valid and all `profile_checks` passed, including
    `ADBE Iris Hightlight Saturation = 50`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_camera_iris_highlight_saturation\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_camera_iris_highlight_saturation\aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Seventy-first follow-up completed:

- Light layer recipes now support the first light option:
  - `light.intensity` -> `Layer.SetLightIntensity`
- Boundary: `light` options are valid only on `type: "light"` layers. This
  slice covers Intensity only; other light options remain separate recipe
  slices.
- `examples/recipes/minimal-light-intensity.json` is a dedicated two-layer
  recipe: a `Light` layer with intensity `140` plus a visible text layer. See
  `knowledge/layer/recipe-light-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-light-intensity.json -out
    tmp_debug\recipes\minimal-light-intensity.aep -json` returned valid and
    all `profile_checks` passed, including `ADBE Light Intensity = 140`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_light_intensity\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_light_intensity\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`

Seventy-second follow-up completed:

- Light layer recipes now support:
  - `light.color` -> `Layer.SetLightColor`
- Boundary: `light` options remain valid only on `type: "light"` layers. This
  slice adds Color; cone, falloff, shadow, and other light options remain
  separate recipe slices.
- `examples/recipes/minimal-light-color.json` is a dedicated two-layer recipe:
  a `Light` layer with color `[255, 51, 102, 204]` plus a visible text layer.
  The light color follows the underlying raw channel order `[A, R, G, B]` in
  the 0..255 range. See `knowledge/layer/recipe-light-options.md`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-light-color.json -out
    tmp_debug\recipes\minimal-light-color.aep -json` returned valid and all
    `profile_checks` passed, including `ADBE Light Color = [255, 51, 102, 204]`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_light_color\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_light_color\aeoracle_render.done` = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`
  - Full gate passed: `go test ./...`, `go vet ./...`,
    `Invoke-Pester -Path scripts\ae_run.Tests.ps1`, and `git diff --check`.

Seventy-third follow-up completed:

- Light layer recipes now support:
  - `light.casts_shadows` -> `Layer.SetLightCastsShadows`
- Boundary: `light` options remain valid only on `type: "light"` layers. This
  slice adds the Casts Shadows toggle; cone, falloff, shadow
  darkness/diffusion, and other light options remain separate recipe slices.
- `examples/recipes/minimal-light-casts-shadows.json` is a dedicated two-layer
  recipe: a `Light` layer with `casts_shadows: true` plus a visible text layer.
  The embedded expected profile checks `ADBE Casts Shadows = 1`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-light-casts-shadows.json -out
    tmp_debug\recipes\minimal-light-casts-shadows.aep -json` returned valid
    and all `profile_checks` passed, including `ADBE Casts Shadows = 1`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_light_casts_shadows\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_light_casts_shadows\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`
  - Full gate passed: `go test ./...`, `go vet ./...`,
    `Invoke-Pester -Path scripts\ae_run.Tests.ps1`, and `git diff --check`.

Seventy-fourth follow-up completed:

- Light layer recipes now support:
  - `light.shadow_darkness` -> `Layer.SetLightShadowDarkness`
- Boundary: `light` options remain valid only on `type: "light"` layers. This
  slice adds Shadow Darkness; cone, falloff, shadow diffusion, and other light
  options remain separate recipe slices.
- `examples/recipes/minimal-light-shadow-darkness.json` is a dedicated
  two-layer recipe: a `Light` layer with `shadow_darkness: 75` plus a visible
  text layer. The embedded expected profile checks
  `ADBE Light Shadow Darkness = 75`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-light-shadow-darkness.json -out
    tmp_debug\recipes\minimal-light-shadow-darkness.aep -json` returned valid
    and all `profile_checks` passed, including
    `ADBE Light Shadow Darkness = 75`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_light_shadow_darkness\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_light_shadow_darkness\aeoracle_render.done` =
      `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`
  - Full gate passed: `go test ./...`, `go vet ./...`,
    `Invoke-Pester -Path scripts\ae_run.Tests.ps1`, and `git diff --check`.

Seventy-fifth follow-up completed:

- Light layer recipes now support:
  - `light.shadow_diffusion` -> `Layer.SetLightShadowDiffusion`
- Boundary: `light` options remain valid only on `type: "light"` layers. This
  slice adds Shadow Diffusion; cone, falloff, and other light options remain
  separate recipe slices.
- `examples/recipes/minimal-light-shadow-diffusion.json` is a dedicated
  two-layer recipe: a `Light` layer with `shadow_diffusion: 18` plus a visible
  text layer. The embedded expected profile checks
  `ADBE Light Shadow Diffusion = 18`.
- Verification:
  - `go test ./internal/recipe ./cmd/aeprecipe` passed.
  - `cmd/aeprecipe compile -recipe
    examples\recipes\minimal-light-shadow-diffusion.json -out
    tmp_debug\recipes\minimal-light-shadow-diffusion.aep -json` returned valid
    and all `profile_checks` passed, including
    `ADBE Light Shadow Diffusion = 18`.
  - AE 2025 render oracle passed:
    - request:
      `tmp_debug\aeoracle\minimal_light_shadow_diffusion\request.json`
    - done marker:
      `tmp_debug\aeoracle\minimal_light_shadow_diffusion\aeoracle_render.done`
      = `ok`
    - metadata: status `ok`, AE `25.1x68`
    - PNG outputs:
      `f000000.png`, `f000015.png`, `f000030.png`
  - Full gate passed: `go test ./...`, `go vet ./...`,
    `Invoke-Pester -Path scripts\ae_run.Tests.ps1`, and `git diff --check`.

Next generation work should broaden recipe coverage in small proven slices:
additional transform keyframe channels/ease where writer support exists,
expression support, or shape filters. Do not start automated correction loops.
