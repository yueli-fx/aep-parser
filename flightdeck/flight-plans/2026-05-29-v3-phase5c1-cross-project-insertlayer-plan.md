# V3 Phase 5C.1 Cross-Project InsertLayer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Lift `InsertLayer`'s cross-Project refuse (R7) so `destComp.InsertLayer(srcLayer, atIdx)` works when `src` lives in a different `*Project` — importing the source layer's reachable item closure (footage + precomp, transitively) into dest at root level and remapping the inserted layer's `SourceID`/`AlternateSourceID` to the freshly-imported dest item IDs.

**Architecture:** Two phases inside a new cross-Project branch. (1) Closure import: BFS over the source layer's item refs, deep-cloning footage/comp Item blocks into dest `rootFold` with fresh item IDs, dedup'ing file-backed footage by `Path`, building a `srcItemID→destItemID` map. (2) Layer splice: the existing same-Project clone logic, extracted into a shared `spliceLayerClone` core parameterized by a `sourceRemap` function (identity for same-Project, itemIDMap lookup for cross-Project). Single-Project (dest) write with a seven-way snapshot + warnings-as-failure rollback covering both phases.

**Tech Stack:** Go, `internal/rifx` chunk tree, `internal/aep` parser. No new dependencies.

**Spec:** `flightdeck/specs/2026-05-29-v3-phase5c1-cross-project-insertlayer-design.md` (read §0, §3, §6 first).

---

## File Structure

- **Create** `internal/aep/import_closure.go` — `insertLayerCrossProject` + closure helpers (`isFileBacked`, `destFootageByPath`, `locateItemBlockByID`, `remapClonedCompLayerLayrs`, `importFootageBlock`).
- **Create** `internal/aep/import_closure_test.go` — helper unit tests + cross-Project refuse/happy-path/round-trip tests.
- **Modify** `internal/aep/insert_layer.go` — lift R7 to a branch; extract the clone+splice+reparse body into `spliceLayerClone(c, src, atIdx, sourceRemap)`; same-Project calls it with identity remap.
- **Do NOT modify** `internal/aep/duplicate_composition.go` (5D Stable — its layer-ID remap pattern is re-implemented locally, not shared).
- **Create** `tmp_debug/ge_cross_project_insert/main.go` — ge emitter for ship-gate.
- **Create** `test_data/re_cross_project_insert.jsx` — builds the AE-native src+dest fixtures.
- **Create** `test_data/verify_ge_cross_project_insert.jsx` — AE acceptance assertions.
- **Modify** `flightdeck/flight-plans/coverage.md`, godoc on `InsertLayer`, `flightdeck/cockpit.md` (post-gate).

**Key existing symbols reused** (all intra-package, already defined): `deepCloneChunk`, `indexOfChunk`, `isItemList`, `idtaItemID` (= comp/footage idta item-ID offset @0x10), `parseFootage(item, id, fallbackName)`, `parseComposition(item, id, name, &warnings)`, `findLayrIndexInItemList`, `insertLayrPosition`, `findAlternateSourceBlsi(layr)`, `chunkIDString`, `newParseCtxFPS`, `parseLayer`, `assignTransformDefaults`, `allocItemID`, `AVItemByID`, rifx IDs (`IDItem`/`IDLayr`/`IDLdta`/`IDIdta`/`IDEwst`/`IDUtf8`).

**ldta byte offsets** (from `parse_layer.go` / 5C design): `@0x00` layer ID, `@0x28` SourceID, `@0x6B` TrackMatte mode, `@0x84` ParentID, `@0xA0` explicit TrackMatteLayerID (AE 23+, guard `len>=0xA4`). blsi chunk `@0x00` = AlternateSourceID.

---

## Task 1: Extract `spliceLayerClone` core from `InsertLayer` (refactor, behavior-preserving)

**Files:**
- Modify: `internal/aep/insert_layer.go`
- Test (regression only): `internal/aep/insert_layer_test.go` (existing — do not edit)

- [ ] **Step 1: Run existing InsertLayer tests to capture the green baseline**

Run: `go test ./internal/aep/ -run TestInsertLayer -count=1 -v`
Expected: all `TestInsertLayer_*` PASS (or SKIP if fixtures absent). Note which PASS — they must stay PASS after refactor.

- [ ] **Step 2: Extract the clone+splice+reparse body into `spliceLayerClone`**

In `insert_layer.go`, replace everything in `InsertLayer` **after the R1–R11 validation block** (from the `// === Adaptive block end ===` comment through the final `return cloneLayer, nil`) with a call, and add the extracted function. The extraction adds ONE new behavior knob — a `sourceRemap func(uint32) uint32` applied to the cloned `SourceID @0x28` and `AlternateSourceID` (blsi). For the existing path the caller passes identity, so emitted bytes are unchanged.

Replace the tail of `InsertLayer` (after R11 corruption checks) with:

