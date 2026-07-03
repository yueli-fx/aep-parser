# Recipe Fill Rule Profile Enums

SUMMARY: Recipe support note for Fill Rule Profile Enums.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for shape fields related to Fill Rule Profile Enums

---

Context: recipe `shape.fill_rule` support, proven with
`examples/recipes/minimal-shape-fill-rule.json`.

Recipe authoring:

- `fill_rule` is a string enum:
  - `"nonzero_winding"` -> `FillRuleNonzeroWinding`
  - `"even_odd"` -> `FillRuleEvenOdd`

Profile readback:

- `ADBE Vector Fill Rule` is numeric:
  - `1` = nonzero winding
  - `2` = even-odd

Validation:

- `fill_rule` must be `nonzero_winding` or `even_odd`.
