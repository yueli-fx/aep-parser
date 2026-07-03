# Pseudo Behavior Payload Application Proof v1

## Goal

Apply extracted pseudo behavior payloads from
`tmp/pseudo_behavior_payloads/payloads.json` to generated pseudo controls and
prove the applied AEP round-trips through the Go parser.

This is a generated-family proof, not a full rig reconstruction. It wires
source keyframe facts onto the generated pseudo control properties that already
exist in the pseudo-controller proof.

## Inputs

Primary inputs:

```text
tmp/pseudo_controller_rebuild/proof.json
tmp/pseudo_behavior_payloads/payloads.json
```

## Output Contract

Add a Go selfhost command that writes:

```text
tmp/pseudo_behavior_application/application.json
tmp/pseudo_behavior_application/generated/<family>.aep
```

The JSON must include:

- source proof path and payload path
- per-family application status
- per-control applied/skipped/error status
- generated AEP path for families that can be written
- round-trip verification counts from reparsed profiles
- boundaries:
  - only generated pseudo families are applied
  - keyframe application is supported for scalar and vector controls via
    existing effect-param animation APIs
  - expression payload application is deferred
  - this does not prove full render-equivalent rig behavior

## Acceptance

The package is complete when:

- The command creates `tmp/pseudo_behavior_application/application.json`.
- The generated top pseudo family AEP reparses with the expected keyframed
  controls.
- Unit tests cover keyframe application planning, skipped non-generated
  families, malformed input, and CLI wiring.
- Verification passes:

```powershell
go test ./internal/selfhost ./cmd/aepselfhost -run "PseudoBehaviorApplication|pseudo-behavior-application" -count=1
go test ./internal/selfhost ./cmd/aepselfhost -count=1
go test ./internal/serializer ./internal/aep_test ./internal/selfhost ./cmd/aepselfhost -count=1
git diff --check
```

## Non-Goals

- No expression parser/evaluator/application.
- No third-party effect render equivalence.
- No AE host run.
- No Markdown report as source of truth.
