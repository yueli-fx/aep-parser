package aep_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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
	var destComp *aep.Composition
	for _, c := range destProj.Compositions {
		if c.Name != "compA_main" {
			destComp = c
			break
		}
	}
	if destComp == nil {
		destComp = destMain
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

	// A folder-nested item (e.g. a solid in the "Solids" folder) must also be
	// found via recursion into the Sfdr sub-container.
	var nested *aep.Footage
	for _, f := range srcProj.Footage {
		if f.IsSolid {
			nested = f
			break
		}
	}
	if nested != nil {
		if s, _ := aep.LocateItemBlockByIDForTest(srcProj, nested.ID); s < 0 {
			t.Errorf("nested solid footage id=%d not located — recursion into folder Sfdr failed", nested.ID)
		}
	}
}

// openFileBacked opens path twice, returning (srcProj, destProj) if the file
// exists and contains at least one file-backed footage item. Returns nil,nil
// if the file is missing or has no file-backed footage.
func openFileBacked(t *testing.T, path string) (*aep.Project, *aep.Project) {
	t.Helper()
	src, err := aep.Open(path)
	if err != nil {
		return nil, nil
	}
	for _, f := range src.Footage {
		if f.Path != "" && !f.IsSolid && !f.IsPlaceholder {
			dest, err2 := aep.Open(path)
			if err2 != nil {
				t.Fatalf("second open of %s: %v", path, err2)
			}
			return src, dest
		}
	}
	return nil, nil
}

// X3: dangling source — src layer references an item id absent from srcProj.
func TestInsertLayerXProj_RefuseDanglingSource(t *testing.T) {
	srcProj, srcMain, _, destComp := openXProjPair(t)
	if srcProj == nil {
		return
	}
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
	_, err := aep.InsertLayer(destComp, src, 0)
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
	_, _, _, realDest := openXProjPair(t)
	aep.SetCompProjForTest(realDest, &aep.Project{}) // bare Project: back == nil
	_, err := aep.InsertLayer(realDest, src, 0)
	if err == nil || !strings.Contains(err.Error(), "root Fold") {
		t.Fatalf("want 'root Fold' refuse, got %v", err)
	}
}

