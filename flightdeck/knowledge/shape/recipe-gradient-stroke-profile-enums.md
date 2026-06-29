# Recipe Gradient Stroke Profile Enums

Context: recipe `shape.gradient_stroke` support, proven with
`examples/recipes/minimal-shape-gradient-stroke.json`.

Recipe authoring:

- `gradient_stroke.type` is a string enum:
  - `"linear"` -> `GradientLinear`
  - `"radial"` -> `GradientRadial`
- `gradient_stroke.start_point` and `gradient_stroke.end_point` are two-value
  shape-local vectors.
- `gradient_stroke.highlight_length` is a radial highlight offset percentage
  from `-100` to `100`; `0` is centered.
- `gradient_stroke.highlight_angle` is the radial highlight direction in
  degrees.
- `gradient_stroke.width` is a non-negative stroke width in pixels.
- `gradient_stroke.line_cap` is a string enum:
  - `"butt"` -> `StrokeLineCapButt`
  - `"round"` -> `StrokeLineCapRound`
  - `"projecting"` -> `StrokeLineCapProjecting`
- `gradient_stroke.line_join` is a string enum:
  - `"miter"` -> `StrokeLineJoinMiter`
  - `"round"` -> `StrokeLineJoinRound`
  - `"bevel"` -> `StrokeLineJoinBevel`
- `gradient_stroke.miter_limit` must be at least `1`.
- `gradient_stroke.color_stops[]` requires at least two stops. Stop offsets and
  midpoints are unit values from `0` to `1`; colors are recipe RGB values from
  `0` to `255`.

Profile readback:

- `ADBE Vector Grad Type` is numeric:
  - `1` = linear
  - `2` = radial
- `ADBE Vector Grad Start Pt` and `ADBE Vector Grad End Pt` read back as
  two-value vectors.
- `ADBE Vector Grad HiLite Length` and `ADBE Vector Grad HiLite Angle` read back
  as numeric scalar properties.
- `ADBE Vector Stroke Width` reads back as a numeric scalar.
- `ADBE Vector Stroke Line Cap` and `ADBE Vector Stroke Line Join` read back as
  numeric enum indexes.
- `ADBE Vector Stroke Miter Limit` reads back as a numeric scalar.

Validation:

- `gradient_stroke.type` must be `linear` or `radial`.
- `start_point` and `end_point` must contain two numbers when present.
- `highlight_length` must be between `-100` and `100` when present.
- `width` must be non-negative when present.
- `line_cap` must be `butt`, `round`, or `projecting`.
- `line_join` must be `miter`, `round`, or `bevel`.
- `miter_limit` must be at least `1`.
- `color_stops` must include at least two valid stops when present.
