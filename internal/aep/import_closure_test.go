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
	// re_duplicate_item_after.aep only contains solids; importFootageBlock cannot
	// locate solid items in the root Fold (they are not stored as Item LISTs).
	// Fall back to re_batch.aep which has file-backed footage.
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
	preNext := destProj.NextItemIDForTest()
	destID, err := aep.ImportFootageBlockForTest(destProj, srcProj, srcF.ID, srcF.Name)
	if err != nil {
		t.Fatalf("importFootageBlock: %v", err)
	}
	if destID != preNext {
		t.Errorf("imported footage destID = %d, want %d (head counter)", destID, preNext)
	}
	if destProj.NextItemIDForTest() != preNext+1 {
		t.Errorf("nextItemID = %d, want %d (+1)", destProj.NextItemIDForTest(), preNext+1)
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
