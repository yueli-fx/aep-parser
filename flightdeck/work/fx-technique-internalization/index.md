# Index — fx-technique-internalization

## State

This is the upstream "understand a project" research arc. It defines the reusable pipeline for turning reference `.aep` files into cross-domain FX techniques and phenomenon recipes.

The fire sample proved the mechanism once; the arc is parked while Booyah
review and versioned migration take priority. Technique facts, portraits,
corpus notes, and report specs now live in this package instead of the old
monolithic AEP-understanding package.

Current selected package `rain-showcase-gate-v1` is implemented.
`rain-reference-phenomenon-v1` is implemented.
`glitch-showcase-gate-v1` is implemented.
`glitch-reference-phenomenon-v1` is implemented.
`expression-pseudo-behavior-boundary-v1` is implemented.
`vector-pseudo-behavior-application-v1` is implemented.
`multi-family-pseudo-application-proof-v1` is implemented.
`pseudo-behavior-payload-application-proof-v1` is implemented.
`sample-fact-behavior-extraction-v1` is implemented.
`pseudo-behavior-wiring-proof-v1` is implemented.
`pseudo-controller-rebuild-proof-v1` is implemented.
The prior `effect-field-report-surface-v1` package is implemented. The
raw inventory command and understanding command generate JSON reports under
`tmp/`; technique report rendering now surfaces the understanding summary as
machine-readable report artifacts.

## Next

Recommended next package: another reference phenomenon package, a richer rain
generator upgrade from the native rain showcase, or a focused generator upgrade
from the validated glitch recipe. Do not resume by selecting a single effect or
field.

## Read now

- `design.md` — pipeline, two-layer knowledge structure, sample findings, and next steps
- `goal.md` — direct goal entry for the latest completed package and next package handoff
- `rain-showcase-gate-v1.md` — spec for validating the rain recipe through a from-scratch native showcase
- `rain-reference-phenomenon-v1.md` — spec for processing the Motionbox rain sample into JSON artifacts and a reusable recipe
- `glitch-showcase-gate-v1.md` — spec for validating the glitch recipe through the from-scratch showcase
- `glitch-reference-phenomenon-v1.md` — spec for processing Motionbox glitch samples into JSON artifacts and a reusable recipe
- `expression-pseudo-behavior-boundary-v1.md` — spec for recording expression payloads as explicit deferred application facts
- `vector-pseudo-behavior-application-v1.md` — spec for applying 2/3/4-component pseudo keyframe payloads
- `multi-family-pseudo-application-proof-v1.md` — spec for generating/applying every selected supported pseudo family
- `pseudo-behavior-payload-application-proof-v1.md` — spec for applying extracted keyframe payloads to generated pseudo controls
- `sample-fact-behavior-extraction-v1.md` — spec for extracting concrete keyframe/expression payload examples
- `pseudo-behavior-wiring-proof-v1.md` — spec for turning rebuild proof behavior notes into wiring tasks
- `pseudo-controller-rebuild-proof-v1.md` — spec for mapping pseudo inventory into `BuildPseudoEffect` proof artifacts
- `effect-field-report-surface-v1.md` — spec for exposing understanding JSON through report artifacts
- `effect-field-understanding-v1.md` — spec for classifying inventory rows into reproducibility, generation policy, study actions, and boundaries
- `technique-facts-v1.md` — machine-usable fact model for learned techniques
- `technique-portrait-v1.md` — portrait/report shape for a learned technique
- `technique-portrait-corpus-v1.md` — corpus notes and examples
- `flightdeck/knowledge/techniques/understand-a-project.md` — current reusable checklist for processing each reference `.aep`
- `flightdeck/knowledge/techniques/fx-techniques.md` — current technique library
- `flightdeck/knowledge/techniques/build-good-rain.md` — rain phenomenon recipe distilled from Motionbox Rain Day
- `flightdeck/knowledge/techniques/build-good-glitch.md` — glitch phenomenon recipe distilled from Booyah and GlitchText

## Read if

- `technique-facts-v1-plan.md` and `technique-portrait-v1-plan.md` — when
  changing the fact or portrait schemas.
