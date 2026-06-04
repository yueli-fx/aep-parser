<!-- Hand-authored lead. Non-symbol prose lives here; per-symbol docs below are
     generated from internal/aep doc comments. -->

Properties are the animatable channels on a layer — Transform (Position /
Scale / Rotation / Anchor Point / Opacity), effect parameters, and mask
Feather / Opacity / Expansion are all `Property` values. Access them via
`layer.Properties[index]`, `layer.PropertyByMatchName(name)`, or the Transform
accessors (`layer.Position()`, `layer.Opacity()`, …; see
[layer.md](layer.md#transform-accessors)).

Value-type and enum reference tables are at the [end of this file](#value-reference).
