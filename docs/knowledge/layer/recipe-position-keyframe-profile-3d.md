# ⚠ Recipe 2D Position keyframes profile as 3D `[x,y,0]`

Recipe `transform.position_keyframes` accepts 2D `[x,y]` values, but `expected_profile.keyframes[]` for `ADBE Position` must compare against profile-read 3D `[x,y,0]` values.

The recipe schema writes layer Position keyframes as 2D values:

```json
"position_keyframes": [
  { "time": 0, "value": [900, 540] },
  { "time": 2, "value": [960, 500] }
]
```

`internal/profile` reads the resulting `ADBE Position` keyframe values as 3D
vectors with a zero Z component:

```json
"expected_profile": {
  "keyframes": [
    {
      "layer_name": "Title",
      "match_name": "ADBE Position",
      "keyframes": [
        { "time": 0, "value": [900, 540, 0] },
        { "time": 2, "value": [960, 500, 0] }
      ]
    }
  ]
}
```

Do not "fix" the recipe input to 3D unless the writer API changes. The split is
intentional for now: authoring stays 2D for 2D layers, while profile contracts
must match the parser's normalized value shape.
