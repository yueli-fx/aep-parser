# ⚠ Recipe 2D Anchor Point keyframes profile as 3D `[x,y,0]`

SUMMARY: Recipe `transform.anchor_point_keyframes` accepts 2D `[x,y]` values, but `expected_profile.keyframes[]` for `ADBE Anchor Point` must compare against profile-read 3D `[x,y,0]` values.
READ WHEN: adding recipe transform anchor point keyframe checks; writing `expected_profile.keyframes[]` for `ADBE Anchor Point`; a recipe keyframe profile check fails with actual `[x y 0]`

---

The recipe schema writes layer Anchor Point keyframes as 2D values:

```json
"anchor_point_keyframes": [
  { "time": 0, "value": [0, 0] },
  { "time": 2, "value": [120, -40] }
]
```

`internal/profile` reads the resulting `ADBE Anchor Point` keyframe values as
3D vectors with a zero Z component:

```json
"expected_profile": {
  "keyframes": [
    {
      "layer_name": "Title",
      "match_name": "ADBE Anchor Point",
      "keyframes": [
        { "time": 0, "value": [0, 0, 0] },
        { "time": 2, "value": [120, -40, 0] }
      ]
    }
  ]
}
```

Do not "fix" the recipe input to 3D unless the writer API changes. The split is
intentional for now: authoring stays 2D for 2D layers, while profile contracts
must match the parser's normalized value shape.
