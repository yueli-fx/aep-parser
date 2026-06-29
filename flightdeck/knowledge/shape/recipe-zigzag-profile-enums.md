# ⚠ Recipe ZigZag points authors as string, profile reads numeric enum

SUMMARY: Recipe `shape.zigzag.points` uses `"corner"` or `"smooth"`, but `expected_profile.properties[]` for `ADBE Vector Zigzag Points` must compare against AE's numeric enum values `1` or `2`.
READ WHEN: adding or debugging recipe ZigZag support; writing `expected_profile.properties[]` for `ADBE Vector Zigzag Points`; a profile check reports `2` instead of `"smooth"`

---

The recipe schema keeps ZigZag points readable:

```json
"zigzag": {
  "size": 40,
  "detail": 8,
  "points": "smooth"
}
```

The writer maps this to `aep.ZigZagPoints` before calling
`ZigZagNode.SetPoints`. `internal/profile` reads the resulting
`ADBE Vector Zigzag Points` leaf as AE's numeric enum:

- `corner` -> `1`
- `smooth` -> `2`

So a profile contract for the example above must assert:

```json
{
  "layer_name": "Underline",
  "match_name": "ADBE Vector Zigzag Points",
  "value": 2
}
```

Do not change recipe authoring to numeric enums unless the recipe contract is
intentionally revised. Keep the recipe human-readable and make profile
contracts match parser output.
