# Recipe Stroke Style Profile Enums

Recipe support note for Stroke Style Profile Enums.

Context: recipe `shape.stroke` line style support, proven with
`examples/recipes/minimal-shape-stroke-style.json`.

Recipe authoring:

- `line_cap` is a string enum:
  - `"butt"` -> `StrokeLineCapButt`
  - `"round"` -> `StrokeLineCapRound`
  - `"projecting"` -> `StrokeLineCapProjecting`
- `line_join` is a string enum:
  - `"miter"` -> `StrokeLineJoinMiter`
  - `"round"` -> `StrokeLineJoinRound`
  - `"bevel"` -> `StrokeLineJoinBevel`
- `miter_limit` is a scalar ratio.

Profile readback:

- `ADBE Vector Stroke Line Cap` is numeric:
  - `1` = butt
  - `2` = round
  - `3` = projecting
- `ADBE Vector Stroke Line Join` is numeric:
  - `1` = miter
  - `2` = round
  - `3` = bevel
- `miter_limit` reads back as `ADBE Vector Stroke Miter Limit`.

Validation:

- `line_cap` must be `butt`, `round`, or `projecting`.
- `line_join` must be `miter`, `round`, or `bevel`.
- `miter_limit` must be at least `1`.
