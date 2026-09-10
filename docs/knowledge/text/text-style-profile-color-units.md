# ⚠ Text-style expected_profile colors use normalized RGBA, not 0..255 property arrays

Text-style `expected_profile.text_styles[]` colors compare against profile `TextStyleRun` normalized `0..1` RGBA, while shape/property checks may expose `0..255` arrays.

`internal/profile.TextStyleRun.FillColor` and `StrokeColor` carry parsed text
run colors as normalized RGBA floats in `0..1`. Recipe input colors may be
authored as `0..255`, but the embedded `expected_profile.text_styles[]` contract
must compare against normalized values.

Example:

```json
"text_style": {
  "fill_color": [64, 128, 255, 255]
},
"expected_profile": {
  "text_styles": [
    {
      "layer_name": "Title",
      "fill_color": [0.25098039215686274, 0.5019607843137255, 1, 1]
    }
  ]
}
```

Do not copy the shape-property convention blindly. Shape `expected_profile.properties[]`
checks can surface some color properties as `0..255` arrays, for example
`ADBE Vector Stroke Color`, but text-style checks are based on `TextStyleRun`.
