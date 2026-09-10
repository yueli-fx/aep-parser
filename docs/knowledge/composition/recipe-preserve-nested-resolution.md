# Recipe Comp Preserve Nested Resolution

Recipe support note for Comp Preserve Nested Resolution.

Context: recipe `comp.preserve_nested_resolution` support, proven with
`examples/recipes/minimal-comp-preserve-nested-resolution.json`.

Recipe authoring:

- `preserve_nested_resolution` -> `SetPreserveNestedResolution`
- Value is a boolean.

Validation:

- No value-domain validation beyond JSON boolean type.

Boundary:

- This is the composition "preserve resolution when nested" toggle.
- It affects behavior when this comp is used as a nested/precomp source.
- The current stable profile schema does not expose this comp flag, so the
  recipe compiler test verifies compiled AEP `cdta @0x8B bit 0x80` readback and
  the example is gated by AE render acceptance plus basic profile count checks.
