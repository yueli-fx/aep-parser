# ⚠ Recipe keyframe ease influence is a fraction, not AE UI percent

Recipe transform keyframes can set optional `in_ease` / `out_ease` objects, but `influence` is authored as the writer fraction `0..1`, not AE UI percent `0..100`.

Recipe transform keyframes support optional easing:

```json
"position_keyframes": [
  {
    "time": 0,
    "value": [760, 540],
    "out_ease": { "speed": 0, "influence": 0.8 }
  },
  {
    "time": 2,
    "value": [1160, 540],
    "in_ease": { "speed": 0, "influence": 0.35 }
  }
]
```

`influence` intentionally matches `aep.TemporalEase.Influence`, where `0.8`
means 80%. Do not write `80` in recipe JSON unless the public recipe contract
is intentionally revised; validation rejects values `<= 0` or `> 1`.

When neither side is present, recipe compilation still emits linear keyframes
through `AddKeyframeLinear`. When either side is present, compilation uses
`AddKeyframeWithEase`, with the missing side left as the zero-value linear side.
