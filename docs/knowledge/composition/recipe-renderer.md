# Recipe Comp Renderer

Recipe support note for Comp Renderer.

Context: recipe `comp.renderer` support, proven with
`examples/recipes/minimal-comp-renderer.json`.

Recipe authoring:

- `renderer` -> `SetRenderer`
- Use binary renderer match names in recipe contracts, such as `ADBE Escher`.

Profile contract:

- `expected_profile.renderer`

Boundary:

- `SetRenderer` can normalize known ExtendScript module names, but recipe
  examples and contracts should prefer the binary match name so profile checks
  stay exact.
- Available engines vary by AE version; unsupported names fail at compile time.
