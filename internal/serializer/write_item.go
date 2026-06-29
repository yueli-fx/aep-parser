package serializer

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

// Item-level setters shared between Composition and Footage. AE's
// project panel exposes a comment column and a color label per item;
// both live in chunks under the Item LIST (cmta for the comment text,
// idta for the label byte at payload @0x3A).

// setItemComment updates / inserts the cmta chunk under the given
// Item LIST and toggles the idta has-comment flag (@0x39). Length-variable
// — the cmta payload uses the same LF→CRLF + double-NUL terminator format as
// Layer.SetComment.
//
// AE silently drops a cmta whose item idta @0x39 flag is 0 (item-level analog
// of the layer ldta @0x3C has-comment flag — see
// incidents/layer-setcomment-cmta-append-position.md; RE-confirmed 2026-06-17
// against re_comp_idta.aep). So the flag MUST be set/cleared alongside the chunk.
func setItemComment(itemList, idta *rifx.Chunk, existing **rifx.Chunk, comment string) error {
	if itemList == nil {
		return fmt.Errorf("no Item LIST reference (built outside parser?)")
	}
	encoded := codec.EncodeCmta(comment)
	if *existing != nil {
		(*existing).Data = encoded
	} else {
		// Insert the new cmta immediately AFTER the Utf8 name chunk — that's
		// where AE places it in an Item LIST (RE re_comp_idta.aep CMT: order is
		// idta Utf8 cmta dats cdta …). A tail-appended cmta (as for a layer's
		// Layr LIST, which AE tolerates) makes AE FAIL TO OPEN the project —
		// the item structure is position-sensitive. (layer-setcomment incident
		// warned this Item-LIST path was unverified; this is the confirmed trap.)
		newCmta := &rifx.Chunk{ID: rifx.IDCmta, Data: encoded}
		insertAt := len(itemList.Children) // fallback: tail
		for i, c := range itemList.Children {
			if c.ID == rifx.IDUtf8 && !c.IsList() {
				insertAt = i + 1
				break
			}
		}
		itemList.Children = append(itemList.Children, nil)
		copy(itemList.Children[insertAt+1:], itemList.Children[insertAt:])
		itemList.Children[insertAt] = newCmta
		*existing = newCmta
	}
	// has-comment flag: 1 when non-empty, 0 when cleared. AE ignores the cmta
	// without it.
	if idta != nil && len(idta.Data) > codec.IdtaHasComment {
		if comment != "" {
			idta.Data[codec.IdtaHasComment] = 1
		} else {
			idta.Data[codec.IdtaHasComment] = 0
		}
	}
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
