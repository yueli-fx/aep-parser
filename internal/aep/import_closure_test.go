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
	if destProj.CompositionByID(destID) != nil {
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
