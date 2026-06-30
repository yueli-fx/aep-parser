# Index — project-index

## State

This work package tracks the dedicated lookup/search/index layer that sits
outside the core `scene.Project` model.

Status: lookup first slice complete, first narrow single-project search slice
implemented, and first corpus fact extraction slice implemented.
`internal/projectindex` exists, recipe/profile hotspots are wired to it, callers
can obtain stable hit locations for source-layer, effect-match, property-match,
expression-substring, and text-font searches, and batch workflows can now
extract durable layer-source/effect-usage/expression/text-style/shape-usage
facts without retaining parsed project graphs. `cmd/aepsearch` is the first
real caller for single-project search and multi-project corpus extraction.
Later work should add broader search/corpus APIs only when a caller needs them.

The current architecture decision from `work/scene-architecture/index.md` is:
keep `Project.Compositions`, `Project.Footage`, `Project.Folders`, and
`Composition.Layers` as the canonical public read model for this phase. Do not
add persistent indexes inside `Project` while callers can mutate those public
slices directly.

Authoritative decision: recipe compilation, profile/diff/clone matching, global
search, and large-corpus pattern extraction should build explicit snapshot
indexes at workflow boundaries.

## Motivation

The performance problem is not ordinary slice iteration. The problem is repeated
linear lookup inside higher-level loops:

- recipe compile/materialization resolving parent/source/layer references
- profile and diff matching layers, sources, effects, and properties
- clone/replication workflows comparing original and generated projects
- global search for effects, expressions, sources, fonts, and properties
- corpus pattern extraction over hundreds or thousands of `.aep` projects

These workflows need stable, reusable lookup structures. They do not need the
core parser object graph to become a mutable indexed store.

## Architecture

Create an index package outside `internal/scene`, tentatively:

- `internal/projectindex`

Terminology:

- **Snapshot** means the index reflects the project graph as observed during
  `Build`. It is not kept in sync with later project mutations.
- **Index** means the concrete lookup object containing maps/inverted indexes.
  Every `projectindex.Index` is a snapshot. Lazy secondary maps built inside the
  same `Index` are part of that same snapshot and must read only from the
  captured project graph.

The single-project index has this lifecycle:

- Built from one parsed `*aep.Project`.
- Fast to query for the duration of one workflow.
- Not automatically synchronized after project mutation.
- Rebuild after structural mutation or `aep.Reopen`.
- The index becomes stale after any structural or semantic mutation that affects
  indexed fields. Examples include adding/removing comps or layers, changing a
  layer name, changing `Layer.SourceID`, changing layer effects, or replacing
  footage identity. Callers are responsible for rebuilding it.
- There is no stale-check API in the first slice because `Project` has no
  mutation version and its public slices/fields are directly writable. Workflow
  owners must rebuild by scope: after they mutate indexed fields or after they
  receive a fresh project from `aep.Reopen`.
- `Build(project *aep.Project) *Index` accepts nil. `Build(nil)` returns a
  non-nil empty index so callers can query safely without nil checks.
- Implementations may build secondary indexes lazily if behavior remains
  deterministic and query results do not mutate the project.
- The index stores pointers to existing project objects, not deep copies.

This keeps `Project` simple and AST-like while giving analysis and generation
code the indexing tools they need.

## First Slice API

Single-project lookup index:

```go
type Index struct {
	project *aep.Project
}

func Build(project *aep.Project) *Index

func (idx *Index) CompositionByID(id uint32) *aep.Composition
func (idx *Index) FootageByID(id uint32) *aep.Footage
func (idx *Index) AVItemByID(id uint32) aep.AVItem
func (idx *Index) LayersByLayerID(id uint32) []*aep.Layer
func (idx *Index) LayersByName(name string) []*aep.Layer
func (idx *Index) LayersBySourceID(sourceID uint32) []*aep.Layer
func (idx *Index) LayersByEffect(matchName string) []*aep.Layer
```

Rules:

- Project item ID lookups are deterministic:
  - `CompositionByID` returns the first matching item in `Project.Compositions`
    slice order.
  - `FootageByID` returns the first matching item in `Project.Footage` slice
    order.
  - `AVItemByID` matches `scene.Project.AVItemByID`: scan compositions first in
    `Project.Compositions` slice order, then footage in `Project.Footage` slice
    order. Folders are not AV items.
- Layer names are not unique, so layer-name queries return slices.
- Layer IDs are not treated as globally unique in the project-index API; use
  `LayersByLayerID` and handle zero or more results.
- Effect and source lookups are inverted indexes and return all matching layers.
  The first slice indexes effects by match name only. Display-name search belongs
  to the later search layer once real query shapes are known.
- The index must tolerate nil projects, nil comps, nil layers, and ID `0`.
  Single-result queries return nil for missing/invalid input. Multi-result
  queries return zero results; callers must not rely on nil-vs-empty slice
  identity.
