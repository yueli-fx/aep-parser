package aep_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
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
