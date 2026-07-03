# FX Technique Internalization Goal

Use this file as the direct goal target for the next AE understanding work.

Suggested user command:

```text
/goal execute flightdeck/work/fx-technique-internalization/goal.md
```

## Objective

Status: completed.

Execute `glitch-showcase-gate-v1.md` as one coherent package. The goal is to
validate the stock-AE glitch recipe through the existing from-scratch showcase
generator and render gate without drifting into single-field validation.

Do not interpret one parameter, one field, or one sample as the goal. This is a
family-level behavior application proof and must not claim full rig behavior
reconstruction.

## Required Reading

Before source changes, read:

1. `flightdeck/work/fx-technique-internalization/index.md`
2. `flightdeck/work/fx-technique-internalization/design.md`
3. `flightdeck/work/fx-technique-internalization/glitch-showcase-gate-v1.md`
4. `flightdeck/work/fx-technique-internalization/glitch-reference-phenomenon-v1.md`
5. `flightdeck/work/fx-technique-internalization/expression-pseudo-behavior-boundary-v1.md`
6. `flightdeck/work/fx-technique-internalization/vector-pseudo-behavior-application-v1.md`
7. `flightdeck/work/fx-technique-internalization/multi-family-pseudo-application-proof-v1.md`
8. `flightdeck/work/fx-technique-internalization/pseudo-behavior-payload-application-proof-v1.md`
9. `flightdeck/work/fx-technique-internalization/sample-fact-behavior-extraction-v1.md`
10. `flightdeck/work/fx-technique-internalization/pseudo-behavior-wiring-proof-v1.md`
11. `flightdeck/work/fx-technique-internalization/pseudo-controller-rebuild-proof-v1.md`
12. `flightdeck/work/fx-technique-internalization/effect-field-understanding-v1.md`
13. `flightdeck/work/fx-technique-internalization/technique-facts-v1.md`
14. `flightdeck/work/fx-technique-internalization/technique-portrait-v1.md`
15. `flightdeck/knowledge/techniques/fx-techniques.md`
16. `flightdeck/knowledge/techniques/understand-a-project.md`

If context was compacted or resumed, first reload the Flightdeck preflight
skill/protocol and reread the project Flightdeck state before continuing.

## Execution Contract

Run this as a continuous package:

1. Reconcile workspace and keep existing uncommitted user/agent changes.
2. Ensure source JSON inputs exist; regenerate upstream if missing or stale:

   ```powershell
   go run ./cmd/aepselfhost effect-field-inventory -root data\samples -out tmp\effect_field_inventory\inventory.json
   go run ./cmd/aepselfhost effect-field-understanding -inventory tmp\effect_field_inventory\inventory.json -out tmp\effect_field_understanding\understanding.json
   go run ./cmd/aepselfhost pseudo-controller-rebuild-proof -inventory tmp\effect_field_inventory\inventory.json -understanding tmp\effect_field_understanding\understanding.json -out tmp\pseudo_controller_rebuild -max 8
   go run ./cmd/aepselfhost pseudo-behavior-wiring-plan -proof tmp\pseudo_controller_rebuild\proof.json -out tmp\pseudo_behavior_wiring
   go run ./cmd/aepselfhost pseudo-behavior-payload-extract -plan tmp\pseudo_behavior_wiring\plan.json -root data\samples -out tmp\pseudo_behavior_payloads
   ```

3. Regenerate `flightdeck/showcase/glitch/glitch.aep`.
4. Generate technique explain JSON and inspect dependencies.
5. Run AE render and record package summary.
6. Run focused verification:

   ```powershell
   go test ./internal/selfhost ./cmd/aepselfhost -run "PseudoBehaviorApplication|pseudo-behavior-application" -count=1
   go test ./internal/selfhost ./cmd/aepselfhost -count=1
   ```

7. Run broader verification before final report:

   ```powershell
   go test ./internal/serializer ./internal/aep_test ./internal/selfhost ./cmd/aepselfhost -count=1
   git diff --check
   ```

8. Report generated artifact paths, summary totals, verification commands, and
   remaining boundaries.

## Output Truth

Generated machine-readable outputs for the current glitch showcase package:

- `tmp/technique_showcase_glitch/glitch_explain.json`
- `tmp/technique_showcase_glitch/package_summary.json`
- `flightdeck/showcase/glitch/glitch.aep`
- `flightdeck/showcase/glitch/glitch.png`

Do not create a long Markdown progress report for the run. If a human-readable
view is needed later, generate it from the JSON.

## Latest Execution

Completed `glitch-showcase-gate-v1` as a single package.

Implementation:

- Regenerated `flightdeck/showcase/glitch/glitch.aep` from pure Go.
- Generated `tmp/technique_showcase_glitch/glitch_explain.json`.
- Updated `flightdeck/showcase/glitch/glitch.png` through AE2025 and AE2020
  render gates.
- Recorded the compact gate result in
  `tmp/technique_showcase_glitch/package_summary.json`.

Generated outputs:

