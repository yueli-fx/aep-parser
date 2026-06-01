package aep

import "github.com/example/aep-parser/internal/rifx"

// renderQueueItemBackrefs holds the RIFX chunk references that power the
// length-variable RenderQueueItem.SetComment write. Nil for items built outside
// the parser.
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
}
