# Rain Reference Phenomenon v1

## Purpose

Process the Motionbox rain sample as the next reference phenomenon for the
technique internalization pipeline.

This package turns a real `.aep` project into machine-readable technique
outputs and a reusable rain phenomenon recipe. It is an understanding package,
not a native generator or plugin-free render-equivalence claim.

## Inputs

- `data/samples/motionbox/generative/e8vfb9qjb9n0/雨の日モーショングラフィックス.aep`
- `cmd/aeptechnique`
- `cmd/aepdissect`
- `flightdeck/knowledge/techniques/fx-techniques.md`

## Outputs

Generated JSON artifacts:

- `tmp/technique_reference_rain/rain_summary.json`
- `tmp/technique_reference_rain/rain_explain.json`
- `tmp/technique_reference_rain/rain_explain.jsonl`
- `tmp/technique_reference_rain/package_summary.json`

Persistent knowledge:

- `flightdeck/knowledge/techniques/build-good-rain.md`

## Acceptance

- Rain sample parses through `aeptechnique`.
- Corpus summary records 1 project and 0 errors.
- Sample is classified as a shape/controller/effect-driven rain system.
- `tc Particular` and `KNSW Unmult` are recorded as exact-render plugin
  dependencies.
- The recipe records a stock-AE rain spine separately from the plugin-required
  exact route.
- `fx-techniques.md` points the rain phenomenon row to the recipe.
- Verification passes:

```powershell
go run ./cmd/aeptechnique -in data\samples\motionbox\generative\e8vfb9qjb9n0 -corpus -recursive -mode explain -out tmp\technique_reference_rain\rain_explain.jsonl -summary -summary-out tmp\technique_reference_rain\rain_summary.json
go run ./cmd/aeptechnique -in "data\samples\motionbox\generative\e8vfb9qjb9n0\雨の日モーショングラフィックス.aep" -mode explain -out tmp\technique_reference_rain\rain_explain.json
go test ./internal/technique ./cmd/aeptechnique -count=1
git diff --check
```

## Boundaries

- The sample is readable and useful for technique extraction, but exact render
  needs Trapcode Particular and Unmult.
- Shape/repeater rain strokes and native time/effect treatments are learnable
  without those plugins.
- No AE render gate was run for a generated rain showcase in this package.
- No third-party native approximation is claimed.

## Latest Execution

Completed as a single reference-phenomenon package.

Generated outputs:

- `tmp/technique_reference_rain/rain_summary.json` — 1 project, 5 comps, 45
  layers, 73 effects, 173 shape operators, 101 dependency edges.
- `tmp/technique_reference_rain/rain_explain.json` — readiness `needs_plugins`
  with blockers `tc Particular` x4 and `KNSW Unmult` x1.
- `tmp/technique_reference_rain/rain_explain.jsonl` — corpus record for the
  same sample.
- `tmp/technique_reference_rain/package_summary.json` — compact package
  summary and boundary ledger.

Persistent knowledge:

- `flightdeck/knowledge/techniques/build-good-rain.md` records the recipe.
- `flightdeck/knowledge/techniques/fx-techniques.md` links rain to the recipe
  and adds the shape-repeater rain-streak atom.

