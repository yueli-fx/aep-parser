# ⚠ Recipe Offset Paths line_join authors as string, profile reads numeric enum

Recipe `shape.offset_paths.line_join` uses `"miter"`, `"round"`, or `"bevel"`, but `expected_profile.properties[]` for `ADBE Vector Offset Line Join` must compare against AE's numeric enum values `1`, `2`, or `3`.

The recipe schema keeps Offset Paths join style readable:

```json
"offset_paths": {
  "amount": 24,
  "line_join": "bevel",
  "miter_limit": 2,
  "copies": 3,
  "copy_offset": 1.5
}
```

The writer maps this to `aep.StrokeLineJoin` before calling
`OffsetPathsNode.SetLineJoin`. `internal/profile` reads the resulting
`ADBE Vector Offset Line Join` leaf as AE's numeric enum:

- `miter` -> `1`
- `round` -> `2`
- `bevel` -> `3`

So a profile contract for the example above must assert:

```json
{
  "layer_name": "Underline",
  "match_name": "ADBE Vector Offset Line Join",
  "value": 3
}
```

Do not change recipe authoring to numeric enums unless the recipe contract is
intentionally revised. Keep the recipe human-readable and make profile
contracts match parser output.
