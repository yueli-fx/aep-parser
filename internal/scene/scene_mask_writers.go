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

// SetOpacity sets the mask's Opacity (0..1; AE UI shows 0..100%). The
// `ADBE Mask Opacity` leaf is AE-default-elided; setting it materializes the leaf
// in the mask atom group (synthesis-insert). Requires a mask round-tripped
// through Reopen (the atom chunk must exist). Mask Opacity scales how strongly
// the mask reveals/cuts — at 0.5 a reveal shows the layer at half strength.
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
