# Recipe Layer Label

SUMMARY: Recipe support note for Layer Label.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for layer fields related to Layer Label

---

Context: recipe `layer.label` support, proven with
`examples/recipes/minimal-layer-label.json`.

Recipe authoring:

- `label` is the AE timeline/project-panel label color index for a layer.
- Valid values are integer indexes `0..16`.

Writer capability:

- `Layer.SetLabel`

Boundary:

- Layer label is stored on the layer item metadata (`ldta`, offset `0x3D`).
- The current stable profile shape does not expose layer label, so
  `expected_profile` cannot assert it directly.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  and AE render/open acceptance.