- `flightdeck/showcase/glitch/glitch.aep` — 2 comps, 8 layers, 9 effects.
- `tmp/technique_showcase_glitch/glitch_explain.json` — `analysis_ready`, no
  parser unknowns or third-party effects.
- `flightdeck/showcase/glitch/glitch.png` — updated render with visible RGB
  split, glow, dark background, and horizontal tear distortion.
- `tmp/technique_showcase_glitch/package_summary.json` — pass result with
  AE2025/AE2020 render attempts and boundary ledger.

Verification:

- `go run ./flightdeck/showcase/glitch`
- `go test ./flightdeck/showcase/glitch -count=1`
- `go run ./cmd/aeptechnique -in flightdeck\showcase\glitch\glitch.aep -mode explain -out tmp\technique_showcase_glitch\glitch_explain.json`
- `go run ./cmd/aepdissect flightdeck\showcase\glitch\glitch.aep`
- `pwsh -NoProfile -File scripts\ae-worker\ae_run.ps1 -AeExe "E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe" -Jsx <absolute render.jsx> -Done <absolute glitch.done> -TimeoutSec 300`
- `pwsh -NoProfile -File scripts\ae-worker\ae_run.ps1 -AeExe "E:\adobe\Adobe After Effects 2020\Support Files\AfterFX.exe" -Jsx <absolute render.jsx> -Done <absolute glitch.done> -TimeoutSec 300`
- `go test ./cmd/aeptechnique ./flightdeck/showcase/glitch -count=1`
- `git diff --check` exits 0; output contains only CRLF normalization warnings.

Notes:

- Initial AE render attempts used relative `-Jsx` / `-Done` paths and looked
  like startup timeouts. Retrying with absolute paths passed on both AE2025 and
  AE2020.
- Showcase status remains `待review`; agent render review does not mark it
  `complete`.

Prior execution:

Completed `glitch-reference-phenomenon-v1` as a single package.

Implementation:

- Generated corpus and per-project `aeptechnique -mode explain` JSON artifacts
  for Motionbox Booyah and GlitchText.
- Added `tmp/technique_reference_glitch/package_summary.json` as the compact
  machine-readable package outcome.
- Added `flightdeck/knowledge/techniques/build-good-glitch.md` as the reusable
  phenomenon recipe.
- Updated `fx-techniques.md` so the pure native glitch row points to
  `build-good-glitch.md`.

Generated outputs:

- `tmp/technique_reference_glitch/glitch_summary.json` — 2 projects, 0 errors,
  readiness counts `analysis_ready: 1` and `needs_plugins: 1`.
- `tmp/technique_reference_glitch/booyah_explain.json` — stock-AE native
  glitch sample: 12 comps, 61 layers, 42 effects, 110 dependency edges.
- `tmp/technique_reference_glitch/glitchtext_explain.json` — plugin-heavy
  glitch sample: 7 comps, 35 layers, 68 effects, 129 dependency edges.
- `tmp/technique_reference_glitch/package_summary.json` — compact package
  summary and boundary ledger.

Verification:

- `go run ./cmd/aeptechnique -in data\samples\motionbox\glitch -corpus -recursive -mode explain -out tmp\technique_reference_glitch\glitch_explain.jsonl -summary -summary-out tmp\technique_reference_glitch\glitch_summary.json`
- `go run ./cmd/aeptechnique -in "data\samples\motionbox\glitch\booyah-glitch\Booyah Glitch.aep" -mode explain -out tmp\technique_reference_glitch\booyah_explain.json`
- `go run ./cmd/aeptechnique -in "data\samples\motionbox\glitch\glitchtext\GlitchText_è«É¼.aep" -mode explain -out tmp\technique_reference_glitch\glitchtext_explain.json`
- `go test ./internal/technique ./cmd/aeptechnique -count=1`
- `git diff --check` exits 0; output contains only CRLF normalization warnings.

Prior execution:

Completed `expression-pseudo-behavior-boundary-v1` as a single package.

Implementation:

- Added application JSON fields for deferred expression payloads:
  `expression_deferred`, `deferred_expression`,
  `deferred_expression_enabled`, and `expression_boundary`.
- Expression-only payloads now produce `status: expression_deferred`.
- Keyframe+expression payloads can still apply keyframes while recording the
  expression as deferred in the same control row.
- No expression parser, evaluator, or AE expression injection was added.

Generated outputs:

- `tmp/pseudo_behavior_application/application.json` — real sample run remains
  8 families, 8 generated families, 5 payload-missing skipped families, 3
  generated application AEPs, 12 applied controls, 12 verified keyframed
  controls, 0 expression-deferred controls, 0 errors.

Verification:

- `go test ./internal/selfhost ./cmd/aepselfhost -run "PseudoBehaviorApplication|pseudo-behavior-application" -count=1`
- `go run ./cmd/aepselfhost pseudo-behavior-application -proof tmp\pseudo_controller_rebuild\proof.json -payloads tmp\pseudo_behavior_payloads\payloads.json -out tmp\pseudo_behavior_application`
- `go test ./internal/selfhost ./cmd/aepselfhost -count=1`
- `go test ./internal/serializer ./internal/aep_test ./internal/selfhost ./cmd/aepselfhost -count=1`
- `git diff --check` exits 0; output contains only CRLF normalization warnings.

