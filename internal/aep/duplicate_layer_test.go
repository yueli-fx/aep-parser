package aep_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// Most happy-path tests reuse the DeleteLayer baseline fixture
// (re_delete_layer_baseline.aep, 3 solid layers L1_top / L2_mid / L3_bot,
// no refs) — identical structural setup to DuplicateLayer's baseline
// state (DuplicateLayer JSX only saves the post-dup state per fixture).
// Tests are skipped when the fixture is absent.

const dupLayerFixtureDir = "../../test_data"

func openDupBaseline(t *testing.T) *aep.Project {
	t.Helper()
	path := filepath.Join(dupLayerFixtureDir, "re_delete_layer_baseline.aep")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("fixture missing: %s (3-solid baseline; produced by test_data/re_delete_layer.jsx RE_DELETE_MODE=baseline)", path)
		return nil
	}
	proj, err := aep.Open(path)
	if err != nil {
		t.Fatalf("aep.Open(%s): %v", path, err)
	}
	return proj
}

func openDupFixture(t *testing.T, mode string) *aep.Project {
	t.Helper()
	path := filepath.Join(dupLayerFixtureDir, "re_duplicate_layer_"+mode+".aep")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("fixture missing: %s (run test_data/re_duplicate_layer.jsx with RE_DUP_MODE=%s)", path, mode)
		return nil
	}
	proj, err := aep.Open(path)
	if err != nil {
		t.Fatalf("aep.Open(%s): %v", path, err)
	}
	return proj
}

func TestDuplicateLayer_RefuseEmptyName(t *testing.T) {
	proj := openDupBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	_, err := c.DuplicateLayer(1, "")
	if err == nil {
		t.Fatal("expected refuse on empty name, got nil")
	}
	if !strings.Contains(err.Error(), "name cannot be empty") {
		t.Errorf("error should mention 'name cannot be empty', got %q", err.Error())
	}
}

func TestDuplicateLayer_RefuseOutOfRange(t *testing.T) {
	proj := openDupBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	preLen := len(c.Layers)
	for _, idx := range []int{-1, preLen, preLen + 50} {
		_, err := c.DuplicateLayer(idx, "X")
		if err == nil {
			t.Errorf("DuplicateLayer(%d): expected error, got nil", idx)
			continue
		}
		if !strings.Contains(err.Error(), "out of range") {
			t.Errorf("DuplicateLayer(%d): error should mention 'out of range', got %q", idx, err.Error())
		}
	}
	if len(c.Layers) != preLen {
		t.Errorf("Layers count changed after refused dup: got %d, want %d", len(c.Layers), preLen)
	}
}

func TestDuplicateLayer_RefuseMissingBackref(t *testing.T) {
	c := &aep.Composition{
		Layers: []*aep.Layer{
			{ID: 1, Type: aep.LayerTypeAV},
			{ID: 2, Type: aep.LayerTypeAV},
		},
	}
	_, err := c.DuplicateLayer(0, "X")
	if err == nil {
		t.Fatal("expected refuse on missing itemList back-ref, got nil")
	}
	if !strings.Contains(err.Error(), "itemList back-ref") {
		t.Errorf("error should mention 'itemList back-ref', got %q", err.Error())
	}
}

func TestDuplicateLayer_RefuseNonAV(t *testing.T) {
	proj := openDupBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	c.Layers[0].Type = aep.LayerTypeCamera
	_, err := c.DuplicateLayer(0, "X")
	if err == nil {
		t.Fatal("expected refuse on non-AV layer, got nil")
	}
	if !strings.Contains(err.Error(), "non-AV") {
		t.Errorf("error should mention 'non-AV', got %q", err.Error())
	}
	if len(c.Layers) != 3 {
		t.Errorf("Layers count changed after refused dup: got %d, want 3", len(c.Layers))
	}
}

func TestDuplicateLayer_RefuseTrackMatte(t *testing.T) {
	proj := openDupBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	c.Layers[1].TrackMatte = aep.TrackMatteAlpha
	_, err := c.DuplicateLayer(1, "X")
	if err == nil {
		t.Fatal("expected refuse on layer with TrackMatte set, got nil")
	}
	if !strings.Contains(err.Error(), "TrackMatte") {
		t.Errorf("error should mention 'TrackMatte', got %q", err.Error())
	}
	if len(c.Layers) != 3 {
		t.Errorf("Layers count changed after refused dup: got %d, want 3", len(c.Layers))
	}
}

