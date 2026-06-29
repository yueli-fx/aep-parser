// DuplicateMask — clone a mask atom in place within a layer's "ADBE Mask
// Parade", mirroring AE's PropertyBase.duplicate() on a mask.
//
// Like RemoveMask, this is triple-aware: a mask is a (tdmn "ADBE Mask Atom",
// mkif[48B], LIST:tdgp) TRIPLE, so the generic DuplicatePropertyGroup refuses
// it (its (tdmn, payload) pair detection sees the mkif before the tdgp).
// DuplicateMask anchors on the source mask's mkif, deep-clones all three chunks
// (opaque content rides along verbatim, per the opaque-preservation invariant), bumps only the clone's
// internal mask index (mkif @0x08) to max+1 so it stays unique, splices the
// clone triple in immediately after the source atom tdgp, and re-parses it into
// a back-ref-correct *Mask inserted right after the source in layer.Masks.
//
// length-variable: the parade tdgp grows; WriteAEP recomputes ancestor LIST
// sizes (same path AddMask exercises). Atomic: every precondition is checked
// before the first byte is spliced, and any parser warning from the clone
// re-parse rolls the whole mutation back.
package serializer

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// DuplicateMask inserts a copy of mask m immediately after it in the layer's
// Mask Parade and returns the clone.
// (Full contract + RE notes live on the aep.DuplicateMask facade — docgen source.)
func DuplicateMask(layer *Layer, m *Mask) (*Mask, error) {
	if layer == nil {
		return nil, fmt.Errorf("DuplicateMask: layer is nil")
	}
	if m == nil {
		return nil, fmt.Errorf("DuplicateMask: mask is nil")
	}

	idx := -1
	for i, x := range layer.Masks {
		if x == m {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, fmt.Errorf("DuplicateMask: mask %q not found in layer %q (already removed?)", m.Name, layer.Name)
	}

	mb := maskBack(m)
	if mb == nil || mb.mkif == nil {
		return nil, fmt.Errorf("DuplicateMask: mask %q has no mkif back-ref (built outside parser)", m.Name)
	}

	parade := layer.MaskParade()
	if parade == nil {
		return nil, fmt.Errorf("DuplicateMask: layer %q has no Mask Parade", layer.Name)
	}
	pgb := propertyGroupBack(parade)
	if pgb == nil || pgb.chunk == nil {
		return nil, fmt.Errorf("DuplicateMask: Mask Parade for layer %q has no chunk back-ref", layer.Name)
	}

	children := pgb.chunk.Children
	mi := indexOfChunk(children, mb.mkif)
	if mi < 1 || mi+1 >= len(children) {
		return nil, fmt.Errorf("DuplicateMask: mask %q mkif is not framed by a triple in the parade LIST", m.Name)
	}
	tdmnCh := children[mi-1]
	atomTdgp := children[mi+1]
	if tdmnCh.ID != rifx.IDTdmn || string(bytes.TrimRight(tdmnCh.Data, "\x00")) != "ADBE Mask Atom" {
		return nil, fmt.Errorf("DuplicateMask: mask %q mkif is not preceded by an \"ADBE Mask Atom\" tdmn", m.Name)
	}
	if !atomTdgp.IsList() || atomTdgp.FormType != rifx.IDTdgp {
		return nil, fmt.Errorf("DuplicateMask: mask %q mkif is not followed by an atom tdgp", m.Name)
	}

	// Find the source's scene node (matched by chunk identity) before
	// committing, so the clone can be inserted right after it.
	var srcNode *AEPropertyGroup
	for _, c := range parade.Children {
		if g, ok := c.(*AEPropertyGroup); ok {
			if gb := propertyGroupBack(g); gb != nil && gb.chunk == atomTdgp {
				srcNode = g
				break
			}
		}
	}

	// Allocate a fresh internal mask index (monotonic; AE tolerates gaps).
	index := uint32(1)
	for _, x := range layer.Masks {
		if x.Index >= index {
			index = x.Index + 1
		}
	}

	// Deep-clone the triple; bump only the clone's mkif internal index.
	tdmnClone := deepCloneChunk(tdmnCh)
	mkifClone := deepCloneChunk(mb.mkif)
	tdgpClone := deepCloneChunk(atomTdgp)
	if len(mkifClone.Data) >= 0x0C {
		binary.BigEndian.PutUint32(mkifClone.Data[0x08:0x0C], index)
	}

	// Snapshot for atomic rollback.
	oldChunkChildren := append([]*rifx.Chunk(nil), children...)
	oldSceneChildren := append([]PropertyBase(nil), parade.Children...)
	oldMasks := append([]*Mask(nil), layer.Masks...)
	oldWarningsLen := warningsLen(layer)
	rollback := func() {
		pgb.chunk.Children = oldChunkChildren
		parade.Children = oldSceneChildren
		layer.Masks = oldMasks
		rollbackWarnings(layer, oldWarningsLen)
	}

	// Chunk: splice the clone triple in just after the source atom tdgp.
	spliced := make([]*rifx.Chunk, 0, len(children)+3)
	spliced = append(spliced, children[:mi+2]...)
	spliced = append(spliced, tdmnClone, mkifClone, tdgpClone)
	spliced = append(spliced, children[mi+2:]...)
	pgb.chunk.Children = spliced

	// Scene property tree: insert a stand-in group node after the source.
	cloneNode := &AEPropertyGroup{MatchName: "ADBE Mask Atom", Name: m.Name}
	scene.SetPropertyGroupParent(cloneNode, parade)
	scene.SetPropertyGroupBack(cloneNode, &propertyGroupBackrefs{chunk: tdgpClone})
	if srcNode != nil {
		insertChildAfter(parade, srcNode, cloneNode)
	} else {
		parade.Children = append(parade.Children, cloneNode)
	}

	// Flat mirror: re-parse the clone triple so the new *Mask's back-refs point
	// at the CLONE's chunks (never aliased to the source).
	var newMask *Mask
	if comp := scene.LayerComp(layer); comp != nil && scene.CompositionProj(comp) != nil {
		ctx := newParseCtxFPS(comp.TickRate, comp.FrameRate, comp.Name, &scene.CompositionProj(comp).Warnings)
		tmpParade := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{tdmnClone, mkifClone, tdgpClone}}
		tmpLayr := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
			makeTdmn(MatchNameGroupMaskParade), tmpParade,
		}}
		masks := parseMasks(tmpLayr, ctx)
		if len(masks) != 1 {
			rollback()
			return nil, fmt.Errorf("DuplicateMask: clone re-parse produced %d masks (want 1)", len(masks))
		}
		newMask = masks[0]
		layer.Masks = insertMaskAt(layer.Masks, idx+1, newMask)
	}

	if newWarn := newWarningsSince(layer, oldWarningsLen); len(newWarn) > 0 {
		rollback()
		return nil, fmt.Errorf("DuplicateMask: produced %d parser warning(s), rolled back: %v", len(newWarn), newWarn)
	}
	return newMask, nil
}
