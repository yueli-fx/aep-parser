package serializer

import (
	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// syncGuides copies each composition's scene-owned guide buffers back into the
// owning Gide ldat (single source of truth → chunk) before serialization,
// pairing guides to 16-byte ldat slots by index. Guide has no per-object
// back-ref (§F D-U2); the comp serialization path re-derives the ldat from the
// comp's owning Item LIST. Length-preserving: an unmutated buffer equals the
// parsed bytes, so a parse→write round-trip is unchanged. Called by WriteAEP;
// no-op for comps/guides built outside the parser. Free function (serializer
// stage): it reads the live Gide ldat through the CompositionWriter interface.
func syncGuides(p *Project) {
	for _, c := range p.Compositions {
		cb := scene.CompositionBack(c)
		if len(c.Guides) == 0 || cb == nil {
			continue
		}
		ldat := cb.GuideLdatData()
		if ldat == nil {
			continue
		}
		for i, g := range c.Guides {
			off := i * codec.GuideItemSize
			block := scene.GuideBlock(g)
			if len(block) != codec.GuideItemSize || off+codec.GuideItemSize > len(ldat) {
				continue
			}
			copy(ldat[off:off+codec.GuideItemSize], block)
		}
	}
}
