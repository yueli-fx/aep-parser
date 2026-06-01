package aep

import (
	"encoding/binary"
	"math"
)

// write_guide.go — length-preserving setters for existing composition guides.
// Each patches the guide's 16-byte slot inside the shared Gide ldat (block
// aliases the chunk bytes), so no chunk size changes and WriteAEP re-emits the
// mutation. Non-structural → no AE ship-gate (same model as the other Set*
// value patches). Adding / removing guides is structural and deferred.
//
// Concurrency: like all Set* patches these mutate shared chunk bytes; callers
// serialize their own access (see incidents/concurrency-unsafe-shared-chunk-bytes).

// SetPosition sets the guide's pixel offset (>= 0). No-op when the guide has no
// backing block (built outside the parser) or position is negative.
func (g *Guide) SetPosition(px float64) {
	if g == nil || len(g.block) < guideItemSize || px < 0 {
		return
	}
	binary.BigEndian.PutUint64(g.block[8:16], math.Float64bits(px))
	g.Position = px
}

// SetOrientation sets the guide's orientation (GuideHorizontal / GuideVertical).
// No-op when the guide has no backing block or the orientation is unrecognized.
func (g *Guide) SetOrientation(o GuideOrientation) {
	if g == nil || len(g.block) < guideItemSize {
		return
	}
	if o != GuideHorizontal && o != GuideVertical {
		return
	}
	binary.BigEndian.PutUint32(g.block[0:4], uint32(o))
	g.Orientation = o
}
