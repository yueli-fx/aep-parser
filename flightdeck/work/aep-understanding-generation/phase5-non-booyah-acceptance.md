# Phase 5 Non-Booyah Acceptance Run

## Inputs

- Primary project: `data/samples/motionbox/motion-graphics/circle-animation/circle animation.aep`
- Cross-diff project: `data/samples/motionbox/motion-graphics/seabox/seabox.aep`
- Reports:
  - `tmp_debug/aepslices/circle_animation_plan.json`
  - `tmp_debug/aepslices/circle_animation_self.json`
  - `tmp_debug/aepslices/circle_animation_vs_seabox.json`
  - `tmp_debug/aeoracle/circle_animation/request.json`

## Commands

```powershell
go run ./cmd/aepslices plan -aep 'data\samples\motionbox\motion-graphics\circle-animation\circle animation.aep' -json -out tmp_debug\aepslices\circle_animation_plan.json
go run ./cmd/aepslices diagnose -expected 'data\samples\motionbox\motion-graphics\circle-animation\circle animation.aep' -actual 'data\samples\motionbox\motion-graphics\circle-animation\circle animation.aep' -json -out tmp_debug\aepslices\circle_animation_self.json
go run ./cmd/aepslices diagnose -expected 'data\samples\motionbox\motion-graphics\circle-animation\circle animation.aep' -actual 'data\samples\motionbox\motion-graphics\seabox\seabox.aep' -json -out tmp_debug\aepslices\circle_animation_vs_seabox.json
go run ./cmd/aeoracle plan -aep 'data\samples\motionbox\motion-graphics\circle-animation\circle animation.aep' -out tmp_debug\aeoracle\circle_animation -json
go run ./cmd/aeoracle render -request tmp_debug\aeoracle\circle_animation\request.json -dry-run
go run ./cmd/aeoracle render -request tmp_debug\aeoracle\circle_animation\request.json -ae 'E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe' -timeout-sec 600
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
  invocation with absolute runner, JSX, and `.done` paths.
- Hard AE render gate passed on After Effects `25.1x68`:
  - `.done`: `tmp_debug/aeoracle/circle_animation/aeoracle_render.done` = `ok`
  - metadata: `tmp_debug/aeoracle/circle_animation/metadata.json`, status `ok`
  - PNG outputs: `f000000.png`, `f000024.png`, `f000048.png`
- Render gate fix recorded in code:
  - `scripts/aeoracle_render.jsx` injects a small JSON parser/stringifier because
    this AE ExtendScript runtime does not provide global `JSON`.
  - `cmd/aeoracle render` resolves request/runner/JSX/done paths against the
    invocation cwd and treats non-`ok` `.done` content as failure.

## Interpretation

This validates the generic non-Booyah path for profile, slice selection, diff,
gap ledger generation, render-frame planning, render request validation, and
hard AE frame extraction.

This still does not prove clone pixel fidelity by itself: the current hard gate
renders sentinel frames from the source project. Phase 6 should pair this with
generated-clone rendering and `cmd/aeoracle compare` before starting automated
recipe correction loops.
