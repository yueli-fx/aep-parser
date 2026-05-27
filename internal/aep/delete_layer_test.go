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

// All tests use re_delete_layer_*.aep produced by test_data/re_delete_layer.jsx
// (see scars/ae-deletelayer-re.md). Fixtures are gitignored; t.Skipf when missing.
// To regenerate: set $env:RE_DELETE_MODE = "baseline"/"middle"/"parent"/"matte"
// then run scripts/ae_run.ps1 against the JSX (baseline/middle/parent under
// AE 2020; matte under AE 2025 — TrackMatteLayerID field is AE 23+).

const deleteLayerFixtureDir = "../../test_data"

func openDeleteLayerFixture(t *testing.T, mode string) *aep.Project {
	t.Helper()
	path := filepath.Join(deleteLayerFixtureDir, "re_delete_layer_"+mode+".aep")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("fixture missing: %s (run test_data/re_delete_layer.jsx with RE_DELETE_MODE=%s)", path, mode)
		return nil
	}
	proj, err := aep.Open(path)
	if err != nil {
		t.Fatalf("aep.Open(%s): %v", path, err)
	}
	return proj
}

func TestDeleteLayer_RefuseOutOfRange(t *testing.T) {
	proj := openDeleteLayerFixture(t, "baseline")
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	preLen := len(c.Layers)

	for _, idx := range []int{-1, preLen, preLen + 50} {
		err := c.DeleteLayer(idx)
		if err == nil {
			t.Errorf("DeleteLayer(%d): expected error, got nil", idx)
			continue
		}
		if !strings.Contains(err.Error(), "out of range") {
			t.Errorf("DeleteLayer(%d): error should mention 'out of range', got %q", idx, err.Error())
		}
	}
	if len(c.Layers) != preLen {
		t.Errorf("Layers count changed after refused delete: got %d, want %d", len(c.Layers), preLen)
	}
}

func TestDeleteLayer_RefuseMissingBackref(t *testing.T) {
	c := &aep.Composition{
		Layers: []*aep.Layer{
			{ID: 1, Type: aep.LayerTypeAV},
			{ID: 2, Type: aep.LayerTypeAV},
		},
	}
	err := c.DeleteLayer(0)
	if err == nil {
		t.Fatal("expected refuse on missing itemList back-ref, got nil")
	}
	if !strings.Contains(err.Error(), "itemList back-ref") {
		t.Errorf("error should mention 'itemList back-ref', got %q", err.Error())
	}
}

func TestDeleteLayer_RefuseLastLayer(t *testing.T) {
	proj := openDeleteLayerFixture(t, "baseline")
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	// Shrink to single layer artificially — refuse should fire BEFORE
	// any structural inspection of itemList, so the inconsistency
	// between c.Layers and itemList doesn't matter.
	c.Layers = c.Layers[:1]
	err := c.DeleteLayer(0)
	if err == nil {
		t.Fatal("expected refuse on single-layer comp, got nil")
	}
	if !strings.Contains(err.Error(), "last layer") {
		t.Errorf("error should mention 'last layer', got %q", err.Error())
	}
}

func TestDeleteLayer_RefuseNonAV(t *testing.T) {
	proj := openDeleteLayerFixture(t, "baseline")
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	c.Layers[0].Type = aep.LayerTypeCamera
	err := c.DeleteLayer(0)
	if err == nil {
		t.Fatal("expected refuse on non-AV layer, got nil")
	}
	if !strings.Contains(err.Error(), "non-AV") {
		t.Errorf("error should mention 'non-AV', got %q", err.Error())
	}
	if len(c.Layers) != 3 {
		t.Errorf("Layers count changed after refused delete: got %d, want 3", len(c.Layers))
	}
}

