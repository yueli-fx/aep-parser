package aep

import "fmt"

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
