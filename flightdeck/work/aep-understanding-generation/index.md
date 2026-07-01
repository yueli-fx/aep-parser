# Index — aep-understanding-generation

## State

This is the active long-running AEP understanding/generation work package. It
started as a pipeline to turn `aep-parser` from a parser/writer plus bespoke
showcase tooling into reusable project understanding, replication diagnostics,
recipe generation, corpus learning, and versioned AEP migration.

Current focus: versioned AEP migration. The implemented surface now includes
`cmd/aepmigrate assess`, conservative `convert`, and `matrix` gates for
AE2020/AE2022/AE2025 writer targets. Matrix AE-open validation can now use a
separate AE host-version dimension across installed AE2020-AE2025 hosts.

History lives in `history.md`; keep this index short enough to recover from.

## Next

Continue versioned migration by choosing one of:

- expand the supported `convert` surface beyond the current conservative layer
  and shape slices;
- decide whether AE2021/AE2023/AE2024 native writer templates are worth adding;
- promote matrix coverage into a recurring verification command once the pass /
  blocked boundary is useful enough.

Do not start automated correction loops.

## Read now

- `versioned-aep-migration-spec.md` — product and architecture spec for
  explicit AE-version upgrade/downgrade assessment, conversion, diff, and loss
  reporting.
- `versioned-aep-migration-convert-plan.md` — first conservative convert
  slices: target-version skeleton rebuild for no-layer comp projects and
  default null-layer projects, with unsupported layer-bearing projects blocked
  before output.
- `versioned-aep-migration-matrix-plan.md` — matrix gate scope, AE-open case
  guard, and verification record.
- `versioned-aep-migration-validation-plan.md` — validation axes, evidence
  levels, and result-recording rules.
- `versioned-aep-migration-validation-summary.md` — reviewed validation ledger
  grouped by domain/capability/version coverage.
- `versioned-aep-migration-remaining-blockers-plan.md` — current total/slice/
  total plan for refreshing the post-text-animator blocker boundary and
  attacking the next smallest blocked family.
- `docs/capabilities.md` — current write capability matrix.

## Read if

- `history.md` — if you need detailed completed-slice history.
- `versioned-aep-migration-assess-plan.md` — if changing `aepmigrate assess`.
- `comp-recipe-execution-strategy.md` — before adding comp-level recipe/profile
  fields.
- `layer-recipe-execution-strategy.md` — before adding layer-level recipe/profile
  fields.
- `design.md` and old phase plans — only for archaeology or if the high-level
  pipeline direction is questioned.
- `flightdeck/knowledge/techniques/understand-a-project.md` and
  `flightdeck/knowledge/techniques/fx-techniques.md` — when resuming reference
  project internalization or technique-library work.

## Progress

Detailed historical progress moved to `history.md`.

Done at a high level:
- Phase 0-6 understanding pipeline foundation is implemented: stable profile,
  structural diff, AE render oracle, gap ledger, slice workflow, and
  source-vs-clone render compare.
- Minimal recipe IR grew into broad recipe/profile coverage for comp settings,
  layer creation/switches/timing/parent refs, camera/light options, effects,
  shape primitives/operators/stroke details/gradients, transform
  keyframes/ease/expressions, text style, and text animators.
- Technique facts, portrait, corpus, explain mode, and selfhost report/verify
  flows are implemented for project understanding and corpus learning.
- Versioned AEP migration now has `assess`, conservative `convert`, and
  `matrix` gates for AE2020/AE2022/AE2025 writer targets.

Current:
- Current `cmd/aepmigrate convert` surface includes conservative comp/layer
  skeleton rebuild, stable comp settings, supported non-shape layer metadata /
  switches / refs / timing / transforms, transform keyframes/ease/expressions,
  single-run/single-paragraph text style, supported text animator slices,
  supported camera/light options, supported parametric graphic/filter shape
  slices, supported gradient fill/stroke slices, precomp refs, supported layer
  matte/mask slices, and supported built-in effects.
- Effect reconstruction now covers supported static params, same-comp layer-ref
  params, parameter expressions, scalar keyframes, and vector keyframes.
- Text animator reconstruction now covers the current recipe-owned animator
  surface by profile-visible properties: opacity, position, scale, rotation
  Z/X/Y, fill/stroke color, tracking, character offset, fill/stroke opacity,
  stroke width, skew, range-offset keyframes, and supported value keyframes.
- Latest full no-AE matrix: 423 total, 420 pass, 3 intentionally blocked,
  0 failed, 0 skipped.
- `aepmigrate ledger` can generate per-recipe/domain Markdown ledgers from
  matrix JSON outputs; current generated ledgers live beside the smoke and
  all-host representative matrix artifacts under `tmp/`.
- Latest AE-open checks: dynamic effect-param and transform dynamic fixtures
  plus text style fixtures pass AE2025 open smoke; representative AE2020 writer
  outputs for vector effect keyframes, transform ease, text style, text
  animator skew/color-keyframes, classic track matte, layer mask, and base
  shape gradient stroke open across AE2020, AE2021, AE2022, AE2023, AE2024, and
  AE2025 hosts.
- Continue versioned migration by expanding supported conversion surface,
  turning the useful matrix boundary into a recurring verification command, or
  deciding whether native AE2021/AE2023/AE2024 writer templates are worth adding.
- Do not start automated correction loops.

## Open questions

- How much capability linking can be automated from current profile paths
  before a richer path-to-capability index exists.
- What the smallest Phase 5 replication slice should be, so it exercises both
  structural and render gaps without forcing a full-project rebuild first.
- Which new recipe-owned family should enter next once there is a proven
  parser/profile surface and a bounded writer contract.
