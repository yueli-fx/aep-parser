package scene

import (
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"
)

// WriteAEP serializes the (possibly mutated) project back to RIFX binary
// form. Sizes are recomputed from the current chunk data, so mutations
// such as Footage.SetPath that change byte lengths are handled correctly.
//
// This is best-effort write-back. The library only understands a small
// subset of the .aep format; chunks we don't know about pass through
// byte-for-byte. If After Effects rejects the output, file a sample.
//
// The scene-graph → chunk sync (shape layers / render queue / guides / head
// counters) runs inside the writer back-ref (serializer stage) before the
// RIFX tree is written, so this thin scene method stays serializer-free.
func (p *Project) WriteAEP(w io.Writer) error {
	if p.back == nil {
		return fmt.Errorf("aep: project has no underlying RIFX tree (was it built from FromReader?)")
	}
	return p.back.WriteAEP(p, w)
}

// SetBitsPerChannel writes the project's color depth (8 / 16 / 32 bpc)
// to BOTH the nhed @0x0F and nnhd @0x18 header bytes. AE stores the
// enum redundantly; we keep both in sync.
//
// Accepts the existing `BPC8` / `BPC16` / `BPC32` constants. Other
// values are written verbatim (in case AE introduces e.g. half-float
// later) but produce a less obvious AE UI state.
//
// length-preserving (2 bytes total).
func (p *Project) SetBitsPerChannel(bpc BitsPerChannel) error {
	if p.back == nil {
		return fmt.Errorf("project: header chunks missing (built outside parser?)")
	}
	if err := p.back.SetBitsPerChannel(bpc); err != nil {
		return err
	}
	p.BitsPerChannel = bpc
	return nil
}

// SetPath updates the footage's source path. The change is propagated to
// the underlying RIFX chunks (the alas JSON's "fullpath" field is rewritten
// in-place; a legacy Cpth chunk, if any, is fully replaced). The next call
// to Project.WriteAEP will serialize the new path.
//
// Returns an error if no writable path chunk exists for this footage
// (e.g. solids and placeholders never had one).
// SetSolidColor sets a solid footage item's color (RGB, each channel 0..1 —
// alpha is pinned to 1.0, matching AE). length-preserving: the value lives
// inside the fixed-size opti "Soli" chunk. Every layer using this solid
// changes color, exactly like editing the solid's settings in AE.
//
// Returns an error when the footage is not a solid or the channel values are
// out of range.
//
// Stable — AE 2020 + AE 2025 ship-gate green as a standalone setter on a
// parsed solid (AE reads the new color via SolidSource.color and keeps it
// across its own resave); the byte patch is also the one the ship-gated
// NewSolidLayer create path applies.
func (f *Footage) SetSolidColor(rgb [3]float64) error {
	if !f.IsSolid {
		return fmt.Errorf("footage %d (%q): not a solid", f.ID, f.Name)
	}
	for i, v := range rgb {
		if math.IsNaN(v) || v < 0 || v > 1 {
			return fmt.Errorf("footage %d (%q): color[%d]=%v out of range 0..1", f.ID, f.Name, i, v)
		}
	}
	if f.back == nil {
		return fmt.Errorf("footage %d (%q): built outside the parser (no chunk back-refs)", f.ID, f.Name)
	}
	if err := f.back.SetSolidColor(rgb); err != nil {
		return err
	}
	f.SolidColor = rgb
	return nil
}

// SetSolidSize sets a solid footage item's pixel dimensions (1..30000 each,
// AE's solid ceiling). length-preserving: u16 fields inside the fixed sspc
// chunk. Layers using the solid are not repositioned (same as resizing a
// solid in AE's settings dialog).
//
// Stable — AE 2020 + AE 2025 ship-gate green as a standalone setter on a
// parsed solid (AE reads the new dimensions via FootageItem.width/height and
// keeps them across its own resave), same gate as SetSolidColor.
func (f *Footage) SetSolidSize(width, height int) error {
	if !f.IsSolid {
		return fmt.Errorf("footage %d (%q): not a solid", f.ID, f.Name)
	}
	if width < 1 || width > 30000 || height < 1 || height > 30000 {
		return fmt.Errorf("footage %d (%q): dimensions %dx%d out of range 1..30000", f.ID, f.Name, width, height)
	}
	if f.back == nil {
		return fmt.Errorf("footage %d (%q): built outside the parser (no chunk back-refs)", f.ID, f.Name)
	}
	if err := f.back.SetSolidSize(uint16(width), uint16(height)); err != nil {
		return err
	}
	f.Width, f.Height = uint16(width), uint16(height)
	return nil
}

func (f *Footage) SetPath(newPath string) error {
	if f.back == nil {
		return fmt.Errorf("footage %d (%q): no path chunks present (solid/placeholder?)", f.ID, f.Name)
	}
	if err := f.back.SetPath(newPath); err != nil {
		return err
	}
	f.Path = newPath
	if base := filepath.Base(strings.ReplaceAll(newPath, `\`, `/`)); base != "" && base != "." {
		f.Name = base
	}
	return nil
}
