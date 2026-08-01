# Recipe Light Layer

Recipe support note for Light Layer.

Context: recipe `type: "light"` support, proven with
`examples/recipes/minimal-light-layer.json`.

Recipe authoring:

- A layer with `type: "light"` creates a light layer through `NewLightLayer`.
- The dedicated example creates a `Light` layer plus a visible text layer so AE
  render acceptance has frame output to inspect.

Writer capability:

- `NewLightLayer`

Boundary:

- This slice covers high-level light layer creation only.
- Recipe transform validation still requires 2-value vectors, so 3D light
  placement is intentionally deferred.
- Light option setters such as kind, intensity, cone angle, falloff, and shadow
  controls are also deferred to separate recipe slices.