```go
	return spliceLayerClone(c, src, atIdx, srcLayrIdx, srcChildren, func(id uint32) uint32 { return id })
}

// spliceLayerClone deep-clones the source Layr block at srcLayrIdx (within
// srcChildren — src.comp's itemList) into c at atIdx, applying the standard
// cross-comp ldta mutations (new layer ID, ParentID/matte reset) plus
// sourceRemap to SourceID @0x28 and AlternateSourceID (blsi). sourceRemap is
// identity for same-Project inserts (bytes unchanged) and an itemIDMap lookup
// for cross-Project inserts. Atomic over c.itemList / c.Layers / proj.nextItemID
// / proj.Warnings.
func spliceLayerClone(c *Composition, src *Layer, atIdx, srcLayrIdx int, srcChildren []*rifx.Chunk, sourceRemap func(uint32) uint32) (*Layer, error) {
	// === Adaptive block end — scan leaf followers until next LIST/EOF ===
	endIdx := srcLayrIdx + 2
	for endIdx < len(srcChildren) && !srcChildren[endIdx].IsList() {
		endIdx++
	}

	// === Snapshot for rollback ===
	oldDestChildren := append([]*rifx.Chunk(nil), c.back.itemList.Children...)
	oldDestLayers := append([]*Layer(nil), c.Layers...)
	oldNextItemID := c.proj.nextItemID
	oldWarningsLen := len(c.proj.Warnings)

	// === Deep-clone source block (fresh Data slices) ===
	cloneBlock := make([]*rifx.Chunk, endIdx-srcLayrIdx)
	for k := srcLayrIdx; k < endIdx; k++ {
		cloneBlock[k-srcLayrIdx] = deepCloneChunk(srcChildren[k])
	}

	// === Allocate new ID + per-byte ldta mutations ===
	newID := c.proj.allocItemID()
	clonedLayr := cloneBlock[0]
	clonedLdta := clonedLayr.FindFirst(rifx.IDLdta)
	if clonedLdta == nil {
		c.proj.nextItemID = oldNextItemID
		return nil, fmt.Errorf("InsertLayer: cloned Layr has no ldta chunk")
	}
	if len(clonedLdta.Data) < 0x88 {
		c.proj.nextItemID = oldNextItemID
		return nil, fmt.Errorf("InsertLayer: cloned Layr ldta too short for ParentID write (got %d bytes, need >=0x88)", len(clonedLdta.Data))
	}
	binary.BigEndian.PutUint32(clonedLdta.Data[0x00:0x04], newID)
	clonedLdta.Data[0x6B] = byte(TrackMatteNone)
	binary.BigEndian.PutUint32(clonedLdta.Data[0x84:0x88], 0)
	if len(clonedLdta.Data) >= 0xA4 {
		binary.BigEndian.PutUint32(clonedLdta.Data[0xA0:0xA4], 0)
	}
	// SourceID remap (identity for same-Project — byte-preserving).
	srcSourceID := binary.BigEndian.Uint32(clonedLdta.Data[0x28:0x2C])
	binary.BigEndian.PutUint32(clonedLdta.Data[0x28:0x2C], sourceRemap(srcSourceID))
	// AlternateSourceID remap (Media Replacement override; blsi @0x00).
	if blsi := findAlternateSourceBlsi(clonedLayr); blsi != nil && len(blsi.Data) >= 4 {
		altID := binary.BigEndian.Uint32(blsi.Data[0:4])
		binary.BigEndian.PutUint32(blsi.Data[0:4], sourceRemap(altID))
	}

	// === Compute dest splice index ===
	destChildren := c.back.itemList.Children
	var insertChunkIdx int
	switch {
	case len(c.Layers) == 0:
		insertChunkIdx = insertLayrPosition(destChildren)
	case atIdx < len(c.Layers):
		target := c.Layers[atIdx]
		if target.back == nil || target.back.layrList == nil {
			c.proj.nextItemID = oldNextItemID
			return nil, fmt.Errorf("InsertLayer: dest Layers[%d] %q has no Layr backref", atIdx, target.Name)
		}
		insertChunkIdx = indexOfChunk(destChildren, target.back.layrList)
		if insertChunkIdx < 0 {
			c.proj.nextItemID = oldNextItemID
			return nil, fmt.Errorf("InsertLayer: dest Layers[%d] %q Layr chunk not found in dest itemList", atIdx, target.Name)
		}
	default:
		last := c.Layers[len(c.Layers)-1]
		if last.back == nil || last.back.layrList == nil {
			c.proj.nextItemID = oldNextItemID
			return nil, fmt.Errorf("InsertLayer: last dest layer %q has no Layr backref", last.Name)
		}
		lastLayrIdx := indexOfChunk(destChildren, last.back.layrList)
		if lastLayrIdx < 0 {
			c.proj.nextItemID = oldNextItemID
			return nil, fmt.Errorf("InsertLayer: last dest layer %q Layr chunk not found", last.Name)
		}
		insertChunkIdx = lastLayrIdx + 2
		for insertChunkIdx < len(destChildren) && !destChildren[insertChunkIdx].IsList() {
			insertChunkIdx++
		}
	}

	// === Splice cloneBlock into dest itemList ===
	newDestChildren := make([]*rifx.Chunk, 0, len(destChildren)+len(cloneBlock))
	newDestChildren = append(newDestChildren, destChildren[:insertChunkIdx]...)
	newDestChildren = append(newDestChildren, cloneBlock...)
	newDestChildren = append(newDestChildren, destChildren[insertChunkIdx:]...)
	c.back.itemList.Children = newDestChildren

	// === Re-parse cloned Layr → fresh *Layer ===
	var localWarnings []string
	ctx := newParseCtxFPS(c.TickRate, c.FrameRate, c.Name, &localWarnings)
	cloneLayer, parseErr := parseLayer(clonedLayr, atIdx, ctx)
	if parseErr != nil {
		c.back.itemList.Children = oldDestChildren
		c.proj.nextItemID = oldNextItemID
		return nil, fmt.Errorf("InsertLayer: re-parse cloned layer: %w", parseErr)
	}
	cloneLayer.comp = c
	assignTransformDefaults(cloneLayer.Properties, c, cloneLayer.Type)

	// === Insert cloneLayer into c.Layers ===
	newLayers := make([]*Layer, 0, len(c.Layers)+1)
	newLayers = append(newLayers, c.Layers[:atIdx]...)
	newLayers = append(newLayers, cloneLayer)
	newLayers = append(newLayers, c.Layers[atIdx:]...)
	c.Layers = newLayers

	// === Warnings-as-failure rollback ===
	if len(localWarnings) > 0 {
		c.proj.Warnings = append(c.proj.Warnings, localWarnings...)
	}
	if len(c.proj.Warnings) > oldWarningsLen {
		c.back.itemList.Children = oldDestChildren
		c.Layers = oldDestLayers
		c.proj.nextItemID = oldNextItemID
		newWarnings := append([]string(nil), c.proj.Warnings[oldWarningsLen:]...)
		c.proj.Warnings = c.proj.Warnings[:oldWarningsLen]
		return nil, fmt.Errorf("InsertLayer: produced %d parser warning(s), rolled back: %v", len(newWarnings), newWarnings)
	}

	return cloneLayer, nil
}
```

Note: `InsertLayer`'s existing R10/R11 block already computes `srcChildren` and `srcLayrIdx` (and validates the Layr/Ewst form types) — keep that block in `InsertLayer` and pass both into `spliceLayerClone`. Remove the now-duplicated adaptive-scan + body from `InsertLayer`.

- [ ] **Step 3: Run the regression tests — bytes must be unchanged**

Run: `go test ./internal/aep/ -run TestInsertLayer -count=1 -v`
Expected: every test that PASSed in Step 1 still PASSes. `TestInsertLayer_RoundTrip` and `TestInsertLayer_FreshDataSlices` are the byte-stability guards.

- [ ] **Step 4: Vet + full package build**

Run: `go vet ./internal/aep/ && go build ./...`
Expected: no output (clean).

- [ ] **Step 5: Commit**

```bash
git add internal/aep/insert_layer.go
git commit -m "refactor(aep): extract spliceLayerClone core from InsertLayer (sourceRemap hook, identity = byte-identical)"
```

---

## Task 2: Closure helpers + unit tests

**Files:**
- Create: `internal/aep/import_closure.go`
- Create: `internal/aep/import_closure_test.go`

- [ ] **Step 1: Write failing tests for the pure helpers**

Create `internal/aep/import_closure_test.go`:

