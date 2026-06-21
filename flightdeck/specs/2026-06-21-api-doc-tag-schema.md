---
status: active
summary: Replace free-prose doc comments + the separate aep:cap directive with ONE swaggo-style @tag block per exported symbol (docgen reads @summary/@description/@param/@returns; capindex reads @domain/@stability/@verify/@gate/@since/@boundary/@incident/@alias). Plus repo-wide cleanup of internal jargon (version codenames V2.2/V3/M8/wave-N, project-file/process refs CLAUDE.md/spec/probe/py-aep) from ALL comments, enforced by lint. Migrate behind a dual-read window (CI green throughout); 486 tags/26 files for schema + ~114 files for cleanup.
last_updated: 2026-06-21
---

# API doc-comment @tag schema (unified with capindex) + repo-wide jargon cleanup

## Problem

Two complaints, both real:

1. **Comment format is inconsistent + mixes machine-data with prose.** Each exported
   symbol carries free-prose doc comments (docgen turns the first sentence into a
   summary, the rest into a description) PLUS a separate `//aep:cap …` directive
   (capindex's CI-enforced capability truth source). The two are unrelated formats; the
   prose carries parameter descriptions inline (no structure); and the `aep:cap
   boundary="…"` field is written in Chinese (64 occurrences in facade.go alone) — a
   language mix.

2. **Internal jargon leaks into comments.** Version codenames (`V2.2`, `V2_2`, `V3`,
   `M8`, `wave 2`), project-file references (`CLAUDE.md #2/#5`, `rules.md`, spec /
   incident filenames written in prose), and RE/process noise (`(probe)`, `RE'd from
   X.aep`, `tmp_debug/…`, fixture line numbers, `py-aep`) appear throughout — meaningless
   to an API consumer. Measured: V2.2/V2_2 in **53 files**; any jargon token in **114
   files** (69 non-test .go); **328 occurrences in doc-comment lines** (non-test), plus
   more in internal implementation comments.

## Goals

- **One annotation block** per exported symbol, swaggo-style `@tag` lines, consumed by
  BOTH docgen (human docs) and capindex (machine capability data). No second directive,
  no drift between two sources of truth.
- **English everywhere** in comments (real CJK *string examples* like a Japanese layer
  name are allowed as quoted data; prose/field text is English).
- **No project-internal jargon** in any comment, repo-wide, enforced by lint.
- **CI green throughout the migration** (no flag day).

Non-goals: changing the public API surface, changing what capindex enforces (same
fields, same crosscheck rules), changing `Example*`-function-based examples.

## The @tag schema

One `//`-comment block per exported symbol. `@`-prefixed lines carry structured fields;
any non-`@` lines before the first tag are ignored (transitional). Fields:

| tag | consumer | required | replaces | notes |
|---|---|---|---|---|
| `@summary <one line>` | docgen | yes | prose first sentence | the symbol's one-line doc |
| `@description <multi-line>` | docgen | no | rest of prose | rich domain prose (AE mechanics, gotchas) |
| `@param <name> <english desc>` | docgen | one per param | inline prose | rendered as a param table |
| `@returns <english desc>` | docgen | no | inline prose | |
| `@domain <name>` | capindex | yes (write surface) | `domain=` | e.g. mask, layer, effect |
| `@stability stable\|alpha` | capindex | yes (write surface) | `tier=` | |
| `@verify ae-accept\|render-pixel\|roundtrip\|none` | capindex | yes | `verify=` | crosscheck: ae-accept/render-pixel needs a live gate |
| `@gate <Test1,Test2>` | capindex | when verify≠none/roundtrip | `gate=` | comma list of ship-gate test names |
| `@since AE2020\|AE2025` | capindex | yes | `AE≥`/`minVer` | min AE read/write version |
| `@boundary <english>` | capindex | no | `boundary=` | coverage edge; **English** |
| `@incident <file>` | capindex/docgen | no | `incident=` | the real home for an internal RE pointer |
| `@alias <comma list>` | capindex | no | `alias=` | extra capindex search terms |

**Rendering note (accepted trade-off):** unlike the current `//aep:cap` (no space →
go/doc strips it), `// @tag` lines are ordinary doc comments and DO appear in native
`go doc`. Acceptable here: the doc surface is `internal/aep` (not pkg.go.dev public) and
the real docs are rendered by docgen into `docs/*.md`; `@summary`/`@description` still
read naturally in `go doc`.

**Example:**

```go
// @summary    Add a vector mask to a layer
// @description Appends a closed Bezier mask to the layer's Mask Parade. The
//   outline is in layer-pixel space; AE recomputes the parade bounds on open.
// @param      layer  owning layer (from a parsed project)
// @param      path   outline in layer-pixel space (closed Bezier)
// @returns    the created *Mask
// @domain     mask
// @stability  stable
// @verify     ae-accept
// @gate       TestAddMask_AEShipGate_AE2020,TestAddMask_AEShipGate_AE2025
// @since      AE2020
// @boundary   removing the last mask leaves an empty parade (AE tolerates it)
// @incident   add-mask-create-re
func AddMask(layer *Layer, path BezierPath) (*Mask, error) { … }
```

## Jargon-cleanup rule (repo-wide, all comments)

Comments (doc, implementation, and package comments) must not name project-internal
artifacts or process. Forbidden token classes:

- **Version/milestone codenames:** `V2.2`, `V2_2`, `v2.2`, `V3`, `M8`, `Phase <n>`
  (as a codename), `wave <n>`.
- **Project-file/process references:** `CLAUDE.md`, `rules.md`, `flightdeck`, `cockpit`,
  spec/incident **filenames** written in prose, `py-aep`.
- **RE/scratch noise:** `(probe)`, `RE'd from <x>.aep`, `tmp_debug/…`, fixture line-number
  refs (`tolerance.aep dump line 145`).

Replacements:

- A genuine cross-reference to a recorded RE finding → the `@incident <file>` field (not
  inline prose).
- Domain mechanics that ARE useful to a reader (AE elides default channels; coordinate
  spaces; chunk-level behavior) stay, rephrased without the codename. (Binary-format
  identifiers like chunk IDs are domain, not jargon — kept.)
- Real, user-facing AE version numbers (`AE 2020`, `AE 2025`) stay.

Enforced by a lint test (token blocklist over all `internal/**/*.go` comments).

## Tooling changes

1. **Shared `@tag` parser.** One package both capindex and docgen import (candidate:
   `internal/apidoc` or a small `cmd`-shared lib), parsing a comment block into a struct
   carrying every field above. Single parse implementation, no duplication.
2. **capindex.** Replace `parseCapTag` (`//aep:cap …`) with the shared parser reading
   `@domain/@stability/@verify/@gate/@since/@boundary/@incident/@alias`. Crosscheck,
   surface-coverage, and generated-doc tests keep the same *semantics* (every write-surface
   symbol tagged; verify↔gate consistency); only the source-format changes.
3. **docgen.** Extend `extract` to read `@summary/@description/@param/@returns`; render
   `@param` as a parameter table in `docs/*.md`. `Example*` functions unchanged.
4. **Lint test.** Token-blocklist test for the jargon rule + a "schema completeness" test
   (every write-surface symbol has `@summary/@domain/@stability`).

## Migration (CI green throughout)

- **Dual-read window.** During migration the shared parser accepts EITHER the legacy
  `//aep:cap` block OR the new `@tag` block per symbol. capindex/docgen prefer `@tag`
  when present, fall back to `aep:cap`. Files flip one at a time; CI never breaks.
- **Mechanical pre-fill converter.** A one-shot tool reads each symbol's existing
  `aep:cap` + prose and emits a `@tag` block: `@domain/@stability/@verify/@gate/@since/
  @boundary/@incident/@alias` from `aep:cap` (mechanical), `@summary` from the prose first
  sentence, `@description` from the remaining prose, and `@param <name> TODO` / `@returns
  TODO` placeholders (params can't be auto-extracted from free prose).
- **Manual pass per symbol.** Fill `@param`/`@returns`, translate `@boundary` to English,
  strip internal jargon, polish `@description`. This is the real labour (~486 symbols).
- **Flip CI.** When all symbols are converted: remove the legacy `aep:cap` parse path,
  enable the jargon-blocklist lint + schema-completeness test as required CI.

## Phasing

- **P1 — schema + tooling + proof:** shared parser; dual-read in capindex & docgen;
  converter tool; `@param` rendering in docgen; convert **facade.go** (77 funcs, the public
  API) fully (manual `@param` + de-jargon) as the proven pattern. CI green via dual-read.
- **P2 — bulk migration:** mechanical pre-fill + manual pass over the remaining 25
  aep:cap files (~400 symbols) and the non-aep:cap files carrying jargon (~114 files total
  for cleanup; cleanup-only files get the jargon pass, no `@tag` block since they have no
  write surface).
- **P3 — flip CI:** drop the legacy `aep:cap` parser; turn on the jargon-blocklist lint
  and schema-completeness test as required; update CLAUDE.md / rules.md to document the new
  convention (the doc-source-of-truth section).

## Scope (measured)

- `@tag` schema migration: **486 `aep:cap` tags / 26 files** (the write surface).
- docgen reads 3 packages: `internal/aep`, `internal/scene`, `internal/codec`.
- Jargon cleanup: **~114 files** (53 with V2.2; 69 non-test .go); **328 doc-comment
  occurrences** + implementation-comment occurrences.

## Risks

- **Manual `@param` labour** is the bulk and is not automatable — the converter only
  placeholders it. Mitigated by phasing (P1 proves the pattern on facade.go before bulk).
- **capindex is CI-enforced truth source.** Rewriting its parser risks coverage gaps.
  Mitigated by keeping the *semantics* identical and the dual-read window (tests stay green
  on every commit; the parser swap is internal).
- **go/doc visibility** of `@tag` lines (accepted above).
- **Churn across a working system.** Most fields already exist in `aep:cap`; the genuine
  new value is `@param`/`@returns` structure + de-jargon. Accepted by the user as worth it
  for a single consistent annotation.

## Self-review

- Placeholders: none (P-phases and fields concrete).
- Consistency: schema fields ↔ capindex fields ↔ migration converter all enumerate the same
  set; docgen vs capindex consumer split stated once and reused.
- Scope: one focused convention change; large but single-purpose. Decomposed into P1/P2/P3
  for the plan.
- Ambiguity: cleanup boundary fixed to "all comments repo-wide"; `@tag` vs `aep:cap` fixed
  to "unified, @tag absorbs aep:cap"; spec home = flightdeck/specs.
