# Effect Field Understanding v1

## Goal

Turn the raw effect-field inventory from real sample projects into a
machine-readable understanding layer that can drive project learning,
generation planning, and honest replication boundaries.

This is the next layer after:

```text
go run ./cmd/aepselfhost effect-field-inventory -root data\samples -out tmp\effect_field_inventory\inventory.json
```

The inventory answers "what exists". This spec answers "what can we do with it".

## Background

Current inventory evidence from `data/samples`:

- 90 projects parsed, 0 failures.
- 159 effect kinds, 3595 effect occurrences.
- 7248 effect-param kinds, 45215 effect-param occurrences.
- Classes by occurrence: `native_supported=3203`, `third_party=260`,
  `pseudo=132`.

Native/bundled sample-shell effect template gaps are currently cleared. The
remaining learning value is not "add one more native template", but classifying
effect fields into reproducible capability, plugin boundary, pseudo-control
intent, and study priority.

## Inputs

Primary input:

- `tmp/effect_field_inventory/inventory.json`

Regenerating command:

```powershell
go run ./cmd/aepselfhost effect-field-inventory -root data\samples -out tmp\effect_field_inventory\inventory.json
```

Secondary context:

- `internal/profile.Profile` effect and param fields.
- `flightdeck/work/fx-technique-internalization/technique-facts-v1.md`
- `flightdeck/work/fx-technique-internalization/technique-portrait-v1.md`
- `flightdeck/knowledge/techniques/fx-techniques.md`
- `flightdeck/knowledge/techniques/understand-a-project.md`

## Output Contract

Create one generated JSON report:

```text
tmp/effect_field_understanding/understanding.json
```

Minimum shape:

```json
{
  "schema_version": 1,
  "source_inventory": "tmp/effect_field_inventory/inventory.json",
  "summary": {
    "effect_kinds": 159,
    "effect_occurrences": 3595,
    "param_kinds": 7248,
    "param_occurrences": 45215,
    "classes": [],
    "reproducibility": [],
    "study_actions": []
  },
  "effects": [
    {
      "match_name": "tc Particular",
      "class": "third_party",
      "occurrences": 12,
      "param_kinds": 1881,
      "capabilities": ["color", "expression", "keyframes", "position", "scalar"],
      "reproducibility": "third_party_plugin_required",
      "generation_policy": "preserve_as_dependency",
      "field_understanding": "param_names_and_values_parseable",
      "study_priority": 100,
      "study_actions": ["summarize_param_groups", "map_controller_inputs"],
      "boundary": "Rendering equivalence requires the plugin unless a separate native approximation recipe is authored."
    }
  ]
}
```

The report must be deterministic and queryable. Markdown can be generated later
from this JSON, but Markdown is not the source of truth.

## Classification Rules

### Effect Class

Use existing class labels from the inventory unless a more specific rule is
needed:

- `native_supported`
- `native_template_gap`
- `pseudo`
- `third_party`
- `unknown_or_alias`

### Reproducibility

Every effect gets exactly one reproducibility label:

- `go_native_exact` — supported native/bundled effect can be emitted by current
  Go writer with known template/API behavior.
- `native_template_possible` — native/bundled effect appears parseable but
  needs a template/API support gap before generation.
- `pseudo_rebuildable` — pseudo effect fields can be understood as controls;
  rebuild may use `BuildPseudoEffect` when the control model is sufficient.
- `third_party_plugin_required` — fields are parseable, but render-equivalent
  output requires the plugin or an authored native approximation.
- `unknown_or_alias_pending` — class or alias needs review before any generation
  claim.

Do not use "unsupported" as a blanket label for third-party or pseudo effects.
They are parseable facts with explicit generation boundaries.

### Generation Policy

Each effect gets one generation policy:

- `emit_native`
- `add_native_template`
- `rebuild_pseudo_controls`
- `preserve_as_dependency`
- `author_native_approximation`
- `classify_or_alias`

The default for third-party effects is `preserve_as_dependency`, not
`author_native_approximation`. Approximation is a separate authored technique,
not an automatic claim.

### Study Priority

Sort study candidates by a stable score:

```text
occurrences * 10
+ param_kinds
+ 100 if class is third_party or pseudo
+ 50 if capabilities include keyframes
+ 50 if capabilities include expression
+ 30 if capabilities include layer_ref
+ 20 if capabilities include color or position
```

This favors high-signal controller/effect systems without letting one huge
plugin parameter table hide all pseudo controllers.

## Required Implementation Shape

Prefer Go.

Expected code boundaries:

- `internal/selfhost/effect_field_understanding.go`
  - read or accept `EffectFieldInventory`
  - build deterministic `EffectFieldUnderstanding`
  - classify reproducibility, generation policy, study actions, and boundaries
- `internal/selfhost/effect_field_understanding_test.go`
  - synthetic inventory tests for native, pseudo, third-party, and unknown
  - study priority ordering test
- `cmd/aepselfhost`
  - add `effect-field-understanding`
  - flags: `-inventory`, `-out`

Expected command:

```powershell
go run ./cmd/aepselfhost effect-field-understanding -inventory tmp\effect_field_inventory\inventory.json -out tmp\effect_field_understanding\understanding.json
```

## Acceptance

The phase is complete when:

- `tmp/effect_field_inventory/inventory.json` can be regenerated from
  `data/samples`.
- `tmp/effect_field_understanding/understanding.json` is generated from the
  inventory.
- The understanding summary includes class, reproducibility, and study-action
  totals.
- Top third-party and pseudo effects appear as parseable study targets with
  explicit render/generation boundaries.
- No claim is made that third-party effects can render without the plugin.
- Tests pass:

```powershell
go test ./internal/selfhost ./cmd/aepselfhost -run "EffectField" -count=1
go test ./internal/selfhost ./cmd/aepselfhost -count=1
```

Recommended broader verification before committing implementation:

```powershell
go test ./internal/serializer ./internal/aep_test ./internal/selfhost ./cmd/aepselfhost -count=1
git diff --check
```

## Non-Goals

- Do not add one effect template at a time.
- Do not attempt plugin-free render equivalence for third-party effects.
- Do not mutate `technique-facts-v1.md` or `technique-portrait-v1.md` unless
  the new understanding output is intentionally wired into those models.
- Do not write long mutable Markdown reports for inventory results.
- Do not commit intermediate validation-only checkpoints unless the user asks.

## Next After v1

After this report exists, the next coherent package is either:

- wire high-level class/reproducibility summaries into technique reports, or
- select one high-priority pseudo/controller family and prove a rebuild path, or
- select one high-priority third-party family and author a native approximation
  as a separate technique package.
