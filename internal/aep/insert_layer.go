package aep

import (
	"fmt"
)

// InsertLayer deep-clones src (from a sibling comp in the SAME Project)
// into c.Layers at atIdx. ALPHA — pending AE 2020+2025 ship-gate
// (3 modes × 2 versions = 6 PASS) before Stable promotion. See
// flightdeck/specs/2026-05-29-v3-phase5c-insertlayer-design.md.
func (c *Composition) InsertLayer(src *Layer, atIdx int) (*Layer, error) {
	if src == nil {
		return nil, fmt.Errorf("InsertLayer: src cannot be nil")
	}
	if c.back == nil || c.back.itemList == nil {
		return nil, fmt.Errorf("InsertLayer: dest comp %q has no itemList back-ref (built outside parser?)", c.Name)
	}
	if c.proj == nil {
		return nil, fmt.Errorf("InsertLayer: dest comp %q has no project back-ref", c.Name)
	}
	if atIdx < 0 || atIdx > len(c.Layers) {
		return nil, fmt.Errorf("InsertLayer: atIdx %d out of range (have %d layers; %d is append)", atIdx, len(c.Layers), len(c.Layers))
	}
	if src.comp == nil {
		return nil, fmt.Errorf("InsertLayer: src.comp is nil (layer detached from any comp)")
	}
	if src.comp == c {
		return nil, fmt.Errorf("InsertLayer: src and dest are the same comp %q — use DuplicateLayer instead", c.Name)
	}
	if src.comp.proj != c.proj {
		return nil, fmt.Errorf("InsertLayer: src and dest in different Projects — cross-Project insert deferred to Phase 5C.1")
	}
	return nil, fmt.Errorf("InsertLayer: not yet implemented")
}
