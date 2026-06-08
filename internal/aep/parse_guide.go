package aep

import (
	"encoding/binary"
	"math"

	"github.com/example/aep-parser/internal/rifx"
)

// parse_guide.go — decode composition ruler guides from the comp's Item-level
// LIST:Gide → LIST:list → ldat(GuideItem × N). Each GuideItem is 16 bytes BE:
//
//	@0x00 u4  orientation_type  (2 = horizontal, 1 = vertical)
//	@0x04 u4  position_type     (always 0 = pixels)
//	@0x08 f8  position          (pixels from top/left edge)
//
// The layer-side LIST:Gide (a Layr child carrying an empty gdta) uses the same
// formType but is nested under DLay/SLay, so FindFirstList on the Item LIST only
// sees the comp-level guides container.

const guideItemSize = 16

// parseGuides reads the Item-level guides for a comp. item is the comp's owning
// Item LIST chunk. Returns nil when the comp has no guides.
func parseGuides(item *rifx.Chunk) []*Guide {
	gide := item.FindFirstList(rifx.IDGide)
	if gide == nil {
		return nil
	}
	list := gide.FindFirstList(rifx.IDkfl)
	if list == nil {
		return nil
	}
	ldat := list.FindFirst(rifx.IDLdat)
	if ldat == nil {
		return nil
	}
	n := len(ldat.Data) / guideItemSize
	if n == 0 {
		return nil
	}
	guides := make([]*Guide, 0, n)
	for i := 0; i < n; i++ {
		block := append([]byte(nil), ldat.Data[i*guideItemSize:(i+1)*guideItemSize]...)
		guides = append(guides, &Guide{
			Orientation: GuideOrientation(binary.BigEndian.Uint32(block[0:4])),
			Position:    math.Float64frombits(binary.BigEndian.Uint64(block[8:16])),
			block:       block,
		})
	}
	return guides
}
