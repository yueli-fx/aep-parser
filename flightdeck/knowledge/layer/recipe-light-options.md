# Recipe Light Options

Context: recipe light option support, starting with
`examples/recipes/minimal-light-intensity.json` and
`examples/recipes/minimal-light-color.json`, plus
`examples/recipes/minimal-light-casts-shadows.json` and
`examples/recipes/minimal-light-shadow-darkness.json`.

Recipe authoring:

- `light.intensity` on a `type: "light"` layer sets the light's Intensity
  property through `Layer.SetLightIntensity`.
- The value is a scalar percentage, matching the underlying writer API.
- `light.color` sets the light's Color property through `Layer.SetLightColor`.
- Color values use the underlying light color channel order `[A, R, G, B]` in
  the 0..255 range. Three-channel values are also allowed by the writer API
  when the target property has three components.
- `light.casts_shadows` toggles the light's Casts Shadows property through
  `Layer.SetLightCastsShadows`.
- `light.shadow_darkness` sets the light's Shadow Darkness percentage through
  `Layer.SetLightShadowDarkness`.

Writer capability:

- `SetLightIntensity`
- `SetLightColor`
- `SetLightCastsShadows`
- `SetLightShadowDarkness`

Boundary:

- `light` options are only valid on `type: "light"` layers.
- Current coverage includes `intensity`, `color`, `casts_shadows`, and
  `shadow_darkness`. Cone, falloff, shadow diffusion, and other light options
  remain separate recipe slices.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  embedded expected-profile property checks, and AE render/open acceptance.
