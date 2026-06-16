package scene

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
//
//aep:cap domain=comp tier=stable verify=roundtrip boundary="length-variable(cmta 整片替换+父 LIST size 重算;无 cmta 时插入新块);无专门 AE gate→round-trip" alias="comp comment,合成备注,项目面板备注,item comment"
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
//
//aep:cap domain=comp tier=stable verify=roundtrip boundary="length-preserving(1B);越界值照写(AE 视为 0);无专门 AE gate→round-trip" alias="label,color label,颜色标签,项目面板标签,comp label"
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
//
//aep:cap domain=io tier=stable verify=roundtrip boundary="length-variable(cmta 整片替换+父 LIST size 重算);无专门 AE gate→round-trip" alias="footage comment,素材备注,项目面板备注,footage item comment"
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
//
//aep:cap domain=io tier=stable verify=roundtrip boundary="length-preserving(1B);越界值照写;无专门 AE gate→round-trip" alias="footage label,素材颜色标签,label,color label"
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
