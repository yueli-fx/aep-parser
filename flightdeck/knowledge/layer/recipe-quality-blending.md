# Recipe Layer Quality And Blending

Context: recipe `layer.quality` and `layer.blending_mode` support, proven with
`examples/recipes/minimal-layer-quality-blending.json`.

Recipe authoring:

- `quality` accepts `wireframe`, `draft`, or `best`.
- `blending_mode` accepts AE blending mode names as lowercase snake case, for
  example `normal`, `multiply`, `screen`, `overlay`, `add`, `subtract`, or
  `divide`.

Writer capabilities:

- `Layer.SetQuality`
- `Layer.SetBlendingMode`

Boundary:

- Layer quality is a length-preserving uint16 in `ldta` offset `0x04`.
- Blending mode is a length-preserving byte in `ldta` offset `0x63`.
- Both setters are stable and AE 2020/2025 accept-gated.
- Contract coverage uses schema capability reporting, validation refusals,
  compiled AEP readback, and AE render/open acceptance.
