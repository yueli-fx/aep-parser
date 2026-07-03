# Glitch Showcase Gate v1

## Purpose

Validate the `build-good-glitch.md` stock-AE route with the existing
`flightdeck/showcase/glitch` from-scratch generator.

This package is a generated showcase/gate, not another reference analysis pass.
It upgrades the glitch recipe from "distilled from samples" toward "executable
stock-AE example" while keeping the review-gate honest: agent render review
lands as `待review`, not `complete`.

## Inputs

- `flightdeck/knowledge/techniques/build-good-glitch.md`
- `flightdeck/showcase/glitch/gen.go`
- `flightdeck/showcase/glitch/render.jsx`

## Outputs

Generated artifacts:

- `flightdeck/showcase/glitch/glitch.aep`
- `flightdeck/showcase/glitch/glitch.png`
- `tmp/technique_showcase_glitch/glitch_explain.json`
- `tmp/technique_showcase_glitch/package_summary.json`

Persistent handoff:

- `flightdeck/showcase/glitch/INDEX.md`
- `flightdeck/showcase/INDEX.md`

## Acceptance

- `go run ./flightdeck/showcase/glitch` regenerates `glitch.aep`.
- `aeptechnique -mode explain` parses the generated AEP with no errors.
- `aepdissect` reports no third-party render dependencies.
- AE render script writes `glitch.png` and `glitch.done`.
- Showcase status remains `待review` until user true-machine review.
- Verification passes:

```powershell
go run ./flightdeck/showcase/glitch
go run ./cmd/aeptechnique -in flightdeck\showcase\glitch\glitch.aep -mode explain -out tmp\technique_showcase_glitch\glitch_explain.json
go run ./cmd/aepdissect flightdeck\showcase\glitch\glitch.aep
pwsh -NoProfile -File scripts\ae-worker\ae_run.ps1 -AeExe "E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe" -Jsx flightdeck\showcase\glitch\render.jsx -Done flightdeck\showcase\glitch\glitch.done -TimeoutSec 180
go test ./cmd/aeptechnique -count=1
git diff --check
```

## Boundaries

- This is a stock-AE route: no Twitch/PEDG/Colorama plugin equivalence.
- `T18 temporal-glitch` is not validated by a single-frame showcase.
- User review is still required before marking showcase `complete`.
