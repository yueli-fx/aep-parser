# Recipe Layer Timing

SUMMARY: Recipe support note for Layer Timing.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for layer fields related to Layer Timing

---

Context: recipe `layer.start_time`, `layer.in_point`, and `layer.out_point`
support, proven with `examples/recipes/minimal-layer-timing.json`.

Recipe authoring:

- `start_time` sets the layer timeline start time in seconds. Negative values
  are allowed by the writer for pre-roll.
- `in_point` sets the source-relative in point in seconds.
- `out_point` sets the source-relative out point in seconds.
- `in_point` and `out_point` are validated against the comp duration and
  `out_point` must be greater than or equal to `in_point`.

Writer capabilities:

- `Layer.SetStartTime`
- `Layer.SetInPoint`
- `Layer.SetOutPoint`

Boundary:

- These are length-preserving dividend/divisor pairs in `ldta`.
- `stretch` is intentionally excluded: its current capability is alpha and AE
  recomputes stretch from coordinated in/out spans.
- Contract coverage uses schema capability reporting, validation refusals,
  compiled AEP readback, and AE render/open acceptance.
