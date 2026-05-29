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
//     @0x00..0x03 ← newID
//     @0x6B       ← TrackMatteNone (cross-comp matte source is invalid)
//     @0x84..0x87 ← 0 (ParentID; src's ParentID named a layer in src.comp)
//     @0xA0..0xA3 ← 0 (explicit matte ID, guarded by len(ldta) >= 0xA4)
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
// Stable — passed AE 2020 + AE 2025 ship-gate (3 modes [basic/footage/precomp]
// × 2 versions = 6/6 PASS, 2026-05-29): AE accepts the Go-emitted file and the
// clone references the source item verbatim with parent + track matte reset.
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
	if clonedLdta == nil {
		c.proj.nextItemID = oldNextItemID
		return nil, fmt.Errorf("InsertLayer: cloned Layr has no ldta chunk")
	}
	if len(clonedLdta.Data) < 0x88 {
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
