# Recipe Layer Shy

SUMMARY: Recipe support note for Layer Shy.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for layer fields related to Layer Shy

---

Context: recipe `layer.shy` support, proven with
`examples/recipes/minimal-layer-shy.json`.

Recipe authoring:

- `shy` is a boolean layer timeline switch.
- It marks the layer as shy; hiding shy layers also depends on the composition
  `hide_shy_layers` switch.

Writer capability:

- `Layer.SetShy`

Boundary:

- Layer shy is a length-preserving single bit in `ldta` offset `0x27`.
- The setter is stable and AE 2020/2025 accept-gated.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  and AE render/open acceptance.
