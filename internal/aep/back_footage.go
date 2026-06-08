package aep

import (
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

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
//     (per CLAUDE.md hard constraint #5); currently nil.
type footageBackrefs struct {
	// itemID / itemName are stored for error-message context (mirrors
	// Footage.ID / Footage.Name at parse time; not used for byte writes).
	itemID   uint32
	itemName string

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

var _ FootageWriter = (*footageBackrefs)(nil)

func (b *footageBackrefs) SetPath(newPath string) error {
	if b.aliasChunk == nil && b.cpthChunk == nil {
		return fmt.Errorf("footage %d (%q): no path chunks present (solid/placeholder?)", b.itemID, b.itemName)
	}
	if b.aliasChunk != nil {
		newData, err := replaceJSONStringField(b.aliasChunk.Data, "fullpath", newPath)
		if err != nil {
			return fmt.Errorf("footage %d (%q): rewrite alas fullpath: %w", b.itemID, b.itemName, err)
		}
		b.aliasChunk.Data = newData
	}
	if b.cpthChunk != nil {
		buf := make([]byte, len(newPath)+1)
		copy(buf, newPath)
		b.cpthChunk.Data = buf
	}
	return nil
}

func (b *footageBackrefs) SetComment(comment string) error {
	if err := setItemComment(b.itemLayrParent, &b.itemCmtaChunk, comment); err != nil {
		return fmt.Errorf("footage %q: %w", b.itemName, err)
	}
	return nil
}

func (b *footageBackrefs) SetLabel(index uint8) error {
	if err := setItemLabel(b.itemIdtaChunk, index); err != nil {
		return fmt.Errorf("footage %q: %w", b.itemName, err)
	}
	return nil
}

