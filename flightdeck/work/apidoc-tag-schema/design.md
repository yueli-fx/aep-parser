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

**The legacy form being replaced** (prose + a separate hidden directive, params described
inline, machine-data Chinese):

```go
// AddMask adds a vector mask ... <prose, params buried in sentences> ...
//
//aep:cap domain=mask tier=stable verify=ae-accept gate=TestAddMask_AEShipGate_AE2020,TestAddMask_AEShipGate_AE2025 boundary="删最后一个 mask 留空 parade(AE 容忍);mask 须来自 parsed 工程" alias="remove mask,删蒙版"
```

## Field rules (the validation schema)

Every field has a machine-checkable rule so "clean / consistent / English" stops being a
subjective reviewer call and becomes a validator pass/fail. Enums are **frozen lists**
(extending one is a deliberate edit, so typos are caught).

| field | required | type / format | rule |
|---|---|---|---|
| `@summary` | yes | single line | imperative-mood opening; **≤ 80 chars**; no trailing period; English; no jargon |
| `@description` | no | multi-line | English; no jargon (no hard length cap) |
| `@param <name> <desc>` | one per non-receiver parameter | one line each | `<name>` **must equal a parameter name in the signature**; `<desc>` non-empty, English, ≥ 3 words; the literal `TODO` is rejected |
| `@returns <desc>` | required iff the func returns ≥ 1 non-`error` value | one line | non-empty, English |
| `@domain` | yes (write surface) | frozen enum (16) | `shape · layer-set · layer-create · text · mask · effect · gradient · keyframe · comp · project · render-queue · structural · eg · expr · io · meta` |
| `@stability` | yes (write surface) | enum | `stable · alpha` |
| `@verify` | yes (write surface) | enum | `ae-accept · render-pixel · roundtrip · none` |
| `@gate` | required iff `@verify ∈ {ae-accept, render-pixel}` | comma list of test names | each matches `Test\w+` AND exists in test sources (existing crosscheck) |
| `@since` | yes (write surface) | format `AE<year>` | frozen enum `AE2020 · AE2025` |
| `@boundary` | no | one or more lines | English; no jargon |
| `@incident` | no | incidents/ filename (no ext) | file must exist under `flightdeck/incidents/` (or `archive/incidents/`); **maintainer field — docgen does NOT render it** |
| `@alias` | no | comma list | lowercase tokens; CJK search terms allowed |

"Write surface" is exactly what capindex already defines in `cmd/capindex/surface.go`
(`requires()`): every `pkg=aep` exported func + every exported method/getter on the
covered types. The schema-completeness check reuses that predicate verbatim — no new
definition.

`@since` carries the READ floor (matches the current `aep:cap` single min-version; no
regression). If a symbol ever needs a distinct write floor, that is a future additive
field (`@since-write`), not blocked here.

## One-command validation

`go run ./cmd/capindex --validate` — prints every violation with `file:line symbol: msg`
and exits non-zero; the same checks run under `go test ./cmd/capindex` for CI. Checks:

1. **Completeness:** every write-surface symbol has `@summary @domain @stability @verify
   @since`.
2. **Enums/format:** `@domain @stability @verify @since` values legal; `@summary` ≤ 80
   chars and single line; `@since` matches `AE<year>`.
3. **`@param` ↔ signature:** the set of `@param` names equals the function's non-receiver
   parameter names — missing, extra, mis-named, or `TODO` all fail.
4. **`@returns`:** present when the signature returns a non-`error` value.
5. **`@gate` existence:** each named test exists (the existing crosscheck) and is present
   when `@verify` demands it.
6. **`@incident` existence:** referenced file exists.
7. **Single-format invariant (migration):** no symbol carries BOTH a legacy `//aep:cap`
   block and a `@tag` block — the converter replaces in place; this removes any
   "which wins / they disagree" ambiguity.
8. **Jargon blocklist:** no comment contains a blocklisted token (see below).
9. **Exit criteria (gates the format flip):** count of remaining `//aep:cap` blocks and
   count of `@param … TODO` placeholders — the legacy parser is removed and `--validate`
   becomes required only when BOTH reach **0**. This is the objective "migration done"
   judgment the reviewers asked for.

Until the flip, `--validate` runs in **warn mode** for not-yet-converted symbols (reports,
does not fail) so CI stays green during the two-step migration; converted symbols are held
to the full schema immediately.

## Schema home — one place to change types

All schema knowledge lives in ONE package, `internal/apidoc`, imported by both
`cmd/capindex` and `cmd/docgen`. No enum or rule is duplicated anywhere else.

- `internal/apidoc/schema.go` — the **single source of truth**, declared as data:
  - frozen enums as Go slices: `Domains` (the 16), `Stabilities`, `Verifies`,
    `SinceVersions`;
  - the field table: for each tag, whether it is required (a predicate over the symbol),
    its kind, and its validator func;
  - the path to the jargon blocklist data file.
