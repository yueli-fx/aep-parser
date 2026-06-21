---
status: active
summary: Step 1 of api-doc-tag-schema: build internal/apidoc (parse+schema+validate+jargon), wire capindex + docgen for dual-read (@tag with aep:cap fallback), add capindex --validate (warn mode) + docgen self-validation, ship the tagconvert tool, and fully convert facade.go (77 funcs) as the end-to-end proof. CI green throughout. Step 2 (bulk 25 files + jargon cleanup + flip) is a follow-on.
last_updated: 2026-06-21
implements: specs/2026-06-21-api-doc-tag-schema.md
---

# @tag 文档 schema 实现 (Step 1)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stand up `internal/apidoc` as the single source of truth for the exported-API annotation vocabulary, wire `cmd/capindex` and `cmd/docgen` to read the new `@tag` block (falling back to legacy `//aep:cap` during a dual-read window), ship `capindex --validate` + the `tagconvert` tool, and fully convert `internal/aep/facade.go` (77 funcs) as the end-to-end proof — all with CI green throughout.

**Architecture:** One new internal package (`internal/apidoc`) owns the parser, the frozen enums, the field-rule validator, and the jargon blocklist; both command tools import it (no duplicated enum or rule). capindex maps a parsed `Annotation` onto its existing `Cap`; docgen reads `@summary/@description/@param/@returns` and renders a param table. A symbol carries EITHER a `@tag` block OR a legacy `aep:cap` directive, never both (single-format invariant), so the migration flips file-by-file with tests staying green.

**Tech Stack:** Go (`go/ast`, `go/parser`, `go/doc`, `regexp`); existing `cmd/capindex` + `cmd/docgen` AST pipelines; `uv`-run flightdeck index script for deck bookkeeping.

> **Progress (2026-06-21, STEP 1 DONE):** facade conversion **77/77** — final batch 7
> (the last 11 funcs: AddEssentialProperty, RemoveEffect, AddMask, SetMaskPath,
> SetMaskPathKeyframes, RemoveMask, DuplicateMask, MoveMask, AddItem, RemoveItem,
> SetRenderer) converted + committed (c3753d4). `grep -c aep:cap internal/aep/facade.go`
> == 0. Gate loop green: `--validate` clean · docgen + capabilities (486) regenerated ·
> go vet + apidoc/capindex/docgen tests pass. **Task 9 (Step 1 facade proof) = DONE.**
> Step 1 (toolchain + facade proof) is complete; remaining work is **Step 2 only**
> (bulk ~25 files / ~400 symbols + repo-wide jargon cleanup + the flip). Plan stays
> `active` until Step 2 lands.
>
> **Cost lesson (carry into Step 2):** hand-transcribing each dense legacy comment →
> English @tag is slow + token-heavy. **Cheaper strategy for Step 2:** now that
> `--validate` (strict) + docgen self-validate are a deterministic pass/fail gate,
> dispatch a **subagent per file/batch** with (a) the field-rule schema, (b) 2–3
> converted facade examples as the pattern, (c) the jargon blocklist, (d) instruction to
> loop `go run ./cmd/capindex --validate` + `go run ./cmd/docgen` + commit until green.
> The validator is the safety net that makes code-via-subagent safe here. Note the jargon
> lint also flags `wave-NN` etc. in your OWN @boundary prose — keep boundaries
> codename-free. (Batch 7 done inline by hand; the 11 were few enough not to warrant a
> dispatch.)
>
> **Progress (2026-06-21):** Tasks 1–8 DONE + committed — the full `internal/apidoc`
> package (parse/schema/validate/jargon), capindex dual-read + `--validate`, docgen
> read/render/self-validate, and the `tagconvert` tool; all packages green (one
> pre-existing out-of-arc RED unit test `TestSynthControlEntries_PardLayout/label`
> unrelated). Plus a parser refinement (preserve blank-line paragraph breaks in
> `@description`). Task 9 IN PROGRESS — pattern proven end-to-end on **batch 1 (6/77
> funcs:** Open/FromReader/Reopen/NewProject/NewComposition/DuplicateComposition):
> `--validate` strict-clean, docs render param tables + multi-paragraph descriptions.
> **Remaining: 71 facade funcs** (same de-jargon + Chinese→English `@boundary` work),
> then Step 2 (bulk 25 files + repo-wide jargon lint + flip).

## Handoff — next session (start here)

**State:** Step 1 COMPLETE — toolchain DONE+committed; facade.go **77/77** converted (batch 7
committed c3753d4), `grep -c aep:cap internal/aep/facade.go` == 0, all green, `--validate` clean.
**Next session starts Step 2.**

**Step 2 = the only remaining work** (the other ~25 `aep:cap` files / ~400 symbols + repo-wide
jargon cleanup + the flip). Same gate-loop + cheap subagent strategy as above; see the Follow-on
section at the bottom for the full plan. Key reminders before bulk-converting:

1. **Find the surface:** `grep -rln "aep:cap" internal/` (facade.go is now 0; the rest are scene
   + serializer + codec files).
2. **Cheap path — subagent + validator gate:** one file/batch per dispatch, each given (a) the
   field-rule schema, (b) 2–3 converted facade funcs as the pattern, (c) the jargon blocklist, (d)
   loop `go run ./cmd/capindex --validate` → `go run ./cmd/docgen -manifest docs/docgen.json` →
   `go run ./cmd/capindex` (regen) → `go test ./cmd/capindex/ ./cmd/docgen/` → commit, until green.
   Per-func rules: `@summary` imperative ≤80 no trailing period · one `@param` per non-receiver arg
   (≥3 words, no `TODO`) · `@returns` only when the func returns a non-error value · machine fields
   `@domain/@stability/@verify/@since AE2020` (+`@gate` when verify is ae-accept/render-pixel,
   +`@incident`/`@boundary`/`@alias` carried) · translate Chinese `@boundary` to English · no jargon
   in YOUR prose (lint flags `wave-NN`, `V2.2`, `M8`, `CLAUDE.md`, `py-aep`, `tmp_debug`). Map
   old→new: `tier`→`@stability` · `minver=N`→`@since AEN` · `gate`/`incident`/`alias` as comma lists.
3. **Resolve the negative-tier question FIRST** (see Follow-on): scene files use
   `tier=planned/missing/negative` but `@stability` enum is stable/alpha only —
   `grep -rn 'tier=\(planned\|missing\|negative\)' internal/` to find them.
4. **Commit hygiene:** keep schema-conversion commits separate from the repo-wide jargon-cleanup
   commit. **Flip (invariant #9):** remaining `aep:cap` == 0 AND `@param … TODO` == 0 → remove the
   legacy parser, make `--validate` + jargon lint required CI, update the doc-source-of-truth note.
   When Step 2 lands, the whole plan is done → run `/flightdeck:landing`.

**For Step 2 (the other ~25 files / ~400 symbols)** — see the Follow-on section at the bottom. Same
subagent+validator loop, one file per dispatch; keep schema-conversion commits separate from the
repo-wide jargon-cleanup commit (commit-hygiene). Resolve the open negative-tier question (Follow-on)
before bulk-converting scene files that use `tier=planned/missing/negative` (the `@stability` enum is
stable/alpha only — `grep -rn 'tier=\(planned\|missing\|negative\)' internal/` to find them).

**Don't re-derive:** the toolchain (`internal/apidoc`, capindex `--validate`, docgen self-validate,
`tools/debug/tagconvert`) is built and tested — just use it. `git log --oneline -20` shows the 14
session commits.

## Global Constraints

- **Module import path:** `github.com/example/aep-parser/internal/apidoc` (matches the `github.com/example/aep-parser/...` paths already used in `cmd/docgen/extract.go`).
- **Frozen enums live in exactly one place** — `internal/apidoc/schema.go`. No `@domain`/`@stability`/`@verify`/`@since` string literal may be re-declared in `cmd/capindex` or `cmd/docgen` once converted (the legacy `cmd/capindex/tag.go` maps are tolerated only until the Step-2 flip).
- **`@domain` frozen enum (16):** `shape · layer-set · layer-create · text · mask · effect · gradient · keyframe · comp · project · render-queue · structural · eg · expr · io · meta`.
- **`@stability` enum:** `stable · alpha`. **`@verify` enum:** `ae-accept · render-pixel · roundtrip · none`. **`@since` enum:** `AE2020 · AE2025` (format `AE<year>`).
- **`@summary`:** single line, ≤ 80 chars, no trailing period, English, no jargon.
- **`@param <name> <desc>`:** `<name>` must equal a non-receiver parameter name in the signature; `<desc>` ≥ 3 words, non-empty, literal `TODO` rejected.
- **Single-format invariant:** no symbol carries BOTH a legacy `//aep:cap` block AND a `@tag` block.
- **CI green throughout:** `go vet ./... && go test ./...` passes after every task. `capindex --validate` runs in **warn mode** for un-converted symbols, **strict** for converted ones.
- **Comments:** English; internal doc comments on exported symbols are the documentation source (docgen). No new inline implementation comments unless WHY is non-obvious (project rule).
- **Commits:** conventional-commit style, direct to local `main`, never `push`. One logical unit per commit.

## File Structure

