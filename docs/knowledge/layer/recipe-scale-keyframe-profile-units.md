# ⚠ Recipe scale keyframes author as percent, profile as 3D unit scale

Recipe `transform.scale_keyframes` values are authored as 2D AE percent `[x,y]`, but `expected_profile.keyframes[]` for `ADBE Scale` must compare against profile-read 3D unit values `[x/100,y/100,1]`.

The recipe schema writes layer Scale keyframes as 2D percent values:

```json
"scale_keyframes": [
  { "time": 0, "value": [80, 80] },
  { "time": 2, "value": [100, 120] },
  { "time": 4, "value": [130, 90] }
]
```

`internal/profile` reads the resulting `ADBE Scale` keyframe values as 3D unit
scale vectors:

```json
"expected_profile": {
  "keyframes": [
    {
      "layer_name": "Title",
      "match_name": "ADBE Scale",
      "keyframes": [
        { "time": 0, "value": [0.8, 0.8, 1] },
        { "time": 2, "value": [1, 1.2, 1] },
        { "time": 4, "value": [1.3, 0.9, 1] }
      ]
    }
  ]
}
```

Do not change recipe authoring to unit scale unless the public recipe contract
is intentionally revised. `SetLayerTransform` receives scale in AE percent,
while profile contracts must match parser-normalized 3D unit values.
