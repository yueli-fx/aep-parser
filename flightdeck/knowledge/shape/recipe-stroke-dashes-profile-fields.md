# Recipe Stroke Dashes Profile Fields

SUMMARY: Recipe support note for Stroke Dashes Profile Fields.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for shape fields related to Stroke Dashes Profile Fields

---

Context: recipe `shape.stroke.dashes` support, proven with
`examples/recipes/minimal-shape-stroke-dashes.json`.

Recipe authoring:

- `shape.stroke.dashes.dash` maps to the first dash length.
- `shape.stroke.dashes.gap` maps to the first gap length.
- Setting either field enables the stroke Dashes group in the writer.

Profile readback:

- `dash` -> `ADBE Vector Stroke Dash 1`
- `gap` -> `ADBE Vector Stroke Gap 1`

Validation:

- `dash` and `gap` must be non-negative.
- The current writer models exactly one Dash/Gap pair.
- Additional dash/gap pairs and dash offset are not modeled in recipe IR yet.
