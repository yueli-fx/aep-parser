# Phase 5 Booyah Acceptance Run

## Inputs

- Original: `samples/motionbox/glitch/booyah-glitch/Booyah Glitch.aep`
- Current clone: `flightdeck/showcase/booyah-clone/booyah-clone.aep`
- Reports:
  - `tmp_debug/aepslices/booyah_plan.json`
  - `tmp_debug/aepslices/booyah_vs_clone_after_ref_fallback.json`
  - `tmp_debug/aeoracle/booyah/request.json`

## Commands

```powershell
go run ./cmd/aepslices plan -aep 'samples\motionbox\glitch\booyah-glitch\Booyah Glitch.aep' -json -out tmp_debug\aepslices\booyah_plan.json
go run ./cmd/aeoracle plan -aep 'samples\motionbox\glitch\booyah-glitch\Booyah Glitch.aep' -out tmp_debug\aeoracle\booyah -json
go run ./cmd/aepslices diagnose -expected 'samples\motionbox\glitch\booyah-glitch\Booyah Glitch.aep' -actual 'flightdeck\showcase\booyah-clone\booyah-clone.aep' -json -out tmp_debug\aepslices\booyah_vs_clone_after_ref_fallback.json
```

## Result

The generic Phase 5 workflow handles the real Booyah project without bespoke
project code:

- Profile/plan: 12 comps, 61 layers, 14 footage items.
- Selected representative slices:
  - comp 85 `グリッチテキスト`: footage-assembly, score 347, 27 layers, 22 footage layers, 27 effects.
  - comp 43 `背景変えるならココ！`: footage-assembly, score 106, 6 layers, 6 footage layers, 5 effects.
  - comp 69 `マップ用フラクタルノイズ`: footage-assembly, score 72, 2 layers, 2 footage layers, 2 effects.
- Render oracle plan selected 8 sentinel frames: 0, 4, 40, 75, 77, 78, 79, 108.
- Diagnose after layer-index and source-ref fallback produced 141 gaps:
  - 140 `write-gap`
  - 1 `investigate-gap`
  - 140 `wrong_value`
  - 1 `missing_object`

## Diagnostic Quality Fix

The first run produced 123 gaps, but 104 of them were `missing_object` /
`extra_object` pairs. That was a matching-quality problem: from-scratch clone
projects do not preserve source layer IDs, and many generated layer names differ
from the original. `profilediff` now falls back to unique layer index after
path and `index:name` matching fail.

The second run then exposed another matching-quality issue: source item IDs
differ in from-scratch clones even when the referenced comp name and kind are
the same. `profilediff` now compares `source_ref` by `kind + name` when both
names are available, falling back to full object equality only when names are
missing.

After both fixes, the report mostly contains concrete field deltas such as
timing, footage source, text, and effect differences. This is more actionable
than object presence or stable-ID noise.

## Remaining Gate

This validates Phase 5 on the real Booyah source project and the current clone.
It does not replace the original Phase 6 readiness gate that calls for at least
one non-Booyah project to pass through the same workflow before recipe IR work.
