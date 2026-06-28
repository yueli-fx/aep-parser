# Phase 5 Non-Booyah Acceptance Run

## Inputs

- Primary project: `samples/motionbox/motion-graphics/circle-animation/circle animation.aep`
- Cross-diff project: `samples/motionbox/motion-graphics/seabox/seabox.aep`
- Reports:
  - `tmp_debug/aepslices/circle_animation_plan.json`
  - `tmp_debug/aepslices/circle_animation_self.json`
  - `tmp_debug/aepslices/circle_animation_vs_seabox.json`
  - `tmp_debug/aeoracle/circle_animation/request.json`

## Commands

```powershell
go run ./cmd/aepslices plan -aep 'samples\motionbox\motion-graphics\circle-animation\circle animation.aep' -json -out tmp_debug\aepslices\circle_animation_plan.json
go run ./cmd/aepslices diagnose -expected 'samples\motionbox\motion-graphics\circle-animation\circle animation.aep' -actual 'samples\motionbox\motion-graphics\circle-animation\circle animation.aep' -json -out tmp_debug\aepslices\circle_animation_self.json
go run ./cmd/aepslices diagnose -expected 'samples\motionbox\motion-graphics\circle-animation\circle animation.aep' -actual 'samples\motionbox\motion-graphics\seabox\seabox.aep' -json -out tmp_debug\aepslices\circle_animation_vs_seabox.json
go run ./cmd/aeoracle plan -aep 'samples\motionbox\motion-graphics\circle-animation\circle animation.aep' -out tmp_debug\aeoracle\circle_animation -json
go run ./cmd/aeoracle render -request tmp_debug\aeoracle\circle_animation\request.json -dry-run
```

## Result

The generic Phase 5 workflow handles a non-Booyah Motionbox project without
bespoke project code:

- Profile/plan: 6 comps, 92 layers, 10 footage items.
- Selected representative slices:
  - comp 283 `Structure 02[ywugqu]`: footage-assembly, score 510, 45 layers, 43 shape layers, 1 footage layer, 20 effects.
  - comp 1 `コンポ 1`: footage-assembly, score 222, 32 layers, 26 shape layers, 3 footage layers, 1 effect.
  - comp 250 `Tiled Background 05[xhcbdf]`: procedural, score 55, 5 layers, 4 shape layers, 3 effects.
- Self-diagnose returned exit 0 and no gaps.
- Cross-diagnose against `seabox.aep` returned exit 1 and 101 structured gaps:
  - 73 `write-gap`
  - 27 `investigate-gap`
  - 1 `semantic-gap`
  - 46 `wrong_value`
  - 31 `missing_object`
  - 24 `extra_object`
- Render oracle planning selected 3 sentinel frames: 0, 24, 48.
- Render dry-run validated the request and emitted the `scripts/ae_run.ps1`
  invocation.

## Interpretation

This validates the generic non-Booyah path for profile, slice selection, diff,
gap ledger generation, render-frame planning, and render request validation. It
does not prove pixel fidelity because no AE render was launched in this run.

Actual AE render was intentionally not started because an existing AfterFX
process was present (`AfterFX`, pid 57036, AE 2020 untitled project). The
tracked `scripts/ae_run.ps1` guard refuses concurrent AE sessions by default to
avoid reading the wrong modal windows or disrupting user state. Close or clear
that AE process before running the hard render gate.

If Phase 6 needs a hard render-readiness gate, run `cmd/aeoracle render` against
the generated request and compare the output images before starting recipe
correction loops.
