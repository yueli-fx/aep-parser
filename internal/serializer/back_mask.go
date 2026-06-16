package serializer

import (
	"encoding/binary"
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// maskBack returns the concrete back-refs behind a Mask's writer interface for
// serializer-stage raw chunk access. Returns nil when the mask was built
// outside the parser. Free function (the receiver is a scene type post split).
func maskBack(m *Mask) *maskBackrefs {
	if mb, ok := scene.MaskBack(m).(*maskBackrefs); ok {
		return mb
	}
	return nil
}

// maskBackrefs holds the rifx.Chunk references that power a parsed Mask's
// length-preserving setters (SetMode / SetInverted / SetColor / SetLocked /
// SetMaskMotionBlur / SetClosed). Lives in a separate shard so the Mask scene
// type stays chunk-free.
//
// Lifecycle:
//   - Populated by parseMasks when a Mask is built from a parsed .aep file.
//   - nil for masks built outside the parser; every Set* refuses with an
//     error when the chunk it needs is missing.
type maskBackrefs struct {
	// maskName is stored for error-message context (mirrors Mask.Name at
	// parse time; not used for byte writes).
	maskName string

	// mkif is the mask info chunk (Mode @0x04 / Inverted @0x00 / Locked @0x01
	// / MotionBlur @0x02 / Color @0x2D-0x2F).
	mkif *rifx.Chunk
	// shph is the first shph path-header chunk (Closed @0x14 write — only
	// meaningful for static masks).
	shph *rifx.Chunk
	// atomTdgp is the mask atom's inner property group (holds Mask Shape +
	// the Feather/Opacity/Offset leaves). Needed by SetMaskOption to overwrite
	// or splice those default-elided option leaves.
	atomTdgp *rifx.Chunk
}

var _ MaskWriter = (*maskBackrefs)(nil)

// ShphData exposes the live shph chunk bytes so the scene SetClosed accessor
// refreshes Mask.ShphRaw after a back-side write. Returns nil for masks built
// outside the parser / static masks without a shph reference.
func (b *maskBackrefs) ShphData() []byte {
	if b == nil || b.shph == nil {
		return nil
	}
	return b.shph.Data
}

func (b *maskBackrefs) SetMode(mode MaskMode) error {
	if b.mkif == nil {
		return fmt.Errorf("mask %q: no mkif chunk (built outside parser?)", b.maskName)
	}
	if len(b.mkif.Data) < 0x08 {
		return fmt.Errorf("mask %q: mkif too short for Mode write (len=%d)", b.maskName, len(b.mkif.Data))
	}
	binary.BigEndian.PutUint32(b.mkif.Data[0x04:0x08], uint32(mode))
	return nil
}

func (b *maskBackrefs) SetInverted(v bool) error {
	if b.mkif == nil {
		return fmt.Errorf("mask %q: no mkif chunk", b.maskName)
	}
	if len(b.mkif.Data) < 1 {
		return fmt.Errorf("mask %q: mkif empty", b.maskName)
	}
	if v {
		b.mkif.Data[0x00] = 1
	} else {
		b.mkif.Data[0x00] = 0
	}
	return nil
}

func (b *maskBackrefs) SetColor(rgb [3]uint8) error {
	if b.mkif == nil {
		return fmt.Errorf("mask %q: no mkif chunk", b.maskName)
	}
	if len(b.mkif.Data) < 0x30 {
		return fmt.Errorf("mask %q: mkif too short for Color write (len=%d)", b.maskName, len(b.mkif.Data))
	}
	b.mkif.Data[0x2D] = rgb[0]
	b.mkif.Data[0x2E] = rgb[1]
	b.mkif.Data[0x2F] = rgb[2]
	return nil
}

func (b *maskBackrefs) SetLocked(v bool) error {
	if b.mkif == nil {
		return fmt.Errorf("mask %q: no mkif chunk", b.maskName)
	}
	if len(b.mkif.Data) < 2 {
		return fmt.Errorf("mask %q: mkif too short for Locked write (len=%d)", b.maskName, len(b.mkif.Data))
	}
	if v {
		b.mkif.Data[0x01] = 1
	} else {
		b.mkif.Data[0x01] = 0
	}
	return nil
}

func (b *maskBackrefs) SetMaskMotionBlur(mode MaskMotionBlurMode) error {
	if b.mkif == nil {
		return fmt.Errorf("mask %q: no mkif chunk", b.maskName)
	}
	if len(b.mkif.Data) < 3 {
		return fmt.Errorf("mask %q: mkif too short for MaskMotionBlur write (len=%d)", b.maskName, len(b.mkif.Data))
	}
	b.mkif.Data[0x02] = byte(mode)
	return nil
}

func (b *maskBackrefs) SetFeatherFalloff(falloff MaskFeatherFalloff) error {
	if b.mkif == nil {
		return fmt.Errorf("mask %q: no mkif chunk", b.maskName)
	}
	if len(b.mkif.Data) < 4 {
		return fmt.Errorf("mask %q: mkif too short for FeatherFalloff write (len=%d)", b.maskName, len(b.mkif.Data))
	}
	b.mkif.Data[0x03] = byte(falloff)
	return nil
}

func (b *maskBackrefs) SetClosed(v bool) error {
	if b.shph == nil {
		return fmt.Errorf("mask %q: no shph chunk", b.maskName)
	}
	if len(b.shph.Data) < 4 {
		return fmt.Errorf("mask %q: shph too short for Closed write (len=%d)", b.maskName, len(b.shph.Data))
	}
	// Open flag = shph[3] bit3 (0x01 closed / 0x09 open on every AE-saved
	// mask; the former @0x14 target is a constant 0x01, writing it was a
	// semantic no-op AE ignored).
	if v {
		b.shph.Data[3] &^= 0x08
	} else {
		b.shph.Data[3] |= 0x08
	}
	return nil
}
