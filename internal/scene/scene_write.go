package scene

import (
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"
)

// @summary     Serialize the project back to RIFX binary form
// @description Sizes are recomputed from the current chunk data, so
//   mutations such as Footage.SetPath that change byte lengths are handled
//   correctly. This is best-effort write-back: only a small subset of the
//   .aep format is understood, and chunks that aren't recognized pass
//   through byte-for-byte. The scene-graph to chunk sync (shape layers /
//   render queue / guides / head counters) runs in the writer back-ref
//   before the RIFX tree is written, so this method itself never touches
//   chunk bytes directly.
// @param       w  the destination to write the binary project to
// @domain      io
// @stability   stable
// @verify      roundtrip
// @since       AE2020
// @boundary    this is the core write-back path, indirectly covered by every
//   structural ship gate; no single AE gate exercises it in isolation
// @alias       write aep,写回,serialize,binary write,RIFX write,WriteAEP
func (p *Project) WriteAEP(w io.Writer) error {
	if p.back == nil {
		return fmt.Errorf("aep: project has no underlying RIFX tree (was it built from FromReader?)")
	}
	return p.back.WriteAEP(p, w)
}

// @summary     Set the project's color bit depth
// @description Writes to BOTH the nhed offset 0x0F and nnhd offset 0x18
//   header bytes — AE stores the enum redundantly, so both copies are kept
//   in sync. Accepts the BPC8 / BPC16 / BPC32 constants; other values are
//   written verbatim (in case AE introduces e.g. half-float later) but
//   produce a less obvious AE UI state.
// @param       bpc  the new bit depth (8 / 16 / 32 bits per channel)
// @domain      project
// @stability   stable
// @verify      ae-accept
// @gate        TestProjectSettings_AEShipGate_AE2020,TestProjectSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving (2 bytes total, 1 in each header); verified
//   by reading app.project.bitsPerChannel back through the AE DOM
// @alias       bits per channel,颜色深度,color depth,bpc,8bpc,16bpc,32bpc
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

// @summary     Set a solid footage item's color
// @description RGB, each channel 0..1 — alpha is pinned to 1.0, matching AE.
//   Every layer using this solid changes color, exactly like editing the
//   solid's settings in AE. The same byte patch is applied by the
//   NewSolidLayer create path.
// @param       rgb  the new color, each channel in range 0..1
// @domain      project
// @stability   stable
// @verify      ae-accept
// @gate        TestSolidSetters_AEShipGate_AE2020,TestSolidSetters_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving — the value lives inside the fixed-size
//   opti "Soli" chunk; verified by reading SolidSource.color back through
//   the AE scripting API
// @alias       solid color,固态层颜色,solid colour,background color,纯色颜色
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

// @summary     Set a solid footage item's pixel dimensions
// @description Layers using the solid are not repositioned — same behavior
//   as resizing a solid through AE's settings dialog. Shares its ship gate
//   with SetSolidColor.
// @param       width   the new width in pixels, in range 1..30000
// @param       height  the new height in pixels, in range 1..30000
// @domain      project
// @stability   stable
// @verify      ae-accept
// @gate        TestSolidSetters_AEShipGate_AE2020,TestSolidSetters_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving — u16 fields inside the fixed sspc chunk,
//   AE's solid ceiling is 30000; verified by reading
//   FootageItem.width/height back through the AE scripting API
// @alias       solid size,固态层尺寸,solid dimensions,solid width,solid height,纯色大小
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

// @summary     Set the footage's source path
// @description The change propagates to the underlying RIFX chunks: the
//   alas JSON's "fullpath" field is rewritten in-place, and a legacy Cpth
//   chunk, if any, is fully replaced. The next call to Project.WriteAEP
//   serializes the new path. Returns an error if no writable path chunk
//   exists for this footage (solids and placeholders never had one).
// @param       newPath  the new source path
// @domain      project
// @stability   stable
// @verify      roundtrip
// @since       AE2020
// @boundary    length-variable (alas JSON fullpath rewrite plus a full Cpth
//   chunk replacement) — relinks the footage path; no dedicated AE gate,
//   verified by round-trip
// @alias       set path,footage path,素材路径,relink footage,replace footage,文件路径
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
