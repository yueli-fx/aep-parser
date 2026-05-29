package aep

import (
	"encoding/binary"
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// InsertLayer deep-clones src into c.Layers at atIdx (0-based; atIdx ==
// len(c.Layers) appends). Returns the inserted clone *Layer on success. src may
// live in a sibling comp of the same Project (Phase 5C) or in a different
// Project (Phase 5C.1, cross-Project).
//
// Same-Project clone semantics (src.comp.proj == c.proj — see
// flightdeck/specs/2026-05-29-v3-phase5c-insertlayer-design.md):
//
//   - new layer ID = c.proj.allocItemID() (head counter +1, monotonic)
//   - clone block = deep byte-clone of src's [Layr, Ewst, leaf-followers)
//     range, with per-byte ldta mutations:
//     @0x00..0x03 ← newID
//     @0x6B       ← TrackMatteNone (cross-comp matte source is invalid)
//     @0x84..0x87 ← 0 (ParentID; src's ParentID named a layer in src.comp)
//     @0xA0..0xA3 ← 0 (explicit matte ID, guarded by len(ldta) >= 0xA4)
//   - clone.SourceID = src.SourceID (verbatim — the shared Footage/Comp item).
//   - clone.Name = src.Name (verbatim — matches AE's layer.copyToComp).
//
// Cross-Project semantics (src.comp.proj != c.proj — Phase 5C.1, see
// flightdeck/specs/2026-05-29-v3-phase5c1-cross-project-insertlayer-design.md):
// additionally imports src's reachable ITEM CLOSURE (footage + precomp,
// transitively) into c's Project at root level with fresh dest item IDs, then
// remaps the inserted clone's SourceID @0x28 + AlternateSourceID through the
// srcItemID→destItemID map. File-backed footage already present in dest (matched
// by Path) is reused, not re-cloned; comps and solids/placeholders are always
// cloned. ParentID / track matte are still reset (cross-comp). Folders are not
// recreated. ALPHA — pending AE 2020 + AE 2025 ship-gate.
//
// Refuse-cases (R1..R11; spec §2): nil src, dest backref missing, atIdx out of
// range, src detached, same-comp redirect, non-AV, direct pre-comp loop
// (same-Project only), src backref missing, structural corruption. Cross-Project
// adds: dest/src Project has no root Fold; dangling closure source.
//
// Atomic mutation (Inv-10 / Inv-11): snapshot dest itemList.Children +
// c.Layers + c.proj.nextItemID + len(c.proj.Warnings) (cross-Project also
// snapshots rootFold.Children + Compositions + Footage); on any new
// parser warning during re-parse, roll all back including the
// nextItemID bump.
//
// Same-Project: Stable — passed AE 2020 + AE 2025 ship-gate (3 modes
// [basic/footage/precomp] × 2 versions = 6/6 PASS, 2026-05-29): AE accepts the
// Go-emitted file and the clone references the source item verbatim with parent
// + track matte reset.
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
	crossProject := src.comp.proj != c.proj
	if src.Type != LayerTypeAV {
		return nil, fmt.Errorf("InsertLayer: refuse non-AV src (Type=%s); only AV layers supported in Phase 5C", src.Type)
	}
	if !crossProject && src.SourceID != 0 && src.SourceID == c.ID {
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

	if crossProject {
		return insertLayerCrossProject(c, src, atIdx, srcLayrIdx, srcChildren)
	}
	return spliceLayerClone(c, atIdx, srcLayrIdx, srcChildren, func(id uint32) uint32 { return id })
}

// spliceLayerClone deep-clones the source Layr block at srcLayrIdx (within
// srcChildren — src.comp's itemList) into c at atIdx, applying the standard
// cross-comp ldta mutations (new layer ID, ParentID/matte reset) plus
// sourceRemap to SourceID @0x28 and AlternateSourceID (blsi). sourceRemap is
// identity for same-Project inserts (bytes unchanged) and an itemIDMap lookup
// for cross-Project inserts. Atomic over c.itemList / c.Layers / proj.nextItemID
// / proj.Warnings.
func spliceLayerClone(c *Composition, atIdx, srcLayrIdx int, srcChildren []*rifx.Chunk, sourceRemap func(uint32) uint32) (*Layer, error) {
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
	// SourceID remap (identity for same-Project — byte-preserving).
	srcSourceID := binary.BigEndian.Uint32(clonedLdta.Data[0x28:0x2C])
	binary.BigEndian.PutUint32(clonedLdta.Data[0x28:0x2C], sourceRemap(srcSourceID))
	// AlternateSourceID remap (Media Replacement override; blsi @0x00).
	if blsi := findAlternateSourceBlsi(clonedLayr); blsi != nil && len(blsi.Data) >= 4 {
		altID := binary.BigEndian.Uint32(blsi.Data[0:4])
		binary.BigEndian.PutUint32(blsi.Data[0:4], sourceRemap(altID))
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

	// === Re-parse cloned Layr → fresh *Layer ===
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
