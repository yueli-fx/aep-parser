# Recipe Layer Common Switches

SUMMARY: Recipe support note for Layer Common Switches.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for layer fields related to Layer Common Switches

---

Context: recipe support for common layer switches, proven with
`examples/recipes/minimal-layer-common-switches.json`.

Recipe authoring:

- `visible` controls the layer video/eye switch.
- `solo` controls the layer solo flag.
- `locked` controls the layer lock flag.
- `effects_enabled` controls the layer fx switch.
- `audio_enabled` controls the layer audio switch.
- `frame_blend_enabled` controls the layer frame blending switch.

Writer capabilities:

- `Layer.SetVisible`
- `Layer.SetSolo`
- `Layer.SetLocked`
- `Layer.SetEffectsEnabled`
- `Layer.SetAudioEnabled`
- `Layer.SetFrameBlendEnabled`

Boundary:

- These are length-preserving layer switches in `ldta`, mostly at offset
  `0x27`; solo uses `ldta` offset `0x26`.
- The setters are stable and AE accept-gated.
- Some switches only have a visual effect with matching context:
  `effects_enabled` needs effects, `audio_enabled` needs audio, and
  `frame_blend_enabled` needs time-based frames.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  and AE render/open acceptance.