func TestDeleteLayer_MiddleSplice(t *testing.T) {
	proj := openDeleteLayerFixture(t, "baseline")
	if proj == nil {
		return
	}
	c := proj.Compositions[0]

	if len(c.Layers) != 3 {
		t.Fatalf("baseline fixture: want 3 layers, got %d", len(c.Layers))
	}
	preL1, preL3 := c.Layers[0], c.Layers[2] // L2_mid is at index 1
	itemList := c.ItemListForTest()
	if itemList == nil {
		t.Fatal("itemList back-ref nil")
	}
	preChildCount := len(itemList.Children)

	if err := c.DeleteLayer(1); err != nil {
		t.Fatalf("DeleteLayer(1): %v", err)
	}

	if len(c.Layers) != 2 {
		t.Fatalf("after delete: want 2 layers, got %d", len(c.Layers))
	}
	if c.Layers[0] != preL1 {
		t.Errorf("c.Layers[0] pointer: surviving L1 changed identity")
	}
	if c.Layers[1] != preL3 {
		t.Errorf("c.Layers[1] pointer: surviving L3 changed identity")
	}

	postChildCount := len(itemList.Children)
	delta := preChildCount - postChildCount
	// AE-saved baseline carries the 16-chunk-per-layer pattern (F4):
	// Layr + Ewst + 14 follower leaves. Adaptive splice consumes them.
	if delta != 16 {
		t.Errorf("itemList children delta: got %d, want 16 (16-chunk AE-saved block)", delta)
	}
}

func TestDeleteLayer_ResetsParentIDOnNeighbor(t *testing.T) {
	proj := openDeleteLayerFixture(t, "baseline")
	if proj == nil {
		return
	}
	c := proj.Compositions[0]

	l1, l2 := c.Layers[0], c.Layers[1]
	if err := l1.SetParent(l2.ID); err != nil {
		t.Fatalf("SetParent setup: %v", err)
	}
	if l1.ParentID != l2.ID {
		t.Fatalf("setup: l1.ParentID = %d, want %d", l1.ParentID, l2.ID)
	}

	if err := c.DeleteLayer(1); err != nil {
		t.Fatalf("DeleteLayer(1): %v", err)
	}

	if l1.ParentID != 0 {
		t.Errorf("l1.ParentID after delete: got %d, want 0", l1.ParentID)
	}
	ldta := l1.LdtaForTest()
	if ldta == nil {
		t.Fatal("l1 ldta nil")
	}
	if got := binary.BigEndian.Uint32(ldta.Data[0x84:0x88]); got != 0 {
		t.Errorf("l1 ldta @0x84..0x87: got %d, want 0", got)
	}
}

func TestDeleteLayer_ResetsTrackMatteIDButPreservesTypeByte(t *testing.T) {
	proj := openDeleteLayerFixture(t, "baseline")
	if proj == nil {
		return
	}
	c := proj.Compositions[0]

	l1, l3 := c.Layers[0], c.Layers[2]
	if l1.LdtaForTest() == nil || len(l1.LdtaForTest().Data) < 0xA4 {
		t.Skip("L1 ldta too short for TrackMatteLayerID write (AE 2020 baseline — field is AE 23+); RE'd in matte fixture instead")
	}
	if err := l1.SetTrackMatteLayer(l3.ID, aep.TrackMatteAlpha); err != nil {
		t.Fatalf("SetTrackMatteLayer setup: %v", err)
	}
	if l1.TrackMatteLayerID != l3.ID {
		t.Fatalf("setup: l1.TrackMatteLayerID = %d, want %d", l1.TrackMatteLayerID, l3.ID)
	}

	typeByteBefore := l1.LdtaForTest().Data[0x6B]
	if typeByteBefore != byte(aep.TrackMatteAlpha) {
		t.Fatalf("setup: l1 ldta @0x6B = 0x%02x, want 0x%02x", typeByteBefore, byte(aep.TrackMatteAlpha))
	}

	if err := c.DeleteLayer(2); err != nil {
		t.Fatalf("DeleteLayer(2): %v", err)
	}

	if l1.TrackMatteLayerID != 0 {
		t.Errorf("l1.TrackMatteLayerID after delete: got %d, want 0", l1.TrackMatteLayerID)
	}
	if got := binary.BigEndian.Uint32(l1.LdtaForTest().Data[0xA0:0xA4]); got != 0 {
		t.Errorf("l1 ldta @0xA0..0xA3: got %d, want 0", got)
	}
	// F3 finding — type byte STAYS SET after matte source is deleted.
	if got := l1.LdtaForTest().Data[0x6B]; got != typeByteBefore {
		t.Errorf("l1 ldta @0x6B (TrackMatte type): got 0x%02x, want 0x%02x (unchanged per F3)", got, typeByteBefore)
	}
}

