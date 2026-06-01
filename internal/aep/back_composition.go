package aep

import "github.com/example/aep-parser/internal/rifx"

// compositionBackrefs holds the rifx.Chunk references that power
// Composition's length-preserving write paths (SetBGColor / SetShutterAngle
// / SetMotionBlur* / SetWorkArea cdta writes, length-variable name writes,
// item-level Comment / Label edits) and NewComposition's re-parse closed
// loop.
//
// Lifecycle:
//   - Populated by parseComposition + parseItem when a Composition is built
//     from a parsed .aep file.
//   - Nil for comps built outside the parser (NewComposition builds it
//     explicitly via the re-parse closed loop).
//   - opaque is reserved for future V3 phases that need to round-trip
//     unrecognized sibling chunks under the comp's owning Item LIST
//     (per CLAUDE.md hard constraint #5); currently nil.
type compositionBackrefs struct {
	// cdta is the underlying cdta chunk reference, captured by
	// parseComposition. Used by SetBGColor / SetShutterAngle /
	// SetMotionBlur* / SetWorkArea for length-preserving writes.
	cdta *rifx.Chunk

	// nameChunk is the comp's Utf8 name chunk (length-variable Set name).
	nameChunk *rifx.Chunk

	// prinChunk / prdaChunk are the comp's renderer chunks under the PRin
	// LIST sibling. prin is a fixed 104-byte chunk (renderer match-name +
	// display name, NUL-padded); prda carries renderer-specific options and
	// is variable-length. SetRenderer patches prin in place (length-
	// preserving name fields) and replaces prda wholesale (structural).
	// Both nil when the comp has no PRin LIST.
	prinChunk *rifx.Chunk
	prdaChunk *rifx.Chunk

	// itemList is the cached owning Item LIST chunk; populated by
	// parseComposition. Used by NewComposition (re-parse closed loop) +
	// future structural mutations. derived cache, never owned (see
	// Invariant #8).
	itemList *rifx.Chunk

	// Item-level chunk refs shared with Footage / Folder. Populated by
	// parseItem from the surrounding Item LIST.
	itemCmtaChunk  *rifx.Chunk // cmta sibling under the Item LIST
	itemIdtaChunk  *rifx.Chunk // idta sibling — Label byte at payload @0x3A
	itemLayrParent *rifx.Chunk // Item LIST itself — needed for cmta insertion when missing

	opaque map[rifx.ChunkID]*rifx.Chunk
}
