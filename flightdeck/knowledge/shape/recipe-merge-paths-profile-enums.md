# Recipe Merge Paths Profile Enums

SUMMARY: Recipe support note for Merge Paths Profile Enums.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for shape fields related to Merge Paths Profile Enums

---

Context: recipe `shape.merge_paths` support, proven with
`examples/recipes/minimal-shape-merge-paths.json`.

Recipe authoring:

- `type` is a string enum:
  - `"merge"` -> `MergeTypeMerge`
  - `"add"` -> `MergeTypeAdd`
  - `"subtract"` -> `MergeTypeSubtract`
  - `"intersect"` -> `MergeTypeIntersect`
  - `"exclude"` -> `MergeTypeExclude`

Profile readback:

- `ADBE Vector Merge Type` is numeric:
  - `1` = merge
  - `2` = add
  - `3` = subtract
  - `4` = intersect
  - `5` = exclude

Recipe boundary:

- The current recipe IR has one `shape` primitive per shape layer. This slice
  proves the Merge Paths node and type enum round-trip through profile and AE
  render. A visible multi-path boolean recipe requires a later multi-shape or
  vector-group recipe structure.
