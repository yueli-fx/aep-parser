package aep_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const dupItemFixtureDir = "../../test_data"

// openDupItemProject opens the RE baseline that holds compA_main (3 AV layers:
// A_precomp→compC, A_footage→F1, A_solid with A_solid.parent = A_precomp) plus
// the shared precomp/footage. Returns (proj, compA_main). Skips if missing.
func openDupItemProject(t *testing.T) (*aep.Project, *aep.Composition) {
	t.Helper()
	// Prefer the dedicated before baseline; fall back to after (also holds
	// a valid compA_main alongside its AE-made dup).
	for _, fn := range []string{"re_duplicate_item_before.aep", "re_duplicate_item_after.aep"} {
		path := filepath.Join(dupItemFixtureDir, fn)
		if _, err := os.Stat(path); err != nil {
			continue
		}
		proj, err := aep.Open(path)
		if err != nil {
			t.Fatalf("aep.Open(%s): %v", path, err)
		}
		for _, c := range proj.Compositions {
			if c.Name == "compA_main" {
				return proj, c
			}
		}
		t.Fatalf("fixture %s: comp 'compA_main' not found", path)
	}
	t.Skipf("fixture missing: re_duplicate_item_{before,after}.aep (run test_data/re_duplicate_item.jsx)")
	return nil, nil
}

func compLayerIDs(c *aep.Composition) []uint32 {
	ids := make([]uint32, len(c.Layers))
	for i, l := range c.Layers {
		ids[i] = l.ID
	}
	return ids
}

func TestDuplicateComposition_RefuseNilSrc(t *testing.T) {
	p := &aep.Project{}
	_, err := aep.DuplicateComposition(p, nil, "x")
	if err == nil || !strings.Contains(err.Error(), "src cannot be nil") {
		t.Fatalf("want 'src cannot be nil', got %v", err)
	}
}

func TestDuplicateComposition_RefuseEmptyName(t *testing.T) {
	proj, src := openDupItemProject(t)
	if proj == nil {
		return
	}
	_, err := aep.DuplicateComposition(proj, src, "")
	if err == nil || !strings.Contains(err.Error(), "name cannot be empty") {
		t.Fatalf("want 'name cannot be empty', got %v", err)
	}
}

func TestDuplicateComposition_RefuseCrossProject(t *testing.T) {
	proj, src := openDupItemProject(t)
	if proj == nil {
		return
	}
	other, _ := openDupItemProject(t)
	if other == nil {
		return
	}
	// src belongs to proj; calling on other must refuse.
	_, err := aep.DuplicateComposition(other, src, "x")
	if err == nil || !strings.Contains(err.Error(), "does not belong to this Project") {
		t.Fatalf("want cross-Project refuse, got %v", err)
	}
}

func TestDuplicateComposition_HappyPath(t *testing.T) {
	proj, src := openDupItemProject(t)
	if proj == nil {
		return
	}
	preComps := len(proj.Compositions)
	preNextID := aep.NextItemIDForTest(proj)
	srcIDs := compLayerIDs(src)
	srcLayerCount := len(src.Layers)
	if srcLayerCount != 3 {
		t.Fatalf("fixture precondition: compA_main want 3 layers, got %d", srcLayerCount)
	}

	dup, err := aep.DuplicateComposition(proj, src, "compA_dup")
	if err != nil {
		t.Fatalf("DuplicateComposition: %v", err)
	}
	if dup == nil {
		t.Fatal("nil dup with no error")
	}
	if dup.Name != "compA_dup" {
		t.Errorf("dup.Name = %q, want %q", dup.Name, "compA_dup")
	}
	if len(proj.Compositions) != preComps+1 {
		t.Errorf("Compositions count = %d, want %d", len(proj.Compositions), preComps+1)
	}
	if dup.ID == src.ID {
		t.Errorf("dup.ID must differ from src.ID (both %d)", dup.ID)
	}
	if len(dup.Layers) != srcLayerCount {
		t.Fatalf("dup layer count = %d, want %d", len(dup.Layers), srcLayerCount)
	}

	// Every dup layer ID must be fresh (not equal to any src layer ID).
	srcIDset := map[uint32]bool{}
	for _, id := range srcIDs {
		srcIDset[id] = true
	}
	for i, l := range dup.Layers {
		if srcIDset[l.ID] {
			t.Errorf("dup.Layers[%d].ID=%d collides with a src layer ID", i, l.ID)
		}
	}

	// Source sharing: dup layers keep the same SourceID as the matching src
	// layer (positional). No item duplication.
	for i := range dup.Layers {
		if dup.Layers[i].SourceID != src.Layers[i].SourceID {
			t.Errorf("dup.Layers[%d].SourceID=%d, want %d (shared verbatim)",
				i, dup.Layers[i].SourceID, src.Layers[i].SourceID)
		}
	}

	// Parent remap: find the src layer with a non-zero ParentID; the matching
	// dup layer must parent to the DUP's own layer (an ID in dup, not src).
	dupIDset := map[uint32]bool{}
	for _, l := range dup.Layers {
		dupIDset[l.ID] = true
	}
	foundParent := false
	for i, sl := range src.Layers {
		if sl.ParentID == 0 {
			continue
		}
		foundParent = true
		dp := dup.Layers[i].ParentID
		if dp == 0 {
			t.Errorf("dup.Layers[%d].ParentID=0, want a remapped parent", i)
		}
		if dp == sl.ParentID {
			t.Errorf("dup.Layers[%d].ParentID=%d NOT remapped (still points at src layer)", i, dp)
		}
		if !dupIDset[dp] {
			t.Errorf("dup.Layers[%d].ParentID=%d is not one of the dup's own layer IDs %v", i, dp, compLayerIDs(dup))
		}
	}
	if !foundParent {
		t.Errorf("fixture precondition: expected a src layer with ParentID != 0")
	}

	if aep.NextItemIDForTest(proj) != preNextID+uint32(srcLayerCount)+1 {
		t.Errorf("nextItemID = %d, want %d (+1 comp +%d layers)",
			aep.NextItemIDForTest(proj), preNextID+uint32(srcLayerCount)+1, srcLayerCount)
	}
}

