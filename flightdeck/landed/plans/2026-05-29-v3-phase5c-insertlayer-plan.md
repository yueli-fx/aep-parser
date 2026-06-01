# V3 Phase 5C InsertLayer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship `Composition.InsertLayer(src *Layer, atIdx int) (*Layer, error)` — same-Project sibling-comp deep-clone — Alpha-tagged in godoc, gated on AE 2020 + AE 2025 ship-gate (3 modes × 2 versions = 6 PASS) before Stable promotion.

**Architecture:** Spec at `flightdeck/specs/2026-05-29-v3-phase5c-insertlayer-design.md` (committed `d5ba860`). Builds on DuplicateLayer (same-comp clone, Stable post-5A/5B) by adding 3 ldta byte resets (ParentID @0x84=0, TrackMatte mode @0x6B=None, explicit matte ID @0xA0=0 when ldta ≥ 0xA4) and a different splice-index strategy (atIdx into dest comp vs srcIdx for self). Source comp never mutated (one-sided write). Reuses existing helpers: `deepCloneChunk`, `findLayrIndexInItemList`, `indexOfChunk`, `insertLayrPosition`, `parseLayer`, `newParseCtxFPS`, `assignTransformDefaults`, `chunkIDString`, `Project.allocItemID`.

**Tech Stack:** Go 1.21+, `internal/aep/` single package, `internal/rifx/` framing, ExtendScript (JSX) for RE fixtures, AfterFX.exe (2020/2025) for ship-gate (`scripts/ae_run.ps1`).

**Spec deltas applied during planning:**
- Spec §7 says fixtures live in `test_data/项目/` — **actual repo convention is `test_data/` (flat)**. Plan uses `test_data/`.
- Spec §3 step 7 invents `findFirstLayerSlot` — **codebase already has `insertLayrPosition(children []*rifx.Chunk) int`** (`internal/aep/new_layer.go:88`). Plan reuses it.

**Pre-flight check (before any task):**

```bash
go vet ./...
go test -count=1 ./internal/aep/... -run 'TestDuplicateLayer|TestDeleteLayer|TestMoveLayer' -v
git status --short
```

Expected: `go vet` clean, all DuplicateLayer/DeleteLayer/MoveLayer tests PASS, working tree clean.

---

## File map

**Create:**
- `internal/aep/insert_layer.go` — `func (c *Composition) InsertLayer(src *Layer, atIdx int) (*Layer, error)` + no new exported helpers; uses existing intra-package helpers.
- `internal/aep/insert_layer_test.go` — refuse-case matrix (R1–R11) + happy-path × 3 splice positions + identity-field assertions + concurrent-mutate safety + round-trip + structural-equivalence (skipped without fixtures).
- `test_data/re_insert_layer.jsx` — produces 3 mode pairs (`basic` / `footage` / `precomp`) as before/after `.aep` baselines.

**Modify:**
- `flightdeck/flight-plans/coverage.md` — add `InsertLayer` row under structural mutation API (Alpha).
- `flightdeck/cockpit.md` — bump Active focus to "Phase 5C InsertLayer Alpha shipped, awaiting AE ship-gate".

**Out of plan scope (deferred per spec §8):**
- Cross-Project (`src.comp.proj != c.proj`) — Phase 5C.1.
- Same-comp redirect — caller uses `DuplicateLayer`.
- Non-AV archetypes (Shape/Text/Camera/Light) — refused via R8.
- `(l *Layer) CopyTo(dest, atIdx)` convenience wrapper — post-Stable ergonomics decision.
- AE ship-gate run itself — user invokes `scripts/ae_run.ps1`; plan terminates at fixture authoring + Alpha doc tag.

---

# Phase A — Walking skeleton (refuse-case matrix)

Goal: land an `InsertLayer` that returns errors for every refuse-case in spec §2, with full TDD coverage. No happy-path yet. Establishes the file + test scaffold that Phase B fills in.

### Task A.1: Test scaffold + first failing test (R1 nil src)

**Files:**
- Create: `internal/aep/insert_layer_test.go`

- [ ] **Step 1: Write the failing test**

```go
package aep_test

import (
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestInsertLayer_RefuseNilSrc(t *testing.T) {
	c := &aep.Composition{Layers: []*aep.Layer{}}
	_, err := c.InsertLayer(nil, 0)
	if err == nil {
		t.Fatal("expected refuse on nil src, got nil error")
	}
	if !strings.Contains(err.Error(), "src cannot be nil") {
		t.Errorf("error should mention 'src cannot be nil', got %q", err.Error())
	}
}
```

- [ ] **Step 2: Run test — must fail (no InsertLayer method yet)**

Run: `go test ./internal/aep/ -run TestInsertLayer_RefuseNilSrc -v`
Expected: build error `c.InsertLayer undefined` OR test FAIL.

- [ ] **Step 3: Create insert_layer.go with the minimum to compile + pass R1**

```go
package aep

import (
	"fmt"
)

// InsertLayer deep-clones src (from a sibling comp in the SAME Project)
// into c.Layers at atIdx. ALPHA — pending AE 2020+2025 ship-gate
// (3 modes × 2 versions = 6 PASS) before Stable promotion. See
// flightdeck/specs/2026-05-29-v3-phase5c-insertlayer-design.md.
func (c *Composition) InsertLayer(src *Layer, atIdx int) (*Layer, error) {
	if src == nil {
		return nil, fmt.Errorf("InsertLayer: src cannot be nil")
	}
	return nil, fmt.Errorf("InsertLayer: not yet implemented")
}
```

- [ ] **Step 4: Run test — must pass**

Run: `go test ./internal/aep/ -run TestInsertLayer_RefuseNilSrc -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/aep/insert_layer.go internal/aep/insert_layer_test.go
git commit -m "feat(aep): V3 Phase 5C InsertLayer scaffold + R1 (nil src)"
```

### Task A.2: Refuse-cases R2-R7 (missing backrefs, range, cross-comp validity)

**Files:**
- Modify: `internal/aep/insert_layer_test.go` (append tests)
- Modify: `internal/aep/insert_layer.go` (extend refuse block)