**New package `internal/apidoc/`** (the schema home — imported by both tools):
- `schema.go` — frozen enums (`Domains`/`Stabilities`/`Verifies`/`SinceVersions`) + the `has` helper + the "only place to edit the vocabulary" header. Single source of truth.
- `parse.go` — `Annotation`/`Param` structs + `Parse(raw string) (*Annotation, error)` (the `@tag`-block parser) + `HasLegacyCap(raw string) bool`.
- `validate.go` — `Mode`, `Symbol`, `Context` types + `Validate(...) []error` (completeness, enums, summary format, `@param`↔signature, `@returns`, `@gate`, `@incident`).
- `jargon.go` — the regex blocklist (the one auditable file) + `LintJargon(comment string) []JargonViolation`.
- `apidoc_test.go` — parser + validator + jargon table tests.

**`cmd/capindex/` (modify):**
- `model.go` — extend `Entry` with `Params`, `ReturnsNonError`, `Pos`, `Ann *apidoc.Annotation`.
- `extract.go` — `attachCap` reads `@tag` first (fallback to legacy); add `extractParamNames` + `returnsNonError`; stamp `Pos`.
- `tag.go` — add `capFromAnnotation(*apidoc.Annotation) Cap`.
- `main.go` — add `-validate` flag → run `apidoc.Validate` + `apidoc.LintJargon` over the surface.
- `validate_cmd.go` (new) — the `-validate` driver.

**`cmd/docgen/` (modify):**
- `model.go` — add `summary`, `params []apidoc.Param`, `returns`, `annotated bool` to `symbol`.
- `extract.go` — parse `@tag` in `funcSymbol` + `withMethods`; fall back to legacy prose.
- `render.go` — render `@summary`/`@param` table/`@returns` (and not `@incident`).
- `main.go` — call `apidoc.Validate` over rendered symbols before writing any file.

**`tools/debug/tagconvert/` (new, tracked):**
- `main.go` — one-shot converter: a file's `aep:cap` + prose → `@tag` blocks with placeholders.
- `convert.go` + `convert_test.go` — the rewrite logic + a golden test.

---

### Task 1: `internal/apidoc` — schema, model, parser

**Files:**
- Create: `internal/apidoc/schema.go`
- Create: `internal/apidoc/parse.go`
- Test: `internal/apidoc/apidoc_test.go`

**Interfaces:**
- Produces: `apidoc.Domains/Stabilities/Verifies/SinceVersions []string`; `apidoc.Param{Name, Desc string}`; `apidoc.Annotation{Summary, Description string; Params []Param; Returns, Domain, Stability, Verify string; Gate []string; Since, Boundary string; Incident, Alias []string; HasTags bool}`; `apidoc.Parse(raw string) (*Annotation, error)`; `apidoc.HasLegacyCap(raw string) bool`.

- [ ] **Step 1: Write the failing parser test**

`internal/apidoc/apidoc_test.go`:

```go
package apidoc

import "testing"

func TestParse_FullBlock(t *testing.T) {
	raw := `@summary    Add a vector mask to a layer
@description Appends a closed Bezier mask to the layer's Mask Parade.
  The outline is in layer-pixel space.
@param      layer  owning layer (from a parsed project)
@param      path   outline in layer-pixel space (closed Bezier)
@returns    the created *Mask
@domain     mask
@stability  stable
@verify     ae-accept
@gate       TestAddMask_AEShipGate_AE2020,TestAddMask_AEShipGate_AE2025
@since      AE2020
@boundary   removing the last mask leaves an empty parade
@incident   add-mask-create-re
@alias      remove mask,删蒙版
`
	a, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !a.HasTags {
		t.Fatal("HasTags = false, want true")
	}
	if a.Summary != "Add a vector mask to a layer" {
		t.Errorf("Summary = %q", a.Summary)
	}
	if want := "Appends a closed Bezier mask to the layer's Mask Parade.\nThe outline is in layer-pixel space."; a.Description != want {
		t.Errorf("Description = %q, want %q", a.Description, want)
	}
	if len(a.Params) != 2 || a.Params[0].Name != "layer" || a.Params[1].Name != "path" {
		t.Fatalf("Params = %+v", a.Params)
	}
	if a.Params[1].Desc != "outline in layer-pixel space (closed Bezier)" {
		t.Errorf("Params[1].Desc = %q", a.Params[1].Desc)
	}
	if a.Returns != "the created *Mask" {
		t.Errorf("Returns = %q", a.Returns)
	}
	if a.Domain != "mask" || a.Stability != "stable" || a.Verify != "ae-accept" || a.Since != "AE2020" {
		t.Errorf("caps = %q/%q/%q/%q", a.Domain, a.Stability, a.Verify, a.Since)
	}
	if len(a.Gate) != 2 || a.Gate[0] != "TestAddMask_AEShipGate_AE2020" {
		t.Errorf("Gate = %v", a.Gate)
	}
	if len(a.Incident) != 1 || a.Incident[0] != "add-mask-create-re" {
		t.Errorf("Incident = %v", a.Incident)
	}
	if len(a.Alias) != 2 || a.Alias[1] != "删蒙版" {
		t.Errorf("Alias = %v", a.Alias)
	}
}

func TestParse_NoTags(t *testing.T) {
	a, err := Parse("AddMask adds a vector mask to a layer.\nNo @ lines here.")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if a.HasTags {
		t.Error("HasTags = true, want false (no @tag lines)")
	}
}

func TestParse_UnknownTag(t *testing.T) {
	if _, err := Parse("@bogus something"); err == nil {
		t.Fatal("Parse: want error for unknown @tag")
	}
}

func TestHasLegacyCap(t *testing.T) {
	if !HasLegacyCap("Open parses an .aep file.\n\naep:cap domain=meta tier=stable verify=roundtrip") {
		t.Error("HasLegacyCap = false, want true")
	}
	if HasLegacyCap("@summary Open an aep file\n@domain io") {
		t.Error("HasLegacyCap = true, want false")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/apidoc/ -run TestParse -v`
Expected: build failure — `undefined: Parse`, `undefined: Annotation`.

- [ ] **Step 3: Write `schema.go`**

```go
// Package apidoc is the single source of truth for the exported-API annotation
// vocabulary. Both cmd/capindex and cmd/docgen import it; no enum or field rule
// is duplicated anywhere else.
//
// THIS FILE IS THE ONLY PLACE TO EDIT THE ANNOTATION VOCABULARY. To add a domain,
// append to Domains; to retire a verify value, remove it from Verifies. capindex,
// docgen, and `capindex --validate` all read these slices, so a change takes
// effect everywhere at once.
package apidoc

// Domains is the frozen set of capability buckets (16). "meta" is the lane for
// getters/readers/aliases/enum consts — present in the API but not verified
// capabilities.
var Domains = []string{
	"shape", "layer-set", "layer-create", "text", "mask", "effect",
	"gradient", "keyframe", "comp", "project", "render-queue", "structural",
	"eg", "expr", "io", "meta",
}

// Stabilities is the frozen set of API-maturity values.
var Stabilities = []string{"stable", "alpha"}

// Verifies is the frozen set of evidence levels.
var Verifies = []string{"ae-accept", "render-pixel", "roundtrip", "none"}

// SinceVersions is the frozen set of minimum-AE-version values (format AE<year>).
var SinceVersions = []string{"AE2020", "AE2025"}

func has(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Write `parse.go`**

```go
package apidoc

import (
	"fmt"
	"strings"
)

// Param is one @param entry: a signature parameter name plus its English desc.
type Param struct {
	Name string
	Desc string
}

// Annotation is the parsed @tag block for one exported symbol. HasTags is false
// when the comment carries no @tag line at all (un-converted legacy prose).
type Annotation struct {
	Summary     string
	Description string
	Params      []Param
	Returns     string
	Domain      string
	Stability   string
	Verify      string
	Gate        []string
	Since       string
	Boundary    string
	Incident    []string
	Alias       []string
	HasTags     bool
}

// Parse reads a raw doc-comment body (each line already stripped of its leading
// "//" and one space) and extracts the @tag block. Lines before the first @tag
// are ignored (transitional prose). A line "@name value" opens a tag; subsequent
// lines that do NOT start with "@" are continuation lines appended to the current
// multi-line tag (@description / @boundary).
func Parse(raw string) (*Annotation, error) {
	a := &Annotation{}
	cur := "" // current multi-line tag ("" = none)
	for _, ln := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(ln)
		if strings.HasPrefix(trimmed, "@") {
			name, val := splitTag(trimmed)
			a.HasTags = true
			cur = ""
			switch name {
			case "summary":
				a.Summary = val
			case "description":
				a.Description = val
				cur = "description"
			case "param":
				p, err := parseParam(val)
				if err != nil {
					return nil, err
				}
				a.Params = append(a.Params, p)
			case "returns":
				a.Returns = val
			case "domain":
				a.Domain = val
			case "stability":
				a.Stability = val
			case "verify":
				a.Verify = val
			case "gate":
				a.Gate = splitList(val)
			case "since":
				a.Since = val
			case "boundary":
				a.Boundary = val
				cur = "boundary"
			case "incident":
				a.Incident = splitList(val)
			case "alias":
				a.Alias = splitList(val)
			default:
				return nil, fmt.Errorf("unknown @tag %q", name)
			}
			continue
		}
		if cur == "" || trimmed == "" {
			continue
		}
		switch cur {
		case "description":
			a.Description += "\n" + trimmed
		case "boundary":
			a.Boundary += "\n" + trimmed
		}
	}
	return a, nil
}

// HasLegacyCap reports whether the raw comment carries a legacy aep:cap directive.
func HasLegacyCap(raw string) bool {
	for _, ln := range strings.Split(raw, "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), "aep:cap") {
			return true
		}
	}
	return false
}

