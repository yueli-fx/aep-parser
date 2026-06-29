# Recipe Comp Background Color

Context: recipe `comp.background_color` support, proven with
`examples/recipes/minimal-comp-background-color.json`.

Recipe authoring:

- `background_color` is an RGB triplet in 0..255 channel units.

Writer capability:

- `SetBGColor`

Boundary:

- Background color is a composition setting, not a profile-visible property in
  the current stable profile shape.
- Contract coverage uses compiled AEP readback and AE render acceptance rather
  than `expected_profile.properties[]`.
