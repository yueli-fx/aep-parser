# Recipe Stroke Taper Profile Fields

SUMMARY: Recipe support note for Stroke Taper Profile Fields.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for shape fields related to Stroke Taper Profile Fields

---

Context: recipe `shape.stroke.taper` support, proven with
`examples/recipes/minimal-shape-stroke-taper.json`.

Recipe authoring:

- `start_length` -> taper start length percent
- `end_length` -> taper end length percent
- `start_width` -> taper start width percent
- `end_width` -> taper end width percent
- `start_ease` -> taper start ease percent
- `end_ease` -> taper end ease percent

Profile readback:

- `ADBE Vector Taper Start Length`
- `ADBE Vector Taper End Length`
- `ADBE Vector Taper Start Width`
- `ADBE Vector Taper End Width`
- `ADBE Vector Taper Start Ease`
- `ADBE Vector Taper End Ease`

Boundary:

- The writer models the always-active percent-mode taper controls.
- Pixel-mode taper mirror streams and length unit enum are not modeled in
  recipe IR yet.
