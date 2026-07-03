# Multi-Family Pseudo Application Proof v1

## Goal

Upgrade the pseudo controller/application proof from top-family only to every
selected pseudo family whose controls are supported.

The current real proof has 8 selected pseudo families and 0 unsupported
controls, but only the first family is generated because the proof runner stops
after top family generation. This package removes that artificial limit and
lets the application proof apply extracted keyframes across all generated
families.

## Inputs

```text
tmp/effect_field_inventory/inventory.json
tmp/effect_field_understanding/understanding.json
tmp/pseudo_behavior_payloads/payloads.json
```

## Output Contract

Regenerate and preserve these JSON/AEP outputs:

```text
tmp/pseudo_controller_rebuild/proof.json
tmp/pseudo_controller_rebuild/generated/*.aep
tmp/pseudo_behavior_wiring/plan.json
tmp/pseudo_behavior_payloads/payloads.json
tmp/pseudo_behavior_application/application.json
tmp/pseudo_behavior_application/generated/*.aep
```

The JSON must show:

- all selected supported pseudo families marked `generated`
- one generated controller AEP per generated family
- application proof attempts every generated family that has payloads
- skipped families are only unsupported or payload-missing, not top-family
  policy misses

## Acceptance

- Unit tests prove `RunPseudoControllerRebuildProof` generates every selected
  supported family, not only index 0.
- Unit tests prove application handles multiple generated families.
- Real run increases generated controller families from 1 to 8.
- Real application run applies all currently extractable scalar keyframe
  payload controls, or reports exact per-control non-scalar/expression
  boundaries.
- Verification passes:

```powershell
go test ./internal/selfhost ./cmd/aepselfhost -run "PseudoController|PseudoBehaviorApplication|pseudo-controller|pseudo-behavior-application" -count=1
go test ./internal/selfhost ./cmd/aepselfhost -count=1
go test ./internal/serializer ./internal/aep_test ./internal/selfhost ./cmd/aepselfhost -count=1
git diff --check
```

## Non-Goals

- No AE host run.
- No third-party render equivalence.
- No expression application.
- No Markdown report as source of truth.