- The index must not mutate the project.
- The source project pointer is kept private. Query callers should use index
  methods instead of reaching back through `idx.Project`, which would encourage
  stale-index bugs.
- The first slice exposes lookup primitives only. Callers that need project
  fields outside those primitives should keep their own `*aep.Project` reference
  from the original workflow, not retrieve it through the index.
- Building the index is best-effort over the parsed object graph. Malformed or
  partial projects are represented by nil/missing entries, not build errors.
  The first slice does not follow references recursively, so circular references
  are not an index-build error.

## Later Slices

### Single-Project Search Result Schema

Search should be built on top of `Index`, but the result schema should not be a
bare list of pointers. The canonical search result is a stable location plus
the matched value. Single-project callers may also receive pointers for
convenience, but pointers are not the identity of a hit.

First schema:

```go
type HitKind string

const (
	HitProjectItem HitKind = "project_item"
	HitComposition HitKind = "composition"
	HitLayer       HitKind = "layer"
	HitEffect      HitKind = "effect"
	HitProperty    HitKind = "property"
	HitExpression  HitKind = "expression"
	HitTextStyle   HitKind = "text_style"
)

type Hit struct {
	Kind     HitKind
	Match    Match
	Location Location
	Pointers HitPointers
}

type Match struct {
	Field string
	Value string
}

type Location struct {
	ItemID   uint32
	ItemKind string
	ItemName string

	CompID   uint32
	CompName string

	LayerID    uint32
	LayerIndex int
	LayerName  string

	EffectMatchName string
	EffectName      string
	EffectOccurrence int

	PropertyMatchName string
	PropertyName      string
	PropertyPath      string
}

type HitPointers struct {
	Item        aep.AVItem
	Comp        *aep.Composition
	Layer       *aep.Layer
	Effect      *aep.Effect
	Property    *aep.Property
}
```

Rules:

- `Location` is the canonical machine-readable identity of a hit.
- `Pointers` are a single-project convenience only and must not be serialized as
  the corpus format.
- Search results must be ordered by project order: composition slice order,
  layer slice order within the composition, then effect/property occurrence.
- Search APIs should start narrow:
  - layers by source item ID
  - layers/effects by effect match name
  - later: expressions containing text, text layers by font, properties by match
    name/display name
- Display-name search is a search-layer concern, not part of the first
  `LayersByEffect` primitive.
- Corpus search should reuse `Location` and add project path/source evidence
  instead of returning raw pointers.

### Recipe integration

Use the index in recipe workflows where references are repeatedly resolved:

- layer parent names
- layer source names/IDs
- expected-profile matching
- materialization after `aep.Reopen`

The recipe compiler can build one index per compilation phase instead of
scanning the project graph repeatedly.

### Profile and diff integration

Use indexes for:

- source reference resolution
- layer matching by ID/name/index fallback
- effect and property search in expected-profile checks
- clone/diff workflows that compare two project graphs

### Global search

Add a query layer over a single `Index`:

- effects by match name, and later display name when needed
- properties by match name, and later display name when needed
- expressions containing text
- text layers using a font
- layers referencing a source item

This can power CLI search and future project browsers.

### Corpus index

Add a corpus-level index only after the single-project index proves useful.
Corpus-level indexing means extracting reusable evidence/pattern facts from many
projects for search, statistics, clustering, and technique discovery. It is not
an ML training plan by itself.

The corpus layer should not keep `[]*Index` as the primary representation. A
single-project `Index` is useful while one project is being analyzed; corpus
search needs durable facts that can survive after the project graph is released.

```go
type Corpus struct {
	Projects []CorpusProject
	Facts    []CorpusFact
}

type CorpusProject struct {
	ID          string
	Path        string
	Fingerprint string
	CompCount   int
	LayerCount  int
	EffectCount int
}

type CorpusFactKind string

const (
	FactLayerSource CorpusFactKind = "layer_source"
	FactEffectUsage CorpusFactKind = "effect_usage"
	FactExpression  CorpusFactKind = "expression"
	FactTextStyle   CorpusFactKind = "text_style"
	FactShapeUsage  CorpusFactKind = "shape_usage"
)

type CorpusFact struct {
	ProjectID string
	Kind      CorpusFactKind
	Location  Location
	Match     Match
	Summary   string
}

type CorpusBuilder interface {
	AddProject(path string, project *aep.Project) error
	Build() (*Corpus, error)
}
```

Build lifecycle:

1. Open one project.
2. Build a single-project `Index`.
3. Extract facts and aggregate counts/pattern summaries.
4. Release the project and index unless the caller explicitly asks to keep them.
5. Repeat for the next project.

Corpus queries should return project path + comp/layer/effect/property evidence,
not raw pointers alone. This is the likely foundation for learning from 1000+
projects because it can answer questions like:

