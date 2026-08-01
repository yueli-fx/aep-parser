# Recipe Comp Work Area

Recipe support note for Comp Work Area.

Context: recipe `comp.work_area` support, proven with
`examples/recipes/minimal-comp-work-area.json`.

Recipe authoring:

- `start` -> `SetWorkArea` start seconds
- `end` -> `SetWorkArea` end seconds

Validation:

- `start` and `end` are both required when `work_area` is present.
- Values must satisfy `0 <= start <= end <= comp.duration`.

Profile contract:

- `expected_profile.work_area.start`
- `expected_profile.work_area.end`

Boundary:

- This slice uses second-based work area authoring only.
- Frame-based work area setters remain available in the lower-level `aep`
  package but are not modeled in recipe IR here.
- `SetWorkArea` is applied after shutter/sample setters because shutter setter
  writes can touch work-area cdta bytes as a cosmetic side effect.
