# Rain Showcase Gate v1

## Purpose

Validate the `build-good-rain.md` stock-AE route with a new from-scratch
showcase generator under `flightdeck/showcase/rain`.

This package is a generated showcase/gate, not an exact clone of the Motionbox
Rain Day sample. The reference exact route still depends on Trapcode Particular
and Unmult; this package validates the plugin-free native spine.

## Inputs

- `flightdeck/knowledge/techniques/build-good-rain.md`
- `flightdeck/showcase/rain/gen.go`
- `flightdeck/showcase/rain/render.jsx`

## Outputs

Generated artifacts:

- `flightdeck/showcase/rain/rain.aep`
- `flightdeck/showcase/rain/rain.png`
- `flightdeck/showcase/rain/rain.done`
- `tmp/technique_showcase_rain/rain_explain.json`
- `tmp/technique_showcase_rain/package_summary.json`

Persistent handoff:

- `flightdeck/showcase/rain/INDEX.md`
- `flightdeck/showcase/INDEX.md`
- `flightdeck/knowledge/techniques/build-good-rain.md`
- `flightdeck/knowledge/techniques/fx-techniques.md`

## Acceptance

- `go run ./flightdeck/showcase/rain` regenerates `rain.aep`.
- `go test ./flightdeck/showcase -run TestRainShowcaseGeneratorBuildsNativeRainSpine -count=1` passes after first failing for the missing generator.
- `aeptechnique -mode explain` parses the generated AEP.
- `aepdissect` reports no third-party render dependencies.
- AE2020 render script writes `rain.png` and `rain.done`.
- Agent visual review confirms the generated frame contains readable slanted
  rain streaks and not just noise or endpoint dots.
- Showcase status remains `待review` until user true-machine review.

## Boundaries

- No plugin-free exact render claim for the Motionbox reference.
- No Trapcode Particular or Unmult embed-template support added.
- Shape/repeater rain streaks are validated by the generated showcase, while
  the broader reference recipe remains sample-observed until more rain samples
  or a richer generator are added.

## Latest Execution

Completed as a single showcase package.

Implementation:

- Added `flightdeck/showcase/rain/gen.go`, a pure Go AE2020 generator.
- Added `flightdeck/showcase/rain/render.jsx`, rendering `RAIN` at t=2.0s.
- Added `flightdeck/showcase/rain/INDEX.md` and updated the top-level showcase
  index.
- Added `flightdeck/showcase/rain_showcase_test.go`; RED failed because
  `flightdeck/showcase/rain` did not exist, GREEN passes after implementation.

Generated outputs:

- `flightdeck/showcase/rain/rain.aep` — 2 comps, native effects only.
- `flightdeck/showcase/rain/rain.png` — AE2020 render with visible slanted rain
  streaks, glow/blur trail, and a dark wet-ground band.
- `tmp/technique_showcase_rain/rain_explain.json` — explain artifact for the
  generated showcase.
- `tmp/technique_showcase_rain/package_summary.json` — compact package result.

Verification:

```powershell
go test ./flightdeck/showcase -run TestRainShowcaseGeneratorBuildsNativeRainSpine -count=1
go test ./flightdeck/showcase ./flightdeck/showcase/rain -count=1
go run ./flightdeck/showcase/rain
go run ./cmd/aeptechnique -in flightdeck\showcase\rain\rain.aep -mode explain -out tmp\technique_showcase_rain\rain_explain.json
go run ./cmd/aepdissect flightdeck\showcase\rain\rain.aep
pwsh -NoProfile -File scripts\ae-worker\ae_run.ps1 -AeExe "E:\adobe\Adobe After Effects 2020\Support Files\AfterFX.exe" -Jsx <absolute render.jsx> -Done <absolute rain.done> -TimeoutSec 300
git diff --check
```

