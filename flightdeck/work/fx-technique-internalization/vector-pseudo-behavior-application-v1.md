# Vector Pseudo Behavior Application v1

## Purpose

Extend pseudo behavior payload application from scalar-only keyframes to
2/3/4-component vector keyframes, using the existing Go AEP vector animation
API instead of treating vector payloads as unsupported errors.

This is a behavior application slice inside the pseudo-controller rebuild arc.
It is not a promise of full pseudo effect behavior reconstruction: expression
payloads and plugin-specific runtime behavior remain explicit boundaries.

## Inputs

- `tmp/pseudo_controller_rebuild/proof.json`
- `tmp/pseudo_behavior_payloads/payloads.json`

If those files are missing or stale, regenerate the upstream JSON chain before
running application.

## Implementation Scope

- Detect scalar payload keyframes and keep the existing scalar path.
- Detect 2/3/4-component numeric vector payload keyframes.
- Apply vector payloads with `aep.AnimateEffectParamVec`.
- Reject mixed-width vectors and non-numeric payload values with a per-control
  JSON error row.
- Preserve expression payloads as `expression_deferred`.
- Keep real output in `tmp/pseudo_behavior_application/application.json`.

## Verification

Required commands:

```powershell
go test ./internal/selfhost ./cmd/aepselfhost -run "PseudoBehaviorApplication|pseudo-behavior-application" -count=1
go run ./cmd/aepselfhost pseudo-behavior-application -proof tmp\pseudo_controller_rebuild\proof.json -payloads tmp\pseudo_behavior_payloads\payloads.json -out tmp\pseudo_behavior_application
go test ./internal/selfhost ./cmd/aepselfhost -count=1
go test ./internal/serializer ./internal/aep_test ./internal/selfhost ./cmd/aepselfhost -count=1
git diff --check
```

## Acceptance

- A synthetic vector-keyframe proof applies and verifies at least one vector
  pseudo control through `AnimateEffectParamVec`.
- Current real sample payloads still apply without regression.
- `application.json` remains the machine-readable outcome ledger.
- No intermediate commit is made unless the user explicitly asks.
