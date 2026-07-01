# Versioned AEP Migration Validation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans when implementing validation tooling changes. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make migration validation answerable by domain, capability, writer target, AE host version, and evidence level.

**Architecture:** Treat writer targets and AE host versions as separate axes. The matrix runner produces raw case artifacts under `tmp/`; the Flightdeck validation summary is the source of truth for reviewed results. Each domain row links to a concrete matrix artifact and states exactly which version axis is covered.

**Tech Stack:** Go `cmd/aepmigrate matrix`, recipe fixtures in `examples/recipes`, profile diff reports, AE open smoke reports, Flightdeck markdown registry.

---

## Validation Axes

### Writer Targets

Writer targets are the AEP templates the Go writer can currently produce:

- `W2020` = `AE2020`
- `W2022` = `AE2022`
- `W2025` = `AE2025`

Current code does not provide native `AE2021`, `AE2023`, or `AE2024` writer templates. Those years must not be reported as writer targets until serializer templates exist.

### AE Host Versions

AE host versions are installed After Effects applications used to open a produced `.aep`:

- `H2020`
- `H2021`
- `H2022`
- `H2023`
- `H2024`
- `H2025`

Host validation answers compatibility: "Can this produced AEP open in AE version X?"

### Evidence Levels

- `PD-3x3`: profile-diff matrix passed for source writers `W2020,W2022,W2025` into target writers `W2020,W2022,W2025`.
- `PD-1x3`: profile-diff matrix passed for source writer `W2020` into target writers `W2020,W2022,W2025`.
- `OPEN-H2025`: converted output opened in AE2025.
- `OPEN-ALL-HOSTS`: one chosen writer output opened in `H2020,H2021,H2022,H2023,H2024,H2025`.
- `OPEN-ENDPOINTS`: one chosen writer output opened in `H2020` and `H2025`.
- `INFER-MID-HOSTS`: `H2021,H2022,H2023,H2024` were not run, but are treated as low-risk inferred compatibility because both `H2020` and `H2025` passed. This must be marked as inference, not as direct validation.
- `SHIPGATE-2020/2025`: lower-level writer API ship-gate exists for AE2020 and AE2025. This is useful supporting evidence, but it is not a migration matrix result.

## Result Registry

Reviewed validation results must be recorded in:

- `flightdeck/work/aep-understanding-generation/versioned-aep-migration-validation-summary.md`

Raw matrix artifacts may live under `tmp/`, but `tmp/` is not the project memory. A result is not considered reported until the summary includes:

- domain
- capability or recipe set
- source writer coverage
- target writer coverage
- AE host coverage
- raw artifact path
- date and status

## Canonical Commands

Profile-diff for one domain across all writer axes:

```powershell
$recipes = Get-ChildItem examples\recipes\minimal-text-animator*.json | ForEach-Object { @('-recipe', $_.FullName) }
go run ./cmd/aepmigrate matrix @recipes -sources all -targets all -out tmp\migration_matrix_text_animators_all_writers
```

AE host fanout for one representative output:

```powershell
go run ./cmd/aepmigrate matrix `
  -recipe examples\recipes\minimal-text-animator-skew.json `
  -sources AE2020 `
  -targets AE2020 `
  -out tmp\migration_matrix_text_animator_skew_all_hosts `
  -ae-root E:\adobe `
  -ae-open `
  -ae-versions all `
  -ae-timeout-sec 240 `
  -max-ae-open-cases 6
```

Full no-AE migration boundary:

```powershell
go run ./cmd/aepmigrate matrix -recipes examples\recipes -sources AE2020 -targets AE2020,AE2022,AE2025 -out tmp\migration_matrix_smoke_all
```

Do not run a broad AE-open matrix without an explicit case count decision.

## Domain Buckets

Use these buckets in the summary:

- `project`: project settings and display settings.
- `comp`: comp settings, work area, renderer, metadata.
- `layer`: default layer creation, switches, refs, timing, parent/matte, source refs.
- `transform`: transform static values, keyframes, ease, expressions.
- `text`: text layer baseline, text style, text animators.
- `effect`: built-in effects, params, keyframes, expressions, layer refs.
- `shape`: parametric shape graphics, fills/strokes, filters, gradients.
- `camera-light`: camera options, light options, light source refs.
- `precomp`: precomp items and precomp layers.

## Execution Rules

- Use the project-level "总 → 分 → 总" rule for this validation work:
  first keep this validation plan and summary as the total control surface,
  then execute domain matrices one bucket at a time, then return to the summary
  to record status and gaps before moving to the next bucket.
- Before implementing a new migration slice, add or identify its recipe bucket.
- Before claiming validation, run the smallest domain matrix that proves the slice.
- After the run, update the validation summary in the same change.
- When a matrix run is intentionally interrupted, do not promote its result.
- When a matrix result comes from `tmp/`, include the exact artifact path and summary counts.
- When AE crashes or a host is missing, record the host status separately from writer support.
- If a capability only runs `H2020` and `H2025`, mark intermediate hosts as
  `INFER-MID-HOSTS`; do not report `H2021-H2024` as directly run unless their
  case reports exist.

## Current Gap

The existing matrix runner can produce raw case artifacts, but it does not yet emit a domain coverage ledger. Until that tooling exists, the markdown summary is the reviewed ledger.

Planned tooling upgrade:

- [ ] Add a recipe catalog mapping recipe names to domain and capability.
- [ ] Add `aepmigrate matrix -coverage-out coverage.json`.
- [ ] Add a generated markdown coverage view grouped by domain/capability.
- [ ] Add tests proving `-sources all`, `-targets all`, and `-ae-versions all` do not conflate writer and host axes.