func TestDuplicateComposition_RoundTrip(t *testing.T) {
	proj, src := openDupItemProject(t)
	if proj == nil {
		return
	}
	dup, err := aep.DuplicateComposition(proj, src, "compA_dup")
	if err != nil {
		t.Fatalf("DuplicateComposition: %v", err)
	}
	wantID := dup.ID
	wantName := dup.Name
	wantLayerCount := len(dup.Layers)
	// remembered remapped parent expectation
	var wantParentNonZero bool
	for _, l := range dup.Layers {
		if l.ParentID != 0 {
			wantParentNonZero = true
		}
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader after dup+write: %v", err)
	}
	var reopened *aep.Composition
	for _, c := range proj2.Compositions {
		if c.Name == wantName && c.ID == wantID {
			reopened = c
			break
		}
	}
	if reopened == nil {
		t.Fatalf("dup comp %q (id %d) not found after round-trip", wantName, wantID)
	}
	if len(reopened.Layers) != wantLayerCount {
		t.Errorf("reopened dup layer count = %d, want %d", len(reopened.Layers), wantLayerCount)
	}
	// Parent ref must still resolve to one of the reopened dup's own layers.
	reIDset := map[uint32]bool{}
	for _, l := range reopened.Layers {
		reIDset[l.ID] = true
	}
	gotParentNonZero := false
	for i, l := range reopened.Layers {
		if l.ParentID != 0 {
			gotParentNonZero = true
			if !reIDset[l.ParentID] {
				t.Errorf("reopened dup.Layers[%d].ParentID=%d does not resolve within the dup", i, l.ParentID)
			}
		}
	}
	if wantParentNonZero != gotParentNonZero {
		t.Errorf("parent-ref presence changed across round-trip: want %v got %v", wantParentNonZero, gotParentNonZero)
	}
}

func TestDuplicateComposition_FreshDataSlices(t *testing.T) {
	proj, src := openDupItemProject(t)
	if proj == nil {
		return
	}
	// Capture src's first layer ldta bytes.
	srcLdta := aep.LdtaForTest(src.Layers[0])
	if srcLdta == nil {
		t.Skip("src layer has no ldta backref")
	}
	srcBefore := append([]byte(nil), srcLdta.Data...)

	dup, err := aep.DuplicateComposition(proj, src, "compA_dup")
	if err != nil {
		t.Fatalf("DuplicateComposition: %v", err)
	}
	dupLdta := aep.LdtaForTest(dup.Layers[0])
	if dupLdta == nil {
		t.Fatal("dup layer has no ldta backref")
	}
	if &dupLdta.Data[0] == &srcLdta.Data[0] {
		t.Fatal("dup ldta shares Data slice header with src — must be fresh allocation")
	}
	for i := 4; i < len(dupLdta.Data); i++ {
		dupLdta.Data[i] ^= 0xFF
	}
	if !bytes.Equal(srcBefore, srcLdta.Data) {
		t.Fatal("src ldta bytes changed after mutating dup — Data slice sharing detected")
	}
}
