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
//
//aep:cap domain=comp tier=alpha verify=roundtrip boundary="length-preserving(16B scene-owned block;syncGuides 在 WriteAEP 时回写 ldat);负值/无 block 时静默 no-op;无专门 AE gate→round-trip" alias="guide position,参考线位置,ruler guide,标尺参考线"
func (g *Guide) SetPosition(px float64) {
	if g == nil || len(g.block) < codec.GuideItemSize || px < 0 {
		return
	}
	binary.BigEndian.PutUint64(g.block[8:16], math.Float64bits(px))
	g.Position = px
}

// SetOrientation sets the guide's orientation (GuideHorizontal / GuideVertical).
// No-op when the guide has no backing block or the orientation is unrecognized.
//
//aep:cap domain=comp tier=alpha verify=roundtrip boundary="length-preserving(16B scene-owned block;syncGuides 在 WriteAEP 时回写 ldat);非法 orientation 值静默 no-op;无专门 AE gate→round-trip" alias="guide orientation,参考线方向,horizontal guide,vertical guide,水平参考线,垂直参考线"
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
