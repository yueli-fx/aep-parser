<!-- Hand-authored reference tables. go/doc comments have no table syntax, so
     these concept tables are maintained here and appended by docgen. -->

## Value reference

### Components → Go type

`Property.Components` fixes the Go type of `Keyframe.Value` and `StaticValue`:

| Components | Meaning | Value / StaticValue type |
|---|---|---|
| 1 | scalar (Opacity / Rotation / effect knob) | `float64` |
| 2 | 2D point (Mask Feather X+Y) | `[]float64` (len 2) |
| 3 | 3D point (Position / Scale / Anchor Point) | `[]float64` (len 3) |
| 4 | RGBA color (Tritone Highlights, …) | `[]float64` (len 4) |

### PropertyControlType (`ControlType()`)

| Constant | Value | Meaning |
|---|---|---|
| `PCTLLayer` | 0 | layer reference |
| `PCTLInteger` | 1 | integer |
| `PCTLScalar` | 2 | scalar slider |
| `PCTLAngle` | 3 | angle dial |
| `PCTLBoolean` | 4 | checkbox |
| `PCTLColor` | 5 | color picker |
| `PCTLTwoD` | 6 | 2D point |
| `PCTLEnum` | 7 | dropdown |
| `PCTLThreeD` | 18 | 3D point |
| `PCTLUnknown` | 15 | unknown |

### PropertyValueType (`ValuePropertyType()`)

Mirrors ExtendScript's `Property.propertyValueType`.

| Constant | Value | Meaning |
|---|---|---|
| `PVTUnknown` | 0 | unknown |
| `PVTNoValue` | 6412 | no value (separator / button) |
| `PVTThreeDSpatial` | 6413 | 3D spatial (Position) |
| `PVTThreeD` | 6414 | 3D non-spatial (Scale) |
| `PVTTwoDSpatial` | 6415 | 2D spatial |
| `PVTTwoD` | 6416 | 2D non-spatial |
| `PVTOneD` | 6417 | scalar |
| `PVTColor` | 6418 | RGBA color |

### Keyframe temporal-ease length

The length of `Keyframe.InTemporalEase` / `OutTemporalEase` depends on the
property kind:

| Property kind | Ease slice length |
|---|---|
| Spatial (Position / Anchor Point) | 1 (speed along the motion path) |
| Non-spatial 1D (Opacity) | 1 |
| Non-spatial N-D (Scale 3D / Feather 2D / Color 4D) | N |
