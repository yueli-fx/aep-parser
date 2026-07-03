# Sample Fact Behavior Extraction v1

## Goal

Recover concrete keyframe/expression payload examples from source sample
profiles for the pseudo controls listed in
`tmp/pseudo_behavior_wiring/plan.json`.

This package extracts behavior facts only. It does not apply those payloads to
the generated pseudo controller AEP and does not claim full rig behavior
reconstruction.

## Inputs

Primary inputs:

```text
tmp/pseudo_behavior_wiring/plan.json
data/samples
```

Regenerate upstream inputs with:

```powershell
go run ./cmd/aepselfhost pseudo-controller-rebuild-proof -inventory tmp\effect_field_inventory\inventory.json -understanding tmp\effect_field_understanding\understanding.json -out tmp\pseudo_controller_rebuild -max 8
go run ./cmd/aepselfhost pseudo-behavior-wiring-plan -proof tmp\pseudo_controller_rebuild\proof.json -out tmp\pseudo_behavior_wiring
```

## Output Contract

Add a Go selfhost command that writes:

```text
tmp/pseudo_behavior_payloads/payloads.json
```

The JSON must include:

- source plan path and sample root
- scan totals
- per-family/per-control extraction status
- concrete examples for controls with behavior tasks:
  - `keyframes`: keyframe time/value/ease data as parsed by the profile layer
  - `expression`: expression text and enabled state when present
- explicit misses when a wiring task has no payload in scanned samples
- boundaries:
  - extracted payloads are source facts
  - applying them to generated controls is a later package
  - expression semantic correctness is not proven here

## Acceptance

The package is complete when:

- The command creates `tmp/pseudo_behavior_payloads/payloads.json`.
- Real sample data produces non-zero keyframe payload examples for at least the
  top generated pseudo family.
- Unit tests cover keyframe extraction, expression extraction, misses, malformed
  plan input, and CLI wiring.
- Verification passes:

```powershell
go test ./internal/selfhost ./cmd/aepselfhost -run "PseudoBehaviorPayload|pseudo-behavior-payload" -count=1
go test ./internal/selfhost ./cmd/aepselfhost -count=1
go test ./internal/serializer ./internal/aep_test ./internal/selfhost ./cmd/aepselfhost -count=1
git diff --check
```

## Non-Goals

- No behavior application to generated AEP.
- No expression parser or evaluator.
- No AE host run.
- No Markdown report as source of truth.
