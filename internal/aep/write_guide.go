package aep

import (
	"encoding/binary"
	"math"

	"github.com/example/aep-parser/internal/rifx"
)

// write_guide.go — length-preserving setters for existing composition guides.
// Each patches the guide's 16-byte scene-owned copy (block — the single source
// of truth), so no chunk size changes. syncGuides copies the mutated copies back
// into each comp's Gide ldat (paired by index) at WriteAEP time. Non-structural
// → no AE ship-gate (same model as the other Set* value patches). Adding /
// removing guides is structural and deferred.
//
// Concurrency: like all Set* patches these mutate shared scene buffers; callers
// serialize their own access (see incidents/concurrency-unsafe-shared-chunk-bytes).

// syncGuides copies each composition's scene-owned guide buffers back into the
// owning Gide ldat (single source of truth → chunk) before serialization,
// pairing guides to 16-byte ldat slots by index. Guide has no per-object
// back-ref (§F D-U2); the comp serialization path re-derives the ldat from the
// comp's owning Item LIST. Length-preserving: an unmutated buffer equals the
// parsed bytes, so a parse→write round-trip is unchanged. Called by
// Project.WriteAEP; no-op for comps/guides built outside the parser.
func (p *Project) syncGuides() {
	for _, c := range p.Compositions {
		if len(c.Guides) == 0 {
			continue
		}
		cb, ok := c.back.(*compositionBackrefs)
		if !ok || cb.itemList == nil {
			continue
		}
		gide := cb.itemList.FindFirstList(rifx.IDGide)
		if gide == nil {
			continue
		}
		list := gide.FindFirstList(rifx.IDkfl)
		if list == nil {
			continue
		}
		ldat := list.FindFirst(rifx.IDLdat)
		if ldat == nil {
			continue
		}
		for i, g := range c.Guides {
			off := i * guideItemSize
			if len(g.block) != guideItemSize || off+guideItemSize > len(ldat.Data) {
				continue
			}
			copy(ldat.Data[off:off+guideItemSize], g.block)
		}
	}
}

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
