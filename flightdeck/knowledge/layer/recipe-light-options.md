# Recipe Light Options

Context: recipe light option support, starting with
`examples/recipes/minimal-light-intensity.json` and
`examples/recipes/minimal-light-color.json`.

Recipe authoring:

- `light.intensity` on a `type: "light"` layer sets the light's Intensity
  property through `Layer.SetLightIntensity`.
- The value is a scalar percentage, matching the underlying writer API.
- `light.color` sets the light's Color property through `Layer.SetLightColor`.
- Color values use the underlying light color channel order `[A, R, G, B]` in
  the 0..255 range. Three-channel values are also allowed by the writer API
  when the target property has three components.

Writer capability:

- `SetLightIntensity`
- `SetLightColor`

Boundary:

- `light` options are only valid on `type: "light"` layers.
- Current coverage includes `intensity` and `color`. Cone, falloff, shadow, and
  other light options remain separate recipe slices.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  embedded expected-profile property checks, and AE render/open acceptance.
