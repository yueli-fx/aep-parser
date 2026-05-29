package aep

import (
	"encoding/binary"
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// isFileBacked reports whether f is file footage eligible for path-based dedup
// (not solid/placeholder, with a non-empty Path).
func isFileBacked(f *Footage) bool {
	return f != nil && !f.IsSolid && !f.IsPlaceholder && f.Path != ""
}

// destFootageByPath returns the first file-backed footage in p whose Path
// equals path, or nil. Used for cross-Project footage dedup (5C.1).
func destFootageByPath(p *Project, path string) *Footage {
	if path == "" {
		return nil
	}
	for _, f := range p.Footage {
		if isFileBacked(f) && f.Path == path {
			return f
		}
	}
	return nil
}

// locateItemBlockByID searches container.Children (recursing into folder Sfdr
// sub-containers) for the Item LIST whose idta item-ID (@idtaItemID) equals id.
// On match it returns the CONTAINER holding the Item plus [start, end) covering
// the Item LIST and its trailing non-Item sibling run within that container.
// Returns (nil, -1, -1) if not found. Recursion handles items nested in project
// folders (e.g. solids in the "Solids" folder), which parseProject also walks
// recursively; imports flatten such items to the dest root regardless.
func locateItemBlockByID(container *rifx.Chunk, id uint32) (*rifx.Chunk, int, int) {
	children := container.Children
	for i, ch := range children {
		if !isItemList(ch) {
			continue
		}
		idta := ch.FindFirst(rifx.IDIdta)
		if idta != nil && len(idta.Data) >= idtaItemID+4 &&
			binary.BigEndian.Uint32(idta.Data[idtaItemID:idtaItemID+4]) == id {
			end := i + 1
			for end < len(children) && !isItemList(children[end]) {
				end++
			}
			return container, i, end
		}
		// Folder items hold their member items under an Sfdr LIST — recurse.
		if sfdr := ch.FindFirstList(rifx.IDSfdr); sfdr != nil {
			if c, s, e := locateItemBlockByID(sfdr, id); c != nil {
				return c, s, e
			}
		}
	}
	return nil, -1, -1
}

// importFootageBlock deep-clones srcID's footage Item block from src's root
// Fold into dest's root Fold with a fresh dest item ID (idta @idtaItemID),
// parses it into dest.Footage, and returns the new dest item ID. On error the
// CALLER (insertLayerCrossProject) restores dest via its outer snapshot — this
// helper does not self-rollback.
func importFootageBlock(dest, src *Project, srcID uint32, name string) (uint32, error) {
	container, start, end := locateItemBlockByID(src.back.rootFold, srcID)
	if container == nil {
		return 0, fmt.Errorf("footage Item block id=%d not found in src project", srcID)
	}
	dup := deepCloneChunk(container.Children[start])
	destID := dest.allocItemID()
	idta := dup.FindFirst(rifx.IDIdta)
	if idta == nil || len(idta.Data) < idtaItemID+4 {
		return 0, fmt.Errorf("cloned footage id=%d idta missing/short", srcID)
	}
	binary.BigEndian.PutUint32(idta.Data[idtaItemID:idtaItemID+4], destID)
	destRoot := dest.back.rootFold
	destRoot.Children = append(destRoot.Children, dup)
	for k := start + 1; k < end; k++ {
		destRoot.Children = append(destRoot.Children, deepCloneChunk(container.Children[k]))
	}
	f, err := parseFootage(dup, destID, name)
	if err != nil {
		return 0, fmt.Errorf("re-parse cloned footage id=%d: %w", srcID, err)
	}
	dest.Footage = append(dest.Footage, f)
	return destID, nil
}

// remapClonedCompLayerLayrs walks dupItemList's Layr LIST children, allocates a
// fresh dest layer ID per layer (rewriting ldta @0x00 and remapping intra-comp
// ParentID @0x84 / explicit matte @0xA0 through the local srcLayerID→destLayerID
// map), and returns the Layr LIST chunks (for the later cross-comp source-ref
// remap pass). Mirrors DuplicateComposition's two-pass pattern locally (the 5D
// file is deliberately left untouched).
func remapClonedCompLayerLayrs(p *Project, dupItemList *rifx.Chunk) ([]*rifx.Chunk, error) {
	idMap := make(map[uint32]uint32)
	var layrs []*rifx.Chunk
	for _, ch := range dupItemList.Children {
		if !ch.IsList() || ch.FormType != rifx.IDLayr {
			continue
		}
		ldta := ch.FindFirst(rifx.IDLdta)
		if ldta == nil {
			continue
		}
		if len(ldta.Data) < 0x88 {
			return nil, fmt.Errorf("cloned Layr ldta too short for ParentID write (got %d bytes, need >=0x88)", len(ldta.Data))
		}
		oldID := binary.BigEndian.Uint32(ldta.Data[0x00:0x04])
		newID := p.allocItemID()
		idMap[oldID] = newID
		binary.BigEndian.PutUint32(ldta.Data[0x00:0x04], newID)
		layrs = append(layrs, ch)
	}
	for _, ch := range layrs {
		ldta := ch.FindFirst(rifx.IDLdta)
		if parent := binary.BigEndian.Uint32(ldta.Data[0x84:0x88]); parent != 0 {
			if mapped, ok := idMap[parent]; ok {
				binary.BigEndian.PutUint32(ldta.Data[0x84:0x88], mapped)
			}
		}
		if len(ldta.Data) >= 0xA4 {
			if matte := binary.BigEndian.Uint32(ldta.Data[0xA0:0xA4]); matte != 0 {
				if mapped, ok := idMap[matte]; ok {
					binary.BigEndian.PutUint32(ldta.Data[0xA0:0xA4], mapped)
				}
			}
		}
	}
	return layrs, nil
}