```go
package aep_test

import (
	"os"
	"path/filepath"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// openXProjPair opens the re_duplicate_item baseline TWICE, yielding two
// INDEPENDENT *Project instances (distinct proj pointers) so InsertLayer sees
// a genuine cross-Project call. Returns (srcProj, srcCompAMain, destProj,
// destComp). destComp is a comp in destProj distinct from compA_main. Skips if
// the fixture is missing.
func openXProjPair(t *testing.T) (*aep.Project, *aep.Composition, *aep.Project, *aep.Composition) {
	t.Helper()
	path := filepath.Join("../../test_data", "re_duplicate_item_before.aep")
	if _, err := os.Stat(path); err != nil {
		path = filepath.Join("../../test_data", "re_duplicate_item_after.aep")
	}
	if _, err := os.Stat(path); err != nil {
		t.Skipf("fixture missing: re_duplicate_item_{before,after}.aep (run test_data/re_duplicate_item.jsx)")
		return nil, nil, nil, nil
	}
	open := func() (*aep.Project, *aep.Composition) {
		p, err := aep.Open(path)
		if err != nil {
			t.Fatalf("aep.Open(%s): %v", path, err)
		}
		var main *aep.Composition
		for _, c := range p.Compositions {
			if c.Name == "compA_main" {
				main = c
			}
		}
		if main == nil {
			t.Fatalf("compA_main not found in %s", path)
		}
		return p, main
	}
	srcProj, srcMain := open()
	destProj, destMain := open()
	// dest comp = the precomp compC (distinct from compA_main) so we insert into
	// a comp that is NOT the closure target.
	var destComp *aep.Composition
	for _, c := range destProj.Compositions {
		if c.Name != "compA_main" {
			destComp = c
			break
		}
	}
	if destComp == nil {
		destComp = destMain // fallback: any comp
	}
	return srcProj, srcMain, destProj, destComp
}

func TestImportHelpers_DestFootageByPath(t *testing.T) {
	_, _, destProj, _ := openXProjPair(t)
	if destProj == nil {
		return
	}
	var withPath *aep.Footage
	for _, f := range destProj.Footage {
		if f.Path != "" && !f.IsSolid && !f.IsPlaceholder {
			withPath = f
			break
		}
	}
	if withPath == nil {
		t.Skip("fixture has no file-backed footage to match")
	}
	got := aep.DestFootageByPathForTest(destProj, withPath.Path)
	if got == nil || got.ID != withPath.ID {
		t.Fatalf("DestFootageByPath(%q) = %v, want footage id %d", withPath.Path, got, withPath.ID)
	}
	if aep.DestFootageByPathForTest(destProj, `\\no\such\path.xyz`) != nil {
		t.Errorf("DestFootageByPath of unknown path should be nil")
	}
}

func TestImportHelpers_LocateItemBlockByID(t *testing.T) {
	srcProj, srcMain, _, _ := openXProjPair(t)
	if srcProj == nil {
		return
	}
	start, end := aep.LocateItemBlockByIDForTest(srcProj, srcMain.ID)
	if start < 0 || end <= start {
		t.Fatalf("LocateItemBlockByID(compA_main id=%d) = (%d,%d), want a valid range", srcMain.ID, start, end)
	}
	if s2, _ := aep.LocateItemBlockByIDForTest(srcProj, 0xFFFFFF); s2 != -1 {
		t.Errorf("LocateItemBlockByID(unknown) start = %d, want -1", s2)
	}
}
```

- [ ] **Step 2: Add the test-only accessors**

Append to `internal/aep/testutil_insert_test.go`:

```go
// DestFootageByPathForTest exposes destFootageByPath for cross-Project tests.
func DestFootageByPathForTest(p *Project, path string) *Footage { return destFootageByPath(p, path) }

// LocateItemBlockByIDForTest exposes locateItemBlockByID over a Project's root Fold.
func LocateItemBlockByIDForTest(p *Project, id uint32) (int, int) {
	if p.back == nil || p.back.rootFold == nil {
		return -1, -1
	}
	return locateItemBlockByID(p.back.rootFold, id)
}
```

- [ ] **Step 3: Run tests to verify they fail (undefined helpers)**

Run: `go test ./internal/aep/ -run 'TestImportHelpers' -count=1`
Expected: COMPILE FAIL — `undefined: destFootageByPath`, `undefined: locateItemBlockByID`.

- [ ] **Step 4: Implement the helpers**

Create `internal/aep/import_closure.go`:

```go
package aep

import (
	"encoding/binary"
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// isFileBacked reports whether f is file footage eligible for path-based dedup
// (not solid/placeholder, with a non-empty Path).
func isFileBacked(f *Footage) bool {
	return f != nil && !f.IsSolid && !f.IsPlaceholder && f.Path != ""
}

// destFootageByPath returns the first file-backed footage in p whose Path
// equals path, or nil. Used for cross-Project footage dedup (5C.1).
func destFootageByPath(p *Project, path string) *Footage {
	if path == "" {
		return nil
	}
	for _, f := range p.Footage {
		if isFileBacked(f) && f.Path == path {
			return f
		}
	}
	return nil
}

// locateItemBlockByID finds the Item LIST in rootFold.Children whose idta
// item-ID (@idtaItemID) equals id, returning [start, end) covering the Item
// LIST plus its trailing non-Item sibling run. Returns (-1,-1) if not found.
func locateItemBlockByID(rootFold *rifx.Chunk, id uint32) (int, int) {
	children := rootFold.Children
	for i, ch := range children {
		if !isItemList(ch) {
			continue
		}
		idta := ch.FindFirst(rifx.IDIdta)
		if idta == nil || len(idta.Data) < idtaItemID+4 {
			continue
		}
		if binary.BigEndian.Uint32(idta.Data[idtaItemID:idtaItemID+4]) != id {
			continue
		}
		end := i + 1
		for end < len(children) && !isItemList(children[end]) {
			end++
		}
		return i, end
	}
	return -1, -1
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/aep/ -run 'TestImportHelpers' -count=1 -v`
Expected: PASS (or SKIP if fixture missing).

- [ ] **Step 6: Commit**

```bash
git add internal/aep/import_closure.go internal/aep/import_closure_test.go internal/aep/testutil_insert_test.go
git commit -m "feat(aep): cross-Project closure helpers (footage dedup, item-block locate)"
```

---

## Task 3: Comp-clone layer-ID remap helper + footage importer

**Files:**
- Modify: `internal/aep/import_closure.go`
- Modify: `internal/aep/import_closure_test.go`

- [ ] **Step 1: Write the failing test for the footage importer**

Append to `internal/aep/import_closure_test.go`:

```go
func TestImportFootageBlock(t *testing.T) {
	srcProj, _, destProj, _ := openXProjPair(t)
	if srcProj == nil {
		return
	}
	// pick a file-backed footage in srcProj that is NOT already in destProj by path
	var srcF *aep.Footage
	for _, f := range srcProj.Footage {
		if f.Path != "" && !f.IsSolid && !f.IsPlaceholder && aep.DestFootageByPathForTest(destProj, f.Path) == nil {
			srcF = f
			break
		}
	}
	if srcF == nil {
		t.Skip("no importable (non-dup) file footage in fixture")
	}
	preCount := len(destProj.Footage)
	preNext := destProj.NextItemIDForTest()
	destID, err := aep.ImportFootageBlockForTest(destProj, srcProj, srcF.ID, srcF.Name)
	if err != nil {
		t.Fatalf("importFootageBlock: %v", err)
	}
	if destID != preNext {
		t.Errorf("imported footage destID = %d, want %d (head counter)", destID, preNext)
	}
	if len(destProj.Footage) != preCount+1 {
		t.Errorf("destProj.Footage count = %d, want %d", len(destProj.Footage), preCount+1)
	}
	imported := destProj.CompositionByID(destID) // should be nil (it's footage)
	if imported != nil {
		t.Errorf("destID resolved as a comp, want footage")
	}
	var ok bool
	for _, f := range destProj.Footage {
		if f.ID == destID && f.Path == srcF.Path {
			ok = true
		}
	}
	if !ok {
		t.Errorf("imported footage not found with id=%d path=%q", destID, srcF.Path)
	}
}
```

- [ ] **Step 2: Add the test accessor**

Append to `internal/aep/testutil_insert_test.go`:

```go
// ImportFootageBlockForTest exposes importFootageBlock for cross-Project tests.
func ImportFootageBlockForTest(dest, src *Project, srcID uint32, name string) (uint32, error) {
	return importFootageBlock(dest, src, srcID, name)
}
```

- [ ] **Step 3: Run to verify failure**

Run: `go test ./internal/aep/ -run 'TestImportFootageBlock' -count=1`
Expected: COMPILE FAIL — `undefined: importFootageBlock` / `remapClonedCompLayerLayrs`.

- [ ] **Step 4: Implement both helpers**

Append to `internal/aep/import_closure.go`:

```go
// importFootageBlock deep-clones srcID's footage Item block from src's root
// Fold into dest's root Fold with a fresh dest item ID (idta @idtaItemID),
// parses it into dest.Footage, and returns the new dest item ID. On error the
// CALLER (insertLayerCrossProject) restores dest via its outer snapshot — this
// helper does not self-rollback.
func importFootageBlock(dest, src *Project, srcID uint32, name string) (uint32, error) {
	srcRoot := src.back.rootFold
	start, end := locateItemBlockByID(srcRoot, srcID)
	if start < 0 {
		return 0, fmt.Errorf("footage Item block id=%d not found in src root Fold", srcID)
	}
	dup := deepCloneChunk(srcRoot.Children[start])
	destID := dest.allocItemID()
	idta := dup.FindFirst(rifx.IDIdta)
	if idta == nil || len(idta.Data) < idtaItemID+4 {
		return 0, fmt.Errorf("cloned footage id=%d idta missing/short", srcID)
	}
	binary.BigEndian.PutUint32(idta.Data[idtaItemID:idtaItemID+4], destID)

	destRoot := dest.back.rootFold
	destRoot.Children = append(destRoot.Children, dup)
	for k := start + 1; k < end; k++ {
		destRoot.Children = append(destRoot.Children, deepCloneChunk(srcRoot.Children[k]))
	}

	f, err := parseFootage(dup, destID, name)
	if err != nil {
		return 0, fmt.Errorf("re-parse cloned footage id=%d: %w", srcID, err)
	}
	dest.Footage = append(dest.Footage, f)
	return destID, nil
}

// remapClonedCompLayerLayrs walks dupItemList's Layr LIST children, allocates a
// fresh dest layer ID per layer (rewriting ldta @0x00 and remapping intra-comp
// ParentID @0x84 / explicit matte @0xA0 through the local srcLayerID→destLayerID
// map), and returns the Layr LIST chunks (for the later cross-comp source-ref
// remap pass). Mirrors DuplicateComposition's two-pass pattern locally (5D file
// untouched per spec trade-off #1).
func remapClonedCompLayerLayrs(p *Project, dupItemList *rifx.Chunk) ([]*rifx.Chunk, error) {
	idMap := make(map[uint32]uint32)
	var layrs []*rifx.Chunk
	for _, ch := range dupItemList.Children {
		if !ch.IsList() || ch.FormType != rifx.IDLayr {
			continue
		}
		ldta := ch.FindFirst(rifx.IDLdta)
		if ldta == nil {
			continue
		}
		if len(ldta.Data) < 0x88 {
			return nil, fmt.Errorf("cloned Layr ldta too short for ParentID write (got %d bytes, need >=0x88)", len(ldta.Data))
		}
		oldID := binary.BigEndian.Uint32(ldta.Data[0x00:0x04])
		newID := p.allocItemID()
		idMap[oldID] = newID
		binary.BigEndian.PutUint32(ldta.Data[0x00:0x04], newID)
		layrs = append(layrs, ch)
	}
	for _, ch := range layrs {
		ldta := ch.FindFirst(rifx.IDLdta)
		if parent := binary.BigEndian.Uint32(ldta.Data[0x84:0x88]); parent != 0 {
			if mapped, ok := idMap[parent]; ok {
				binary.BigEndian.PutUint32(ldta.Data[0x84:0x88], mapped)
			}
		}
		if len(ldta.Data) >= 0xA4 {
			if matte := binary.BigEndian.Uint32(ldta.Data[0xA0:0xA4]); matte != 0 {
				if mapped, ok := idMap[matte]; ok {
					binary.BigEndian.PutUint32(ldta.Data[0xA0:0xA4], mapped)
				}
			}
		}
	}
	return layrs, nil
}
```

- [ ] **Step 5: Run to verify pass**

Run: `go test ./internal/aep/ -run 'TestImportFootageBlock' -count=1 -v`
Expected: PASS (or SKIP).

- [ ] **Step 6: Commit**

```bash
git add internal/aep/import_closure.go internal/aep/import_closure_test.go internal/aep/testutil_insert_test.go
git commit -m "feat(aep): cross-Project footage importer + comp layer-ID remap helper"
```

---

## Task 4: `insertLayerCrossProject` orchestration + lift R7 branch in `InsertLayer`

**Files:**
- Modify: `internal/aep/import_closure.go`
- Modify: `internal/aep/insert_layer.go`

- [ ] **Step 1: Lift R7 to a branch and gate R9 on same-Project in `InsertLayer`**

In `insert_layer.go`, replace the R7 refuse block:

```go
	if src.comp.proj != c.proj {
		return nil, fmt.Errorf("InsertLayer: src and dest in different Projects — cross-Project insert deferred to Phase 5C.1")
	}
```

with a `crossProject` flag computed right after the R6 same-comp check, and gate R9 on it:

```go
	crossProject := src.comp.proj != c.proj
```

Then change the R9 block to:

```go
	if !crossProject && src.SourceID != 0 && src.SourceID == c.ID {
		return nil, fmt.Errorf("InsertLayer: refuse direct pre-comp loop (src.SourceID=%d == dest.ID=%d)", src.SourceID, c.ID)
	}
```

And, immediately after the R10/R11 block computes `srcChildren` and `srcLayrIdx` (and validates Layr/Ewst form types), branch before splicing:

```go
	if crossProject {
		return insertLayerCrossProject(c, src, atIdx, srcLayrIdx, srcChildren)
	}
	return spliceLayerClone(c, src, atIdx, srcLayrIdx, srcChildren, func(id uint32) uint32 { return id })
```

- [ ] **Step 2: Implement `insertLayerCrossProject`**

Append to `internal/aep/import_closure.go`:

```go
// insertLayerCrossProject handles InsertLayer when src lives in a different
// Project than c. It imports src's reachable item closure (footage + precomp,
// transitively) into c's Project at root level with fresh item IDs, dedup'ing
// file-backed footage by Path, then splices the layer via spliceLayerClone with
// the SourceID/AlternateSourceID remapped to the imported dest items. Atomic:
// a seven-way dest snapshot + warnings-as-failure rollback covers both phases.
// srcChildren/srcLayrIdx are the located src Layr position from InsertLayer.
func insertLayerCrossProject(c *Composition, src *Layer, atIdx, srcLayrIdx int, srcChildren []*rifx.Chunk) (*Layer, error) {
	destProj := c.proj
	srcProj := src.comp.proj

	// X1 / X2
	if destProj.back == nil || destProj.back.rootFold == nil {
		return nil, fmt.Errorf("InsertLayer: dest Project has no root Fold back-ref (built outside parser?)")
	}
	if srcProj == nil {
		return nil, fmt.Errorf("InsertLayer: src layer's Project is unknown (src.comp.proj == nil)")
	}
	if srcProj.back == nil || srcProj.back.rootFold == nil {
		return nil, fmt.Errorf("InsertLayer: src Project has no root Fold back-ref")
	}
	rootFold := destProj.back.rootFold

	// === Outer snapshot (covers closure import + the layer splice) ===
	oldRootChildren := append([]*rifx.Chunk(nil), rootFold.Children...)
	oldComps := append([]*Composition(nil), destProj.Compositions...)
	oldFootage := append([]*Footage(nil), destProj.Footage...)
	oldDestItemList := append([]*rifx.Chunk(nil), c.back.itemList.Children...)
	oldDestLayers := append([]*Layer(nil), c.Layers...)
	oldNextItemID := destProj.nextItemID
	oldWarningsLen := len(destProj.Warnings)
	rollback := func() {
		rootFold.Children = oldRootChildren
		destProj.Compositions = oldComps
		destProj.Footage = oldFootage
		c.back.itemList.Children = oldDestItemList
		c.Layers = oldDestLayers
		destProj.nextItemID = oldNextItemID
		if len(destProj.Warnings) > oldWarningsLen {
			destProj.Warnings = destProj.Warnings[:oldWarningsLen]
		}
	}

	// === PHASE 1: import the source item closure (BFS) ===
	itemIDMap := make(map[uint32]uint32)
	type pendingComp struct {
		dup   *rifx.Chunk
		id    uint32
		name  string
		layrs []*rifx.Chunk
	}
	var pending []pendingComp

	worklist := make([]uint32, 0, 2)
	if src.SourceID != 0 {
		worklist = append(worklist, src.SourceID)
	}
	if src.AlternateSourceID != 0 {
		worklist = append(worklist, src.AlternateSourceID)
	}

	for len(worklist) > 0 {
		srcID := worklist[0]
		worklist = worklist[1:]
		if srcID == 0 {
			continue
		}
		if _, done := itemIDMap[srcID]; done {
			continue
		}
		item := srcProj.AVItemByID(srcID)
		if item == nil {
			rollback()
			return nil, fmt.Errorf("InsertLayer: cross-Project source item id=%d not found in src Project (dangling)", srcID)
		}
		switch it := item.(type) {
		case *Footage:
			if isFileBacked(it) {
				if existing := destFootageByPath(destProj, it.Path); existing != nil {
					itemIDMap[srcID] = existing.ID // dedup hit — reuse, no clone
					continue
				}
			}
			destID, err := importFootageBlock(destProj, srcProj, srcID, it.Name)
			if err != nil {
				rollback()
				return nil, err
			}
			itemIDMap[srcID] = destID
		case *Composition:
			container, start, end := locateItemBlockByID(srcProj.back.rootFold, srcID)
			if container == nil {
				rollback()
				return nil, fmt.Errorf("InsertLayer: cross-Project comp id=%d Item block not found in src Project", srcID)
			}
			dup := deepCloneChunk(container.Children[start])
			destID := destProj.allocItemID()
			idta := dup.FindFirst(rifx.IDIdta)
			if idta == nil || len(idta.Data) < idtaItemID+4 {
				rollback()
				return nil, fmt.Errorf("InsertLayer: cloned comp id=%d idta missing/short", srcID)
			}
			binary.BigEndian.PutUint32(idta.Data[idtaItemID:idtaItemID+4], destID)
			layrs, err := remapClonedCompLayerLayrs(destProj, dup)
			if err != nil {
				rollback()
				return nil, fmt.Errorf("InsertLayer: comp id=%d: %w", srcID, err)
			}
			// splice into dest root Fold (append after existing items — flatten
			// nested-folder source comps to the dest root per spec §0)
			rootFold.Children = append(rootFold.Children, dup)
			for k := start + 1; k < end; k++ {
				rootFold.Children = append(rootFold.Children, deepCloneChunk(container.Children[k]))
			}
			itemIDMap[srcID] = destID
			// collect this comp's layers' source refs (still SRC-project IDs)
			for _, layr := range layrs {
				ldta := layr.FindFirst(rifx.IDLdta)
				if sid := binary.BigEndian.Uint32(ldta.Data[0x28:0x2C]); sid != 0 {
					worklist = append(worklist, sid)
				}
				if blsi := findAlternateSourceBlsi(layr); blsi != nil && len(blsi.Data) >= 4 {
					if aid := binary.BigEndian.Uint32(blsi.Data[0:4]); aid != 0 {
						worklist = append(worklist, aid)
					}
				}
			}
			pending = append(pending, pendingComp{dup: dup, id: destID, name: it.Name, layrs: layrs})
		}
	}

	// === PHASE 1 Pass 2: remap imported comps' layer source refs ===
	remap := func(id uint32) uint32 {
		if id == 0 {
			return 0
		}
		if mapped, ok := itemIDMap[id]; ok {
			return mapped
		}
		return id // src's own direct refs are always mapped; leave unknowns verbatim
	}
	for _, pc := range pending {
		for _, layr := range pc.layrs {
			ldta := layr.FindFirst(rifx.IDLdta)
			if sid := binary.BigEndian.Uint32(ldta.Data[0x28:0x2C]); sid != 0 {
				binary.BigEndian.PutUint32(ldta.Data[0x28:0x2C], remap(sid))
			}
			if blsi := findAlternateSourceBlsi(layr); blsi != nil && len(blsi.Data) >= 4 {
				if aid := binary.BigEndian.Uint32(blsi.Data[0:4]); aid != 0 {
					binary.BigEndian.PutUint32(blsi.Data[0:4], remap(aid))
				}
			}
		}
	}

	// === Reparse imported comps (after source remap so refs resolve) ===
	for _, pc := range pending {
		dupComp, err := parseComposition(pc.dup, pc.id, pc.name, &destProj.Warnings)
		if err != nil {
			rollback()
			return nil, fmt.Errorf("InsertLayer: re-parse imported comp %q: %w", pc.name, err)
		}
		dupComp.proj = destProj
		destProj.Compositions = append(destProj.Compositions, dupComp)
	}

	// === PHASE 2: splice the layer with SourceID/AltSourceID remapped ===
	clone, err := spliceLayerClone(c, src, atIdx, srcLayrIdx, srcChildren, remap)
	if err != nil {
		rollback()
		return nil, err
	}

	// === Warnings-as-failure (covers import + splice) ===
	if len(destProj.Warnings) > oldWarningsLen {
		newWarnings := append([]string(nil), destProj.Warnings[oldWarningsLen:]...)
		rollback()
		return nil, fmt.Errorf("InsertLayer: cross-Project produced %d parser warning(s), rolled back: %v", len(newWarnings), newWarnings)
	}

	return clone, nil
}
```

