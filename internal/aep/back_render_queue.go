package aep

import "github.com/example/aep-parser/internal/rifx"

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
