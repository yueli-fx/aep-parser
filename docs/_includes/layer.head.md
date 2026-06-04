<!-- Hand-authored lead. Per-symbol docs below are generated from internal/aep
     doc comments. -->

A `Layer` is one layer in a composition's timeline, reached via `comp.Layers[i]`,
`comp.LayerByID`, or `comp.LayerByName`. The `Type` field says what kind it is
(AV / shape / text / camera / light / null / adjustment); the type-specific
accessors (camera, light, material, text) only return meaningful values for the
matching layer type.

Transform properties have convenience accessors — `Position()`, `Scale()`,
`Rotation()`, `AnchorPoint()`, `Opacity()` — each returning a
[`Property`](property.md). Text layers expose content via `TextSource` plus the
`SetText` / `SetRun*` setters documented here (they live on `Layer`); see also
[text.md](text.md). Enum field types (`LayerType`, `BlendingMode`,
`TrackMatteType`, …) are catalogued in [constants.md](constants.md).
