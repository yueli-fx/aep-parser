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

// @summary     Set the guide's pixel offset
// @param       px  the new offset in pixels from the top or left edge (must be >= 0)
// @domain      comp
// @stability   alpha
// @verify      roundtrip
// @since       AE2020
// @boundary    length-preserving: patches the guide's 16-byte scene-owned
//   block, which is copied back into the composition's ldat at write time.
//   Silently no-ops when the guide has no backing block (built outside the
//   parser) or px is negative.
// @alias       guide position,参考线位置,ruler guide,标尺参考线
func (g *Guide) SetPosition(px float64) {
	if g == nil || len(g.block) < codec.GuideItemSize || px < 0 {
		return
	}
	binary.BigEndian.PutUint64(g.block[8:16], math.Float64bits(px))
	g.Position = px
}

// @summary     Set the guide's orientation
// @param       o  the new orientation (GuideHorizontal or GuideVertical)
// @domain      comp
// @stability   alpha
// @verify      roundtrip
// @since       AE2020
// @boundary    length-preserving: patches the guide's 16-byte scene-owned
//   block, which is copied back into the composition's ldat at write time.
//   Silently no-ops when the guide has no backing block (built outside the
//   parser) or o is not a recognized orientation value.
// @alias       guide orientation,参考线方向,horizontal guide,vertical guide,水平参考线,垂直参考线
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