- `internal/apidoc/parse.go` — the `@tag`-block → struct parser.
- `internal/apidoc/validate.go` — the checks, consuming `schema.go`'s enums + rules.

**How to add or remove a type** (the question this answers): editing exactly one slice in
`internal/apidoc/schema.go` is the whole change. Add a `@domain` → append to `Domains`;
retire a `@verify` value → remove it from `Verifies`. capindex, docgen, and `--validate`
all read these slices, so the new/removed type takes effect everywhere at once. A header
comment in `schema.go` states "this file is the only place to edit the annotation
vocabulary", and a test asserts no domain/verify/stability string literals exist outside
this package (so a future edit can't silently fork the enum).

## Generation self-validates

Doc generation does not trust its input. `cmd/docgen` calls `apidoc.Validate` over the
symbols it is about to render and **fails (non-zero) before writing any file** if an
annotation violates the schema — so `go generate ./...` / `go run ./cmd/docgen` cannot
emit docs from invalid or incomplete tags. It is the *same* `apidoc.Validate` that
`cmd/capindex --validate` and `go test ./cmd/capindex` run — one implementation, enforced
at generation time AND in CI. (During the migration window `Validate` takes a mode flag:
warn for un-converted symbols, strict for converted ones and for the final flip.)

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

**Deterministic blocklist (not semantic judgment).** The lint matches concrete,
case-insensitive regexes, covering spelling variants — e.g. `v2[._]?\d`, `\bV3\b`,
`\bM8\b`, `wave[ _-]?\d`, `phase[ _-]?\d` (codename use), `CLAUDE\.md`, `rules\.md`,
`flightdeck`, `cockpit`, `py-?aep`, `\(probe\)`, `tmp_debug`, `RE'?d from`. The list lives
in one file (`cmd/capindex` lint data) so it is auditable and extendable.

**Escape hatch + history distinction.** A comment line ending with `//nolint:jargon`
is exempt — for the rare case where a token is genuinely required (e.g. a compatibility
note that must name a real external version). The intent is to ban *meaningless internal
codenames and process noise*, NOT valuable history: a real RE finding worth keeping goes
to `@incident <file>`; durable domain history worth a reader's time stays in
`@description`, rephrased without the codename. If neither fits and the token must stay,
`//nolint:jargon` makes that an explicit, reviewable decision rather than a silent leak.

**Scope:** codename + process tokens are linted across **all** `internal/**/*.go` comments
(tests included — `V2.2` is noise everywhere). The `@tag` schema validation itself applies
only to the non-test write surface (test files carry no exported API tags).

## Tooling changes

1. **`internal/apidoc` (new, the schema home).** Holds `schema.go` (enums + field rules,
   the single source of truth — see *Schema home*), `parse.go` (the `@tag`-block parser),
   and `validate.go` (`apidoc.Validate`). Imported by both `cmd/capindex` and `cmd/docgen`;
   one parse + one validate implementation, no duplication.
2. **capindex.** Replace `parseCapTag` (`//aep:cap …`) with `apidoc` reading
   `@domain/@stability/@verify/@gate/@since/@boundary/@incident/@alias`. Crosscheck,
   surface-coverage, and generated-doc tests keep the same *semantics* (every write-surface
   symbol tagged; verify↔gate consistency); only the source-format changes.
3. **docgen.** Extend `extract` to read `@summary/@description/@param/@returns`; render
   `@param` as a parameter table in `docs/*.md`; `@incident` parsed but NOT rendered
   (maintainer field); `Example*` unchanged. **Calls `apidoc.Validate` before writing any
   file and fails on violation** (see *Generation self-validates*).
4. **Validator surface.** `cmd/capindex --validate` (human "一键校验") and
   `go test ./cmd/capindex` (CI) both call `apidoc.Validate` — the same function docgen
   runs. The jargon blocklist data lives in one auditable file under `internal/apidoc`.
5. **Converter (`tools/debug/tagconvert`).** One-shot, tracked: rewrites a file's
   `aep:cap` + prose into `@tag` blocks with placeholders, for the manual pass to finish.

## Migration — two steps (CI green throughout)

A **dual-read window** spans both steps: the shared parser accepts EITHER a legacy
`//aep:cap` block OR a `@tag` block per symbol (never both — invariant #7). `--validate`
runs in warn mode for un-converted symbols, full-strict for converted ones, so CI stays
green while the surface flips incrementally.

**Step 1 — schema + tooling + facade proof.**
Shared parser; capindex + docgen both read `@tag` (fall back to `aep:cap`); converter
tool; `--validate` (warn mode); docgen `@param` table rendering; convert **facade.go**
(77 funcs, the public API) fully — including translating its ~64 Chinese `boundary`
fields to English `@boundary` and writing real `@param`/`@returns`. Deliverable: the
schema, the validator, and one fully-converted file proving the pattern end-to-end.

**Step 2 — bulk convert + cleanup + flip.**
Mechanical pre-fill + manual pass over the remaining ~25 `aep:cap` files (~400 symbols);
repo-wide jargon cleanup (the codename/process lint, all `internal/**` incl. tests);
then the **flip**, gated objectively on invariant #9 (remaining `aep:cap` = 0 AND
`@param … TODO` = 0): remove the legacy `aep:cap` parser, make `--validate` + jargon lint
required CI, and update the project doc-source-of-truth note.

**Commit hygiene (addresses the rollback concern):** schema-conversion commits and
jargon-cleanup commits are kept **separate** — a doc/comment-cleanup commit never mixes
with a `@tag`-conversion commit — so the schema work can be reverted without dragging the
(large, mechanical) cleanup diff, and vice-versa.

The reviewers' "this is really three projects" critique is real; the response is **strict
separation by commit stream + objective exit gates**, not a second spec — the three
threads (schema, validator, cleanup) share one parser and one validator, so splitting the
*spec* would just fragment one tightly-coupled toolchain.

## Scope (measured)

- `@tag` schema migration: **486 `aep:cap` tags / 26 files** (the write surface).
- docgen reads 3 packages: `internal/aep`, `internal/scene`, `internal/codec`.
- Jargon cleanup: **~114 files** (53 with V2.2; 69 non-test .go); **328 doc-comment
  occurrences** + implementation-comment occurrences.

## Why now (benefit)

The library only grows; a second annotation format and ad-hoc prose get more expensive to
keep consistent every release. Converging on ONE machine-validated annotation now caps that
debt: a single `--validate` becomes the gate for *all* doc correctness (params match
signatures, enums legal, no jargon, gates exist), so future API additions can't drift —
the cost is paid once, against a surface that is still small enough to convert.

## Risks (and how the schema retires them)

- **Manual editing consistency** (the reviewers' top risk, correctly) — 486 symbols hand-
  edited invites quality drift. **Retired by the validator:** `@param`↔signature matching,
  enum legality, `@summary` length, jargon blocklist, and no-`TODO` are all machine-checked,
  so a sloppy manual edit fails CI rather than slipping through. This is the core reason the
  strict field schema exists.
- **capindex parser rewrite** (CI truth source) — mitigated by identical *semantics*
  (same `requires()` surface, same crosscheck) + the single-format invariant + dual-read;
  the parser swap is internal and every step keeps tests green.
- **`go doc` visibility of machine `@tag` lines** — accepted: `internal/aep` is not a
  pkg.go.dev surface, docgen renders the real docs, and `@summary`/`@description` still read
  naturally; the trade buys one annotation system instead of two.
- **Converter covers only mechanical fields** — true; `@param`/`@returns`/English/de-jargon
  stay manual. Accounted for in Step 1's facade estimate (incl. the ~64 Chinese `boundary`
  translations) so the proof reflects real per-symbol cost before bulk.

## External-review points folded in

Three independent reviews (do not assume full-context) converged on real gaps; resolutions:

- **Over-bundled / hard rollback** → separate commit streams + objective exit gates (kept
  one spec since schema/validator/cleanup share one toolchain).
- **Subjective acceptance ("English", "no jargon", "polish")** → the field-rule schema +
  `--validate` make every one machine-checkable.
- **No conflict rule / mixed dual-read** → single-format invariant (#7): never both formats.
- **No objective "done"** → exit gate (#9): remaining `aep:cap` = 0 AND `@param TODO` = 0.
- **`@param` could be meaningless** → name-must-match-signature + ≥3-word + no-`TODO` check.
- **Jargon rules too semantic** → concrete regex blocklist + `//nolint:jargon` escape +
  `@incident`/rephrase for valuable history.
- **`@incident` vs no-internal-refs contradiction** → `@incident` is a structured maintainer
  field (docgen does not render it), not reader prose.
- **test-file scope contradiction** → codename lint repo-wide incl. tests; schema validation
  only on the non-test write surface.
- **`@since` can't split read/write floor** → documented as the read floor (no regression);
  `@since-write` is a future additive field.
- **Legacy format shown** → added alongside the new-format example.
- Not adopted: "split into 3 specs" (one toolchain), "@alias → config" (capindex is a
  source-inline truth index by design), "use AEP:domain prefix instead of @" (cosmetic).

## Self-review

- Placeholders: none; every field has a concrete rule and a validator check.
- Consistency: field-rule table ↔ `--validate` checks ↔ converter output enumerate the same
  field set; docgen vs capindex consumer split stated once and reused.
- Scope: one tightly-coupled toolchain (parser + validator + converter) + a cleanup pass;
  two-step migration, separate commit streams.
- Ambiguity: cleanup boundary = all comments repo-wide (codename tokens incl. tests);
  `@tag` absorbs `aep:cap` (unified); single-format invariant during migration; spec home =
  flightdeck/specs.
