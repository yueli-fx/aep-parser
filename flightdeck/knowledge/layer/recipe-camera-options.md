# Recipe Camera Options

Context: recipe camera option support, starting with
`examples/recipes/minimal-camera-zoom.json` and
`examples/recipes/minimal-camera-depth-of-field.json`.

Recipe authoring:

- `camera.zoom` on a `type: "camera"` layer sets the camera's Zoom property
  through `Layer.SetCameraZoom`.
- The value is a scalar in pixels, matching the underlying writer API.
- `camera.depth_of_field` sets the camera's Depth of Field toggle through
  `Layer.SetCameraDepthOfField`.

Writer capability:

- `Layer.SetCameraZoom`
- `Layer.SetCameraDepthOfField`

Boundary:

- `camera` options are only valid on `type: "camera"` layers.
- Current coverage includes `zoom` and `depth_of_field`. Focus distance,
  aperture, iris, and other camera options remain separate recipe slices.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  embedded expected-profile property checks, and AE render/open acceptance.
