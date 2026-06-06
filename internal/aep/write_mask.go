package aep

import (
	"encoding/binary"
	"fmt"
)

// Mask setters — length-preserving mkif / shph byte writes. Each
// setter validates the underlying chunk is long enough for the target
// offset, mutates the bytes in place, then syncs the matching Go
// field. WriteAEP serializes the change.
//
// Note: vertex / path mutations are NOT supported here (requires
// rewriting variable-length shap kfl streams).

// SetMode writes a new mask Mode enum (uint32 BE) to mkif @0x04.
// length-preserving (4 bytes).
func (m *Mask) SetMode(mode MaskMode) error {
	if m.back == nil || m.back.mkif == nil {
		return fmt.Errorf("mask %q: no mkif chunk (built outside parser?)", m.Name)
	}
	if len(m.back.mkif.Data) < 0x08 {
		return fmt.Errorf("mask %q: mkif too short for Mode write (len=%d)", m.Name, len(m.back.mkif.Data))
	}
	binary.BigEndian.PutUint32(m.back.mkif.Data[0x04:0x08], uint32(mode))
	m.Mode = mode
	return nil
}

// SetInverted toggles the mask Inverted flag (mkif @0x00).
// length-preserving (1 byte).
func (m *Mask) SetInverted(v bool) error {
	if m.back == nil || m.back.mkif == nil {
		return fmt.Errorf("mask %q: no mkif chunk", m.Name)
	}
	if len(m.back.mkif.Data) < 1 {
		return fmt.Errorf("mask %q: mkif empty", m.Name)
	}
	if v {
		m.back.mkif.Data[0x00] = 1
	} else {
		m.back.mkif.Data[0x00] = 0
	}
	m.Inverted = v
	return nil
}

// SetColor writes the mask timeline label color RGB to mkif
// @0x2D/@0x2E/@0x2F. AE always keeps alpha (0x2C) at 0xFF;
// this setter does not touch it.
// length-preserving (3 bytes).
func (m *Mask) SetColor(rgb [3]uint8) error {
	if m.back == nil || m.back.mkif == nil {
		return fmt.Errorf("mask %q: no mkif chunk", m.Name)
	}
	if len(m.back.mkif.Data) < 0x30 {
		return fmt.Errorf("mask %q: mkif too short for Color write (len=%d)", m.Name, len(m.back.mkif.Data))
	}
	m.back.mkif.Data[0x2D] = rgb[0]
	m.back.mkif.Data[0x2E] = rgb[1]
	m.back.mkif.Data[0x2F] = rgb[2]
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
	if m.back == nil || m.back.mkif == nil {
		return fmt.Errorf("mask %q: no mkif chunk", m.Name)
	}
	if len(m.back.mkif.Data) < 2 {
		return fmt.Errorf("mask %q: mkif too short for Locked write (len=%d)", m.Name, len(m.back.mkif.Data))
	}
	if v {
		m.back.mkif.Data[0x01] = 1
	} else {
		m.back.mkif.Data[0x01] = 0
	}
	m.Locked = v
	return nil
}

// SetMaskMotionBlur writes the per-mask motion-blur override at mkif
// @0x02. Valid values: MaskMotionBlurSameAsLayer (0),
// MaskMotionBlurOn (2), MaskMotionBlurOff (3).
// length-preserving (1 byte).
func (m *Mask) SetMaskMotionBlur(mode MaskMotionBlurMode) error {
	if m.back == nil || m.back.mkif == nil {
		return fmt.Errorf("mask %q: no mkif chunk", m.Name)
	}
	if len(m.back.mkif.Data) < 3 {
		return fmt.Errorf("mask %q: mkif too short for MaskMotionBlur write (len=%d)", m.Name, len(m.back.mkif.Data))
	}
	m.back.mkif.Data[0x02] = byte(mode)
	m.MotionBlur = mode
	return nil
}

// SetClosed toggles whether the (first) path is closed (shph @0x14).
// length-preserving (1 byte). For animated masks this only affects
// the first snapshot; per-keyframe closed flags aren't exposed yet.
func (m *Mask) SetClosed(v bool) error {
	if m.back == nil || m.back.shph == nil {
		return fmt.Errorf("mask %q: no shph chunk", m.Name)
	}
	if len(m.back.shph.Data) <= 0x14 {
		return fmt.Errorf("mask %q: shph too short for Closed write (len=%d)", m.Name, len(m.back.shph.Data))
	}
	if v {
		m.back.shph.Data[0x14] = 1
	} else {
		m.back.shph.Data[0x14] = 0
	}
	m.Closed = v
	// Also refresh ShphRaw to mirror the new state for callers reading
	// the cached snapshot.
	m.ShphRaw = append([]byte(nil), m.back.shph.Data...)
	return nil
}