Prior execution:

Completed `vector-pseudo-behavior-application-v1` as a single package.

Implementation:

- Added vector payload conversion for 2/3/4-component numeric keyframe values.
- Applied vector payloads through `aep.AnimateEffectParamVec`.
- Kept scalar payload application on the existing `aep.AnimateEffectParam`
  path.
- Kept expression payloads deferred and malformed payloads as per-control JSON
  errors.

Generated outputs:

- `tmp/pseudo_behavior_application/application.json` — real sample run remains
  8 families, 8 generated families, 5 payload-missing skipped families, 3
  generated application AEPs, 12 applied controls, 12 verified keyframed
  controls, 0 expression-deferred controls, 0 errors.
- `tmp/pseudo_behavior_application/generated/*.aep` — generated application
  proof AEPs for `Pseudo/YanKFB2`, `Pseudo/YanKRA3`, and
  `Pseudo/0.9527614069952335`.

Verification:

- `go test ./internal/selfhost ./cmd/aepselfhost -run "PseudoBehaviorApplication|pseudo-behavior-application" -count=1`
- `go run ./cmd/aepselfhost pseudo-behavior-application -proof tmp\pseudo_controller_rebuild\proof.json -payloads tmp\pseudo_behavior_payloads\payloads.json -out tmp\pseudo_behavior_application`
- `go test ./internal/selfhost ./cmd/aepselfhost -count=1`
- `go test ./internal/serializer ./internal/aep_test ./internal/selfhost ./cmd/aepselfhost -count=1`
- `git diff --check` exits 0; output contains only CRLF normalization warnings.

Prior execution:

Completed `multi-family-pseudo-application-proof-v1` as a single package.

Generated inputs and output used:

- `tmp/effect_field_inventory/inventory.json` — 90 projects, 0 failed, 159
  effect kinds, 7248 param kinds.
- `tmp/effect_field_understanding/understanding.json` — same source totals.
- `tmp/pseudo_controller_rebuild/proof.json` — 8 pseudo families selected, 1
  8 generated families, 119 generated controls, 0 unsupported controls.
- `tmp/pseudo_controller_rebuild/generated/*.aep` — 8 generated controller
  proof AEPs.
- `tmp/pseudo_behavior_wiring/plan.json` — 8 families, 119 controls, 107
  static controls, 12 keyframe tasks, 0 expression tasks.
- `tmp/pseudo_behavior_payloads/payloads.json` — 8 planned families, 119
  planned controls, 12 behavior tasks, 90 scanned projects, 0 failed projects,
  12 controls with payloads, 0 missing controls, 12 keyframe payloads, 0
  expression payloads.
- `tmp/pseudo_behavior_application/application.json` — 8 families, 8 generated
  families, 5 payload-missing skipped families, 3 generated application AEPs,
  12 applied controls, 12 verified keyframed controls, 0 expression-deferred
  controls, 0 errors.
- `tmp/pseudo_behavior_application/generated/*.aep` — generated application
  proof AEPs for `Pseudo/YanKFB2`, `Pseudo/YanKRA3`, and
  `Pseudo/0.9527614069952335`.

Verification:

- `go test ./internal/selfhost ./cmd/aepselfhost -run "PseudoController|PseudoBehaviorApplication|pseudo-controller|pseudo-behavior-application" -count=1`
- `go test ./internal/selfhost ./cmd/aepselfhost -count=1`
- `go test ./internal/serializer ./internal/aep_test ./internal/selfhost ./cmd/aepselfhost -count=1`
- `git diff --check` exits 0; output contains only CRLF normalization warnings.

## Commit Rule

Default for this goal:

```text
no intermediate commits
```

Commit only if the user explicitly asks, or after the whole package is
validated and the user agrees to land the batch.

## Stop Conditions

Stop and report instead of coding when:

- The work starts drifting into one-effect or one-field implementation.
- The source inventory cannot be regenerated.
- The proposed output cannot be represented as JSON.
- Third-party effects are being described as plugin-free render-equivalent.
- Pseudo fields are being described as full behavior reconstruction when only
  source payloads have been applied.
- Markdown logs start replacing JSON state.
- Commits start appearing before final package validation without explicit user
  approval.

## Completion Definition

This goal is complete only when:

- Both Motionbox glitch samples parse through `aeptechnique`.
- Generated JSON artifacts record 2 projects and 0 errors.
- Booyah is recorded as the stock-AE/native glitch learning sample.
- GlitchText is recorded as plugin-required for exact render, not plugin-free.
- `build-good-glitch.md` captures the reusable recipe without replacing JSON
  artifacts as the truth source.
- Focused verification passes, or any failure is reported with exact evidence.
- The final response clearly distinguishes:
  - parseable fields,
  - native exact generation,
  - plugin-required exact render,
  - third-party plugin-required boundaries.
