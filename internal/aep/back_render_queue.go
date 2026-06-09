package aep

import (
	"encoding/binary"
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// renderQueueBackrefs holds the RIFX chunk reference that powers the structural
// RenderQueue.RemoveItem write. Nil for queues built outside the parser.
type renderQueueBackrefs struct {
	// lrdr is the owning LIST:LRdr container. RemoveItem reaches the
	// LRdr-level settings list (lhd3 + ldat of 2246B blocks) and the Rout
	// per-item flags chunk through it.
	lrdr *rifx.Chunk
}

// renderQueueItemBackrefs holds the RIFX chunk references that power the
// length-variable RenderQueueItem.SetComment write plus the write-time settings
// sync. Nil for items built outside the parser.
type renderQueueItemBackrefs struct {
	// litm is the owning LIST:LItm container — the parent a fresh RCom is
	// inserted into when the item has no comment yet.
	litm *rifx.Chunk

	// itemListChunk is this item's LIST:list (output-module settings) chunk. It
	// is the insertion anchor: a new RCom goes immediately before it, matching
	// AE / py-aep per-item ordering [RCom] + list + 'LOm '.
	itemListChunk *rifx.Chunk

	// rcomChunk is the existing RCom wrapper leaf, or nil when the item carries
	// no comment. Patched in place on replace.
	rcomChunk *rifx.Chunk

	// settingsSlice aliases this item's 2246-byte region inside the shared
	// LRdr settings ldat. The scene-side RenderQueueItem.settingsBlock is an
	// independent copy (single source of truth); syncRenderQueue copies it back
	// into this alias at WriteAEP time (length-preserving). The structural
	// AddItem / RemoveItem ops re-point it when the shared ldat is spliced.
	settingsSlice []byte
}

// outputModuleBackrefs locates where syncRenderQueue copies a scene-owned
// OutputModule's settings buffers back at WriteAEP time. There is no writer
// interface: every OutputModule setter is a pure scene-buffer mutation (single
// source of truth), so the back exists purely for the write-time sync. Nil for
// modules built outside the parser.
type outputModuleBackrefs struct {
	// settingsSlice aliases this module's 128B OutputModuleSettingsItem inside
	// its owning per-item LIST:list ldat.
	settingsSlice []byte
	// roouSlice aliases this module's Roou chunk Data.
	roouSlice []byte
}

var (
	_ RenderQueueItemWriter = (*renderQueueItemBackrefs)(nil)
	_ RenderQueueWriter     = (*renderQueueBackrefs)(nil)
	_ OutputModuleWriter    = (*outputModuleBackrefs)(nil)
)

func (b *renderQueueBackrefs) isRenderQueueWriter() {}

func (b *outputModuleBackrefs) isOutputModuleWriter() {}

// renderQueueBack returns the concrete backrefs behind a RenderQueue's writer
// interface for serializer-stage raw chunk access (AddItem / RemoveItem reach
// the LRdr container through it). Returns nil when the queue was built outside
// the parser.
func (rq *RenderQueue) renderQueueBack() *renderQueueBackrefs {
	if b, ok := rq.back.(*renderQueueBackrefs); ok {
		return b
	}
	return nil
}

// outputModuleBack returns the concrete backrefs behind an OutputModule's
// writer interface for the write-time settings sync. Returns nil when the
// module was built outside the parser.
func (om *OutputModule) outputModuleBack() *outputModuleBackrefs {
	if b, ok := om.back.(*outputModuleBackrefs); ok {
		return b
	}
	return nil
}

// renderQueueItemBack returns the concrete backrefs behind a RenderQueueItem's
// writer interface for serializer-stage raw chunk access (settings sync +
// structural AddItem / RemoveItem). Returns nil when the item was built outside
// the parser.
func (it *RenderQueueItem) renderQueueItemBack() *renderQueueItemBackrefs {
	if rb, ok := it.back.(*renderQueueItemBackrefs); ok {
		return rb
	}
	return nil
}

// SetComment writes the comment into the item's RCom wrapper chunk: it replaces
// an existing RCom's payload, or inserts a fresh RCom into the LItm LIST
// immediately before the item's settings list (matching AE / py-aep per-item
// ordering [RCom] + list + 'LOm '). Setting "" on an item with no RCom is a
// no-op. WriteAEP recomputes the LItm/LRdr LIST sizes. The scene-side Comment
// field is synced by the RenderQueueItem delegate.
func (b *renderQueueItemBackrefs) SetComment(comment string) error {
	if b.rcomChunk != nil {
		b.rcomChunk.Data = encodeRComData(comment)
		return nil
	}
	if comment == "" {
		return nil
	}
	if b.litm == nil || b.itemListChunk == nil {
		return fmt.Errorf("render queue item: no LItm reference to insert RCom into")
	}
	newRcom := &rifx.Chunk{ID: rifx.IDRCom, Data: encodeRComData(comment)}
	children := b.litm.Children
	pos := len(children)
	for i, ch := range children {
		if ch == b.itemListChunk {
			pos = i
			break
		}
	}
	children = append(children, nil)
	copy(children[pos+1:], children[pos:])
	children[pos] = newRcom
	b.litm.Children = children
	b.rcomChunk = newRcom
	return nil
}

// encodeRComData builds the RCom wrapper body: one embedded Utf8 chunk
// ("Utf8" + big-endian u32 length + raw UTF-8 payload), padded to an even
// length with a trailing NUL when the payload length is odd. This mirrors how
// AE / py-aep serialize the chunk (Utf8 carries no NUL terminator; the pad is
// the RIFX even-boundary pad). The inner chunk's even length keeps the RCom
// body even, so the RCom leaf itself never needs an outer pad.
func encodeRComData(comment string) []byte {
	payload := []byte(comment)
	out := make([]byte, 0, 8+len(payload)+1)
	out = append(out, 'U', 't', 'f', '8')
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(payload)))
	out = append(out, lenBuf[:]...)
	out = append(out, payload...)
	if len(payload)%2 != 0 {
		out = append(out, 0)
	}
	return out
}
