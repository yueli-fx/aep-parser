# Recipe Camera Options

Context: recipe camera option support, starting with
`examples/recipes/minimal-camera-zoom.json` and
`examples/recipes/minimal-camera-depth-of-field.json` and
`examples/recipes/minimal-camera-focus-distance.json` and
`examples/recipes/minimal-camera-aperture.json` and
`examples/recipes/minimal-camera-blur-level.json` and
`examples/recipes/minimal-camera-iris-shape.json` and
`examples/recipes/minimal-camera-iris-rotation.json` and
`examples/recipes/minimal-camera-iris-roundness.json` and
`examples/recipes/minimal-camera-iris-aspect-ratio.json` and
`examples/recipes/minimal-camera-iris-diffraction-fringe.json` and
`examples/recipes/minimal-camera-iris-highlight-gain.json`.

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
- `camera.iris_shape` sets the camera's Iris Shape enum through
  `Layer.SetIrisShape`.
- `camera.iris_rotation` sets the camera's Iris Rotation property through
  `Layer.SetIrisRotation`.
- `camera.iris_roundness` sets the camera's Iris Roundness property through
  `Layer.SetIrisRoundness`.
- `camera.iris_aspect_ratio` sets the camera's Iris Aspect Ratio property
  through `Layer.SetIrisAspectRatio`.
- `camera.iris_diffraction_fringe` sets the camera's Iris Diffraction Fringe
  property through `Layer.SetIrisDiffractionFringe`.
- `camera.iris_highlight_gain` sets the camera's Iris Highlight Gain property
  through `Layer.SetIrisHighlightGain`.

Writer capability:

- `Layer.SetCameraZoom`
- `Layer.SetCameraDepthOfField`
- `Layer.SetCameraFocusDistance`
- `Layer.SetCameraAperture`
- `Layer.SetCameraBlurLevel`
- `SetIrisShape`
- `SetIrisRotation`
- `SetIrisRoundness`
- `SetIrisAspectRatio`
- `SetIrisDiffractionFringe`
- `SetIrisHighlightGain`

Boundary:

- `camera` options are only valid on `type: "camera"` layers.
- Current coverage includes `zoom`, `depth_of_field`, `focus_distance`,
  `aperture`, `blur_level`, `iris_shape`, `iris_rotation`, and
  `iris_roundness`, `iris_aspect_ratio`, and `iris_diffraction_fringe`. Other
  iris and camera options, including highlight threshold/saturation, remain
  separate recipe slices.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  embedded expected-profile property checks, and AE render/open acceptance.
