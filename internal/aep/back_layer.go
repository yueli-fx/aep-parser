package aep

import "github.com/example/aep-parser/internal/rifx"

// layerBackrefs holds the rifx.Chunk references that power Layer's
// length-preserving write paths (SetName / SetComment / SetVisible /
// SetBlendingMode / SetAlternateSource / per-text-run setters).
//
// Lifecycle:
//   - Populated by parseLayer when a Layer is built from a parsed .aep file.
//   - Nil for layers built outside the parser (NewProject / NewShapeLayer
//     builders construct it explicitly per their archetype).
//   - opaque is reserved for future V3 phases that need to round-trip
//     unrecognized sibling chunks under the layer's owning Layr LIST
//     (per CLAUDE.md hard constraint #5). Phase 1 leaves it nil.
type layerBackrefs struct {
	// ldta is the underlying ldta chunk reference, captured by parseLayer.
	// Used by SetVisible/SetBlendingMode/etc. for length-preserving
	// flag-bit and byte-field writes.
	ldta *rifx.Chunk

	// nameChunk is the layer's name Utf8 chunk. Used by SetName for
	// length-variable text replacement.
	nameChunk *rifx.Chunk

	// commentChunk is the layer's cmta chunk (may be nil when no
	// comment was set). SetComment replaces its data or creates one.
	commentChunk *rifx.Chunk

	// layrList is the owning Layr LIST itself — needed when SetComment
	// has to insert a fresh cmta chunk (no existing one to mutate).
	layrList *rifx.Chunk

	// btdsChunk is the btds LIST holding the text source bytes
	// (TextSourceRaw is an alias of this chunk's Data). Length-variable
	// text writes (per-run setters) update this chunk's Data to point
	// at a fresh splice; WriteAEP recomputes parent LIST sizes.
	btdsChunk *rifx.Chunk

	// alternateSourceBlsi is the underlying blsi chunk (4-byte BE uint32
	// holding the alt source AVItem id). nil for layers without an
	// Essential Properties media-replacement slot. SetAlternateSource
	// rewrites its first 4 data bytes in place (length-preserving).
	alternateSourceBlsi *rifx.Chunk

	opaque map[rifx.ChunkID]*rifx.Chunk
}