- [ ] **Step 3: Vet + build**

Run: `go vet ./internal/aep/ && go build ./...`
Expected: clean.

- [ ] **Step 4: Run the full InsertLayer suite (same-Project regression must hold)**

Run: `go test ./internal/aep/ -run 'TestInsertLayer|TestImport' -count=1 -v`
Expected: existing same-Project tests still PASS; new helper tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/aep/import_closure.go internal/aep/insert_layer.go
git commit -m "feat(aep): cross-Project InsertLayer — closure import + SourceID remap (Alpha)"
```

---

## Task 5: Cross-Project behavior tests (refuse + happy-path + round-trip)

**Files:**
- Modify: `internal/aep/import_closure_test.go`

- [ ] **Step 1: Write the refuse tests (X2, X3)**

Append to `internal/aep/import_closure_test.go`:

```go
import "bytes" // add to existing import block

// X3: dangling source — src layer references an item id absent from srcProj.
func TestInsertLayerXProj_RefuseDanglingSource(t *testing.T) {
	srcProj, srcMain, destProj, destComp := openXProjPair(t)
	if srcProj == nil {
		return
	}
	_ = destProj
	// find an AV layer with a real source, then point it at a bogus id
	var src *aep.Layer
	for _, l := range srcMain.Layers {
		if l.Type == aep.LayerTypeAV && l.SourceID != 0 {
			src = l
			break
		}
	}
	if src == nil {
		t.Skip("no AV layer with a source in fixture")
	}
	orig := src.SourceID
	src.SourceID = 0x00FFFFFF
	defer func() { src.SourceID = orig }()
	_, err := destComp.InsertLayer(src, 0)
	if err == nil || !strings.Contains(err.Error(), "dangling") {
		t.Fatalf("want 'dangling' refuse, got %v", err)
	}
}

// X1: dest Project has no root Fold (built outside parser).
func TestInsertLayerXProj_RefuseDestNoRootFold(t *testing.T) {
	srcProj, srcMain, _, _ := openXProjPair(t)
	if srcProj == nil {
		return
	}
	var src *aep.Layer
	for _, l := range srcMain.Layers {
		if l.Type == aep.LayerTypeAV {
			src = l
			break
		}
	}
	if src == nil {
		t.Skip("no AV layer in fixture")
	}
	// Build a dest comp whose proj is a bare Project (back == nil) but which
	// still has a real itemList (so R2 passes and we reach the X1 check).
	_, _, realDestProj, realDest := openXProjPair(t)
	bare := &aep.Project{}
	aep.SetCompProjForTest(realDest, bare)
	_ = realDestProj
	_, err := realDest.InsertLayer(src, 0)
	if err == nil || !strings.Contains(err.Error(), "root Fold") {
		t.Fatalf("want 'root Fold' refuse, got %v", err)
	}
}
```

(`strings` is already imported by other test files in the package; add it to this file's import block.)

- [ ] **Step 2: Write the happy-path tests (footage import, precomp closure, dedup, round-trip)**

Append:

```go
func TestInsertLayerXProj_HappyPath_Closure(t *testing.T) {
	srcProj, srcMain, destProj, destComp := openXProjPair(t)
	if srcProj == nil {
		return
	}
	// choose an AV layer whose source resolves in srcProj
	var src *aep.Layer
	for _, l := range srcMain.Layers {
		if l.Type == aep.LayerTypeAV && srcProj.AVItemByID(l.SourceID) != nil {
			src = l
			break
		}
	}
	if src == nil {
		t.Skip("no AV layer with a resolvable source in fixture")
	}
	preDestItems := len(destProj.Compositions) + len(destProj.Footage)
	preLayers := len(destComp.Layers)

	clone, err := destComp.InsertLayer(src, 0)
	if err != nil {
		t.Fatalf("cross-Project InsertLayer: %v", err)
	}
	if clone == nil || len(destComp.Layers) != preLayers+1 || destComp.Layers[0] != clone {
		t.Fatalf("clone not inserted at slot 0 (layers=%d)", len(destComp.Layers))
	}
	// closure imported: dest item count grew by >=1
	postDestItems := len(destProj.Compositions) + len(destProj.Footage)
	if postDestItems <= preDestItems {
		t.Errorf("dest item count did not grow: pre=%d post=%d", preDestItems, postDestItems)
	}
	// SourceID remapped: clone.SourceID resolves to a DEST item (not the src id)
	if clone.SourceID == 0 {
		t.Fatalf("clone.SourceID is 0")
	}
	if destProj.AVItemByID(clone.SourceID) == nil {
		t.Errorf("clone.SourceID=%d does not resolve in destProj (remap failed)", clone.SourceID)
	}
	// parent/matte reset (cross-comp orphan)
	if clone.ParentID != 0 || clone.TrackMatteLayerID != 0 {
		t.Errorf("clone refs not reset: ParentID=%d matte=%d", clone.ParentID, clone.TrackMatteLayerID)
	}
}

