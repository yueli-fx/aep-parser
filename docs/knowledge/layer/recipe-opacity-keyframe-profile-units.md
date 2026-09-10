# ⚠ Recipe opacity keyframes author as percent, profile as unit opacity

Recipe `transform.opacity_keyframes` values are authored in AE-style percent `0..100`, but `expected_profile.keyframes[]` for `ADBE Opacity` must compare against profile-read unit values `0..1`.

The recipe schema writes layer Opacity keyframes in percent:

```json
"opacity_keyframes": [
  { "time": 0, "value": 20 },
  { "time": 2, "value": 100 },
  { "time": 4, "value": 45 }
]
```

`internal/profile` reads the resulting `ADBE Opacity` keyframe values as unit
opacity floats:

```json
"expected_profile": {
  "keyframes": [
    {
      "layer_name": "Title",
      "match_name": "ADBE Opacity",
      "keyframes": [
        { "time": 0, "value": 0.2 },
        { "time": 2, "value": 1 },
        { "time": 4, "value": 0.45 }
      ]
    }
  ]
}
```

Do not change recipe authoring to `0..1` unless the public recipe contract is
intentionally revised. `SetLayerTransform` receives opacity in percent, while
profile contracts must match parser-normalized values.
