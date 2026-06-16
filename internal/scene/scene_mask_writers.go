package scene

import "fmt"

// Mask setters — length-preserving mkif / shph byte writes. Each
// setter delegates byte-patch work to maskBackrefs (MaskWriter), then
// syncs the matching Go field. WriteAEP serializes the change.
//
// Note: vertex / path mutations are NOT supported here (requires
// rewriting variable-length shap kfl streams).

// SetMode writes a new mask Mode enum (uint32 BE) to mkif @0x04.
// length-preserving (4 bytes).
//
//aep:cap domain=mask tier=stable verify=roundtrip boundary="length-preserving(4B mkif @0x04);无专门 AE gate→round-trip" alias="mask mode,遮罩模式,blend mode,add subtract intersect"
func (m *Mask) SetMode(mode MaskMode) error {
	if m.back == nil {
		return fmt.Errorf("mask %q: no mkif chunk (built outside parser?)", m.Name)
	}
	if err := m.back.SetMode(mode); err != nil {
		return err
	}
	m.Mode = mode
	return nil
}

// SetInverted toggles the mask Inverted flag (mkif @0x00).
// length-preserving (1 byte).
//
//aep:cap domain=mask tier=stable verify=roundtrip boundary="length-preserving(1B mkif @0x00);无专门 AE gate→round-trip" alias="mask inverted,遮罩反转,invert mask"
func (m *Mask) SetInverted(v bool) error {
	if m.back == nil {
		return fmt.Errorf("mask %q: no mkif chunk", m.Name)
	}
	if err := m.back.SetInverted(v); err != nil {
		return err
	}
	m.Inverted = v
	return nil
}

// SetColor writes the mask timeline label color RGB to mkif
// @0x2D/@0x2E/@0x2F. AE always keeps alpha (0x2C) at 0xFF;
// this setter does not touch it.
// length-preserving (3 bytes).
//
//aep:cap domain=mask tier=stable verify=roundtrip boundary="length-preserving(3B mkif @0x2D-0x2F);alpha 不动;无专门 AE gate→round-trip" alias="mask color,遮罩颜色,label color,时间线标签色"
func (m *Mask) SetColor(rgb [3]uint8) error {
	if m.back == nil {
		return fmt.Errorf("mask %q: no mkif chunk", m.Name)
	}
	if err := m.back.SetColor(rgb); err != nil {
		return err
	}
	m.Color = rgb
	return nil
}

// MaskMotionBlurMode describes the per-mask motion-blur override.
// AE's MaskMotionBlur enum uses these stored values:
//
//	0 = SameAsLayer (default — mask renders blur iff layer has it)
//	2 = On (force ON regardless of layer)
//	3 = Off (force OFF regardless of layer)
//
// Stored as a single byte at mkif @0x02.
type MaskMotionBlurMode uint8

const (
	MaskMotionBlurSameAsLayer MaskMotionBlurMode = 0
	MaskMotionBlurOn          MaskMotionBlurMode = 2
	MaskMotionBlurOff         MaskMotionBlurMode = 3
)

// SetLocked toggles the mask's lock flag (mkif @0x01). When locked, AE
// refuses edits to the mask in the timeline UI; the bytes are still
// mutable through this library.
// length-preserving (1 byte).
//
//aep:cap domain=mask tier=stable verify=roundtrip boundary="length-preserving(1B mkif @0x01);AE UI 锁定但字节仍可写;无专门 AE gate→round-trip" alias="mask locked,遮罩锁定,lock mask"
func (m *Mask) SetLocked(v bool) error {
	if m.back == nil {
		return fmt.Errorf("mask %q: no mkif chunk", m.Name)
	}
	if err := m.back.SetLocked(v); err != nil {
		return err
	}
	m.Locked = v
	return nil
}

// SetMaskMotionBlur writes the per-mask motion-blur override at mkif
// @0x02. Valid values: MaskMotionBlurSameAsLayer (0),
// MaskMotionBlurOn (2), MaskMotionBlurOff (3).
// length-preserving (1 byte).
//
//aep:cap domain=mask tier=stable verify=roundtrip boundary="length-preserving(1B mkif @0x02);无专门 AE gate→round-trip" alias="mask motion blur,遮罩运动模糊,motion blur override"
func (m *Mask) SetMaskMotionBlur(mode MaskMotionBlurMode) error {
	if m.back == nil {
		return fmt.Errorf("mask %q: no mkif chunk", m.Name)
	}
	if err := m.back.SetMaskMotionBlur(mode); err != nil {
		return err
	}
	m.MotionBlur = mode
	return nil
}

// MaskFeatherFalloff describes the mask feather decay curve (AE's
// MaskFeatherFalloff enum FFO_SMOOTH/FFO_LINEAR). Stored as a single byte at
// mkif @0x03 (RE'd 2026-06-17 by diffing two AE-native masks Smooth vs Linear;
// the parser previously did not read this byte).
//
//	0 = Smooth (FFO_SMOOTH, AE default)
//	1 = Linear (FFO_LINEAR)
type MaskFeatherFalloff uint8

const (
	MaskFeatherFalloffSmooth MaskFeatherFalloff = 0
	MaskFeatherFalloffLinear MaskFeatherFalloff = 1
)

