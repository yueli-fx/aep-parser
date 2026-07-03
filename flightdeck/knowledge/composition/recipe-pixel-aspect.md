# Recipe Comp Pixel Aspect

SUMMARY: Recipe support note for Comp Pixel Aspect.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for composition fields related to Comp Pixel Aspect

---

Context: recipe `comp.pixel_aspect` support, proven with
`examples/recipes/minimal-comp-pixel-aspect.json`.

Recipe authoring:

- `pixel_aspect` -> `SetPixelAspect`
- Value is a positive pixel aspect ratio, for example `1`, `1.33`, or `2`.

Validation:

- The value must be positive.

Boundary:

- The current stable profile schema does not expose pixel aspect, so the
  recipe compiler test verifies compiled AEP readback and the example is gated
  by AE render acceptance plus basic profile count checks.
