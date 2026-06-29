# Recipe Comp Display Start Time

Context: recipe `comp.display_start_time` support, proven with
`examples/recipes/minimal-comp-display-start-time.json`.

Recipe authoring:

- `display_start_time` -> `SetDisplayStartTime`
- Value is seconds.

Validation:

- The value must be non-negative.

Boundary:

- This slice models second-based display start only.
- Frame-based display start remains available in the lower-level `aep` package
  but is not modeled in recipe IR here.
- The current stable profile schema does not expose display start time, so the
  recipe compiler test verifies compiled AEP readback and the example is gated
  by AE render acceptance plus basic profile count checks.