// Happy path: clone has new ID, supplied name, source pushed down,
// itemList grew by the per-layer block size (16 for AE-saved baseline).
func TestDuplicateLayer_HappyPath_Middle(t *testing.T) {
	proj := openDupBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	if len(c.Layers) != 3 {
		t.Fatalf("baseline: want 3 layers, got %d", len(c.Layers))
	}

	srcL2 := c.Layers[1]
	srcID := srcL2.ID
	srcSourceID := srcL2.SourceID
	srcParentID := srcL2.ParentID
	preChildCount := len(c.ItemListForTest().Children)
	preNextItemID := proj.NextItemIDForTest()

	clone, err := c.DuplicateLayer(1, "L2_clone")
	if err != nil {
		t.Fatalf("DuplicateLayer(1, \"L2_clone\"): %v", err)
	}
	if clone == nil {
		t.Fatal("DuplicateLayer returned nil clone with no error")
	}

	// Count + positioning.
	if len(c.Layers) != 4 {
		t.Fatalf("post-dup: want 4 layers, got %d", len(c.Layers))
	}
	if c.Layers[1] != clone {
		t.Errorf("c.Layers[1] should be the clone (pushed source down to idx 2)")
	}
	if c.Layers[2] != srcL2 {
		t.Errorf("c.Layers[2] should be the original source L2 (pushed down)")
	}

	// Identity fields.
	if clone.Name != "L2_clone" {
		t.Errorf("clone.Name: got %q, want %q", clone.Name, "L2_clone")
	}
	if clone.ID == srcID {
		t.Errorf("clone.ID must differ from source (collision); both = %d", clone.ID)
	}
	if clone.ID != preNextItemID {
		t.Errorf("clone.ID: got %d, want %d (preNextItemID = head counter +1)", clone.ID, preNextItemID)
	}
	// Verbatim body copies (F4 + F7).
	if clone.SourceID != srcSourceID {
		t.Errorf("clone.SourceID: got %d, want %d (verbatim from source per F4)", clone.SourceID, srcSourceID)
	}
	if clone.ParentID != srcParentID {
		t.Errorf("clone.ParentID: got %d, want %d (verbatim from source per F7)", clone.ParentID, srcParentID)
	}

	// itemList growth = block size (16 for AE-saved baseline per F5).
	postChildCount := len(c.ItemListForTest().Children)
	delta := postChildCount - preChildCount
	if delta != 16 {
		t.Errorf("itemList children delta: got %d, want 16 (AE-saved block per F5)", delta)
	}

	// proj.nextItemID bumped.
	postNextItemID := proj.NextItemIDForTest()
	if postNextItemID != preNextItemID+1 {
		t.Errorf("proj.nextItemID: got %d, want %d (pre+1)", postNextItemID, preNextItemID+1)
	}
}

// Concurrent-mutate safety: cloned chunks must not share Data slices with
// source — per scars/concurrency-unsafe-shared-chunk-bytes.md. Verify by
// mutating clone's ldta and checking source's ldta is untouched.
func TestDuplicateLayer_FreshDataSlices(t *testing.T) {
	proj := openDupBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	src := c.Layers[1]
	srcLdtaBefore := append([]byte(nil), src.LdtaForTest().Data...)

	clone, err := c.DuplicateLayer(1, "Clone")
	if err != nil {
		t.Fatalf("DuplicateLayer: %v", err)
	}
	cloneLdta := clone.LdtaForTest()
	if cloneLdta == nil {
		t.Fatal("clone has no ldta backref")
	}
	if &cloneLdta.Data[0] == &src.LdtaForTest().Data[0] {
		t.Fatal("clone ldta shares Data slice header with source — must be fresh allocation")
	}

	// Mutate every byte of clone's ldta past the ID field — source must remain pristine.
	for i := 4; i < len(cloneLdta.Data); i++ {
		cloneLdta.Data[i] ^= 0xFF
	}
	srcLdtaAfter := src.LdtaForTest().Data
	if !bytes.Equal(srcLdtaBefore, srcLdtaAfter) {
		t.Fatal("source ldta bytes changed after mutating clone — Data slice sharing detected")
	}
}

