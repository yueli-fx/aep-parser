# Recipe Layer Motion Blur

Recipe support note for Layer Motion Blur.

Context: recipe `layer.motion_blur` support, proven with
`examples/recipes/minimal-layer-motion-blur.json`.

Recipe authoring:

- `motion_blur` is a boolean layer switch.
- It controls the layer-level motion blur checkbox; visible blur still depends
  on the composition motion blur switch/settings.

Writer capability:

- `Layer.SetMotionBlur`

Boundary:

- Layer motion blur is a length-preserving single bit in `ldta` offset `0x27`.
- The setter is stable and AE 2020/2025 accept-gated.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  and AE render/open acceptance.
