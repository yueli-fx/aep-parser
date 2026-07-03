# Recipe Comp Hide Shy Layers

SUMMARY: Recipe support note for Comp Hide Shy Layers.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for composition fields related to Comp Hide Shy Layers

---

Context: recipe `comp.hide_shy_layers` support, proven with
`examples/recipes/minimal-comp-hide-shy-layers.json`.

Recipe authoring:

- `hide_shy_layers` -> `SetHideShyLayers`
- Value is a boolean.

Validation:

- No value-domain validation beyond JSON boolean type.

Boundary:

- This is the composition master "Hide Shy Layers" timeline toggle.
- It only affects layers that are themselves marked shy; recipe layer-level shy
  flags are not modeled in this slice.
- The current stable profile schema does not expose this comp flag, so the
  recipe compiler test verifies compiled AEP `cdta @0x8B bit 0x01` readback and
  the example is gated by AE render acceptance plus basic profile count checks.
