# Recipe Stroke Composite Order Profile Enums

Recipe support note for Stroke Composite Order Profile Enums.

Context: recipe `shape.stroke.composite_order` support, proven with
`examples/recipes/minimal-shape-stroke-composite-order.json`.

Recipe authoring:

- `shape.stroke.composite_order` is a string enum:
  - `"above_previous"` -> `ShapeCompositeOrderAbovePrevious`
  - `"below_previous"` -> `ShapeCompositeOrderBelowPrevious`

Profile readback:

- `ADBE Vector Composite Order` is numeric:
  - `1` = above previous
  - `2` = below previous

Validation:

- `shape.stroke.composite_order` must be `above_previous` or
  `below_previous`.
