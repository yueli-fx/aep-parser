# Recipe Repeater Profile Enums

Context: recipe `shape.repeater` support, proven with
`examples/recipes/minimal-shape-repeater.json`.

Recipe authoring:

- `order` is a string enum:
  - `"below"` -> `RepeaterOrderBelow`
  - `"above"` -> `RepeaterOrderAbove`

Profile readback:

- `ADBE Vector Repeater Order` is numeric:
  - `1` = below
  - `2` = above
- Repeater transform opacity profile names are:
  - `ADBE Vector Repeater Opacity 1` for `start_opacity`
  - `ADBE Vector Repeater Opacity 2` for `end_opacity`

Validation:

- `copies` must be at least `1`.
- `order` must be `below` or `above`.
- `anchor`, `position`, and `scale` are 2D vectors.
- `start_opacity` and `end_opacity` must be between `0` and `100`.
