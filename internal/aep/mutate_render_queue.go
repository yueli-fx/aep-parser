package aep

import (
	"encoding/binary"
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// RemoveItem deletes the render queue item at index (0-based), mirroring
// ExtendScript RenderQueueItem.remove(). Alpha / structural.
//
// Byte mechanics REd from AE 2020 (test_data/re_rq_delete.jsx, 2-item→1-item
// diff): removing item i drops, in lock-step,
//
//   - the item's [RCom?] + LIST:list + LIST:'LOm ' from the LItm container,
//   - the item's 2246-byte block from the LRdr-level settings ldat, and
//     decrements the settings lhd3 count (@0x08 and @0x0C),
//   - the item's per-item block from the Rout flags chunk (4-byte header +
//     uniform per-item stride), decrementing the header proportionally.
//
// Because every item's settingsBlock aliases a sub-slice of the shared settings
// ldat, the surviving items are re-aliased to their new offsets after the splice.
//
// Alpha: structural delete is not yet AE-ship-gated. Only items with one output
// module are covered by the Rout RE (uniform per-item stride); see
// incidents/render-queue-delete-mechanics.md.
func (rq *RenderQueue) RemoveItem(index int) error {
	if rq == nil {
		return fmt.Errorf("RemoveItem: nil render queue")
	}
	if index < 0 || index >= len(rq.Items) {
		return fmt.Errorf("RemoveItem: index %d out of range (have %d items)", index, len(rq.Items))
	}
	if rq.back == nil || rq.back.lrdr == nil {
		return fmt.Errorf("RemoveItem: render queue built outside parser (no LRdr back-ref)")
	}
	item := rq.Items[index]
	if item.back == nil || item.back.litm == nil || item.back.itemListChunk == nil {
		return fmt.Errorf("RemoveItem: item %d has no LItm back-refs", index)
	}

	litm := item.back.litm
	listIdx := indexOfChunk(litm.Children, item.back.itemListChunk)
	if listIdx < 0 {
		return fmt.Errorf("RemoveItem: item %d list chunk not found in LItm", index)
	}
	// The item's 'LOm ' group is the next LOm sibling after its settings list.
	lomIdx := -1
	for j := listIdx + 1; j < len(litm.Children); j++ {
		if litm.Children[j].IsList() && litm.Children[j].FormType == rifx.IDLOm {
			lomIdx = j
			break
		}
	}
	if lomIdx < 0 {
		return fmt.Errorf("RemoveItem: item %d 'LOm ' group not found", index)
	}

	// === Locate the LRdr-level settings list (sibling of LItm) ===
	settingsList := rq.back.lrdr.FindFirstList(rifx.IDkfl)
	if settingsList == nil {
		return fmt.Errorf("RemoveItem: LRdr settings list missing")
	}
	ldat := settingsList.FindFirst(rifx.IDLdat)
	lhd3 := settingsList.FindFirst(rifx.IDLhd3)
	if ldat == nil || lhd3 == nil {
		return fmt.Errorf("RemoveItem: settings ldat/lhd3 missing")
	}
	off := index * renderSettingsItemSize
	if off+renderSettingsItemSize > len(ldat.Data) {
		return fmt.Errorf("RemoveItem: settings ldat too short for item %d (len=%d)", index, len(ldat.Data))
	}

	// === Commit: remove the item's [RCom?] + list + LOm from LItm ===
	remove := map[*rifx.Chunk]bool{
		item.back.itemListChunk: true,
		litm.Children[lomIdx]:   true,
	}
	if item.back.rcomChunk != nil {
		remove[item.back.rcomChunk] = true
	}
	kept := make([]*rifx.Chunk, 0, len(litm.Children)-len(remove))
	for _, ch := range litm.Children {
		if !remove[ch] {
			kept = append(kept, ch)
		}
	}
	litm.Children = kept

	// === settings ldat: splice out the 2246B block (fresh slice) ===
	newLdat := make([]byte, 0, len(ldat.Data)-renderSettingsItemSize)
	newLdat = append(newLdat, ldat.Data[:off]...)
	newLdat = append(newLdat, ldat.Data[off+renderSettingsItemSize:]...)
	ldat.Data = newLdat

	// lhd3 count fields @0x08 and @0x0C both track the item count.
	if len(lhd3.Data) >= 0x10 {
		decU32(lhd3.Data[0x08:])
		decU32(lhd3.Data[0x0C:])
	}

	// === Rout: drop the item's per-item block + decrement header ===
	if rout := rq.back.lrdr.FindFirst(rifx.IDRout); rout != nil && len(rout.Data) >= 4 {
		const routHeader = 4
		n := len(rq.Items) // pre-removal count
		if stride := (len(rout.Data) - routHeader) / n; stride > 0 {
			rOff := routHeader + index*stride
			if rOff+stride <= len(rout.Data) {
				newRout := make([]byte, 0, len(rout.Data)-stride)
				newRout = append(newRout, rout.Data[:rOff]...)
				newRout = append(newRout, rout.Data[rOff+stride:]...)
				h := binary.BigEndian.Uint32(newRout[0:4])
				binary.BigEndian.PutUint32(newRout[0:4], h-h/uint32(n))
				rout.Data = newRout
			}
		}
	}

	// === scene: drop item, re-index survivors' settings aliases ===
	rq.Items = append(rq.Items[:index], rq.Items[index+1:]...)
	item.back = nil
	for i, it := range rq.Items {
		if it.settingsBlock != nil {
			it.settingsBlock = ldat.Data[i*renderSettingsItemSize : (i+1)*renderSettingsItemSize]
		}
	}
	return nil
}

// decU32 decrements the big-endian u32 at the start of b in place.
func decU32(b []byte) {
	if len(b) >= 4 {
		binary.BigEndian.PutUint32(b, binary.BigEndian.Uint32(b)-1)
	}
}

// incU32 increments the big-endian u32 at the start of b in place.
func incU32(b []byte) {
	if len(b) >= 4 {
		binary.BigEndian.PutUint32(b, binary.BigEndian.Uint32(b)+1)
	}
}

// AddItem appends a render queue item for comp, mirroring ExtendScript
// RenderQueue.items.add(comp). Alpha / structural.
//
// Strategy (clone + remap, like InsertLayer): the queue's last item is the
// template — its 2246B settings block, [LIST:list + 'LOm '] group, and Rout
// per-item block are deep-cloned, then the clone's comp_id (settings @0x08) is
// repointed at comp. The settings ldat / Rout / lhd3 count grow in lock-step,
// mirroring AE's own items.add() delta (REd from a 1-item→2-item diff). The
// cloned output module keeps the template's path/template (AE accepts it; a
// fresh add would name it after comp — deferred).
//
// Requires at least one existing item to clone from (an empty queue has no
// template). The grown settings ldat reallocates, so every item's settingsBlock
// alias is re-pointed afterward. See incidents/render-queue-delete-mechanics.md.
func (rq *RenderQueue) AddItem(comp *Composition) (*RenderQueueItem, error) {
	if rq == nil {
		return nil, fmt.Errorf("AddItem: nil render queue")
	}
	if comp == nil || comp.proj == nil {
		return nil, fmt.Errorf("AddItem: comp is nil or detached from a project")
	}
	if rq.back == nil || rq.back.lrdr == nil {
		return nil, fmt.Errorf("AddItem: render queue built outside parser (no LRdr back-ref)")
	}
	if len(rq.Items) == 0 {
		return nil, fmt.Errorf("AddItem: empty queue has no template item to clone")
	}
	template := rq.Items[len(rq.Items)-1]
	if template.back == nil || template.back.litm == nil || template.back.itemListChunk == nil {
		return nil, fmt.Errorf("AddItem: template item has no LItm back-refs")
	}

	litm := template.back.litm
	listIdx := indexOfChunk(litm.Children, template.back.itemListChunk)
	if listIdx < 0 {
		return nil, fmt.Errorf("AddItem: template list chunk not found in LItm")
	}
	lomIdx := -1
	for j := listIdx + 1; j < len(litm.Children); j++ {
		if litm.Children[j].IsList() && litm.Children[j].FormType == rifx.IDLOm {
			lomIdx = j
			break
		}
	}
	if lomIdx < 0 {
		return nil, fmt.Errorf("AddItem: template 'LOm ' group not found")
	}

	settingsList := rq.back.lrdr.FindFirstList(rifx.IDkfl)
	if settingsList == nil {
		return nil, fmt.Errorf("AddItem: LRdr settings list missing")
	}
	ldat := settingsList.FindFirst(rifx.IDLdat)
	lhd3 := settingsList.FindFirst(rifx.IDLhd3)
	if ldat == nil || lhd3 == nil {
		return nil, fmt.Errorf("AddItem: settings ldat/lhd3 missing")
	}
	n := len(rq.Items)
	tOff := (n - 1) * renderSettingsItemSize
	if tOff+renderSettingsItemSize > len(ldat.Data) {
		return nil, fmt.Errorf("AddItem: settings ldat too short for template block")
	}

	// === Clone template settings block + remap comp_id ===
	newBlock := append([]byte(nil), ldat.Data[tOff:tOff+renderSettingsItemSize]...)
	binary.BigEndian.PutUint32(newBlock[rsCompID:], comp.ID)
	ldat.Data = append(ldat.Data, newBlock...)
	incU32(lhd3.Data[0x08:])
	incU32(lhd3.Data[0x0C:])

	// === Clone the [list, LOm] group + append to LItm ===
	clonedList := deepCloneChunk(template.back.itemListChunk)
	clonedLOm := deepCloneChunk(litm.Children[lomIdx])
	litm.Children = append(litm.Children, clonedList, clonedLOm)

	// === Rout: clone template's per-item block + grow header ===
	if rout := rq.back.lrdr.FindFirst(rifx.IDRout); rout != nil && len(rout.Data) >= 4 {
		const routHeader = 4
		if stride := (len(rout.Data) - routHeader) / n; stride > 0 {
			tRoutOff := routHeader + (n-1)*stride
			if tRoutOff+stride <= len(rout.Data) {
				block := append([]byte(nil), rout.Data[tRoutOff:tRoutOff+stride]...)
				h := binary.BigEndian.Uint32(rout.Data[0:4])
				rout.Data = append(rout.Data, block...)
				binary.BigEndian.PutUint32(rout.Data[0:4], h+h/uint32(n))
			}
		}
	}

	// === scene: build the new item, re-alias ALL settings blocks (ldat grew) ===
	blocks := renderSettingsBlocks(rq.back.lrdr)
	newItem := buildRenderQueueItem(blocks, n, "", clonedList, clonedLOm, comp.proj)
	newItem.back = &renderQueueItemBackrefs{litm: litm, itemListChunk: clonedList}
	rq.Items = append(rq.Items, newItem)
	for i, it := range rq.Items {
		if it.settingsBlock != nil || i == n {
			it.settingsBlock = ldat.Data[i*renderSettingsItemSize : (i+1)*renderSettingsItemSize]
		}
	}
	return newItem, nil
}
