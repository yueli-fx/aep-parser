# ⚠ Recipe effect param expressions need a materializing value

SUMMARY: Recipe effect params can set an optional `expression`, but `value` is still required so `SetEffectParam` can materialize the parameter before `SetExpression` runs.
READ WHEN: adding recipe effect-parameter expressions; debugging `unsupported_effect_param_value`; writing `expected_profile.effects[].params[]` expression checks

---

Effect parameters are usually elided until a writer materializes them. Recipe
expression support therefore keeps `value` required:

```json
"effects": [
  {
    "match_name": "ADBE Gaussian Blur 2",
    "params": [
      {
        "match_name": "ADBE Gaussian Blur 2-0001",
        "value": 0,
        "expression": {
          "source": "time * 40",
          "enabled": false
        }
      }
    ]
  }
]
```

The compiler calls `SetEffectParam` first, then applies `Property.SetExpression`
and optional `Property.SetExpressionEnabled` to the returned property.

`expected_profile.effects[].params[]` can assert source:

```json
{
  "match_name": "ADBE Gaussian Blur 2-0001",
  "value": 0,
  "expression": "time * 40"
}
```

The profile contract checks source only. It does not currently expose the
enabled/disabled expression toggle; tests that need that proof should reopen the
compiled AEP and inspect `aep.Property.ExpressionEnabled`.
