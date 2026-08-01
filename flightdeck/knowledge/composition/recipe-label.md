# Recipe Comp Label

Recipe support note for Comp Label.

Context: recipe `comp.label` support, proven with
`examples/recipes/minimal-comp-label.json`.

Recipe authoring:

- `label` is the AE project-panel label color index.
- Valid values are integer indexes `0..16`.

Writer capability:

- `SetLabel`

Boundary:

- Composition label is an item-level project panel field, not a cdta comp
  setting.
- `SetLabel` needs parsed item backrefs (`idta` carrier). From-scratch recipe
  compilation must build the base project first, `Reopen` it, then apply the
  label on the reopened composition.
- Label is not exposed in the current stable profile shape. Contract coverage
  uses compiled AEP readback and AE render/open acceptance rather than
  `expected_profile.properties[]`.