// SetFeatherFalloff writes the mask's feather-falloff curve at mkif @0x03.
// Valid values: MaskFeatherFalloffSmooth (0, default), MaskFeatherFalloffLinear (1).
// length-preserving (1 byte).
//
//aep:cap domain=mask tier=stable verify=ae-accept gate=TestMaskFeatherFalloff_AEShipGate_AE2020,TestMaskFeatherFalloff_AEShipGate_AE2025 boundary="length-preserving(1B mkif @0x03);RE'd 2026-06-17(parser 此前漏读该字节);非渲染→AE DOM readback gate(maskFeatherFalloff enum)" alias="mask feather falloff,遮罩羽化衰减,feather falloff curve,smooth linear feather"
func (m *Mask) SetFeatherFalloff(falloff MaskFeatherFalloff) error {
	if m.back == nil {
		return fmt.Errorf("mask %q: no mkif chunk", m.Name)
	}
	if err := m.back.SetFeatherFalloff(falloff); err != nil {
		return err
	}
	m.FeatherFalloff = falloff
	return nil
}

// SetOpacity sets the mask's Opacity (0..1; AE UI shows 0..100%). The
// `ADBE Mask Opacity` leaf is AE-default-elided; setting it materializes the leaf
// in the mask atom group (synthesis-insert). Requires a mask round-tripped
// through Reopen (the atom chunk must exist). Mask Opacity scales how strongly
// the mask reveals/cuts — at 0.5 a reveal shows the layer at half strength.
//
//aep:cap domain=mask tier=stable verify=render-pixel gate=TestMGMaskOpacity_AEShipGate_AE2020,TestMGMaskOpacity_AEShipGate_AE2025 boundary="synthesis-insert;需 Reopen 后调用;0..1 范围拒绝越界" alias="mask opacity,遮罩不透明度,mask transparency,蒙版透明度"
func (m *Mask) SetOpacity(v float64) error {
	if v < 0 || v > 1 {
		return fmt.Errorf("mask %q: opacity %g out of range [0,1]", m.Name, v)
	}
	if m.back == nil {
		return fmt.Errorf("mask %q: no atom chunk (built outside parser?)", m.Name)
	}
	if err := m.back.SetMaskOption("ADBE Mask Opacity", v); err != nil {
		return err
	}
	m.Opacity = v
	return nil
}

// SetFeather sets the mask's Feather softness (X, Y in pixels). The
// `ADBE Mask Feather` leaf is AE-default-elided; setting it materializes the leaf
// (synthesis-insert). Requires a mask round-tripped through Reopen.
//
//aep:cap domain=mask tier=stable verify=roundtrip boundary="synthesis-insert;需 Reopen 后调用;xy 不得为负;无专门 AE gate→round-trip" alias="mask feather,遮罩羽化,feather softness,边缘柔化"
func (m *Mask) SetFeather(xy [2]float64) error {
	if xy[0] < 0 || xy[1] < 0 {
		return fmt.Errorf("mask %q: feather %v must be >= 0", m.Name, xy)
	}
	if m.back == nil {
		return fmt.Errorf("mask %q: no atom chunk (built outside parser?)", m.Name)
	}
	if err := m.back.SetMaskOption("ADBE Mask Feather", xy); err != nil {
		return err
	}
	m.Feather = xy
	return nil
}

// SetExpansion sets the mask's Expansion (AE "Mask Expansion", internally
// `ADBE Mask Offset`) in pixels — positive grows the masked region, negative
// shrinks it. The leaf is AE-default-elided; setting it materializes the leaf
// (synthesis-insert). Requires a mask round-tripped through Reopen.
//
//aep:cap domain=mask tier=stable verify=roundtrip boundary="synthesis-insert;需 Reopen 后调用;正值扩张负值收缩;无专门 AE gate→round-trip" alias="mask expansion,遮罩扩展,mask offset,扩展收缩"
func (m *Mask) SetExpansion(v float64) error {
	if m.back == nil {
		return fmt.Errorf("mask %q: no atom chunk (built outside parser?)", m.Name)
	}
	if err := m.back.SetMaskOption("ADBE Mask Offset", v); err != nil {
		return err
	}
	m.Expansion = v
	return nil
}

// SetClosed toggles whether the (first) path is closed (shph @0x14).
// length-preserving (1 byte). For animated masks this only affects
// the first snapshot; per-keyframe closed flags aren't exposed yet.
//
//aep:cap domain=mask tier=stable verify=roundtrip boundary="length-preserving(1B shph @0x14);动画遮罩仅影响第一帧 snapshot;无专门 AE gate→round-trip" alias="mask closed,路径闭合,closed path,开放路径"
func (m *Mask) SetClosed(v bool) error {
	if m.back == nil {
		return fmt.Errorf("mask %q: no shph chunk", m.Name)
	}
	if err := m.back.SetClosed(v); err != nil {
		return err
	}
	m.Closed = v
	// Also refresh ShphRaw to mirror the new state for callers reading
	// the cached snapshot.
	if d := m.back.ShphData(); d != nil {
		m.ShphRaw = append([]byte(nil), d...)
	}
	return nil
}
