package scene

import "fmt"

// Mask setters — length-preserving mkif / shph byte writes. Each
// setter delegates byte-patch work to maskBackrefs (MaskWriter), then
// syncs the matching Go field. WriteAEP serializes the change.
//
// Note: vertex / path mutations are NOT supported here (requires
// rewriting variable-length shap kfl streams).

// @summary     Set a mask's blend mode
// @param       mode  the new mask mode
// @domain      mask
// @stability   stable
// @verify      ae-accept
// @gate        TestMaskOpts_AEShipGate_AE2020,TestMaskOpts_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving 4-byte write to mkif offset 0x04
// @alias       mask mode,遮罩模式,blend mode,add subtract intersect
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

// @summary     Set a mask's Inverted flag
// @param       v  the new inverted state
// @domain      mask
// @stability   stable
// @verify      ae-accept
// @gate        TestMaskOpts_AEShipGate_AE2020,TestMaskOpts_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving 1-byte write to mkif offset 0x00
// @alias       mask inverted,遮罩反转,invert mask
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

// @summary     Set a mask's timeline label color
// @description AE always keeps the alpha byte at 0xFF; this setter does
//   not touch it.
// @param       rgb  the new label color (red, green, blue)
// @domain      mask
// @stability   stable
// @verify      ae-accept
// @gate        TestMaskOpts_AEShipGate_AE2020,TestMaskOpts_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving 3-byte write to mkif offsets 0x2D-0x2F
// @alias       mask color,遮罩颜色,label color,时间线标签色
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

// @summary     Set a mask's lock flag
// @description When locked, AE refuses edits to the mask in the timeline UI;
//   the underlying bytes are still mutable through this library.
// @param       v  the new locked state
// @domain      mask
// @stability   stable
// @verify      ae-accept
// @gate        TestMaskOpts_AEShipGate_AE2020,TestMaskOpts_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving 1-byte write to mkif offset 0x01
// @alias       mask locked,遮罩锁定,lock mask
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

// @summary     Set a mask's per-mask motion-blur override
// @param       mode  the new motion-blur mode (SameAsLayer, On, or Off)
// @domain      mask
// @stability   stable
// @verify      ae-accept
// @gate        TestMaskOpts_AEShipGate_AE2020,TestMaskOpts_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving 1-byte write to mkif offset 0x02
// @alias       mask motion blur,遮罩运动模糊,motion blur override
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

// @summary     Set a mask's feather-falloff curve
// @param       falloff  the new falloff curve (Smooth or Linear)
// @domain      mask
// @stability   stable
// @verify      ae-accept
// @gate        TestMaskFeatherFalloff_AEShipGate_AE2020,TestMaskFeatherFalloff_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving 1-byte write to mkif offset 0x03; verified
//   via AE DOM readback of the maskFeatherFalloff enum (not a render check)
// @alias       mask feather falloff,遮罩羽化衰减,feather falloff curve,smooth linear feather
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

// @summary     Set a mask's opacity
// @description Mask Opacity scales how strongly the mask reveals or cuts —
//   at 0.5 a reveal shows the layer at half strength. The `ADBE Mask
//   Opacity` leaf is elided by AE when at its default, so setting it
//   materializes the leaf in the mask's atom group.
// @param       v  the new opacity, normalized 0..1 (AE's UI shows 0..100%)
// @domain      mask
// @stability   stable
// @verify      render-pixel
// @gate        TestMGMaskOpacity_AEShipGate_AE2020,TestMGMaskOpacity_AEShipGate_AE2025
// @since       AE2020
// @boundary    synthesis-insert; requires a mask that has been round-tripped
//   through Reopen so the atom chunk exists; rejects values outside 0..1
// @alias       mask opacity,遮罩不透明度,mask transparency,蒙版透明度
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

// @summary     Set a mask's feather softness
// @description The `ADBE Mask Feather` leaf is elided by AE when at its
//   default, so setting it materializes the leaf in the mask's atom group.
// @param       xy  the new feather softness in pixels (x, y); must be >= 0
// @domain      mask
// @stability   stable
// @verify      ae-accept
// @gate        TestMaskOpts_AEShipGate_AE2020,TestMaskOpts_AEShipGate_AE2025
// @since       AE2020
// @boundary    synthesis-insert; requires a mask that has been round-tripped
//   through Reopen so the atom chunk exists; rejects negative components
// @alias       mask feather,遮罩羽化,feather softness,边缘柔化
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

// @summary     Set a mask's expansion
// @description Internally this is the `ADBE Mask Offset` property (AE's UI
//   calls it "Mask Expansion"). Positive values grow the masked region,
//   negative values shrink it. The leaf is elided by AE when at its
//   default, so setting it materializes the leaf in the mask's atom group.
// @param       v  the new expansion in pixels (positive grows, negative shrinks)
// @domain      mask
// @stability   stable
// @verify      ae-accept
// @gate        TestMaskOpts_AEShipGate_AE2020,TestMaskOpts_AEShipGate_AE2025
// @since       AE2020
// @boundary    synthesis-insert; requires a mask that has been round-tripped
//   through Reopen so the atom chunk exists
// @alias       mask expansion,遮罩扩展,mask offset,扩展收缩
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

// @summary     Set whether a mask's path is closed
// @description For animated masks this only affects the first snapshot;
//   per-keyframe closed flags aren't exposed yet.
// @param       v  the new closed state
// @domain      mask
// @stability   stable
// @verify      ae-accept
// @gate        TestMaskOpts_AEShipGate_AE2020,TestMaskOpts_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving 1-byte write to shph offset 0x14
// @alias       mask closed,路径闭合,closed path,开放路径
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