func splitTag(line string) (name, val string) {
	line = strings.TrimPrefix(line, "@")
	if i := strings.IndexAny(line, " \t"); i >= 0 {
		return line[:i], strings.TrimSpace(line[i:])
	}
	return line, ""
}

func parseParam(val string) (Param, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return Param{}, fmt.Errorf("@param requires a name")
	}
	name := strings.Fields(val)[0]
	desc := strings.TrimSpace(strings.TrimPrefix(val, name))
	return Param{Name: name, Desc: desc}, nil
}

func splitList(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/apidoc/ -v`
Expected: PASS (`TestParse_FullBlock`, `TestParse_NoTags`, `TestParse_UnknownTag`, `TestHasLegacyCap`).

- [ ] **Step 6: Commit**

```bash
git add internal/apidoc/schema.go internal/apidoc/parse.go internal/apidoc/apidoc_test.go
git commit -m "feat(apidoc): @tag block parser + frozen schema enums (step 1)"
```

---

### Task 2: `internal/apidoc` — validator

**Files:**
- Create: `internal/apidoc/validate.go`
- Test: `internal/apidoc/validate_test.go`

**Interfaces:**
- Consumes: `apidoc.Annotation` (Task 1), `apidoc.Domains/Stabilities/Verifies/SinceVersions`, `has`.
- Produces: `apidoc.Mode` (`ModeWarn`/`ModeStrict`); `apidoc.Symbol{Name, Pos string; Params []string; ReturnsNonError, WriteSurface bool}`; `apidoc.Context{Gates map[string]bool; IncidentsDir string}`; `apidoc.Validate(ann *Annotation, sym Symbol, ctx Context, mode Mode) []error`.

- [ ] **Step 1: Write the failing validator test**

`internal/apidoc/validate_test.go`:

```go
package apidoc

import (
	"strings"
	"testing"
)

func goodAnn() *Annotation {
	return &Annotation{
		Summary: "Add a vector mask to a layer", Domain: "mask", Stability: "stable",
		Verify: "ae-accept", Since: "AE2020",
		Gate:   []string{"TestAddMask_AEShipGate_AE2020"},
		Params: []Param{{"layer", "owning layer reference"}, {"path", "closed Bezier outline path"}},
		Returns: "the created mask", HasTags: true,
	}
}

func goodSym() Symbol {
	return Symbol{Name: "AddMask", Pos: "facade.go:10", Params: []string{"layer", "path"}, ReturnsNonError: true, WriteSurface: true}
}

func ctx() Context {
	return Context{Gates: map[string]bool{"TestAddMask_AEShipGate_AE2020": false}}
}

func msgs(errs []error) string {
	var b strings.Builder
	for _, e := range errs {
		b.WriteString(e.Error())
		b.WriteByte('\n')
	}
	return b.String()
}

func TestValidate_Clean(t *testing.T) {
	if errs := Validate(goodAnn(), goodSym(), ctx(), ModeStrict); len(errs) != 0 {
		t.Fatalf("want clean, got:\n%s", msgs(errs))
	}
}

func TestValidate_SummaryRules(t *testing.T) {
	a := goodAnn()
	a.Summary = strings.Repeat("x", 81)
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "80 chars") {
		t.Errorf("want 80-char error, got:\n%s", msgs(errs))
	}
	a = goodAnn()
	a.Summary = "Adds a mask."
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "period") {
		t.Errorf("want period error, got:\n%s", msgs(errs))
	}
}

func TestValidate_ParamMismatch(t *testing.T) {
	a := goodAnn()
	a.Params = []Param{{"layer", "owning layer reference"}} // missing "path"
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), `missing @param "path"`) {
		t.Errorf("want missing-param error, got:\n%s", msgs(errs))
	}
	a = goodAnn()
	a.Params = append(a.Params, Param{"bogus", "extra param desc here"})
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), `"bogus" is not a parameter`) {
		t.Errorf("want extra-param error, got:\n%s", msgs(errs))
	}
}

func TestValidate_ParamDescRules(t *testing.T) {
	a := goodAnn()
	a.Params[0].Desc = "TODO"
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "TODO") {
		t.Errorf("want TODO error, got:\n%s", msgs(errs))
	}
	a = goodAnn()
	a.Params[0].Desc = "owning layer"
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "3 words") {
		t.Errorf("want 3-word error, got:\n%s", msgs(errs))
	}
}

func TestValidate_Enums(t *testing.T) {
	a := goodAnn()
	a.Domain = "bogus"
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), `invalid @domain "bogus"`) {
		t.Errorf("want domain enum error, got:\n%s", msgs(errs))
	}
	a = goodAnn()
	a.Since = "AE99"
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "@since") {
		t.Errorf("want since error, got:\n%s", msgs(errs))
	}
}

func TestValidate_GateRequiredAndExists(t *testing.T) {
	a := goodAnn()
	a.Gate = nil
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "requires @gate") {
		t.Errorf("want gate-required error, got:\n%s", msgs(errs))
	}
	a = goodAnn()
	a.Gate = []string{"TestNonexistent"}
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "not found") {
		t.Errorf("want gate-existence error, got:\n%s", msgs(errs))
	}
}