- which effects and parameter combinations appear together
- which layer/source/precomp structures are common
- which expressions and fonts recur across projects
- which shape/text/matte techniques are common in a category

Memory rules:

- Do not assume 1000 full `Project` graphs plus 1000 full `Index` snapshots
  should stay resident at once.
- Prefer streaming extraction into facts, summarized counters, or persisted
  shards.
- A corpus hit must be reloadable by `Project.Path + Location`, but it must not
  require the original pointer to still be alive.
- If an interactive browser needs live pointers, load one project on demand and
  rebuild its single-project `Index`.

## Non-Goals

- Do not make `scene.Project` maintain persistent maps in this phase.
- Do not privatize public slices as part of the first index package.
- Do not add locks in the first slice. An `Index` is immutable after build and
  can be shared for concurrent read-only queries once safely published. It is
  not safe to mutate the underlying project concurrently with index queries.
- Do not hide duplicate names or duplicate IDs; expose deterministic behavior
  and multi-result APIs where identity is not unique.
- Do not add batch query combinators until a real recipe/profile/search caller
  needs one.
- Do not add stale detection until `Project` has a mutation version or a wrapper
  owns all mutations.

## Execution Order

1. Add `internal/projectindex` with single-project snapshot maps and tests.
2. Wire the recipe compiler to use the index in one repeated lookup hotspot.
3. Wire profile/diff expected-profile matching where it currently builds ad hoc
   indexes.
4. Add benchmarks for repeated lookup hotspots before broad integration.
5. Add single-project search APIs once real query shapes are confirmed.
6. Add corpus-level indexing for batch learning after single-project search is
   stable.

## Progress

- [x] Promoted indexing/search into a dedicated work package.
- [x] Decided the index is external and snapshot-based, not persistent inside
      `Project`.
- [x] Implement `internal/projectindex` first slice.
  - [x] Project item ID lookup with deterministic first-match semantics.
  - [x] Layer inverted lookups by layer ID, name, source ID, and effect match
        name.
  - [x] Build/query benchmarks for the first slice.
- [x] Integrate one recipe lookup hotspot.
  - [x] Parent/light source layer-name resolution now goes through
        `projectindex`, while preserving recipe's old duplicate-name
        last-match behavior.
- [x] Integrate one profile/diff lookup hotspot.
  - [x] Profile layer `source_ref` resolution now goes through
        `projectindex.AVItemByID` instead of ad hoc comp/footage maps.
- [x] Design single-project search result schema.
- [x] Implement first single-project search API slice.
  - [x] `SearchLayersBySourceID` returns stable layer hits with source item
        evidence and pointers for single-project convenience.
  - [x] `SearchEffectsByMatchName` returns stable effect hits with deterministic
        per-layer effect occurrence.
  - [x] `SearchPropertiesByMatchName` returns layer-property and effect-param
        hits with stable property paths.
  - [x] `SearchExpressionsContaining` returns layer-property and effect-param
        expression hits using case-sensitive substring matching.
  - [x] `SearchTextStylesByFont` returns text-style run hits with stable run
        paths and font resolution from either `FontName` or `Fonts[FontIndex]`.
  - [x] Search hits serialize with stable snake_case JSON names and do not
        serialize raw Go pointers.
- [x] Design corpus-level learning/search index.
- [x] Implement first corpus fact extraction slice.
  - [x] `NewCorpusBuilder` ingests projects one at a time and emits
        serializable `CorpusProject` summaries plus `CorpusFact` rows.
  - [x] First facts cover layer source references, effect usage, property
        expressions, text style font usage, and shape primitive/path usage.
  - [x] Corpus JSON contains stable project IDs, locations, matches, and
        summaries, but no `Project`, `Index`, or pointer graph.
- [x] Add first CLI caller.
  - [x] `cmd/aepsearch search` opens one `.aep` and emits stable search hits
        for effect/source/property/expression/font queries.
  - [x] `cmd/aepsearch corpus` opens one or more `.aep` files and emits
        durable corpus facts without exposing live project pointers.

## Verification

Current first implementation slice:

- `go test ./internal/projectindex -count=1`
- `go test ./internal/recipe -run 'TestCompileToFile(Set|LayerParent|ChecksLayerParent|LightSource)' -count=1`
- `go test ./internal/profile -run TestBuildSyntheticProjectIncludesLayerSourceRefs -count=1`
- `go test ./internal/projectindex ./internal/profile -count=1`
- `git diff --check` (CRLF warnings only)

Current search slice:

- `go test ./internal/projectindex -run 'TestSearch' -count=1`
- `go test ./internal/projectindex -count=1`

Current corpus slice:

- `go test ./internal/projectindex -run 'TestCorpus' -count=1`
- `go test ./internal/projectindex -count=1`

Current CLI slice:

- `go test ./cmd/aepsearch -count=1`
- `go test ./...`
- `go vet ./...`
