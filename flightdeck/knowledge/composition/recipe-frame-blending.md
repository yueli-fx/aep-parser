# Recipe Comp Frame Blending

SUMMARY: Recipe support note for Comp Frame Blending.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for composition fields related to Comp Frame Blending

---

Context: recipe `comp.frame_blending` support, proven with
`examples/recipes/minimal-comp-frame-blending.json`.

Recipe authoring:

- `frame_blending` -> `SetFrameBlending`
- Value is a boolean.

Validation:

- No value-domain validation beyond JSON boolean type.

Boundary:

- This is the composition master frame-blending switch only.
- Visible interpolation still requires frame-blending-capable layers with their
  own layer frame-blend flags/modes enabled.
- The current stable profile schema does not expose this comp flag, so the
  recipe compiler test verifies compiled AEP `cdta @0x8B bit 0x10` readback and
  the example is gated by AE render acceptance plus basic profile count checks.