func TestValidate_WarnSkipsUntagged(t *testing.T) {
	a := &Annotation{HasTags: false}
	if errs := Validate(a, goodSym(), ctx(), ModeWarn); len(errs) != 0 {
		t.Errorf("warn mode should skip untagged symbol, got:\n%s", msgs(errs))
	}
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "missing @tag") {
		t.Errorf("strict mode should flag untagged write-surface symbol, got:\n%s", msgs(errs))
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/apidoc/ -run TestValidate -v`
Expected: build failure — `undefined: Validate`, `undefined: Symbol`, `undefined: Context`, `undefined: ModeStrict`.

- [ ] **Step 3: Write `validate.go`**

```go
package apidoc

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Mode controls strictness during the migration window.
type Mode int

const (
	// ModeWarn skips completeness for un-converted (no-@tag) symbols so CI stays
	// green during migration. A converted symbol is always fully validated.
	ModeWarn Mode = iota
	// ModeStrict additionally flags an un-converted write-surface symbol as
	// missing its @tag annotation — used at the flip.
	ModeStrict
)

// Symbol carries the signature facts Validate cross-checks an Annotation against.
type Symbol struct {
	Name            string   // for error messages
	Pos             string   // "file:line" for error messages
	Params          []string // non-receiver parameter names, in order
	ReturnsNonError bool     // signature returns >= 1 non-error value
	WriteSurface    bool     // capindex requires() predicate result
}

// Context carries external facts for the gate/incident crosschecks.
type Context struct {
	Gates        map[string]bool // test name -> unconditionally-skipped
	IncidentsDir string          // flightdeck/incidents (for @incident existence)
}

var sinceRe = regexp.MustCompile(`^AE\d{4}$`)

// Validate checks ann against the schema for sym. An un-converted symbol (no
// @tag) yields no errors in ModeWarn (or a single "missing @tag" in ModeStrict
// for a write-surface symbol); a converted symbol is held to the full schema in
// both modes. Each error is prefixed "file:line symbol:".
func Validate(ann *Annotation, sym Symbol, ctx Context, mode Mode) []error {
	var errs []error
	add := func(format string, args ...any) {
		prefix := fmt.Sprintf("%s %s: ", sym.Pos, sym.Name)
		errs = append(errs, fmt.Errorf(prefix+format, args...))
	}

	if !ann.HasTags {
		if mode == ModeStrict && sym.WriteSurface {
			add("missing @tag annotation")
		}
		return errs
	}

	// Completeness (write surface only).
	if sym.WriteSurface {
		for _, m := range []struct {
			ok   bool
			name string
		}{
			{ann.Summary != "", "@summary"},
			{ann.Domain != "", "@domain"},
			{ann.Stability != "", "@stability"},
			{ann.Verify != "", "@verify"},
			{ann.Since != "", "@since"},
		} {
			if !m.ok {
				add("missing %s", m.name)
			}
		}
	}

	// Enums.
	if ann.Domain != "" && !has(Domains, ann.Domain) {
		add("invalid @domain %q", ann.Domain)
	}
	if ann.Stability != "" && !has(Stabilities, ann.Stability) {
		add("invalid @stability %q", ann.Stability)
	}
	if ann.Verify != "" && !has(Verifies, ann.Verify) {
		add("invalid @verify %q", ann.Verify)
	}
	if ann.Since != "" && (!sinceRe.MatchString(ann.Since) || !has(SinceVersions, ann.Since)) {
		add("invalid @since %q (want one of %v)", ann.Since, SinceVersions)
	}

	// @summary format.
	if ann.Summary != "" {
		if strings.Contains(ann.Summary, "\n") {
			add("@summary must be a single line")
		}
		if n := len([]rune(ann.Summary)); n > 80 {
			add("@summary exceeds 80 chars (%d)", n)
		}
		if strings.HasSuffix(ann.Summary, ".") {
			add("@summary must not end with a period")
		}
	}

	validateParams(ann, sym, add)

	// @returns presence (signature returns a non-error value).
	if sym.ReturnsNonError && ann.Returns == "" {
		add("missing @returns (signature returns a non-error value)")
	}

	// @gate presence + existence.
	if (ann.Verify == "ae-accept" || ann.Verify == "render-pixel") && len(ann.Gate) == 0 {
		add("@verify %s requires @gate", ann.Verify)
	}
	for _, g := range ann.Gate {
		skipped, ok := ctx.Gates[g]
		if !ok {
			add("@gate %q not found in any *_test.go", g)
		} else if skipped {
			add("@gate %q is unconditionally skipped (disabled)", g)
		}
	}

	// @incident existence.
	for _, slug := range ann.Incident {
		if ctx.IncidentsDir != "" && !incidentExists(ctx.IncidentsDir, slug) {
			add("@incident %q not found under incidents/", slug)
		}
	}

	return errs
}

func validateParams(ann *Annotation, sym Symbol, add func(string, ...any)) {
	want := map[string]bool{}
	for _, p := range sym.Params {
		want[p] = true
	}
	got := map[string]bool{}
	for _, p := range ann.Params {
		got[p.Name] = true
		if !want[p.Name] {
			add("@param %q is not a parameter of the signature", p.Name)
			continue
		}
		switch {
		case strings.TrimSpace(p.Desc) == "":
			add("@param %q has an empty description", p.Name)
		case strings.Contains(p.Desc, "TODO"):
			add("@param %q description is a TODO placeholder", p.Name)
		case len(strings.Fields(p.Desc)) < 3:
			add("@param %q description must be >= 3 words", p.Name)
		}
	}
	for _, p := range sym.Params {
		if !got[p] {
			add("missing @param %q", p)
		}
	}
}

func incidentExists(dir, slug string) bool {
	candidates := []string{
		filepath.Join(dir, slug+".md"),
		filepath.Join(filepath.Dir(dir), "archive", "incidents", slug+".md"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/apidoc/ -run TestValidate -v`
Expected: PASS (all `TestValidate_*`).

- [ ] **Step 5: Commit**

```bash
git add internal/apidoc/validate.go internal/apidoc/validate_test.go
git commit -m "feat(apidoc): field-rule validator (completeness, enums, @param/signature match, gate/incident crosscheck)"
```

---

### Task 3: `internal/apidoc` — jargon blocklist lint

**Files:**
- Create: `internal/apidoc/jargon.go`
- Test: `internal/apidoc/jargon_test.go`

**Interfaces:**
- Produces: `apidoc.JargonViolation{Line int; Token, Pattern string}`; `apidoc.LintJargon(comment string) []JargonViolation`.

- [ ] **Step 1: Write the failing jargon test**

`internal/apidoc/jargon_test.go`:

```go
package apidoc

import "testing"

func TestLintJargon_Hits(t *testing.T) {
	cases := []string{
		"This was the V2.2 builder transplant",
		"see CLAUDE.md #5 for the rule",
		"derived in tmp_debug/foo",
		"RE'd from tolerance.aep",
		"the wave 2 ldta pass",
		"per the flightdeck cockpit",
		"py-aep golden disagrees",
	}
	for _, c := range cases {
		if v := LintJargon(c); len(v) == 0 {
			t.Errorf("LintJargon(%q) = clean, want a hit", c)
		}
	}
}

func TestLintJargon_Clean(t *testing.T) {
	cases := []string{
		"Appends a closed Bezier mask to the Mask Parade.",
		"AE 2020 and AE 2025 both accept this.", // real user-facing versions stay
		"the tdb4 keyframe layout differs by 3 offsets", // chunk IDs are domain, not jargon
	}
	for _, c := range cases {
		if v := LintJargon(c); len(v) != 0 {
			t.Errorf("LintJargon(%q) = %+v, want clean", c, v)
		}
	}
}

func TestLintJargon_NolintEscape(t *testing.T) {
	if v := LintJargon("compatible back to V2.2 //nolint:jargon"); len(v) != 0 {
		t.Errorf("nolint line should be exempt, got %+v", v)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/apidoc/ -run TestLintJargon -v`
Expected: build failure — `undefined: LintJargon`.

- [ ] **Step 3: Write `jargon.go`**

```go
package apidoc

import (
	"regexp"
	"strings"
)

// jargonPatterns is the auditable blocklist of project-internal codenames and
// process noise forbidden in comments (case-insensitive). This file is the one
// place the blocklist lives; extend it here.
var jargonPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)v2[._]?2`),     // V2.2 / V2_2 / V22
	regexp.MustCompile(`(?i)\bv3\b`),       // V3 milestone codename
	regexp.MustCompile(`(?i)\bm8\b`),       // M8 milestone codename
	regexp.MustCompile(`(?i)wave[ _-]?\d`), // "wave 2"
	regexp.MustCompile(`(?i)phase[ _-]?\d`),// "Phase 0" as a codename
	regexp.MustCompile(`(?i)claude\.md`),
	regexp.MustCompile(`(?i)rules\.md`),
	regexp.MustCompile(`(?i)flightdeck`),
	regexp.MustCompile(`(?i)cockpit`),
	regexp.MustCompile(`(?i)py-?aep`),
	regexp.MustCompile(`(?i)\(probe\)`),
	regexp.MustCompile(`(?i)tmp_debug`),
	regexp.MustCompile(`(?i)re'?d from`),
}

// JargonViolation is one blocklist hit on a comment line.
type JargonViolation struct {
	Line    int
	Token   string
	Pattern string
}

// LintJargon scans comment text (lines joined by \n) for blocklisted tokens. A
// line ending with //nolint:jargon is exempt — the explicit escape for a token
// that genuinely must stay (e.g. a real external-version compatibility note).
func LintJargon(comment string) []JargonViolation {
	var out []JargonViolation
	for i, ln := range strings.Split(comment, "\n") {
		if strings.HasSuffix(strings.TrimSpace(ln), "//nolint:jargon") {
			continue
		}
		for _, re := range jargonPatterns {
			if m := re.FindString(ln); m != "" {
				out = append(out, JargonViolation{Line: i + 1, Token: m, Pattern: re.String()})
			}
		}
	}
	return out
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/apidoc/ -v`
Expected: PASS (parser + validator + jargon).

- [ ] **Step 5: Commit**

```bash
git add internal/apidoc/jargon.go internal/apidoc/jargon_test.go
git commit -m "feat(apidoc): deterministic jargon blocklist lint + //nolint:jargon escape"
```

---

### Task 4: capindex reads `@tag` (dual-read)

**Files:**
- Modify: `cmd/capindex/model.go` (extend `Entry`)
- Modify: `cmd/capindex/extract.go` (`attachCap`, add param/return extraction + `Pos`)
- Modify: `cmd/capindex/tag.go` (add `capFromAnnotation`)
- Test: `cmd/capindex/tag_test.go` (append cases)

**Interfaces:**
- Consumes: `apidoc.Parse`, `apidoc.HasLegacyCap`, `apidoc.Annotation`.
- Produces: `Entry.Params []string`, `Entry.ReturnsNonError bool`, `Entry.Pos string`, `Entry.Ann *apidoc.Annotation`; `capFromAnnotation(*apidoc.Annotation) Cap`.

- [ ] **Step 1: Write the failing dual-read test**

Append to `cmd/capindex/tag_test.go`:

```go
func TestCapFromAnnotation(t *testing.T) {
	a := &apidoc.Annotation{
		Domain: "mask", Stability: "stable", Verify: "ae-accept", Since: "AE2020",
		Gate: []string{"TestAddMask_AEShipGate_AE2020"}, Boundary: "empty parade ok",
		Incident: []string{"add-mask-create-re"}, Alias: []string{"add mask"}, HasTags: true,
	}
	c := capFromAnnotation(a)
	if c.Domain != "mask" || c.Tier != "stable" || c.Verify != "ae-accept" || c.MinVer != "2020" {
		t.Fatalf("cap = %+v", c)
	}
	if len(c.Gate) != 1 || c.Boundary != "empty parade ok" || len(c.Incident) != 1 {
		t.Errorf("cap fields = %+v", c)
	}
	// capFromAnnotation output must still satisfy the existing tier/verify rules.
	if err := validateCap(&c); err != nil {
		t.Errorf("validateCap on @tag-sourced cap: %v", err)
	}
}
```

Add `"github.com/example/aep-parser/internal/apidoc"` to the test file's imports.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./cmd/capindex/ -run TestCapFromAnnotation -v`
Expected: build failure — `undefined: capFromAnnotation`.

- [ ] **Step 3: Extend `Entry` in `model.go`**

Add the import `"github.com/example/aep-parser/internal/apidoc"` and these fields to `Entry`:

```go
	Pos             string                `json:"-"` // "file:line" for --validate messages
	Params          []string              `json:"-"` // non-receiver parameter names
	ReturnsNonError bool                  `json:"-"` // signature returns a non-error value
	Ann             *apidoc.Annotation    `json:"-"` // parsed @tag block (nil for legacy aep:cap)
```

- [ ] **Step 4: Add `capFromAnnotation` to `tag.go`**

```go
// capFromAnnotation maps a parsed @tag Annotation onto the legacy Cap shape so
// the existing crosscheck/render/query code is unchanged. @since "AE2020" maps
// to the bare MinVer "2020".
func capFromAnnotation(a *apidoc.Annotation) Cap {
	return Cap{
		Domain:   a.Domain,
		Tier:     a.Stability,
		Verify:   a.Verify,
		MinVer:   strings.TrimPrefix(a.Since, "AE"),
		Gate:     a.Gate,
		Boundary: a.Boundary,
		Incident: a.Incident,
		Alias:    a.Alias,
	}
}
```

Add `"github.com/example/aep-parser/internal/apidoc"` to `tag.go` imports (it already imports `strings`).

- [ ] **Step 5: Rewrite `attachCap` + add signature extraction in `extract.go`**

Replace `attachCap`:

```go
func attachCap(e *Entry, cg *ast.CommentGroup) {
	raw := rawCommentText(cg)
	ann, err := apidoc.Parse(raw)
	if err != nil {
		e.HasCap, e.parseErr = true, err
		return
	}
	if ann.HasTags {
		if apidoc.HasLegacyCap(raw) {
			e.HasCap = true
			e.parseErr = fmt.Errorf("carries BOTH @tag and legacy aep:cap (single-format invariant)")
			return
		}
		e.HasCap, e.Ann, e.Cap = true, ann, capFromAnnotation(ann)
		return
	}
	c, ok, err := parseCapTag(raw)
	if err != nil {
		e.HasCap, e.parseErr = true, err
		return
	}
	if ok {
		e.HasCap, e.Cap = true, *c
	}
}
```

Add to imports of `extract.go`: `"fmt"` and `"github.com/example/aep-parser/internal/apidoc"`.

In `funcEntry`, after building `e`, capture signature facts + position:

```go
	e.Pos = fmt.Sprintf("%s:%d", filepath.Base(fset.Position(d.Pos()).Filename), fset.Position(d.Pos()).Line)
	e.Params = extractParamNames(d)
	e.ReturnsNonError = returnsNonError(d)
```

(Place these lines before `attachCap(&e, d.Doc)`; `funcEntry` already receives `fset`.) Add helpers:

```go
// extractParamNames returns the non-receiver parameter names of d, in order
// (unnamed params are skipped — they cannot carry an @param).
func extractParamNames(d *ast.FuncDecl) []string {
	var out []string
	if d.Type.Params == nil {
		return out
	}
	for _, f := range d.Type.Params.List {
		for _, n := range f.Names {
			if n.Name != "_" {
				out = append(out, n.Name)
			}
		}
	}
	return out
}

// returnsNonError reports whether d returns at least one result whose type is
// not the builtin error.
func returnsNonError(d *ast.FuncDecl) bool {
	if d.Type.Results == nil {
		return false
	}
	for _, f := range d.Type.Results.List {
		if id, ok := f.Type.(*ast.Ident); ok && id.Name == "error" {
			continue
		}
		return true
	}
	return false
}
```

- [ ] **Step 6: Run the existing + new capindex tests**

Run: `go test ./cmd/capindex/ -v`
Expected: PASS — `TestCapFromAnnotation` plus the whole existing suite (legacy `aep:cap` still parsed; no symbol converted yet, so no behavior change to generated output).

- [ ] **Step 7: Verify generated docs are unchanged**

Run: `go run ./cmd/capindex -check`
Expected: exit 0 (no `out of date` message — nothing converted yet).

- [ ] **Step 8: Commit**

```bash
git add cmd/capindex/model.go cmd/capindex/extract.go cmd/capindex/tag.go cmd/capindex/tag_test.go
git commit -m "feat(capindex): dual-read @tag with aep:cap fallback + single-format invariant"
```

---

### Task 5: `capindex --validate` (warn mode) + CI test

**Files:**
- Create: `cmd/capindex/validate_cmd.go`
- Modify: `cmd/capindex/main.go` (add `-validate` flag)
- Test: `cmd/capindex/validate_cmd_test.go`

**Interfaces:**
- Consumes: `apidoc.Validate`, `apidoc.LintJargon`, `apidoc.Symbol`, `apidoc.Context`, `apidoc.Mode`; `publicSurface.requires` (surface.go); `scanTests` (gatescan.go); `Entry.{Ann,Params,ReturnsNonError,Pos}`.
- Produces: `runValidate(entries []Entry, surface *publicSurface, ctx apidoc.Context, mode apidoc.Mode) []error`.

- [ ] **Step 1: Write the failing `runValidate` test**

`cmd/capindex/validate_cmd_test.go`:

```go
package main

import (
	"strings"
	"testing"

	"github.com/example/aep-parser/internal/apidoc"
)

func TestRunValidate_ConvertedSymbolStrict(t *testing.T) {
	surface := &publicSurface{funcs: map[string]bool{"AddThing": true}}
	good := Entry{
		Symbol: "AddThing", Kind: "func", Pkg: "aep", Pos: "f.go:1",
		Params: []string{"x"}, ReturnsNonError: true,
		Ann: &apidoc.Annotation{
			Summary: "Add a thing to the project", Domain: "structural", Stability: "stable",
			Verify: "roundtrip", Since: "AE2020", Returns: "the created thing",
			Params: []apidoc.Param{{"x", "the thing to add"}}, HasTags: true,
		},
	}
	ctx := apidoc.Context{Gates: map[string]bool{}}
	if errs := runValidate([]Entry{good}, surface, ctx, apidoc.ModeWarn); len(errs) != 0 {
		t.Fatalf("clean converted symbol, got: %v", errs)
	}

	bad := good
	badAnn := *good.Ann
	badAnn.Summary = "Add a thing." // trailing period
	bad.Ann = &badAnn
	errs := runValidate([]Entry{bad}, surface, ctx, apidoc.ModeWarn)
	if len(errs) == 0 || !strings.Contains(errs[0].Error(), "period") {
		t.Fatalf("want period violation even in warn mode (symbol is converted), got: %v", errs)
	}
}

func TestRunValidate_UntaggedWarnVsStrict(t *testing.T) {
	surface := &publicSurface{funcs: map[string]bool{"AddThing": true}}
	untagged := Entry{Symbol: "AddThing", Kind: "func", Pkg: "aep", Pos: "f.go:1"}
	ctx := apidoc.Context{Gates: map[string]bool{}}
	if errs := runValidate([]Entry{untagged}, surface, ctx, apidoc.ModeWarn); len(errs) != 0 {
		t.Errorf("warn mode must ignore untagged surface symbol, got: %v", errs)
	}
	if errs := runValidate([]Entry{untagged}, surface, ctx, apidoc.ModeStrict); len(errs) == 0 {
		t.Error("strict mode must flag untagged surface symbol")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./cmd/capindex/ -run TestRunValidate -v`
Expected: build failure — `undefined: runValidate`.

- [ ] **Step 3: Write `validate_cmd.go`**

```go
package main

import (
	"github.com/example/aep-parser/internal/apidoc"
)

// runValidate validates the parsed entries against the apidoc schema. A converted
// symbol (Ann != nil) is held to the full schema in either mode; an un-converted
// write-surface symbol passes in warn mode and fails only in strict mode. It also
// runs the jargon blocklist over every converted symbol's annotation prose.
func runValidate(entries []Entry, surface *publicSurface, ctx apidoc.Context, mode apidoc.Mode) []error {
	var errs []error
	for _, e := range entries {
		if e.parseErr != nil {
			errs = append(errs, e.parseErr)
			continue
		}
		sym := apidoc.Symbol{
			Name: symbolName(e), Pos: e.Pos, Params: e.Params,
			ReturnsNonError: e.ReturnsNonError, WriteSurface: surface.requires(e),
		}
		ann := e.Ann
		if ann == nil {
			ann = &apidoc.Annotation{} // un-converted: HasTags false
		}
		errs = append(errs, apidoc.Validate(ann, sym, ctx, mode)...)
		if e.Ann != nil {
			for _, v := range apidoc.LintJargon(jargonText(e.Ann)) {
				errs = append(errs, fmtJargon(sym, v))
			}
		}
	}
	return errs
}

func symbolName(e Entry) string {
	if e.Recv != "" {
		return trimStar(e.Recv) + "." + e.Symbol
	}
	return e.Symbol
}

// jargonText joins the human-facing annotation prose for the jargon lint
// (machine fields like @domain/@gate are not prose).
func jargonText(a *apidoc.Annotation) string {
	parts := []string{a.Summary, a.Description, a.Returns, a.Boundary}
	for _, p := range a.Params {
		parts = append(parts, p.Desc)
	}
	return joinNonEmpty(parts, "\n")
}
```

Add small helpers `trimStar`, `joinNonEmpty`, `fmtJargon` (or inline with `strings`); `fmtJargon` returns `fmt.Errorf("%s %s: jargon token %q (pattern %s)", sym.Pos, sym.Name, v.Token, v.Pattern)`.

- [ ] **Step 4: Wire the `-validate` flag in `main.go`**

In `main()`, add alongside the other flags:

```go
	validate := flag.Bool("validate", false, "validate @tag annotations against the schema and exit")
```

After `entries, err := extractEntries(...)` and before the `*coverage` branch, add:

```go
	if *validate {
		surface, err := loadSurface(docgenPath)
		if err != nil {
			fatal(err)
		}
		gates, err := scanTests(filepath.Join(root, "internal"))
		if err != nil {
			fatal(err)
		}
		ctx := apidoc.Context{Gates: gates, IncidentsDir: incidentsDir}
		errs := runValidate(entries, surface, ctx, apidoc.ModeWarn)
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "capindex:", e)
		}
		if len(errs) > 0 {
			os.Exit(1)
		}
		fmt.Println("capindex: @tag validation clean")
		return
	}
```

Add `"github.com/example/aep-parser/internal/apidoc"` to `main.go` imports.

- [ ] **Step 5: Add a CI test that `--validate` (warn mode) passes on the real tree**

Append to `cmd/capindex/validate_cmd_test.go`:

```go
func TestValidate_RealTree_WarnClean(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	entries, err := extractEntries(capindexPkgDirs(root)...)
	if err != nil {
		t.Fatal(err)
	}
	surface, err := loadSurface(filepath.Join(root, "docs", "docgen.json"))
	if err != nil {
		t.Fatal(err)
	}
	gates, err := scanTests(filepath.Join(root, "internal"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := apidoc.Context{Gates: gates, IncidentsDir: filepath.Join(root, "flightdeck", "incidents")}
	if errs := runValidate(entries, surface, ctx, apidoc.ModeWarn); len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("validate: %v", e)
		}
	}
}
```

Add `"path/filepath"` to the test imports.

- [ ] **Step 6: Run the validator tests + the whole suite**

Run: `go test ./cmd/capindex/ -v`
Expected: PASS — including `TestValidate_RealTree_WarnClean` (nothing converted yet, so warn mode is clean).

- [ ] **Step 7: Smoke-run the command**

Run: `go run ./cmd/capindex --validate`
Expected: prints `capindex: @tag validation clean`, exit 0.

- [ ] **Step 8: Commit**

```bash
git add cmd/capindex/validate_cmd.go cmd/capindex/main.go cmd/capindex/validate_cmd_test.go
git commit -m "feat(capindex): --validate (warn mode) + CI test running apidoc.Validate over the surface"
```

---

### Task 6: docgen reads `@tag` + renders the `@param` table

**Files:**
- Modify: `cmd/docgen/model.go` (extend `symbol`)
- Modify: `cmd/docgen/extract.go` (`funcSymbol`, `withMethods`)
- Modify: `cmd/docgen/render.go` (render summary/param/returns)
- Test: `cmd/docgen/render_test.go` + `cmd/docgen/extract_test.go` (append cases)

**Interfaces:**
- Consumes: `apidoc.Parse`, `apidoc.Annotation`, `apidoc.Param`; `rawDoc *ast.CommentGroup` (already captured in `RawFuncDocs`).
- Produces: `symbol.summary string`, `symbol.params []apidoc.Param`, `symbol.returns string`, `symbol.annotated bool`; `renderParams(b, params)` table.

- [ ] **Step 1: Write the failing render test**

Append to `cmd/docgen/render_test.go`:

```go
func TestRenderSymbol_ParamTable(t *testing.T) {
	s := symbol{
		name: "AddMask", kind: kindMethod, signature: "func (l *Layer) AddMask(path BezierPath) (*Mask, error)",
		annotated: true, summary: "Add a vector mask to a layer",
		doc:    "Appends a closed Bezier mask to the layer's Mask Parade.",
		params: []apidoc.Param{{Name: "path", Desc: "outline in layer-pixel space"}},
		returns: "the created mask",
	}
	var b strings.Builder
	renderSymbol(&b, "Layer", s)
	out := b.String()
	for _, want := range []string{
		"Add a vector mask to a layer",                 // summary line
		"Appends a closed Bezier mask",                 // description prose
		"| Parameter | Description |",                  // param table header
		"| `path` | outline in layer-pixel space |",    // param row
		"**Returns:** the created mask",                // returns line
	} {
		if !strings.Contains(out, want) {
			t.Errorf("render missing %q in:\n%s", want, out)
		}
	}
}
```

Add `"github.com/example/aep-parser/internal/apidoc"` to the test imports.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./cmd/docgen/ -run TestRenderSymbol_ParamTable -v`
Expected: build failure — `symbol` has no field `annotated`/`summary`/`params`/`returns`.

- [ ] **Step 3: Extend `symbol` in `model.go`**

Add the import `"github.com/example/aep-parser/internal/apidoc"` and these fields to `symbol`:

```go
	summary   string          // @summary line (annotated symbols)
	params    []apidoc.Param  // @param entries (annotated symbols)
	returns   string          // @returns line (annotated symbols)
	annotated bool            // true when the comment carried a @tag block
```

- [ ] **Step 4: Parse `@tag` in `extract.go`**

Add a helper that turns a raw comment group into annotation-derived symbol fields, and call it from both `funcSymbol` and the method loop in `withMethods`:

```go
// applyAnnotation parses a raw doc-comment group for @tag fields. When present,
// it fills the symbol's summary/description/params/returns from the annotation;
// otherwise it falls back to the legacy directive-stripped prose.
func applyAnnotation(s *symbol, rawDoc *ast.CommentGroup) {
	ann, err := apidoc.Parse(rawCommentOf(rawDoc))
	if err != nil || !ann.HasTags {
		s.doc = directiveStrippedText(rawDoc)
		return
	}
	s.annotated = true
	s.summary = ann.Summary
	s.doc = ann.Description
	s.params = ann.Params
	s.returns = ann.Returns
}

// rawCommentOf reconstructs comment text WITHOUT go/doc directive stripping, so
// `// @tag` lines survive for apidoc.Parse (mirrors capindex's rawCommentText).
func rawCommentOf(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	var b strings.Builder
	for _, c := range cg.List {
		t := strings.TrimPrefix(c.Text, "//")
		t = strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(t, "/*"), "*/"), "")
		b.WriteString(strings.TrimPrefix(t, " "))
		b.WriteByte('\n')
	}
	return b.String()
}
```

In `funcSymbol`, replace `doc: directiveStrippedText(rawDoc),` with a post-construction `applyAnnotation(s, rawDoc)`:

```go
func funcSymbol(lp *loadedPackage, fn *doc.Func) *symbol {
	rawDoc := lp.RawFuncDocs[fn.Decl.Pos()]
	s := &symbol{
		name:      fn.Name,
		kind:      kindMethod,
		signature: normalizeSignature(lp.Fset, fn.Decl),
	}
	applyAnnotation(s, rawDoc)
	return s
}
```

In `withMethods`, where the method `sym` is built (the `for _, fn := range ty.Methods` loop), replace `doc: directiveStrippedText(rawDoc),` with `applyAnnotation(&sym, rawDoc)` after constructing `sym` (keep `name` + `signature`).

Add `"github.com/example/aep-parser/internal/apidoc"` to `extract.go` imports.

- [ ] **Step 5: Render summary/param/returns in `render.go`**

In `renderSymbol`, after the signature code-fence and before the existing `renderProse(s.doc)` block, add the summary; after the prose, add the param table + returns:

```go
	if s.summary != "" {
		fmt.Fprintf(b, "\n%s\n", s.summary)
	}
	if p := renderProse(s.doc); p != "" {
		fmt.Fprintf(b, "\n%s\n", p)
	}
	renderParams(b, s.params)
	if s.returns != "" {
		fmt.Fprintf(b, "\n**Returns:** %s\n", s.returns)
	}
```

(Remove the now-duplicated standalone `renderProse(s.doc)` block that previously sat here.) Add the table renderer:

```go
// renderParams emits a markdown parameter table (nothing when params is empty).
func renderParams(b *strings.Builder, params []apidoc.Param) {
	if len(params) == 0 {
		return
	}
	b.WriteString("\n| Parameter | Description |\n|---|---|\n")
	for _, p := range params {
		fmt.Fprintf(b, "| `%s` | %s |\n", p.Name, p.Desc)
	}
}
```

Apply the same summary/param/returns block to `renderFuncs` (package-level functions) right after the signature fence. Add `"github.com/example/aep-parser/internal/apidoc"` to `render.go` imports.

- [ ] **Step 6: Run the docgen tests**

Run: `go test ./cmd/docgen/ -v`
Expected: PASS — `TestRenderSymbol_ParamTable` plus the existing suite (un-converted symbols still render legacy prose via the fallback, so `docs/*.gen.md` output is unchanged for now).

- [ ] **Step 7: Verify generated docs unchanged**

Run: `go run ./cmd/docgen -manifest docs/docgen.json` then `git diff --stat docs/`
Expected: no changes (nothing converted yet — fallback preserves legacy prose). If any `docs/*.gen.md` changed, investigate before continuing.

- [ ] **Step 8: Commit**

```bash
git add cmd/docgen/model.go cmd/docgen/extract.go cmd/docgen/render.go cmd/docgen/render_test.go
git commit -m "feat(docgen): read @summary/@description/@param/@returns + render param table (legacy prose fallback)"
```

---

### Task 7: docgen self-validates before writing

**Files:**
- Modify: `cmd/docgen/main.go` (`generateFile` → validate before write)
- Create: `cmd/docgen/validate.go` (the per-symbol validation gather)
- Test: `cmd/docgen/main_test.go` (append a case)

**Interfaces:**
- Consumes: `apidoc.Parse`, `apidoc.Validate`, `apidoc.Symbol`, `apidoc.Context`, `apidoc.ModeWarn`; the loaded packages' AST.
- Produces: `validateAnnotations(lps []*loadedPackage) []error` — validates every annotated func/method's `@param`/`@summary`/`@returns`/enum well-formedness (WriteSurface=false: docgen checks well-formedness, capindex owns completeness).

- [ ] **Step 1: Write the failing self-validation test**

Append to `cmd/docgen/main_test.go` (create the file if absent, `package main`):

```go
func TestValidateAnnotations_RealTree(t *testing.T) {
	dirs := []string{
		"../../internal/aep", "../../internal/scene", "../../internal/codec",
	}
	lps, err := loadPackages(dirs)
	if err != nil {
		t.Fatal(err)
	}
	if errs := validateAnnotations(lps); len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("docgen self-validate: %v", e)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./cmd/docgen/ -run TestValidateAnnotations_RealTree -v`
Expected: build failure — `undefined: validateAnnotations`.

- [ ] **Step 3: Write `validate.go`**

```go
package main

import (
	"fmt"
	"go/ast"
	"go/doc"

	"github.com/example/aep-parser/internal/apidoc"
)

// validateAnnotations runs apidoc.Validate over every annotated package-level
// func + type method in the loaded packages, in warn mode with WriteSurface
// false: docgen checks that what it is about to RENDER is well-formed
// (@param matches the signature, @summary length/format, enum legality), while
// capindex --validate owns capability-completeness. Un-annotated symbols are
// skipped (legacy prose path).
func validateAnnotations(lps []*loadedPackage) []error {
	var errs []error
	ctx := apidoc.Context{} // no gate/incident crosscheck here (capindex owns it)
	check := func(lp *loadedPackage, fn *doc.Func) {
		raw := rawCommentOf(lp.RawFuncDocs[fn.Decl.Pos()])
		ann, err := apidoc.Parse(raw)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %v", fn.Name, err))
			return
		}
		if !ann.HasTags {
			return
		}
		sym := apidoc.Symbol{
			Name:            fn.Name,
			Pos:             posOf(lp, fn.Decl),
			Params:          declParamNames(fn.Decl),
			ReturnsNonError: declReturnsNonError(fn.Decl),
			WriteSurface:    false,
		}
		errs = append(errs, apidoc.Validate(ann, sym, ctx, apidoc.ModeWarn)...)
	}
	for _, lp := range lps {
		for _, fn := range lp.Doc.Funcs {
			check(lp, fn)
		}
		for _, ty := range lp.Doc.Types {
			for _, fn := range ty.Funcs {
				check(lp, fn)
			}
			for _, fn := range ty.Methods {
				check(lp, fn)
			}
		}
	}
	return errs
}

func posOf(lp *loadedPackage, d *ast.FuncDecl) string {
	p := lp.Fset.Position(d.Pos())
	return fmt.Sprintf("%s:%d", p.Filename, p.Line)
}

func declParamNames(d *ast.FuncDecl) []string {
	var out []string
	if d.Type.Params == nil {
		return out
	}
	for _, f := range d.Type.Params.List {
		for _, n := range f.Names {
			if n.Name != "_" {
				out = append(out, n.Name)
			}
		}
	}
	return out
}

func declReturnsNonError(d *ast.FuncDecl) bool {
	if d.Type.Results == nil {
		return false
	}
	for _, f := range d.Type.Results.List {
		if id, ok := f.Type.(*ast.Ident); ok && id.Name == "error" {
			continue
		}
		return true
	}
	return false
}
```

- [ ] **Step 4: Gate generation on validation in `main.go`**

In `generateFile`, after `lps, err := loadPackages(dirs)` succeeds, add:

```go
	if errs := validateAnnotations(lps); len(errs) > 0 {
		var b strings.Builder
		for _, e := range errs {
			fmt.Fprintf(&b, "  %v\n", e)
		}
		return "", fmt.Errorf("docgen: annotation validation failed before writing:\n%s", b.String())
	}
```

(`generateFile` already imports `fmt` + `strings`.)

- [ ] **Step 5: Run the docgen suite**

Run: `go test ./cmd/docgen/ -v`
Expected: PASS — `TestValidateAnnotations_RealTree` (nothing converted, so no annotated symbols → clean) plus the existing suite.

- [ ] **Step 6: Commit**

```bash
git add cmd/docgen/validate.go cmd/docgen/main.go cmd/docgen/main_test.go
git commit -m "feat(docgen): self-validate annotations (apidoc.Validate) before writing any file"
```

---

### Task 8: `tools/debug/tagconvert` — one-shot converter

**Files:**
- Create: `tools/debug/tagconvert/main.go`
- Create: `tools/debug/tagconvert/convert.go`
- Test: `tools/debug/tagconvert/convert_test.go`

**Interfaces:**
- Consumes: `cmd/capindex`'s legacy parse semantics (re-implemented minimally here — the converter is throwaway and must not import a `main` package); `apidoc` enums for sanity.
- Produces: `convertSource(src []byte) ([]byte, error)` — rewrites each exported symbol's `//aep:cap …` + leading prose into a `@tag` block with `@param … TODO` placeholders for each signature param.

- [ ] **Step 1: Write the failing converter test**

`tools/debug/tagconvert/convert_test.go`:

```go
package main

import (
	"strings"
	"testing"
)

func TestConvertSource_Basic(t *testing.T) {
	src := `package aep

// AddThing adds a thing to the project. It does the needful.
//
//aep:cap domain=structural tier=stable verify=roundtrip alias="add thing,加东西"
func AddThing(p *Project, x int) (*Thing, error) { return nil, nil }
`
	out, err := convertSource([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	for _, want := range []string{
		"// @summary",                 // a summary tag was emitted
		"// @param   p ",              // each signature param gets a placeholder
		"// @param   x ",
		"// @returns ",                // non-error return → @returns placeholder
		"// @domain     structural",   // machine fields carried over
		"// @stability  stable",
		"// @verify     roundtrip",
		"// @since      AE2020",        // minver default 2020 → AE2020
		"// @alias      add thing,加东西",
		"TODO",                         // placeholders present for manual fill
	} {
		if !strings.Contains(got, want) {
			t.Errorf("convert missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "aep:cap") {
		t.Errorf("legacy aep:cap not removed:\n%s", got)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./tools/debug/tagconvert/ -run TestConvertSource_Basic -v`
Expected: build failure — `undefined: convertSource`.

- [ ] **Step 3: Write `convert.go`**

Parse the file with `go/parser` (ParseComments), and for each exported `FuncDecl` whose doc group contains an `aep:cap` line: build a `@tag` block (summary placeholder from the legacy first sentence; `@param <name> TODO` per signature param; `@returns TODO` when a non-error result exists; machine fields mapped from the parsed `aep:cap` k/v — `tier`→`@stability` if `stable|alpha` else left for manual, `minver`→`@since AE<minver>`, `boundary`→`@boundary` with a `TODO: translate` marker if it contains non-ASCII). Replace the original doc group text in the source bytes. Key implementation notes for the engineer:

```go
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

// convertSource rewrites every exported func whose doc comment carries an
// aep:cap directive into a @tag block with placeholders. It edits the raw bytes
// (not the AST) to preserve all unrelated formatting; edits are applied
// back-to-front so byte offsets stay valid.
func convertSource(src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "src.go", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	type edit struct{ start, end int; text string }
	var edits []edit
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || !fd.Name.IsExported() || fd.Doc == nil {
			continue
		}
		legacy := capLine(fd.Doc)
		if legacy == "" {
			continue
		}
		block := buildTagBlock(fd, legacy)
		start := fset.Position(fd.Doc.Pos()).Offset
		end := fset.Position(fd.Doc.End()).Offset
		edits = append(edits, edit{start, end, block})
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	out := append([]byte(nil), src...)
	for _, e := range edits {
		out = append(out[:e.start], append([]byte(e.text), out[e.end:]...)...)
	}
	return out, nil
}

// capLine returns the aep:cap directive body (after "aep:cap"), or "".
func capLine(cg *ast.CommentGroup) string {
	for _, c := range cg.List {
		t := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
		if strings.HasPrefix(t, "aep:cap") {
			return strings.TrimSpace(strings.TrimPrefix(t, "aep:cap"))
		}
	}
	return ""
}

// buildTagBlock renders the new // @tag block. Machine fields come from the
// parsed aep:cap k/v; prose/param/returns are TODO placeholders for the manual
// pass. (kvParse + paramNames + hasNonErrReturn are small local helpers.)
func buildTagBlock(fd *ast.FuncDecl, legacy string) string {
	kv := kvParse(legacy)
	var b bytes.Buffer
	fmt.Fprintf(&b, "// @summary    TODO summary for %s\n", fd.Name.Name)
	for _, p := range paramNames(fd) {
		fmt.Fprintf(&b, "// @param   %s TODO describe %s\n", p, p)
	}
	if hasNonErrReturn(fd) {
		b.WriteString("// @returns TODO describe the return value\n")
	}
	emit := func(tag, key string) {
		if v := kv[key]; v != "" {
			fmt.Fprintf(&b, "// @%-10s %s\n", tag, v)
		}
	}
	emit("domain", "domain")
	if t := kv["tier"]; t == "stable" || t == "alpha" {
		fmt.Fprintf(&b, "// @%-10s %s\n", "stability", t)
	} else if t != "" {
		fmt.Fprintf(&b, "// @%-10s TODO map tier=%s\n", "stability", t)
	}
	emit("verify", "verify")
	emit("gate", "gate")
	minver := kv["minver"]
	if minver == "" {
		minver = "2020"
	}
	fmt.Fprintf(&b, "// @%-10s AE%s\n", "since", minver)
	if v := kv["boundary"]; v != "" {
		if isASCII(v) {
			fmt.Fprintf(&b, "// @%-10s %s\n", "boundary", v)
		} else {
			fmt.Fprintf(&b, "// @%-10s TODO translate: %s\n", "boundary", v)
		}
	}
	emit("incident", "incident")
	emit("alias", "alias")
	return strings.TrimRight(b.String(), "\n")
}
```

The engineer adds the small helpers `kvParse` (reuse the tokenize logic from `cmd/capindex/tag.go` — copy `tokenizeKV` minimally), `paramNames`, `hasNonErrReturn`, `isASCII`. The converter is intentionally lossy (everything human is `TODO`); its job is to remove boilerplate, not finish the prose.

- [ ] **Step 4: Write `main.go`**

```go
// Command tagconvert rewrites a Go source file's exported-symbol aep:cap
// directives into @tag blocks with placeholders, for the manual conversion pass.
// One-shot, in place: `go run ./tools/debug/tagconvert internal/aep/facade.go`.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: tagconvert <file.go>")
		os.Exit(2)
	}
	path := os.Args[1]
	src, err := os.ReadFile(path)
	if err != nil {
		fatal(err)
	}
	out, err := convertSource(src)
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("tagconvert: rewrote %s\n", path)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "tagconvert:", err)
	os.Exit(1)
}
```

- [ ] **Step 5: Run the converter test + confirm it builds**

Run: `go test ./tools/debug/tagconvert/ -v && go build ./tools/debug/tagconvert/`
Expected: PASS + clean build.

- [ ] **Step 6: Commit**

```bash
git add tools/debug/tagconvert/
git commit -m "feat(tagconvert): one-shot aep:cap+prose -> @tag block converter with TODO placeholders"
```

---

### Task 9: convert `internal/aep/facade.go` fully (the end-to-end proof)

**Files:**
- Modify: `internal/aep/facade.go` (77 funcs: `aep:cap` → `@tag`, Chinese `boundary` → English `@boundary`, real `@param`/`@returns`/`@summary`/`@description`)
- Regenerate: `docs/capabilities.{json,md}`, `docs/*.gen.md`

This task is gated by validators, not by a single unit test: a clean `capindex --validate` (strict on facade) + green `go test ./...` + a clean docgen regen is the proof.

- [ ] **Step 1: Run the converter over facade.go**

Run: `go run ./tools/debug/tagconvert internal/aep/facade.go`
Expected: `tagconvert: rewrote internal/aep/facade.go` — every func now has a `@tag` block with `TODO` placeholders; no `aep:cap` remains.

Verify: `grep -c aep:cap internal/aep/facade.go` → `0`.

- [ ] **Step 2: Confirm the tree still builds + dual-read still parses**

Run: `go build ./... && go run ./cmd/capindex --validate`
Expected: build clean; `--validate` (warn mode) now reports the facade symbols' `TODO` placeholders + any missing-field violations (these are the manual-pass worklist — expected, non-blocking in warn mode? No: converted symbols are strict). The `TODO` rows ARE the to-do list. Capture the output.

- [ ] **Step 3: Manually finish each facade symbol**

For each of the 77 funcs, replace the placeholders with real content (use the legacy prose + the spec example as the model — see `internal/aep/facade.go:16-24` for `Open`/`FromReader`, `:55` for `NewProject`):
- `@summary` — imperative, ≤ 80 chars, no trailing period, no jargon (e.g. `Parse an .aep file by path`).
- `@description` — the rich prose from the existing comment, de-jargoned (drop `M8`/`V2.2`/`CLAUDE.md` etc.; a genuine RE pointer becomes `@incident <slug>`).
- `@param <name> <desc>` — one per signature param, ≥ 3 words, English, no `TODO`.
- `@returns` — when the func returns a non-error value.
- Translate the ~64 Chinese `@boundary` lines to English.

Work in batches (~10 funcs), re-running `go run ./cmd/capindex --validate` after each batch to drive the violation count to zero.

- [ ] **Step 4: Validate facade strict-clean**

Run: `go run ./cmd/capindex --validate`
Expected: `capindex: @tag validation clean` (warn mode is clean because every converted facade symbol now passes the full schema, and the rest of the tree is still legacy `aep:cap`).

To prove strict on the facade specifically, run the suite (the real-tree warn test) and additionally spot-check with a temporary strict run if desired. The objective gate: zero `--validate` violations + zero `@param … TODO` remaining in facade.go (`grep -c "TODO" internal/aep/facade.go` → `0`).

- [ ] **Step 5: Regenerate docs + capindex outputs**

Run:
```bash
go run ./cmd/docgen -manifest docs/docgen.json
go run ./cmd/capindex
```
Expected: docgen writes the `docs/*.gen.md` files (now with `@param` tables + summaries for facade funcs); capindex writes `docs/capabilities.{json,md}`. Review `git diff docs/` — the facade funcs should gain param tables; capability rows should be byte-identical in meaning (same domain/tier/verify/gate; boundary now English).

- [ ] **Step 6: Full verification**

Run: `go vet ./... && go test ./...`
Expected: PASS across the module (capindex `-check`/generated tests, docgen `docs_uptodate` test, apidoc, tagconvert).

- [ ] **Step 7: Commit (schema conversion — kept separate from any cleanup)**

```bash
git add internal/aep/facade.go docs/
git commit -m "refactor(aep): convert facade.go (77 funcs) to @tag schema — proof of end-to-end pipeline

BREAKING: none (public API unchanged). Doc-comment format only: aep:cap+prose -> @tag.
Chinese boundary fields translated to English @boundary. Step 1 of api-doc-tag-schema."
```

---

## Self-Review

**Spec coverage check** (spec → task):
- *The @tag schema / field table* → Task 1 (parser + structs) + Task 2 (rules). ✓
- *Field rules (the validation schema)* → Task 2 (`Validate`). ✓
- *One-command validation `--validate`* → Task 5. ✓
- *Schema home — one package* → Tasks 1–3 (`internal/apidoc`). ✓ (The cross-package "no enum literal outside apidoc" guard test is deferred to the Step-2 flip — capindex's legacy `tag.go` maps legitimately coexist during dual-read; see Follow-on.)
- *Generation self-validates* → Task 7. ✓
- *Jargon-cleanup rule + deterministic blocklist + //nolint escape* → Task 3 (`LintJargon`); wired into `--validate` in Task 5. Repo-wide application across all `internal/**` is Step 2. ✓ (lint engine done; bulk application follow-on)
- *Tooling changes 1–5* → apidoc (Tasks 1–3), capindex (Tasks 4–5), docgen (Tasks 6–7), converter (Task 8). ✓
- *Migration Step 1 (schema+tooling+facade proof)* → Tasks 1–9. ✓
- *Migration Step 2 (bulk + cleanup + flip)* → Follow-on section (not bite-sized TDD; it is iterative mechanical conversion gated on objective exit criteria). ✓
- *Single-format invariant #7* → Task 4 (`attachCap` errors on both formats). ✓
- *Commit hygiene (separate streams)* → Task 9 commit is schema-only; jargon cleanup commits are separate in Step 2. ✓

**Type consistency:** `apidoc.Annotation`/`apidoc.Param`/`apidoc.Symbol`/`apidoc.Context`/`apidoc.Mode` are defined in Tasks 1–2 and consumed verbatim in Tasks 4–8. `capFromAnnotation` (Task 4) feeds the existing `Cap`/`validateCap` unchanged. `symbol.{summary,params,returns,annotated}` (Task 6) are consumed by `renderSymbol`/`renderParams` in the same task. `rawCommentOf` is defined in Task 6 and reused in Task 7. No name drift.

**Placeholder scan:** no `TBD`/"implement later"/"add error handling" left; every code step shows full code. (The `TODO` literals in Task 8/9 are intentional — they are the converter's placeholder output that the manual pass in Task 9 removes, and the validator rejects them, which is the point.)

---

## Follow-on: Step 2 (bulk convert + jargon cleanup + flip) — not in this plan's TDD scope

Step 2 is iterative mechanical work gated on objective exit criteria, not bite-sized TDD; track it as its own execution arc once Step 1 lands:

1. **Bulk convert the remaining ~25 `aep:cap` files (~400 symbols)** — run `tagconvert` per file, manual-finish each, drive `capindex --validate` to zero per file. Commit per file or small batch (schema-conversion stream).
2. **Repo-wide jargon cleanup** — extend the `LintJargon` invocation to scan all `internal/**/*.go` comments (incl. tests). Fix or `//nolint:jargon` each hit. Commit separately (cleanup stream — never mixed with a `@tag`-conversion commit, per spec § Commit hygiene).
3. **The flip** — gated on the objective exit criteria (spec § validation check #9): remaining `aep:cap` blocks = 0 **AND** `@param … TODO` = 0. Then: remove the legacy `parseCapTag` path from `attachCap`; delete the legacy enum maps in `cmd/capindex/tag.go`; add the "no `@domain`/`@stability`/`@verify`/`@since` string literal outside `internal/apidoc`" guard test; make `capindex --validate` + the jargon lint **required** CI (strict mode, fail on any violation); update the project doc-source-of-truth note (CLAUDE.md / rules.md) to name `@tag` as the single annotation format.

**Open question to resolve before Step 2 bulk conversion** (surfaced, not decided here — do not invent enum values): the legacy `aep:cap` tier set includes `planned`/`missing`/`negative` (negative-finding capabilities), but the frozen `@stability` enum is `stable · alpha` only. facade.go has none of these (Step 1 is unaffected), but some scene-file caps do. Before converting those, decide the mapping — candidates: (a) negative findings move to `@domain meta` + `@verify none` and drop the maturity axis; (b) add a separate `@status` field; (c) keep them as legacy `aep:cap` permanently (the flip's "remaining aep:cap = 0" then needs an explicit allowlist). This is a spec follow-up, not a Step-1 blocker.
