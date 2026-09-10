# Recipe Wiggle Paths points authors as string, profile reads numeric enum

Recipe `shape.wiggle_paths.points` uses `"corner"` or `"smooth"`, but `expected_profile.properties[]` for `ADBE Vector Roughen Points` must compare against AE's numeric enum values `1` or `2`.

The recipe schema keeps Wiggle Paths points readable:

```json
"wiggle_paths": {
  "size": 60,
  "detail": 30,
  "wiggles_per_second": 4,
  "random_seed": 9,
  "points": "smooth",
  "correlation": 80,
  "temporal_phase": 45,
  "spatial_phase": 20
}
```

The writer maps this to `aep.RoughenPoints` before calling
`WigglePathsNode.SetPoints`. `internal/profile` reads the resulting
`ADBE Vector Roughen Points` leaf as AE's numeric enum:

- `corner` -> `1`
- `smooth` -> `2`

So a profile contract for the example above must assert:

```json
{
  "layer_name": "Underline",
  "match_name": "ADBE Vector Roughen Points",
  "value": 2
}
```

Do not change recipe authoring to numeric enums unless the recipe contract is
intentionally revised. Keep the recipe human-readable and make profile
contracts match parser output.
