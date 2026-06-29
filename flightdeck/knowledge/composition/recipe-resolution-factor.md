# Recipe Comp Resolution Factor

Context: recipe `comp.resolution_factor` support, proven with
`examples/recipes/minimal-comp-resolution-factor.json`.

Recipe authoring:

- `resolution_factor` -> `SetResolutionFactor`
- Shape is `[x, y]`, matching AE scripting's two-value resolution factor.

Validation:

- Exactly two values are required.
- Both values must be positive integers in `uint16` range.

Boundary:

- The current stable profile schema does not expose resolution factor, so the
  recipe compiler test verifies compiled AEP readback and the example is gated
  by AE render acceptance plus basic profile count checks.