- [ ] **Step 1: Write failing tests for R2-R7**

Append to `internal/aep/insert_layer_test.go`:

```go
// R2: dest comp has no itemList back-ref
func TestInsertLayer_RefuseDestMissingItemList(t *testing.T) {
	dest := &aep.Composition{Layers: []*aep.Layer{}}
	srcComp := &aep.Composition{}
	src := &aep.Layer{ID: 10, Type: aep.LayerTypeAV}
	aep.SetLayerCompForTest(src, srcComp)
	_, err := dest.InsertLayer(src, 0)
	if err == nil || !strings.Contains(err.Error(), "itemList back-ref") {
		t.Fatalf("want 'itemList back-ref' error, got %v", err)
	}
}

// R3: dest comp has no Project back-ref (can't allocItemID)
// Reuses a real opened baseline then nils proj for the test — direct synth
// would have nil itemList first (R2 fires earlier).
func TestInsertLayer_RefuseDestMissingProject(t *testing.T) {
	dest, src := openInsertPair(t)
	if dest == nil {
		return
	}
	aep.SetCompProjForTest(dest, nil)
	_, err := dest.InsertLayer(src, 0)
	if err == nil || !strings.Contains(err.Error(), "project back-ref") {
		t.Fatalf("want 'project back-ref' error, got %v", err)
	}
}

// R4: atIdx out of range (< 0 or > len). atIdx == len is "append" (allowed).
func TestInsertLayer_RefuseAtIdxOutOfRange(t *testing.T) {
	dest, src := openInsertPair(t)
	if dest == nil {
		return
	}
	preLen := len(dest.Layers)
	for _, idx := range []int{-1, preLen + 1, preLen + 50} {
		_, err := dest.InsertLayer(src, idx)
		if err == nil || !strings.Contains(err.Error(), "out of range") {
			t.Errorf("atIdx=%d: want 'out of range' error, got %v", idx, err)
		}
	}
	if len(dest.Layers) != preLen {
		t.Errorf("dest.Layers count changed after refused inserts: got %d, want %d", len(dest.Layers), preLen)
	}
}

// R5: src.comp == nil
func TestInsertLayer_RefuseSrcDetached(t *testing.T) {
	dest, _ := openInsertPair(t)
	if dest == nil {
		return
	}
	orphan := &aep.Layer{ID: 99, Type: aep.LayerTypeAV}
	// orphan.comp stays nil.
	_, err := dest.InsertLayer(orphan, 0)
	if err == nil || !strings.Contains(err.Error(), "src.comp") {
		t.Fatalf("want 'src.comp' error, got %v", err)
	}
}

// R6: same-comp — caller should use DuplicateLayer
func TestInsertLayer_RefuseSameComp(t *testing.T) {
	dest, _ := openInsertPair(t)
	if dest == nil {
		return
	}
	if len(dest.Layers) == 0 {
		t.Fatalf("fixture precondition: dest.Layers empty")
	}
	_, err := dest.InsertLayer(dest.Layers[0], 0)
	if err == nil || !strings.Contains(err.Error(), "DuplicateLayer") {
		t.Fatalf("want 'DuplicateLayer' redirect error, got %v", err)
	}
}

// R7: cross-Project (src and dest in different Projects)
func TestInsertLayer_RefuseCrossProject(t *testing.T) {
	dest, _ := openInsertPair(t)
	if dest == nil {
		return
	}
	otherDest, otherSrc := openInsertPair(t)
	if otherDest == nil {
		return
	}
	_ = otherDest
	_, err := dest.InsertLayer(otherSrc, 0)
	if err == nil || !strings.Contains(err.Error(), "cross-Project") {
		t.Fatalf("want 'cross-Project' error, got %v", err)
	}
}
```

- [ ] **Step 2: Add fixture helper + test-only setters at the top of insert_layer_test.go**

Insert above `TestInsertLayer_RefuseNilSrc`:

```go
import (
	"os"
	"path/filepath"
)

const insertLayerFixtureDir = "../../test_data"

// openInsertPair opens the baseline that contains 2 sibling comps in
// one Project (compA = src holder, compB = dest), returns (dest=compB, src=compA.Layers[0]).
// Skips if fixture missing.
func openInsertPair(t *testing.T) (*aep.Composition, *aep.Layer) {
	t.Helper()
	path := filepath.Join(insertLayerFixtureDir, "re_insert_layer_basic_before.aep")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("fixture missing: %s (run test_data/re_insert_layer.jsx with RE_INSERT_MODE=basic_before)", path)
		return nil, nil
	}
	proj, err := aep.Open(path)
	if err != nil {
		t.Fatalf("aep.Open(%s): %v", path, err)
	}
	if len(proj.Compositions) < 2 {
		t.Fatalf("fixture %s: want >=2 comps, got %d", path, len(proj.Compositions))
	}
	compA := proj.Compositions[0]
	compB := proj.Compositions[1]
	if len(compA.Layers) == 0 {
		t.Fatalf("fixture %s: compA has no layers", path)
	}
	return compB, compA.Layers[0]
}
```

