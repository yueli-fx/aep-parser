# Pseudo Behavior Wiring Proof v1

## Goal

Turn `pseudo-controller-rebuild-proof-v1` control plans into a machine-readable
behavior wiring plan for pseudo/controller families. This package records how
keyframed/expression-bearing controls should be handled after controller
structure rebuild, without claiming full rig behavior reconstruction.

## Inputs

Primary input:

```text
tmp/pseudo_controller_rebuild/proof.json
```

Regenerate with:

```powershell
go run ./cmd/aepselfhost pseudo-controller-rebuild-proof -inventory tmp\effect_field_inventory\inventory.json -understanding tmp\effect_field_understanding\understanding.json -out tmp\pseudo_controller_rebuild -max 8
```

## Output Contract

Add a Go selfhost proof command that writes:

```text
tmp/pseudo_behavior_wiring/plan.json
```

`plan.json` must include:

- source proof path
- family/control totals
- a per-family wiring plan
- wiring tasks derived from behavior notes:
  - `keyframed` -> `extract_keyframes_then_apply`
  - `expression` -> `extract_expression_then_apply`
  - both -> keyframes first, expression second, with explicit conflict note
- controls without behavior notes listed as `static_control_preserved`
- explicit boundaries:
  - controller structure is rebuildable
  - keyframe/expression data is not reconstructed by this package
  - exact behavior requires later extraction from source project facts

## Acceptance

The package is complete when:

- The command creates `tmp/pseudo_behavior_wiring/plan.json`.
- Real proof data reports non-zero static controls and non-zero keyframe wiring
  tasks for the generated `Pseudo/YanKFB2` family.
- Unit tests cover keyframe/expression/static task derivation and malformed
  input handling.
- CLI tests cover command wiring.
- Verification passes:

```powershell
go test ./internal/selfhost ./cmd/aepselfhost -run "PseudoBehavior|pseudo-behavior" -count=1
go test ./internal/selfhost ./cmd/aepselfhost -count=1
go test ./internal/serializer ./internal/aep_test ./internal/selfhost ./cmd/aepselfhost -count=1
git diff --check
```

## Non-Goals

- No expression parser.
- No keyframe synthesis or application to the generated AEP.
- No AE host run.
- No Markdown report as source of truth.
