// internal/aep/import_placeholder.go
//
// Task 5: ImportPlaceholder implementation following the NewComposition pattern.
// Creates a placeholder footage item in the project's root folder.
package aep

import (
	"encoding/binary"
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// buildPlaceholderIdta constructs a 20-byte idta chunk for a placeholder footage item.
// Type code 0x07 = footage (per workshop/plans/finish/2026-05-22-v2-1-foundation-plan.md).
func buildPlaceholderIdta(itemID uint32) *rifx.Chunk {
	data := make([]byte, 20)
	binary.BigEndian.PutUint16(data[0:2], 0x07) // type code = footage
	binary.BigEndian.PutUint32(data[16:20], itemID)
	return &rifx.Chunk{
		ID:   rifx.IDIdta,
		Data: data,
	}
}

// buildPlaceholderOpti constructs the opti chunk for a placeholder footage.
// Format: "Plac" tag at offset 0, name at offset 0x0A.
func buildPlaceholderOpti(name string) *rifx.Chunk {
	// opti chunk is 266 bytes (from AE 2020 fixture)
	data := make([]byte, 266)
	copy(data[0:4], []byte("Plac"))
	// Name at offset 0x0A
	copy(data[0x0A:0x0A+len(name)], []byte(name))
	return &rifx.Chunk{
		ID:   rifx.IDOpti,
		Data: data,
	}
}

// buildPlaceholderSspc constructs the sspc chunk for a placeholder footage.
// Real AE sspc is 222 bytes with width/height at @0x20/@0x24.
func buildPlaceholderSspc(width, height uint16) *rifx.Chunk {
	data := make([]byte, 222)
	binary.BigEndian.PutUint16(data[0x20:0x22], width)
	binary.BigEndian.PutUint16(data[0x24:0x26], height)
	return &rifx.Chunk{
		ID:   rifx.IDSspc,
		Data: data,
	}
}

// buildPlaceholderPin constructs the Pin LIST containing the placeholder footage descriptor.
func buildPlaceholderPin(name string, width, height uint16) *rifx.Chunk {
	sspc := buildPlaceholderSspc(width, height)
	opti := buildPlaceholderOpti(name)
	utf8 := &rifx.Chunk{
		ID:   rifx.IDUtf8,
		Data: []byte(name),
	}

	return &rifx.Chunk{
		ID:       rifx.IDList,
		FormType: rifx.IDPin,
		Children: []*rifx.Chunk{sspc, opti, utf8},
	}
}

// buildPlaceholderItem constructs the Item LIST for a placeholder footage.
func buildPlaceholderItem(target AETarget, itemID uint32, name string, width, height uint16) *rifx.Chunk {
	iide := &rifx.Chunk{
		ID:   rifx.ChunkID{'i', 'i', 'd', 'e'},
		Data: []byte{0x01, 0x00, 0x00, 0x00},
	}
	idpc := &rifx.Chunk{
		ID:   rifx.ChunkID{'i', 'd', 'p', 'c'},
		Data: make([]byte, 8),
	}
	idta := buildPlaceholderIdta(itemID)
	utf8 := &rifx.Chunk{
		ID:   rifx.IDUtf8,
		Data: []byte(name),
	}
	pin := buildPlaceholderPin(name, width, height)

	return &rifx.Chunk{
		ID:       rifx.IDList,
		FormType: rifx.IDItem,
		Children: []*rifx.Chunk{iide, idpc, idta, utf8, pin},
	}
}

// ImportPlaceholder creates a new placeholder footage item in the project's root folder.
// This mirrors py-aep's Project.ImportPlaceholder API.
//
// Parameters:
//   name     - placeholder name (empty string defaults to "Placeholder")
//   width    - width in pixels (must be >= 4 and <= 30000)
//   height   - height in pixels (must be >= 4 and <= 30000)
//   frameRate - frame rate in Hz (must be >= 1.0 and <= 99.0)
//   duration - duration in seconds (must be > 0 and <= 10800)
//
// Returns the new Footage item or an error if validation fails or parsing fails.
//
// Atomic mutation (Invariant #10): if chunk parse fails or warnings appear,
// rollback chunk-tree + typed index + warnings to pre-call state.
func (p *Project) ImportPlaceholder(
	name string,
	width, height int,
	frameRate, duration float64,
) (*Footage, error) {
	// 1. Validate inputs
	if err := validateImportPlaceholderInputs(name, width, height, frameRate, duration); err != nil {
		return nil, err
	}

	// 2. Normalize name
	if name == "" {
		name = "Placeholder"
	}

	// 3. Allocate ID (monotonic; never reuses)
	id := p.allocItemID()

	// 4. Build chunks
	itemList := buildPlaceholderItem(p.target, id, name, uint16(width), uint16(height))

	// 5. Atomic mutation prep
	if p.rootFold == nil {
		return nil, fmt.Errorf("internal: project missing root Fold (template malformed?)")
	}
	oldChildLen := len(p.rootFold.Children)
	oldWarningsLen := len(p.Warnings)

	// 6. Append to rootFold + reparse closed loop
	p.rootFold.Children = append(p.rootFold.Children, itemList)
	p.rootFold.Children = append(p.rootFold.Children, lowerItemSiblings(nil)...)

	// Reparse the new footage item
	footage, err := parseFootage(itemList, id, name)
	if err != nil {
		// Rollback
		p.rootFold.Children = p.rootFold.Children[:oldChildLen]
		p.Warnings = p.Warnings[:oldWarningsLen]
		return nil, fmt.Errorf("internal: re-parsing new placeholder: %w", err)
	}

	// 7. Warnings-as-failure (Invariant #11)
	if len(p.Warnings) != oldWarningsLen {
		newWarnings := append([]string(nil), p.Warnings[oldWarningsLen:]...)
		p.rootFold.Children = p.rootFold.Children[:oldChildLen]
		p.Warnings = p.Warnings[:oldWarningsLen]
		return nil, fmt.Errorf("internal: builder produced %d parser warning(s): %v", len(newWarnings), newWarnings)
	}

	// 8. Register in typed index
	p.Footage = append(p.Footage, footage)

	return footage, nil
}

// validateImportPlaceholderInputs returns nil if all inputs are valid, or
// an error naming the offending field + value.
func validateImportPlaceholderInputs(name string, width, height int, frameRate, duration float64) error {
	if width < 4 || width > 30000 {
		return fmt.Errorf("width must be between 4 and 30000 (got %d)", width)
	}
	if height < 4 || height > 30000 {
		return fmt.Errorf("height must be between 4 and 30000 (got %d)", height)
	}
	if frameRate < 1.0 || frameRate > 99.0 {
		return fmt.Errorf("frameRate must be between 1.0 and 99.0 (got %g)", frameRate)
	}
	if duration <= 0 || duration > 10800 {
		return fmt.Errorf("duration must be > 0 and <= 10800 (got %g)", duration)
	}
	return nil
}