Add a new test helper file for the cross-package setters required by R2 + R3 (we can't reach `Layer.comp` or `Composition.proj` from `_test` package directly). Create `internal/aep/testutil_insert_test.go`:

```go
package aep

// SetLayerCompForTest assigns the back-ref used by InsertLayer's R5 check.
// Test-only; production code never calls this.
func SetLayerCompForTest(l *Layer, c *Composition) { l.comp = c }

// SetCompProjForTest assigns the project back-ref used by InsertLayer's R3.
// Test-only.
func SetCompProjForTest(c *Composition, p *Project) { c.proj = p }
```

- [ ] **Step 3: Run tests — all R2-R7 must fail**

Run: `go test ./internal/aep/ -run 'TestInsertLayer_Refuse' -v`
Expected: tests for R2-R7 FAIL with "not yet implemented".

- [ ] **Step 4: Extend InsertLayer refuse block for R2-R7**

Replace the placeholder `return nil, fmt.Errorf("InsertLayer: not yet implemented")` in `internal/aep/insert_layer.go` with:

```go
	if c.back == nil || c.back.itemList == nil {
		return nil, fmt.Errorf("InsertLayer: dest comp %q has no itemList back-ref (built outside parser?)", c.Name)
	}
	if c.proj == nil {
		return nil, fmt.Errorf("InsertLayer: dest comp %q has no project back-ref", c.Name)
	}
	if atIdx < 0 || atIdx > len(c.Layers) {
		return nil, fmt.Errorf("InsertLayer: atIdx %d out of range (have %d layers; %d is append)", atIdx, len(c.Layers), len(c.Layers))
	}
	if src.comp == nil {
		return nil, fmt.Errorf("InsertLayer: src.comp is nil (layer detached from any comp)")
	}
	if src.comp == c {
		return nil, fmt.Errorf("InsertLayer: src and dest are the same comp %q — use DuplicateLayer instead", c.Name)
	}
	if src.comp.proj != c.proj {
		return nil, fmt.Errorf("InsertLayer: src and dest in different Projects — cross-Project insert deferred to Phase 5C.1")
	}
	return nil, fmt.Errorf("InsertLayer: not yet implemented")
```

- [ ] **Step 5: Run R1-R7 tests — all PASS**

Run: `go test ./internal/aep/ -run 'TestInsertLayer_Refuse(NilSrc|DestMissing|AtIdx|SrcDetached|SameComp|CrossProject)' -v`
Expected: 6 PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/aep/insert_layer.go internal/aep/insert_layer_test.go internal/aep/testutil_insert_test.go
git commit -m "feat(aep): V3 Phase 5C InsertLayer R2-R7 refuse-cases + test helpers"
```

### Task A.3: Refuse-cases R8-R11 (semantic + structural defense)

**Files:**
- Modify: `internal/aep/insert_layer_test.go`
- Modify: `internal/aep/insert_layer.go`

- [ ] **Step 1: Write failing tests for R8-R11**

Append to `internal/aep/insert_layer_test.go`:

```go
// R8: src not AV (camera/light/text/shape refused in Phase 5C)
func TestInsertLayer_RefuseNonAV(t *testing.T) {
	dest, src := openInsertPair(t)
	if dest == nil {
		return
	}
	origType := src.Type
	src.Type = aep.LayerTypeCamera
	defer func() { src.Type = origType }()
	_, err := dest.InsertLayer(src, 0)
	if err == nil || !strings.Contains(err.Error(), "non-AV") {
		t.Fatalf("want 'non-AV' error, got %v", err)
	}
}

// R9: direct pre-comp loop — src.SourceID points at dest comp itself
func TestInsertLayer_RefuseDirectPrecompLoop(t *testing.T) {
	dest, src := openInsertPair(t)
	if dest == nil {
		return
	}
	origSourceID := src.SourceID
	src.SourceID = dest.ID
	defer func() { src.SourceID = origSourceID }()
	if dest.ID == 0 {
		t.Skip("fixture dest comp has ID 0; can't trigger R9 (SourceID==0 means 'no source')")
	}
	_, err := dest.InsertLayer(src, 0)
	if err == nil || !strings.Contains(err.Error(), "pre-comp loop") {
		t.Fatalf("want 'pre-comp loop' error, got %v", err)
	}
}

// R10: src has no layrList backref (built outside parser)
func TestInsertLayer_RefuseSrcMissingLayrList(t *testing.T) {
	dest, src := openInsertPair(t)
	if dest == nil {
		return
	}
	aep.ClearLayerLayrListForTest(src)
	_, err := dest.InsertLayer(src, 0)
	if err == nil || !strings.Contains(err.Error(), "Layr chunk back-ref") {
		t.Fatalf("want 'Layr chunk back-ref' error, got %v", err)
	}
}

// R11: structural corruption (FormType mismatch) — exercised via test hook.
func TestInsertLayer_RefuseStructuralCorruption(t *testing.T) {
	dest, src := openInsertPair(t)
	if dest == nil {
		return
	}
	aep.CorruptSrcLayrFormTypeForTest(src)
	_, err := dest.InsertLayer(src, 0)
	if err == nil || !strings.Contains(err.Error(), "non-Layr") {
		t.Fatalf("want 'non-Layr' corruption error, got %v", err)
	}
}
```

- [ ] **Step 2: Extend testutil_insert_test.go with the new helpers**

Append to `internal/aep/testutil_insert_test.go`:

```go
import "github.com/example/aep-parser/internal/rifx"

func ClearLayerLayrListForTest(l *Layer) {
	if l.back != nil {
		l.back.layrList = nil
	}
}

func CorruptSrcLayrFormTypeForTest(l *Layer) {
	if l.back != nil && l.back.layrList != nil {
		l.back.layrList.FormType = rifx.ChunkID{'X', 'X', 'X', 'X'}
	}
}
```

(If the import is unused in the first file revision, fold both helpers into the same file; verify package compiles after step 4.)

- [ ] **Step 3: Run new tests — must FAIL ("not yet implemented")**

Run: `go test ./internal/aep/ -run 'TestInsertLayer_Refuse(NonAV|DirectPrecompLoop|SrcMissingLayrList|StructuralCorruption)' -v`
Expected: 4 FAIL.

- [ ] **Step 4: Extend InsertLayer refuse block for R8-R11**

In `internal/aep/insert_layer.go`, replace the trailing `return nil, fmt.Errorf("InsertLayer: not yet implemented")` with:

```go
	if src.Type != LayerTypeAV {
		return nil, fmt.Errorf("InsertLayer: refuse non-AV src (Type=%s); only AV layers supported in Phase 5C", src.Type)
	}
	if src.SourceID != 0 && src.SourceID == c.ID {
		return nil, fmt.Errorf("InsertLayer: refuse direct pre-comp loop (src.SourceID=%d == dest.ID=%d)", src.SourceID, c.ID)
	}
	if src.back == nil || src.back.layrList == nil {
		return nil, fmt.Errorf("InsertLayer: src layer %q has no Layr chunk back-ref", src.Name)
	}
	srcChildren := src.comp.back.itemList.Children
	srcLayrIdx := findLayrIndexInItemList(src.comp.back.itemList, src.back.layrList)
	if srcLayrIdx < 0 {
		return nil, fmt.Errorf("InsertLayer: src layer %q Layr chunk not found in its comp's itemList", src.Name)
	}
	if !srcChildren[srcLayrIdx].IsList() || srcChildren[srcLayrIdx].FormType != rifx.IDLayr {
		return nil, fmt.Errorf("InsertLayer: src layer %q backref points to non-Layr chunk (FormType=%s)", src.Name, chunkIDString(srcChildren[srcLayrIdx].FormType))
	}
	if srcLayrIdx+1 >= len(srcChildren) {
		return nil, fmt.Errorf("InsertLayer: src layer %q Layr at end of itemList (no Ewst sibling)", src.Name)
	}
	if !srcChildren[srcLayrIdx+1].IsList() || srcChildren[srcLayrIdx+1].FormType != rifx.IDEwst {
		return nil, fmt.Errorf("InsertLayer: src layer %q expected Ewst sibling after Layr, found %s", src.Name, chunkIDString(srcChildren[srcLayrIdx+1].FormType))
	}
	return nil, fmt.Errorf("InsertLayer: not yet implemented (happy path forthcoming in Phase B)")
```

Add `"github.com/example/aep-parser/internal/rifx"` to the imports of `insert_layer.go` (needed for `rifx.IDLayr` / `rifx.IDEwst`).

- [ ] **Step 5: Run all refuse tests — 11 PASS**

Run: `go test ./internal/aep/ -run 'TestInsertLayer_Refuse' -v`
Expected: 11 PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/aep/insert_layer.go internal/aep/insert_layer_test.go internal/aep/testutil_insert_test.go
git commit -m "feat(aep): V3 Phase 5C InsertLayer R8-R11 refuse-cases (semantic + structural)"
```

---

# Phase B — Happy path: clone block construction + splice

Goal: complete the InsertLayer implementation so a basic same-Project cross-comp clone produces the correct chunk shape + identity fields. Splice positions (empty / mid / append) all handled.

### Task B.1: First happy-path test (basic mode, atIdx=0)

**Files:**
- Modify: `internal/aep/insert_layer_test.go`

- [ ] **Step 1: Write failing happy-path test**

Append to `internal/aep/insert_layer_test.go`:

```go
func TestInsertLayer_HappyPath_Basic_AtIdxZero(t *testing.T) {
	dest, src := openInsertPair(t)
	if dest == nil {
		return
	}
	srcID := src.ID
	srcName := src.Name
	srcSourceID := src.SourceID
	preLen := len(dest.Layers)
	preChildCount := len(dest.ItemListForTest().Children)
	preNextItemID := dest.ProjForTest().NextItemIDForTest()

	clone, err := dest.InsertLayer(src, 0)
	if err != nil {
		t.Fatalf("InsertLayer(src, 0): %v", err)
	}
	if clone == nil {
		t.Fatal("InsertLayer returned nil clone with no error")
	}
	if len(dest.Layers) != preLen+1 {
		t.Fatalf("dest.Layers count: got %d, want %d", len(dest.Layers), preLen+1)
	}
	if dest.Layers[0] != clone {
		t.Errorf("dest.Layers[0] should be the clone")
	}
	if clone.ID == srcID {
		t.Errorf("clone.ID must differ from source; both = %d", clone.ID)
	}
	if clone.ID != preNextItemID {
		t.Errorf("clone.ID: got %d, want %d (preNextItemID head counter+1)", clone.ID, preNextItemID)
	}
	if clone.Name != srcName {
		t.Errorf("clone.Name: got %q, want %q (verbatim from src)", clone.Name, srcName)
	}
	if clone.SourceID != srcSourceID {
		t.Errorf("clone.SourceID: got %d, want %d (verbatim from src)", clone.SourceID, srcSourceID)
	}
	if clone.ParentID != 0 {
		t.Errorf("clone.ParentID: got %d, want 0 (reset for cross-comp)", clone.ParentID)
	}
	if clone.TrackMatteLayerID != 0 {
		t.Errorf("clone.TrackMatteLayerID: got %d, want 0 (reset for cross-comp)", clone.TrackMatteLayerID)
	}
	if clone.TrackMatte != aep.TrackMatteNone {
		t.Errorf("clone.TrackMatte: got %d, want TrackMatteNone (reset for cross-comp)", clone.TrackMatte)
	}
	postChildCount := len(dest.ItemListForTest().Children)
	if postChildCount <= preChildCount {
		t.Errorf("dest itemList children should grow; pre=%d post=%d", preChildCount, postChildCount)
	}
	postNextItemID := dest.ProjForTest().NextItemIDForTest()
	if postNextItemID != preNextItemID+1 {
		t.Errorf("proj.nextItemID: got %d, want %d (pre+1)", postNextItemID, preNextItemID+1)
	}
}
```

- [ ] **Step 2: Add the missing test accessor `ProjForTest` if not present**

Run: `grep -n 'func (c \*Composition) ProjForTest' internal/aep/*.go`

If absent, append to `internal/aep/testutil_insert_test.go`:

```go
func (c *Composition) ProjForTest() *Project { return c.proj }
```

- [ ] **Step 3: Run happy-path test — must FAIL ("not yet implemented")**

Run: `go test ./internal/aep/ -run TestInsertLayer_HappyPath_Basic_AtIdxZero -v`
Expected: FAIL with "not yet implemented" OR fixture-skip (if `re_insert_layer_basic_before.aep` absent — Phase C produces it; that's OK for now, the implementation still needs to be correct).

- [ ] **Step 4: Commit the test (impl in B.2)**

```bash
git add internal/aep/insert_layer_test.go internal/aep/testutil_insert_test.go
git commit -m "test(aep): InsertLayer basic happy-path expectations (atIdx=0)"
```

### Task B.2: Implement happy path (clone block + ldta byte deltas + splice + reparse)

**Files:**
- Modify: `internal/aep/insert_layer.go`

- [ ] **Step 1: Replace the trailing "not yet implemented" with full impl**

Final `internal/aep/insert_layer.go` (replace the file body wholesale — keeps refuse block intact, adds happy path):

```go
package aep

import (
	"encoding/binary"
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// InsertLayer deep-clones src (from a sibling comp in the SAME Project as c)
// into c.Layers at atIdx (0-based; atIdx == len(c.Layers) appends). Returns
// the inserted clone *Layer on success.
//
// Cross-comp clone semantics (Phase 5C; same-Project only — see
// flightdeck/specs/2026-05-29-v3-phase5c-insertlayer-design.md):
//
//   - new layer ID = c.proj.allocItemID() (head counter +1, monotonic)
//   - clone block = deep byte-clone of src's [Layr, Ewst, leaf-followers)
//     range, with per-byte ldta mutations:
//       @0x00..0x03 ← newID
//       @0x6B       ← TrackMatteNone (cross-comp matte source is invalid)
//       @0x84..0x87 ← 0 (ParentID; src's ParentID named a layer in src.comp)
//       @0xA0..0xA3 ← 0 (explicit matte ID, guarded by len(ldta) >= 0xA4)
//   - clone.SourceID = src.SourceID (verbatim — Footage/Comp item lives in
//     the shared Project; no item duplication).
//   - clone.Name = src.Name (verbatim — matches AE's layer.copyToComp).
//
// Refuse-cases (R1..R11; spec §2): nil src, dest backref missing, atIdx
// out of range, src detached, same-comp redirect, cross-Project,
// non-AV, direct pre-comp loop, src backref missing, structural
// corruption.
//
// Atomic mutation (Inv-10 / Inv-11): snapshot dest itemList.Children +
// c.Layers + c.proj.nextItemID + len(c.proj.Warnings); on any new
// parser warning during re-parse, roll all back including the
// nextItemID bump.
//
// ALPHA — pending AE 2020 + AE 2025 ship-gate (3 modes × 2 versions =
// 6 PASS). Will be promoted to Stable in godoc after ship-gate green.
func (c *Composition) InsertLayer(src *Layer, atIdx int) (*Layer, error) {
	// === Refuse-case matrix R1-R11 ===
	if src == nil {
		return nil, fmt.Errorf("InsertLayer: src cannot be nil")
	}
	if c.back == nil || c.back.itemList == nil {
		return nil, fmt.Errorf("InsertLayer: dest comp %q has no itemList back-ref (built outside parser?)", c.Name)
	}
	if c.proj == nil {
		return nil, fmt.Errorf("InsertLayer: dest comp %q has no project back-ref", c.Name)
	}
	if atIdx < 0 || atIdx > len(c.Layers) {
		return nil, fmt.Errorf("InsertLayer: atIdx %d out of range (have %d layers; %d is append)", atIdx, len(c.Layers), len(c.Layers))
	}
	if src.comp == nil {
		return nil, fmt.Errorf("InsertLayer: src.comp is nil (layer detached from any comp)")
	}
	if src.comp == c {
		return nil, fmt.Errorf("InsertLayer: src and dest are the same comp %q — use DuplicateLayer instead", c.Name)
	}
	if src.comp.proj != c.proj {
		return nil, fmt.Errorf("InsertLayer: src and dest in different Projects — cross-Project insert deferred to Phase 5C.1")
	}
	if src.Type != LayerTypeAV {
		return nil, fmt.Errorf("InsertLayer: refuse non-AV src (Type=%s); only AV layers supported in Phase 5C", src.Type)
	}
	if src.SourceID != 0 && src.SourceID == c.ID {
		return nil, fmt.Errorf("InsertLayer: refuse direct pre-comp loop (src.SourceID=%d == dest.ID=%d)", src.SourceID, c.ID)
	}
	if src.back == nil || src.back.layrList == nil {
		return nil, fmt.Errorf("InsertLayer: src layer %q has no Layr chunk back-ref", src.Name)
	}
	srcChildren := src.comp.back.itemList.Children
	srcLayrIdx := findLayrIndexInItemList(src.comp.back.itemList, src.back.layrList)
	if srcLayrIdx < 0 {
		return nil, fmt.Errorf("InsertLayer: src layer %q Layr chunk not found in its comp's itemList", src.Name)
	}
	if !srcChildren[srcLayrIdx].IsList() || srcChildren[srcLayrIdx].FormType != rifx.IDLayr {
		return nil, fmt.Errorf("InsertLayer: src layer %q backref points to non-Layr chunk (FormType=%s)", src.Name, chunkIDString(srcChildren[srcLayrIdx].FormType))
	}
	if srcLayrIdx+1 >= len(srcChildren) {
		return nil, fmt.Errorf("InsertLayer: src layer %q Layr at end of itemList (no Ewst sibling)", src.Name)
	}
	if !srcChildren[srcLayrIdx+1].IsList() || srcChildren[srcLayrIdx+1].FormType != rifx.IDEwst {
		return nil, fmt.Errorf("InsertLayer: src layer %q expected Ewst sibling after Layr, found %s", src.Name, chunkIDString(srcChildren[srcLayrIdx+1].FormType))
	}

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
	if clonedLdta == nil || len(clonedLdta.Data) < 0x88 {
		c.proj.nextItemID = oldNextItemID
		return nil, fmt.Errorf("InsertLayer: cloned Layr ldta too short for ParentID write (got %d bytes, need >=0x88)", len(clonedLdta.Data))
	}
	binary.BigEndian.PutUint32(clonedLdta.Data[0x00:0x04], newID)
	clonedLdta.Data[0x6B] = byte(TrackMatteNone)
	binary.BigEndian.PutUint32(clonedLdta.Data[0x84:0x88], 0)
	if len(clonedLdta.Data) >= 0xA4 {
		binary.BigEndian.PutUint32(clonedLdta.Data[0xA0:0xA4], 0)
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

	// === Re-parse cloned Layr → fresh *Layer with backrefs into clones ===
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

- [ ] **Step 2: `go vet` + compile**

Run: `go vet ./internal/aep/...`
Expected: clean.

- [ ] **Step 3: Run the basic happy-path test**

Run: `go test ./internal/aep/ -run TestInsertLayer_HappyPath_Basic_AtIdxZero -v`

Expected outcomes:
- **If `re_insert_layer_basic_before.aep` exists**: PASS.
- **If fixture missing**: SKIP (which is the current state until Phase C runs the JSX). That's expected; we'll re-run after Phase C.

- [ ] **Step 4: Run the full refuse suite to confirm no regression**

Run: `go test ./internal/aep/ -run 'TestInsertLayer_Refuse' -v`
Expected: 11 PASS.

- [ ] **Step 5: Confirm no regression elsewhere**

Run: `go test -count=1 ./internal/aep/...`
Expected: all existing tests still PASS (no fixture-availability regression for unrelated tests).

- [ ] **Step 6: Commit**

```bash
git add internal/aep/insert_layer.go
git commit -m "feat(aep): V3 Phase 5C InsertLayer happy-path impl (clone+splice+reparse)"
```

### Task B.3: Splice-position tests (mid, append, empty-dest)

**Files:**
- Modify: `internal/aep/insert_layer_test.go`

- [ ] **Step 1: Write failing/skipped tests for the 3 splice positions**

Append:

```go
func TestInsertLayer_HappyPath_Basic_MidAndAppend(t *testing.T) {
	for _, tc := range []struct {
		name     string
		atIdxFn  func(dest *aep.Composition) int
		wantSlot func(dest *aep.Composition, clone *aep.Layer) bool
	}{
		{
			name:     "middle",
			atIdxFn:  func(dest *aep.Composition) int { return len(dest.Layers) / 2 },
			wantSlot: func(dest *aep.Composition, clone *aep.Layer) bool { return dest.Layers[len(dest.Layers)/2-0] == clone },
		},
		{
			name:     "append",
			atIdxFn:  func(dest *aep.Composition) int { return len(dest.Layers) },
			wantSlot: func(dest *aep.Composition, clone *aep.Layer) bool { return dest.Layers[len(dest.Layers)-1] == clone },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dest, src := openInsertPair(t)
			if dest == nil {
				return
			}
			atIdx := tc.atIdxFn(dest)
			clone, err := dest.InsertLayer(src, atIdx)
			if err != nil {
				t.Fatalf("InsertLayer(src, %d): %v", atIdx, err)
			}
			if !tc.wantSlot(dest, clone) {
				t.Errorf("clone not at expected slot for %s; dest.Layers=%v", tc.name, layerIDs(dest.Layers))
			}
		})
	}
}

func TestInsertLayer_HappyPath_EmptyDest(t *testing.T) {
	dest, src := openInsertPair(t)
	if dest == nil {
		return
	}
	// Manually drain dest.Layers to simulate empty-dest path. We keep the
	// underlying itemList chunks intact (insertLayrPosition walks them);
	// only the Go-side []*Layer slice is empty. This exercises the
	// "case len(c.Layers) == 0" branch in InsertLayer.
	dest.Layers = nil
	clone, err := dest.InsertLayer(src, 0)
	if err != nil {
		t.Fatalf("InsertLayer into emptied dest: %v", err)
	}
	if len(dest.Layers) != 1 || dest.Layers[0] != clone {
		t.Errorf("empty-dest result: want single clone, got %v", layerIDs(dest.Layers))
	}
}

func layerIDs(ls []*aep.Layer) []uint32 {
	ids := make([]uint32, len(ls))
	for i, l := range ls {
		ids[i] = l.ID
	}
	return ids
}
```

- [ ] **Step 2: Run new splice tests**

Run: `go test ./internal/aep/ -run 'TestInsertLayer_HappyPath_(Basic_MidAndAppend|EmptyDest)' -v`
Expected: SKIP if fixture absent; PASS once fixture exists.

- [ ] **Step 3: Commit**

```bash
git add internal/aep/insert_layer_test.go
git commit -m "test(aep): InsertLayer splice-position coverage (mid/append/empty-dest)"
```

### Task B.4: Concurrent-mutate safety + round-trip

**Files:**
- Modify: `internal/aep/insert_layer_test.go`

- [ ] **Step 1: Write failing/skipped tests**

Append:

```go
import "bytes"

func TestInsertLayer_FreshDataSlices(t *testing.T) {
	dest, src := openInsertPair(t)
	if dest == nil {
		return
	}
	srcLdtaBefore := append([]byte(nil), src.LdtaForTest().Data...)
	clone, err := dest.InsertLayer(src, 0)
	if err != nil {
		t.Fatalf("InsertLayer: %v", err)
	}
	cloneLdta := clone.LdtaForTest()
	if cloneLdta == nil {
		t.Fatal("clone has no ldta backref")
	}
	if &cloneLdta.Data[0] == &src.LdtaForTest().Data[0] {
		t.Fatal("clone ldta shares Data slice header with src — must be fresh allocation")
	}
	for i := 4; i < len(cloneLdta.Data); i++ {
		cloneLdta.Data[i] ^= 0xFF
	}
	if !bytes.Equal(srcLdtaBefore, src.LdtaForTest().Data) {
		t.Fatal("src ldta bytes changed after mutating clone — Data slice sharing detected")
	}
}

func TestInsertLayer_RoundTrip(t *testing.T) {
	dest, src := openInsertPair(t)
	if dest == nil {
		return
	}
	proj := dest.ProjForTest()
	clone, err := dest.InsertLayer(src, 0)
	if err != nil {
		t.Fatalf("InsertLayer: %v", err)
	}
	wantCloneID := clone.ID
	wantCloneName := clone.Name

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader after insert+write: %v", err)
	}
	if len(proj2.Compositions) < 2 {
		t.Fatalf("re-opened comps=%d, want >=2", len(proj2.Compositions))
	}
	destReopen := proj2.Compositions[1]
	if destReopen.Layers[0].ID != wantCloneID {
		t.Errorf("re-opened dest.Layers[0].ID: got %d, want %d", destReopen.Layers[0].ID, wantCloneID)
	}
	if destReopen.Layers[0].Name != wantCloneName {
		t.Errorf("re-opened dest.Layers[0].Name: got %q, want %q", destReopen.Layers[0].Name, wantCloneName)
	}
	if destReopen.Layers[0].ParentID != 0 || destReopen.Layers[0].TrackMatteLayerID != 0 {
		t.Errorf("re-opened clone refs not reset: ParentID=%d TrackMatteLayerID=%d",
			destReopen.Layers[0].ParentID, destReopen.Layers[0].TrackMatteLayerID)
	}
}
```

- [ ] **Step 2: Run new tests**

Run: `go test ./internal/aep/ -run 'TestInsertLayer_(FreshDataSlices|RoundTrip)' -v`
Expected: SKIP if fixture absent; PASS once fixture exists.

- [ ] **Step 3: Commit**

```bash
git add internal/aep/insert_layer_test.go
git commit -m "test(aep): InsertLayer concurrent-mutate safety + round-trip"
```

---

# Phase C — RE fixture authoring

Goal: produce the 3 mode-paired `re_insert_layer_*.aep` baselines the Phase B tests need. **The JSX file authored here is committed; the `.aep` outputs are user-produced via AE.**

### Task C.1: Author re_insert_layer.jsx (3 modes)

**Files:**
- Create: `test_data/re_insert_layer.jsx`

- [ ] **Step 1: Write the JSX driver**

```jsx
// RE fixture for V3 Phase 5C — Composition.InsertLayer (cross-comp clone, same-Project).
//
// Produces 3 mode pairs (before/after) the Go ship-gate compares against:
//   basic    — compA: 1 solid "Src" no parent/no matte; compB: 1 unrelated solid.
//              After: compB has Src clone inserted at index 1.
//   footage  — compA: 1 AV layer over Footage F1; compB: 1 unrelated solid + F1
//              already imported. After: compB has Src clone refs same F1.
//   precomp  — compA: 1 AV layer whose source is precomp compC; compB: 1
//              unrelated solid + compC. After: compB has Src clone refs compC.
//
// Mode via $.getenv("RE_INSERT_MODE"); state ("before" | "after") via
// $.getenv("RE_INSERT_STATE").
//
// Outputs:
//   test_data/re_insert_layer_<mode>_<state>.aep
//   test_data/re_insert_layer_<mode>_<state>.done
//
// AE 2020 compatible (no AE 23+ APIs).
(function () {
    var mode  = $.getenv("RE_INSERT_MODE");
    var state = $.getenv("RE_INSERT_STATE");
    if (mode === null || mode === "")  mode  = "basic";
    if (state === null || state === "") state = "before";
    var validModes = "basic|footage|precomp";
    var validStates = "before|after";
    if (validModes.indexOf(mode) === -1 || validStates.indexOf(state) === -1) {
        var errFile = new File("e:/projects/tools/aep-parser/test_data/re_insert_layer_unknown.done");
        errFile.open("w");
        errFile.write("ERR unknown mode/state: " + mode + "/" + state);
        errFile.close();
        return;
    }

    var outFile  = new File("e:/projects/tools/aep-parser/test_data/re_insert_layer_" + mode + "_" + state + ".aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_insert_layer_" + mode + "_" + state + ".done");
    var log = ["mode=" + mode, "state=" + state];
    try { log.push("ae=" + app.version); } catch (e) {}

    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    var compA, compB, compC, footageF1, src, dst, clone;

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(outFile);
    });

    step("setup_shared_items", function () {
        if (mode === "footage") {
            // Solid footage as shared asset (no external file required).
            footageF1 = app.project.items.addSolid([0.25, 0.25, 0.75], "F1_shared", 200, 200, 1, 5);
        } else if (mode === "precomp") {
            compC = app.project.items.addComp("compC_precomp", 320, 240, 1, 5, 24);
            compC.layers.addSolid([0.5, 0.5, 0.5], "compC_filler", 100, 100, 1);
        }
    });

    step("setup_compA", function () {
        compA = app.project.items.addComp("compA_src_holder", 1920, 1080, 1, 5, 24);
        if (mode === "basic") {
            src = compA.layers.addSolid([1, 0, 0], "Src", 100, 100, 1);
        } else if (mode === "footage") {
            src = compA.layers.add(footageF1);
            src.name = "Src_footage";
        } else if (mode === "precomp") {
            src = compA.layers.add(compC);
            src.name = "Src_precomp";
        }
    });

    step("setup_compB", function () {
        compB = app.project.items.addComp("compB_dest", 1920, 1080, 1, 5, 24);
        compB.layers.addSolid([0, 1, 0], "B_unrelated", 100, 100, 1);
        if (mode === "footage") {
            // pre-touch compB's relationship with F1 (used elsewhere already)
            var extra = compB.layers.add(footageF1);
            extra.enabled = false;
        } else if (mode === "precomp") {
            var extra2 = compB.layers.add(compC);
            extra2.enabled = false;
        }
    });

    if (state === "after") {
        step("perform_insert", function () {
            // Insert clone of compA's Src at top of compB.
            clone = src.copyToComp(compB);
            // copyToComp inserts at the top of dest (index 1 in AE 1-based).
        });
    }

    step("save", function () {
        app.project.save(outFile);
    });

    step("write_done", function () {
        doneFile.open("w");
        doneFile.write(log.join("\n"));
        doneFile.close();
    });
})();
```

- [ ] **Step 2: Verify JSX is syntactically loadable (smoke test)**

Run (PowerShell): `Get-Content test_data/re_insert_layer.jsx | Select-Object -First 5`
Expected: header comment visible — confirms file written, no encoding mismatch.

- [ ] **Step 3: Commit JSX (the `.aep` outputs come from user-run)**

```bash
git add test_data/re_insert_layer.jsx
git commit -m "chore(aep): V3 Phase 5C re_insert_layer.jsx (3 modes × 2 states)"
```

### Task C.2: Document the user-run sequence (handoff to ship-gate)

**Files:**
- Modify: `flightdeck/cockpit.md` (update Next session)

- [ ] **Step 1: Update cockpit's Next session block**

In `flightdeck/cockpit.md`, replace the existing `## Next session` block with:

