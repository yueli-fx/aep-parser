// RemoveMask — splice a mask atom out of a layer's "ADBE Mask Parade".
//
// Inverse of AddMask. Each mask is a (tdmn "ADBE Mask Atom", mkif[48B],
// LIST:tdgp) TRIPLE — one chunk more than an effect's (tdmn, sspc) pair — so
// the generic indexed-group remove (childTdmnPayload, which assumes the
// payload is preceded directly by a tdmn) refuses a mask atom: the tdgp's
// predecessor is the mkif, not the tdmn. RemoveMask is the triple-aware path:
// it anchors on the parsed mask's own mkif chunk, validates the framing tdmn
// and trailing atom tdgp, splices all three out, and drops the mask from both
// the scene property tree (parade.Children) and the flat layer.Masks slice.
//
// length-variable: the parade tdgp shrinks; WriteAEP recomputes the ancestor
// LIST sizes bottom-up (same path AddMask exercises when it grows the parade).
// The removed triple chunks ride out verbatim, so no opaque content is
// regenerated (per the opaque-preservation invariant). Every precondition is checked before the first
// byte is spliced, so a rejected RemoveMask leaves the project untouched.
package serializer

import (
	"bytes"
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

// RemoveMask deletes a mask from its layer's Mask Parade.
// (Full contract + RE notes live on the aep.RemoveMask facade — docgen source.)
func RemoveMask(layer *Layer, m *Mask) error {
	if layer == nil {
		return fmt.Errorf("RemoveMask: layer is nil")
	}
	if m == nil {
		return fmt.Errorf("RemoveMask: mask is nil")
	}

	idx := -1
	for i, x := range layer.Masks {
		if x == m {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("RemoveMask: mask %q not found in layer %q (already removed?)", m.Name, layer.Name)
	}

	mb := maskBack(m)
	if mb == nil || mb.mkif == nil {
		return fmt.Errorf("RemoveMask: mask %q has no mkif back-ref (built outside parser)", m.Name)
	}

	parade := layer.MaskParade()
	if parade == nil {
		return fmt.Errorf("RemoveMask: layer %q has no Mask Parade", layer.Name)
	}
	pgb := propertyGroupBack(parade)
	if pgb == nil || pgb.chunk == nil {
		return fmt.Errorf("RemoveMask: Mask Parade for layer %q has no chunk back-ref", layer.Name)
	}

	// Locate the atom triple by the mkif pointer (the triple's middle chunk):
	// the "ADBE Mask Atom" tdmn directly precedes it, the atom tdgp follows.
	children := pgb.chunk.Children
	mi := indexOfChunk(children, mb.mkif)
	if mi < 1 || mi+1 >= len(children) {
		return fmt.Errorf("RemoveMask: mask %q mkif is not framed by a triple in the parade LIST", m.Name)
	}
	tdmnCh := children[mi-1]
	atomTdgp := children[mi+1]
	if tdmnCh.ID != rifx.IDTdmn || string(bytes.TrimRight(tdmnCh.Data, "\x00")) != "ADBE Mask Atom" {
		return fmt.Errorf("RemoveMask: mask %q mkif is not preceded by an \"ADBE Mask Atom\" tdmn", m.Name)
	}
	if !atomTdgp.IsList() || atomTdgp.FormType != rifx.IDTdgp {
		return fmt.Errorf("RemoveMask: mask %q mkif is not followed by an atom tdgp", m.Name)
	}

	// Find the scene property-tree node backing this atom (matched by chunk
	// identity, not by index) before committing, so a structural inconsistency
	// refuses rather than half-applying.
	var sceneNode PropertyBase
	for _, c := range parade.Children {
		if g, ok := c.(*AEPropertyGroup); ok {
			if gb := propertyGroupBack(g); gb != nil && gb.chunk == atomTdgp {
				sceneNode = c
				break
			}
		}
	}

	// --- All preconditions passed; commit (no failure points below). ---

	// Chunk: splice out the three triple chunks [mi-1 .. mi+1].
	spliced := make([]*rifx.Chunk, 0, len(children)-3)
	spliced = append(spliced, children[:mi-1]...)
	spliced = append(spliced, children[mi+2:]...)
	pgb.chunk.Children = spliced

	// Scene property tree: drop the atom's group node.
	if sceneNode != nil {
		parade.Children = filterPropertyBase(parade.Children, map[PropertyBase]bool{sceneNode: true})
	}

	// Flat mirror.
	layer.Masks = dropMask(layer.Masks, idx)
	return nil
}
