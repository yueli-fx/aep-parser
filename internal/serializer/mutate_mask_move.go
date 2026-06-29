// MoveMask — reorder a mask within a layer's "ADBE Mask Parade", mirroring AE's
// PropertyBase.moveTo() on a mask.
//
// Triple-aware, like RemoveMask / DuplicateMask: each mask is a (tdmn "ADBE
// Mask Atom", mkif[48B], LIST:tdgp) TRIPLE, so the generic MovePropertyGroup
// (which rebuilds the group LIST from (tdmn, payload) pairs) can't reorder it.
// MoveMask locates every mask's triple by its mkif pointer, confirms they form
// one contiguous run (parade header prefix + N triples + Group End suffix),
// then re-emits the run in the target order — the SAME chunk pointers, so all
// opaque content rides along unchanged (per the opaque-preservation invariant) — and applies the same
// permutation to the scene property tree and the flat layer.Masks slice.
//
// No chunk is created or destroyed, so no LIST size changes and no re-parse is
// needed. Every precondition is checked before the first reorder, so a rejected
// MoveMask leaves the project untouched.
package serializer

import (
	"bytes"
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

// maskTriple holds one mask atom's three consecutive parade chunks plus the
// index of its leading tdmn / trailing tdgp in the parade LIST.
type maskTriple struct {
	tdmn, mkif, tdgp *rifx.Chunk
	lo, hi           int
}

// MoveMask moves mask m to position toIndex (0-based) among the layer's masks,
// other masks keeping their relative order.
// (Full contract + RE notes live on the aep.MoveMask facade — docgen source.)
func MoveMask(layer *Layer, m *Mask, toIndex int) error {
	if layer == nil {
		return fmt.Errorf("MoveMask: layer is nil")
	}
	if m == nil {
		return fmt.Errorf("MoveMask: mask is nil")
	}
	n := len(layer.Masks)
	cur := -1
	for i, x := range layer.Masks {
		if x == m {
			cur = i
			break
		}
	}
	if cur < 0 {
		return fmt.Errorf("MoveMask: mask %q not found in layer %q (already removed?)", m.Name, layer.Name)
	}
	if toIndex < 0 || toIndex >= n {
		return fmt.Errorf("MoveMask: toIndex %d out of range (have %d masks)", toIndex, n)
	}
	if toIndex == cur {
		return nil
	}

	parade := layer.MaskParade()
	if parade == nil {
		return fmt.Errorf("MoveMask: layer %q has no Mask Parade", layer.Name)
	}
	pgb := propertyGroupBack(parade)
	if pgb == nil || pgb.chunk == nil {
		return fmt.Errorf("MoveMask: Mask Parade for layer %q has no chunk back-ref", layer.Name)
	}
	if len(parade.Children) != n {
		return fmt.Errorf("MoveMask: Mask Parade scene children (%d) != mask count (%d)", len(parade.Children), n)
	}
	children := pgb.chunk.Children

	// Locate every mask's triple by its mkif pointer.
	triples := make([]maskTriple, n)
	first, last := len(children), -1
	for i, x := range layer.Masks {
		mb := maskBack(x)
		if mb == nil || mb.mkif == nil {
			return fmt.Errorf("MoveMask: mask %q has no mkif back-ref (built outside parser)", x.Name)
		}
		mi := indexOfChunk(children, mb.mkif)
		if mi < 1 || mi+1 >= len(children) {
			return fmt.Errorf("MoveMask: mask %q mkif is not framed by a triple in the parade LIST", x.Name)
		}
		tdmn := children[mi-1]
		tdgp := children[mi+1]
		if tdmn.ID != rifx.IDTdmn || string(bytes.TrimRight(tdmn.Data, "\x00")) != "ADBE Mask Atom" {
			return fmt.Errorf("MoveMask: mask %q mkif is not preceded by an \"ADBE Mask Atom\" tdmn", x.Name)
		}
		if !tdgp.IsList() || tdgp.FormType != rifx.IDTdgp {
			return fmt.Errorf("MoveMask: mask %q mkif is not followed by an atom tdgp", x.Name)
		}
		triples[i] = maskTriple{tdmn, mb.mkif, tdgp, mi - 1, mi + 1}
		if mi-1 < first {
			first = mi - 1
		}
		if mi+1 > last {
			last = mi + 1
		}
	}
	// The triples must occupy one contiguous run so the parade header prefix and
	// the Group End suffix stay put around them.
	if last-first+1 != 3*n {
		return fmt.Errorf("MoveMask: parade triples are not contiguous (span %d for %d masks)", last-first+1, n)
	}

	order := moveIndexPermutation(n, cur, toIndex)

	// --- All preconditions passed; commit (no failure points below). ---

	prefix := append([]*rifx.Chunk(nil), children[:first]...)
	suffix := append([]*rifx.Chunk(nil), children[last+1:]...)
	rebuilt := make([]*rifx.Chunk, 0, len(prefix)+3*n+len(suffix))
	rebuilt = append(rebuilt, prefix...)
	for _, oldIdx := range order {
		tr := triples[oldIdx]
		rebuilt = append(rebuilt, tr.tdmn, tr.mkif, tr.tdgp)
	}
	rebuilt = append(rebuilt, suffix...)
	pgb.chunk.Children = rebuilt

	// Scene: reorder parade.Children + flat layer.Masks by the same permutation.
	newSceneChildren := make([]PropertyBase, n)
	for newIdx, oldIdx := range order {
		newSceneChildren[newIdx] = parade.Children[oldIdx]
	}
	parade.Children = newSceneChildren
	layer.Masks = reorderMasks(layer.Masks, order)
	return nil
}
