# Recipe Gradient Fill Profile Enums

Context: recipe `shape.gradient_fill` support, proven with
`examples/recipes/minimal-shape-gradient-fill.json`.

Recipe authoring:

- `gradient_fill.type` is a string enum:
  - `"linear"` -> `GradientLinear`
  - `"radial"` -> `GradientRadial`
- `gradient_fill.start_point` and `gradient_fill.end_point` are two-value
  shape-local vectors.
- `gradient_fill.highlight_length` is a radial highlight offset percentage from
  `-100` to `100`; `0` is centered.
- `gradient_fill.highlight_angle` is the radial highlight direction in degrees.
- `gradient_fill.color_stops[]` requires at least two stops. Stop offsets and
  midpoints are unit values from `0` to `1`; colors are recipe RGB values from
  `0` to `255`.
- `gradient_fill.alpha_stops[]` requires at least two stops. Stop offsets and
  midpoints are unit values from `0` to `1`; alpha values are unit opacity from
  `0` to `1`.

Profile readback:

- `ADBE Vector Grad Type` is numeric:
  - `1` = linear
  - `2` = radial
- `ADBE Vector Grad Start Pt` and `ADBE Vector Grad End Pt` read back as
  two-value vectors.
- `ADBE Vector Grad HiLite Length` and `ADBE Vector Grad HiLite Angle` read back
  as numeric scalar properties.
- Alpha stops are stored in the gradient XML payload; recipe coverage asserts
  them through compiled AEP readback (`Gradient().AlphaStops`) plus AE render
  oracle instead of `expected_profile.properties[]`.

Validation:

- `gradient_fill.type` must be `linear` or `radial`.
- `start_point` and `end_point` must contain two numbers when present.
- `highlight_length` must be between `-100` and `100` when present.
- `color_stops` must include at least two valid stops when present.
- `alpha_stops` must include at least two stops when present; offset, midpoint,
  and alpha must be between `0` and `1`.
