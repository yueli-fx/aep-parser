package serializer

import (
	"encoding/binary"
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// isFileBacked reports whether f is file footage eligible for path-based dedup
// (not solid/placeholder, with a non-empty Path).
func isFileBacked(f *Footage) bool {
	return f != nil && !f.IsSolid && !f.IsPlaceholder && f.Path != ""
}

// destFootageByPath returns the first file-backed footage in p whose Path
// equals path, or nil. Used for cross-Project footage dedup.
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
// sub-containers) for the Item LIST whose idta item-ID (@codec.IdtaItemID) equals id.
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
		if idta != nil && len(idta.Data) >= codec.IdtaItemID+4 &&
			binary.BigEndian.Uint32(idta.Data[codec.IdtaItemID:codec.IdtaItemID+4]) == id {
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
// Fold into dest's root Fold with a fresh dest item ID (idta @codec.IdtaItemID),
// parses it into dest.Footage, and returns the new dest item ID. On error the
// CALLER (insertLayerCrossProject) restores dest via its outer snapshot — this
// helper does not self-rollback.
func importFootageBlock(dest, src *Project, srcID uint32, name string) (uint32, error) {
	container, start, end := locateItemBlockByID(projectBack(src).rootFold, srcID)
	if container == nil {
		return 0, fmt.Errorf("footage Item block id=%d not found in src project", srcID)
	}
	dup := deepCloneChunk(container.Children[start])
	destID := allocItemID(dest)
	idta := dup.FindFirst(rifx.IDIdta)
	if idta == nil || len(idta.Data) < codec.IdtaItemID+4 {
		return 0, fmt.Errorf("cloned footage id=%d idta missing/short", srcID)
	}
	binary.BigEndian.PutUint32(idta.Data[codec.IdtaItemID:codec.IdtaItemID+4], destID)
	destRoot := projectBack(dest).rootFold
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

// insertLayerCrossProject handles InsertLayer when src lives in a different
// Project than c. It imports src's reachable item closure (footage + precomp,
// transitively) into c's Project at root level with fresh item IDs, dedup'ing
// file-backed footage by Path, then splices the layer via spliceLayerClone with
// SourceID/AlternateSourceID remapped to the imported dest items. Atomic: a
// seven-way dest snapshot + warnings-as-failure rollback covers both phases.
// srcChildren/srcLayrIdx are the located src Layr position from InsertLayer.
func insertLayerCrossProject(c *Composition, src *Layer, atIdx, srcLayrIdx int, srcChildren []*rifx.Chunk) (*Layer, error) {
	destProj := scene.CompositionProj(c)
	srcProj := scene.CompositionProj(scene.LayerComp(src))

	destPb := projectBack(destProj)
	if destPb == nil || destPb.rootFold == nil {
		return nil, fmt.Errorf("InsertLayer: dest Project has no root Fold back-ref (built outside parser?)")
	}
	if srcProj == nil {
		return nil, fmt.Errorf("InsertLayer: src layer's Project is unknown (src.comp.proj == nil)")
	}
	srcPb := projectBack(srcProj)
	if srcPb == nil || srcPb.rootFold == nil {
		return nil, fmt.Errorf("InsertLayer: src Project has no root Fold back-ref")
	}
	rootFold := destPb.rootFold

	// === Outer snapshot (covers closure import + the layer splice) ===
	oldRootChildren := append([]*rifx.Chunk(nil), rootFold.Children...)
	oldComps := append([]*Composition(nil), destProj.Compositions...)
	oldFootage := append([]*Footage(nil), destProj.Footage...)
	destCompCb := compositionBack(c)
	if destCompCb == nil || destCompCb.itemList == nil {
		return nil, fmt.Errorf("InsertLayer: dest comp %q has no itemList back-ref", c.Name)
	}
	oldDestItemList := append([]*rifx.Chunk(nil), destCompCb.itemList.Children...)
	oldDestLayers := append([]*Layer(nil), c.Layers...)
	oldNextItemID := scene.ProjectNextItemID(destProj)
	oldWarningsLen := len(destProj.Warnings)
	rollback := func() {
		rootFold.Children = oldRootChildren
		destProj.Compositions = oldComps
		destProj.Footage = oldFootage
		destCompCb.itemList.Children = oldDestItemList
		c.Layers = oldDestLayers
		scene.SetProjectNextItemID(destProj, oldNextItemID)
		if len(destProj.Warnings) > oldWarningsLen {
			destProj.Warnings = destProj.Warnings[:oldWarningsLen]
		}
	}

	// === Import the source item closure (BFS) ===
	itemIDMap := make(map[uint32]uint32)
	type pendingComp struct {
		dup   *rifx.Chunk
		id    uint32
		name  string
		layrs []*rifx.Chunk
	}
	var pending []pendingComp

	worklist := make([]uint32, 0, 2)
	if src.SourceID != 0 {
		worklist = append(worklist, src.SourceID)
	}
	if src.AlternateSourceID != 0 {
		worklist = append(worklist, src.AlternateSourceID)
	}

	for len(worklist) > 0 {
		srcID := worklist[0]
		worklist = worklist[1:]
		if srcID == 0 {
			continue
		}
		if _, done := itemIDMap[srcID]; done {
			continue
		}
		item := srcProj.AVItemByID(srcID)
		if item == nil {
			rollback()
			return nil, fmt.Errorf("InsertLayer: cross-Project source item id=%d not found in src Project (dangling)", srcID)
		}
		switch it := item.(type) {
		case *Footage:
			if isFileBacked(it) {
				if existing := destFootageByPath(destProj, it.Path); existing != nil {
					itemIDMap[srcID] = existing.ID // dedup hit — reuse, no clone
					continue
				}
			}
			destID, err := importFootageBlock(destProj, srcProj, srcID, it.Name)
			if err != nil {
				rollback()
				return nil, err
			}
			itemIDMap[srcID] = destID
		case *Composition:
			container, start, end := locateItemBlockByID(srcPb.rootFold, srcID)
			if container == nil {
				rollback()
				return nil, fmt.Errorf("InsertLayer: cross-Project comp id=%d Item block not found in src Project", srcID)
			}
			dup := deepCloneChunk(container.Children[start])
			destID := allocItemID(destProj)
			idta := dup.FindFirst(rifx.IDIdta)
			if idta == nil || len(idta.Data) < codec.IdtaItemID+4 {
				rollback()
				return nil, fmt.Errorf("InsertLayer: cloned comp id=%d idta missing/short", srcID)
			}
			binary.BigEndian.PutUint32(idta.Data[codec.IdtaItemID:codec.IdtaItemID+4], destID)
			layrs, err := remapClonedCompLayerLayrs(destProj, dup)
			if err != nil {
				rollback()
				return nil, fmt.Errorf("InsertLayer: comp id=%d: %w", srcID, err)
			}
			rootFold.Children = append(rootFold.Children, dup)
			for k := start + 1; k < end; k++ {
				rootFold.Children = append(rootFold.Children, deepCloneChunk(container.Children[k]))
			}
			itemIDMap[srcID] = destID
			// layrs come from remapClonedCompLayerLayrs, which already errored
			// out unless every layer's ldta is >= 0x88 — so the @0x28 SourceID
			// read here (and in Pass 2 below) is safe without a length guard.
			for _, layr := range layrs {
				ldta := layr.FindFirst(rifx.IDLdta)
				if sid := binary.BigEndian.Uint32(ldta.Data[0x28:0x2C]); sid != 0 {
					worklist = append(worklist, sid)
				}
				if blsi := findAlternateSourceBlsi(layr); blsi != nil && len(blsi.Data) >= 4 {
					if aid := binary.BigEndian.Uint32(blsi.Data[0:4]); aid != 0 {
						worklist = append(worklist, aid)
					}
				}
			}
			pending = append(pending, pendingComp{dup: dup, id: destID, name: it.Name, layrs: layrs})
		}
	}

	// === Second pass: remap imported comps' layer source refs ===
	remap := func(id uint32) uint32 {
		if id == 0 {
			return 0
		}
		if mapped, ok := itemIDMap[id]; ok {
			return mapped
		}
		return id
	}
	for _, pc := range pending {
		for _, layr := range pc.layrs {
			ldta := layr.FindFirst(rifx.IDLdta)
			if sid := binary.BigEndian.Uint32(ldta.Data[0x28:0x2C]); sid != 0 {
				binary.BigEndian.PutUint32(ldta.Data[0x28:0x2C], remap(sid))
			}
			if blsi := findAlternateSourceBlsi(layr); blsi != nil && len(blsi.Data) >= 4 {
				if aid := binary.BigEndian.Uint32(blsi.Data[0:4]); aid != 0 {
					binary.BigEndian.PutUint32(blsi.Data[0:4], remap(aid))
				}
			}
		}
	}

	// === Reparse imported comps (after source remap so refs resolve) ===
	for _, pc := range pending {
		dupComp, err := parseComposition(pc.dup, pc.id, pc.name, &destProj.Warnings)
		if err != nil {
			rollback()
			return nil, fmt.Errorf("InsertLayer: re-parse imported comp %q: %w", pc.name, err)
		}
		scene.SetCompositionProj(dupComp, destProj)
		destProj.Compositions = append(destProj.Compositions, dupComp)
	}

	// === Splice the layer with SourceID/AltSourceID remapped ===
	clone, err := spliceLayerClone(c, atIdx, srcLayrIdx, srcChildren, remap)
	if err != nil {
		rollback()
		return nil, err
	}

	// === Warnings-as-failure (covers import + splice) ===
	if len(destProj.Warnings) > oldWarningsLen {
		newWarnings := append([]string(nil), destProj.Warnings[oldWarningsLen:]...)
		rollback()
		return nil, fmt.Errorf("InsertLayer: cross-Project produced %d parser warning(s), rolled back: %v", len(newWarnings), newWarnings)
	}

	return clone, nil
}

// remapClonedCompLayerLayrs walks dupItemList's Layr LIST children, allocates a
// fresh dest layer ID per layer (rewriting ldta @0x00 and remapping intra-comp
// ParentID @0x84 / explicit matte @0xA0 through the local srcLayerID→destLayerID
// map), and returns the Layr LIST chunks (for the later cross-comp source-ref
// remap pass). Runs the same two-pass remap (alloc IDs, then rewrite refs)
// that DuplicateComposition uses.
func remapClonedCompLayerLayrs(p *Project, dupItemList *rifx.Chunk) ([]*rifx.Chunk, error) {
	idMap := make(map[uint32]uint32)
	var layrs []*rifx.Chunk
	var ldtas []*rifx.Chunk
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
		newID := allocItemID(p)
		idMap[oldID] = newID
		binary.BigEndian.PutUint32(ldta.Data[0x00:0x04], newID)
		layrs = append(layrs, ch)
		ldtas = append(ldtas, ldta)
	}
	for _, ldta := range ldtas {
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
