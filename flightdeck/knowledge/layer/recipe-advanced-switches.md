# Recipe Layer Advanced Switches

SUMMARY: Recipe support note for Layer Advanced Switches.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for layer fields related to Layer Advanced Switches

---

Context: recipe support for advanced layer switches, proven with
`examples/recipes/minimal-layer-advanced-switches.json`.

Recipe authoring:

- `collapse_transform` controls Collapse Transformations / Continuously
  Rasterize.
- `is_3d` controls the 3D-layer switch.
- `is_adjust` controls the Adjustment Layer switch.
- `is_guide` controls the Guide Layer switch.
- `sampling_bicubic` switches sampling from bilinear to bicubic.
- `frame_blend_pixel_motion` selects Pixel Motion when frame blending is
  enabled.
- `preserve_transparency` controls Preserve Underlying Transparency.

Writer capabilities:

- `Layer.SetCollapseTransform`
- `Layer.SetIs3D`
- `Layer.SetIsAdjust`
- `Layer.SetIsGuide`
- `Layer.SetSamplingBicubic`
- `Layer.SetFrameBlendPixelMotion`
- `Layer.SetPreserveTransparency`

Boundary:

- These are length-preserving layer switches in `ldta`.
- `frame_blend_pixel_motion` needs `frame_blend_enabled` to have a visible
  effect, so the dedicated recipe enables both.
- `is_guide` excludes the layer from rendered output even though AE can show it
  in the comp viewer; AE render/open acceptance is still the relevant gate.
- `markers_locked` is intentionally not part of this recipe slice because its
  current capability verification is `roundtrip`, not `ae-accept`.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  and AE render/open acceptance.
