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

func TestInsertLayer_HappyPath_Basic_MidAndAppend(t *testing.T) {
	for _, tc := range []struct {
		name    string
		atIdxFn func(dest *aep.Composition) int
	}{
		{
			name:    "middle",
			atIdxFn: func(dest *aep.Composition) int { return len(dest.Layers) / 2 },
		},
		{
			name:    "append",
			atIdxFn: func(dest *aep.Composition) int { return len(dest.Layers) },
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
			if dest.Layers[atIdx] != clone {
				t.Errorf("clone not at expected slot %d for %s; dest.Layers=%v", atIdx, tc.name, layerIDs(dest.Layers))
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
