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
- `cmd/aepmigrate matrix` is the latest completed slice. It batches recipe
  fixtures across source/target writer labels, writes aggregate/per-case
  reports, discovers AE hosts, supports explicit `-ae-versions` host-version
  fanout for `-ae-open`, and protects broad AE-open runs with
  `-max-ae-open-cases`.
- `cmd/aepmigrate convert` now also preserves supported layer-level label and
  comment metadata for reconstructed non-shape layers; the layer label/comment
  recipe fixtures pass AE2020/AE2022/AE2025 matrix conversion and AE2025 open
  smoke.
- Supported text-layer switch reconstruction now covers visible, solo, shy,
  locked, effects/audio switches, motion blur, frame blend, blending mode, and
  quality; the common-switches/shy/motion-blur/quality-blending recipe fixtures
  pass AE2020/AE2022/AE2025 matrix conversion and AE2025 open smoke.
- Supported text-layer advanced switch reconstruction now covers collapse
  transform, 3D, adjustment, guide, bicubic sampling, pixel-motion frame blend,
  and preserve transparency; `minimal-layer-advanced-switches.json` passes
  AE2020/AE2022/AE2025 matrix conversion and AE2025 open smoke.
- Text-layer parent refs are now remapped onto the target project's layer IDs
  after same-comp layer reconstruction; `minimal-layer-parent.json` and the
  null-parent `minimal-null-layer.json` pass AE2020/AE2022/AE2025 matrix
  conversion and AE2025 open smoke.
- Solid-backed null-controller layers now preserve source solid dimensions,
  color, visibility, and null flag while rebuilding as null layers;
  `minimal-layer-null-flag.json` passes AE2020/AE2022/AE2025 matrix conversion
  and AE2025 open smoke.
- Camera and light static option reconstruction now preserves profile-visible
  camera options and light kind/color/intensity/cone/falloff/shadow options;
  camera option fixtures plus light option fixtures except `minimal-light-source.json`
  pass AE2020/AE2022/AE2025 matrix conversion, with camera/light object-profile
  AE2025 open smokes passing.
- Continue versioned migration by expanding supported conversion surface or
  deciding whether native AE2021/AE2023/AE2024 writer templates are worth
  adding.
- Do not start automated correction loops.

## Open questions

- How much capability linking can be automated from current profile paths
  before a richer path-to-capability index exists.
- What the smallest Phase 5 replication slice should be, so it exercises both
  structural and render gaps without forcing a full-project rebuild first.
- Which new recipe-owned family should enter next once there is a proven
  parser/profile surface and a bounded writer contract.
