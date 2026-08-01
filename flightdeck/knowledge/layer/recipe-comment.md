# Recipe Layer Comment

Recipe support note for Layer Comment.

Context: recipe `layer.comment` support, proven with
`examples/recipes/minimal-layer-comment.json`.

Recipe authoring:

- `comment` writes the AE timeline Comments column / Layer Settings comment.
- Empty string means no comment write.

Writer capability:

- `Layer.SetComment`

Boundary:

- Layer comment is length-variable metadata stored in a `cmta` chunk.
- The writer must keep the `cmta` double-NUL terminator and the `ldta` offset
  `0x3C` has-comment flag in sync; otherwise AE can read back an empty comment.
- The current stable profile shape does not expose layer comments, so
  `expected_profile` cannot assert it directly.
- Contract coverage uses schema capability reporting, compiled AEP readback,
  and AE render/open acceptance.
