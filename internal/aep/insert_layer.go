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
	return nil, fmt.Errorf("InsertLayer: not yet implemented")
}
