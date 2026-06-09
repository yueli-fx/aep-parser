package aep

import (
	"encoding/binary"
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// InsertLayer deep-clones src into c.Layers at atIdx (0-based; atIdx ==
// len(c.Layers) appends). Returns the inserted clone *Layer on success. src may
// live in a sibling comp of the same Project, or in a different Project
// (cross-Project).
//
// Same-Project clone semantics (scene.CompositionProj(scene.LayerComp(src)) == scene.CompositionProj(c)):
//
//   - new layer ID = allocItemID(scene.CompositionProj(c)) (head counter +1, monotonic)
//   - clone block = deep byte-clone of src's [Layr, Ewst, leaf-followers)
//     range, with per-byte ldta mutations:
//     @0x00..0x03 ← newID
//     @0x6B       ← TrackMatteNone (cross-comp matte source is invalid)
//     @0x84..0x87 ← 0 (ParentID; src's ParentID named a layer in scene.LayerComp(src))
//     @0xA0..0xA3 ← 0 (explicit matte ID, guarded by len(ldta) >= 0xA4)
//   - clone.SourceID = src.SourceID (verbatim — the shared Footage/Comp item).
//   - clone.Name = src.Name (verbatim — matches AE's layer.copyToComp).
//
// Cross-Project semantics (scene.CompositionProj(scene.LayerComp(src)) != scene.CompositionProj(c)):
// additionally imports src's reachable ITEM CLOSURE (footage + precomp,
// transitively) into c's Project at root level with fresh dest item IDs, then
// remaps the inserted clone's SourceID @0x28 + AlternateSourceID through the
// srcItemID→destItemID map. File-backed footage already present in dest (matched
// by Path) is reused, not re-cloned; comps and solids/placeholders are always
// cloned. ParentID / track matte are still reset (cross-comp). Folders are not
// recreated.
//
// Refuse-cases: nil src, dest backref missing, atIdx out of range, src
// detached, same-comp redirect, non-AV, direct pre-comp loop (same-Project
// only), src backref missing, structural corruption. Cross-Project adds:
// dest/src Project has no root Fold; dangling closure source.
//
// Atomic mutation: snapshot dest itemList.Children + c.Layers +
// scene.ProjectNextItemID(scene.CompositionProj(c)) + len(scene.CompositionProj(c).Warnings) (cross-Project also snapshots
// rootFold.Children + Compositions + Footage); on any new parser warning
// during re-parse, roll all back including the nextItemID bump.
//
// Stable (both paths) — same-Project passed AE 2020 + AE 2025 ship-gate (3 modes
// [basic/footage/precomp] × 2 = 6/6 PASS); cross-Project passed the assert-based
// AE 2020 + AE 2025 gate (3 modes [footage/precomp/dedup] × 2 = 6/6 PASS): AE
// accepts the Go-emitted file, the inserted clone's source resolves (imported /
// dedup'd), and footage is not duplicated on path match.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. BREAKING vs the former Composition.InsertLayer method form.
func InsertLayer(c *Composition, src *Layer, atIdx int) (*Layer, error) {
	// === Refuse-case matrix R1-R11 ===
	if src == nil {
		return nil, fmt.Errorf("InsertLayer: src cannot be nil")
	}
	destCb := compositionBack(c)
	if destCb == nil || destCb.itemList == nil {
		return nil, fmt.Errorf("InsertLayer: dest comp %q has no itemList back-ref (built outside parser?)", c.Name)
	}
	if scene.CompositionProj(c) == nil {
		return nil, fmt.Errorf("InsertLayer: dest comp %q has no project back-ref", c.Name)
	}
	if atIdx < 0 || atIdx > len(c.Layers) {
		return nil, fmt.Errorf("InsertLayer: atIdx %d out of range (have %d layers; %d is append)", atIdx, len(c.Layers), len(c.Layers))
	}
	if scene.LayerComp(src) == nil {
		return nil, fmt.Errorf("InsertLayer: src.comp is nil (layer detached from any comp)")
	}
	if scene.LayerComp(src) == c {
		return nil, fmt.Errorf("InsertLayer: src and dest are the same comp %q — use DuplicateLayer instead", c.Name)
	}
	crossProject := scene.CompositionProj(scene.LayerComp(src)) != scene.CompositionProj(c)
	if src.Type != LayerTypeAV {
		return nil, fmt.Errorf("InsertLayer: refuse non-AV src (Type=%s); only AV layers supported", src.Type)
	}
	if !crossProject && src.SourceID != 0 && src.SourceID == c.ID {
		return nil, fmt.Errorf("InsertLayer: refuse direct pre-comp loop (src.SourceID=%d == dest.ID=%d)", src.SourceID, c.ID)
	}
	srcBack := layerBack(src)
	if srcBack == nil || srcBack.layrList == nil {
		return nil, fmt.Errorf("InsertLayer: src layer %q has no Layr chunk back-ref", src.Name)
	}
	srcCb, ok2 := scene.CompositionBack(scene.LayerComp(src)).(*compositionBackrefs)
	if !ok2 || srcCb == nil || srcCb.itemList == nil {
		return nil, fmt.Errorf("InsertLayer: src comp %q has no itemList back-ref", scene.LayerComp(src).Name)
	}
	srcChildren := srcCb.itemList.Children
	srcLayrIdx := findLayrIndexInItemList(srcCb.itemList, srcBack.layrList)
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
// srcChildren — scene.LayerComp(src)'s itemList) into c at atIdx, applying the standard
// cross-comp ldta mutations (new layer ID, ParentID/matte reset) plus
// sourceRemap to SourceID @0x28 and AlternateSourceID (blsi). sourceRemap is
// identity for same-Project inserts (bytes unchanged) and an itemIDMap lookup
// for cross-Project inserts. Atomic over c.itemList / c.Layers / scene.ProjectNextItemID(proj)
// / proj.Warnings.
func spliceLayerClone(c *Composition, atIdx, srcLayrIdx int, srcChildren []*rifx.Chunk, sourceRemap func(uint32) uint32) (*Layer, error) {
	// === Adaptive block end — scan leaf followers until next LIST/EOF ===
	endIdx := srcLayrIdx + 2
	for endIdx < len(srcChildren) && !srcChildren[endIdx].IsList() {
		endIdx++
	}

	// === Snapshot for rollback ===
	cb := compositionBack(c)
	if cb == nil || cb.itemList == nil {
		return nil, fmt.Errorf("InsertLayer: dest comp %q has no itemList back-ref", c.Name)
	}
	oldDestChildren := append([]*rifx.Chunk(nil), cb.itemList.Children...)
	oldDestLayers := append([]*Layer(nil), c.Layers...)
	oldNextItemID := scene.ProjectNextItemID(scene.CompositionProj(c))
	oldWarningsLen := len(scene.CompositionProj(c).Warnings)

	// === Deep-clone source block (fresh Data slices) ===
	cloneBlock := make([]*rifx.Chunk, endIdx-srcLayrIdx)
	for k := srcLayrIdx; k < endIdx; k++ {
		cloneBlock[k-srcLayrIdx] = deepCloneChunk(srcChildren[k])
	}

	// === Allocate new ID + per-byte ldta mutations ===
	newID := allocItemID(scene.CompositionProj(c))
	clonedLayr := cloneBlock[0]
	clonedLdta := clonedLayr.FindFirst(rifx.IDLdta)
	if clonedLdta == nil {
		scene.SetProjectNextItemID(scene.CompositionProj(c), oldNextItemID)
		return nil, fmt.Errorf("InsertLayer: cloned Layr has no ldta chunk")
	}
	if len(clonedLdta.Data) < 0x88 {
		scene.SetProjectNextItemID(scene.CompositionProj(c), oldNextItemID)
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
	destChildren := cb.itemList.Children
	var insertChunkIdx int
	switch {
	case len(c.Layers) == 0:
		insertChunkIdx = insertLayrPosition(destChildren)
	case atIdx < len(c.Layers):
		target := c.Layers[atIdx]
		targetBack := layerBack(target)
		if targetBack == nil || targetBack.layrList == nil {
			scene.SetProjectNextItemID(scene.CompositionProj(c), oldNextItemID)
			return nil, fmt.Errorf("InsertLayer: dest Layers[%d] %q has no Layr backref", atIdx, target.Name)
		}
		insertChunkIdx = indexOfChunk(destChildren, targetBack.layrList)
		if insertChunkIdx < 0 {
			scene.SetProjectNextItemID(scene.CompositionProj(c), oldNextItemID)
			return nil, fmt.Errorf("InsertLayer: dest Layers[%d] %q Layr chunk not found in dest itemList", atIdx, target.Name)
		}
	default:
		last := c.Layers[len(c.Layers)-1]
		lastBack := layerBack(last)
		if lastBack == nil || lastBack.layrList == nil {
			scene.SetProjectNextItemID(scene.CompositionProj(c), oldNextItemID)
			return nil, fmt.Errorf("InsertLayer: last dest layer %q has no Layr backref", last.Name)
		}
		lastLayrIdx := indexOfChunk(destChildren, lastBack.layrList)
		if lastLayrIdx < 0 {
			scene.SetProjectNextItemID(scene.CompositionProj(c), oldNextItemID)
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
	cb.itemList.Children = newDestChildren

	// === Re-parse cloned Layr → fresh *Layer ===
	var localWarnings []string
	ctx := newParseCtxFPS(c.TickRate, c.FrameRate, c.Name, &localWarnings)
	cloneLayer, parseErr := parseLayer(clonedLayr, atIdx, ctx)
	if parseErr != nil {
		cb.itemList.Children = oldDestChildren
		scene.SetProjectNextItemID(scene.CompositionProj(c), oldNextItemID)
		return nil, fmt.Errorf("InsertLayer: re-parse cloned layer: %w", parseErr)
	}
	scene.SetLayerComp(cloneLayer, c)
	scene.AssignTransformDefaults(cloneLayer.Properties, c, cloneLayer.Type)

	// === Insert cloneLayer into c.Layers ===
	newLayers := make([]*Layer, 0, len(c.Layers)+1)
	newLayers = append(newLayers, c.Layers[:atIdx]...)
	newLayers = append(newLayers, cloneLayer)
	newLayers = append(newLayers, c.Layers[atIdx:]...)
	c.Layers = newLayers

	// === Warnings-as-failure rollback ===
	if len(localWarnings) > 0 {
		scene.CompositionProj(c).Warnings = append(scene.CompositionProj(c).Warnings, localWarnings...)
	}
	if len(scene.CompositionProj(c).Warnings) > oldWarningsLen {
		cb.itemList.Children = oldDestChildren
		c.Layers = oldDestLayers
		scene.SetProjectNextItemID(scene.CompositionProj(c), oldNextItemID)
		newWarnings := append([]string(nil), scene.CompositionProj(c).Warnings[oldWarningsLen:]...)
		scene.CompositionProj(c).Warnings = scene.CompositionProj(c).Warnings[:oldWarningsLen]
		return nil, fmt.Errorf("InsertLayer: produced %d parser warning(s), rolled back: %v", len(newWarnings), newWarnings)
	}

	return cloneLayer, nil
}
