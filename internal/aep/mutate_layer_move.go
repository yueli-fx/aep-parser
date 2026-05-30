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

// Layer-level convenience wrappers over Composition.MoveLayer, mirroring
// AE ScriptingAPI's layer.moveAfter / moveBefore / moveToBeginning /
// moveToEnd. All delegate to the comp's MoveLayer (already ship-gated
// against AE 2020 + AE 2025; see scars/ae-deletelayer-re.md F4 + the
// move_layer ship-gate).
//
// Each wrapper finds the current slice index of the receiver via pointer
// identity in `l.comp.Layers`; this avoids relying on `Layer.Index`,
// which is parse-time and may be stale if the comp was previously
// mutated by DeleteLayer / DuplicateLayer (those don't re-index, see
// V3 Phase 4 plan §1 commentary).

// MoveToBeginning moves the receiver to position 0 (top of layer stack
// in AE's display, AE-index 1).
func (l *Layer) MoveToBeginning() error {
	c, idx, err := l.locateInComp("MoveToBeginning")
	if err != nil {
		return err
	}
	return c.MoveLayer(idx, 0)
}

// MoveToEnd moves the receiver to the last position in c.Layers
// (bottom of layer stack in AE's display, AE-index c.numLayers).
func (l *Layer) MoveToEnd() error {
	c, idx, err := l.locateInComp("MoveToEnd")
	if err != nil {
		return err
	}
	return c.MoveLayer(idx, len(c.Layers)-1)
}

// MoveAfter moves the receiver to the slot immediately after `other`
// (i.e., other.Index < receiver.Index post-call, both viewed in
// c.Layers slice order — receiver lands just below other in the stack).
// Returns an error if other belongs to a different comp, other == l,
// or either layer is missing a comp back-ref.
func (l *Layer) MoveAfter(other *Layer) error {
	c, fromIdx, otherIdx, err := l.locatePair("MoveAfter", other)
	if err != nil {
		return err
	}
	// Target slice index: position of `other` after we've cut `l`.
	// If l is currently before other (fromIdx < otherIdx), cutting l
	// shifts other down by 1 — and we want to land at otherIdx (which
	// puts l immediately after other's new position). MoveLayer's `to`
	// argument is the FINAL index in c.Layers, so target = otherIdx.
	// If l is currently after other (fromIdx > otherIdx), cutting l
	// doesn't shift other — target = otherIdx + 1 to land right after.
	var to int
	if fromIdx < otherIdx {
		to = otherIdx
	} else {
		to = otherIdx + 1
	}
	return c.MoveLayer(fromIdx, to)
}

// MoveBefore moves the receiver to the slot immediately before `other`
// (receiver lands just above other in the stack).
func (l *Layer) MoveBefore(other *Layer) error {
	c, fromIdx, otherIdx, err := l.locatePair("MoveBefore", other)
	if err != nil {
		return err
	}
	// Final index = position of other minus the post-cut shift.
	// If l < other (cutting l shifts other down by 1): target = otherIdx - 1
	// If l > other (cutting l doesn't shift other):    target = otherIdx
	var to int
	if fromIdx < otherIdx {
		to = otherIdx - 1
	} else {
		to = otherIdx
	}
	return c.MoveLayer(fromIdx, to)
}

func (l *Layer) locateInComp(op string) (*Composition, int, error) {
	if l.comp == nil {
		return nil, 0, fmt.Errorf("%s: layer %q has no comp back-ref (built outside parser?)", op, l.Name)
	}
	for i, other := range l.comp.Layers {
		if other == l {
			return l.comp, i, nil
		}
	}
	return nil, 0, fmt.Errorf("%s: layer %q not present in its own comp %q (parse-tree inconsistency)", op, l.Name, l.comp.Name)
}

func (l *Layer) locatePair(op string, other *Layer) (*Composition, int, int, error) {
	if other == nil {
		return nil, 0, 0, fmt.Errorf("%s: other layer is nil", op)
	}
	if other == l {
		return nil, 0, 0, fmt.Errorf("%s: cannot %s self", op, op)
	}
	c, fromIdx, err := l.locateInComp(op)
	if err != nil {
		return nil, 0, 0, err
	}
	if other.comp != c {
		return nil, 0, 0, fmt.Errorf("%s: other layer %q belongs to a different comp", op, other.Name)
	}
	for i, x := range c.Layers {
		if x == other {
			return c, fromIdx, i, nil
		}
	}
	return nil, 0, 0, fmt.Errorf("%s: other layer %q not present in comp %q", op, other.Name, c.Name)
}