// Round-trip: Open → Dup → Write → Reopen produces a parseable file with
// the expected 4 layers in display order.
func TestDuplicateLayer_RoundTrip(t *testing.T) {
	proj := openDupBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	if _, err := c.DuplicateLayer(1, "L2_clone"); err != nil {
		t.Fatalf("DuplicateLayer(1): %v", err)
	}
	// After dup: [L1, clone, source_L2, L3]
	wantIDs := []uint32{c.Layers[0].ID, c.Layers[1].ID, c.Layers[2].ID, c.Layers[3].ID}
	wantNames := []string{c.Layers[0].Name, c.Layers[1].Name, c.Layers[2].Name, c.Layers[3].Name}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}

	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader after dup+write: %v", err)
	}
	if len(proj2.Compositions) != 1 {
		t.Fatalf("re-opened: comps=%d, want 1", len(proj2.Compositions))
	}
	c2 := proj2.Compositions[0]
	if len(c2.Layers) != 4 {
		t.Fatalf("re-opened: layers=%d, want 4", len(c2.Layers))
	}
	for i, want := range wantIDs {
		if c2.Layers[i].ID != want {
			t.Errorf("re-opened Layers[%d].ID: got %d, want %d", i, c2.Layers[i].ID, want)
		}
	}
	for i, want := range wantNames {
		if c2.Layers[i].Name != want {
			t.Errorf("re-opened Layers[%d].Name: got %q, want %q", i, c2.Layers[i].Name, want)
		}
	}
}

// Cloned ldta @0x00..0x03 carries the NEW ID; @0x04+ verbatim from source
// (Finding 10 — the only mutation is the ID field).
func TestDuplicateLayer_LdtaBodyVerbatim(t *testing.T) {
	proj := openDupBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	src := c.Layers[1]
	srcLdtaBytes := append([]byte(nil), src.LdtaForTest().Data...)

	clone, err := c.DuplicateLayer(1, "Clone")
	if err != nil {
		t.Fatalf("DuplicateLayer: %v", err)
	}
	cloneLdtaBytes := clone.LdtaForTest().Data
	if len(cloneLdtaBytes) != len(srcLdtaBytes) {
		t.Fatalf("ldta length differs: clone=%d, src=%d", len(cloneLdtaBytes), len(srcLdtaBytes))
	}
	gotCloneID := binary.BigEndian.Uint32(cloneLdtaBytes[0x00:0x04])
	if gotCloneID != clone.ID {
		t.Errorf("clone ldta @0x00..0x03: got %d, want %d (clone.ID)", gotCloneID, clone.ID)
	}
	// Everything past the ID field must match source byte-for-byte (F10).
	if !bytes.Equal(srcLdtaBytes[4:], cloneLdtaBytes[4:]) {
		// Find the first divergence for a useful error.
		for i := 4; i < len(srcLdtaBytes); i++ {
			if srcLdtaBytes[i] != cloneLdtaBytes[i] {
				t.Fatalf("ldta @0x%02X: src=0x%02X clone=0x%02X — F10 says only @0x00..0x03 should differ", i, srcLdtaBytes[i], cloneLdtaBytes[i])
			}
		}
	}
}

// Structural equivalence: Go's Open(re_delete_layer_baseline) →
// DuplicateLayer(1, "L2_mid") should produce an itemList shape (chunk
// IDs / IsList / FormType in order) matching AE's re_duplicate_layer_solo
// post-dup state. Byte-identical comparison is impossible (timestamp /
// UUID variance + different JSX scripts produced the two baselines), but
// chunk shape is the deterministic surface that decides AE-side parse
// success.
func TestDuplicateLayer_StructuralEquivalence_Solo(t *testing.T) {
	projBaseline := openDupBaseline(t)
	if projBaseline == nil {
		return
	}
	projSolo := openDupFixture(t, "solo")
	if projSolo == nil {
		return
	}

	cb := projBaseline.Compositions[0]
	if _, err := cb.DuplicateLayer(1, "L2_mid"); err != nil {
		t.Fatalf("DuplicateLayer(1, \"L2_mid\"): %v", err)
	}
	cs := projSolo.Compositions[0]

	if len(cb.Layers) != len(cs.Layers) {
		t.Errorf("layer count: Go=%d, AE-solo=%d", len(cb.Layers), len(cs.Layers))
	}

	goChildren := cb.ItemListForTest().Children
	aeChildren := cs.ItemListForTest().Children

	if len(goChildren) != len(aeChildren) {
		t.Fatalf("itemList children count: Go=%d, AE-solo=%d (F5 says AE post-dup = baseline+16)", len(goChildren), len(aeChildren))
	}
	for i, ae := range aeChildren {
		goCh := goChildren[i]
		if goCh.ID != ae.ID {
			t.Errorf("children[%d] ID: Go=%q, AE-solo=%q", i, string(goCh.ID[:]), string(ae.ID[:]))
		}
		if goCh.IsList() != ae.IsList() {
			t.Errorf("children[%d] IsList: Go=%v, AE-solo=%v", i, goCh.IsList(), ae.IsList())
		}
		if goCh.IsList() && goCh.FormType != ae.FormType {
			t.Errorf("children[%d] FormType: Go=%q, AE-solo=%q", i, string(goCh.FormType[:]), string(ae.FormType[:]))
		}
	}
}