func TestInsertLayerXProj_FootageDedup(t *testing.T) {
	srcProj, srcMain, destProj, destComp := openXProjPair(t)
	if srcProj == nil {
		return
	}
	// pick a file-backed footage layer whose Path ALSO exists in destProj
	var src *aep.Layer
	for _, l := range srcMain.Layers {
		f := srcProj.AVItemByID(l.SourceID)
		ff, ok := f.(*aep.Footage)
		if ok && ff.Path != "" && aep.DestFootageByPathForTest(destProj, ff.Path) != nil {
			src = l
			break
		}
	}
	if src == nil {
		t.Skip("no file-footage layer whose Path is shared with dest")
	}
	preFootage := len(destProj.Footage)
	clone, err := destComp.InsertLayer(src, 0)
	if err != nil {
		t.Fatalf("InsertLayer (dedup): %v", err)
	}
	// dedup: footage count must NOT grow
	if len(destProj.Footage) != preFootage {
		t.Errorf("footage duplicated despite path match: pre=%d post=%d", preFootage, len(destProj.Footage))
	}
	// SourceID points at the pre-existing dest footage
	if destProj.AVItemByID(clone.SourceID) == nil {
		t.Errorf("dedup clone.SourceID=%d unresolved in dest", clone.SourceID)
	}
}

func TestInsertLayerXProj_RoundTrip(t *testing.T) {
	srcProj, srcMain, destProj, destComp := openXProjPair(t)
	if srcProj == nil {
		return
	}
	var src *aep.Layer
	for _, l := range srcMain.Layers {
		if l.Type == aep.LayerTypeAV && srcProj.AVItemByID(l.SourceID) != nil {
			src = l
			break
		}
	}
	if src == nil {
		t.Skip("no resolvable AV source layer")
	}
	clone, err := destComp.InsertLayer(src, 0)
	if err != nil {
		t.Fatalf("InsertLayer: %v", err)
	}
	wantID, wantSource := clone.ID, clone.SourceID

	var buf bytes.Buffer
	if err := destProj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	// find the comp with our clone and assert the source resolves in proj2
	found := false
	for _, c := range proj2.Compositions {
		for _, l := range c.Layers {
			if l.ID == wantID {
				found = true
				if l.SourceID != wantSource {
					t.Errorf("re-opened clone SourceID=%d, want %d", l.SourceID, wantSource)
				}
				if proj2.AVItemByID(l.SourceID) == nil {
					t.Errorf("re-opened clone source %d unresolved", l.SourceID)
				}
			}
		}
	}
	if !found {
		t.Errorf("clone layer id=%d not found after round-trip", wantID)
	}
}
```

- [ ] **Step 3: Run the cross-Project tests**

Run: `go test ./internal/aep/ -run 'TestInsertLayerXProj' -count=1 -v`
Expected: PASS (or SKIP if fixture missing). If a closure assertion fails, bisect per spec §6 bisection candidates.

- [ ] **Step 4: Full package test + vet**

Run: `go vet ./... && go test ./internal/aep/ -count=1`
Expected: `ok`.

- [ ] **Step 5: Commit**

```bash
git add internal/aep/import_closure_test.go
git commit -m "test(aep): cross-Project InsertLayer refuse + closure/dedup/round-trip"
```

---

## Task 6: Ship-gate artifacts (ge emitter + JSX fixtures)

**Files:**
- Create: `tmp_debug/ge_cross_project_insert/main.go`
- Create: `test_data/re_cross_project_insert.jsx`
- Create: `test_data/verify_ge_cross_project_insert.jsx`

Read `flightdeck/checklists/re-fixture.md` (§ GDI 自动化, per-version `.done` tag) and `flightdeck/incident-reports/ae25-acceptance-gate.md` before this task. The assert-based gate (spec §6) replaces byte-diff — there is no AE "after" baseline.

- [ ] **Step 1: Author `re_cross_project_insert.jsx` — builds src + dest projects per mode**

The JSX (driven by an env/arg `RE_XPROJ_MODE` like the other RE scripts) saves two AE-native files per mode:
- `re_xproj_src_<mode>.aep`: a comp `src_comp` with the layer(s) to copy + its footage/precomp closure.
- `re_xproj_dest_<mode>.aep`: a comp `dest_comp` to receive the clone (+ for `dedup`, a footage already importing the same file as the src's footage).

Modes: `footage` (layer → 1 file footage), `precomp` (layer → precomp → nested footage), `dedup` (layer → footage whose file is also imported in dest). Mirror the structure of `test_data/re_duplicate_item.jsx` for comp/footage/precomp creation. Save via `project.save(File(...))`.

- [ ] **Step 2: Author `tmp_debug/ge_cross_project_insert/main.go` — Go-side emitter**

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	aep "github.com/example/aep-parser/internal/aep"
)

// ge_cross_project_insert opens re_xproj_src_<mode> + re_xproj_dest_<mode>,
// inserts the src comp's first AV layer into the dest comp (cross-Project),
// writes ge_cross_project_insert_<mode>.aep. Go-side assertions guard before
// write; AE acceptance is verified by verify_ge_cross_project_insert.jsx.
func main() {
	dir := `e:\projects\tools\aep-parser\test_data`
	for _, mode := range []string{"footage", "precomp", "dedup"} {
		srcProj, err := aep.Open(filepath.Join(dir, "re_xproj_src_"+mode+".aep"))
		must(err, mode, "open src")
		destProj, err := aep.Open(filepath.Join(dir, "re_xproj_dest_"+mode+".aep"))
		must(err, mode, "open dest")

		srcComp := srcProj.CompositionByName("src_comp")
		destComp := destProj.CompositionByName("dest_comp")
		if srcComp == nil || destComp == nil {
			fail(mode, "src_comp/dest_comp not found")
		}
		var src *aep.Layer
		for _, l := range srcComp.Layers {
			if l.Type == aep.LayerTypeAV {
				src = l
				break
			}
		}
		if src == nil {
			fail(mode, "no AV layer in src_comp")
		}
		preFootage := len(destProj.Footage)
		clone, err := destComp.InsertLayer(src, 0)
		must(err, mode, "InsertLayer")
		if destProj.AVItemByID(clone.SourceID) == nil {
			fail(mode, fmt.Sprintf("clone.SourceID=%d unresolved in dest", clone.SourceID))
		}
		if mode == "dedup" && len(destProj.Footage) != preFootage {
			fail(mode, "footage duplicated in dedup mode")
		}
		out := filepath.Join(dir, "ge_cross_project_insert_"+mode+".aep")
		f, err := os.Create(out)
		must(err, mode, "create out")
		must(destProj.WriteAEP(f), mode, "WriteAEP")
		f.Close()
		fmt.Printf("[%s] OK → %s (clone id=%d src=%d)\n", mode, out, clone.ID, clone.SourceID)
	}
}

func must(err error, mode, what string) {
	if err != nil {
		fail(mode, what+": "+err.Error())
	}
}
func fail(mode, msg string) {
	fmt.Fprintf(os.Stderr, "[%s] FAIL: %s\n", mode, msg)
	os.Exit(1)
}
```