```markdown
## Next session

1. **User runs `re_insert_layer.jsx`** under AE 2020 + AE 2025 — 12 invocations total:
   `for mode in basic footage precomp; for state in before after: $env:RE_INSERT_MODE=$mode; $env:RE_INSERT_STATE=$state; afterfx.exe -r test_data/re_insert_layer.jsx`
   Produces `test_data/re_insert_layer_{basic,footage,precomp}_{before,after}.aep` (6 files).
2. **Re-run Go tests** to lift fixture skips: `go test -count=1 ./internal/aep/ -run TestInsertLayer -v` — expect 11 refuse + 6 happy/structural PASS.
3. **AE ship-gate** — `scripts/ae_run.ps1` opens each `ge_insert_layer_<mode>.aep` (Go-emitted post-InsertLayer) in both AE versions and byte-diffs against the `_after` baseline. 6/6 PASS → promote to Stable.
4. **Promote godoc tag** Alpha → Stable in `internal/aep/insert_layer.go` + add coverage row.

**Phase 5 后续候选**（5C 完后回到三选一）：
- **`Project.DuplicateItem(item Item, name string)`** — comp / footage / folder 通用
- **V2.2.1 ShapeLayer 拓展**
- Phase 5C.1 cross-Project InsertLayer
```

- [ ] **Step 2: Commit cockpit refresh + bump Last updated**

