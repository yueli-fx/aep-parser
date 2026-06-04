<!-- Hand-authored lead. Per-symbol docs below are generated from internal/aep
     doc comments. -->

Masks are a layer's vector masks. Each mask has one Bezier path (closed or
open), a mode (Add / Subtract / …), and optional Feather / Opacity / Expansion.
Access them via `layer.Masks[i]`.

A path is either **static** (`Vertices`, a single snapshot) or **animated**
(`PathKeyframes`, one full path snapshot per time with a scalar ease). For an
animated mask, `Vertices` mirrors `PathKeyframes[0].Vertices` so callers that
don't care about animation can ignore the distinction.
