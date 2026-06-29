# Recipe Gradient Fill Profile Enums

Context: recipe `shape.gradient_fill` support, proven with
`examples/recipes/minimal-shape-gradient-fill.json`.

Recipe authoring:

- `gradient_fill.type` is a string enum:
  - `"linear"` -> `GradientLinear`
  - `"radial"` -> `GradientRadial`
- `gradient_fill.start_point` and `gradient_fill.end_point` are two-value
  shape-local vectors.
- `gradient_fill.color_stops[]` requires at least two stops. Stop offsets and
  midpoints are unit values from `0` to `1`; colors are recipe RGB values from
  `0` to `255`.

Profile readback:

- `ADBE Vector Grad Type` is numeric:
  - `1` = linear
  - `2` = radial
- `ADBE Vector Grad Start Pt` and `ADBE Vector Grad End Pt` read back as
  two-value vectors.

Validation:

- `gradient_fill.type` must be `linear` or `radial`.
- `start_point` and `end_point` must contain two numbers when present.
- `color_stops` must include at least two valid stops when present.
