# Recipe Polystar Profile Fields

Context: recipe `shape.kind` support for `star` / `polygon`, proven with
`examples/recipes/minimal-shape-polystar.json`.

Recipe authoring:

- `kind: "star"` emits a default Star node.
- `kind: "polygon"` emits the same Star node with `StarTypePolygon`.
- Shared fields:
  - `points`
  - `position`
  - `rotation`
  - `inner_radius`
  - `outer_radius`
  - `inner_roundness`
  - `outer_roundness`

Profile readback:

- `ADBE Vector Star Type` is numeric:
  - `1` = star
  - `2` = polygon
- Profile property names use AE's on-disk spelling:
  - `ADBE Vector Star Inner Roundess`
  - `ADBE Vector Star Outer Roundess`

Validation:

- `points` must be at least `3`.
- `position` is a 2D vector.
- `inner_radius` and `outer_radius` must be non-negative.
- Legacy `roundness` remains rect-only; polystar uses the explicit inner/outer
  roundness fields.
