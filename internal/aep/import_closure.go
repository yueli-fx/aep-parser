package aep

import (
	"encoding/binary"

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

// locateItemBlockByID finds the Item LIST in rootFold.Children whose idta
// item-ID (@idtaItemID) equals id, returning [start, end) covering the Item
// LIST plus its trailing non-Item sibling run. Returns (-1,-1) if not found.
func locateItemBlockByID(rootFold *rifx.Chunk, id uint32) (int, int) {
	children := rootFold.Children
	for i, ch := range children {
		if !isItemList(ch) {
			continue
		}
		idta := ch.FindFirst(rifx.IDIdta)
		if idta == nil || len(idta.Data) < idtaItemID+4 {
			continue
		}
		if binary.BigEndian.Uint32(idta.Data[idtaItemID:idtaItemID+4]) != id {
			continue
		}
		end := i + 1
		for end < len(children) && !isItemList(children[end]) {
			end++
		}
		return i, end
	}
	return -1, -1
}
