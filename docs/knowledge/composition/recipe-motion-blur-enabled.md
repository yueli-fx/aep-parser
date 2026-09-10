# Recipe Comp Motion Blur Enabled

Recipe support note for Comp Motion Blur Enabled.

Context: recipe `comp.motion_blur.enabled` support, proven with
`examples/recipes/minimal-comp-motion-blur-enabled.json`.

Recipe authoring:

- `motion_blur.enabled` -> `SetCompMotionBlur`
- Value is a boolean.

Validation:

- No value-domain validation beyond JSON boolean type.

Boundary:

- This is the composition motion-blur master switch only.
- Layers still need their own per-layer motion-blur flag to render motion blur.
- The current stable profile schema does not expose this comp flag, so the
  recipe compiler test verifies compiled AEP `cdta @0x8B bit 0x08` readback and
  the example is gated by AE render acceptance plus basic profile count checks.
