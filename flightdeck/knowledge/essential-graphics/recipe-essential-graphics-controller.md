# Recipe Essential Graphics controller slice

SUMMARY: Recipe `effects[].params[].essential_graphics` exposes a static effect parameter as an Essential Graphics controller through `AddEssentialProperty`.
READ WHEN: adding recipe Essential Graphics fields; debugging `expected_profile.essential_graphics[]`; deciding whether keyframed, expression, color, layer, or point params are in scope

---

Recipe syntax:

```json
{
  "match_name": "ADBE Slider Control-0001",
  "value": 42,
  "essential_graphics": {
    "name": "Amount"
  }
}
```

This first recipe slice is intentionally narrow:

- The parameter must have a static `value`.
- `target_layer`, `keyframes`, and `expression` are refused when
  `essential_graphics` is present.
- The compiler materializes the effect param, then calls
  `AddEssentialProperty(layer, fx, paramMatchName, name)`.
- `expected_profile.essential_graphics[]` checks the first comp's EG panel order
  by controller `name` and optional `type`.

The coverage recipe is
`examples/recipes/minimal-essential-graphics-controller.json`.

This does not mean recipe IR supports full EG panel authoring yet. Color
controllers, point/dropdown/text controllers, Transform-source controllers, and
keyframed/expression-backed exposure need separate slices and evidence.
