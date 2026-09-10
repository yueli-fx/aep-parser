# Recipe Null Layer

Recipe support note for Null Layer.

Context: recipe `type: "null"` support, proven with
`examples/recipes/minimal-null-layer.json`.

Recipe authoring:

- A layer with `type: "null"` creates a standard null helper through
  `NewNullLayer`.
- The layer can still receive common recipe fields such as `name`, `parent`,
  switches, and `transform`.
- The dedicated example creates `Controller` as a null layer and parents a text
  layer to it.

Writer capability:

- `NewNullLayer`

Boundary:

- This is the higher-level null creation path. It differs from `is_null`,
  which only flips the low-level Null-Object marker bit on an already-created
  recipe layer.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  parent linkage, and AE render/open acceptance.