- [ ] **Step 3: Author `verify_ge_cross_project_insert.jsx` — AE acceptance assertions**

For each mode, AE opens `ge_cross_project_insert_<mode>.aep` and asserts (writing a per-version `.done` tag with last line `PASS`/`FAIL` per re-fixture.md):
- `dest_comp` exists and `dest_comp.layer(1)` is the inserted clone;
- `dest_comp.layer(1).source` is non-null and `.source.name` matches the expected imported item;
- the source has no missing-footage flag (`source.footageMissing` false for file footage);
- `app.project.numItems` grew by the expected closure size;
- `dedup` mode: footage item count grew by closure-minus-1 (shared footage reused).

- [ ] **Step 4: Generate fixtures (self-serve AE run)**

Run (per `feedback_ae_ship_gate_self_serve` — agent runs unattended; AE 2020 cold-start may need a warm retry):
```
pwsh scripts/ae_run.ps1 -Jsx test_data/re_cross_project_insert.jsx -Mode footage
pwsh scripts/ae_run.ps1 -Jsx test_data/re_cross_project_insert.jsx -Mode precomp
pwsh scripts/ae_run.ps1 -Jsx test_data/re_cross_project_insert.jsx -Mode dedup
```
(Adjust to the actual `ae_run.ps1` flag names — confirm against `flightdeck/checklists/re-fixture.md`.)
Expected: six `re_xproj_{src,dest}_*.aep` files in `test_data/` (gitignored).

- [ ] **Step 5: Build ge files**

Run: `go run ./tmp_debug/ge_cross_project_insert`
Expected: `[footage] OK`, `[precomp] OK`, `[dedup] OK` and three `ge_cross_project_insert_*.aep`.

- [ ] **Step 6: Commit (scripts + emitter only — fixtures are gitignored)**

```bash
git add tmp_debug/ge_cross_project_insert/main.go test_data/re_cross_project_insert.jsx test_data/verify_ge_cross_project_insert.jsx
git commit -m "test(aep): cross-Project InsertLayer ship-gate emitter + JSX fixtures"
```

---

## Task 7: Ship-gate run + Stable promotion + doc sync

**Files:**
- Modify: `internal/aep/insert_layer.go` (godoc), `internal/aep/import_closure.go` (godoc)
- Modify: `flightdeck/flight-plans/coverage.md`
- Modify: `flightdeck/cockpit.md`

- [ ] **Step 1: Run the cross-Project acceptance gate (3 modes × AE 2020 + AE 2025)**

Run:
```
pwsh scripts/ae_run.ps1 -Jsx test_data/verify_ge_cross_project_insert.jsx -AE 2020
pwsh scripts/ae_run.ps1 -Jsx test_data/verify_ge_cross_project_insert.jsx -AE 2025
```
Expected: per-version `.done` tags, last line `PASS`, for all 3 modes × 2 versions = 6 PASS.
On `FAIL` / silent-drop / exit 2 (cold-start flake) → warm retry once; if still failing, bisect per spec §6.

- [ ] **Step 2: Run the same-Project InsertLayer regression gate (shared core touched in Task 1)**

Run:
```
go run ./tmp_debug/ge_insert_layer
pwsh scripts/ae_run.ps1 -Jsx test_data/verify_ge_insert_layer.jsx -AE 2020
pwsh scripts/ae_run.ps1 -Jsx test_data/verify_ge_insert_layer.jsx -AE 2025
```
Expected: the existing `verify_ge_insert_layer_ae20{20,25}_{basic,footage,precomp}.done` all stay `PASS` (6/6) — confirms the `spliceLayerClone` extraction did not regress same-Project bytes.

- [ ] **Step 3: Promote godoc Alpha → Stable**

In `insert_layer.go`, update the `InsertLayer` godoc to document the cross-Project semantics and, once Step 1–2 are green, mark the cross-Project path Stable (cite the gate result + date). Remove any Alpha caveat.

- [ ] **Step 4: Sync coverage + cockpit (per `commits.md` § 命令一致性)**

- `flightdeck/flight-plans/coverage.md`: extend the InsertLayer row to note cross-Project closure import; mark Stable with the 6/6 PASS.
- `flightdeck/cockpit.md`: update Active focus + add the ship-gate PASS line to 自验留痕; move this plan + the 5C.1 design to `landed/` (per the archive convention used this session).

- [ ] **Step 5: Final full verification**

Run: `go vet ./... && go test ./internal/aep/ -count=1`
Expected: `ok`.

- [ ] **Step 6: Commit**

```bash
git add internal/aep/insert_layer.go internal/aep/import_closure.go flightdeck/flight-plans/coverage.md flightdeck/cockpit.md
git commit -m "feat(aep): cross-Project InsertLayer Alpha→Stable — 6/6 AE 2020+2025 ship-gate PASS"
```

---

## Notes for the implementer

- **Atomicity is the trap.** The cross-Project path mutates dest `rootFold`, `Compositions`, `Footage`, `nextItemID`, the dest comp's `itemList`, and `c.Layers`. The single `rollback()` closure in `insertLayerCrossProject` must restore ALL of them on ANY failure (dangling source, parse error, warnings). `spliceLayerClone`'s own inner rollback is a subset and harmless to layer on top.
- **Two-pass ordering matters.** Imported comps' layer source refs hold SRC-project item IDs until Pass 2 remaps them. parse comps AFTER Pass 2 so `SourceComposition()`/`SourceFootage()` resolve against final dest IDs.
- **Don't touch `duplicate_composition.go`.** `remapClonedCompLayerLayrs` is a deliberate local re-implementation of its two-pass pattern (spec trade-off #1) — keep them independent.
- **Fixtures are gitignored.** All `re_*` / `ge_*` `.aep` files regenerate via the JSX + ge emitter. Only source code + `.jsx` + `tmp_debug` are committed.
- **Cold-start flake.** AE 2020 cold-start may hit the splash/About screen → exit 2 false negative. Warm retry clears it (per `feedback_ae_ship_gate_self_serve` + cockpit known-flake note).
```
