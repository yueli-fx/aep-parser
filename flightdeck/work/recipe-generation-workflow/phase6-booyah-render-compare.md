# Phase 6 Booyah Render Compare

## Inputs

- Original: `data/samples/motionbox/glitch/booyah-glitch/Booyah Glitch.aep`
- Current clone: `flightdeck/showcase/booyah-clone/booyah-clone.aep`
- Compared comp: `グリッチテキスト`
- AE executable: `E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe`
- AE version reported by render metadata: `25.1x68`

## Commands

```powershell
go run ./cmd/aeoracle plan -aep 'data\samples\motionbox\glitch\booyah-glitch\Booyah Glitch.aep' -comp 'グリッチテキスト' -out tmp_debug\aeoracle\booyah_compare\source -json
go run ./cmd/aeoracle clone-request -from tmp_debug\aeoracle\booyah_compare\source\request.json -aep flightdeck\showcase\booyah-clone\booyah-clone.aep -out tmp_debug\aeoracle\booyah_compare\clone -json
go run ./cmd/aeoracle render -request tmp_debug\aeoracle\booyah_compare\source\request.json -ae 'E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe' -timeout-sec 900
go run ./cmd/aeoracle render -request tmp_debug\aeoracle\booyah_compare\clone\request.json -ae 'E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe' -timeout-sec 900
go run ./cmd/aeoracle compare-set -expected-meta tmp_debug\aeoracle\booyah_compare\source\metadata.json -actual-meta tmp_debug\aeoracle\booyah_compare\clone\metadata.json -threshold 0 -json -out tmp_debug\aeoracle\booyah_compare\compare_set.json
go run ./cmd/aepslices diagnose -expected 'data\samples\motionbox\glitch\booyah-glitch\Booyah Glitch.aep' -actual 'flightdeck\showcase\booyah-clone\booyah-clone.aep' -render-set tmp_debug\aeoracle\booyah_compare\compare_set.json -json -out tmp_debug\aepslices\booyah_vs_clone_with_render.json
```

`compare-set` and final `diagnose` return exit 1 because they produced valid
reports with fidelity gaps.

## Render Result

Both source and clone rendered successfully:

- Source metadata: `tmp_debug/aeoracle/booyah_compare/source/metadata.json`,
  status `ok`, comp `グリッチテキスト`, frames `8`.
- Clone metadata: `tmp_debug/aeoracle/booyah_compare/clone/metadata.json`,
  status `ok`, comp `グリッチテキスト`, frames `8`.
- Frame-set report: `tmp_debug/aeoracle/booyah_compare/compare_set.json`.

Frame-set summary:

- Total frames: 8
- OK frames: 2 (`f000000`, `f000180`)
- Different frames: 6
- Missing expected frames: 0
- Missing actual frames: 0

Different frames:

| Tag | Frame | Seconds | Different pixels | Different % | Max channel delta |
| --- | ---: | ---: | ---: | ---: | ---: |
| `f000001` | 1 | 0.033367 | 1548911 | 74.6967 | 255 |
| `f000003` | 3 | 0.100100 | 1598870 | 77.1060 | 255 |
| `f000004` | 4 | 0.133467 | 980023 | 47.2619 | 236 |
| `f000005` | 5 | 0.166834 | 1349221 | 65.0666 | 255 |
| `f000006` | 6 | 0.200200 | 431848 | 20.8260 | 255 |
| `f000007` | 7 | 0.233567 | 1231208 | 59.3754 | 255 |

## Combined Diagnose Result

Combined report: `tmp_debug/aepslices/booyah_vs_clone_with_render.json`

- Total gaps: 147
- Profile gap report: 141 gaps
- Render frame-set gap report: 6 gaps
- Render gap type: `semantic-gap`
- Render evidence: `L4_render`, source `internal/aeoracle`

## Interpretation

The generic Phase 6 render-compare loop is now proven on the real Booyah source
and current clone:

- A single source sentinel plan can be cloned for a generated AEP.
- Source and clone can be rendered through the same `cmd/aeoracle render`
  command.
- `cmd/aeoracle compare-set` can produce a multi-frame pixel report.
- `cmd/aepslices diagnose -render-set` can merge profile gaps and render gaps.

The current clone still has major visual deltas in early frames. That is
expected from the Phase 5 profile gaps: many layer timing and source-reference
differences remain. Phase 6 can now proceed to recipe IR planning because there
is a measured source-vs-clone render objective, but automated correction loops
should remain blocked until their parameter space is explicitly bounded.
