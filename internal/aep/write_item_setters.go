package aep

import "fmt"

// Item-level scene-method setters shared between Composition and Footage. AE's
// project panel exposes a comment column and a color label per item; both
// delegate through the writer interface (CompositionWriter / FootageWriter) so
// these scene methods never name the concrete back-ref or touch a chunk. The
// serializer helpers they ultimately drive (setItemComment / setItemLabel) live
// in write_item.go.

// SetComment writes a project-panel comment on the composition (Item-
// level, distinct from Layer.SetComment). length-variable: the cmta
// chunk's Data is replaced wholesale; WriteAEP recomputes the parent
// Item LIST size. When no cmta chunk exists yet, a fresh one is
// inserted into the Item LIST.
func (c *Composition) SetComment(comment string) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no Item LIST reference (built outside parser?)", c.Name)
	}
	if err := c.back.SetComment(comment); err != nil {
		return err
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
	if err := c.back.SetLabel(index); err != nil {
		return err
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
