# Recipe Layer Null Flag

Context: recipe `layer.is_null` support, proven with
`examples/recipes/minimal-layer-null-flag.json`.

Recipe authoring:

- `is_null` is a boolean and maps to `Layer.SetIsNull`.
- The dedicated example uses a solid layer named `Controller` with
  `is_null: true`, hides it, and parents a text layer to it.

Writer capability:

- `Layer.SetIsNull`

Boundary:

- `is_null` flips AE's Null-Object marker bit at `ldta` offset `0x26`.
- This exposes the low-level flag on an existing recipe layer. It is not the
  same as AE's UI command or the library's higher-level `NewNullLayer`, which
  creates the standard 100x100 solid-backed null helper.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  and AE render/open acceptance.
