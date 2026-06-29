# ⚠ Recipe transform expressions require a reopen-backed property

SUMMARY: Recipe `transform.expressions` writes expression source to parsed transform properties after a project reopen; do not call `SetExpression` on freshly generated properties without backrefs.
READ WHEN: adding recipe expression support; moving expression writes inside recipe compilation; debugging `no tdbs reference` from `Property.SetExpression`

---

Recipe transform expressions are authored under `transform.expressions`:

```json
"transform": {
  "position": [960, 540],
  "expressions": {
    "position": {
      "source": "[value[0] + time * 10, value[1], value[2]]"
    },
    "opacity": {
      "source": "time * 50",
      "enabled": false
    }
  }
}
```

`Property.SetExpression` requires parsed property backing (`tdbs` / `tdb4`).
The recipe compiler therefore writes the base layer transform first, reopens the
project with `aep.Reopen`, then applies expression source and optional enabled
state to the reopened transform properties.

`expected_profile.properties[]` can assert expression source:

```json
{
  "layer_name": "Title",
  "match_name": "ADBE Position",
  "expression": "[value[0] + time * 10, value[1], value[2]]"
}
```

The profile contract checks expression source only. It does not currently expose
or assert `ExpressionEnabled`; tests that need enabled/disabled proof should
reopen the compiled AEP and inspect `aep.Property.ExpressionEnabled` directly.
