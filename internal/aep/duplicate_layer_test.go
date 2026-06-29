package aep_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// Most happy-path tests reuse the DeleteLayer baseline fixture
// (re_delete_layer_baseline.aep, 3 solid layers L1_top / L2_mid / L3_bot,
// no refs) — identical structural setup to DuplicateLayer's baseline
// state (DuplicateLayer JSX only saves the post-dup state per fixture).
// Tests are skipped when the fixture is absent.

const dupLayerFixtureDir = "../../test_data/generated/fixtures"

func openDupBaseline(t *testing.T) *aep.Project {
	t.Helper()
	path := filepath.Join(dupLayerFixtureDir, "re_delete_layer_baseline.aep")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("fixture missing: %s (3-solid baseline; produced by test_data/generators/re_delete_layer.jsx RE_DELETE_MODE=baseline)", path)
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
		t.Skipf("fixture missing: %s (run test_data/generators/re_duplicate_layer.jsx with RE_DUP_MODE=%s)", path, mode)
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
	_, err := aep.DuplicateLayer(c, 1, "")
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
		_, err := aep.DuplicateLayer(c, idx, "X")
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
	_, err := aep.DuplicateLayer(c, 0, "X")
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
	_, err := aep.DuplicateLayer(c, 0, "X")
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

// Implicit matte (TrackMatte != None && TrackMatteLayerID == 0)
// is still refused — F2 quirk applies to positional "layer-above" matte.
func TestDuplicateLayer_RefuseImplicitTrackMatte(t *testing.T) {
	proj := openDupBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	c.Layers[1].TrackMatte = aep.TrackMatteAlpha
	// TrackMatteLayerID stays 0 — implicit matte path.
	if c.Layers[1].TrackMatteLayerID != 0 {
		t.Fatalf("test precondition: TrackMatteLayerID should be 0 for implicit case, got %d", c.Layers[1].TrackMatteLayerID)
	}
	_, err := aep.DuplicateLayer(c, 1, "X")
	if err == nil {
		t.Fatal("expected refuse on layer with implicit TrackMatte, got nil")
	}
	if !strings.Contains(err.Error(), "implicit TrackMatte") {
		t.Errorf("error should mention 'implicit TrackMatte', got %q", err.Error())
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
	preChildCount := len(aep.ItemListForTest(c).Children)
	preNextItemID := aep.NextItemIDForTest(proj)

	clone, err := aep.DuplicateLayer(c, 1, "L2_clone")
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
	postChildCount := len(aep.ItemListForTest(c).Children)
	delta := postChildCount - preChildCount
	if delta != 16 {
		t.Errorf("itemList children delta: got %d, want 16 (AE-saved block per F5)", delta)
	}

	// proj.nextItemID bumped.
	postNextItemID := aep.NextItemIDForTest(proj)
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
	srcLdtaBefore := append([]byte(nil), aep.LdtaForTest(src).Data...)

	clone, err := aep.DuplicateLayer(c, 1, "Clone")
	if err != nil {
		t.Fatalf("DuplicateLayer: %v", err)
	}
	cloneLdta := aep.LdtaForTest(clone)
	if cloneLdta == nil {
		t.Fatal("clone has no ldta backref")
	}
	if &cloneLdta.Data[0] == &aep.LdtaForTest(src).Data[0] {
		t.Fatal("clone ldta shares Data slice header with source — must be fresh allocation")
	}

	// Mutate every byte of clone's ldta past the ID field — source must remain pristine.
	for i := 4; i < len(cloneLdta.Data); i++ {
		cloneLdta.Data[i] ^= 0xFF
	}
	srcLdtaAfter := aep.LdtaForTest(src).Data
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
	if _, err := aep.DuplicateLayer(c, 1, "L2_clone"); err != nil {
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
	srcLdtaBytes := append([]byte(nil), aep.LdtaForTest(src).Data...)

	clone, err := aep.DuplicateLayer(c, 1, "Clone")
	if err != nil {
		t.Fatalf("DuplicateLayer: %v", err)
	}
	cloneLdtaBytes := aep.LdtaForTest(clone).Data
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
	if _, err := aep.DuplicateLayer(cb, 1, "L2_mid"); err != nil {
		t.Fatalf("DuplicateLayer(1, \"L2_mid\"): %v", err)
	}
	cs := projSolo.Compositions[0]

	if len(cb.Layers) != len(cs.Layers) {
		t.Errorf("layer count: Go=%d, AE-solo=%d", len(cb.Layers), len(cs.Layers))
	}

	goChildren := aep.ItemListForTest(cb).Children
	aeChildren := aep.ItemListForTest(cs).Children

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

// layerIndexInComp returns the 0-based c.Layers index of `target` via
// pointer identity, or -1 if not found.
func layerIndexInComp(c *aep.Composition, target *aep.Layer) int {
	for i, l := range c.Layers {
		if l == target {
			return i
		}
	}
	return -1
}

// Happy path: duplicating an explicit-matte layer (AE 23+
// TrackMatteLayerID != 0) succeeds and produces a clone with the same
// matte source/mode. F2 position-shift quirk does NOT apply — clone is
// inserted at source's old slice index like the solo/dup_parent/dup_child
// modes.
func TestDuplicateLayer_ExplicitMatte_HappyPath(t *testing.T) {
	proj, c := openTrackMatteAE24(t)
	if proj == nil || c == nil {
		return
	}
	src := layerBySourceName(proj, c, "mt_alpha_to_solidA")
	if src == nil {
		t.Fatal("fixture missing layer 'mt_alpha_to_solidA'")
	}
	idx := layerIndexInComp(c, src)
	if idx < 0 {
		t.Fatal("can't locate mt_alpha_to_solidA in c.Layers")
	}
	if src.TrackMatte == aep.TrackMatteNone {
		t.Fatalf("test precondition: src.TrackMatte should be non-None, got %d", src.TrackMatte)
	}
	if src.TrackMatteLayerID == 0 {
		t.Fatalf("test precondition: src.TrackMatteLayerID should be non-zero (explicit matte), got 0")
	}
	srcMode := src.TrackMatte
	srcMatteID := src.TrackMatteLayerID
	preLayerCount := len(c.Layers)
	preChildCount := len(aep.ItemListForTest(c).Children)
	preNextItemID := aep.NextItemIDForTest(proj)

	clone, err := aep.DuplicateLayer(c, idx, "mt_alpha_clone")
	if err != nil {
		t.Fatalf("DuplicateLayer on explicit-matte layer: %v", err)
	}
	if clone == nil {
		t.Fatal("clone is nil with no error")
	}
	if len(c.Layers) != preLayerCount+1 {
		t.Errorf("layer count: got %d, want %d", len(c.Layers), preLayerCount+1)
	}
	if c.Layers[idx] != clone {
		t.Errorf("clone should occupy source's old idx %d; got different layer", idx)
	}
	if c.Layers[idx+1] != src {
		t.Errorf("source should be pushed to idx %d; got different layer", idx+1)
	}
	if clone.Name != "mt_alpha_clone" {
		t.Errorf("clone.Name: got %q, want %q", clone.Name, "mt_alpha_clone")
	}
	if clone.ID == src.ID {
		t.Errorf("clone.ID must differ from source; both = %d", clone.ID)
	}
	if clone.ID != preNextItemID {
		t.Errorf("clone.ID: got %d, want %d", clone.ID, preNextItemID)
	}
	if clone.TrackMatte != srcMode {
		t.Errorf("clone.TrackMatte: got %d, want %d (verbatim from source)", clone.TrackMatte, srcMode)
	}
	if clone.TrackMatteLayerID != srcMatteID {
		t.Errorf("clone.TrackMatteLayerID: got %d, want %d (verbatim from source)", clone.TrackMatteLayerID, srcMatteID)
	}
	// itemList grew by per-layer block size (16 for AE-saved layer).
	delta := len(aep.ItemListForTest(c).Children) - preChildCount
	if delta != 16 {
		t.Errorf("itemList children delta: got %d, want 16", delta)
	}
}

// Byte-verbatim: clone's ldta @0xA0..0xA3 (TrackMatteLayerID) and
// @0x6B (TrackMatte mode) match source byte-for-byte. Only @0x00..0x03
// (layer ID) differs per F10.
func TestDuplicateLayer_ExplicitMatte_VerbatimBytes(t *testing.T) {
	proj, c := openTrackMatteAE24(t)
	if proj == nil || c == nil {
		return
	}
	src := layerBySourceName(proj, c, "mt_luma_to_solidB")
	if src == nil {
		t.Fatal("fixture missing layer 'mt_luma_to_solidB'")
	}
	idx := layerIndexInComp(c, src)
	if idx < 0 {
		t.Fatal("can't locate mt_luma_to_solidB in c.Layers")
	}
	srcLdtaBytes := append([]byte(nil), aep.LdtaForTest(src).Data...)
	if len(srcLdtaBytes) < 0xA4 {
		t.Fatalf("source ldta too short for AE 23+ matte slot: %d bytes", len(srcLdtaBytes))
	}

	clone, err := aep.DuplicateLayer(c, idx, "mt_luma_clone")
	if err != nil {
		t.Fatalf("DuplicateLayer: %v", err)
	}
	cloneLdtaBytes := aep.LdtaForTest(clone).Data
	if len(cloneLdtaBytes) != len(srcLdtaBytes) {
		t.Fatalf("ldta length mismatch: clone=%d src=%d", len(cloneLdtaBytes), len(srcLdtaBytes))
	}
	// @0x00..0x03 — clone ID (must differ from source).
	gotCloneID := binary.BigEndian.Uint32(cloneLdtaBytes[0x00:0x04])
	if gotCloneID != clone.ID {
		t.Errorf("clone ldta @0x00..0x03: got %d, want %d", gotCloneID, clone.ID)
	}
	// @0x6B — TrackMatte mode byte (verbatim).
	if cloneLdtaBytes[0x6B] != srcLdtaBytes[0x6B] {
		t.Errorf("ldta @0x6B (TrackMatte): clone=0x%02X src=0x%02X — must be verbatim copy", cloneLdtaBytes[0x6B], srcLdtaBytes[0x6B])
	}
	// @0xA0..0xA3 — TrackMatteLayerID (verbatim).
	if !bytes.Equal(cloneLdtaBytes[0xA0:0xA4], srcLdtaBytes[0xA0:0xA4]) {
		t.Errorf("ldta @0xA0..0xA3 (TrackMatteLayerID): clone=% X src=% X — must be verbatim copy", cloneLdtaBytes[0xA0:0xA4], srcLdtaBytes[0xA0:0xA4])
	}
	// F10 strict: everything past the ID field must match byte-for-byte.
	if !bytes.Equal(srcLdtaBytes[4:], cloneLdtaBytes[4:]) {
		for i := 4; i < len(srcLdtaBytes); i++ {
			if srcLdtaBytes[i] != cloneLdtaBytes[i] {
				t.Fatalf("ldta @0x%02X: src=0x%02X clone=0x%02X — F10 says only @0x00..0x03 should differ", i, srcLdtaBytes[i], cloneLdtaBytes[i])
			}
		}
	}
}

// Round-trip: explicit matte survives WriteAEP + reparse.
func TestDuplicateLayer_ExplicitMatte_RoundTrip(t *testing.T) {
	proj, c := openTrackMatteAE24(t)
	if proj == nil || c == nil {
		return
	}
	src := layerBySourceName(proj, c, "mt_alphainv_to_solidC")
	if src == nil {
		t.Fatal("fixture missing layer 'mt_alphainv_to_solidC'")
	}
	idx := layerIndexInComp(c, src)
	if idx < 0 {
		t.Fatal("can't locate mt_alphainv_to_solidC in c.Layers")
	}
	srcMode := src.TrackMatte
	srcMatteID := src.TrackMatteLayerID

	clone, err := aep.DuplicateLayer(c, idx, "mt_alphainv_clone")
	if err != nil {
		t.Fatalf("DuplicateLayer: %v", err)
	}
	cloneID := clone.ID

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	var clone2 *aep.Layer
	for _, c2 := range proj2.Compositions {
		for _, l := range c2.Layers {
			if l.ID == cloneID {
				clone2 = l
				break
			}
		}
		if clone2 != nil {
			break
		}
	}
	if clone2 == nil {
		t.Fatalf("clone (ID=%d) not found after round-trip", cloneID)
	}
	if clone2.Name != "mt_alphainv_clone" {
		t.Errorf("post-roundtrip clone.Name: got %q, want %q", clone2.Name, "mt_alphainv_clone")
	}
	if clone2.TrackMatte != srcMode {
		t.Errorf("post-roundtrip clone.TrackMatte: got %d, want %d", clone2.TrackMatte, srcMode)
	}
	if clone2.TrackMatteLayerID != srcMatteID {
		t.Errorf("post-roundtrip clone.TrackMatteLayerID: got %d, want %d", clone2.TrackMatteLayerID, srcMatteID)
	}
}

// Sibling-matte: clone and source BOTH carry the same explicit
// matte pointer post-dup; AE should render two matted layers from the
// same source layer.
func TestDuplicateLayer_ExplicitMatte_BothPointToSameSource(t *testing.T) {
	proj, c := openTrackMatteAE24(t)
	if proj == nil || c == nil {
		return
	}
	src := layerBySourceName(proj, c, "mt_alpha_to_solidA")
	if src == nil {
		t.Fatal("fixture missing layer 'mt_alpha_to_solidA'")
	}
	idx := layerIndexInComp(c, src)
	if idx < 0 {
		t.Fatal("can't locate mt_alpha_to_solidA in c.Layers")
	}
	originalMatteID := src.TrackMatteLayerID
	if originalMatteID == 0 {
		t.Fatalf("test precondition: source must have explicit matte")
	}

	clone, err := aep.DuplicateLayer(c, idx, "mt_alpha_clone")
	if err != nil {
		t.Fatalf("DuplicateLayer: %v", err)
	}
	// Source unchanged.
	if src.TrackMatteLayerID != originalMatteID {
		t.Errorf("source matte ID changed after dup: was %d, now %d", originalMatteID, src.TrackMatteLayerID)
	}
	// Clone matches.
	if clone.TrackMatteLayerID != originalMatteID {
		t.Errorf("clone matte ID: got %d, want %d (same source as original)", clone.TrackMatteLayerID, originalMatteID)
	}
	// Both reference a real layer in the comp.
	if c.LayerByID(originalMatteID) == nil {
		t.Errorf("matte source layer (ID=%d) not found in comp — broken explicit matte ref", originalMatteID)
	}
}
