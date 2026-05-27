package aep

import "github.com/example/aep-parser/internal/rifx"

// footageBackrefs holds the rifx.Chunk references that power Footage's
// length-preserving write paths (SetPath, SSPC flag setters) plus the
// length-variable item-level Comment / Label writers.
//
// Lifecycle:
//   - Populated by parseFootage + parseItem when a Footage is built from a
//     parsed .aep file.
//   - Nil for footage items built outside the parser.
//   - opaque is reserved for future V3 phases that need to round-trip
//     unrecognized sibling chunks under the footage's owning Item LIST
//     (per CLAUDE.md hard constraint #5). Phase 1 leaves it nil.
type footageBackrefs struct {
	// Underlying chunks holding the source path. Set by the parser when
	// found; SetPath mutates these for write-back.
	aliasChunk *rifx.Chunk // Pin/Als2/alas — JSON with "fullpath"
	cpthChunk  *rifx.Chunk // legacy Cpth chunk, when present

	// sspcChunk is the source-settings chunk (~222 bytes). Holds width /
	// height (already on Footage) plus audio sample rate, start/end
	// frame, footage_missing flag, etc. — used by P1 1H convenience
	// helpers (FootageMissing / HasAudio / StartFrame / EndFrame).
	// Offsets per py-aep binary/footage_chunks.py::SspcChunk.
	sspcChunk *rifx.Chunk

	// Item-level write-back references (used by SetComment / SetLabel).
	itemCmtaChunk  *rifx.Chunk
	itemIdtaChunk  *rifx.Chunk
	itemLayrParent *rifx.Chunk

	opaque map[rifx.ChunkID]*rifx.Chunk
}
