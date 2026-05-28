# V3 Phase 1 — backrefs shard extraction Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Spec**: [`../specs/2026-05-27-v3-deep-think.md`](../specs/2026-05-27-v3-deep-think.md) § 4 (Recommended starting slice)
**Predecessor**: P2c followup#2 (PASS 259, 2026-05-27)

**Goal**: Detach chunk-ref fields (`*rifx.Chunk`) from `Property` / `Layer` / `Composition` / `Footage` / `Project` / `Keyframe` struct top-level into per-type `back *<type>Backrefs` shards in new `back_*.go` files. Pure refactor — zero public API delta, zero behavior change, byte-identical WriteAEP, PASS 259 unchanged.

**Architecture**: Each scene struct keeps its logical fields at top-level (Dimensions / StaticValue / Keyframes / Name / Type / ...). The chunk-ref fields used for length-preserving writes move into a private `back *<type>Backrefs` shard. Setters / parser / writer code reads chunk refs as `obj.back.tdb4` instead of `obj.tdb4`. Locks in the V3 scene-vs-serialization boundary at the struct shape level. Each backref struct also gains an `opaque map[rifx.ChunkID]*rifx.Chunk` field (CLAUDE.md hard constraint #5) — declared now per spec § 6 Q5, populated by future phases.

**Tech Stack**: Go 1.x, `internal/aep` single-package convention (CLAUDE.md hard constraint #3), `internal/rifx` chunk type.

**Non-goals (Phase 1)**:
- No `aep.Open` / `Project` rename — soft migration (spec § 6 Q1).
- No `opaque` population — field exists but stays nil-map; future Phase 2+ parsers capture unrecognized siblings.
- No behavior change — all 259 tests must pass byte-identical without code edits.
- No new methods, no new public types, no `Stable` API additions.

---

## Pre-flight check

Run from `e:/projects/tools/aep-parser`:

```powershell
go vet ./...                                                              # clean (modernize warnings ignored)
go test -count=1 ./internal/aep/... 2>&1 | Select-String '^--- PASS' | Measure-Object | Select-Object -ExpandProperty Count   # expect 259
go test -count=1 ./internal/aep/... 2>&1 | Select-String '^--- FAIL' | Measure-Object | Select-Object -ExpandProperty Count   # expect 0
git status --short                                                        # expect clean
git branch --show-current                                                 # expect main
```

If any of these don't match expectations, STOP and reconcile with board.md before starting Task 1.

---

## File Structure (created in this plan)

- `internal/aep/back_property.go` (new) — `propertyBackrefs` struct
- `internal/aep/back_keyframe.go` (new) — `keyframeBackrefs` struct
- `internal/aep/back_layer.go` (new) — `layerBackrefs` struct
- `internal/aep/back_composition.go` (new) — `compositionBackrefs` struct
- `internal/aep/back_footage.go` (new) — `footageBackrefs` struct
- `internal/aep/back_project.go` (new) — `projectBackrefs` struct

Files modified per task: see each Task's `Files` block.

**Naming convention (locked here)**: backref struct is `<type>Backrefs` (lower-case type prefix, plural `Backrefs`). Field on parent struct is `back *<type>Backrefs`. Opaque map is `opaque map[rifx.ChunkID]*rifx.Chunk`.

---

## Task 1: `propertyBackrefs` extraction

**Files:**
- Create: `internal/aep/back_property.go`
- Modify: `internal/aep/types_core.go` (Property struct definition, lines 808-873)
- Modify: `internal/aep/property_flags.go` (15 chunk-ref reads)
- Modify: `internal/aep/property_group.go` (2 chunk-ref reads)
- Modify: `internal/aep/write_property.go` (47 chunk-ref reads/writes)
- Modify: `internal/aep/testhelpers_test.go` (4 chunk-ref reads — verify whether these are test-private fixture builders that need updating)
- Modify: `internal/aep/parse_keyframe.go` (uses `prop.bytesPerKF` write site)
- Modify: any other file the compiler flags (whole-package grep next)

**Chunk fields to migrate** (10 total):

```go
// FROM Property struct top-level:
cdat       *rifx.Chunk
ldat       *rifx.Chunk
lhd3       *rifx.Chunk
bytesPerKF int
tdbs       *rifx.Chunk
tdb4       *rifx.Chunk
tdsb       *rifx.Chunk
exprChunk  *rifx.Chunk
tdum       *rifx.Chunk
tduM       *rifx.Chunk
```

- [ ] **Step 1: Create the backrefs struct file**

Write `internal/aep/back_property.go`:

```go
package aep

import "github.com/example/aep-parser/internal/rifx"

// propertyBackrefs holds the rifx.Chunk references that power Property's
// length-preserving write paths (SetStaticValue / SetExpression / per-keyframe
// setters). Lives in a separate shard so V3 scene types can mutate logical
// fields without dragging serialization state through every accessor.
//
// Lifecycle:
//   - Populated by parseLeafProperty when a Property is built from a parsed
//     .aep file.
//   - Nil for properties built outside the parser (NewProject builders set
//     it explicitly per their archetype).
//   - opaque captures parsed-but-undecoded sibling chunks under the property's
//     owning tdbs/tdgp container; serializer re-emits them in original order
//     to satisfy CLAUDE.md hard constraint #5 (opaque preservation). Populated
//     by future V3 phases; nil-map in Phase 1.
type propertyBackrefs struct {
	tdbs      *rifx.Chunk
	tdb4      *rifx.Chunk
	cdat      *rifx.Chunk
	ldat      *rifx.Chunk
	lhd3      *rifx.Chunk
	tdsb      *rifx.Chunk
	tdum      *rifx.Chunk
	tduM      *rifx.Chunk
	exprChunk *rifx.Chunk

	// bytesPerKF mirrors the lhd3 @0x10 keyframe stride (decoded once during
	// parse, cached so write paths don't re-read the header on every setter).
	bytesPerKF int

	opaque map[rifx.ChunkID]*rifx.Chunk
}
```

- [ ] **Step 2: Remove chunk fields from Property struct, add `back`**

Edit `internal/aep/types_core.go` Property struct (currently lines 808-873). Delete these 10 field declarations:

```go
cdat       *rifx.Chunk
ldat       *rifx.Chunk
lhd3       *rifx.Chunk
bytesPerKF int
tdbs       *rifx.Chunk
tdb4       *rifx.Chunk
tdsb       *rifx.Chunk
exprChunk  *rifx.Chunk
tdum       *rifx.Chunk
tduM       *rifx.Chunk
```

Replace with a single field at the same location (preserve surrounding comments; move them to the new file if they explain serialization invariants rather than logical semantics):

```go
// back holds the underlying RIFX chunk refs that power length-preserving
// writes. nil for properties built outside the parser. See back_property.go.
back *propertyBackrefs
```

- [ ] **Step 3: Migrate parser write sites**

Inside `internal/aep/parse_property.go` (or wherever `parseLeafProperty` lives — grep first), find every `prop.tdb4 = ...` / `prop.cdat = ...` / etc. line. Before the first assignment, insert:

```go
if prop.back == nil {
    prop.back = &propertyBackrefs{}
}
```

Then rewrite each assignment:

```go
prop.tdb4 = tdb4Chunk        →  prop.back.tdb4 = tdb4Chunk
prop.cdat = cdatChunk        →  prop.back.cdat = cdatChunk
prop.ldat = ldatChunk        →  prop.back.ldat = ldatChunk
prop.lhd3 = lhd3Chunk        →  prop.back.lhd3 = lhd3Chunk
prop.bytesPerKF = bpk        →  prop.back.bytesPerKF = bpk
prop.tdbs = tdbsChunk        →  prop.back.tdbs = tdbsChunk
prop.tdsb = tdsbChunk        →  prop.back.tdsb = tdsbChunk
prop.exprChunk = exprChunk   →  prop.back.exprChunk = exprChunk
prop.tdum = tdumChunk        →  prop.back.tdum = tdumChunk
prop.tduM = tduMChunk        →  prop.back.tduM = tduMChunk
```

To find them all, run from repo root:

```powershell
go build ./internal/aep/ 2>&1
```

Each `prop.<field> undefined` error points to a line needing migration.

- [ ] **Step 4: Migrate reader/setter sites — property_flags.go**

`internal/aep/property_flags.go` has 15 chunk-ref reads. Each one looks like one of:

```go
if p.tdb4 == nil { ... }      →  if p.back == nil || p.back.tdb4 == nil { ... }
data := p.tdb4.Data           →  data := p.back.tdb4.Data
p.tdsb.Data[...] |= ...       →  p.back.tdsb.Data[...] |= ...
```

For methods that already check `p.tdb4 == nil` first, the simplest mechanical replacement: replace `p.tdb4` with `p.back.tdb4`, and the nil-check becomes a chained `if p.back == nil || p.back.tdb4 == nil`.

- [ ] **Step 5: Migrate writer sites — write_property.go**

`internal/aep/write_property.go` has 47 chunk-ref reads/writes. Same mechanical rewrite. Most are inside methods receiving `*Property`; rename `p.cdat` → `p.back.cdat`, `p.ldat` → `p.back.ldat`, etc.

Guard nil-back at method entry only when the method is reachable from a path that could see a parser-less property:

```go
func (p *Property) SetStaticValue(v any) error {
    if p.back == nil {
        return errors.New("property has no backing chunks (constructed outside parser)")
    }
    if p.back.cdat == nil {
        return errors.New("property has no static-value chunk")
    }
    // ... rest unchanged, with p.cdat → p.back.cdat
}
```

(Today's code already has the second check; just lift the first check to top-of-method.)

- [ ] **Step 6: Migrate property_group.go (2 reads)**

Look at the 2 occurrences in `internal/aep/property_group.go`. Mechanical rewrite — `p.tdgp` style refs become `p.back.<field>`.

- [ ] **Step 7: Migrate parse_keyframe.go**

Find the `bytesPerKF` write site (one line). Update:

```go
prop.bytesPerKF = ...  →  prop.back.bytesPerKF = ...   (after the nil-init guard added in Step 3)
```

- [ ] **Step 8: Migrate testhelpers_test.go (4 reads)**

These are test-only fixture builders that construct `&Property{...}` directly with chunk fields. Two options:
- (a) Update the literal to `&Property{back: &propertyBackrefs{cdat: ..., ldat: ...}}` if the test relied on those chunks being non-nil.
- (b) Leave the test untouched if it doesn't reference those fields after construction; just delete the lines that assigned chunk-ref fields.

Pick (a) when the test exercises a writer path; (b) when the test reads logical fields only.

- [ ] **Step 9: Compile-check**

```powershell
go build ./internal/aep/
```

Expected output: no errors. If `prop.<field> undefined` errors remain, repeat the migration for the offending file.

- [ ] **Step 10: Run vet**

```powershell
go vet ./...
```

Expected: clean (existing modernize warnings tolerated, no new warnings).

- [ ] **Step 11: Run full test suite**

```powershell
go test -count=1 ./internal/aep/...
```

Expected: `ok ... aep` with no FAILs. PASS count unchanged from baseline (259).

If any test fails, diagnose root cause — most likely a nil-back guard missed in a setter path that test exercises. Don't paper over with a broader nil check; trace the call chain to the specific accessor that needs back-init.

- [ ] **Step 12: Commit**

```bash
git add internal/aep/back_property.go internal/aep/types_core.go internal/aep/property_flags.go internal/aep/property_group.go internal/aep/write_property.go internal/aep/parse_keyframe.go internal/aep/parse_property.go internal/aep/testhelpers_test.go
git commit -m "refactor(aep): V3 Phase 1 — extract Property chunk refs into propertyBackrefs shard"
```

(Adjust `git add` set to whatever files the migration actually touched.)

---

## Task 2: `keyframeBackrefs` extraction

**Files:**
- Create: `internal/aep/back_keyframe.go`
- Modify: `internal/aep/types_core.go` (Keyframe struct lines 1023-1042)
- Modify: `internal/aep/frame_time_accessors.go` (3 reads)
- Modify: `internal/aep/parse_keyframe.go` (1 write)
- Modify: `internal/aep/write_keyframe.go` (41 reads/writes)

**Fields to migrate** (5 total, all under Keyframe):

```go
ldat     *rifx.Chunk
offset   int
dims     int
tickRate float64
compFps  float64
```

- [ ] **Step 1: Create the backrefs struct**

Write `internal/aep/back_keyframe.go`:

```go
package aep

import "github.com/example/aep-parser/internal/rifx"

// keyframeBackrefs holds the rifx.Chunk reference plus the cached layout
// metadata (offset within ldat.Data, dimensionality, owning comp's TickRate
// and FrameRate) that powers Keyframe.SetTime / SetValue / SetFrameTime
// in-place writes.
//
// Cached fields (offset / dims / tickRate / compFps) duplicate state that
// could be re-derived (offset from Property.lhd3 bpk + keyframe index;
// dims from Property.Components; rates from owning composition). They're
// here because the parser already computed them once per keyframe block.
//
// Nil for keyframes built outside the parser.
type keyframeBackrefs struct {
	ldat     *rifx.Chunk
	offset   int
	dims     int
	tickRate float64
	compFps  float64

	opaque map[rifx.ChunkID]*rifx.Chunk
}
```

- [ ] **Step 2: Replace Keyframe struct fields**

Edit `internal/aep/types_core.go` Keyframe struct. Delete the 5 lines:

```go
ldat     *rifx.Chunk
offset   int
dims     int
tickRate float64
compFps  float64
```

Add:

```go
// back holds the underlying RIFX chunk ref + cached layout metadata that
// power length-preserving keyframe writes. Nil for keyframes built outside
// the parser. See back_keyframe.go.
back *keyframeBackrefs
```

- [ ] **Step 3: Migrate parse_keyframe.go (1 write)**

Find the keyframe construction site. Currently sets `k.ldat = ldat; k.offset = offset; ...`. Replace the block with:

```go
k.back = &keyframeBackrefs{
    ldat:     ldat,
    offset:   offset,
    dims:     dims,
    tickRate: tickRate,
    compFps:  compFps,
}
```

- [ ] **Step 4: Migrate write_keyframe.go (41 reads)**

Every `k.ldat` / `k.offset` / `k.dims` / `k.tickRate` / `k.compFps` becomes `k.back.ldat` / `k.back.offset` / etc.

For setters that currently early-return when chunk is nil:

```go
func (k *Keyframe) SetTime(t float64) error {
    if k.ldat == nil { return errors.New("...") }
    ...
}
```

becomes:

```go
func (k *Keyframe) SetTime(t float64) error {
    if k.back == nil || k.back.ldat == nil { return errors.New("...") }
    ...
}
```

- [ ] **Step 5: Migrate frame_time_accessors.go (3 reads)**

Same pattern. `k.tickRate`, `k.compFps`, `k.ldat` → through `k.back`.

- [ ] **Step 6: Compile + vet + test**

```powershell
go build ./internal/aep/
go vet ./...
go test -count=1 ./internal/aep/...
```

Expected: clean build, clean vet, PASS=259.

- [ ] **Step 7: Commit**

```bash
git add internal/aep/back_keyframe.go internal/aep/types_core.go internal/aep/frame_time_accessors.go internal/aep/parse_keyframe.go internal/aep/write_keyframe.go
git commit -m "refactor(aep): V3 Phase 1 — extract Keyframe chunk ref + layout cache into keyframeBackrefs"
```

---

## Task 3: `layerBackrefs` extraction

**Files:**
- Create: `internal/aep/back_layer.go`
- Modify: `internal/aep/types_core.go` (Layer struct lines 503-642)
- Modify: `internal/aep/parse_layer.go` (10 reads/writes)
- Modify: `internal/aep/parse_composition.go` (1 read)
- Modify: `internal/aep/layer_accessors.go` (1 read)
- Modify: `internal/aep/new_layer.go` (1 read)
- Modify: `internal/aep/sync_shape_layers.go` (2 reads)
- Modify: `internal/aep/write_layer.go` (83 reads — biggest single file)
- Modify: `internal/aep/write_composition.go` (2 reads)
- Modify: `internal/aep/write_text.go` (9 reads)
- Modify: `internal/aep/frame_time_accessors.go` (2 reads)
- Modify: `internal/aep/write_layer_lightsource_test.go` (2 reads — verify whether these are test fixture builders)

**Fields to migrate** (6 chunk refs on Layer):

```go
ldta                *rifx.Chunk
nameChunk           *rifx.Chunk
commentChunk        *rifx.Chunk
layrList            *rifx.Chunk
btdsChunk           *rifx.Chunk
alternateSourceBlsi *rifx.Chunk
```

NOT migrated (these are scene-side / runtime, keep on Layer top-level):
- `comp *Composition` — scene back-pointer
- `shapeRootGroup`, `shapeTransform`, `shapeDirty` — runtime shape state
- `propertyTree *AEPropertyGroup` — scene-level tree mirror
- `TextSourceRaw`, `TextSource` — public logical fields

- [ ] **Step 1: Create back_layer.go**

```go
package aep

import "github.com/example/aep-parser/internal/rifx"

// layerBackrefs holds the rifx.Chunk references that power Layer's
// length-preserving write paths (SetName / SetComment / SetVisible /
// SetBlendingMode / SetAlternateSource / per-text-run setters).
//
// Nil for layers built outside the parser. Setters that mutate length-
// preserving bytes guard `if l.back == nil || l.back.ldta == nil`.
type layerBackrefs struct {
	ldta                *rifx.Chunk
	nameChunk           *rifx.Chunk
	commentChunk        *rifx.Chunk
	layrList            *rifx.Chunk
	btdsChunk           *rifx.Chunk
	alternateSourceBlsi *rifx.Chunk

	opaque map[rifx.ChunkID]*rifx.Chunk
}
```

- [ ] **Step 2: Replace Layer struct fields**

Delete the 6 chunk-ref declarations from Layer struct. Replace with:

```go
// back holds the underlying RIFX chunk refs that power length-preserving
// writes. Nil for layers built outside the parser. See back_layer.go.
back *layerBackrefs
```

- [ ] **Step 3: Migrate parse_layer.go (10 sites)**

At the Layer construction site, initialize:

```go
layer.back = &layerBackrefs{
    ldta:                ldta,
    nameChunk:           nameChunk,
    commentChunk:        commentChunk,  // may be nil — OK
    layrList:            layrList,
    btdsChunk:           btdsChunk,     // may be nil for non-text layers
    alternateSourceBlsi: blsi,          // may be nil
}
```

Replace per-field assignments throughout the file.

- [ ] **Step 4: Migrate write_layer.go (83 sites)**

Every `l.ldta` / `l.nameChunk` / `l.commentChunk` / `l.layrList` / `l.btdsChunk` / `l.alternateSourceBlsi` → through `l.back`.

For methods like `SetVisible`:

```go
func (l *Layer) SetVisible(v bool) error {
    if l.ldta == nil { return errSetterNotSupported }
    ...
}
```

becomes:

```go
func (l *Layer) SetVisible(v bool) error {
    if l.back == nil || l.back.ldta == nil { return errSetterNotSupported }
    ...
}
```

Use Edit's `replace_all` for the high-frequency receiver-prefix substitutions when the prefix is unambiguous:

```
old: l.ldta
new: l.back.ldta
```

But re-read the file afterward — `replace_all` over `l.ldta` does NOT touch struct field declarations or comments, so verify no stray `// l.ldta` line got broken. Run `go build` after each replace.

- [ ] **Step 5: Migrate the remaining files (sync_shape_layers.go, write_composition.go, write_text.go, frame_time_accessors.go, parse_composition.go, layer_accessors.go, new_layer.go)**

Same mechanical rewrite. The grep counts above tell you what scope each file needs.

- [ ] **Step 6: Compile + vet + test**

```powershell
go build ./internal/aep/
go vet ./...
go test -count=1 ./internal/aep/...
```

Expected: PASS=259, vet clean.

- [ ] **Step 7: Commit**

```bash
git add -A internal/aep/
git commit -m "refactor(aep): V3 Phase 1 — extract Layer chunk refs into layerBackrefs shard"
```

(`git add -A internal/aep/` is OK here because all changes are part of the same refactor; verify `git status` shows only the expected files before committing.)

---

## Task 4: `compositionBackrefs` extraction

**Files:**
- Create: `internal/aep/back_composition.go`
- Modify: `internal/aep/types_core.go` (Composition struct lines 296-404)
- Modify: `internal/aep/parse.go` (6 reads — likely the `root` field, double-check via grep before counting)
- Modify: `internal/aep/parse_composition.go` (1 read)
- Modify: `internal/aep/new_composition.go` (3 reads)
- Modify: `internal/aep/new_layer.go` (15 reads)
- Modify: `internal/aep/sync_shape_layers.go` (1 read)
- Modify: `internal/aep/write_composition.go` (69 reads)
- Modify: `internal/aep/write_item.go` (4 reads)
- Modify: `internal/aep/testhelpers_test.go` (2 reads)

**Caveat**: the grep counts above for `.cdta` / `.itemList` / etc. include matches inside method names and docstrings (e.g. `cdtaSomething`, `// cdta @0xA8`). Re-grep with word-boundary patterns when verifying scope.

**Fields to migrate** (6 chunk refs on Composition):

```go
cdta           *rifx.Chunk
nameChunk      *rifx.Chunk
itemList       *rifx.Chunk
itemCmtaChunk  *rifx.Chunk
itemIdtaChunk  *rifx.Chunk
itemLayrParent *rifx.Chunk
```

NOT migrated (keep on Composition):
- `proj *Project` — scene back-pointer
- `Layers`, `Markers`, all public logical fields
- `Comment`, `Label` — public

- [ ] **Step 1: Create back_composition.go**

```go
package aep

import "github.com/example/aep-parser/internal/rifx"

// compositionBackrefs holds the rifx.Chunk references that power
// Composition's length-preserving write paths (cdta setters, name change,
// item-level Comment / Label edits, NewComposition's re-parse closed loop).
//
// Nil for compositions built outside the parser.
type compositionBackrefs struct {
	cdta           *rifx.Chunk
	nameChunk      *rifx.Chunk
	itemList       *rifx.Chunk
	itemCmtaChunk  *rifx.Chunk
	itemIdtaChunk  *rifx.Chunk
	itemLayrParent *rifx.Chunk

	opaque map[rifx.ChunkID]*rifx.Chunk
}
```

- [ ] **Step 2: Replace Composition struct fields**

Delete 6 chunk-ref declarations, add:

```go
// back holds the underlying RIFX chunk refs that power length-preserving
// writes. Nil for compositions built outside the parser. See back_composition.go.
back *compositionBackrefs
```

- [ ] **Step 3-6: Migrate construction + write sites**

Mechanical rewrite as in Tasks 1-3. Special consideration: `CdtaRawBytes` is a public method at line 409:

```go
func (c *Composition) CdtaRawBytes() []byte {
    if c.cdta == nil { return nil }
    return c.cdta.Data
}
```

becomes:

```go
func (c *Composition) CdtaRawBytes() []byte {
    if c.back == nil || c.back.cdta == nil { return nil }
    return c.back.cdta.Data
}
```

Signature unchanged → public API preserved.

- [ ] **Step 7: Compile + vet + test**

```powershell
go build ./internal/aep/ && go vet ./... && go test -count=1 ./internal/aep/...
```

Expected: PASS=259.

- [ ] **Step 8: Commit**

```bash
git add -A internal/aep/
git commit -m "refactor(aep): V3 Phase 1 — extract Composition chunk refs into compositionBackrefs"
```

---

## Task 5: `footageBackrefs` extraction

**Files:**
- Create: `internal/aep/back_footage.go`
- Modify: `internal/aep/types_core.go` (Footage struct lines 458-494)
- Modify: `internal/aep/parse_footage.go` (3 reads)
- Modify: `internal/aep/footage_convenience.go` (8 reads)
- Modify: `internal/aep/write.go` (6 reads — verify whether these are footage or project; grep)

**Fields to migrate** (6 chunk refs on Footage):

```go
aliasChunk     *rifx.Chunk
cpthChunk      *rifx.Chunk
sspcChunk      *rifx.Chunk
itemCmtaChunk  *rifx.Chunk
itemIdtaChunk  *rifx.Chunk
itemLayrParent *rifx.Chunk
```

- [ ] **Step 1: Create back_footage.go**

```go
package aep

import "github.com/example/aep-parser/internal/rifx"

// footageBackrefs holds the rifx.Chunk references that power Footage's
// length-preserving write paths (SetPath / SetComment / SetLabel / SSPC
// flag setters).
//
// Nil for footage items built outside the parser.
type footageBackrefs struct {
	aliasChunk     *rifx.Chunk // Pin/Als2/alas — JSON with "fullpath"
	cpthChunk      *rifx.Chunk // legacy Cpth chunk, when present
	sspcChunk      *rifx.Chunk // source-settings chunk
	itemCmtaChunk  *rifx.Chunk
	itemIdtaChunk  *rifx.Chunk
	itemLayrParent *rifx.Chunk

	opaque map[rifx.ChunkID]*rifx.Chunk
}
```

- [ ] **Step 2: Replace Footage struct fields**

Same pattern: delete 6 chunk-refs, add `back *footageBackrefs`.

- [ ] **Step 3: Migrate parser write site**

In `parse_footage.go`, initialize `f.back = &footageBackrefs{...}` at the Footage construction site.

- [ ] **Step 4: Migrate footage_convenience.go + write.go**

Mechanical rewrite. `f.aliasChunk` → `f.back.aliasChunk`, etc.

- [ ] **Step 5: Compile + vet + test**

```powershell
go build ./internal/aep/ && go vet ./... && go test -count=1 ./internal/aep/...
```

Expected: PASS=259.

- [ ] **Step 6: Commit**

```bash
git add internal/aep/back_footage.go internal/aep/types_core.go internal/aep/parse_footage.go internal/aep/footage_convenience.go internal/aep/write.go
git commit -m "refactor(aep): V3 Phase 1 — extract Footage chunk refs into footageBackrefs"
```

---

## Task 6: `projectBackrefs` extraction

**Files:**
- Create: `internal/aep/back_project.go`
- Modify: `internal/aep/types_core.go` (Project struct lines 200-236)
- Modify: `internal/aep/application.go` (1 read: `a.Project.root`)
- Modify: `internal/aep/parse.go` (~9 reads)
- Modify: `internal/aep/new_composition.go` (1 read)
- Modify: `internal/aep/project_settings.go` (74 reads — biggest single file)
- Modify: `internal/aep/project_views.go` (2 reads)
- Modify: `internal/aep/write.go` (11 reads)
- Modify: `internal/aep/testhelpers_test.go` (2 reads)

**Fields to migrate** (9 fields on Project):

```go
root      *rifx.Chunk
nhedChunk *rifx.Chunk
nnhdChunk *rifx.Chunk
acerChunk *rifx.Chunk
adfrChunk *rifx.Chunk
dwgaChunk *rifx.Chunk
gpugUtf8  *rifx.Chunk
exenUtf8  *rifx.Chunk
cmsUtf8   *rifx.Chunk
```

NOT migrated (keep on Project — these are V2 derived scene state, not chunk back-refs):
- `nextItemID uint32` — monotonic ID counter
- `rootFold *rifx.Chunk` — cached root Fold LIST (derived cache, not raw chunk back-ref... but it IS a chunk ref — judgement call)
- `target AETarget` — builder-template selector

**Decision on `rootFold`**: it IS a `*rifx.Chunk`, so for shape consistency it should move into `projectBackrefs`. Keep it there. `nextItemID` and `target` stay on Project (logical state).

So `projectBackrefs` actually has 10 fields:

```go
type projectBackrefs struct {
	root      *rifx.Chunk
	rootFold  *rifx.Chunk
	nhedChunk *rifx.Chunk
	nnhdChunk *rifx.Chunk
	acerChunk *rifx.Chunk
	adfrChunk *rifx.Chunk
	dwgaChunk *rifx.Chunk
	gpugUtf8  *rifx.Chunk
	exenUtf8  *rifx.Chunk
	cmsUtf8   *rifx.Chunk

	opaque map[rifx.ChunkID]*rifx.Chunk
}
```

- [ ] **Step 1: Create back_project.go**

Write the struct above (with full docstring matching other backref files: "Holds the underlying RIFX root + project-level single-field chunk refs ... Nil for projects built outside the parser ...").

- [ ] **Step 2: Replace Project struct fields**

Delete 10 chunk-ref declarations from Project struct. Add:

```go
// back holds the underlying RIFX root + project-level single-field chunk
// refs that power length-preserving writes. Nil for projects built outside
// the parser. See back_project.go.
back *projectBackrefs
```

Keep `Warnings`, `nextItemID`, `target` at top-level.

- [ ] **Step 3: Migrate parser (parse.go)**

In `parse.go`, find where the Project is constructed (`proj := &Project{...}` or similar). After construction, populate:

```go
proj.back = &projectBackrefs{
    root: rootChunk,
    // ... other fields populated below as chunks are discovered
}
```

Then each later `proj.nhedChunk = nhed` becomes `proj.back.nhedChunk = nhed`.

- [ ] **Step 4: Migrate project_settings.go (74 reads)**

This is the biggest file. Use `Edit` with `replace_all` per field for receiver-prefix patterns. Example:

```
old: p.acerChunk
new: p.back.acerChunk
```

After each replace_all, run `go build ./internal/aep/` to catch syntax breakage. The setters in this file mostly look like:

```go
func (p *Project) SetSomething(v T) error {
    if p.acerChunk == nil { return errSetterNotSupported }
    p.acerChunk.Data[0] = byte(v)
    return nil
}
```

After rewrite:

```go
func (p *Project) SetSomething(v T) error {
    if p.back == nil || p.back.acerChunk == nil { return errSetterNotSupported }
    p.back.acerChunk.Data[0] = byte(v)
    return nil
}
```

- [ ] **Step 5: Migrate application.go (1 site)**

`application.go:58-62`:

```go
if a.Project == nil || a.Project.root == nil { return "" }
head := a.Project.root.FindFirst(chunkIDHead)
```

becomes:

```go
if a.Project == nil || a.Project.back == nil || a.Project.back.root == nil { return "" }
head := a.Project.back.root.FindFirst(chunkIDHead)
```

- [ ] **Step 6: Migrate write.go, new_composition.go, project_views.go, testhelpers_test.go**

Mechanical rewrite of remaining 16 reads.

- [ ] **Step 7: Compile + vet + test**

```powershell
go build ./internal/aep/ && go vet ./... && go test -count=1 ./internal/aep/...
```

Expected: PASS=259, vet clean.

- [ ] **Step 8: Commit**

```bash
git add -A internal/aep/
git commit -m "refactor(aep): V3 Phase 1 — extract Project chunk refs into projectBackrefs (final type)"
```

---

## Task 7: Phase 1 acceptance + docs sync

**Files:**
- Modify: `workshop/board.md`
- Modify: `workshop/plans/coverage.md` (optional — add Phase 1 row)
- Verify: no other files

- [ ] **Step 1: Public API delta check**

The Stable public API surface must be byte-for-byte unchanged. Run:

```powershell
go doc -all ./internal/aep/ > /tmp/api-after.txt
```

If there's a baseline `api-before.txt` from before Task 1, diff them. Otherwise, verify by spot-check that none of:

- `aep.Open`, `aep.FromReader`, `aep.Parse`, `aep.ParseReader`
- `Project.WriteAEP`, `Project.WriteJSON`, `Project.Compositions`, `Project.Footage`, `Project.Folders`, `Project.BitsPerChannel`, `Project.Warnings`
- `Composition.{ID,Name,Width,Height,FrameRate,Duration,TickRate,Layers,...,CdtaRawBytes,LayerByID,LayerByName,ActiveCamera}`
- `Layer.{Index,Name,Type,Properties,Effects,Markers,Masks,Position(),Scale(),Opacity(),...,Parent(),SourceComposition(),SourceFootage(),TrackMatteLayer()}`
- `Property.{MatchName,Name,Components,Keyframes,StaticValue,Expression,ExpressionEnabled,DefaultValue,LastValue,NbOptions,Gradient}` + all public methods
- `Footage.{ID,Name,Path,...,IsPlaceholder}` + `Footage.SetPath`
- `Folder.{ID,Name}`
- `Application.{Project,Version}`
- `Keyframe.{Time,Value,InInterp,OutInterp,InSpatialTangent,OutSpatialTangent,InTemporalEase,OutTemporalEase}` + all setters

…changed signature or moved fields. If a method that previously read `p.cdat` now reads `p.back.cdat`, that's an internal change — no public API delta. Confirm.

- [ ] **Step 2: Re-run full test suite for final PASS=259 check**

```powershell
go vet ./...
go test -count=1 ./internal/aep/... -v 2>&1 | Select-String '^--- PASS' | Measure-Object | Select-Object -ExpandProperty Count
go test -count=1 ./internal/aep/... -v 2>&1 | Select-String '^--- FAIL' | Measure-Object | Select-Object -ExpandProperty Count
```

Expected: vet clean, PASS=259, FAIL=0.

- [ ] **Step 3: Byte-identical fixture roundtrip**

The test suite includes `TestRoundtripWrite` and many `Test*Roundtrip` tests that already do byte-identical comparison. If they're all green, byte-identical is verified.

Optional extra: run the RE fixtures directly via the existing roundtrip test harness. Spot-check:

```powershell
go test ./internal/aep/ -run 'TestRoundtripWrite|TestKeyframeRoundtrip|TestSetResolutionFactorRoundtrip' -v
```

Expected: all PASS.

- [ ] **Step 4: Update board.md**

Edit `workshop/board.md`:

- Bump `Last updated` to today's date + describe Phase 1 completion.
- Move "V3 Phase 1 起步" out of `Next session`.
- Add `Recently finished` entry summarizing the 6 backref extractions.
- Set new `Next session` to "V3 Phase 2 candidate: `Composition.DeleteLayer(idx)` — first structural mutation through scene + back-ref split" (per spec § 4 Phase 2 candidate). Add the AE-acceptance-gate dependency note.
- Update `Active focus` to reflect V3 Phase 2 readiness.

- [ ] **Step 5: Commit docs sync**

```bash
git add workshop/board.md
git commit -m "docs(workshop): V3 Phase 1 complete — backrefs extraction done, board.md → Phase 2 (DeleteLayer)"
```

---

## Acceptance criteria (Phase 1 complete)

All four must hold simultaneously after Task 7:

1. `go test -count=1 ./internal/aep/...` reports PASS=259, FAIL=0 (matches pre-Phase-1 baseline)
2. `go vet ./...` is clean (no new warnings vs pre-Phase-1)
3. Each of `Property` / `Layer` / `Composition` / `Footage` / `Project` / `Keyframe` has a `back *<type>Backrefs` field and zero top-level `*rifx.Chunk` / `int bytesPerKF` / `float64 tickRate` / `float64 compFps` / `int offset` / `int dims` fields (verify with grep on `types_core.go`)
4. Public API surface unchanged — every Stable method listed in Task 7 Step 1 exists with the same signature

---

## Risk notes (from spec § 3)

- **Mutation API atomic-rollback edge cases**: Phase 1 doesn't add new mutations, so the V2.1 atomic invariant pattern (`Project.Warnings` rollback) doesn't get exercised differently. Low risk.
- **Public API drift**: highest-risk failure mode is accidentally renaming or moving a Stable field/method. Step 1 of Task 7 mitigates via explicit check.
- **Opaque preservation regression**: Phase 1 adds the `opaque map` field but doesn't populate it. No regression risk — same chunks are written verbatim from the same chunk tree. Future Phase 2+ that synthesizes new chunks must populate opaque to retain AE 2025 ship-gate.
- **Test fixture builders breaking**: `testhelpers_test.go` and `*_test.go` files that build `&Property{...}` literals will fail to compile if they referenced moved fields. The plan calls these out explicitly per task; remaining build errors localize the problem.

---

## Related

- Spec: [`../specs/2026-05-27-v3-deep-think.md`](../specs/2026-05-27-v3-deep-think.md) (§ 4 starting slice, § 6 decision asks)
- Prior plan: [`../plans/2026-05-22-v2-2-layer-creation-plan.md`](2026-05-22-v2-2-layer-creation-plan.md) (V2.2 ShapeLayer — context on Layer struct growth)
- CLAUDE.md hard constraints: #1 (length-preserving), #2 (Stable vs Alpha API), #3 (single package), #5 (opaque preservation), #6 (AE acceptance gate — not exercised in Phase 1 since no new write paths)
- Verify playbook: [`../playbooks/verify.md`](../playbooks/verify.md) (PASS count check + commit-gate commands)
