# Index — project-index

## State

This work package tracks the dedicated lookup/search/index layer that sits
outside the core `scene.Project` model.

The current architecture decision from `work/scene-architecture/index.md` is:
keep `Project.Compositions`, `Project.Footage`, `Project.Folders`, and
`Composition.Layers` as the canonical public read model for this phase. Do not
add persistent indexes inside `Project` while callers can mutate those public
slices directly.

This package is the answer to the real performance needs that remain:
recipe compilation, profile/diff/clone matching, global search, and large
corpus learning should build explicit snapshot indexes at workflow boundaries.

## Motivation

The performance problem is not ordinary slice iteration. The problem is repeated
linear lookup inside higher-level loops:

- recipe compile/materialization resolving parent/source/layer references
- profile and diff matching layers, sources, effects, and properties
- clone/replication workflows comparing original and generated projects
- global search for effects, expressions, sources, fonts, and properties
- corpus learning over hundreds or thousands of `.aep` projects

These workflows need stable, reusable lookup structures. They do not need the
core parser object graph to become a mutable indexed store.

## Architecture

Create an index package outside `internal/scene`, tentatively:

- `internal/projectindex`

The index is a snapshot:

- Built from one parsed `*aep.Project`.
- Fast to query for the duration of one workflow.
- Not automatically synchronized after project mutation.
- Rebuild after structural mutation or `aep.Reopen`.

This keeps `Project` simple and AST-like while giving analysis and generation
code the indexing tools they need.

## First Slice API

Single-project lookup index:

```go
type Index struct {
	Project *aep.Project
}

func Build(project *aep.Project) *Index

func (idx *Index) CompositionByID(id uint32) *aep.Composition
func (idx *Index) FootageByID(id uint32) *aep.Footage
func (idx *Index) AVItemByID(id uint32) aep.AVItem
func (idx *Index) LayersByID(id uint32) []*aep.Layer
func (idx *Index) LayersByName(name string) []*aep.Layer
func (idx *Index) LayersBySourceID(sourceID uint32) []*aep.Layer
func (idx *Index) LayersByEffect(matchName string) []*aep.Layer
```

Rules:

- ID lookups for project items preserve existing first-match semantics where the
  current public API already does so.
- Layer names are not unique, so layer-name queries return slices.
- Effect and source lookups are inverted indexes and return all matching layers.
- The index must tolerate nil projects, nil comps, nil layers, and ID `0`.
- The index must not mutate the project.

## Later Slices

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

- effects by match name / display name
- properties by match name / display name
- expressions containing text
- text layers using a font
- layers referencing a source item

This can power CLI search and future project browsers.

### Corpus index

Add a corpus-level index only after the single-project index proves useful:

```go
type Corpus struct {
	Projects []*Index
}
```

Corpus queries should return project path + comp/layer/effect/property evidence,
not raw pointers alone. This is the likely foundation for learning from 1000+
projects.

## Non-Goals

- Do not make `scene.Project` maintain persistent maps in this phase.
- Do not privatize public slices as part of the first index package.
- Do not make the snapshot index concurrency-safe until a caller needs shared
  cross-goroutine access.
- Do not hide duplicate names or duplicate IDs; expose deterministic behavior
  and multi-result APIs where identity is not unique.

## Execution Order

1. Add `internal/projectindex` with single-project snapshot maps and tests.
2. Wire the recipe compiler to use the index in one repeated lookup hotspot.
3. Wire profile/diff expected-profile matching where it currently builds ad hoc
   indexes.
4. Add single-project search APIs once real query shapes are confirmed.
5. Add corpus-level indexing for batch learning after single-project search is
   stable.

## Progress

- [x] Promoted indexing/search into a dedicated work package.
- [x] Decided the index is external and snapshot-based, not persistent inside
      `Project`.
- [ ] Implement `internal/projectindex` first slice.
- [ ] Integrate one recipe lookup hotspot.
- [ ] Integrate one profile/diff lookup hotspot.
- [ ] Design single-project search result schema.
- [ ] Design corpus-level learning/search index.

## Verification

Current planning slice:

- `git diff --check` (CRLF warnings only)
