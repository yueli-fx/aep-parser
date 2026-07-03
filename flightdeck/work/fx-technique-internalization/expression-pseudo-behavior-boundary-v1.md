# Expression Pseudo Behavior Boundary v1

## Purpose

Make expression payload handling explicit in pseudo behavior application JSON.
Expression text may be extracted as a source fact, but this package must not
claim expression runtime reconstruction.

## Scope

- Preserve extracted expression text and enabled state in the application
  outcome as a deferred payload.
- Count expression-only payloads as deferred.
- Count keyframe+expression conflict payloads as both keyframe-applied and
  expression-deferred.
- Keep generated AEP output focused on keyframe application only.
- Keep real sample results honest: current `data/samples` wiring has 0
  expression tasks and 0 expression payloads.

## Non-Goals

- No expression parser or evaluator.
- No AE expression injection.
- No plugin/runtime behavior reconstruction.
- No Markdown report as source of truth.

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

- Synthetic expression-only payload is written as `expression_deferred`.
- Synthetic keyframe+expression payload applies keyframes and records the
  expression as deferred in the same control row.
- Real application run remains 12 applied, 12 verified, 0 expression-deferred,
  0 errors until real expression payloads appear in `data/samples`.
