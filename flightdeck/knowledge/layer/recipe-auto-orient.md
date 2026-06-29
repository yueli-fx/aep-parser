# Recipe Layer Auto Orient

Context: recipe `layer.auto_orient` support, proven with
`examples/recipes/minimal-layer-auto-orient.json`.

Recipe authoring:

- `auto_orient` accepts one of:
  - `none`
  - `along_path`
  - `camera_or_point_of_interest`
  - `characters_toward_camera`
- The compiler maps the string to `Layer.SetAutoOrient`.
- Invalid values are refused during validation.

Writer capability:

- `Layer.SetAutoOrient`

Boundary:

- Auto-orient is a mutually-exclusive bit group spread across `ldta` offsets
  `0x25` and `0x26`.
- `along_path` needs a motion path to have visible meaning; the dedicated
  example includes Position keyframes so AE has a path carrier.
- Contract coverage uses schema capability reporting, invalid enum refusal,
  compiled AEP readback, and AE render/open acceptance.
