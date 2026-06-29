# Recipe Comp Comment

Context: recipe `comp.comment` support, proven with
`examples/recipes/minimal-comp-comment.json`.

Recipe authoring:

- `comment` writes the AE project-panel comment for the composition item.
- Empty comments are omitted by recipe validation/compilation in the current
  schema; use a non-empty string when exercising `SetComment`.

Writer capability:

- `SetComment`

Boundary:

- Composition comment is an item-level project panel field, not a cdta comp
  setting.
- `SetComment` is length-variable and needs parsed item backrefs (`idta` +
  `cmta` carrier). From-scratch recipe compilation must build the base project
  first, `Reopen` it, then apply the comment on the reopened composition.
- The writer must set the idta has-comment flag together with the cmta payload;
  otherwise AE can silently drop the comment even when bytes appear present.
- Comment is not exposed in the current stable profile shape. Contract coverage
  uses compiled AEP readback and AE render/open acceptance rather than
  `expected_profile.properties[]`.
