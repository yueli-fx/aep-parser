# Recipe Stroke Wave Profile Fields

SUMMARY: Recipe support note for Stroke Wave Profile Fields.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for shape fields related to Stroke Wave Profile Fields

---

Context: recipe `shape.stroke.wave` support, proven with
`examples/recipes/minimal-shape-stroke-wave.json`.

Recipe authoring:

- `amount` -> wave amount percent
- `wavelength` -> wave wavelength
- `phase` -> wave phase degrees

Profile readback:

- `ADBE Vector Taper Wave Amount`
- `ADBE Vector Taper Wavelength`
- `ADBE Vector Taper Wave Phase`

Boundary:

- The writer models the stroke wave wavelength-mode scalar controls.
- Pixel-mode wave mirror streams and the wave unit enum are not modeled in
  recipe IR yet.
