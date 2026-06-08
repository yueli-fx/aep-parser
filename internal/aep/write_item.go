package aep

import (
	"fmt"

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
	encoded := encodeCmta(comment)
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

// SetComment writes a project-panel comment on the composition (Item-
// level, distinct from Layer.SetComment). length-variable: the cmta
// chunk's Data is replaced wholesale; WriteAEP recomputes the parent
// Item LIST size. When no cmta chunk exists yet, a fresh one is
// inserted into the Item LIST.
func (c *Composition) SetComment(comment string) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no Item LIST reference (built outside parser?)", c.Name)
	}
	if err := setItemComment(c.back.itemLayrParent, &c.back.itemCmtaChunk, comment); err != nil {
		return fmt.Errorf("comp %q: %w", c.Name, err)
	}
	c.Comment = comment
	return nil
}

// SetLabel writes the project-panel color label index (0..16) for the
// composition (Item-level). Indices outside 0..16 are written verbatim
// (AE shows index 0 for unknown values). length-preserving (1 byte).
func (c *Composition) SetLabel(index uint8) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no idta chunk reference", c.Name)
	}
	if err := setItemLabel(c.back.itemIdtaChunk, index); err != nil {
		return fmt.Errorf("comp %q: %w", c.Name, err)
	}
	c.Label = index
	return nil
}

// SetComment writes a project-panel comment on the footage item.
// Length-variable; same semantics as Composition.SetComment.
func (f *Footage) SetComment(comment string) error {
	if f.back == nil {
		return fmt.Errorf("footage %q: no Item LIST reference (built outside parser?)", f.Name)
	}
	if err := f.back.SetComment(comment); err != nil {
		return err
	}
	f.Comment = comment
	return nil
}

// SetLabel writes the project-panel color label index for the footage.
func (f *Footage) SetLabel(index uint8) error {
	if f.back == nil {
		return fmt.Errorf("footage %q: no idta chunk reference", f.Name)
	}
	if err := f.back.SetLabel(index); err != nil {
		return err
	}
	f.Label = index
	return nil
}
