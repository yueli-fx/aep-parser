<!-- Hand-authored. Package-level const block + label table (no single owning type). -->

## Property match-name constants

The five Transform properties' ADBE match-names, handy to pass to
`Layer.PropertyByMatchName` instead of typing the raw string:

```go
const (
    MatchNameAnchorPoint = "ADBE Anchor Point"
    MatchNamePosition    = "ADBE Position"
    MatchNameScale       = "ADBE Scale"
    MatchNameRotateZ     = "ADBE Rotate Z"
    MatchNameOpacity     = "ADBE Opacity"
)
```

The same-named accessors on `Layer` (`Position()`, …) are terser — see
[layer.md](layer.md#transform-accessors). 3D X/Y rotation match-names
(`"ADBE Rotate X"` / `"ADBE Rotate Y"`) have no constants; use the literals.

## Marker label index

`Marker.Label` is AE's timeline label-color index, `0..16`:

| Index | Default name (AE default palette) |
|---|---|
| 0 | None (layer default) |
| 1 | Red |
| 2 | Yellow |
| 3 | Aqua |
| 4 | Pink |
| 5 | Lavender |
| 6 | Peach |
| 7 | Sea Foam |
| 8 | Blue |
| 9 | Green |
| 10 | Purple |
| 11 | Orange |
| 12 | Brown |
| 13 | Fuchsia |
| 14 | Cyan |
| 15 | Sandstone |
| 16 | Dark Green |

> Users can remap labels to any RGB in AE Preferences → Labels; the index is
> stable, the RGB is not.
