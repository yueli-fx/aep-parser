# Recipe Adjustment Layer

Context: recipe `type: "adjustment"` support, proven with
`examples/recipes/minimal-adjustment-layer.json`.

Recipe authoring:

- A layer with `type: "adjustment"` creates a comp-sized adjustment layer
  through `NewAdjustmentLayer`.
- The layer can receive common recipe fields such as `name`, switches,
  transforms, and effects.
- The dedicated example creates a `Grade` adjustment layer above a text layer.

Writer capability:

- `NewAdjustmentLayer`

Boundary:

- This is the higher-level adjustment-layer creation path. It differs from
  `is_adjust`, which only flips the low-level adjustment marker bit on an
  already-created recipe layer.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  and AE render/open acceptance.
