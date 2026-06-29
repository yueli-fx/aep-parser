# Recipe Light Options

Context: recipe light option support, starting with
`examples/recipes/minimal-light-intensity.json`.

Recipe authoring:

- `light.intensity` on a `type: "light"` layer sets the light's Intensity
  property through `Layer.SetLightIntensity`.
- The value is a scalar percentage, matching the underlying writer API.

Writer capability:

- `SetLightIntensity`

Boundary:

- `light` options are only valid on `type: "light"` layers.
- Current coverage includes `intensity`. Cone, falloff, shadow, color, and
  other light options remain separate recipe slices.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  embedded expected-profile property checks, and AE render/open acceptance.
