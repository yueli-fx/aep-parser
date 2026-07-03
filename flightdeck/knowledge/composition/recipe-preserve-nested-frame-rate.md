# Recipe Comp Preserve Nested Frame Rate

SUMMARY: Recipe support note for Comp Preserve Nested Frame Rate.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for composition fields related to Comp Preserve Nested Frame Rate

---

Context: recipe `comp.preserve_nested_frame_rate` support, proven with
`examples/recipes/minimal-comp-preserve-nested-frame-rate.json`.

Recipe authoring:

- `preserve_nested_frame_rate` -> `SetPreserveNestedFrameRate`
- Value is a boolean.

Validation:

- No value-domain validation beyond JSON boolean type.

Boundary:

- This is the composition "preserve frame rate when nested" toggle.
- It affects behavior when this comp is used as a nested/precomp source.
- The current stable profile schema does not expose this comp flag, so the
  recipe compiler test verifies compiled AEP `cdta @0x8B bit 0x20` readback and
  the example is gated by AE render acceptance plus basic profile count checks.
