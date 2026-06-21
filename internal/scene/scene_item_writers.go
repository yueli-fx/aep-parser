package scene

import "fmt"

// Item-level scene-method setters shared between Composition and Footage. AE's
// project panel exposes a comment column and a color label per item; both
// delegate through the writer interface (CompositionWriter / FootageWriter) so
// these scene methods never name the concrete back-ref or touch a chunk. The
// serializer helpers they ultimately drive (setItemComment / setItemLabel) live
// in write_item.go.

// @summary    Set the project-panel comment on a composition
// @param      comment  the comment text to write
// @domain     comp
// @stability  stable
// @verify     ae-accept
// @gate       TestCompIdta_AEShipGate_AE2020,TestCompIdta_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-variable: the cmta chunk's data is replaced wholesale
//   and WriteAEP recomputes the parent Item LIST size. When no cmta chunk
//   exists yet, a fresh one is inserted right after the Utf8 chunk
//   (position-sensitive — appending at the end of the LIST produces a file
//   AE refuses to open), and the idta byte at offset 0x39 has-comment flag
//   is set, otherwise AE silently drops the comment.
// @alias      comp comment,合成备注,项目面板备注,item comment
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

// @summary    Set the project-panel color label index on a composition
// @param      index  the color label index (0..16)
// @domain     comp
// @stability  stable
// @verify     ae-accept
// @gate       TestCompIdta_AEShipGate_AE2020,TestCompIdta_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, 1 byte at idta offset 0x3A. Values outside
//   0..16 are written verbatim and AE displays them as index 0.
// @alias      label,color label,颜色标签,项目面板标签,comp label
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

// @summary    Set the project-panel comment on a footage item
// @param      comment  the comment text to write
// @domain     io
// @stability  stable
// @verify     ae-accept
// @gate       TestFootageIdta_AEShipGate_AE2020,TestFootageIdta_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-variable: the cmta chunk's data is replaced wholesale
//   and the parent Item LIST size is recomputed. When no cmta chunk exists
//   yet, a fresh one is inserted right after the Utf8 chunk and the idta
//   byte at offset 0x39 has-comment flag is set (shared mechanics with
//   Composition.SetComment). On a solid footage item the Utf8 chunk is
//   empty, so the cmta chunk is inserted right after that empty Utf8 chunk.
// @alias      footage comment,素材备注,项目面板备注,footage item comment
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

// @summary    Set the project-panel color label index on a footage item
// @param      index  the color label index (0..16)
// @domain     io
// @stability  stable
// @verify     ae-accept
// @gate       TestFootageIdta_AEShipGate_AE2020,TestFootageIdta_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, 1 byte at idta offset 0x3A. Values outside
//   0..16 are written verbatim and AE displays them as index 0.
// @alias      footage label,素材颜色标签,label,color label
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