- `technique-report-dashboard-plan.md` and `technique-report-dashboard-spec.md`
  — when resuming report/dashboard work.
- `flightdeck/knowledge/techniques/add-reference-sample.md` — when adding or classifying a new reference sample
- `flightdeck/knowledge/techniques/build-good-fire.md` — when comparing against the fire validation case

## Progress

Done:
- `rain-showcase-gate-v1` adds a plugin-free native rain showcase under
  `flightdeck/showcase/rain`, with shape/repeater rain streaks, native mist,
  Echo/Blur/Glow trail, AE2020 render output, and a TDD structure test.
- `rain-reference-phenomenon-v1` adds Motionbox Rain Day reference artifacts
  under `tmp/technique_reference_rain`, records the reusable rain recipe in
  `flightdeck/knowledge/techniques/build-good-rain.md`, and links the rain
  phenomenon plus `shape-repeater-rain-streaks` atom in `fx-techniques.md`.
- The five-step internalization pipeline is defined.
- Fire and lightning sample lessons are recorded.
- The two element classes are recognized: procedural elements vs footage plus assembly.
- Technique facts, portrait, corpus, and report-dashboard specs were moved
  here from the old umbrella package.
- `effect-field-inventory` and `effect-field-understanding` generate JSON
  totals for native, pseudo, and third-party effect fields from `data/samples`.
- `effect-field-report-surface-v1` adds `effect_field_summary.json` and
  `effect_field_study_queue.csv` to technique report outputs when the
  understanding JSON exists; verification checks these artifacts when listed in
  the manifest.
- `pseudo-controller-rebuild-proof-v1` adds a selfhost proof command that maps
  pseudo inventory fields into `BuildPseudoEffect` control plans, writes
  `tmp/pseudo_controller_rebuild/proof.json`, and generates
  `tmp/pseudo_controller_rebuild/generated/yankfb2.aep` from the top family.
- `pseudo-behavior-wiring-proof-v1` adds a selfhost proof command that turns
  rebuild proof behavior notes into `tmp/pseudo_behavior_wiring/plan.json`
  with static/keyframe/expression wiring tasks.
- `sample-fact-behavior-extraction-v1` adds a selfhost extraction command that
  writes `tmp/pseudo_behavior_payloads/payloads.json`; current real run scanned
  90 projects with 0 failures and found payloads for all 12 keyframe behavior
  tasks.
- `pseudo-behavior-payload-application-proof-v1` adds a selfhost application
  command that writes `tmp/pseudo_behavior_application/application.json` and
  `tmp/pseudo_behavior_application/generated/yankfb2.aep`; current real run
  applied and verified 2 scalar keyframed controls for the generated
  `Pseudo/YanKFB2` family, skipped 7 non-generated families, and had 0 errors.
- `multi-family-pseudo-application-proof-v1` removes the top-family-only
  generation limit: current real run generates 8 controller proof AEPs for all
  selected supported pseudo families, then writes 3 application proof AEPs for
  payload-bearing families and applies/verifies all 12 scalar keyframed
  controls with 0 errors.
- `vector-pseudo-behavior-application-v1` adds vector payload detection and
  application through `aep.AnimateEffectParamVec`; synthetic vector payload
  coverage applies/verifies a 2-component point control, while the current real
  sample payload run remains 12 scalar controls applied/verified with 0 errors.
- `expression-pseudo-behavior-boundary-v1` makes expression payload application
  explicit in JSON: expression-only payloads are `expression_deferred`, and
  keyframe+expression conflict payloads can apply keyframes while preserving the
  expression as deferred. Current real samples still report 0 expression tasks,
  0 expression-deferred controls, and 0 errors.

Current:
- No active package is selected inside this topic. The latest completed package
  regenerated the stock-AE rain showcase, confirmed no third-party dependencies,
  rendered via AE2020, and recorded the package result in
  `tmp/technique_showcase_rain/package_summary.json`.

## Open questions

- Which next reference phenomenon should be used to validate the pipeline at larger scale.
- Whether technique atoms should remain in project-local knowledge or graduate into a broader shared library later.
