package serializer

import (
	"encoding/binary"
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// DeleteLayer removes the layer at the given 0-based index in c.Layers.
// Returns nil on success, or an error if a refuse-case triggers (index
// out of range / comp lacks itemList back-ref / target is the last
// layer / target is not an AV layer / backref corruption).
// (Full contract + RE notes live on the aep.DeleteLayer facade — docgen source.)
func DeleteLayer(c *Composition, index int) error {
	// 1. Validate refuse-cases.
	if index < 0 || index >= len(c.Layers) {
		return fmt.Errorf("DeleteLayer: index %d out of range (have %d layers)", index, len(c.Layers))
	}
	cb := compositionBack(c)
	if cb == nil || cb.itemList == nil {
		return fmt.Errorf("DeleteLayer: comp %q has no itemList back-ref (built outside parser?)", c.Name)
	}
	if len(c.Layers) == 1 {
		return fmt.Errorf("DeleteLayer: refuse to delete the last layer of comp %q (single-layer-comp behavior not RE'd)", c.Name)
	}

	deleted := c.Layers[index]
	if deleted.Type != LayerTypeAV {
		return fmt.Errorf("DeleteLayer: refuse non-AV layer (idx=%d Type=%s); only AV layers supported", index, deleted.Type)
	}
	deletedBack := layerBack(deleted)
	if deletedBack == nil || deletedBack.layrList == nil {
		return fmt.Errorf("DeleteLayer: layer %q at idx %d has no Layr chunk back-ref", deleted.Name, index)
	}

	// 2. Locate Layr in itemList.Children.
	children := cb.itemList.Children
	layrIdx := findLayrIndexInItemList(cb.itemList, deletedBack.layrList)
	if layrIdx < 0 {
		return fmt.Errorf("DeleteLayer: layer %q Layr chunk not found in itemList", deleted.Name)
	}

	// 3. Defensive structural assertions — FormType and Ewst sibling.
	if !children[layrIdx].IsList() || children[layrIdx].FormType != rifx.IDLayr {
		return fmt.Errorf("DeleteLayer: layer %q backref points to non-Layr chunk (FormType=%s)", deleted.Name, chunkIDString(children[layrIdx].FormType))
	}
	if layrIdx+1 >= len(children) {
		return fmt.Errorf("DeleteLayer: layer %q Layr at end of itemList (no Ewst sibling)", deleted.Name)
	}
	ewstCandidate := children[layrIdx+1]
	if !ewstCandidate.IsList() || ewstCandidate.FormType != rifx.IDEwst {
		return fmt.Errorf("DeleteLayer: layer %q expected Ewst sibling after Layr, found %s", deleted.Name, chunkIDString(ewstCandidate.FormType))
	}

	// 4. Adaptive splice — find end of trailing leaf block. Consumes the
	//    14 fvdv/fiop/ftts/foac/fiac/fipc/fifl follower leaves in AE-saved
	//    files; consumes 0 in Go-built layers (NewShapeLayer inserts only
	//    Layr+Ewst).
	endIdx := layrIdx + 2
	for endIdx < len(children) && !children[endIdx].IsList() {
		endIdx++
	}

	// 5. Snapshot for rollback.
	oldChildren := append([]*rifx.Chunk(nil), children...)
	oldLayers := append([]*Layer(nil), c.Layers...)
	oldWarningsLen := 0
	if scene.CompositionProj(c) != nil {
		oldWarningsLen = len(scene.CompositionProj(c).Warnings)
	}

	// Snapshot every neighbor whose ref we'll clear: struct fields +
	// ldta bytes (when present). Track distinct ldta chunks so the
	// rollback restore happens once per chunk even if multiple fields
	// were cleared on the same neighbor.
	type neighborSnap struct {
		layer             *Layer
		parentID          uint32
		trackMatteLayerID uint32
	}
	type ldtaSnap struct {
		chunk *rifx.Chunk
		data  []byte
	}
	var neighborSnaps []neighborSnap
	var ldtaSnaps []ldtaSnap
	snapLdta := func(chunk *rifx.Chunk) {
		if chunk == nil {
			return
		}
		for _, s := range ldtaSnaps {
			if s.chunk == chunk {
				return
			}
		}
		ldtaSnaps = append(ldtaSnaps, ldtaSnap{
			chunk: chunk,
			data:  append([]byte(nil), chunk.Data...),
		})
	}

	deletedID := deleted.ID

	// 6. Reference cleanup pass.
	for k, neighbor := range c.Layers {
		if k == index {
			continue
		}
		needsParent := neighbor.ParentID == deletedID
		needsMatte := neighbor.TrackMatteLayerID == deletedID
		if !needsParent && !needsMatte {
			continue
		}
		neighborSnaps = append(neighborSnaps, neighborSnap{
			layer:             neighbor,
			parentID:          neighbor.ParentID,
			trackMatteLayerID: neighbor.TrackMatteLayerID,
		})
		neighborBack := layerBack(neighbor)
		if needsParent {
			if neighborBack != nil && neighborBack.ldta != nil &&
				len(neighborBack.ldta.Data) >= 0x88 {
				snapLdta(neighborBack.ldta)
				binary.BigEndian.PutUint32(neighborBack.ldta.Data[0x84:0x88], 0)
			}
			neighbor.ParentID = 0
		}
		if needsMatte {
			if neighborBack != nil && neighborBack.ldta != nil &&
				len(neighborBack.ldta.Data) >= 0xA4 {
				snapLdta(neighborBack.ldta)
				binary.BigEndian.PutUint32(neighborBack.ldta.Data[0xA0:0xA4], 0)
			}
			neighbor.TrackMatteLayerID = 0
		}
	}

	// 7. Splice itemList.Children — drop [layrIdx, endIdx).
	cb.itemList.Children = append(append([]*rifx.Chunk(nil), children[:layrIdx]...), children[endIdx:]...)

	// 8. Splice c.Layers.
	c.Layers = append(append([]*Layer(nil), c.Layers[:index]...), c.Layers[index+1:]...)

	// 9. Warnings-as-failure. DeleteLayer doesn't re-parse so warnings
	//    won't increase in practice — this is a defensive rollback path
	//    matching the V2.1 pattern, so future re-parse extensions get it
	//    for free.
	if scene.CompositionProj(c) != nil && len(scene.CompositionProj(c).Warnings) > oldWarningsLen {
		cb.itemList.Children = oldChildren
		c.Layers = oldLayers
		for _, s := range neighborSnaps {
			s.layer.ParentID = s.parentID
			s.layer.TrackMatteLayerID = s.trackMatteLayerID
		}
		for _, s := range ldtaSnaps {
			s.chunk.Data = s.data
		}
		newWarnings := append([]string(nil), scene.CompositionProj(c).Warnings[oldWarningsLen:]...)
		scene.CompositionProj(c).Warnings = scene.CompositionProj(c).Warnings[:oldWarningsLen]
		return fmt.Errorf("DeleteLayer: produced %d parser warning(s), rolled back: %v", len(newWarnings), newWarnings)
	}

	return nil
}

// findLayrIndexInItemList returns the index of layrChunk in
// itemList.Children, or -1 if not present. Pointer identity match.
func findLayrIndexInItemList(itemList, layrChunk *rifx.Chunk) int {
	if itemList == nil || layrChunk == nil {
		return -1
	}
	for i, ch := range itemList.Children {
		if ch == layrChunk {
			return i
		}
	}
	return -1
}

// chunkIDString renders a ChunkID as a 4-char string for error
// messages. Local to delete_layer to avoid polluting rifx with a
// debugging helper.
func chunkIDString(id rifx.ChunkID) string {
	return string(id[:])
}
