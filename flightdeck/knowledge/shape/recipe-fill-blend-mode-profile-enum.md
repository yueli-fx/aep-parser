# Recipe Fill Blend Mode Profile Enum

Context: recipe `shape.fill_blend_mode` support, proven with
`examples/recipes/minimal-shape-fill-blend-mode.json`.

Recipe authoring:

- `shape.fill_blend_mode` is AE's 1-based shape blend mode index.
- The value must be an integer of at least `1`.
- `1` is Normal/default. Larger values use AE's shape blend mode enum directly;
  this recipe layer does not introduce string aliases for the full AE enum.

Profile readback:

- `ADBE Vector Blend Mode` reads back as a numeric enum index.
- Fill and stroke both use the same profile match name; dedicated examples avoid
  ambiguous multiple occurrences.

Validation:

- `fill_blend_mode` must be a whole number and at least `1`.