Update the `**Last updated**: 2026-05-29 by claude (Phase 5C InsertLayer design committed, awaiting user review → plan)` line to:

```markdown
**Last updated**: 2026-05-29 by claude (Phase 5C InsertLayer impl + JSX shipped Alpha; awaiting AE ship-gate)
**Active focus**: V3 Phase 5C InsertLayer Alpha — Go-side complete (refuse R1-R11 + happy-path × 3 splice positions + round-trip + concurrent-mutate). Waiting on user JSX run → AE ship-gate (3 modes × 2 versions = 6 PASS) for Stable promotion.
```

```bash
git add flightdeck/cockpit.md
git commit -m "chore(flightdeck): cockpit refresh — Phase 5C InsertLayer Alpha shipped, awaiting ship-gate"
```

---

# Phase D — Coverage doc + Alpha tag finalize

### Task D.1: Coverage doc row + godoc Alpha tag confirmation

**Files:**
- Modify: `flightdeck/flight-plans/coverage.md`
- Modify: `internal/aep/insert_layer.go` (verify Alpha tag wording)

- [ ] **Step 1: Open coverage.md and locate the structural mutation API section**

Run: `grep -n 'structural mutation\|DuplicateLayer\|DeleteLayer\|MoveLayer' flightdeck/flight-plans/coverage.md`

