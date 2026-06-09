package scene

import (
	"encoding/binary"
	"math"

	"github.com/example/aep-parser/internal/codec"
)

// Guide length-preserving setters. Each patches the guide's 16-byte scene-owned
// copy (block — the single source of truth), so no chunk size changes. The
// serializer-side syncGuides copies the mutated copies back into each comp's
// Gide ldat at WriteAEP time.

// SetPosition sets the guide's pixel offset (>= 0). No-op when the guide has no
// backing block (built outside the parser) or position is negative.
func (g *Guide) SetPosition(px float64) {
	if g == nil || len(g.block) < codec.GuideItemSize || px < 0 {
		return
	}
	binary.BigEndian.PutUint64(g.block[8:16], math.Float64bits(px))
	g.Position = px
}

// SetOrientation sets the guide's orientation (GuideHorizontal / GuideVertical).
// No-op when the guide has no backing block or the orientation is unrecognized.
func (g *Guide) SetOrientation(o GuideOrientation) {
	if g == nil || len(g.block) < codec.GuideItemSize {
		return
	}
	if o != GuideHorizontal && o != GuideVertical {
		return
	}
	binary.BigEndian.PutUint32(g.block[0:4], uint32(o))
	g.Orientation = o
}