func TestInsertLayerXProj_HappyPath_Closure(t *testing.T) {
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
		t.Skip("no AV layer with a resolvable source in fixture")
	}
	preDestItems := len(destProj.Compositions) + len(destProj.Footage)
	preLayers := len(destComp.Layers)

	clone, err := aep.InsertLayer(destComp, src, 0)
	if err != nil {
		t.Fatalf("cross-Project InsertLayer: %v", err)
	}
	if clone == nil || len(destComp.Layers) != preLayers+1 || destComp.Layers[0] != clone {
		t.Fatalf("clone not inserted at slot 0 (layers=%d)", len(destComp.Layers))
	}
	postDestItems := len(destProj.Compositions) + len(destProj.Footage)
	if postDestItems <= preDestItems {
		t.Errorf("dest item count did not grow: pre=%d post=%d", preDestItems, postDestItems)
	}
	if clone.SourceID == 0 {
		t.Fatalf("clone.SourceID is 0")
	}
	if destProj.AVItemByID(clone.SourceID) == nil {
		t.Errorf("clone.SourceID=%d does not resolve in destProj (remap failed)", clone.SourceID)
	}
	if clone.ParentID != 0 || clone.TrackMatteLayerID != 0 {
		t.Errorf("clone refs not reset: ParentID=%d matte=%d", clone.ParentID, clone.TrackMatteLayerID)
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
	clone, err := aep.InsertLayer(destComp, src, 0)
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

// Footage dedup: insert a layer whose source is FILE-backed footage that also
// exists (by Path) in dest → dest footage count must NOT grow. Uses re_batch.aep
// (file footage) via openFileBacked; SKIPs if no comp layer sources file footage
// there (the ship-gate covers dedup definitively with purpose-built fixtures).
func TestInsertLayerXProj_FootageDedup(t *testing.T) {
	srcProj, destProj := openFileBacked(t, "../../test_data/re_batch.aep")
	if srcProj == nil {
		t.Skip("re_batch.aep absent or has no file-backed footage")
	}
	// find a layer (in any srcProj comp) whose source is file-backed footage
	var src *aep.Layer
	var destComp *aep.Composition
	for _, c := range srcProj.Compositions {
		for _, l := range c.Layers {
			if f, ok := srcProj.AVItemByID(l.SourceID).(*aep.Footage); ok && f.Path != "" && !f.IsSolid && !f.IsPlaceholder {
				src = l
				break
			}
		}
		if src != nil {
			break
		}
	}
	for _, c := range destProj.Compositions {
		destComp = c
		break
	}
	if src == nil || destComp == nil {
		t.Skip("no file-footage-sourced layer in re_batch.aep")
	}
	// dest opened from same file → already has the footage by Path → dedup hit
	preFootage := len(destProj.Footage)
	clone, err := aep.InsertLayer(destComp, src, 0)
	if err != nil {
		t.Fatalf("InsertLayer (dedup): %v", err)
	}
	if len(destProj.Footage) != preFootage {
		t.Errorf("footage duplicated despite path match: pre=%d post=%d", preFootage, len(destProj.Footage))
	}
	if destProj.AVItemByID(clone.SourceID) == nil {
		t.Errorf("dedup clone.SourceID=%d unresolved in dest", clone.SourceID)
	}
}

func TestImportFootageBlock(t *testing.T) {
	srcProj, _, destProj, _ := openXProjPair(t)
	if srcProj == nil {
		return
	}
	// pick any file-backed footage in srcProj
	var srcF *aep.Footage
	for _, f := range srcProj.Footage {
		if f.Path != "" && !f.IsSolid && !f.IsPlaceholder {
			srcF = f
			break
		}
	}
	// re_duplicate_item_after.aep contains only solid footage; this test imports
	// file-backed footage (the !IsSolid filter above leaves srcF nil here), so
	// fall back to re_batch.aep which has file-backed footage.
	if srcF == nil {
		batchPath := filepath.Join("../../test_data", "re_batch.aep")
		srcProj, destProj = openFileBacked(t, batchPath)
		if srcProj == nil {
			t.Skip("no file-backed footage in any available fixture")
		}
		for _, f := range srcProj.Footage {
			if f.Path != "" && !f.IsSolid && !f.IsPlaceholder {
				srcF = f
				break
			}
		}
	}
	if srcF == nil {
		t.Skip("no file-backed footage in fixture")
	}
	// openXProjPair / openFileBacked open the same file twice, so destProj
	// already has srcF's Path → it would be a dedup hit. Drop dest footage
	// sharing that Path so the fresh-import path is genuinely exercised.
	var kept []*aep.Footage
	for _, f := range destProj.Footage {
		if f.Path != srcF.Path {
			kept = append(kept, f)
		}
	}
	destProj.Footage = kept
	if aep.DestFootageByPathForTest(destProj, srcF.Path) != nil {
		t.Fatalf("precondition: dest still has footage with path %q", srcF.Path)
	}

	preCount := len(destProj.Footage)
	preNext := aep.NextItemIDForTest(destProj)
	destID, err := aep.ImportFootageBlockForTest(destProj, srcProj, srcF.ID, srcF.Name)
	if err != nil {
		t.Fatalf("importFootageBlock: %v", err)
	}
	if destID != preNext {
		t.Errorf("imported footage destID = %d, want %d (head counter)", destID, preNext)
	}
	if aep.NextItemIDForTest(destProj) != preNext+1 {
		t.Errorf("nextItemID = %d, want %d (+1)", aep.NextItemIDForTest(destProj), preNext+1)
	}
	if len(destProj.Footage) != preCount+1 {
		t.Errorf("destProj.Footage count = %d, want %d", len(destProj.Footage), preCount+1)
	}
	if destProj.CompositionByID(destID) != nil {
		t.Errorf("destID resolved as a comp, want footage")
	}
	var found *aep.Footage
	for _, f := range destProj.Footage {
		if f.ID == destID {
			found = f
		}
	}
	if found == nil {
		t.Fatalf("imported footage id=%d not found in destProj.Footage", destID)
	}
	if found.Path != srcF.Path {
		t.Errorf("imported footage Path = %q, want %q", found.Path, srcF.Path)
	}
}
