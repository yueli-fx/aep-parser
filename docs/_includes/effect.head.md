<!-- Hand-authored lead. Per-symbol docs below are generated from internal/aep
     doc comments. -->

Effects are the effect instances applied to a layer (Gaussian Blur, Tritone,
Curves, …), accessed via `layer.Effects[i]`. Each effect has a `MatchName` (its
type) and `Parameters` — every knob is a `Property`, so read/write them through
the [Property API](property.md) (`SetStaticValue`, keyframe setters, …).

Parameter match-names use the `<EffectMatchName>-NNNN` form (e.g.
`"ADBE Gaussian Blur 2-0001"` = Blurriness), where `NNNN` is the parameter's
internal index in the effect template.
