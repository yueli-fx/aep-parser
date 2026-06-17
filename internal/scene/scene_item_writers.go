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
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompIdta_AEShipGate_AE2020,TestCompIdta_AEShipGate_AE2025 boundary="length-variable(cmta 整片替换+父 LIST size 重算);无 cmta 时插入到 Utf8 之后(位置敏感,RE re_comp_idta.aep:末尾 append→AE 打不开)+ 设 idta @0x39 has-comment flag(否则 AE 丢 comment);双版本 AE gated(comp_idta,comp.comment DOM readback)" alias="comp comment,合成备注,项目面板备注,item comment"
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
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompIdta_AEShipGate_AE2020,TestCompIdta_AEShipGate_AE2025 boundary="length-preserving(1B,idta @0x3A);越界值照写(AE 视为 0);双版本 AE gated(comp_idta,comp.label DOM readback=9)" alias="label,color label,颜色标签,项目面板标签,comp label"
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
//aep:cap domain=io tier=stable verify=ae-accept gate=TestFootageIdta_AEShipGate_AE2020,TestFootageIdta_AEShipGate_AE2025 boundary="length-variable(cmta 整片替换+父 LIST size 重算;无 cmta 时插入到 Utf8 之后 + 设 idta @0x39 flag——与 comp setItemComment 同源);双版本 AE gated(footage_idta,单 solid 载体 footage.comment DOM readback,solid Item Utf8 空→cmta 插在该空 Utf8 后)" alias="footage comment,素材备注,项目面板备注,footage item comment"
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
//aep:cap domain=io tier=stable verify=ae-accept gate=TestFootageIdta_AEShipGate_AE2020,TestFootageIdta_AEShipGate_AE2025 boundary="length-preserving(1B,idta @0x3A);越界值照写(AE 视为 0);双版本 AE gated(footage_idta,footage.label DOM readback=9)" alias="footage label,素材颜色标签,label,color label"
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
