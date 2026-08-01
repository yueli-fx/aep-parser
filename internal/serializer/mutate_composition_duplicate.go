package serializer

import (
	"encoding/binary"
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// DuplicateComposition deep-clones src (a comp in this Project) as a new
// sibling comp named name, appended to p.Compositions. The dup contains a
// fresh copy of every layer (new layer IDs), with intra-comp parent +
// track-matte refs remapped to the dup's own layers; layer SOURCES
// (footage / precomp items) are shared verbatim, not duplicated — matching
// AE ScriptingAPI's CompItem.duplicate(). Returns the new *Composition.
// (Full contract + RE notes live on the aep.DuplicateComposition facade — docgen source.)
func DuplicateComposition(p *Project, src *Composition, name string) (*Composition, error) {
	// === Refuse-case matrix R1-R7 ===
	if src == nil {
		return nil, fmt.Errorf("DuplicateComposition: src cannot be nil")
	}
	pb := projectBack(p)
	if pb == nil || pb.rootFold == nil {
		return nil, fmt.Errorf("DuplicateComposition: project has no root Fold back-ref (built outside parser?)")
	}
	srcCb := compositionBack(src)
	if srcCb == nil || srcCb.itemList == nil {
		return nil, fmt.Errorf("DuplicateComposition: src comp %q has no itemList back-ref", src.Name)
	}
	if scene.CompositionProj(src) != p {
		return nil, fmt.Errorf("DuplicateComposition: src comp %q does not belong to this Project", src.Name)
	}
	if name == "" {
		return nil, fmt.Errorf("DuplicateComposition: name cannot be empty")
	}

	rootChildren := pb.rootFold.Children
	srcItemIdx := indexOfChunk(rootChildren, srcCb.itemList)
	if srcItemIdx < 0 {
		return nil, fmt.Errorf("DuplicateComposition: src comp %q Item LIST not found in root Fold", src.Name)
	}

	// === Trailing-sibling run: chunks after the Item until the next Item/EOF ===
	// (AE keeps a run of FEE LIST + small flag chunks after every comp Item;
	// see new_composition.go lowerItemSiblings. Clone src's actual run verbatim.)
	sibEnd := srcItemIdx + 1
	for sibEnd < len(rootChildren) && !isItemList(rootChildren[sibEnd]) {
		sibEnd++
	}

	// === Snapshot for rollback ===
	oldRootChildren := append([]*rifx.Chunk(nil), rootChildren...)
	oldComps := append([]*Composition(nil), p.Compositions...)
	oldItems := append([]ProjectItem(nil), p.Items...)
	oldNextItemID := scene.ProjectNextItemID(p)
	oldWarningsLen := len(p.Warnings)

	// === Deep-clone comp Item block (fresh Data slices) ===
	dupItemList := deepCloneChunk(srcCb.itemList)
	dupSiblings := make([]*rifx.Chunk, 0, sibEnd-(srcItemIdx+1))
	for k := srcItemIdx + 1; k < sibEnd; k++ {
		dupSiblings = append(dupSiblings, deepCloneChunk(rootChildren[k]))
	}

	// === New comp item ID (idta @0x10) ===
	newCompID := allocItemID(p)
	dupIdta := dupItemList.FindFirst(rifx.IDIdta)
	if dupIdta == nil || len(dupIdta.Data) < codec.IdtaItemID+4 {
		scene.SetProjectNextItemID(p, oldNextItemID)
		return nil, fmt.Errorf("DuplicateComposition: cloned comp idta missing or too short for item ID write")
	}
	binary.BigEndian.PutUint32(dupIdta.Data[codec.IdtaItemID:codec.IdtaItemID+4], newCompID)

	// === Remap pass — the per-comp-clone machinery vs InsertLayer ===
	// Pass A: each Layr LIST child is one layer; alloc a fresh ID and record
	// the srcLayerID→dupLayerID map.
	idMap := make(map[uint32]uint32)
	var layerLdtas []*rifx.Chunk
	for _, ch := range dupItemList.Children {
		if !ch.IsList() || ch.FormType != rifx.IDLayr {
			continue
		}
		ldta := ch.FindFirst(rifx.IDLdta)
		if ldta == nil {
			continue // Ewst-paired empty / non-layer Layr — skip defensively
		}
		if len(ldta.Data) < 0x88 {
			scene.SetProjectNextItemID(p, oldNextItemID)
			return nil, fmt.Errorf("DuplicateComposition: a cloned Layr ldta too short for ParentID write (got %d bytes, need >=0x88)", len(ldta.Data))
		}
		oldID := binary.BigEndian.Uint32(ldta.Data[0x00:0x04])
		newID := allocItemID(p)
		idMap[oldID] = newID
		binary.BigEndian.PutUint32(ldta.Data[0x00:0x04], newID)
		layerLdtas = append(layerLdtas, ldta)
	}

	// Pass B: rewrite intra-comp parent + matte refs through idMap. Refs not
	// in the map (cross-comp / corruption) are left verbatim — the reparse
	// warning path catches anything AE-invalid.
	for _, ldta := range layerLdtas {
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

	// === Name rewrite (length-variable Utf8) ===
	dupUtf8 := dupItemList.FindFirst(rifx.IDUtf8)
	if dupUtf8 == nil {
		scene.SetProjectNextItemID(p, oldNextItemID)
		return nil, fmt.Errorf("DuplicateComposition: cloned comp has no Utf8 name chunk")
	}
	dupUtf8.Data = []byte(name)

	// === Splice dup block into rootFold after src's sibling run ===
	insertAt := sibEnd
	newRootChildren := make([]*rifx.Chunk, 0, len(rootChildren)+1+len(dupSiblings))
	newRootChildren = append(newRootChildren, rootChildren[:insertAt]...)
	newRootChildren = append(newRootChildren, dupItemList)
	newRootChildren = append(newRootChildren, dupSiblings...)
	newRootChildren = append(newRootChildren, rootChildren[insertAt:]...)
	pb.rootFold.Children = newRootChildren

	// === Reparse closed loop ===
	dupComp, parseErr := parseComposition(dupItemList, newCompID, name, &p.Warnings)
	if parseErr != nil {
		pb.rootFold.Children = oldRootChildren
		scene.SetProjectNextItemID(p, oldNextItemID)
		return nil, fmt.Errorf("DuplicateComposition: re-parse cloned comp: %w", parseErr)
	}
	scene.SetCompositionProj(dupComp, p)

	// === Register + warnings-as-failure rollback ===
	p.Compositions = append(p.Compositions, dupComp)
	if len(p.Warnings) > oldWarningsLen {
		pb.rootFold.Children = oldRootChildren
		p.Compositions = oldComps
		scene.SetProjectNextItemID(p, oldNextItemID)
		newWarnings := append([]string(nil), p.Warnings[oldWarningsLen:]...)
		p.Warnings = p.Warnings[:oldWarningsLen]
		return nil, fmt.Errorf("DuplicateComposition: produced %d parser warning(s), rolled back: %v", len(newWarnings), newWarnings)
	}
	if err := rebuildProjectItems(p); err != nil {
		pb.rootFold.Children = oldRootChildren
		p.Compositions = oldComps
		p.Items = oldItems
		scene.SetProjectNextItemID(p, oldNextItemID)
		return nil, fmt.Errorf("DuplicateComposition: rebuild project items: %w", err)
	}

	return dupComp, nil
}

// isItemList reports whether ch is a comp/footage/folder Item LIST.
func isItemList(ch *rifx.Chunk) bool {
	return ch.IsList() && ch.FormType == rifx.IDItem
}
