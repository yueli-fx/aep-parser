# Recipe Gradient Stroke Profile Enums

Context: recipe `shape.gradient_stroke` support, proven with
`examples/recipes/minimal-shape-gradient-stroke.json`.

Recipe authoring:

- `gradient_stroke.type` is a string enum:
  - `"linear"` -> `GradientLinear`
  - `"radial"` -> `GradientRadial`
- `gradient_stroke.start_point` and `gradient_stroke.end_point` are two-value
  shape-local vectors.
- `gradient_stroke.width` is a non-negative stroke width in pixels.
- `gradient_stroke.color_stops[]` requires at least two stops. Stop offsets and
  midpoints are unit values from `0` to `1`; colors are recipe RGB values from
  `0` to `255`.

Profile readback:

- `ADBE Vector Grad Type` is numeric:
  - `1` = linear
  - `2` = radial
- `ADBE Vector Grad Start Pt` and `ADBE Vector Grad End Pt` read back as
  two-value vectors.
- `ADBE Vector Stroke Width` reads back as a numeric scalar.

Validation:

- `gradient_stroke.type` must be `linear` or `radial`.
- `start_point` and `end_point` must contain two numbers when present.
- `width` must be non-negative when present.
- `color_stops` must include at least two valid stops when present.
