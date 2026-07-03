# Effect Field Report Surface v1

## Goal

Expose `effect-field-understanding` results through the selfhost technique
report artifact set so the learning/report workflow can consume effect
reproducibility and study-priority data without manually opening a separate
temporary JSON.

This is the next coherent package after `effect-field-understanding-v1`.

## Inputs

Primary generated input:

```text
tmp/effect_field_understanding/understanding.json
```

Regenerating commands:

```powershell
go run ./cmd/aepselfhost effect-field-inventory -root data\samples -out tmp\effect_field_inventory\inventory.json
go run ./cmd/aepselfhost effect-field-understanding -inventory tmp\effect_field_inventory\inventory.json -out tmp\effect_field_understanding\understanding.json
```

## Output Contract

When `aepselfhost technique-report` renders report artifacts, it should also
produce effect-field summary artifacts when an understanding JSON is available:

```text
<report-out>/effect_field_summary.json
<report-out>/effect_field_study_queue.csv
```

The JSON summary is the machine-readable report surface:

```json
{
  "schema_version": 1,
  "source_understanding": "tmp/effect_field_understanding/understanding.json",
  "summary": {
    "effect_kinds": 159,
    "effect_occurrences": 3595,
    "param_kinds": 7248,
    "param_occurrences": 45215,
    "reproducibility": [],
    "generation_policies": [],
    "study_actions": []
  },
  "top_study_targets": [
    {
      "match_name": "tc Particular",
      "class": "third_party",
      "study_priority": 2251,
      "reproducibility": "third_party_plugin_required",
      "generation_policy": "preserve_as_dependency",
      "boundary": "Rendering equivalence requires the plugin unless a separate native approximation recipe is authored."
    }
  ]
}
```

The CSV is a compact review queue sorted by study priority.

## Implementation Rules

- Prefer Go.
- Do not rerun sample parsing from inside the report renderer.
- If `tmp/effect_field_understanding/understanding.json` is missing, report
  generation must still succeed and omit these two artifacts.
- If the file exists but is malformed, report generation should fail; a stale or
  corrupt machine input should not silently produce a report.
- Do not claim third-party render equivalence without plugins.
- Do not mutate `technique-facts-v1` or `technique-portrait-v1` for this package.

## Expected Files

- Modify `internal/selfhost/technique_report.go` or a focused helper under
  `internal/selfhost`.
- Modify `internal/selfhost/report_verify.go` only if the manifest/verify layer
  needs to recognize optional artifacts.
- Add focused tests under `internal/selfhost`.
- CLI changes are not required unless the existing report command needs a flag
  to locate the understanding JSON.

## Acceptance

The package is complete when:

- Report rendering writes `effect_field_summary.json` and
  `effect_field_study_queue.csv` when understanding JSON is present.
- Report rendering still succeeds without those artifacts when understanding
  JSON is absent.
- `VerifyTechniqueReport` validates the generated artifacts when they are
  listed in the manifest.
- Tests pass:

```powershell
go test ./internal/selfhost -run "EffectField|TechniqueReport|VerifyTechniqueReport" -count=1
go test ./internal/selfhost ./cmd/aepselfhost -count=1
```

Broader verification:

```powershell
go test ./internal/serializer ./internal/aep_test ./internal/selfhost ./cmd/aepselfhost -count=1
git diff --check
```

## Non-Goals

- No Markdown report as source of truth.
- No UI/dashboard changes.
- No third-party approximation authoring.
- No pseudo rebuild implementation.
