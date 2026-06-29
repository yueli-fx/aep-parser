# Recipe Camera Options

Context: recipe camera option support, starting with
`examples/recipes/minimal-camera-zoom.json` and
`examples/recipes/minimal-camera-depth-of-field.json` and
`examples/recipes/minimal-camera-focus-distance.json` and
`examples/recipes/minimal-camera-aperture.json` and
`examples/recipes/minimal-camera-blur-level.json`.

Recipe authoring:

- `camera.zoom` on a `type: "camera"` layer sets the camera's Zoom property
  through `Layer.SetCameraZoom`.
- The value is a scalar in pixels, matching the underlying writer API.
- `camera.depth_of_field` sets the camera's Depth of Field toggle through
  `Layer.SetCameraDepthOfField`.
- `camera.focus_distance` sets the camera's Focus Distance property through
  `Layer.SetCameraFocusDistance`.
- `camera.aperture` sets the camera's Aperture property through
  `Layer.SetCameraAperture`.
- `camera.blur_level` sets the camera's Blur Level property through
  `Layer.SetCameraBlurLevel`.

Writer capability:

- `Layer.SetCameraZoom`
- `Layer.SetCameraDepthOfField`
- `Layer.SetCameraFocusDistance`
- `Layer.SetCameraAperture`
- `Layer.SetCameraBlurLevel`

Boundary:

- `camera` options are only valid on `type: "camera"` layers.
- Current coverage includes `zoom`, `depth_of_field`, `focus_distance`,
  `aperture`, and `blur_level`. Iris and other camera options remain separate
  recipe slices.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  embedded expected-profile property checks, and AE render/open acceptance.
