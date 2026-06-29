# Recipe Layer Parent

Context: recipe `layer.parent` support, proven with
`examples/recipes/minimal-layer-parent.json`.

Recipe authoring:

- `parent` is a string naming another layer in the same comp.
- The compiler resolves the name to that layer's generated AE layer ID and
  calls `Layer.SetParent`.
- Parent references are applied after all layers in the comp are created, so a
  layer may reference a parent that appears before or after it in the recipe.
- Validation refuses a parent name that does not exist in the same comp.

Writer capability:

- `Layer.SetParent`

Boundary:

- Parent is stored as a length-preserving 4-byte field in `ldta` at offset
  `0x84`.
- Recipe input deliberately uses layer names rather than raw IDs because IDs
  are generated during compilation.
- Contract coverage uses schema capability reporting, unknown-parent refusal,
  compiled AEP readback, and AE render/open acceptance.
