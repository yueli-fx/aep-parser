# Recipe Fill Composite Order Profile Enums

SUMMARY: Recipe support note for Fill Composite Order Profile Enums.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for shape fields related to Fill Composite Order Profile Enums

---

Context: recipe `shape.fill_composite_order` support, proven with
`examples/recipes/minimal-shape-fill-composite-order.json`.

Recipe authoring:

- `fill_composite_order` is a string enum:
  - `"above_previous"` -> `ShapeCompositeOrderAbovePrevious`
  - `"below_previous"` -> `ShapeCompositeOrderBelowPrevious`

Profile readback:

- `ADBE Vector Composite Order` is numeric:
  - `1` = above previous
  - `2` = below previous

Validation:

- `fill_composite_order` must be `above_previous` or `below_previous`.
