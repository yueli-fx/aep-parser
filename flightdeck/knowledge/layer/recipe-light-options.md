# Recipe Light Options

Context: recipe light option support, starting with
`examples/recipes/minimal-light-intensity.json` and
`examples/recipes/minimal-light-color.json`, plus
`examples/recipes/minimal-light-casts-shadows.json` and
`examples/recipes/minimal-light-shadow-darkness.json` /
`examples/recipes/minimal-light-shadow-diffusion.json` /
`examples/recipes/minimal-light-falloff-type.json` /
`examples/recipes/minimal-light-falloff-start.json` /
`examples/recipes/minimal-light-falloff-distance.json` /
`examples/recipes/minimal-light-cone-angle.json` /
`examples/recipes/minimal-light-cone-feather.json` /
`examples/recipes/minimal-light-kind.json`.

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
- `light.shadow_diffusion` sets the light's Shadow Diffusion distance in pixels
  through `Layer.SetLightShadowDiffusion`.
- `light.falloff_type` sets the light's Falloff Type enum through
  `Layer.SetLightFalloffType`.
- `light.falloff_start` sets the light's Falloff Start distance in pixels
  through `Layer.SetLightFalloffStart`. It is only meaningful when
  `falloff_type` is not the none/default mode.
- `light.falloff_distance` sets the light's Falloff Distance in pixels through
  `Layer.SetLightFalloffDistance`.
- `light.cone_angle` sets the spotlight Cone Angle in degrees through
  `Layer.SetLightConeAngle`. It is only meaningful for spot lights.
- `light.cone_feather` sets the spotlight Cone Feather percentage through
  `Layer.SetLightConeFeather`. It is only meaningful for spot lights.
- `light.kind` sets the light kind through `Layer.SetLightKind`. Supported
  strings are `parallel`, `spot`, `point`, and `ambient`.

Writer capability:

- `SetLightIntensity`
- `SetLightColor`
- `SetLightCastsShadows`
- `SetLightShadowDarkness`
- `SetLightShadowDiffusion`
- `SetLightFalloffType`
- `SetLightFalloffStart`
- `SetLightFalloffDistance`
- `SetLightConeAngle`
- `SetLightConeFeather`
- `SetLightKind`

Boundary:

- `light` options are only valid on `type: "light"` layers.
- Current coverage includes `intensity`, `color`, `casts_shadows`, and
  `shadow_darkness` / `shadow_diffusion` plus `falloff_type` /
  `falloff_start` / `falloff_distance` plus `cone_angle` / `cone_feather` plus
  `kind`. Light source and other light options remain separate recipe slices.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  embedded expected-profile property checks, and AE render/open acceptance.
