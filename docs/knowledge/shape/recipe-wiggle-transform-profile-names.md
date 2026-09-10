# Recipe Wiggle Transform profile names mix Xform and shared modulation names

Recipe `shape.wiggle_transform` maps nested transform amplitudes to `ADBE Vector Wiggler ...`; only wiggles-per-second profiles as `ADBE Vector Xform Temporal Freq`, while seed/correlation/phase use shared `ADBE Vector ...` modulation names.

The recipe keeps Wiggle Transform authoring grouped under one field:

```json
"wiggle_transform": {
  "anchor": [10, 12],
  "position": [80, 60],
  "scale": [20, 30],
  "rotation": 25,
  "wiggles_per_second": 4,
  "random_seed": 9,
  "correlation": 80,
  "temporal_phase": 45,
  "spatial_phase": 20
}
```

The profile names are not all prefixed the same way:

- `anchor` -> `ADBE Vector Wiggler Anchor`
- `position` -> `ADBE Vector Wiggler Position`
- `scale` -> `ADBE Vector Wiggler Scale`
- `rotation` -> `ADBE Vector Wiggler Rotation`
- `wiggles_per_second` -> `ADBE Vector Xform Temporal Freq`
- `random_seed` -> `ADBE Vector Random Seed`
- `correlation` -> `ADBE Vector Correlation`
- `temporal_phase` -> `ADBE Vector Temporal Phase`
- `spatial_phase` -> `ADBE Vector Spatial Phase`

Do not assert `ADBE Vector Xform Random Seed` or `ADBE Vector Xform
Correlation`; those are not the parser output for the current writer.
