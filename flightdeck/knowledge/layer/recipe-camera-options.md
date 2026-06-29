# Recipe Camera Options

Context: recipe camera option support, starting with
`examples/recipes/minimal-camera-zoom.json`.

Recipe authoring:

- `camera.zoom` on a `type: "camera"` layer sets the camera's Zoom property
  through `Layer.SetCameraZoom`.
- The value is a scalar in pixels, matching the underlying writer API.

Writer capability:

- `Layer.SetCameraZoom`

Boundary:

- `camera` options are only valid on `type: "camera"` layers.
- This slice covers `zoom` only. Depth of field, focus distance, aperture, iris,
  and other camera options remain separate recipe slices.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  embedded expected-profile property checks, and AE render/open acceptance.
