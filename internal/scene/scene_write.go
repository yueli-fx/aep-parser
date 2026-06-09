package scene

import (
	"fmt"
	"io"
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