// AE 25 variant — uses the matte fixture (AE 2025-saved, ldta long
// enough for TrackMatteLayerID @0xA0..0xA3). The fixture's post-delete
// state has 2 layers (L2, L3); we re-attach an explicit matte ref then
// delete the source to exercise the AE 23+ orphan-cleanup branch.
func TestDeleteLayer_ResetsTrackMatteID_AE25(t *testing.T) {
	proj := openDeleteLayerFixture(t, "matte")
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	if len(c.Layers) != 2 {
		t.Fatalf("matte fixture: want 2 surviving layers, got %d", len(c.Layers))
	}
	l2, l3 := c.Layers[0], c.Layers[1] // matte fixture post-AE-delete: L2, L3
	if l2.LdtaForTest() == nil || len(l2.LdtaForTest().Data) < 0xA4 {
		t.Fatalf("matte fixture L2 ldta too short for @0xA0..0xA3 (AE 25 should write 224B ldta)")
	}

	if err := l2.SetTrackMatteLayer(l3.ID, aep.TrackMatteAlpha); err != nil {
		t.Fatalf("SetTrackMatteLayer setup: %v", err)
	}
	if l2.TrackMatteLayerID != l3.ID {
		t.Fatalf("setup: l2.TrackMatteLayerID = %d, want %d", l2.TrackMatteLayerID, l3.ID)
	}
	typeByteBefore := l2.LdtaForTest().Data[0x6B]

	// L3 is at index 1; deleting it would leave 1 layer in c.Layers —
	// but the refuse fires only at len==1, so a 2→1 delete is OK.
	if err := c.DeleteLayer(1); err != nil {
		t.Fatalf("DeleteLayer(1): %v", err)
	}

	if l2.TrackMatteLayerID != 0 {
		t.Errorf("l2.TrackMatteLayerID after delete: got %d, want 0", l2.TrackMatteLayerID)
	}
	if got := binary.BigEndian.Uint32(l2.LdtaForTest().Data[0xA0:0xA4]); got != 0 {
		t.Errorf("l2 ldta @0xA0..0xA3: got %d, want 0", got)
	}
	if got := l2.LdtaForTest().Data[0x6B]; got != typeByteBefore {
		t.Errorf("l2 ldta @0x6B (TrackMatte type): got 0x%02x, want 0x%02x (unchanged per F3)", got, typeByteBefore)
	}
}

func TestDeleteLayer_RoundTrip(t *testing.T) {
	proj := openDeleteLayerFixture(t, "baseline")
	if proj == nil {
		return
	}
	c := proj.Compositions[0]

	if err := c.DeleteLayer(1); err != nil {
		t.Fatalf("DeleteLayer(1): %v", err)
	}
	wantIDs := []uint32{c.Layers[0].ID, c.Layers[1].ID}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}

	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader after delete+write: %v", err)
	}
	if len(proj2.Compositions) != 1 {
		t.Fatalf("re-opened: comps=%d, want 1", len(proj2.Compositions))
	}
	c2 := proj2.Compositions[0]
	if len(c2.Layers) != 2 {
		t.Fatalf("re-opened: layers=%d, want 2", len(c2.Layers))
	}
	gotIDs := []uint32{c2.Layers[0].ID, c2.Layers[1].ID}
	if gotIDs[0] != wantIDs[0] || gotIDs[1] != wantIDs[1] {
		t.Errorf("re-opened layer IDs: got %v, want %v", gotIDs, wantIDs)
	}
}