- [ ] **Step 2: Add the InsertLayer row**

Insert (typically alongside DuplicateLayer / MoveLayer / DeleteLayer rows — adapt the column shape to whatever already exists):

```markdown
| `Composition.InsertLayer(src, atIdx) (*Layer, error)` | **Alpha** | Cross-comp deep-clone, same-Project; refuses non-AV, cross-Project, same-comp, direct pre-comp loop, structural corruption. Ship-gate pending: 3 modes × 2 AE versions. |
```

(Exact markdown format — header levels, column count — must mirror existing rows; check the file first.)

- [ ] **Step 3: Confirm `internal/aep/insert_layer.go` godoc opens with "ALPHA"**

Run: `grep -n 'ALPHA' internal/aep/insert_layer.go`
Expected: at least one match in the godoc block (verifies tag present).

- [ ] **Step 4: `go vet` + full test pass**

Run:
```bash
go vet ./...
go test -count=1 ./internal/aep/...
```

Expected: clean vet, all tests PASS (InsertLayer tests SKIP if fixtures absent — that's the expected handoff state).

- [ ] **Step 5: Commit**

```bash
git add flightdeck/flight-plans/coverage.md
git commit -m "docs(aep): coverage — InsertLayer Alpha row pending ship-gate"
```

---

## Self-review checklist

Before declaring the plan complete the implementor should verify:

1. **Spec §2 refuse-matrix coverage**: 11 entries R1–R11 ↔ 11 dedicated tests. (Phase A tasks A.1–A.3.)
2. **Spec §3 per-byte mutations**: 4 writes (newID @0x00, TrackMatte @0x6B, ParentID @0x84, explicit matte @0xA0). All 4 land in Task B.2 with the @0xA4 length guard.
3. **Spec §3 step 7 splice branches**: empty (`insertLayrPosition`), mid (`indexOfChunk(target.layrList)`), append (last + adaptive scan). All 3 in Task B.2 + tested in B.3.
4. **Spec §5 atomic invariants**: snapshot quadruple (dest children / dest layers / nextItemID / warnings len) + warnings-as-failure rollback present in Task B.2 final block.
5. **Spec §7 file map**: `insert_layer.go` + `insert_layer_test.go` created (no `findFirstLayerSlot` — replaced with existing `insertLayrPosition`).
6. **Spec §6 ship-gate**: 3 modes JSX in Task C.1, user-run handoff documented in C.2; AE ship-gate execution itself is outside plan scope (user-triggered).
7. **Alpha tag in godoc**: present in Task B.2 impl block; verified in Task D.1.
8. **No placeholders**: every step contains actual code, expected output, and a commit command.

If any of the above don't tie out at the end of a task, the implementor should fix inline before moving on.

---

## Execution handoff

**Plan complete. Saved to `flightdeck/flight-plans/2026-05-29-v3-phase5c-insertlayer-plan.md`.**

Two execution options:

1. **Subagent-Driven (recommended)** — dispatch a fresh subagent per task, review between tasks. Best for catching impl drift early; tasks A.1 → D.1 are independent enough that parallelism on Phase A tests is possible after the scaffold lands.

2. **Inline Execution** — execute tasks in this session using `superpowers:executing-plans`, batched commits with checkpoints after Phase B.2 (full impl) and Phase D.1 (handoff).

Which approach?
