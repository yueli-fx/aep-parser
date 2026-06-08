package aep

import (
	"encoding/binary"
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// DuplicateLayer clones the layer at the given 0-based index in c.Layers
// and inserts the clone at that same position, pushing source and
// everything below down by one (mirrors AE ScriptingAPI's
// layer.duplicate()). Returns the cloned *Layer on success, or an error
// if a refuse-case triggers.
//
// Clone semantics (RE'd via 4 AE-saved fixtures + byte-diff):
//
//   - new layer ID = proj.allocItemID() (head counter +1, monotonic)
//   - clone's 16-chunk block (Layr + Ewst + 14 follower leaves in
//     AE-saved files; 2 chunks in Go-built layers) is a deep byte-clone
//     of source's block, with ldta @0x00..0x03 overwritten with the new
//     ID. All other body bytes (SourceID @0x28, ParentID @0x84,
//     TrackMatte @0x6B) are verbatim from source.
//   - Layer.SourceID/ParentID/TrackMatteLayerID/TrackMatte struct fields
//     on the clone = source values (no footage duplication; no
//     reference rewrites).
//   - Name = caller-supplied (AE keeps source's name verbatim; we
//     require an explicit name to avoid silent duplicate-name confusion).
//   - Children's outgoing ParentID is NOT updated — clone is a fresh
//     sibling shadow; source remains the canonical parent for any
//     incoming refs (F6).
//
// Refuse-cases (conservative):
//
//   - name empty
//   - index out of range
//   - comp lacks parsed itemList back-ref
//   - source is not an AV layer (camera/light/audio behavior not RE'd)
//   - source has implicit TrackMatte (TrackMatte != None &&
//     TrackMatteLayerID == 0). F2 quirk: AE relocates clone above the
//     positional matte source to preserve original's matte; not yet
//     supported. AE 23+ explicit matte (TrackMatteLayerID != 0) is
//     ALLOWED (Stable — clone byte-copies @0xA0 + @0x6B verbatim;
//     passed AE 2025 ship-gate).
//   - backref corruption (Layr formType / Ewst sibling mismatch)
//
// Atomic mutation: snapshot pre-call state of itemList.Children,
// c.Layers, proj.nextItemID, and proj.Warnings; on any parser warning
// surfaced during the re-parse, roll all of them back (including the
// nextItemID bump) and return the warnings as an error.
func (c *Composition) DuplicateLayer(index int, name string) (*Layer, error) {
	// 1. Validate refuse-cases.
	if name == "" {
		return nil, fmt.Errorf("DuplicateLayer: name cannot be empty")
	}
	if index < 0 || index >= len(c.Layers) {
		return nil, fmt.Errorf("DuplicateLayer: index %d out of range (have %d layers)", index, len(c.Layers))
	}
	cb, ok := c.back.(*compositionBackrefs)
	if !ok || cb == nil || cb.itemList == nil {
		return nil, fmt.Errorf("DuplicateLayer: comp %q has no itemList back-ref (built outside parser?)", c.Name)
	}
	if c.proj == nil {
		return nil, fmt.Errorf("DuplicateLayer: comp %q has no project back-ref", c.Name)
	}

	source := c.Layers[index]
	if source.Type != LayerTypeAV {
		return nil, fmt.Errorf("DuplicateLayer: refuse non-AV layer (idx=%d Type=%s); only AV layers supported", index, source.Type)
	}
	// F2 quirk applies only to implicit "layer-above" matte where matte
	// source is positional. AE 23+ explicit matte (TrackMatteLayerID !=
	// 0) decouples matte from layer order — clone keeps the explicit ID
	// via byte-verbatim ldta @0xA0 + @0x6B copy.
	if source.TrackMatte != TrackMatteNone && source.TrackMatteLayerID == 0 {
		return nil, fmt.Errorf("DuplicateLayer: refuse layer %q (idx=%d) with implicit TrackMatte=%d (TrackMatteLayerID=0); AE relocates clone to preserve original's matte (F2 quirk), not yet supported", source.Name, index, source.TrackMatte)
	}
	sourceBack := source.layerBack()
	if sourceBack == nil || sourceBack.layrList == nil {
		return nil, fmt.Errorf("DuplicateLayer: layer %q at idx %d has no Layr chunk back-ref", source.Name, index)
	}

	// 2. Locate source Layr in itemList.Children.
	children := cb.itemList.Children
	srcLayrIdx := findLayrIndexInItemList(cb.itemList, sourceBack.layrList)
	if srcLayrIdx < 0 {
		return nil, fmt.Errorf("DuplicateLayer: layer %q Layr chunk not found in itemList", source.Name)
	}

	// 3. Defensive structural assertions — FormType and Ewst sibling.
	if !children[srcLayrIdx].IsList() || children[srcLayrIdx].FormType != rifx.IDLayr {
		return nil, fmt.Errorf("DuplicateLayer: layer %q backref points to non-Layr chunk (FormType=%s)", source.Name, chunkIDString(children[srcLayrIdx].FormType))
	}
	if srcLayrIdx+1 >= len(children) {
		return nil, fmt.Errorf("DuplicateLayer: layer %q Layr at end of itemList (no Ewst sibling)", source.Name)
	}
	ewstCandidate := children[srcLayrIdx+1]
	if !ewstCandidate.IsList() || ewstCandidate.FormType != rifx.IDEwst {
		return nil, fmt.Errorf("DuplicateLayer: layer %q expected Ewst sibling after Layr, found %s", source.Name, chunkIDString(ewstCandidate.FormType))
	}

	// 4. Adaptive block end — consume leaf followers until next LIST/EOF
	//    (handles AE-saved 16-chunk and Go-built 2-chunk forms alike).
	endIdx := srcLayrIdx + 2
	for endIdx < len(children) && !children[endIdx].IsList() {
		endIdx++
	}

	// 5. Snapshot for rollback. Key diff vs DeleteLayer: we DO snapshot
	//    proj.nextItemID (the clone bumps it; rollback must un-bump so a
	//    subsequent New/Duplicate gets the right ID).
	oldItemChildren := append([]*rifx.Chunk(nil), children...)
	oldLayers := append([]*Layer(nil), c.Layers...)
	oldNextItemID := c.proj.nextItemID
	oldWarningsLen := len(c.proj.Warnings)

	// 6. Deep-clone source's [srcLayrIdx, endIdx) block. Every Data slice
	//    is freshly allocated — Set* mutates share chunk bytes, so a
	//    shared slice would let a later edit corrupt the source layer.
	cloneBlock := make([]*rifx.Chunk, endIdx-srcLayrIdx)
	for k := srcLayrIdx; k < endIdx; k++ {
		cloneBlock[k-srcLayrIdx] = deepCloneChunk(children[k])
	}

	// 7. Allocate new ID, mutate clone's ldta @0x00..0x03 — the ONLY byte
	//    change to the cloned block per Finding 10.
	newID := c.proj.allocItemID()
	clonedLayr := cloneBlock[0]
	clonedLdta := clonedLayr.FindFirst(rifx.IDLdta)
	if clonedLdta == nil || len(clonedLdta.Data) < 4 {
		c.proj.nextItemID = oldNextItemID
		return nil, fmt.Errorf("DuplicateLayer: cloned Layr missing ldta or ldta data too short (got %d bytes)", len(clonedLdta.Data))
	}
	binary.BigEndian.PutUint32(clonedLdta.Data[0x00:0x04], newID)

	// 8. Rewrite clone's name Utf8 chunk (length-variable; same mechanic
	//    as Layer.SetName — replace Data slice, WriteAEP recomputes
	//    ancestor LIST sizes).
	clonedNameUtf8 := clonedLayr.FindFirst(rifx.IDUtf8)
	if clonedNameUtf8 == nil {
		c.proj.nextItemID = oldNextItemID
		return nil, fmt.Errorf("DuplicateLayer: cloned Layr %q missing Utf8 name chunk", source.Name)
	}
	clonedNameUtf8.Data = []byte(name)

	// 9. Splice clone block into itemList.Children at srcLayrIdx (BEFORE
	//    source, pushing source down — matches F1 for solo/dup_parent/
	//    dup_child modes).
	newChildren := make([]*rifx.Chunk, 0, len(children)+len(cloneBlock))
	newChildren = append(newChildren, children[:srcLayrIdx]...)
	newChildren = append(newChildren, cloneBlock...)
	newChildren = append(newChildren, children[srcLayrIdx:]...)
	cb.itemList.Children = newChildren

	// 10. Re-parse cloneLayr to build a fresh *Layer with backrefs into
	//     cloned chunks. parseLayer reads ID from cloned ldta @0x00 (now
	//     newID), name from cloned Utf8 (now caller-supplied), and all
	//     other fields verbatim from cloned bytes.
	var localWarnings []string
	ctx := newParseCtxFPS(c.TickRate, c.FrameRate, c.Name, &localWarnings)
	cloneLayer, parseErr := parseLayer(clonedLayr, index, ctx)
	if parseErr != nil {
		cb.itemList.Children = oldItemChildren
		c.proj.nextItemID = oldNextItemID
		return nil, fmt.Errorf("DuplicateLayer: re-parse cloned layer: %w", parseErr)
	}
	cloneLayer.comp = c
	assignTransformDefaults(cloneLayer.Properties, c, cloneLayer.Type)

	// 11. Insert cloneLayer into c.Layers at index.
	newLayers := make([]*Layer, 0, len(c.Layers)+1)
	newLayers = append(newLayers, c.Layers[:index]...)
	newLayers = append(newLayers, cloneLayer)
	newLayers = append(newLayers, c.Layers[index:]...)
	c.Layers = newLayers

	// 12. Warnings-as-failure. Append local re-parse warnings to project,
	//     then rollback ALL state if any new warnings appeared.
	if len(localWarnings) > 0 {
		c.proj.Warnings = append(c.proj.Warnings, localWarnings...)
	}
	if len(c.proj.Warnings) > oldWarningsLen {
		cb.itemList.Children = oldItemChildren
		c.Layers = oldLayers
		c.proj.nextItemID = oldNextItemID
		newWarnings := append([]string(nil), c.proj.Warnings[oldWarningsLen:]...)
		c.proj.Warnings = c.proj.Warnings[:oldWarningsLen]
		return nil, fmt.Errorf("DuplicateLayer: produced %d parser warning(s), rolled back: %v", len(newWarnings), newWarnings)
	}

	return cloneLayer, nil
}

// deepCloneChunk is defined in new_composition.go — recursive deep copy
// with fresh Data slices (concurrent-mutate safe; Set* mutates share
// chunk bytes, so clones must not alias the source's slices).
