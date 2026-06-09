package serializer

import (
	"fmt"

	"github.com/example/aep-parser/internal/codec"
	"github.com/example/aep-parser/internal/rifx"
)

// Item-level setters shared between Composition and Footage. AE's
// project panel exposes a comment column and a color label per item;
// both live in chunks under the Item LIST (cmta for the comment text,
// idta for the label byte at payload @0x3A).

// setItemComment updates / inserts the cmta chunk under the given
// Item LIST. Length-variable — the cmta payload uses the same LF→CRLF
// + NUL terminator format as Layer.SetComment.
func setItemComment(itemList *rifx.Chunk, existing **rifx.Chunk, comment string) error {
	if itemList == nil {
		return fmt.Errorf("no Item LIST reference (built outside parser?)")
	}
	encoded := codec.EncodeCmta(comment)
	if *existing != nil {
		(*existing).Data = encoded
		return nil
	}
	// Insert a new cmta chunk into the Item LIST's Children.
	newCmta := &rifx.Chunk{ID: rifx.IDCmta, Data: encoded}
	itemList.Children = append(itemList.Children, newCmta)
	*existing = newCmta
	return nil
}

// setItemLabel writes a 1-byte label index to the given idta chunk
// at payload @0x3A. Length-preserving (1 byte). Returns an error if
// the idta is missing or shorter than 0x3B bytes (real-world idta is
// 84 bytes, so this is just a sanity guard).
func setItemLabel(idta *rifx.Chunk, index uint8) error {
	if idta == nil {
		return fmt.Errorf("no idta chunk reference")
	}
	if len(idta.Data) <= 0x3A {
		return fmt.Errorf("idta too short for label write (len=%d)", len(idta.Data))
	}
	idta.Data[0x3A] = index
	return nil
}
