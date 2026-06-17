package serializer

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
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

	// optiChunk is the footage-options chunk. For "Soli" footage it holds
	// the solid's ARGB color (@0x0A, 4×float32 BE) and display name (@0x1A,
	// NUL-terminated inside a fixed 256-byte tail) — patched in place by the
	// NewSolidLayer / NewNullLayer / NewAdjustmentLayer create paths.
	optiChunk *rifx.Chunk

	// Item-level write-back references (used by SetComment / SetLabel).
	itemCmtaChunk  *rifx.Chunk
	itemIdtaChunk  *rifx.Chunk
	itemLayrParent *rifx.Chunk

	opaque map[rifx.ChunkID]*rifx.Chunk
}

var _ FootageWriter = (*footageBackrefs)(nil)

// footageBack returns the concrete back-refs behind a Footage's writer
// interface for serializer-stage raw chunk access. Returns nil when the
// footage was built outside the parser. Free function (the receiver is a scene
// type post package-split).
func footageBack(f *Footage) *footageBackrefs {
	if fb, ok := scene.FootageBack(f).(*footageBackrefs); ok {
		return fb
	}
	return nil
}

func (b *footageBackrefs) SspcData() []byte {
	if b == nil || b.sspcChunk == nil {
		return nil
	}
	return b.sspcChunk.Data
}

// SetSolidColor patches the solid's RGB into the opti "Soli" chunk
// (@0x0A ARGB 4×float32 BE, alpha pinned to 1.0) — the same length-preserving
// byte patch the ship-gated NewSolidLayer create path applies.
func (b *footageBackrefs) SetSolidColor(rgb [3]float64) error {
	if b == nil || b.optiChunk == nil || len(b.optiChunk.Data) < optiSoliName {
		return fmt.Errorf("footage %d (%q): no opti Soli chunk to hold a solid color", b.itemID, b.itemName)
	}
	d := b.optiChunk.Data
	putF32 := func(off int, v float64) {
		binary.BigEndian.PutUint32(d[off:off+4], math.Float32bits(float32(v)))
	}
	putF32(optiSoliColorA, 1)
	putF32(optiSoliColorR, rgb[0])
	putF32(optiSoliColorG, rgb[1])
	putF32(optiSoliColorB, rgb[2])
	return nil
}

// SetSolidSize patches the solid's pixel dimensions into sspc @0x20/@0x24
// (u16 BE) — same length-preserving patch as the NewSolidLayer create path.
func (b *footageBackrefs) SetSolidSize(width, height uint16) error {
	if b == nil || b.sspcChunk == nil || len(b.sspcChunk.Data) < sspcHeightOff+2 {
		return fmt.Errorf("footage %d (%q): no sspc chunk to hold solid dimensions", b.itemID, b.itemName)
	}
	sd := b.sspcChunk.Data
	binary.BigEndian.PutUint16(sd[sspcWidthOff:sspcWidthOff+2], width)
	binary.BigEndian.PutUint16(sd[sspcHeightOff:sspcHeightOff+2], height)
	return nil
}

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
	if err := setItemComment(b.itemLayrParent, b.itemIdtaChunk, &b.itemCmtaChunk, comment); err != nil {
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
