package aep

import (
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// MoveLayer reorders the layer at `from` to position `to` in c.Layers
// (both 0-based). The source layer's entire chunk block — Layr + Ewst
// + leaf followers (adaptive scan to next LIST/EOF, same machinery as
// DeleteLayer / DuplicateLayer) — is spliced out and re-inserted at the
// target slot. After the call, c.Layers[to] == the moved layer, and
// every layer's Layer.Index field is refreshed to match its new slice
// position.
//
// Refuse-cases (Phase 4 conservative):
//
//   - `from` or `to` out of range (note: `to == len(c.Layers)-1` IS in
//     range and means "move to last slot")
//   - comp lacks parsed itemList back-ref
//   - source layer lacks Layr back-ref / corrupted block (Layr formType
//     / Ewst sibling mismatch)
//
// `from == to` is a no-op (returns nil, no state change).
//
// Unlike DeleteLayer / DuplicateLayer, MoveLayer does NOT care about
// layer Type or TrackMatte — pure reorder works for AV / Camera / Light
// / Audio / Shape / Text / matted layers alike.
//
// Atomic mutation (Inv-10 / Inv-11): snapshot pre-call itemList.Children
// + c.Layers + each layer's Index + Warnings count; on any new parser
// warning during the call, roll all of them back. No re-parse and no
// new chunks created, so the warnings path is defensive (mirrors V2.1
// pattern for symmetry).
//
// Stable: no Alpha gate — AE behavior is known (layer order = order of
// Layr LISTs in itemList.Children, same model that DeleteLayer and
// DuplicateLayer already exercise and ship-gate via 14/14 PASS runs
// across AE 2020 + AE 2025). Phase 4 ship-gate (`verify_ge_move_layer.jsx`,
// 3 modes × 2 versions = 6 runs) re-validates AE acceptance for the
// reorder path specifically.
func (c *Composition) MoveLayer(from, to int) error {
	// 1. Validate refuse-cases.
	if c.back == nil || c.back.itemList == nil {
		return fmt.Errorf("MoveLayer: comp %q has no itemList back-ref (built outside parser?)", c.Name)
	}
	n := len(c.Layers)
	if from < 0 || from >= n {
		return fmt.Errorf("MoveLayer: from %d out of range (have %d layers)", from, n)
	}
	if to < 0 || to >= n {
		return fmt.Errorf("MoveLayer: to %d out of range (have %d layers)", to, n)
	}
	if from == to {
		return nil
	}

	source := c.Layers[from]
	if source.back == nil || source.back.layrList == nil {
		return fmt.Errorf("MoveLayer: layer %q at idx %d has no Layr chunk back-ref", source.Name, from)
	}

	// 2. Locate source Layr in itemList.Children.
	children := c.back.itemList.Children
	srcLayrIdx := findLayrIndexInItemList(c.back.itemList, source.back.layrList)
	if srcLayrIdx < 0 {
		return fmt.Errorf("MoveLayer: layer %q Layr chunk not found in itemList", source.Name)
	}

	// 3. Defensive structural assertions — FormType and Ewst sibling.
	if !children[srcLayrIdx].IsList() || children[srcLayrIdx].FormType != rifx.IDLayr {
		return fmt.Errorf("MoveLayer: layer %q backref points to non-Layr chunk (FormType=%s)", source.Name, chunkIDString(children[srcLayrIdx].FormType))
	}
	if srcLayrIdx+1 >= len(children) {
		return fmt.Errorf("MoveLayer: layer %q Layr at end of itemList (no Ewst sibling)", source.Name)
	}
	ewstCandidate := children[srcLayrIdx+1]
	if !ewstCandidate.IsList() || ewstCandidate.FormType != rifx.IDEwst {
		return fmt.Errorf("MoveLayer: layer %q expected Ewst sibling after Layr, found %s", source.Name, chunkIDString(ewstCandidate.FormType))
	}

	// 4. Adaptive block end — consume leaf followers until next LIST/EOF
	//    (mirrors DeleteLayer / DuplicateLayer adaptive splice — handles
	//    AE-saved 16-chunk and Go-built 2-chunk forms alike).
	endIdx := srcLayrIdx + 2
	for endIdx < len(children) && !children[endIdx].IsList() {
		endIdx++
	}

	// 5. Snapshot for rollback (V2.1 pattern). Difference vs
	//    DeleteLayer/DuplicateLayer: no nextItemID bump to roll back,
	//    but we DO snapshot per-layer Index since we'll re-assign them.
	oldChildren := append([]*rifx.Chunk(nil), children...)
	oldLayers := append([]*Layer(nil), c.Layers...)
	oldIndexes := make([]int, n)
	for i, l := range c.Layers {
		oldIndexes[i] = l.Index
	}
	oldWarningsLen := 0
	if c.proj != nil {
		oldWarningsLen = len(c.proj.Warnings)
	}

	// 6. Extract the source block (will be re-inserted at the new slot).
	block := append([]*rifx.Chunk(nil), children[srcLayrIdx:endIdx]...)

	// 7. Build cutChildren = children minus the block.
	cutChildren := make([]*rifx.Chunk, 0, len(children)-len(block))
	cutChildren = append(cutChildren, children[:srcLayrIdx]...)
	cutChildren = append(cutChildren, children[endIdx:]...)

	// 8. Build cutLayers = c.Layers minus source.
	cutLayers := make([]*Layer, 0, n-1)
	cutLayers = append(cutLayers, c.Layers[:from]...)
	cutLayers = append(cutLayers, c.Layers[from+1:]...)

	// 9. Determine insertion index in cutChildren.
	var insertIdx int
	if to < len(cutLayers) {
		// Insert BEFORE the Layr block of cutLayers[to].
		target := cutLayers[to]
		if target.back == nil || target.back.layrList == nil {
			return fmt.Errorf("MoveLayer: target layer %q has no Layr backref", target.Name)
		}
		insertIdx = indexOfChunk(cutChildren, target.back.layrList)
		if insertIdx < 0 {
			return fmt.Errorf("MoveLayer: target layer %q Layr not found in cut itemList", target.Name)
		}
	} else {
		// to == len(cutLayers) (i.e., to == n-1, move to last slot).
		// Insert AFTER the last remaining layer's block — scan to end of
		// that block.
		lastLayer := cutLayers[len(cutLayers)-1]
		if lastLayer.back == nil || lastLayer.back.layrList == nil {
			return fmt.Errorf("MoveLayer: last cut layer %q has no Layr backref", lastLayer.Name)
		}
		lastLayrIdx := indexOfChunk(cutChildren, lastLayer.back.layrList)
		if lastLayrIdx < 0 {
			return fmt.Errorf("MoveLayer: last cut layer %q Layr not found in cut itemList", lastLayer.Name)
		}
		insertIdx = lastLayrIdx + 2
		for insertIdx < len(cutChildren) && !cutChildren[insertIdx].IsList() {
			insertIdx++
		}
	}

	// 10. Splice block back into cutChildren at insertIdx.
	newChildren := make([]*rifx.Chunk, 0, len(children))
	newChildren = append(newChildren, cutChildren[:insertIdx]...)
	newChildren = append(newChildren, block...)
	newChildren = append(newChildren, cutChildren[insertIdx:]...)

	// 11. Splice source back into cutLayers at to.
	newLayers := make([]*Layer, 0, n)
	newLayers = append(newLayers, cutLayers[:to]...)
	newLayers = append(newLayers, source)
	newLayers = append(newLayers, cutLayers[to:]...)

	// 12. Apply.
	c.back.itemList.Children = newChildren
	c.Layers = newLayers
	for i, l := range c.Layers {
		l.Index = i
	}

	// 13. Warnings-as-failure (Inv-11). Defensive — no re-parse here, but
	//     pattern stays consistent with DeleteLayer/DuplicateLayer.
	if c.proj != nil && len(c.proj.Warnings) > oldWarningsLen {
		c.back.itemList.Children = oldChildren
		c.Layers = oldLayers
		for i, l := range c.Layers {
			l.Index = oldIndexes[i]
		}
		newWarnings := append([]string(nil), c.proj.Warnings[oldWarningsLen:]...)
		c.proj.Warnings = c.proj.Warnings[:oldWarningsLen]
		return fmt.Errorf("MoveLayer: produced %d parser warning(s), rolled back: %v", len(newWarnings), newWarnings)
	}

	return nil
}

// indexOfChunk returns the slice index of target in children by pointer
// identity, or -1 if not present.
func indexOfChunk(children []*rifx.Chunk, target *rifx.Chunk) int {
	for i, c := range children {
		if c == target {
			return i
		}
	}
	return -1
}
